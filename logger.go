package gotel

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ZapLogger is a wrapper around *zap.Logger.
type ZapLogger struct {
	logger *zap.Logger
}

func newZapLogger(serviceName, version, level string, debug bool) (*ZapLogger, error) {
	logLevel, err := zapcore.ParseLevel(level)
	if err != nil {
		return nil, fmt.Errorf("failed to parse log level : %w", err)
	}

	encCfg := zap.NewProductionEncoderConfig()
	if debug {
		encCfg = zap.NewDevelopmentEncoderConfig()
		encCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	encCfg.LevelKey = "level"
	encCfg.CallerKey = "caller"
	encCfg.TimeKey = "timestamp"
	encCfg.MessageKey = "message"
	encCfg.StacktraceKey = "stacktrace"
	encCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	encoder := zapcore.NewJSONEncoder(encCfg)
	if debug {
		encoder = zapcore.NewConsoleEncoder(encCfg)
	}

	consoleCore := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), logLevel)
	otelCore := otelzap.NewCore(serviceName)
	core := zapcore.NewTee(consoleCore, otelCore)

	logger := zap.New(
		core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	return &ZapLogger{
		logger: logger.With(
			zap.String("version", version),
			zap.String("service", serviceName),
		),
	}, nil
}

// WithContext returns a child *zap.Logger that has the provided context attached as a field.
func (l *ZapLogger) WithContext(ctx context.Context) *zap.Logger {
	return l.logger.With(zap.Any("context", ctx))
}

// Info logs an info level message and attaches trace/span info.
func (l *ZapLogger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	l.WithContext(ctx).Info(msg, fields...)
}

// Error logs an error level message and attaches trace/span info.
func (l *ZapLogger) Error(ctx context.Context, msg string, fields ...zap.Field) {
	l.WithContext(ctx).Error(msg, fields...)
}

// Warn logs a warning level message and attaches trace/span info.
func (l *ZapLogger) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	l.WithContext(ctx).Warn(msg, fields...)
}

// Debug logs a debug level message and attaches trace/span info.
func (l *ZapLogger) Debug(ctx context.Context, msg string, fields ...zap.Field) {
	l.WithContext(ctx).Debug(msg, fields...)
}

// Sync flushes any buffered log entries.
func (l *ZapLogger) Sync() error {
	return l.logger.Sync()
}
