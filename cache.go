package multicache

import (
	"context"

	"github.com/rumis/multicache/adaptor"
	"github.com/rumis/multicache/logger"
	"github.com/rumis/multicache/metrics"
	"github.com/rumis/multicache/utils"
)

type Cache[K comparable, V adaptor.Metadata] struct {
	name     string
	adaptors []adaptor.Adaptor[K, V]
}

// NewCache 创建一个新的Cache对象
func NewCache[K comparable, V adaptor.Metadata](name string, adaptors ...adaptor.Adaptor[K, V]) *Cache[K, V] {
	return &Cache[K, V]{
		name:     name,
		adaptors: adaptors,
	}
}

// Get 读取对象
func (c *Cache[K, V]) Get(ctx context.Context, key K, value V) (bool, error) {

	ctx = context.WithValue(ctx, metrics.MetricsTraceKey, utils.UUID())
	metrics.Start(ctx, c.name)

	for _, adap := range c.adaptors {
		ok, err := adap.Get(ctx, key, value)
		if err != nil {
			logger.Error(err.Error(), "solution", c.name, "adaptor", adap.Name(), "key", key, "event", adaptor.LogEventGet)
		}
		if ok {
			metrics.Summary(ctx)
			return !value.Zero(), nil
		}
	}
	metrics.Summary(ctx)
	return false, nil
}

// Set 向缓存中写入对象
func (c *Cache[K, V]) Set(ctx context.Context, value V) error {
	for _, adap := range c.adaptors {
		err := adap.Set(ctx, value)
		if err != nil {
			logger.Error(err.Error(), "solution", c.name, "adaptor", adap.Name(), "value", value, "event", adaptor.LogEventSet)
			return err
		}
	}
	return nil
}

// Del 删除缓存对象
func (c *Cache[K, V]) Del(ctx context.Context, key K) error {
	for _, adap := range c.adaptors {
		err := adap.Del(ctx, key)
		if err != nil {
			logger.Error(err.Error(), "solution", c.name, "adaptor", adap.Name(), "key", key, "event", adaptor.LogEventDel)
			return err
		}
	}
	return nil
}
