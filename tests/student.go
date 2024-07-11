package tests

import (
	"github.com/rumis/multicache/encoding/msgpack"

	"github.com/rumis/multicache/adaptor"
)

var _ adaptor.Metadata = (*Student)(nil)

// Student 测试用对象示例
type Student struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
	Time int64  `json:"time"`
}

func (s *Student) Key() string {
	return s.Name
}

func (s *Student) Value() ([]byte, error) {
	buf, err := msgpack.Encode(s)
	return buf, err
}

func (s *Student) Decode(buf []byte) error {
	err := msgpack.Decode(buf, s)
	return err
}

func (s *Student) Zero() bool {
	return s.Name == ""
}
