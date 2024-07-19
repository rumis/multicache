package multicache

import (
	"context"
	"fmt"
	"strconv"
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
	"github.com/rumis/multicache/utils"
)

func TestCacheTwoLevel(t *testing.T) {

	testRemote := RemoteCacheTest(nil)
	testDataSource := DataSourceAdaptorTest(testRemote)

	cacheInst := NewCache[string, *tests.Student]("cache_test", testRemote, testDataSource)
	// cacheInst := NewCacheWithMetric[string, *tests.Student]("cache_test", metrics.NewMetricsLogger(), testRemote, testDataSource)

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

func TestCacheThreeLevel(t *testing.T) {

	// prometheusGateWayHost := tests.PromGatewayHost()
	// metricPrometheus := metrics.NewMetricsPrometheus("multicache", "cache_hitmiss", "multicache_test_solution", prometheus.WithGatewayHost(prometheusGateWayHost))
	// metrics.SetMetrics(metricPrometheus)

	testLocal := LocalCacheTest()
	testRemote := RemoteCacheTest(testLocal)
	testDataSource := DataSourceAdaptorTest(testRemote)

	// cacheInst := NewCache[string, *tests.Student]("cache_test", testLocal, testRemote, testDataSource)
	cacheInst := NewCacheWithMetric[string, *tests.Student]("cache_test", metrics.NewMetricsLogger(), testLocal, testRemote, testDataSource)

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

func TestCacheSingleflight(t *testing.T) {

	testLocal := LocalCacheTest()
	testRemote := RemoteCacheTest(testLocal)
	testDataSource := DataSourceAdaptorTest(testRemote)

	cacheInst := NewCache[string, *tests.Student]("cache_test", testLocal, testRemote, testDataSource)

	var wg sync.WaitGroup

	for i := 0; i < 5; i++ {
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

func TestCacheLoop(t *testing.T) {

	prometheusGateWayHost := tests.PromGatewayHost()
	metricPrometheus := metrics.NewMetricsPrometheus("multicache", "cache_hitmiss", "multicache_test_solution", prometheus.WithGatewayHost(prometheusGateWayHost))
	metrics.SetMetrics(metricPrometheus)

	testLocal := LocalCacheTest()
	testRemote := RemoteCacheTest(testLocal)
	testDataSource := DataSourceAdaptorTest(testRemote)

	cacheInst := NewCache[string, *tests.Student]("cache_test", testLocal, testRemote, testDataSource)

	for i := 0; i < 1000; i++ {
		idxName := utils.SafeRand().Intn(100)
		var s tests.Student
		ok, err := cacheInst.Get(context.Background(), strconv.Itoa(idxName), &s)
		if err != nil {
			panic(err)
		}
		if !ok {
			t.Error("Get Error")
		}
		fmt.Println(s)
		time.Sleep(20 * time.Millisecond)
	}

	time.Sleep(5 * time.Second)

}

func TestCacheSet(t *testing.T) {

	testLocal := LocalCacheTest()
	testRemote := RemoteCacheTest(testLocal)
	testDataSource := DataSourceAdaptorTest(testRemote)

	cacheInst := NewCacheWithMetric[string, *tests.Student]("cache_set_test", metrics.NewMetricsLogger(), testLocal, testRemote, testDataSource)

	s1 := tests.Student{
		Name: "张三",
		Age:  18,
		Time: time.Now().UnixNano(),
	}

	err := cacheInst.Set(context.Background(), &s1)
	if err != nil {
		panic(err)
	}

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
		fmt.Println("DataSourceAdaptorTest,if this print is means the db query may happend")
		return &tests.Student{
			Name: key,
			Age:  18,
			Time: time.Now().UnixNano(),
		}, true, nil
	}, datasource.WithSingleFlightWaitTime(200*time.Millisecond))
}
