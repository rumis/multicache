package metrics

type MetaEvent int

const (
	Hit MetaEvent = iota + 1 // Hit
	Miss
	Set
)

// QueryResultTypeString 返回查询结果类型的字符串表示
func MetaEventString(t MetaEvent) string {
	switch t {
	case Hit:
		return "Hit"
	case Miss:
		return "Miss"
	case Set:
		return "Set"
	default:
		return "Unknown"
	}
}

type ContextKey string

const MetricsTraceKey = ContextKey("multicache_metrics_trace")
const MetricsClient = ContextKey("multicache_metrics_client")

// Meta 适配器单次查询结果
type Meta struct {
	Key         string
	AdaptorName string
	Type        MetaEvent
	TrackTime   int64
}

// MetricsMeta 单流程查询结果
type MetricsMeta struct {
	Name  string
	Trace string
	Metas map[string][]Meta
}
