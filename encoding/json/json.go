package json

import "encoding/json"

// Encode 序列化
func Encode(v any) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// Decode 反序列化
func Decode(data []byte, v any) error {
	err := json.Unmarshal(data, v)
	if err != nil {
		return err
	}
	return nil
}
