package datasource

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rumis/multicache/adaptor"
	"github.com/rumis/multicache/logger"
	"github.com/rumis/multicache/metrics"
	"golang.org/x/sync/singleflight"
)

// ErrNotFound 值不存在
var ErrNotFound = errors.New("entry not found")

// DataSourceFunc 数据源构造函数
type DataSourceFunc[K comparable, V adaptor.Metadata] func(key K) (V, bool, error)

// 类型检测
var _ adaptor.Adaptor[string, adaptor.Metadata] = (*DataSourceAdaptor[string, adaptor.Metadata])(nil)

// DataSourceAdaptor 数据源
type DataSourceAdaptor[K comparable, V adaptor.Metadata] struct {
	name           string
	solutionName   string
	preAdaptor     adaptor.Adaptor[K, V]
	dataSourceFn   DataSourceFunc[K, V]
	sg             singleflight.Group
	sgWaitDuration time.Duration
}

// NewDataSourceAdaptor 创建一个新的数据源适配器对象
func NewDataSourceAdaptor[K comparable, V adaptor.Metadata](preAdaptor adaptor.Adaptor[K, V], dsfn DataSourceFunc[K, V], fns ...DataSourceOptionFunc) *DataSourceAdaptor[K, V] {
	opts := DefaultDataSourceOption()
	for _, fn := range fns {
		fn(&opts)
	}
	return &DataSourceAdaptor[K, V]{
		name:           opts.Name,
		solutionName:   opts.SolutionName,
		preAdaptor:     preAdaptor,
		dataSourceFn:   dsfn,
		sgWaitDuration: opts.SingleFlightWaitTime,
	}
}

// Name 适配器名称，需要在当前业务场景中保证唯一
func (c *DataSourceAdaptor[K, V]) Name() string {
	return c.name
}

// Get 读取对象
func (c *DataSourceAdaptor[K, V]) Get(ctx context.Context, key K, value V) (bool, error) {
	startTime := time.Now()
	done := make(chan ValueWithError[V])
	// 删除singleflight对象中的缓存
	defer func() {
		c.sg.Forget(fmt.Sprint(key))
	}()

	go func() {
		defer close(done) // 关闭通道
		defer func() {
			err := recover()
			if err != nil {
				done <- ValueWithError[V]{Val: value, Err: errors.New(fmt.Sprint(err))}
			}
		}()

		v, err, _ := c.sg.Do(fmt.Sprint(key), func() (interface{}, error) {
			resultVal, ok, err := c.dataSourceFn(key)
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, ErrNotFound
			}
			return resultVal, nil
		})
		done <- ValueWithError[V]{Val: v.(V), Err: err}
	}()

	var result ValueWithError[V]
	select {
	case result = <-done:
	case <-time.After(c.sgWaitDuration):
		r1, ok, err := c.dataSourceFn(key)
		if err == nil && !ok {
			err = ErrNotFound
		}
		result = ValueWithError[V]{Val: r1, Err: err}
	}

	if result.Err == ErrNotFound {
		return false, nil
	}
	if result.Err != nil {
		return false, result.Err
	}

	valBuf, err := result.Val.Value()
	if err != nil {
		return false, err
	}
	err = value.Decode(valBuf)
	if err != nil {
		return false, err
	}

	metrics.AddMeta(ctx, metrics.Meta{
		AdaptorName: c.Name(),
		Key:         fmt.Sprint(key),
		Type:        metrics.Hit,
		TrackTime:   time.Since(startTime).Milliseconds(),
	})

	if c.preAdaptor != nil {
		err := c.preAdaptor.Set(ctx, value)
		if err != nil {
			logger.Error(err.Error(), "solution", c.solutionName, "adaptor", c.Name(), "value", value, "event", adaptor.LogEventRefill)
		}
	}

	return true, nil
}

// Set 写入对象
func (c *DataSourceAdaptor[K, V]) Set(ctx context.Context, value V) error {
	return nil
}

// Del 删除对象 -
func (c *DataSourceAdaptor[K, V]) Del(ctx context.Context, key K) error {
	return nil
}
