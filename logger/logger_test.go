package logger

import (
	"errors"
	"testing"
)

func TestOut(t *testing.T) {
	Debug("debug ", "date", "2024年06月05日")
	Info("info ", "date", "2024年06月06日")
	Warn("warn ", "date", "2024年06月07日")
	Error("error ", "date", "2024年06月08日")
}

func TestError(t *testing.T) {
	err := errors.New("test error 小毛驴")
	Error(err.Error(), "solution", "测试场景", "adaptor", "本地缓存", "keys", []string{"张三", "李四"}, "event", "GET")
}
