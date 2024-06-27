package datasource

import (
	"context"
	"fmt"
	"time"

	"github.com/rumis/multicache/adaptor"
	"github.com/rumis/multicache/logger"
	"github.com/rumis/multicache/metrics"
)

type MultiDataSourceFunc[K comparable, V adaptor.Metadata] func(keys adaptor.Keys[K]) (adaptor.Values[K, V], error)

// 类型检测
var _ adaptor.MultiAdaptor[string, adaptor.Metadata] = (*DataSourceMultiAdaptor[string, adaptor.Metadata])(nil)

// DataSourceMultiAdaptor 多值数据源适配器
type DataSourceMultiAdaptor[K comparable, V adaptor.Metadata] struct {
	solutionName string
	preAdaptor   adaptor.MultiAdaptor[K, V]
	dataSourceFn MultiDataSourceFunc[K, V]
}

// NewDataSourceMultiAdaptor 多值数据源适配器
func NewDataSourceMultiAdaptor[K comparable, V adaptor.Metadata](preAdaptor adaptor.MultiAdaptor[K, V], dsfn MultiDataSourceFunc[K, V], fns ...DataSourceOptionFunc) *DataSourceMultiAdaptor[K, V] {
	opts := DefaultDataSourceOption()
	for _, fn := range fns {
		fn(&opts)
	}
	return &DataSourceMultiAdaptor[K, V]{
		solutionName: opts.SolutionName,
		preAdaptor:   preAdaptor,
		dataSourceFn: dsfn,
	}
}

// Name 适配器名称
func (c *DataSourceMultiAdaptor[K, V]) Name() string {
	return "datasource_database"
}

// Get 读取对象
func (c *DataSourceMultiAdaptor[K, V]) Get(ctx context.Context, keys adaptor.Keys[K], vals adaptor.Values[K, V], fn adaptor.NewValueFunc[V]) (adaptor.Keys[K], error) {
	startTime := time.Now()
	hasKeys := make(adaptor.Keys[K], 0)
	hasValues := make(adaptor.ValueCol[V], 0)

	results, err := c.dataSourceFn(keys)
	if err != nil {
		return hasKeys, err
	}

	for key, val := range results {
		vals[key] = val
		hasKeys = append(hasKeys, key)
		hasValues = append(hasValues, val)

		metrics.AddMeta(ctx, metrics.Meta{
			AdaptorName: c.Name(),
			Key:         fmt.Sprint(val.Key()),
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
func (c *DataSourceMultiAdaptor[K, V]) Set(ctx context.Context, vals adaptor.ValueCol[V]) error {
	return nil
}

// Del 删除对象
func (c *DataSourceMultiAdaptor[K, V]) Del(ctx context.Context, keys adaptor.Keys[K]) error {
	return nil
}
