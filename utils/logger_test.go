// Package utils provides logging utilities
package utils_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"salesforce-splunk-migration/utils"
)

func TestString(t *testing.T) {
	t.Run("Success_CreatesStringField", func(t *testing.T) {
		field := utils.String("key", "value")
		require.NotNil(t, field)
		assert.Equal(t, "key", field.Key())
		assert.Equal(t, utils.StringType, field.Type())
	})

	t.Run("Success_EmptyValue", func(t *testing.T) {
		field := utils.String("key", "")
		require.NotNil(t, field)
		assert.Equal(t, "key", field.Key())
	})

	t.Run("Success_SpecialCharacters", func(t *testing.T) {
		field := utils.String("key", "value with spaces & symbols!")
		require.NotNil(t, field)
		assert.Equal(t, "key", field.Key())
	})
}

func TestInt(t *testing.T) {
	t.Run("Success_PositiveInt", func(t *testing.T) {
		field := utils.Int("count", 42)
		require.NotNil(t, field)
		assert.Equal(t, "count", field.Key())
		assert.Equal(t, utils.IntType, field.Type())
	})

	t.Run("Success_NegativeInt", func(t *testing.T) {
		field := utils.Int("count", -10)
		require.NotNil(t, field)
		assert.Equal(t, "count", field.Key())
	})

	t.Run("Success_ZeroInt", func(t *testing.T) {
		field := utils.Int("count", 0)
		require.NotNil(t, field)
		assert.Equal(t, "count", field.Key())
	})
}

func TestInt64(t *testing.T) {
	t.Run("Success_MaxInt64", func(t *testing.T) {
		field := utils.Int64("bigcount", int64(9223372036854775807))
		require.NotNil(t, field)
		assert.Equal(t, "bigcount", field.Key())
	})

	t.Run("Success_MinInt64", func(t *testing.T) {
		field := utils.Int64("bigcount", int64(-9223372036854775808))
		require.NotNil(t, field)
		assert.Equal(t, "bigcount", field.Key())
	})

	t.Run("Success_ZeroInt64", func(t *testing.T) {
		field := utils.Int64("bigcount", int64(0))
		require.NotNil(t, field)
	})
}

func TestFloat64(t *testing.T) {
	t.Run("Success_PositiveFloat", func(t *testing.T) {
		field := utils.Float64("value", 3.14)
		require.NotNil(t, field)
		assert.Equal(t, "value", field.Key())
		assert.Equal(t, utils.Float64Type, field.Type())
	})

	t.Run("Success_NegativeFloat", func(t *testing.T) {
		field := utils.Float64("value", -2.5)
		require.NotNil(t, field)
	})

	t.Run("Success_ZeroFloat", func(t *testing.T) {
		field := utils.Float64("value", 0.0)
		require.NotNil(t, field)
	})

	t.Run("Success_LargeFloat", func(t *testing.T) {
		field := utils.Float64("value", 1e10)
		require.NotNil(t, field)
	})
}

func TestBool(t *testing.T) {
	t.Run("Success_True", func(t *testing.T) {
		field := utils.Bool("enabled", true)
		require.NotNil(t, field)
		assert.Equal(t, "enabled", field.Key())
		assert.Equal(t, utils.BoolType, field.Type())
	})

	t.Run("Success_False", func(t *testing.T) {
		field := utils.Bool("enabled", false)
		require.NotNil(t, field)
		assert.Equal(t, "enabled", field.Key())
	})
}

func TestErr(t *testing.T) {
	t.Run("Success_NilError", func(t *testing.T) {
		field := utils.Err(nil)
		require.NotNil(t, field)
		// When nil, zap.Error creates a field with nil value
	})

	t.Run("Success_WithError", func(t *testing.T) {
		err := errors.New("test error")
		field := utils.Err(err)
		require.NotNil(t, field)
		assert.Equal(t, "error", field.Key())
	})

	t.Run("Success_WrappedError", func(t *testing.T) {
		baseErr := errors.New("base error")
		wrappedErr := errors.Join(baseErr, errors.New("wrapped"))
		field := utils.Err(wrappedErr)
		require.NotNil(t, field)
	})
}

