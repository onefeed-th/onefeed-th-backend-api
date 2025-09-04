package logger

import (
	"context"
	"log/slog"
	"os"
	"sync"
)

// Logger interface defines logging methods with slog compatibility
type Logger interface {
	Debug(ctx context.Context, message string, args ...any)
	Info(ctx context.Context, message string, args ...any)
	Warn(ctx context.Context, message string, args ...any)
	Error(ctx context.Context, message string, args ...any)
	Fatal(ctx context.Context, message string, args ...any)
	With(args ...any) Logger
	WithService(service string) Logger
	Sync() error
}

type slogLogger struct {
	logger  *slog.Logger
	service string
}

var (
	defaultLogger *slogLogger
	once          sync.Once
)

func init() {
	InitLogger()
}

// InitLogger initializes the global slog logger with optimized configuration
func InitLogger() {
	once.Do(func() {
		// Set log level from environment or default to INFO
		var logLevel slog.Level
		switch os.Getenv("LOG_LEVEL") {
		case "DEBUG":
			logLevel = slog.LevelDebug
		case "INFO":
			logLevel = slog.LevelInfo
		case "WARN":
			logLevel = slog.LevelWarn
		case "ERROR":
			logLevel = slog.LevelError
		default:
			logLevel = slog.LevelInfo
		}

		// Create handler with JSON format and custom options
		opts := &slog.HandlerOptions{
			Level:     logLevel,
			AddSource: true, // Add caller info for debugging
		}

		handler := slog.NewJSONHandler(os.Stdout, opts)
		logger := slog.New(handler)

		defaultLogger = &slogLogger{
			logger:  logger.With("service", "onefeed-backend"),
			service: "onefeed-backend",
		}
	})
}

// New creates a new logger with the specified service name
func New(service string) Logger {
	return &slogLogger{
		logger:  defaultLogger.logger.With("service", service),
		service: service,
	}
}

// GetDefault returns the default logger
func GetDefault() Logger {
	return defaultLogger
}

func (l *slogLogger) WithService(service string) Logger {
	return &slogLogger{
		logger:  l.logger.With("service", service),
		service: service,
	}
}

func (l *slogLogger) With(args ...any) Logger {
	return &slogLogger{
		logger:  l.logger.With(args...),
		service: l.service,
	}
}

func (l *slogLogger) Debug(ctx context.Context, message string, args ...any) {
	l.logger.DebugContext(ctx, message, l.addContextArgs(ctx, args...)...)
}

func (l *slogLogger) Info(ctx context.Context, message string, args ...any) {
	l.logger.InfoContext(ctx, message, l.addContextArgs(ctx, args...)...)
}

func (l *slogLogger) Warn(ctx context.Context, message string, args ...any) {
	l.logger.WarnContext(ctx, message, l.addContextArgs(ctx, args...)...)
}

func (l *slogLogger) Error(ctx context.Context, message string, args ...any) {
	l.logger.ErrorContext(ctx, message, l.addContextArgs(ctx, args...)...)
}

func (l *slogLogger) Fatal(ctx context.Context, message string, args ...any) {
	l.logger.ErrorContext(ctx, message, l.addContextArgs(ctx, args...)...)
	os.Exit(1)
}

func (l *slogLogger) Sync() error {
	// slog doesn't require explicit sync like zap
	// This method exists for interface compatibility
	return nil
}

// addContextArgs extracts fields from context and adds them to log args
func (l *slogLogger) addContextArgs(ctx context.Context, args ...any) []any {
	if ctx == nil {
		return args
	}

	var contextArgs []any
	
	// Extract trace ID from context
	if traceID, ok := ctx.Value("trace_id").(string); ok && traceID != "" {
		contextArgs = append(contextArgs, "trace_id", traceID)
	}
	
	// Extract user ID from context  
	if userID, ok := ctx.Value("user_id").(string); ok && userID != "" {
		contextArgs = append(contextArgs, "user_id", userID)
	}
	
	// Extract request ID from context
	if requestID, ok := ctx.Value("request_id").(string); ok && requestID != "" {
		contextArgs = append(contextArgs, "request_id", requestID)
	}

	// Combine context args with provided args
	return append(contextArgs, args...)
}

// Convenience functions using default logger
func Debug(ctx context.Context, message string, args ...any) {
	defaultLogger.Debug(ctx, message, args...)
}

func Info(ctx context.Context, message string, args ...any) {
	defaultLogger.Info(ctx, message, args...)
}

func Warn(ctx context.Context, message string, args ...any) {
	defaultLogger.Warn(ctx, message, args...)
}

func Error(ctx context.Context, message string, args ...any) {
	defaultLogger.Error(ctx, message, args...)
}

func Fatal(ctx context.Context, message string, args ...any) {
	defaultLogger.Fatal(ctx, message, args...)
}

// Sync flushes any buffered log entries
func Sync() error {
	return defaultLogger.Sync()
}