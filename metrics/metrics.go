package metrics

import "context"

// Metrics 统计接口
type Metrics interface {
	Start(ctx context.Context, name string) error
	AddMeta(ctx context.Context, meta Meta) error
	Summary(ctx context.Context) error
}

var defaultMetrics Metrics = NewMetricsLogger()

// DefaultMetrics 获取当前系统默认的统计器
func DefaultMetrics() Metrics {
	return defaultMetrics
}

// SetMetrics 设置默认的统计器
func SetMetrics(m Metrics) {
	defaultMetrics = m
}
