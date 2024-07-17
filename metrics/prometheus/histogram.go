package prometheus

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

// HistogramMessage 直方图消息
type HistogramMessage struct {
	Value  float64
	Labels []Label
}

// Histogram 直方图指标
type Histogram struct {
	opts PrometheusClientOptions
	hist prometheus.Histogram
	ch   chan HistogramMessage
	job  string
	run  bool
}

// NewHistogram 创建直方图指标
// 每创建一个直方图指标附带创建一个协程消费数据，当不再使用时需要调用Stop方法销毁协程
// 建议创建的指标推送器采用单例，避免协程泄露
func NewHistogram(namespace string, name string, job string, bucket []float64, fns ...PrometheusClientOptionsHandle) *Histogram {
	opts := DefaultPrometheusClientOptions()
	for _, fn := range fns {
		fn(&opts)
	}

	h := &Histogram{
		opts: opts,
		hist: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      name,
			Help:      namespace + "." + name,
			Buckets:   bucket,
		}),
		ch:  make(chan HistogramMessage, opts.CacheSize),
		job: job,
		run: true,
	}
	go h.pusher()
	return h
}

// Observe 记录值
func (h *Histogram) Observe(val float64, labels ...Label) error {
	select {
	case h.ch <- HistogramMessage{
		Value:  val,
		Labels: labels,
	}:
	default:
		return Error_ChannelFull
	}
	return nil
}

// Stop 销毁数据消费协程
func (h *Histogram) Stop() {
	h.run = false
}

// pusher 指标数据推送协程
func (h *Histogram) pusher() {
	defer func() {
		if err := recover(); err != nil {
			if h.opts.PanicErrorHandle != nil {
				h.opts.PanicErrorHandle(errors.New(fmt.Sprint(err)))
			}
			go h.pusher() // 重新创建指标数据推送协程
		}
	}()
	for {
		var histMsg HistogramMessage
		var ok bool
		select {
		case histMsg, ok = <-h.ch:
		default:
		}
		if !ok {
			time.Sleep(h.opts.PusherWaitingTimeout * time.Millisecond)
			if !h.run { // 停止消费协程
				return
			}
			continue
		}
		// 设置值
		h.hist.Observe(histMsg.Value)
		// 创建pusher
		pusher := push.New(h.opts.GateWayHost, h.job).Collector(h.hist)
		// 附加标签
		for _, v := range histMsg.Labels {
			pusher.Grouping(v.Name, v.Value)
		}
		// 推送
		ctx, cfn := context.WithTimeout(context.Background(), time.Second*3)
		err := pusher.PushContext(ctx)
		if err != nil && h.opts.PushErrorHandle != nil {
			h.opts.PushErrorHandle(err)
		}
		cfn() // 调用取消函数，防止内存泄露
	}
}
