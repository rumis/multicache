# 简介
multicache设计为通用缓存框架，其核心功能：

* 支持多级缓存任意组合：本地缓存，分布式缓存，数据库，消息队列等
* 支持自定义编解码，当前项目中实现了原生json，msgpack编码方式
* 本地缓存默认基于FreeCache实现，支持各种自定义实现，支持本地缓存全局更新
* 分布式缓存默认基于go-redis/v9实现
* 支持多种方案解决缓存穿透&缓存击穿&缓存雪崩问题
* 指标采集，支持缓存命中率、响应时间、QPS等各种性能指标采集；默认实现了基于日志打印及Prometheus的适配器
* 支持单Key/批量Keys数据查询
* 【开发中】热Key发现
* 【开发中】Key管理API&UI

# 安装
使用最新版的multicache，可以在项目中导入该库。项目中使用了泛型特性，需要go版本在1.18以上
```
go get github.com/rumis/multicache
```

# 快速开始

### 定义数据对象
数据对象均需实现Metadata元数据接口，该接口定义了数据编解码以及零值的判定等方法
```
// Metadata 原数据定义
type Metadata interface {
	// Key 该对象的Key
	Key() string
	// Value 对象序列化后的值
	Value() ([]byte, error)
	// Decode 对象反序列化
	Decode([]byte) error
	// Zero 判定对象是否为零值
	Zero() bool
}
```

定义示例数据对象
```

import "github.com/rumis/multicache/encoding/json"

// Student 测试用示例对象
type Student struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
	Time int64  `json:"time"`
}
// Key 返回该对象的Key值
func (s *Student) Key() string {
	return s.Name
}
// Value 返回编码后字符串
func (s *Student) Value() ([]byte, error) {
	buf, err := json.Encode(s)
	return buf, err
}
// Decode 解码，字符串初始化该对象
func (s *Student) Decode(buf []byte) error {
	err := json.Decode(buf, s)
	return err
}
// Zero 判定对象是否为零值
func (s *Student) Zero() bool {
	return s.Name == ""
}
```

### 数据读取
#### 分布式缓存<->数据库
```
package multicache

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/rumis/multicache/adaptor"
	"github.com/rumis/multicache/datasource"
	"github.com/rumis/multicache/local"
	"github.com/rumis/multicache/metrics"
	"github.com/rumis/multicache/metrics/prometheus"
	"github.com/rumis/multicache/remote"
	"github.com/rumis/multicache/tests"
)

func TestCacheTwoLevel(t *testing.T) {

	testRemote := RemoteCacheTest(nil)
	testDataSource := DataSourceAdaptorTest(testRemote)

	cacheInst := NewCache[string, *tests.Student]("cache_test", testRemote, testDataSource)

	// 首次获取数据
	var s tests.Student
	ok, err := cacheInst.Get(context.Background(), "张三", &s)
	if err != nil {
		panic(err)
	}
	if !ok {
		t.Error("Get Error")
	}
	fmt.Println(s)

	// 再次获取同一数据
	var s1 tests.Student
	ok1, err1 := cacheInst.Get(context.Background(), "张三", &s1)
	if err1 != nil {
		panic(err1)
	}
	if !ok1 {
		t.Error("Get Error 1")
	}
	fmt.Println(s1)

	// 获取其他数据
	var s2 tests.Student
	ok2, err2 := cacheInst.Get(context.Background(), "李四", &s2)
	if err2 != nil {
		panic(err2)
	}
	if !ok2 {
		t.Error("Get Error 2")
	}
	fmt.Println(s2)
}
// RemoteCacheTest 创建分布式缓存适配器
func RemoteCacheTest(preAdaptor adaptor.Adaptor[string, *tests.Student]) adaptor.Adaptor[string, *tests.Student] {
	return remote.NewRedisAdaptor[string, *tests.Student](tests.NewRedisClient(), preAdaptor)
}
// DataSourceAdaptorTest 创建数据源适配器
func DataSourceAdaptorTest(preAdaptor adaptor.Adaptor[string, *tests.Student]) adaptor.Adaptor[string, *tests.Student] {
	return datasource.NewDataSourceAdaptor[string, *tests.Student](preAdaptor, func(key string) (*tests.Student, bool, error) {
		time.Sleep(100 * time.Millisecond)
		return &tests.Student{
			Name: key,
			Age:  18,
			Time: time.Now().UnixNano(),
		}, true, nil
	}, datasource.WithSingleFlightWaitTime(0))
}
```

