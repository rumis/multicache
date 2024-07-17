package prometheus

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/rumis/multicache/tests"
)

func TestHistogram(t *testing.T) {

	gatewayHost := tests.PromGatewayHost()

	hist := NewHistogram("multicache", "message_process_time", "cache_message_process_time", []float64{10, 20, 50, 100, 200, 500, 1000}, WithGatewayHost(gatewayHost))

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	go func() {
		for {
			err := hist.Observe(r.Float64()*1000, Label{"localcache", "xxx"})
			if err != nil {
				fmt.Println("add error:", err)
			}
			time.Sleep(time.Millisecond * time.Duration(r.Intn(30)))
		}
	}()
	time.Sleep(time.Second * 60)
}
