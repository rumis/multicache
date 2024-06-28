package local

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/coocood/freecache"
	"github.com/rumis/multicache/adaptor"
	"github.com/rumis/multicache/logger"
	"github.com/rumis/multicache/metrics"
	"github.com/rumis/multicache/syncer"
	"github.com/rumis/multicache/utils"
)

// 类型检测
var _ adaptor.Adaptor[string, adaptor.Metadata] = (*FreeCache[string, adaptor.Metadata])(nil)

// FreeCache 基于freecache的本地缓存实现
type FreeCache[K comparable, V adaptor.Metadata] struct {
	// freechache对象
	innerCache *freecache.Cache
	// key前缀
	prefix string
	// 过期时间
	ttl time.Duration
	// 过期时间噪音
	threshold time.Duration
	// 标志是否跳过获取过程
	skipGet      bool
	preAdaptor   adaptor.Adaptor[K, V]
	name         string
	solutionName string
	ttlZero      time.Duration
	syncer       syncer.Syncer
}

// NewFreeCache 创建一个新的FreeCache对象
func NewFreeCache[K comparable, V adaptor.Metadata](icache *freecache.Cache, preAdaptor adaptor.Adaptor[K, V], fns ...LocalCacheOptionFunc) *FreeCache[K, V] {
	// 默认+自定义配置
	opts := DefaultLocalCacheOption()
	for _, fn := range fns {
		fn(&opts)
	}

	cacheInst := &FreeCache[K, V]{
		innerCache:   icache,
		prefix:       opts.Prefix,
		ttl:          opts.TTL,
		threshold:    opts.Threshold,
		skipGet:      opts.SkipGet,
		preAdaptor:   preAdaptor,
		name:         opts.Name,
		solutionName: opts.SolutionName,
		ttlZero:      opts.TTLZero,
		syncer:       opts.Syncer,
	}

	// 订阅数据同步事件
	if cacheInst.syncer != nil {
		cacheInst.syncer.Subscribe(context.Background(), cacheInst.sync)
	}

	return cacheInst
}

// Name 适配器名称
func (c *FreeCache[K, V]) Name() string {
	return c.name
}

// Get 读取对象
func (c *FreeCache[K, V]) Get(ctx context.Context, key K, value V) (bool, error) {
	startTime := time.Now()
	// 跳过
	if c.skipGet {
		metrics.AddMeta(ctx, metrics.Meta{
			AdaptorName: c.Name(),
			Key:         fmt.Sprint(key),
			Type:        metrics.Miss,
		})
		return false, nil
	}

	buf, err := c.innerCache.Get(utils.Bytes(c.key(key)))
	if errors.Is(err, freecache.ErrNotFound) {
		metrics.AddMeta(ctx, metrics.Meta{
			AdaptorName: c.Name(),
			Key:         fmt.Sprint(key),
			Type:        metrics.Miss,
		})
		// key不存在
		return false, nil
	}
	if err != nil {
		metrics.AddMeta(ctx, metrics.Meta{
			AdaptorName: c.Name(),
			Key:         fmt.Sprint(key),
			Type:        metrics.Miss,
		})
		return false, err
	}
	// 反序列化对象
	value.Decode(buf)

	metrics.AddMeta(ctx, metrics.Meta{
		AdaptorName: c.Name(),
		Key:         fmt.Sprint(key),
		Type:        metrics.Hit,
		TrackTime:   time.Since(startTime).Milliseconds(),
	})

	// 数据回写
	if c.preAdaptor != nil {
		err := c.preAdaptor.Set(ctx, value)
		if err != nil {
			// 回写失败 只记录错误，不影响主流程
			logger.Error(err.Error(), "solution", c.solutionName, "adaptor", c.Name(), "value", value, "event", adaptor.LogEventRefill)
		}
	}
	return true, nil
}

// Set 写入对象
func (c *FreeCache[K, V]) Set(ctx context.Context, value V) error {
	startTime := time.Now()
	ttl := int(c.ttl.Seconds()) + utils.SafeRand().Intn(int(c.threshold.Seconds())) // 正常TTL
	ttl = utils.IfExpr(value.Zero(), int(c.ttlZero.Seconds()), ttl)                 // 0值TTL

	valBuf, err := value.Value()
	if err != nil {
		return err
	}
	err = c.innerCache.Set(utils.Bytes(c.key1(value.Key())), valBuf, ttl)
	if err != nil {
		return err
	}
	// 缓存数据同步
	if c.syncer != nil {
		setEvent := &syncer.CacheSyncEvent{
			EventType: syncer.EventTypeAdd,
			Key:       c.key1(value.Key()),
			Val:       valBuf,
			TTL:       time.Duration(ttl) * time.Second,
		}
		err := c.syncer.Emit(ctx, setEvent)
		if err != nil {
			logger.Error(err.Error(), "solution", c.solutionName, "adaptor", c.Name(), "value", setEvent.Encode(), "event", adaptor.LogEventSync)
		}
	}

	metrics.AddMeta(ctx, metrics.Meta{
		AdaptorName: c.Name(),
		Key:         value.Key(),
		Type:        metrics.Set,
		TrackTime:   time.Since(startTime).Milliseconds(),
	})
	return nil
}

// Del 删除对象
func (c *FreeCache[K, V]) Del(ctx context.Context, key K) error {
	// 删除本地缓存
	c.innerCache.Del(utils.Bytes(c.key(key)))

	// 删除缓存数据操作同步
	if c.syncer != nil {
		delEvent := &syncer.CacheSyncEvent{
			EventType: syncer.EventTypeDelete,
			Key:       c.key(key),
		}
		err := c.syncer.Emit(ctx, delEvent)
		if err != nil {
			logger.Error(err.Error(), "solution", c.solutionName, "adaptor", c.Name(), "value", delEvent.Encode(), "event", adaptor.LogEventSync)
		}
	}

	return nil
}

// sync 数据同步
func (c *FreeCache[K, V]) sync(e *syncer.CacheSyncEvent) {
	if c.syncer == nil {
		return
	}
	if e.ClientID == c.syncer.ClientID() {
		return
	}
	switch e.EventType {
	case syncer.EventTypeAdd:
		ttl := int(e.TTL.Seconds())
		err := c.innerCache.Set(utils.Bytes(e.Key), e.Val, ttl)
		if err != nil {
			logger.Error(err.Error(), "solution", c.solutionName, "adaptor", c.Name(), "value", e, "event", adaptor.LogEventSyncAdd)
		}
	case syncer.EventTypeDelete:
		c.innerCache.Del(utils.Bytes(e.Key))
	}
}

// key 生成缓存key
func (c *FreeCache[K, V]) key(key K) string {
	return c.key1(fmt.Sprint(key))
}

// key1 生成缓存key
func (c *FreeCache[K, V]) key1(key string) string {
	return c.prefix + key
}
