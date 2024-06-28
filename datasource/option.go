package datasource

import (
	"time"

	"github.com/rumis/multicache/adaptor"
)

// DataSourceOption 数据源配置选项
type DataSourceOption struct {
	Name                 string
	SolutionName         string
	SingleFlightWaitTime time.Duration // 单飞请求等待时间
}

// DataSourceOptionFunc 数据源配置函数
type DataSourceOptionFunc func(*DataSourceOption)

// DefaultDataSourceOption 默认数据源配置
func DefaultDataSourceOption() DataSourceOption {
	return DataSourceOption{
		Name:                 "datasource_database",
		SolutionName:         "multicache_default",
		SingleFlightWaitTime: 200 * time.Millisecond,
	}
}

// WithSolutionName 场景名称
func WithSolutionName(name string) DataSourceOptionFunc {
	return func(option *DataSourceOption) {
		option.SolutionName = name
	}
}

// WithSingleFlightWaitTime 单飞请求等待时间, 超过等待时长则直接调用底层数据方法(可能会触发查库)
func WithSingleFlightWaitTime(waitTime time.Duration) DataSourceOptionFunc {
	return func(option *DataSourceOption) {
		option.SingleFlightWaitTime = waitTime
	}
}

// WithName 设置适配器名称
func WithName(name string) DataSourceOptionFunc {
	return func(option *DataSourceOption) {
		option.Name = name
	}
}

// ValueWithError 含返回值和错误的结构
type ValueWithError[V adaptor.Metadata] struct {
	Val V
	Err error
}