通过系统默认的日志指标监控，我们可以看到如下输出内容
```
=== RUN   TestCacheTwoLevel
2024/06/23 07:17:41 INFO multicache_metrics name=cache_test trace=6058e50cad674ef19a920742b69a9ffd key=张三 remote_redis=Miss datasource_database=Hit remote_redis=Set
{张三 18 1719127061905550575}
2024/06/23 07:17:41 INFO multicache_metrics name=cache_test trace=7e38475fa59f4b7399223b21e7667bf0 key=张三 remote_redis=Hit
{张三 18 1719127061905550575}
2024/06/23 07:17:42 INFO multicache_metrics name=cache_test trace=2ac92bb6a69942fba35c1285b45549cc key=李四 remote_redis=Miss datasource_database=Hit remote_redis=Set
{李四 18 1719127062009207887}
--- PASS: TestCacheTwoLevel (0.21s)
```

#### 本地缓存<->分布式缓存<数据库>
```
func TestCacheThreeLevel(t *testing.T) {

	testLocal := LocalCacheTest()
	testRemote := RemoteCacheTest(testLocal)
	testDataSource := DataSourceAdaptorTest(testRemote)

	cacheInst := NewCache[string, *tests.Student]("cache_test", testLocal, testRemote, testDataSource)

	// 首次获取数据
	var s tests.Student
	ok, err := cacheInst.Get(context.Background(), "张三", &s)
	if err != nil {
		panic(err)
	}
	if !ok {
		t.Error("Get Error")
	}
	fmt.Println(s)

	// 再次获取同一数据
	var s1 tests.Student
	ok1, err1 := cacheInst.Get(context.Background(), "张三", &s1)
	if err1 != nil {
		panic(err1)
	}
	if !ok1 {
		t.Error("Get Error 1")
	}
	fmt.Println(s1)

	// 第三次获取同一数据
	var s3 tests.Student
	ok3, err3 := cacheInst.Get(context.Background(), "张三", &s3)
	if err3 != nil {
		panic(err3)
	}
	if !ok3 {
		t.Error("Get Error 3")
	}
	fmt.Println(s3)

	// 获取其他数据
	var s2 tests.Student
	ok2, err2 := cacheInst.Get(context.Background(), "李四", &s2)
	if err2 != nil {
		panic(err2)
	}
	if !ok2 {
		t.Error("Get Error 2")
	}
	fmt.Println(s2)
}

// LocalCacheTest 本地缓存
func LocalCacheTest() adaptor.Adaptor[string, *tests.Student] {
	return local.NewFreeCache[string, *tests.Student](local.FreeCacheClient(), nil)
}

// RemoteCacheTest 分布式缓存
func RemoteCacheTest(preAdaptor adaptor.Adaptor[string, *tests.Student]) adaptor.Adaptor[string, *tests.Student] {
	return remote.NewRedisAdaptor[string, *tests.Student](tests.NewRedisClient(), preAdaptor)
}

// DataSourceAdaptorTest 数据源适配器
func DataSourceAdaptorTest(preAdaptor adaptor.Adaptor[string, *tests.Student]) adaptor.Adaptor[string, *tests.Student] {
	return datasource.NewDataSourceAdaptor[string, *tests.Student](preAdaptor, func(key string) (*tests.Student, bool, error) {
		time.Sleep(100 * time.Millisecond)
		return &tests.Student{
			Name: key,
			Age:  18,
			Time: time.Now().UnixNano(),
		}, true, nil
	}, datasource.WithSingleFlightWaitTime(0))
}
```

