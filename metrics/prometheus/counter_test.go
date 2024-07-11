package prometheus

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestCounter(t *testing.T) {

	h := "http://172.17.0.3:9091/"

	groups := []map[string]string{{
		"adaptor": "datasource",
		"key":     "张三",
		"event":   "hit",
	}, {
		"adaptor": "redis",
		"key":     "张三",
		"event":   "miss",
	}, {
		"adaptor": "local",
		"key":     "张三",
		"event":   "miss",
	}, {
		"adaptor": "redis",
		"key":     "李四",
		"event":   "hit",
	}, {
		"adaptor": "local",
		"key":     "李四",
		"event":   "miss",
	}}

	counter := NewCounter("multicache", "cache_hitmiss", "multicache_test_solution", WithGatewayHost(h))
	defer counter.Stop()

	go func() {
		for {
			rand.Seed(time.Now().UnixNano())
			idx := rand.Intn(len(groups))
			group := groups[idx]
			labs := make([]Label, 0)
			for k, v := range group {
				labs = append(labs, Label{
					Name:  k,
					Value: v,
				})
			}
			err := counter.Incr(labs...)
			if err != nil {
				fmt.Println("add error:", err)
			}
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(30)))
		}
	}()

	time.Sleep(time.Minute * 1)

}
