package logger

import (
	"context"
	"os"
	"strings"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger interface defines logging methods with zap compatibility
type Logger interface {
	Debug(ctx context.Context, message string, fields ...zap.Field)
	Info(ctx context.Context, message string, fields ...zap.Field)
	Warn(ctx context.Context, message string, fields ...zap.Field)
	Error(ctx context.Context, message string, fields ...zap.Field)
	Fatal(ctx context.Context, message string, fields ...zap.Field)
	With(fields ...zap.Field) Logger
	WithService(service string) Logger
	Sync() error
}

type zapLogger struct {
	logger  *zap.Logger
	service string
}

var (
	defaultLogger *zapLogger
	once          sync.Once
)

func init() {
	InitLogger()
}

// InitLogger initializes the global zap logger with optimized configuration
func InitLogger() {
	once.Do(func() {
		// Production-optimized configuration
		config := zap.NewProductionConfig()
		
		// Use JSON encoding for structured logs
		config.Encoding = "json"
		
		// Set log level from environment or default to INFO
		logLevel := os.Getenv("LOG_LEVEL")
		switch logLevel {
		case "DEBUG":
			config.Level.SetLevel(zapcore.DebugLevel)
		case "INFO":
			config.Level.SetLevel(zapcore.InfoLevel)
		case "WARN":
			config.Level.SetLevel(zapcore.WarnLevel)
		case "ERROR":
			config.Level.SetLevel(zapcore.ErrorLevel)
		default:
			config.Level.SetLevel(zapcore.InfoLevel)
		}

		// Add caller info for debugging
		config.DisableCaller = false
		config.DisableStacktrace = false

		// Custom time format
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

		// Add service field to all logs
		logger, err := config.Build(
			zap.AddCallerSkip(1), // Skip wrapper function in stack trace
		)
		if err != nil {
			panic("Failed to initialize logger: " + err.Error())
		}

		defaultLogger = &zapLogger{
			logger:  logger.With(zap.String("service", "onefeed-backend")),
			service: "onefeed-backend",
		}
	})
}

// New creates a new logger with the specified service name
func New(service string) Logger {
	return &zapLogger{
		logger:  defaultLogger.logger.With(zap.String("service", service)),
		service: service,
	}
}

// GetDefault returns the default logger
func GetDefault() Logger {
	return defaultLogger
}

// SetLevel sets the global log level
func SetLevel(level zapcore.Level) {
	defaultLogger.logger.Core().Enabled(level)
}

func (l *zapLogger) WithService(service string) Logger {
	return &zapLogger{
		logger:  l.logger.With(zap.String("service", service)),
		service: service,
	}
}

func (l *zapLogger) With(fields ...zap.Field) Logger {
	return &zapLogger{
		logger:  l.logger.With(fields...),
		service: l.service,
	}
}

func (l *zapLogger) Debug(ctx context.Context, message string, fields ...zap.Field) {
	l.logger.Debug(message, l.addContextFields(ctx, fields...)...)
}

func (l *zapLogger) Info(ctx context.Context, message string, fields ...zap.Field) {
	l.logger.Info(message, l.addContextFields(ctx, fields...)...)
}

func (l *zapLogger) Warn(ctx context.Context, message string, fields ...zap.Field) {
	l.logger.Warn(message, l.addContextFields(ctx, fields...)...)
}

func (l *zapLogger) Error(ctx context.Context, message string, fields ...zap.Field) {
	l.logger.Error(message, l.addContextFields(ctx, fields...)...)
}

func (l *zapLogger) Fatal(ctx context.Context, message string, fields ...zap.Field) {
	l.logger.Fatal(message, l.addContextFields(ctx, fields...)...)
}

func (l *zapLogger) Sync() error {
	// Sync can fail on stderr/stdout on some systems, which is not critical
	// We'll ignore these specific errors
	if err := l.logger.Sync(); err != nil {
		// Ignore sync errors for stderr/stdout as they're not critical
		if strings.Contains(err.Error(), "sync /dev/stderr") || 
		   strings.Contains(err.Error(), "sync /dev/stdout") ||
		   strings.Contains(err.Error(), "inappropriate ioctl for device") {
			return nil
		}
		return err
	}
	return nil
}

// addContextFields extracts fields from context and adds them to log fields
func (l *zapLogger) addContextFields(ctx context.Context, fields ...zap.Field) []zap.Field {
	if ctx == nil {
		return fields
	}

	contextFields := []zap.Field{}
	
	// Extract trace ID from context
	if traceID, ok := ctx.Value("trace_id").(string); ok && traceID != "" {
		contextFields = append(contextFields, zap.String("trace_id", traceID))
	}
	
	// Extract user ID from context  
	if userID, ok := ctx.Value("user_id").(string); ok && userID != "" {
		contextFields = append(contextFields, zap.String("user_id", userID))
	}
	
	// Extract request ID from context
	if requestID, ok := ctx.Value("request_id").(string); ok && requestID != "" {
		contextFields = append(contextFields, zap.String("request_id", requestID))
	}

	// Combine context fields with provided fields
	return append(contextFields, fields...)
}

// Convenience functions using default logger
func Debug(ctx context.Context, message string, fields ...zap.Field) {
	defaultLogger.Debug(ctx, message, fields...)
}

func Info(ctx context.Context, message string, fields ...zap.Field) {
	defaultLogger.Info(ctx, message, fields...)
}

func Warn(ctx context.Context, message string, fields ...zap.Field) {
	defaultLogger.Warn(ctx, message, fields...)
}

func Error(ctx context.Context, message string, fields ...zap.Field) {
	defaultLogger.Error(ctx, message, fields...)
}

func Fatal(ctx context.Context, message string, fields ...zap.Field) {
	defaultLogger.Fatal(ctx, message, fields...)
}

// Sync flushes any buffered log entries
func Sync() error {
	return defaultLogger.Sync()
}