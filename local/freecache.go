package local

import (
	"sync"

	"github.com/coocood/freecache"
)

var freeCacheClient *freecache.Cache
var freeCacheOnce sync.Once
var freeCacheSize Size = 128 * MB

// InitFreeCacheSize 初始化FreeCache使用的内存大小
func InitFreeCacheSize(s Size) {
	freeCacheSize = s
}

// FreeCacheClient 获取FreeCache客户端
func FreeCacheClient() *freecache.Cache {
	freeCacheOnce.Do(func() {
		freeCacheClient = freecache.NewCache(int(freeCacheSize))
	})
	return freeCacheClient
}
