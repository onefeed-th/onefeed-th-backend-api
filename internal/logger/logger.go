package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"
)

// LogLevel represents the severity of a log entry
type LogLevel string

const (
	DEBUG LogLevel = "DEBUG"
	INFO  LogLevel = "INFO"
	WARN  LogLevel = "WARN"
	ERROR LogLevel = "ERROR"
	FATAL LogLevel = "FATAL"
)

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     LogLevel               `json:"level"`
	Message   string                 `json:"message"`
	Service   string                 `json:"service"`
	TraceID   string                 `json:"trace_id,omitempty"`
	UserID    string                 `json:"user_id,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`
	File      string                 `json:"file,omitempty"`
	Line      int                    `json:"line,omitempty"`
	Function  string                 `json:"function,omitempty"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Duration  *time.Duration         `json:"duration,omitempty"`
}

// Logger interface defines logging methods
type Logger interface {
	Debug(ctx context.Context, message string, fields ...map[string]interface{})
	Info(ctx context.Context, message string, fields ...map[string]interface{})
	Warn(ctx context.Context, message string, fields ...map[string]interface{})
	Error(ctx context.Context, message string, err error, fields ...map[string]interface{})
	Fatal(ctx context.Context, message string, err error, fields ...map[string]interface{})
	WithFields(fields map[string]interface{}) Logger
	WithService(service string) Logger
}

type structuredLogger struct {
	service    string
	baseFields map[string]interface{}
	level      LogLevel
}

var defaultLogger *structuredLogger

func init() {
	defaultLogger = &structuredLogger{
		service:    "onefeed-backend",
		baseFields: make(map[string]interface{}),
		level:      INFO,
	}
}

// SetLevel sets the global log level
func SetLevel(level LogLevel) {
	defaultLogger.level = level
}

// New creates a new structured logger
func New(service string) Logger {
	return &structuredLogger{
		service:    service,
		baseFields: make(map[string]interface{}),
		level:      INFO,
	}
}

// GetDefault returns the default logger
func GetDefault() Logger {
	return defaultLogger
}

func (l *structuredLogger) WithFields(fields map[string]interface{}) Logger {
	newFields := make(map[string]interface{})
	for k, v := range l.baseFields {
		newFields[k] = v
	}
	for k, v := range fields {
		newFields[k] = v
	}

	return &structuredLogger{
		service:    l.service,
		baseFields: newFields,
		level:      l.level,
	}
}

func (l *structuredLogger) WithService(service string) Logger {
	return &structuredLogger{
		service:    service,
		baseFields: l.baseFields,
		level:      l.level,
	}
}

func (l *structuredLogger) Debug(ctx context.Context, message string, fields ...map[string]interface{}) {
	if l.shouldLog(DEBUG) {
		l.log(ctx, DEBUG, message, nil, fields...)
	}
}

func (l *structuredLogger) Info(ctx context.Context, message string, fields ...map[string]interface{}) {
	if l.shouldLog(INFO) {
		l.log(ctx, INFO, message, nil, fields...)
	}
}

func (l *structuredLogger) Warn(ctx context.Context, message string, fields ...map[string]interface{}) {
	if l.shouldLog(WARN) {
		l.log(ctx, WARN, message, nil, fields...)
	}
}

func (l *structuredLogger) Error(ctx context.Context, message string, err error, fields ...map[string]interface{}) {
	if l.shouldLog(ERROR) {
		l.log(ctx, ERROR, message, err, fields...)
	}
}

func (l *structuredLogger) Fatal(ctx context.Context, message string, err error, fields ...map[string]interface{}) {
	l.log(ctx, FATAL, message, err, fields...)
	os.Exit(1)
}

func (l *structuredLogger) shouldLog(level LogLevel) bool {
	levels := map[LogLevel]int{
		DEBUG: 0,
		INFO:  1,
		WARN:  2,
		ERROR: 3,
		FATAL: 4,
	}

	return levels[level] >= levels[l.level]
}

func (l *structuredLogger) log(ctx context.Context, level LogLevel, message string, err error, fields ...map[string]interface{}) {
	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     level,
		Message:   message,
		Service:   l.service,
		Fields:    make(map[string]interface{}),
	}

	// Add base fields
	for k, v := range l.baseFields {
		entry.Fields[k] = v
	}

	// Add context-specific fields
	if ctx != nil {
		if traceID, ok := ctx.Value("trace_id").(string); ok && traceID != "" {
			entry.TraceID = traceID
		}
		if userID, ok := ctx.Value("user_id").(string); ok && userID != "" {
			entry.UserID = userID
		}
		if requestID, ok := ctx.Value("request_id").(string); ok && requestID != "" {
			entry.RequestID = requestID
		}
	}

	// Add provided fields
	for _, fieldMap := range fields {
		if fieldMap != nil {
			for k, v := range fieldMap {
				entry.Fields[k] = v
			}
		}
	}

	// Add error if present
	if err != nil {
		entry.Error = err.Error()
	}

	// Add caller information for ERROR and FATAL levels
	if level == ERROR || level == FATAL {
		if pc, file, line, ok := runtime.Caller(2); ok {
			entry.File = file
			entry.Line = line
			if fn := runtime.FuncForPC(pc); fn != nil {
				entry.Function = fn.Name()
			}
		}
	}

	// Remove empty fields map
	if len(entry.Fields) == 0 {
		entry.Fields = nil
	}

	// Marshal and output
	jsonData, err := json.Marshal(entry)
	if err != nil {
		log.Printf("Failed to marshal log entry: %v", err)
		log.Printf("[%s] %s: %s", level, l.service, message)
		return
	}

	fmt.Println(string(jsonData))
}

// Convenience functions using default logger
func Debug(ctx context.Context, message string, fields ...map[string]interface{}) {
	defaultLogger.Debug(ctx, message, fields...)
}

func Info(ctx context.Context, message string, fields ...map[string]interface{}) {
	defaultLogger.Info(ctx, message, fields...)
}

func Warn(ctx context.Context, message string, fields ...map[string]interface{}) {
	defaultLogger.Warn(ctx, message, fields...)
}

func Error(ctx context.Context, message string, err error, fields ...map[string]interface{}) {
	defaultLogger.Error(ctx, message, err, fields...)
}

func Fatal(ctx context.Context, message string, err error, fields ...map[string]interface{}) {
	defaultLogger.Fatal(ctx, message, err, fields...)
}