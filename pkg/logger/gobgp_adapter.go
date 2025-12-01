package logger

import (
	"go.uber.org/zap"

	gobgplog "github.com/osrg/gobgp/v3/pkg/log"
)

type gobgpLoggerAdapter struct {
	logger *zap.Logger
	level  gobgplog.LogLevel
}

func NewGoBGPLogger() gobgplog.Logger {
	return &gobgpLoggerAdapter{
		logger: zap.L(),
		level:  gobgplog.InfoLevel,
	}
}

func (l *gobgpLoggerAdapter) fieldsToZap(fields gobgplog.Fields) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	return zapFields
}

func (l *gobgpLoggerAdapter) Panic(msg string, fields gobgplog.Fields) {
	l.logger.Panic(msg, l.fieldsToZap(fields)...)
}

func (l *gobgpLoggerAdapter) Fatal(msg string, fields gobgplog.Fields) {
	l.logger.Fatal(msg, l.fieldsToZap(fields)...)
}

func (l *gobgpLoggerAdapter) Error(msg string, fields gobgplog.Fields) {
	l.logger.Error(msg, l.fieldsToZap(fields)...)
}

func (l *gobgpLoggerAdapter) Warn(msg string, fields gobgplog.Fields) {
	l.logger.Warn(msg, l.fieldsToZap(fields)...)
}

func (l *gobgpLoggerAdapter) Info(msg string, fields gobgplog.Fields) {
	l.logger.Info(msg, l.fieldsToZap(fields)...)
}

func (l *gobgpLoggerAdapter) Debug(msg string, fields gobgplog.Fields) {
	l.logger.Debug(msg, l.fieldsToZap(fields)...)
}

func (l *gobgpLoggerAdapter) SetLevel(level gobgplog.LogLevel) {
	l.level = level
}

func (l *gobgpLoggerAdapter) GetLevel() gobgplog.LogLevel {
	return l.level
}
