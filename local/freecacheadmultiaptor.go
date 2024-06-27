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
var _ adaptor.MultiAdaptor[string, adaptor.Metadata] = (*MultiFreeCache[string, adaptor.Metadata])(nil)

// FreeCache 基于freecache的本地多值缓存实现
type MultiFreeCache[K comparable, V adaptor.Metadata] struct {
	innerCache   *freecache.Cache
	prefix       string
	ttl          time.Duration
	threshold    time.Duration
	skipGet      bool
	preAdaptor   adaptor.MultiAdaptor[K, V]
	solutionName string
	ttlZero      time.Duration
	syncer       syncer.Syncer
}

// NewMultiFreeCache 多值本地缓存
func NewMultiFreeCache[K comparable, V adaptor.Metadata](icache *freecache.Cache, preAdaptor adaptor.MultiAdaptor[K, V], fns ...LocalCacheOptionFunc) *MultiFreeCache[K, V] {
	// 默认+自定义配置
	opts := DefaultLocalCacheOption()
	for _, fn := range fns {
		fn(&opts)
	}

	multiCacheInst := &MultiFreeCache[K, V]{
		innerCache:   icache,
		prefix:       opts.Prefix,
		ttl:          opts.TTL,
		threshold:    opts.Threshold,
		skipGet:      opts.SkipGet,
		preAdaptor:   preAdaptor,
		solutionName: opts.SolutionName,
		ttlZero:      opts.TTLZero,
		syncer:       opts.Syncer,
	}

	// 订阅数据同步事件
	if multiCacheInst.syncer != nil {
		multiCacheInst.syncer.Subscribe(context.Background(), multiCacheInst.sync)
	}

	return multiCacheInst

}

// Name 适配器名称，需要在当前业务场景中保证唯一
func (c *MultiFreeCache[K, V]) Name() string {
	return "local_freecache"
}

// Get 读取对象
func (c *MultiFreeCache[K, V]) Get(ctx context.Context, keys adaptor.Keys[K], vals adaptor.Values[K, V], fn adaptor.NewValueFunc[V]) (adaptor.Keys[K], error) {
	hasKeys := make(adaptor.Keys[K], 0)
	hasValues := make(adaptor.ValueCol[V], 0)
	for _, key := range keys {
		startTime := time.Now()
		buf, err := c.innerCache.Get(utils.Bytes(c.key(key)))
		if errors.Is(err, freecache.ErrNotFound) {
			// key不存在
			metrics.AddMeta(ctx, metrics.Meta{
				AdaptorName: c.Name(),
				Key:         fmt.Sprint(key),
				Type:        metrics.Miss,
			})
			continue
		}
		if err != nil {
			return hasKeys, err
		}
		// 反序列化对象
		val := fn()
		val.Decode(buf)
		vals[key] = val
		hasKeys = append(hasKeys, key)
		hasValues = append(hasValues, val)

		metrics.AddMeta(ctx, metrics.Meta{
			AdaptorName: c.Name(),
			Key:         fmt.Sprint(key),
			Type:        metrics.Hit,
			TrackTime:   time.Since(startTime).Milliseconds(),
		})

	}

	if c.preAdaptor != nil && len(hasValues) > 0 {
		err := c.preAdaptor.Set(ctx, hasValues)
		if err != nil {
			logger.Error(err.Error(), "solution", c.solutionName, "adaptor", c.Name(), "value", hasValues, "event", adaptor.LogEventRefill)
		}
	}

	return hasKeys, nil
}

// Set 写入对象
func (c *MultiFreeCache[K, V]) Set(ctx context.Context, vals adaptor.ValueCol[V]) error {
	for _, val := range vals {
		startTime := time.Now()
		// 序列化对象
		key := val.Key()
		buf, err := val.Value()
		if err != nil {
			logger.Error(err.Error(), "solution", c.solutionName, "adaptor", c.Name(), "value", val, "event", adaptor.LogEventSet)
			continue
		}
		// 写入缓存
		ttl := int(c.ttl.Seconds()) + utils.SafeRand().Intn(int(c.threshold.Seconds()))
		ttl = utils.IfExpr(val.Zero(), int(c.ttlZero.Seconds()), ttl)
		err = c.innerCache.Set(utils.Bytes(c.key1(key)), buf, ttl)
		if err != nil {
			logger.Error(err.Error(), "solution", c.solutionName, "adaptor", c.Name(), "value", val, "event", adaptor.LogEventSet)
			continue
		}

		// 同步缓存数据写入操作
		if c.syncer != nil {
			setEvent := &syncer.CacheSyncEvent{
				EventType: syncer.EventTypeAdd,
				Key:       c.key1(key),
				Val:       buf,
				TTL:       time.Duration(ttl) * time.Second,
			}
			err := c.syncer.Emit(ctx, setEvent)
			if err != nil {
				logger.Error(err.Error(), "solution", c.solutionName, "adaptor", c.Name(), "value", setEvent.Encode(), "event", adaptor.LogEventSync)
			}
		}

		metrics.AddMeta(ctx, metrics.Meta{
			AdaptorName: c.Name(),
			Key:         fmt.Sprint(key),
			Type:        metrics.Set,
			TrackTime:   time.Since(startTime).Milliseconds(),
		})

	}
	return nil
}

// Del 删除对象
func (c *MultiFreeCache[K, V]) Del(ctx context.Context, keys adaptor.Keys[K]) error {
	for _, key := range keys {

		// 删除本地缓存数据
		c.innerCache.Del(utils.Bytes(c.key(key)))

		// 同步缓存数据删除操作
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

	}
	return nil
}

// sync 数据同步
func (c *MultiFreeCache[K, V]) sync(e *syncer.CacheSyncEvent) {
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
func (c *MultiFreeCache[K, V]) key(key K) string {
	return c.key1(fmt.Sprint(key))
}

// key1 生成缓存key
func (c *MultiFreeCache[K, V]) key1(key string) string {
	return c.prefix + key
}
