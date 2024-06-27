package adaptor

import (
	"context"
)

// Adaptor 定义了核心的数据读写删除等适配接口
type Adaptor[K comparable, V Metadata] interface {
	// Name 适配器名称，需要在当前业务场景中保证唯一
	Name() string
	// Get 读取对象
	Get(ctx context.Context, key K, value V) (bool, error)
	// Set 写入对象
	Set(ctx context.Context, value V) error
	// Del 删除对象
	Del(ctx context.Context, key K) error
}
