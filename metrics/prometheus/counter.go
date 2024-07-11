package prometheus

import (
	"errors"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

// Counter 计数器指标
type Counter struct {
	opts PrometheusClientOptions
	cnt  prometheus.Counter
	ch   chan []Label
	job  string
	run  bool
}

// NewCounter 创建计数器
// 每创建一个计数器附带创建一个协程消费数据，当不再使用时需要调用Stop方法销毁协程
func NewCounter(namespace string, name string, job string, fns ...PrometheusClientOptionsHandle) *Counter {
	opts := DefaultPrometheusClientOptions()
	for _, fn := range fns {
		fn(&opts)
	}
	counter := &Counter{
		opts: opts,
		cnt: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      name,
			Help:      namespace + "." + name,
		}),
		ch:  make(chan []Label, opts.CacheSize),
		job: job,
		run: true,
	}
	go counter.pusher()
	return counter
}

// Incr 计数器递增
func (c *Counter) Incr(labels ...Label) error {
	select {
	case c.ch <- labels:
	default:
		return Error_ChannelFull
	}
	return nil
}

// Stop 销毁数据消费协程
func (c *Counter) Stop() {
	c.run = false
}

// pusher 指标数据推送协程
func (c *Counter) pusher() {
	defer func() {
		if err := recover(); err != nil {
			if c.opts.PanicErrorHandle != nil {
				c.opts.PanicErrorHandle(errors.New(fmt.Sprint(err)))
			}
			go c.pusher() // 重新创建指标数据推送协程
		}
	}()
	for {
		var labels []Label
		var ok bool
		select {
		case labels, ok = <-c.ch:
		default:
		}
		if !ok {
			time.Sleep(c.opts.PusherWaitingTimeout)
			if !c.run {
				return
			}
			continue
		}
		// 计数器递增
		c.cnt.Inc()
		// 创建pusher
		pusher := push.New(c.opts.GateWayHost, c.job).Collector(c.cnt)
		// 附加标签
		for _, v := range labels {
			pusher.Grouping(v.Name, v.Value)
		}
		// 推送
		err := pusher.Push()
		if err != nil && c.opts.PushErrorHandle != nil {
			c.opts.PushErrorHandle(err)
		}
	}
}
