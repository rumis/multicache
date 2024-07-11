package prometheus

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

// Query Prometheus数据查询
type Query struct {
	opts      PrometheusClientOptions
	client    api.Client
	clientErr error
}

// NewQuery Prometheus数据查询
func NewQuery(fns ...PrometheusClientOptionsHandle) *Query {
	opts := DefaultPrometheusClientOptions()
	for _, fn := range fns {
		fn(&opts)
	}
	client, err := api.NewClient(api.Config{
		Address: opts.PromHttpApiQueryHost,
	})
	if err != nil {
		return &Query{
			opts:      opts,
			clientErr: err,
		}
	}
	return &Query{
		opts:   opts,
		client: client,
	}
}

// Query 查询
func (q *Query) Query(query string, ts time.Time, opts ...v1.Option) (model.Value, error) {
	if q.client == nil && q.clientErr != nil {
		return nil, q.clientErr
	}
	if q.client == nil && q.clientErr == nil {
		return nil, Error_ClientNil
	}
	v1api := v1.NewAPI(q.client)
	ctx, cancel := context.WithTimeout(context.Background(), q.opts.QueryTimeout)
	defer cancel()

	m, _, err := v1api.Query(ctx, query, ts, opts...)
	return m, err
}

// Range 范围查询
func (q *Query) Range(query string, r v1.Range, opts ...v1.Option) (model.Value, error) {
	if q.client == nil && q.clientErr != nil {
		return nil, q.clientErr
	}
	if q.client == nil && q.clientErr == nil {
		return nil, Error_ClientNil
	}
	v1api := v1.NewAPI(q.client)
	ctx, cancel := context.WithTimeout(context.Background(), q.opts.QueryTimeout)
	defer cancel()

	m, _, err := v1api.QueryRange(ctx, query, r, opts...)

	return m, err
}

// Series 查询时间序列
// query支持正则，re2语法
func (q *Query) Series(startTime time.Time, endTime time.Time, query ...string) ([]model.LabelSet, error) {
	if q.client == nil && q.clientErr != nil {
		return nil, q.clientErr
	}
	if q.client == nil && q.clientErr == nil {
		return nil, Error_ClientNil
	}
	v1api := v1.NewAPI(q.client)
	ctx, cancel := context.WithTimeout(context.Background(), q.opts.QueryTimeout)
	defer cancel()

	set, _, err := v1api.Series(ctx, query, startTime, endTime)
	if err != nil {
		return nil, err
	}
	return set, nil
}

// LabelNames 查询某序列下所有标签名称
func (q *Query) LabelNames(startTime time.Time, endTime time.Time, query ...string) ([]string, error) {
	if q.client == nil && q.clientErr != nil {
		return nil, q.clientErr
	}
	if q.client == nil && q.clientErr == nil {
		return nil, Error_ClientNil
	}
	v1api := v1.NewAPI(q.client)
	ctx, cancel := context.WithTimeout(context.Background(), q.opts.QueryTimeout)
	defer cancel()

	labels, _, err := v1api.LabelNames(ctx, query, startTime, endTime)
	if err != nil {
		return nil, err
	}
	return labels, nil
}

// LabelValues 查询某标签的所有值
func (q *Query) LabelValues(label string, startTime time.Time, endTime time.Time, query ...string) (model.LabelValues, error) {
	if q.client == nil && q.clientErr != nil {
		return nil, q.clientErr
	}
	if q.client == nil && q.clientErr == nil {
		return nil, Error_ClientNil
	}
	v1api := v1.NewAPI(q.client)
	ctx, cancel := context.WithTimeout(context.Background(), q.opts.QueryTimeout)
	defer cancel()

	vals, _, err := v1api.LabelValues(ctx, label, query, startTime, endTime)
	if err != nil {
		return nil, err
	}
	return vals, nil
}
