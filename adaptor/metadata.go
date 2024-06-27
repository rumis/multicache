package adaptor

// Metadata 原数据定义
type Metadata interface {
	// Key 该对象的Key
	Key() string
	// Value 对象序列化后的值
	Value() ([]byte, error)
	// Decode 对象反序列化
	Decode([]byte) error
	// Zero 判定对象是否为零值
	Zero() bool
}
