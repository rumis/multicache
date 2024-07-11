package prometheus

import (
	"fmt"
	"testing"
	"time"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

const queryHost = "http://172.17.0.3:9091/"

func TestQuery(t *testing.T) {

	query := NewQuery(WithPromHttpApiQueryHost(queryHost))
	_, err := query.Query("up", time.Now())
	if err != nil {
		t.Error(err)
	}

}

func TestQueryRange(t *testing.T) {
	query := NewQuery(WithPromHttpApiQueryHost(queryHost))
	_, err := query.Range("sum by(le) (increase(jiaoyan_websocket_client_message_request_time_bucket[1m]))", v1.Range{
		Start: time.Now().Add(-time.Hour * 24),
		End:   time.Now(),
		Step:  time.Minute,
	})
	if err != nil {
		t.Error(err)
	}
}

func TestLableNames(t *testing.T) {
	query := NewQuery(WithPromHttpApiQueryHost(queryHost))
	labels, err := query.LabelNames(time.Now().Add(-time.Hour*24), time.Now(), "jiaoyan_multicache_event")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(labels)
}

func TestLableValues(t *testing.T) {
	query := NewQuery(WithPromHttpApiQueryHost(queryHost))
	labels, err := query.LabelValues("", time.Now().Add(-time.Hour*24), time.Now())
	if err != nil {
		t.Error(err)
	}
	fmt.Println(labels)
}

func TestSeries(t *testing.T) {
	query := NewQuery(WithPromHttpApiQueryHost(queryHost))
	// jiaoyan_websocket_client_alive_count
	// {__name__=~"job:.*"}
	set, err := query.Series(time.Now().Add(-time.Hour*24), time.Now(), `{__name__=~"^jiaoyan_multicache.+"}`)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(set)

	// // 去重
	// nameMap := make(map[string]struct{})
	// for _, v := range m {
	// 	name, ok := v["__name__"]
	// 	if !ok {
	// 		continue
	// 	}
	// 	nameMap[string(name)] = struct{}{}
	// }
	// // 转数组
	// names := make([]string, 0, len(nameMap))
	// for v := range nameMap {
	// 	names = append(names, v)
	// }
}
