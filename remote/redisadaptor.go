package remote

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rumis/multicache/adaptor"
	"github.com/rumis/multicache/logger"
	"github.com/rumis/multicache/metrics"
	"github.com/rumis/multicache/utils"
)

// 类型检测
var _ adaptor.Adaptor[string, adaptor.Metadata] = (*RedisAdaptor[string, adaptor.Metadata])(nil)

// RedisAdaptor 基于Redis的分布式缓存适配实现
type RedisAdaptor[K comparable, V adaptor.Metadata] struct {
	rClient      *redis.Client
	prefix       string
	ttl          time.Duration
	threshold    int
	preAdaptor   adaptor.Adaptor[K, V]
	name         string
	solutionName string
	ttlZero      time.Duration
}

// NewRedisAdaptor 创建一个新的RedisAdaptor对象
func NewRedisAdaptor[K comparable, V adaptor.Metadata](client *redis.Client, preAdaptor adaptor.Adaptor[K, V], fns ...RemoteCacheOptionFunc) *RedisAdaptor[K, V] {
	// 默认+自定义配置
	opts := DefaultRemoteCacheOption()
	for _, fn := range fns {
		fn(&opts)
	}
	return &RedisAdaptor[K, V]{
		rClient:      client,
		prefix:       opts.Prefix,
		ttl:          opts.TTL,
		threshold:    opts.Threshold,
		name:         opts.Name,
		solutionName: opts.SolutionName,
		preAdaptor:   preAdaptor,
		ttlZero:      opts.TTLZero,
	}
}

// Name 适配器名称，需要在当前业务场景中保证唯一
func (c *RedisAdaptor[K, V]) Name() string {
	return c.name
}

// Get 读取对象
func (c *RedisAdaptor[K, V]) Get(ctx context.Context, key K, value V) (bool, error) {
	startTime := time.Now()
	missMeta := metrics.Meta{
		AdaptorName: c.Name(),
		Key:         fmt.Sprint(key),
		Type:        metrics.Miss,
	}

	buf, err := c.rClient.Get(ctx, c.key(key)).Bytes()
	if errors.Is(err, redis.Nil) {
		// key不存在
		metrics.AddMeta(ctx, missMeta)
		return false, nil
	}
	if err != nil {
		metrics.AddMeta(ctx, missMeta)
		return false, err
	}
	// 反序列化对象
	err = value.Decode(buf)
	if err != nil {
		metrics.AddMeta(ctx, missMeta)
		return false, err
	}

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
			logger.Error(err.Error(), "solution", c.solutionName, "adaptor", c.Name(), "value", value, "event", adaptor.LogEventRefill)
		}
	}

	// 返回
	return true, nil
}

// Set 写入对象
func (c *RedisAdaptor[K, V]) Set(ctx context.Context, value V) error {
	startTime := time.Now()
	ttl := c.ttl + time.Second*time.Duration(utils.SafeRand().Intn(c.threshold))
	ttl = utils.IfExpr(value.Zero(), c.ttlZero, ttl)

	valBuf, err := value.Value()
	if err != nil {
		return err
	}

	err = c.rClient.SetEx(ctx, c.key1(value.Key()), utils.String(valBuf), ttl).Err()

	metrics.AddMeta(ctx, metrics.Meta{
		AdaptorName: c.Name(),
		Key:         value.Key(),
		Type:        metrics.Set,
		TrackTime:   time.Since(startTime).Milliseconds(),
	})

	return err
}

// Del 删除对象
func (c *RedisAdaptor[K, V]) Del(ctx context.Context, key K) error {
	err := c.rClient.Del(ctx, c.key(key)).Err()
	return err
}

// key 生成缓存key
func (c *RedisAdaptor[K, V]) key(key K) string {
	return c.key1(fmt.Sprint(key))
}

// key1 生成缓存key
func (c *RedisAdaptor[K, V]) key1(key string) string {
	return c.prefix + key
}
