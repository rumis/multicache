package msgpack

import "github.com/vmihailenco/msgpack/v5"

// Encode 序列化
func Encode(v any) ([]byte, error) {
	b, err := msgpack.Marshal(v)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// Decode 反序列化
func Decode(data []byte, v any) error {
	err := msgpack.Unmarshal(data, v)
	if err != nil {
		return err
	}
	return nil
}
