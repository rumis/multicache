package remote

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/rumis/multicache/adaptor"
	"github.com/rumis/multicache/logger"
	"github.com/rumis/multicache/metrics"
	"github.com/rumis/multicache/utils"
)

// 类型检测
var _ adaptor.MultiAdaptor[string, adaptor.Metadata] = (*RedisMultiAdaptor[string, adaptor.Metadata])(nil)

// RedisMultiAdaptor 基于Redis的分布式多值缓存适配实现
type RedisMultiAdaptor[K comparable, V adaptor.Metadata] struct {
	rClient      *redis.Client
	prefix       string
	ttl          time.Duration
	threshold    int
	skipGet      bool
	preAdaptor   adaptor.MultiAdaptor[K, V]
	name         string
	solutionName string
	ttlZero      time.Duration
}

// NewRedisMultiAdaptor 基于Redis的多值缓存对象
func NewRedisMultiAdaptor[K comparable, V adaptor.Metadata](client *redis.Client, preAdaptor adaptor.MultiAdaptor[K, V], fns ...RemoteCacheOptionFunc) *RedisMultiAdaptor[K, V] {
	// 默认+自定义配置
	opts := DefaultRemoteCacheOption()
	for _, fn := range fns {
		fn(&opts)
	}
	return &RedisMultiAdaptor[K, V]{
		rClient:      client,
		prefix:       opts.Prefix,
		ttl:          opts.TTL,
		threshold:    opts.Threshold,
		skipGet:      opts.SkipGet,
		preAdaptor:   preAdaptor,
		name:         opts.Name,
		solutionName: opts.SolutionName,
		ttlZero:      opts.TTLZero,
	}
}

// Name 适配器名称
func (c *RedisMultiAdaptor[K, V]) Name() string {
	return c.name
}

// Get 读取对象
// 暂不支持pipeline操作(集群部署)
func (c *RedisMultiAdaptor[K, V]) Get(ctx context.Context, keys adaptor.Keys[K], vals adaptor.Values[K, V], fn adaptor.NewValueFunc[V]) (adaptor.Keys[K], error) {
	hasKeys := make(adaptor.Keys[K], 0)
	hasValues := make(adaptor.ValueCol[V], 0)
	for _, key := range keys {
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
			continue
		}
		if err != nil {
			metrics.AddMeta(ctx, missMeta)
			logger.Error(err.Error(), "solution", c.solutionName, "adaptor", c.Name(), "key", key, "event", adaptor.LogEventGet)
			continue
		}
		// 反序列化对象
		val := fn()
		err = val.Decode(buf)
		if err != nil {
			metrics.AddMeta(ctx, missMeta)
			logger.Error(err.Error(), "solution", c.solutionName, "adaptor", c.Name(), "key", key, "event", adaptor.LogEventGet)
			continue
		}

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

	// 数据回写
	if c.preAdaptor != nil && len(hasValues) > 0 {
		err := c.preAdaptor.Set(ctx, hasValues)
		if err != nil {
			logger.Error(err.Error(), "solution", c.solutionName, "adaptor", c.Name(), "value", hasValues, "event", adaptor.LogEventRefill)
		}
	}

	return hasKeys, nil
}

// Set 写入对象
func (c *RedisMultiAdaptor[K, V]) Set(ctx context.Context, vals adaptor.ValueCol[V]) error {
	for _, val := range vals {
		startTime := time.Now()
		// 序列化对象
		key := val.Key()
		buf, err := val.Value()
		if err != nil {
			logger.Error(err.Error(), "solution", c.solutionName, "adaptor", c.Name(), "value", val, "event", adaptor.LogEventSet)
			continue
		}

		ttl := c.ttl + time.Second*time.Duration(utils.SafeRand().Intn(c.threshold))
		ttl = utils.IfExpr(val.Zero(), c.ttlZero, ttl)

		if err := c.rClient.Set(ctx, c.key1(key), buf, ttl).Err(); err != nil {
			logger.Error(err.Error(), "solution", c.solutionName, "adaptor", c.Name(), "value", val, "event", adaptor.LogEventSet)
			continue
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
func (c *RedisMultiAdaptor[K, V]) Del(ctx context.Context, keys adaptor.Keys[K]) error {
	for _, key := range keys {
		if err := c.rClient.Del(ctx, c.key(key)).Err(); err != nil {
			return err
		}
	}
	return nil
}

// key 生成缓存key
func (c *RedisMultiAdaptor[K, V]) key(key K) string {
	return c.key1(fmt.Sprint(key))
}

// key1 生成缓存key
func (c *RedisMultiAdaptor[K, V]) key1(key string) string {
	return c.prefix + key
}
