package syncer

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/rumis/multicache/tests"
)

func TestRedisSyncer(t *testing.T) {

	redisClient := tests.NewRedisClient()

	s1 := NewRedisSyncer(redisClient, "channel_test")

	s2 := NewRedisSyncer(redisClient, "channel_test")

	s1.Subscribe(context.TODO(), func(e *CacheSyncEvent) {

		s1 := tests.Student{}
		s1.Decode(e.Val)

		t.Log("s1", s1)
	})
	for v := 0; v < 10; v++ {
		s := tests.Student{
			Name: "s_" + strconv.Itoa(v),
			Age:  v,
			Time: time.Now().UnixNano(),
		}
		valBuf, _ := s.Value()
		s2.Emit(context.TODO(), &CacheSyncEvent{
			EventType: EventTypeAdd,
			Key:       s.Key(),
			Val:       valBuf,
		})
		time.Sleep(time.Millisecond * 300)
	}
	time.Sleep(time.Millisecond * 1000)

}
