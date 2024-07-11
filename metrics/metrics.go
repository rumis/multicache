package metrics

import "context"

// Metrics 统计接口
type Metrics interface {
	Start(ctx context.Context, name string) error
	AddMeta(ctx context.Context, meta Meta) error
	Summary(ctx context.Context) error
}

var defaultMetrics Metrics = NewMetricsLogger()

// SetMetrics 设置默认的统计器
func SetMetrics(m Metrics) {
	defaultMetrics = m
}

// Start 开始
func Start(ctx context.Context, name string) error {
	return defaultMetrics.Start(ctx, name)
}

// AddMeta 添加单次查询结果
func AddMeta(ctx context.Context, meta Meta) error {
	return defaultMetrics.AddMeta(ctx, meta)
}

// Summary 输出统计信息
func Summary(ctx context.Context) {
	defaultMetrics.Summary(ctx)
}
