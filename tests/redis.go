package tests

import (
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// NewRedisClient 创建一个新的Redis客户端
// 每次单测只创建一次
func NewRedisClient() *redis.Client {
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	return redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})
}
