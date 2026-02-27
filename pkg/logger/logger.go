package logger

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/Raihanarrasyid/iacctl/pkg/errors"
)

// LogLevel represents different logging levels
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal
)

var logLevelNames = map[LogLevel]string{
	LogLevelDebug: "DEBUG",
	LogLevelInfo:  "INFO",
	LogLevelWarn:  "WARN",
	LogLevelError: "ERROR",
	LogLevelFatal: "FATAL",
}

// Logger represents a structured logger
type Logger struct {
	level  LogLevel
	logger *log.Logger
}

var defaultLogger *Logger

// NewLogger creates a new logger with the specified level
func NewLogger(level LogLevel) *Logger {
	return &Logger{
		level:  level,
		logger: log.New(os.Stdout, "", 0), // No prefix, we'll add our own
	}
}

// SetDefault sets the default logger
func SetDefault(l *Logger) {
	defaultLogger = l
}

// GetDefault returns the default logger
func GetDefault() *Logger {
	if defaultLogger == nil {
		defaultLogger = NewLogger(LogLevelInfo)
	}
	return defaultLogger
}

// SetLevel sets the logging level
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// Debug logs a debug message
func (l *Logger) Debug(message string, fields ...interface{}) {
	l.log(LogLevelDebug, message, fields...)
}

// Info logs an info message
func (l *Logger) Info(message string, fields ...interface{}) {
	l.log(LogLevelInfo, message, fields...)
}

// Warn logs a warning message
func (l *Logger) Warn(message string, fields ...interface{}) {
	l.log(LogLevelWarn, message, fields...)
}

// Error logs an error message
func (l *Logger) Error(message string, fields ...interface{}) {
	l.log(LogLevelError, message, fields...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(message string, fields ...interface{}) {
	l.log(LogLevelFatal, message, fields...)
	os.Exit(1)
}

// ErrorWithAppError logs an AppError with structured information
func (l *Logger) ErrorWithAppError(err *errors.AppError, fields ...interface{}) {
	allFields := append([]interface{}{"error_code", err.Code, "error_message", err.Message}, fields...)
	if err.Details != "" {
		allFields = append(allFields, "error_details", err.Details)
	}
	if err.Context != nil {
		for k, v := range err.Context {
			allFields = append(allFields, fmt.Sprintf("ctx_%s", k), v)
		}
	}
	if err.Cause != nil {
		allFields = append(allFields, "cause", err.Cause.Error())
	}
	
	l.log(LogLevelError, "Application error occurred", allFields...)
}

// log is the internal logging method
func (l *Logger) log(level LogLevel, message string, fields ...interface{}) {
	if level < l.level {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	caller := getCaller()
	
	logLine := fmt.Sprintf("[%s] %s %s", timestamp, logLevelNames[level], message)
	
	if len(fields) > 0 {
		logLine += " | " + formatFields(fields...)
	}
	
	if caller != "" {
		logLine += fmt.Sprintf(" [%s]", caller)
	}
	
	l.logger.Println(logLine)
}

// formatFields formats key-value pairs for logging
func formatFields(fields ...interface{}) string {
	if len(fields)%2 != 0 {
		fields = append(fields, "(missing value)")
	}
	
	var parts []string
	for i := 0; i < len(fields); i += 2 {
		key := fmt.Sprintf("%v", fields[i])
		value := fmt.Sprintf("%v", fields[i+1])
		parts = append(parts, fmt.Sprintf("%s=%s", key, value))
	}
	
	return strings.Join(parts, ", ")
}

// getCaller gets the calling function for logging
func getCaller() string {
	pc, file, line, ok := runtime.Caller(3) // Skip 3 frames to get the actual caller
	if !ok {
		return ""
	}
	
	funcName := runtime.FuncForPC(pc).Name()
	if lastSlash := strings.LastIndex(funcName, "/"); lastSlash >= 0 {
		funcName = funcName[lastSlash+1:]
	}
	if lastDot := strings.LastIndex(funcName, "."); lastDot >= 0 {
		funcName = funcName[lastDot+1:]
	}
	
	if lastSlash := strings.LastIndex(file, "/"); lastSlash >= 0 {
		file = file[lastSlash+1:]
	}
	
	return fmt.Sprintf("%s:%s:%d", funcName, file, line)
}

// Convenience functions using the default logger
func Debug(message string, fields ...interface{}) {
	GetDefault().Debug(message, fields...)
}

func Info(message string, fields ...interface{}) {
	GetDefault().Info(message, fields...)
}

func Warn(message string, fields ...interface{}) {
	GetDefault().Warn(message, fields...)
}

func Error(message string, fields ...interface{}) {
	GetDefault().Error(message, fields...)
}

func Fatal(message string, fields ...interface{}) {
	GetDefault().Fatal(message, fields...)
}

func ErrorWithAppError(err *errors.AppError, fields ...interface{}) {
	GetDefault().ErrorWithAppError(err, fields...)
}

// WithFields creates a logger with additional fields
func (l *Logger) WithFields(fields ...interface{}) *Logger {
	// This is a simple implementation - in a real system you might want
	// to create a new logger instance that stores the fields
	return l
}

// WithError logs an error with additional context
func (l *Logger) WithError(err error, message string, fields ...interface{}) {
	allFields := append([]interface{}{"error", err.Error()}, fields...)
	l.Error(message, allFields...)
}

// WithAppError logs an AppError with additional context
func (l *Logger) WithAppError(err *errors.AppError, message string, fields ...interface{}) {
	l.ErrorWithAppError(err, fields...)
}

// SetLogLevelFromString sets log level from string
func SetLogLevelFromString(level string) {
	var logLevel LogLevel
	switch strings.ToUpper(level) {
	case "DEBUG":
		logLevel = LogLevelDebug
	case "INFO":
		logLevel = LogLevelInfo
	case "WARN", "WARNING":
		logLevel = LogLevelWarn
	case "ERROR":
		logLevel = LogLevelError
	case "FATAL":
		logLevel = LogLevelFatal
	default:
		logLevel = LogLevelInfo
	}
	
	GetDefault().SetLevel(logLevel)
}
