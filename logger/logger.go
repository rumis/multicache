package logger

import "errors"

// Logger 日志接口定义
type Logger interface {
	Debug(format string, v ...any)
	Info(format string, v ...any)
	Warn(format string, v ...any)
	Error(format string, v ...any)
}

// 定义日志级别
type Level int

const (
	LevelDebug Level = iota + 1
	LevelInfo
	LevelWarn
	LevelError
)

var defaultLevel Level = LevelDebug
var defaultLogger Logger

func init() {
	defaultLogger, _ = NewZapSugarLogger()
}

// SetLogger 设置日志实例
func SetLevel(l Level) error {
	if l < LevelDebug || l > LevelError {
		return errors.New("invalid log level")
	}
	defaultLevel = l
	return nil
}

// SetLogger 设置日志实例
func SetLogger(l Logger) error {
	if l == nil {
		return errors.New("invalid logger")
	}
	defaultLogger = l
	return nil
}

// Debug 调试
func Debug(format string, v ...any) {
	if defaultLogger == nil {
		return
	}
	if defaultLevel <= LevelDebug {
		defaultLogger.Debug(format, v...)
	}
}

// Info 信息
func Info(format string, v ...any) {
	if defaultLogger == nil {
		return
	}
	if defaultLevel <= LevelInfo {
		defaultLogger.Info(format, v...)
	}
}

// Warn 警告
func Warn(format string, v ...any) {
	if defaultLogger == nil {
		return
	}
	if defaultLevel <= LevelWarn {
		defaultLogger.Warn(format, v...)
	}
}

// Error 错误
func Error(format string, v ...any) {
	if defaultLogger == nil {
		return
	}
	if defaultLevel <= LevelError {
		defaultLogger.Error(format, v...)
	}
}
