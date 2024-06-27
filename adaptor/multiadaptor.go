package adaptor

import "context"

// Keys key集合
type Keys[K comparable] []K

// Values 返回值集合，暂使用map结构
type Values[K comparable, V Metadata] map[K]V

// ValueCol 批量数据操作对象
type ValueCol[V Metadata] []V

// NewValueFunc 创建新对象函数
type NewValueFunc[V Metadata] func() V

// MultiAdaptor 批量数据操作适配器接口
type MultiAdaptor[K comparable, V Metadata] interface {
	// Name 适配器名称，需要在当前业务场景中保证唯一
	Name() string
	// Get 读取对象
	Get(ctx context.Context, keys Keys[K], vals Values[K, V], fn NewValueFunc[V]) (Keys[K], error)
	// Set 写入对象
	Set(ctx context.Context, vals ValueCol[V]) error
	// Del 删除对象
	Del(ctx context.Context, keys Keys[K]) error
}
