package local

import (
	"time"

	"github.com/rumis/multicache/syncer"
)

// LocalCacheOption 本地缓存选项
type LocalCacheOption struct {
	Prefix       string
	TTL          time.Duration
	Threshold    time.Duration
	SkipGet      bool
	Name         string
	SolutionName string
	TTLZero      time.Duration
	Syncer       syncer.Syncer
}

// LocalCacheOptionFunc 本地缓存配置函数
type LocalCacheOptionFunc func(*LocalCacheOption)

// DefaultLocalCacheOption 默认本地缓存配置
func DefaultLocalCacheOption() LocalCacheOption {
	return LocalCacheOption{
		Name:      "local_freecache",
		Prefix:    "multicache_local_",
		TTL:       time.Second * 30,
		Threshold: time.Second * 5,
		TTLZero:   time.Second * 5,
	}
}

// WithPrefix 设置缓存前缀
func WithPrefix(prefix string) LocalCacheOptionFunc {
	return func(option *LocalCacheOption) {
		option.Prefix = prefix
	}
}

// WithTTL 设置缓存过期时间
func WithTTL(ttl time.Duration) LocalCacheOptionFunc {
	return func(option *LocalCacheOption) {
		option.TTL = ttl
	}
}

// WithThreshold 设置缓存过期噪音阈值
func WithThreshold(threshold time.Duration) LocalCacheOptionFunc {
	return func(option *LocalCacheOption) {
		option.Threshold = threshold
	}
}

// WithSkipGet 设置是否跳过Get操作
func WithSkipGet(skip bool) LocalCacheOptionFunc {
	return func(option *LocalCacheOption) {
		option.SkipGet = skip
	}
}

// WithSolutionName 场景名称
func WithSolutionName(name string) LocalCacheOptionFunc {
	return func(option *LocalCacheOption) {
		option.SolutionName = name
	}
}

// WithName 适配器名称
func WithName(name string) LocalCacheOptionFunc {
	return func(option *LocalCacheOption) {
		option.Name = name
	}
}

// WithTTLZero 零值对象缓存过期时间
func WithTTLZero(ttl time.Duration) LocalCacheOptionFunc {
	return func(option *LocalCacheOption) {
		option.TTLZero = ttl
	}
}

// WithSyncer 配置数据同步器
func WithSyncer(syncer syncer.Syncer) LocalCacheOptionFunc {
	return func(option *LocalCacheOption) {
		option.Syncer = syncer
	}
}