func TestDuration(t *testing.T) {
	t.Run("Success_Second", func(t *testing.T) {
		field := utils.Duration("elapsed", time.Second)
		require.NotNil(t, field)
		assert.Equal(t, "elapsed", field.Key())
		assert.Equal(t, utils.DurationType, field.Type())
	})

	t.Run("Success_Millisecond", func(t *testing.T) {
		field := utils.Duration("elapsed", time.Millisecond)
		require.NotNil(t, field)
	})

	t.Run("Success_ZeroDuration", func(t *testing.T) {
		field := utils.Duration("elapsed", 0)
		require.NotNil(t, field)
	})

	t.Run("Success_NegativeDuration", func(t *testing.T) {
		field := utils.Duration("elapsed", -time.Second)
		require.NotNil(t, field)
	})
}

func TestGetLogger(t *testing.T) {
	t.Run("Success_ReturnsLogger", func(t *testing.T) {
		logger := utils.GetLogger()
		require.NotNil(t, logger)
	})

	t.Run("Success_CanLogMessages", func(t *testing.T) {
		logger := utils.GetLogger()
		require.NotNil(t, logger)

		// Test that we can call logging methods without panic
		assert.NotPanics(t, func() {
			logger.Info("Test info message", utils.String("key", "value"))
			logger.Warn("Test warn message")
			logger.Error("Test error message")
			logger.Debug("Test debug message")
		})
	})

	t.Run("Success_MultipleCalls", func(t *testing.T) {
		logger1 := utils.GetLogger()
		logger2 := utils.GetLogger()
		assert.NotNil(t, logger1)
		assert.NotNil(t, logger2)
	})
}

func TestNewLogger(t *testing.T) {
	t.Run("Success_InfoLevel", func(t *testing.T) {
		config := utils.LoggerConfig{
			Level:       utils.InfoLevel,
			ServiceName: "test-service",
			InstanceID:  "test-instance",
			Development: true,
		}

		logger, err := utils.NewLogger(config)
		require.NoError(t, err)
		require.NotNil(t, logger)
	})

	t.Run("Success_DebugLevel", func(t *testing.T) {
		config := utils.LoggerConfig{
			Level:       utils.DebugLevel,
			ServiceName: "test-service",
			InstanceID:  "test-instance",
			Development: false,
		}

		logger, err := utils.NewLogger(config)
		require.NoError(t, err)
		require.NotNil(t, logger)
	})

	t.Run("Success_ProductionMode", func(t *testing.T) {
		config := utils.LoggerConfig{
			Level:       utils.InfoLevel,
			ServiceName: "prod-service",
			InstanceID:  "prod-instance",
			Development: false,
		}

		logger, err := utils.NewLogger(config)
		require.NoError(t, err)
		require.NotNil(t, logger)
	})

	t.Run("Success_CanLogAtDifferentLevels", func(t *testing.T) {
		config := utils.LoggerConfig{
			Level:       utils.InfoLevel,
			ServiceName: "test",
			Development: true,
		}

		logger, err := utils.NewLogger(config)
		require.NoError(t, err)

		assert.NotPanics(t, func() {
			logger.Info("Info message", utils.String("test", "value"))
			logger.Warn("Warning message")
			logger.Error("Error message")
		})
	})
}

func TestNewDevelopmentLogger(t *testing.T) {
	t.Run("Success_CreatesLogger", func(t *testing.T) {
		logger, err := utils.NewDevelopmentLogger("test-service", "test-instance")
		require.NoError(t, err)
		require.NotNil(t, logger)
	})

	t.Run("Success_CanLog", func(t *testing.T) {
		logger, err := utils.NewDevelopmentLogger("test-service", "test-instance")
		require.NoError(t, err)

		assert.NotPanics(t, func() {
			logger.Debug("Development log message")
			logger.Info("Info message")
		})
	})
}

func TestNewProductionLogger(t *testing.T) {
	t.Run("Success_CreatesLogger", func(t *testing.T) {
		logger, err := utils.NewProductionLogger("test-service", "test-instance")
		require.NoError(t, err)
		require.NotNil(t, logger)
	})

	t.Run("Success_CanLog", func(t *testing.T) {
		logger, err := utils.NewProductionLogger("test-service", "test-instance")
		require.NoError(t, err)

		assert.NotPanics(t, func() {
			logger.Info("Production log message")
			logger.Warn("Warning")
		})
	})
}

