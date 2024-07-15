package logger

import (
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// ZapSugarLogger 基于zap实现的日志
type ZapSugarLogger struct {
	zapLogger *zap.SugaredLogger
}

// NewZapSugarLogger 创建一个基于zap实现的日志对象
func NewZapSugarLogger() (*ZapSugarLogger, error) {
	zapLog, err := zap.Config{
		Level:       zap.NewAtomicLevelAt(zap.DebugLevel),
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         "json",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}.Build()
	if err != nil {
		return nil, errors.WithMessage(err, "create zap logger failed")
	}
	return &ZapSugarLogger{
		zapLogger: zapLog.Sugar(),
	}, nil
}

// Debug 调试
func (z *ZapSugarLogger) Debug(format string, v ...any) {
	if z.zapLogger == nil {
		return
	}
	defer z.zapLogger.Sync()
	z.zapLogger.Debugw(format, v...)
}

// Info 信息
func (z *ZapSugarLogger) Info(format string, v ...any) {
	if z.zapLogger == nil {
		return
	}
	defer z.zapLogger.Sync()
	z.zapLogger.Infow(format, v...)
}

// Warn 警告
func (z *ZapSugarLogger) Warn(format string, v ...any) {
	if z.zapLogger == nil {
		return
	}
	defer z.zapLogger.Sync()
	z.zapLogger.Warnw(format, v...)
}

// Error 错误
func (z *ZapSugarLogger) Error(format string, v ...any) {
	if z.zapLogger == nil {
		return
	}
	defer z.zapLogger.Sync()
	z.zapLogger.Errorw(format, v...)
}
