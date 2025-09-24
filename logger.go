package gotel

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ZapLogger is a thin wrapper around *zap.Logger that integrates with OpenTelemetry
// using the otelzap bridge.
type ZapLogger struct {
	logger *zap.Logger
}

// NewZapLogger builds a zap logger configured for the given service name.
func newZapLogger(serviceName, version, level string, debug bool) (*ZapLogger, error) {
	// Parse the log level from the config string.
	logLevel, err := zapcore.ParseLevel(level)
	if err != nil {
		return nil, fmt.Errorf("failed to parse log level : %w", err)
	}

	// Use zap encoder config defaults.
	encCfg := zap.NewProductionEncoderConfig()
	if debug {
		encCfg = zap.NewDevelopmentEncoderConfig()
		// Add color for easier reading in development logs.
		encCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// Customize encoder keys and format.
	encCfg.LevelKey = "level"
	encCfg.CallerKey = "caller"
	encCfg.TimeKey = "timestamp"
	encCfg.MessageKey = "message"
	encCfg.StacktraceKey = "stacktrace"
	encCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	// Console/STDOUT core
	encoder := zapcore.NewJSONEncoder(encCfg)
	if debug {
		encoder = zapcore.NewConsoleEncoder(encCfg)
	}

	consoleCore := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), logLevel)

	// This converts zap entries to OpenTelemetry LogRecords.
	// By default it uses the global LoggerProvider (set via global.SetLoggerProvider).
	otelCore := otelzap.NewCore(serviceName)

	// Combine (tee) cores so logs go to both console and OTLP.
	core := zapcore.NewTee(consoleCore, otelCore)

	logger := zap.New(
		core,
		zap.AddCaller(),      // Include call site
		zap.AddCallerSkip(1), // Skip the wrapper frame
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

// Info logs an info level message and attaches trace/span info (if present in ctx).
func (l *ZapLogger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	l.WithContext(ctx).Info(msg, fields...)
}

// Error logs an error level message and attaches trace/span info (if present in ctx).
func (l *ZapLogger) Error(ctx context.Context, msg string, fields ...zap.Field) {
	l.WithContext(ctx).Error(msg, fields...)
}

// Warn logs a warning level message and attaches trace/span info (if present in ctx).
func (l *ZapLogger) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	l.WithContext(ctx).Warn(msg, fields...)
}

// Debug logs a debug level message and attaches trace/span info (if present in ctx).
func (l *ZapLogger) Debug(ctx context.Context, msg string, fields ...zap.Field) {
	l.WithContext(ctx).Debug(msg, fields...)
}

// Sync flushes any buffered log entries.
func (l *ZapLogger) Sync() error {
	return l.logger.Sync()
}
