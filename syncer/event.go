package syncer

import (
	"encoding/json"
	"time"
)

type EventType int

const (
	EventTypeAdd EventType = iota + 1
	EventTypeDelete
)

// CacheSyncEvent 数据同步事件
type CacheSyncEvent struct {
	ClientID  string        `json:"clientId"`
	EventType EventType     `json:"eventType"`
	Key       string        `json:"key"`
	Val       []byte        `json:"val"`
	TTL       time.Duration `json:"ttl"`
}

// Encode 对象序列化
func (e *CacheSyncEvent) Encode() string {
	buf, err := json.Marshal(e)
	if err != nil {
		return ""
	}
	return string(buf)
}

// Decode 对象反序列化
func (e *CacheSyncEvent) Decode(buf []byte) error {
	return json.Unmarshal(buf, e)
}