#### 开启数据源singleflight支持
singleflight默认开启，等待数据源超时时间为200ms，可以通过数据源选项参数SingleFlightWaitTime进行修改，如果值为零则表示不启用singleflight支持
```
package multicache

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/rumis/multicache/adaptor"
	"github.com/rumis/multicache/datasource"
	"github.com/rumis/multicache/local"
	"github.com/rumis/multicache/metrics"
	"github.com/rumis/multicache/metrics/prometheus"
	"github.com/rumis/multicache/remote"
	"github.com/rumis/multicache/tests"
)

func TestCacheSingleflight(t *testing.T) {

	testLocal := LocalCacheTest()
	testRemote := RemoteCacheTest(testLocal)
	testDataSource := DataSourceAdaptorTest(testRemote)

	cacheInst := NewCache[string, *tests.Student]("cache_test", testLocal, testRemote, testDataSource)

	var wg sync.WaitGroup

	for range 5 {
		wg.Add(1)
		go func() {

			defer func() {
				if err := recover(); err != nil {
					fmt.Println(err)
				}
			}()

			defer wg.Done()
			var s tests.Student
			ok, err := cacheInst.Get(context.Background(), "张三", &s)
			if err != nil {
				panic(err)
			}
			if !ok {
				t.Error("Get Error")
			}
			// fmt.Println(s)
		}()
	}
	wg.Wait()
}

// LocalCacheTest 本地缓存
func LocalCacheTest() adaptor.Adaptor[string, *tests.Student] {
	return local.NewFreeCache[string, *tests.Student](local.FreeCacheClient(), nil)
}

// RemoteCacheTest 分布式缓存
func RemoteCacheTest(preAdaptor adaptor.Adaptor[string, *tests.Student]) adaptor.Adaptor[string, *tests.Student] {
	return remote.NewRedisAdaptor[string, *tests.Student](tests.NewRedisClient(), preAdaptor)
}

// DataSourceAdaptorTest 数据源适配器
// 创建数据源适配器时我们通过设置singleflight等待时长开启singleflight支持
func DataSourceAdaptorTest(preAdaptor adaptor.Adaptor[string, *tests.Student]) adaptor.Adaptor[string, *tests.Student] {
	return datasource.NewDataSourceAdaptor[string, *tests.Student](preAdaptor, func(key string) (*tests.Student, bool, error) {
		time.Sleep(100 * time.Millisecond)
		fmt.Println("DataSourceAdaptorTest,if this print is means the db query may happend")
		return &tests.Student{
			Name: key,
			Age:  18,
			Time: time.Now().UnixNano(),
		}, true, nil
	}, datasource.WithSingleFlightWaitTime(200*time.Millisecond))
}
```
上述单元测试通过默认的日志指标监控输出如下内容
```
=== RUN   TestCacheSingleflight
DataSourceAdaptorTest,if this print is means the db query may happend
2024/06/23 07:59:19 INFO multicache_metrics name=cache_test trace=2a5b0fd144af457698088f4fa01395a8 key=张三 local_freecache=Miss remote_redis=Miss datasource_database=Hit remote_redis=Set
2024/06/23 07:59:19 INFO multicache_metrics name=cache_test trace=2ce366749f6f48ec9268aa7e6c72f187 key=张三 local_freecache=Miss remote_redis=Miss datasource_database=Hit remote_redis=Set
2024/06/23 07:59:19 INFO multicache_metrics name=cache_test trace=e017f268147f46588bac3b299ba28b8a key=张三 local_freecache=Miss remote_redis=Miss datasource_database=Hit remote_redis=Set
2024/06/23 07:59:19 INFO multicache_metrics name=cache_test trace=106b4b8edccd45399cc8cf08f08c772d key=张三 local_freecache=Miss remote_redis=Miss datasource_database=Hit remote_redis=Set
2024/06/23 07:59:19 INFO multicache_metrics name=cache_test trace=fad2d5502e3a4c30a850f70c84381081 key=张三 local_freecache=Miss remote_redis=Miss datasource_database=Hit remote_redis=Set
--- PASS: TestCacheSingleflight (0.12s)
```
通过日志我们可以看到，虽然触发了5次【datasource_database=Hit】，但实际上数据源适配器模拟函数仅被执行了一次

