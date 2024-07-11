package syncer

import (
	"context"
	"errors"
)

// ErrNilClient Redis客户端空
var ErrNilClient = errors.New("nil redis client")

// EventHandler 本地缓存数据同步事件处理函数
type EventHandler func(e *CacheSyncEvent)

// Syncer 数据同步器接口
type Syncer interface {
	// GetClientID 获取客户端ID
	ClientID() string
	// Emit 广播消息
	Emit(ctx context.Context, e *CacheSyncEvent) error
	// Subscribe 订阅消息
	Subscribe(ctx context.Context, fn EventHandler) error
}
