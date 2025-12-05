// Package utils provides logging utilities for the application
package utils

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger provides a standardized logging interface
type Logger interface {
	// Info logs an info level message
	Info(msg string, fields ...Field)
	// Warn logs a warning level message
	Warn(msg string, fields ...Field)
	// Error logs an error level message
	Error(msg string, fields ...Field)
	// Debug logs a debug level message
	Debug(msg string, fields ...Field)
	// Fatal logs a fatal level message and exits
	Fatal(msg string, fields ...Field)
	// With creates a child logger with additional fields
	With(fields ...Field) Logger
}

// Field represents a structured logging field
type Field interface {
	Key() string
	Value() interface{}
	Type() FieldType
}

// FieldType represents the type of a logging field
type FieldType int

const (
	StringType FieldType = iota
	IntType
	Float64Type
	BoolType
	ErrorType
	DurationType
	TimeType
)

// LogLevel represents the logging level
type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

// LoggerConfig holds configuration for creating loggers
type LoggerConfig struct {
	Level       LogLevel
	ServiceName string
	InstanceID  string
	Development bool
}

// field implements the Field interface using Zap fields internally
type field struct {
	zapField zap.Field
}

func (f field) Key() string {
	return f.zapField.Key
}

func (f field) Value() interface{} {
	return f.zapField.Interface
}

func (f field) Type() FieldType {
	switch f.zapField.Type {
	case zapcore.StringType:
		return StringType
	case zapcore.Int64Type, zapcore.Int32Type, zapcore.Int16Type, zapcore.Int8Type:
		return IntType
	case zapcore.Float64Type, zapcore.Float32Type:
		return Float64Type
	case zapcore.BoolType:
		return BoolType
	case zapcore.ErrorType:
		return ErrorType
	case zapcore.DurationType:
		return DurationType
	case zapcore.TimeType:
		return TimeType
	default:
		return StringType
	}
}

// Field creation functions
func String(key, value string) Field {
	return field{zapField: zap.String(key, value)}
}

func Int(key string, value int) Field {
	return field{zapField: zap.Int(key, value)}
}

func Int64(key string, value int64) Field {
	return field{zapField: zap.Int64(key, value)}
}

func Float64(key string, value float64) Field {
	return field{zapField: zap.Float64(key, value)}
}

func Bool(key string, value bool) Field {
	return field{zapField: zap.Bool(key, value)}
}

func Err(err error) Field {
	return field{zapField: zap.Error(err)}
}

func Duration(key string, value time.Duration) Field {
	return field{zapField: zap.Duration(key, value)}
}

// zapLogger implements the Logger interface using Zap
type zapLogger struct {
	zap *zap.Logger
}

func (l *zapLogger) Info(msg string, fields ...Field) {
	l.zap.Info(msg, l.convertFields(fields)...)
}

func (l *zapLogger) Warn(msg string, fields ...Field) {
	l.zap.Warn(msg, l.convertFields(fields)...)
}

func (l *zapLogger) Error(msg string, fields ...Field) {
	l.zap.Error(msg, l.convertFields(fields)...)
}

func (l *zapLogger) Debug(msg string, fields ...Field) {
	l.zap.Debug(msg, l.convertFields(fields)...)
}

func (l *zapLogger) Fatal(msg string, fields ...Field) {
	l.zap.Fatal(msg, l.convertFields(fields)...)
}

func (l *zapLogger) With(fields ...Field) Logger {
	return &zapLogger{
		zap: l.zap.With(l.convertFields(fields)...),
	}
}

func (l *zapLogger) convertFields(fields []Field) []zap.Field {
	zapFields := make([]zap.Field, len(fields))
	for i, f := range fields {
		zapFields[i] = f.(field).zapField
	}
	return zapFields
}

// NewLogger creates a new logger with the specified configuration
func NewLogger(config LoggerConfig) (Logger, error) {
	var zapConfig zap.Config

	if config.Development {
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		zapConfig = zap.NewProductionConfig()
	}

	// Convert our LogLevel to Zap's level
	var zapLevel zapcore.Level
	switch config.Level {
	case DebugLevel:
		zapLevel = zapcore.DebugLevel
	case InfoLevel:
		zapLevel = zapcore.InfoLevel
	case WarnLevel:
		zapLevel = zapcore.WarnLevel
	case ErrorLevel:
		zapLevel = zapcore.ErrorLevel
	case FatalLevel:
		zapLevel = zapcore.FatalLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	zapConfig.Level = zap.NewAtomicLevelAt(zapLevel)

	// Add service name and instance ID as initial fields
	zapConfig.InitialFields = map[string]interface{}{
		"service":     config.ServiceName,
		"instance_id": config.InstanceID,
	}

	zapLog, err := zapConfig.Build()
	if err != nil {
		return nil, err
	}

	return &zapLogger{zap: zapLog}, nil
}

// NewDevelopmentLogger creates a logger configured for development
func NewDevelopmentLogger(serviceName, instanceID string) (Logger, error) {
	config := LoggerConfig{
		Level:       InfoLevel,
		ServiceName: serviceName,
		InstanceID:  instanceID,
		Development: true,
	}
	return NewLogger(config)
}

// NewProductionLogger creates a logger configured for production
func NewProductionLogger(serviceName, instanceID string) (Logger, error) {
	config := LoggerConfig{
		Level:       InfoLevel,
		ServiceName: serviceName,
		InstanceID:  instanceID,
		Development: false,
	}
	return NewLogger(config)
}

// Global logger instance
var globalLogger Logger

// InitializeGlobalLogger initializes the global logger
func InitializeGlobalLogger(serviceName, instanceID string, development bool) error {
	var logger Logger
	var err error

	if development {
		logger, err = NewDevelopmentLogger(serviceName, instanceID)
	} else {
		logger, err = NewProductionLogger(serviceName, instanceID)
	}

	if err != nil {
		return err
	}

	globalLogger = logger
	return nil
}

// GetLogger returns the global logger instance
func GetLogger() Logger {
	if globalLogger == nil {
		// Fallback to a basic development logger if not initialized
		logger, _ := NewDevelopmentLogger("salesforce-splunk-migration", "default")
		globalLogger = logger
	}
	return globalLogger
}