#### 批量数据读取
```
package multicache

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/rumis/multicache/adaptor"
	"github.com/rumis/multicache/datasource"
	"github.com/rumis/multicache/local"
	"github.com/rumis/multicache/remote"
	"github.com/rumis/multicache/tests"
)

func TestMultiCache(t *testing.T) {

	testLocalMulti := MultiLocalCacheTest()
	testRemoteMulti := MultiRemoteCacheTest(testLocalMulti)
	testDataSourceMulti := MultiMysqlAdaptorTest(testRemoteMulti)

	multiCacheInst := NewMultiCache[string, *tests.Student]("multicache_test", testLocalMulti, testRemoteMulti, testDataSourceMulti)

	// 首次获取数据
	s := make(map[string]*tests.Student)
	err := multiCacheInst.Get(context.Background(), []string{"张三", "李四"}, s, func() *tests.Student {
		return &tests.Student{}
	})
	if err != nil {
		panic(err)
	}
	buf, _ := json.Marshal(s)
	fmt.Println(string(buf))

	// 再次数据
	s1 := make(map[string]*tests.Student)
	err = multiCacheInst.Get(context.Background(), []string{"王五", "李四"}, s1, func() *tests.Student {
		return &tests.Student{}
	})
	if err != nil {
		panic(err)
	}
	buf1, _ := json.Marshal(s1)
	fmt.Println(string(buf1))

	// 第三次数据
	s2 := make(map[string]*tests.Student)
	err = multiCacheInst.Get(context.Background(), []string{"王五", "李四"}, s2, func() *tests.Student {
		return &tests.Student{}
	})
	if err != nil {
		panic(err)
	}
	buf2, _ := json.Marshal(s2)
	fmt.Println(string(buf2))

}

func MultiLocalCacheTest() adaptor.MultiAdaptor[string, *tests.Student] {
	return local.NewMultiFreeCache[string, *tests.Student](local.FreeCacheClient(), nil)
}

func MultiRemoteCacheTest(preAdaptor adaptor.MultiAdaptor[string, *tests.Student]) adaptor.MultiAdaptor[string, *tests.Student] {
	return remote.NewRedisMultiAdaptor[string, *tests.Student](tests.NewRedisClient(), preAdaptor)
}

func MultiMysqlAdaptorTest(preAdaptor adaptor.MultiAdaptor[string, *tests.Student]) adaptor.MultiAdaptor[string, *tests.Student] {
	return datasource.NewDataSourceMultiAdaptor[string, *tests.Student](preAdaptor, func(keys adaptor.Keys[string]) (adaptor.Values[string, *tests.Student], error) {
		vals := make(adaptor.Values[string, *tests.Student], 0)
		for _, key := range keys {
			vals[key] = &tests.Student{
				Name: key,
				Age:  18,
			}
		}
		return vals, nil
	})
}
```

# 自定义日志
系统日志模块支持自定义，只需实现如下接口即可
```
// Logger 日志接口定义
type Logger interface {
	Debug(format string, v ...any)
	Info(format string, v ...any)
	Warn(format string, v ...any)
	Error(format string, v ...any)
}
```
默认日志通过官方slog实现，可以通过SetLogger方法进行替换

# 编解码
Metadata元数据接口定义了数据的编解码方式，用户可任意自定义实现。系统默认提供了原生json，msgpack两种。对编解码性能要求较高的场景可以选用字节的[sonic](https://github.com/bytedance/sonic)库，其使用JIT和SIMD加速

# 缓存指标收集和统计
系统支持缓存命中率，查询响应耗时，及QPS等核心性能指标收集。同时系统还支持自定义指标数据的输出方式，通过Metrics接口实现
```
// Metrics 统计接口
type Metrics interface {
	Start(ctx context.Context, name string) error
	AddMeta(ctx context.Context, meta Meta) error
	Summary(ctx context.Context) error
}
```

#### 默认日志方式
系统默认以日志的方式输出监控指标，该方式只收集了行为日志，未进一步整合，建议只测试环境使用
```
2024/06/24 08:43:39 INFO multicache_metrics name=cache_test trace=3105f3de9705484db2c9ae24842fe48c key=张三 remote_redis=Miss remote_redis_track_time=0 datasource_database=Hit datasource_database_track_time=102 remote_redis=Set remote_redis_track_time=0
2024/06/24 08:43:39 INFO multicache_metrics name=cache_test trace=6c5c7615765d4d5f9dd4aaddf2fed6e5 key=张三 remote_redis=Hit remote_redis_track_time=0
2024/06/24 08:43:39 INFO multicache_metrics name=cache_test trace=496ab376cf374613b1d3d10f993ce01a key=李四 remote_redis=Miss remote_redis_track_time=0 datasource_database=Hit datasource_database_track_time=101 remote_redis=Set remote_redis_track_time=0
```

#### 启用Phemotheus数据指标监控
```
prometheusGateWayHost := "http://127.0.0.1:9091"
metricPrometheus := metrics.NewMetricsPrometheus("multicache", "cache_hitmiss", "multicache_test_solution", prometheus.WithGatewayHost(prometheusGateWayHost))
metrics.SetMetrics(metricPrometheus)
```

# 基准测试





