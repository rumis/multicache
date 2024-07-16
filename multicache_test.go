package multicache

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/rumis/multicache/adaptor"
	"github.com/rumis/multicache/datasource"
	"github.com/rumis/multicache/local"
	"github.com/rumis/multicache/metrics"
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

func TestMultiCacheSet(t *testing.T) {

	testLocalMulti := MultiLocalCacheTest()
	testRemoteMulti := MultiRemoteCacheTest(testLocalMulti)
	testDataSourceMulti := MultiMysqlAdaptorTest(testRemoteMulti)

	multiCacheInst := NewMultiCacheWithMetric[string, *tests.Student]("multicache_test", metrics.NewMetricsLogger(), testLocalMulti, testRemoteMulti, testDataSourceMulti)

	err := multiCacheInst.Set(context.Background(), []*tests.Student{
		{
			Name: "张三",
			Age:  18,
			Time: time.Now().Unix(),
		},
		{
			Name: "李四",
			Age:  19,
			Time: time.Now().Unix(),
		},
	})

	if err != nil {
		panic(err)
	}

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
