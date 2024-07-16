package multicache

import (
	"context"

	"github.com/rumis/multicache/adaptor"
	"github.com/rumis/multicache/logger"
	"github.com/rumis/multicache/metrics"
	"github.com/rumis/multicache/utils"
)

type MultiCache[K comparable, V adaptor.Metadata] struct {
	name     string
	adaptors []adaptor.MultiAdaptor[K, V]
	metric   metrics.Metrics
}

// NewMultiCache 创建一个新的MultiCache对象
func NewMultiCache[K comparable, V adaptor.Metadata](name string, adaptors ...adaptor.MultiAdaptor[K, V]) *MultiCache[K, V] {
	return &MultiCache[K, V]{
		name:     name,
		adaptors: adaptors,
		metric:   metrics.DefaultMetrics(),
	}
}

// NewMultiCacheWithMetric 创建一个新的MultiCache对象，包含自定义指标计数器
func NewMultiCacheWithMetric[K comparable, V adaptor.Metadata](name string, metric metrics.Metrics, adaptors ...adaptor.MultiAdaptor[K, V]) *MultiCache[K, V] {
	return &MultiCache[K, V]{
		name:     name,
		adaptors: adaptors,
		metric:   metric,
	}
}

// Get 读取对象
func (c *MultiCache[K, V]) Get(ctx context.Context, keys adaptor.Keys[K], vals adaptor.Values[K, V], fn adaptor.NewValueFunc[V]) error {

	ctx = context.WithValue(ctx, metrics.MetricsTraceKey, utils.UUID())
	ctx = context.WithValue(ctx, metrics.MetricsClient, c.metric)
	c.metric.Start(ctx, c.name)

	tmpKeys := keys
	for _, adap := range c.adaptors {
		_, err := adap.Get(ctx, tmpKeys, vals, fn)
		if err != nil {
			// 错误日志
			logger.Error(err.Error(), "solution", c.name, "adaptor", adap.Name(), "key", tmpKeys, "event", adaptor.LogEventGet)
		}

		if len(vals) == len(keys) {
			// 读取到了所有数据
			c.metric.Summary(ctx)

			// 剔除零值
			for key, val := range vals {
				if val.Zero() {
					delete(vals, key)
				}
			}

			return nil
		}

		// 剩余需要读取的数据
		missKeys := make([]K, 0, len(keys)-len(vals))
		for _, key := range keys {
			if _, ok := vals[key]; !ok {
				missKeys = append(missKeys, key)
			}
		}
		tmpKeys = missKeys
	}

	c.metric.Summary(ctx)

	// 剔除零值
	for key, val := range vals {
		if val.Zero() {
			delete(vals, key)
		}
	}

	return nil
}

// Set 向缓存中写入对象
func (c *MultiCache[K, V]) Set(ctx context.Context, vals adaptor.ValueCol[V]) error {

	ctx = context.WithValue(ctx, metrics.MetricsTraceKey, utils.UUID())
	ctx = context.WithValue(ctx, metrics.MetricsClient, c.metric)
	c.metric.Start(ctx, c.name)

	for _, adap := range c.adaptors {
		err := adap.Set(ctx, vals)
		if err != nil {
			logger.Error(err.Error(), "solution", c.name, "adaptor", adap.Name(), "value", vals, "event", adaptor.LogEventSet)
			return err
		}
	}
	c.metric.Summary(ctx)
	return nil
}

// Del 删除缓存对象
func (c *MultiCache[K, V]) Del(ctx context.Context, keys adaptor.Keys[K]) error {
	for _, adap := range c.adaptors {
		err := adap.Del(ctx, keys)
		if err != nil {
			logger.Error(err.Error(), "solution", c.name, "adaptor", adap.Name(), "key", keys, "event", adaptor.LogEventDel)
			return err
		}
	}
	return nil
}
