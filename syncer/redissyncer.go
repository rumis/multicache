package syncer

import (
	"context"
	"fmt"
	"runtime"

	"github.com/go-redis/redis/v8"
	"github.com/rumis/multicache/logger"
	"github.com/rumis/multicache/utils"
)

// 类型检测
var _ Syncer = (*RedisSyncer)(nil)

// RedisSyncer 基于Redis发布/订阅模式的数据同步器
type RedisSyncer struct {
	clientId string
	channel  string
	// Client对象
	innerClient *redis.Client
}

// NewRedisSyncer 基于Redis发布/订阅模式的数据同步器
func NewRedisSyncer(iclient *redis.Client, channel string) *RedisSyncer {
	return &RedisSyncer{
		innerClient: iclient,
		channel:     channel,
		clientId:    utils.UUID(),
	}
}

// ClientID 获取端ID
func (r *RedisSyncer) ClientID() string {
	return r.clientId
}

// Emit 广播数据
func (r *RedisSyncer) Emit(ctx context.Context, e *CacheSyncEvent) error {
	if r.innerClient == nil {
		return ErrNilClient
	}
	e.ClientID = r.clientId
	err := r.innerClient.Publish(ctx, r.channel, e.Encode()).Err()
	if err != nil {
		return err
	}
	return nil
}

// Read 读取数据
func (r *RedisSyncer) Subscribe(ctx context.Context, fn EventHandler) error {
	if r.innerClient == nil {
		return ErrNilClient
	}
	ch := r.innerClient.Subscribe(ctx, r.channel).Channel()
	go func() {
		defer func() {
			if err := recover(); err != nil {
				buf := make([]byte, 1<<16)
				runtime.Stack(buf, false)
				logger.Error(fmt.Sprint(err), "channel", r.channel, "stack", string(buf))
				r.Subscribe(ctx, fn)
			}
		}()
		for {
			m := <-ch
			msg := &CacheSyncEvent{}
			err := msg.Decode([]byte(m.Payload))
			if err != nil {
				logger.Error(fmt.Sprint(err), "channel", r.channel)
			}
			if msg.ClientID == r.clientId {
				continue
			}
			fn(msg)
		}
	}()
	return nil
}