func TestInitializeGlobalLogger(t *testing.T) {
	t.Run("Success_DevelopmentMode", func(t *testing.T) {
		err := utils.InitializeGlobalLogger("test-service", "test-instance", true)
		assert.NoError(t, err)

		logger := utils.GetLogger()
		require.NotNil(t, logger)
	})

	t.Run("Success_ProductionMode", func(t *testing.T) {
		err := utils.InitializeGlobalLogger("prod-service", "prod-instance", false)
		assert.NoError(t, err)

		logger := utils.GetLogger()
		require.NotNil(t, logger)
	})
}

func TestLoggerWith(t *testing.T) {
	t.Run("Success_AddsField", func(t *testing.T) {
		logger := utils.GetLogger()
		childLogger := logger.With(utils.String("request_id", "123"))

		require.NotNil(t, childLogger)
	})

	t.Run("Success_MultipleFields", func(t *testing.T) {
		logger := utils.GetLogger()
		childLogger := logger.With(
			utils.String("request_id", "123"),
			utils.Int("user_id", 456),
		)

		require.NotNil(t, childLogger)
		assert.NotPanics(t, func() {
			childLogger.Info("Child logger message")
		})
	})

	t.Run("Success_NestedWith", func(t *testing.T) {
		logger := utils.GetLogger()
		child1 := logger.With(utils.String("level1", "value1"))
		child2 := child1.With(utils.String("level2", "value2"))

		require.NotNil(t, child1)
		require.NotNil(t, child2)
	})
}

func TestFieldTypes(t *testing.T) {
	t.Run("StringType", func(t *testing.T) {
		field := utils.String("key", "value")
		assert.Equal(t, utils.StringType, field.Type())
	})

	t.Run("IntType", func(t *testing.T) {
		field := utils.Int("count", 1)
		assert.Equal(t, utils.IntType, field.Type())
	})

	t.Run("Float64Type", func(t *testing.T) {
		field := utils.Float64("value", 1.0)
		assert.Equal(t, utils.Float64Type, field.Type())
	})

	t.Run("BoolType", func(t *testing.T) {
		field := utils.Bool("flag", true)
		assert.Equal(t, utils.BoolType, field.Type())
	})

	t.Run("DurationType", func(t *testing.T) {
		field := utils.Duration("time", time.Second)
		assert.Equal(t, utils.DurationType, field.Type())
	})
}

func TestLogLevels(t *testing.T) {
	levels := []struct {
		name  string
		level utils.LogLevel
	}{
		{"DebugLevel", utils.DebugLevel},
		{"InfoLevel", utils.InfoLevel},
		{"WarnLevel", utils.WarnLevel},
		{"ErrorLevel", utils.ErrorLevel},
		{"FatalLevel", utils.FatalLevel},
	}

	for _, tc := range levels {
		t.Run(tc.name, func(t *testing.T) {
			config := utils.LoggerConfig{
				Level:       tc.level,
				ServiceName: "test",
				Development: true,
			}

			logger, err := utils.NewLogger(config)
			require.NoError(t, err)
			require.NotNil(t, logger)
		})
	}
}

func TestLoggerConcurrency(t *testing.T) {
	t.Run("Success_ConcurrentWrites", func(t *testing.T) {
		logger := utils.GetLogger()

		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func(id int) {
				logger.Info("Concurrent log", utils.Int("id", id))
				done <- true
			}(i)
		}

		for i := 0; i < 10; i++ {
			<-done
		}
	})

	t.Run("Success_ConcurrentWithFields", func(t *testing.T) {
		logger := utils.GetLogger()

		done := make(chan bool)
		for i := 0; i < 5; i++ {
			go func(id int) {
				childLogger := logger.With(utils.Int("goroutine", id))
				childLogger.Info("Message from goroutine")
				done <- true
			}(i)
		}

		for i := 0; i < 5; i++ {
			<-done
		}
	})
}
