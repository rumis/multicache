package remote

import "time"

type RemoteCacheOption struct {
	Prefix       string
	TTL          time.Duration
	Threshold    int // 单位秒
	SkipGet      bool
	Name         string
	SolutionName string
	TTLZero      time.Duration
}

// RemoteCacheOptionFunc 分布式缓存配置函数
type RemoteCacheOptionFunc func(*RemoteCacheOption)

// DefaultRemoteCacheOption 默认分布式缓存配置
func DefaultRemoteCacheOption() RemoteCacheOption {
	return RemoteCacheOption{
		Name:      "remote_redis",
		Prefix:    "mulcache_local_",
		TTL:       time.Second * 90,
		Threshold: 5,
		TTLZero:   time.Second * 5,
	}
}

// WithPrefix 设置缓存前缀
func WithPrefix(prefix string) RemoteCacheOptionFunc {
	return func(option *RemoteCacheOption) {
		option.Prefix = prefix
	}
}

// WithTTL 设置缓存过期时间
func WithTTL(ttl time.Duration) RemoteCacheOptionFunc {
	return func(option *RemoteCacheOption) {
		option.TTL = ttl
	}
}

// WithSkipGet 设置是否跳过Get操作
func WithSkipGet(skip bool) RemoteCacheOptionFunc {
	return func(option *RemoteCacheOption) {
		option.SkipGet = skip
	}
}

// WithSolutionName 设置场景名称
func WithSolutionName(name string) RemoteCacheOptionFunc {
	return func(option *RemoteCacheOption) {
		option.SolutionName = name
	}
}

// WithName 设置适配器名称
func WithName(name string) RemoteCacheOptionFunc {
	return func(option *RemoteCacheOption) {
		option.Name = name
	}
}

// WithThreshold 设置缓存过期噪音阈值
func WithThreshold(threshold int) RemoteCacheOptionFunc {
	return func(option *RemoteCacheOption) {
		option.Threshold = threshold
	}
}

// WithTTLZero 零值对象缓存过期时间
func WithTTLZero(ttl time.Duration) RemoteCacheOptionFunc {
	return func(option *RemoteCacheOption) {
		option.TTLZero = ttl
	}
}
