package metrics

import (
	"context"
	"errors"
	"sync"

	"github.com/rumis/multicache/logger"
)

var _ Metrics = &MetricsLogger{}

// MetricsLogger 以日志的方式输出缓存相关统计信息
type MetricsLogger struct {
	metas sync.Map
}

// NewMetricsLogger 创建一个新的MetricsLogger
func NewMetricsLogger() *MetricsLogger {
	return &MetricsLogger{}
}

// Start 开始统计
func (m *MetricsLogger) Start(ctx context.Context, name string) error {
	traceKey := ctx.Value(MetricsTraceKey).(string)
	if traceKey == "" {
		return errors.New("metrics trace key not found in context")
	}
	m.metas.Store(traceKey, MetricsMeta{
		Name:  name,
		Trace: traceKey,
		Metas: make(map[string][]Meta),
	})
	return nil
}

// AddMeta 添加单次查询结果
func (m *MetricsLogger) AddMeta(ctx context.Context, meta Meta) error {
	traceKey := ctx.Value(MetricsTraceKey).(string)
	if traceKey == "" {
		return errors.New("metrics trace key not found in context")
	}
	item, ok := m.metas.Load(traceKey)
	if !ok {
		return errors.New("metrics item not found")
	}
	metricsItem, ok := item.(MetricsMeta)
	if !ok {
		return errors.New("invalid metrics item")
	}

	metas := metricsItem.Metas[meta.Key]
	if metas == nil {
		metas = make([]Meta, 0)
	}
	metas = append(metas, meta)
	metricsItem.Metas[meta.Key] = metas

	m.metas.Store(traceKey, metricsItem)

	return nil
}

// Summary 输出统计信息
// 多key时输出多条信息
func (m *MetricsLogger) Summary(ctx context.Context) error {
	traceKey := ctx.Value(MetricsTraceKey).(string)
	if traceKey == "" {
		return errors.New("metrics trace key not found in context")
	}
	item, ok := m.metas.Load(traceKey)
	if !ok {
		return errors.New("metrics item not found")
	}
	metricsItem, ok := item.(MetricsMeta)
	if !ok {
		return errors.New("invalid metrics item")
	}

	statItems := []any{"name", metricsItem.Name, "trace", metricsItem.Trace}

	for key, metas := range metricsItem.Metas {
		keyStatItems := append(statItems, "key", key)
		for _, meta := range metas {
			keyStatItems = append(keyStatItems, meta.AdaptorName, MetaEventString(meta.Type), meta.AdaptorName+"_track_time", meta.TrackTime)
		}

		// 每个key输出一条记录
		logger.Info("multicache_metrics", keyStatItems...)
	}

	// 删除
	m.metas.Delete(traceKey)

	return nil

}
