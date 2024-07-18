package prometheus

import (
	"errors"
	"time"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/rumis/multicache/logger"
)

var (
	Error_ChannelFull = errors.New("channel is full,this event will be drop")
	Error_ClientNil   = errors.New("client initialize failed")
)

const (
	LogEventPrometheusPushErr  = "PROMETHEUS_PUSH_ERR"
	LogEventPrometheusQueryErr = "PROMETHEUS_QUERY_ERR"
	LogEventPrometheusPanicErr = "PROMETHEUS_PANIC_ERR"
)

type QueryValue model.Value
type QueryWarnings v1.Warnings
type QueryRange v1.Range

// Label 指标标签
type Label struct {
	Name  string
	Value string
}

type PrometheusClientOptionsHandle func(opts *PrometheusClientOptions)

type PrometheusClientErrorHandle func(err error)

type PrometheusClientOptions struct {
	GateWayHost          string
	PushErrorHandle      PrometheusClientErrorHandle
	QueryErrorHandle     PrometheusClientErrorHandle
	PanicErrorHandle     PrometheusClientErrorHandle
	CacheSize            int
	PromReadHost         string
	PromHttpApiQueryHost string
	QueryTimeout         time.Duration // api查询时http请求超时时间
	// 当Channel中元素空时，pusher协程等待时间
	PusherWaitingTimeout time.Duration
}

// DefaultPrometheusClientOptions 默认配置
func DefaultPrometheusClientOptions() PrometheusClientOptions {
	return PrometheusClientOptions{
		GateWayHost: "http://127.0.0.1:9091",
		CacheSize:   1024,
		PushErrorHandle: func(err error) {
			logger.Error(err.Error(), "event", LogEventPrometheusPushErr)
		},
		QueryErrorHandle: func(err error) {
			logger.Error(err.Error(), "event", LogEventPrometheusQueryErr)
		},
		PanicErrorHandle: func(err error) {
			logger.Error(err.Error(), "event", LogEventPrometheusPanicErr)
		},
		PusherWaitingTimeout: 3 * time.Millisecond,
		QueryTimeout:         3 * time.Second,
	}
}

// WithGatewayHost 配置pushgateway的域名
func WithGatewayHost(h string) PrometheusClientOptionsHandle {
	return func(opts *PrometheusClientOptions) {
		opts.GateWayHost = h
	}
}

// WithPushErrorHandle 配置推送错误时的日志方法
func WithPushErrorHandle(fn PrometheusClientErrorHandle) PrometheusClientOptionsHandle {
	return func(opts *PrometheusClientOptions) {
		opts.PushErrorHandle = fn
	}
}

// WithQueryErrorHandle 配置推送错误时的日志方法
func WithQueryErrorHandle(fn PrometheusClientErrorHandle) PrometheusClientOptionsHandle {
	return func(opts *PrometheusClientOptions) {
		opts.QueryErrorHandle = fn
	}
}

// WithPanicErrorHandle 配置panic错误输出方法
func WithPanicErrorHandle(fn PrometheusClientErrorHandle) PrometheusClientOptionsHandle {
	return func(opts *PrometheusClientOptions) {
		opts.PanicErrorHandle = fn
	}
}

// WithCacheSize 配置缓存大小
func WithCacheSize(size int) PrometheusClientOptionsHandle {
	return func(opts *PrometheusClientOptions) {
		opts.CacheSize = size
	}
}

// WithPusherWaitingTimeout 配置消息推送线程空数据等待时间
func WithPusherWaitingTimeout(ts time.Duration) PrometheusClientOptionsHandle {
	return func(opts *PrometheusClientOptions) {
		opts.PusherWaitingTimeout = ts
	}
}

// WithPromReadHost 设置prometheus数据读取host
func WithPromReadHost(h string) PrometheusClientOptionsHandle {
	return func(opts *PrometheusClientOptions) {
		opts.PromReadHost = h
	}
}

// WithPromHttpApiQueryHost 通过http-api查询阿里prometheus数据配置HOTS
func WithPromHttpApiQueryHost(h string) PrometheusClientOptionsHandle {
	return func(opts *PrometheusClientOptions) {
		opts.PromHttpApiQueryHost = h
	}
}

// WithQueryTimeout 设置查询超时时间
func WithQueryTimeout(ts time.Duration) PrometheusClientOptionsHandle {
	return func(opts *PrometheusClientOptions) {
		opts.QueryTimeout = ts
	}
}
