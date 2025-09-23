package gotel

import (
	"fmt"
	"os"

	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ZapLogger is a thin wrapper around *zap.Logger that integrates with OpenTelemetry
// using the otelzap bridge.
type ZapLogger struct {
	*zap.Logger
}

// NewZapLogger builds a zap logger configured for the given service name.
func NewZapLogger(serviceName, version, level string, debug bool) (*ZapLogger, error) {
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
		Logger: logger.With(
			zap.String("version", version),
			zap.String("service", serviceName),
		),
	}, nil
}
