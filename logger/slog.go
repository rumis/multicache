package logger

import (
	"log"
)

// LocalLogger 基于slog实现的日志
type LocalLogger struct {
}

// Debug 调试
func (l *LocalLogger) Debug(format string, v ...any) {
	// slog.Debug(format, v...)
	log.Println("Debug", format, v)
}

// Info 信息
func (l *LocalLogger) Info(format string, v ...any) {
	// slog.Info(format, v...)
	log.Println("Info", format, v)
}

// Warn 警告
func (l *LocalLogger) Warn(format string, v ...any) {
	// slog.Warn(format, v...)
	log.Println("Warn", format, v)
}

// Error 错误
func (l *LocalLogger) Error(format string, v ...any) {
	// slog.Error(format, v...)
	log.Println("Error", format, v)
}
