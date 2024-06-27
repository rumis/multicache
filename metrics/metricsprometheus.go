package metrics

import (
	"context"

	"github.com/rumis/multicache/metrics/prometheus"
)

// MetricsPrometheus 通过Prometheus采集缓存相关统计信息
type MetricsPrometheus struct {
	promCounter   *prometheus.Counter
	promHistogram *prometheus.Histogram
}

// NewMetricsPrometheus 创建一个新的MetricsPrometheus
func NewMetricsPrometheus(namespace string, name string, job string, fns ...prometheus.PrometheusClientOptionsHandle) *MetricsPrometheus {
	return &MetricsPrometheus{
		promCounter:   prometheus.NewCounter(namespace, name, job, fns...),
		promHistogram: prometheus.NewHistogram(namespace, name, job, []float64{10, 20, 50, 100, 200, 500, 1000}, fns...),
	}
}

// Start 开始统计
func (m *MetricsPrometheus) Start(ctx context.Context, name string) error {
	return nil
}

// AddMeta 添加单次查询结果
func (m *MetricsPrometheus) AddMeta(ctx context.Context, meta Meta) error {
	labs := []prometheus.Label{
		{
			Name:  "adaptor",
			Value: meta.AdaptorName,
		},
		{
			Name:  "key",
			Value: meta.Key,
		},
		{
			Name:  "event",
			Value: MetaEventString(meta.Type),
		},
	}
	// 次数
	err := m.promCounter.Incr(labs...)
	if err != nil {
		return err
	}

	// 耗时
	if meta.TrackTime > 0 {
		err = m.promHistogram.Observe(float64(meta.TrackTime), labs...)
		if err != nil {
			return err
		}
	}

	return nil
}

// Summary 输出统计信息
func (m *MetricsPrometheus) Summary(ctx context.Context) error {
	return nil
}
