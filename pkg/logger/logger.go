package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Driver represents the logging driver type
type Driver string

const (
	DriverSyslog   Driver = "syslog"
	DriverJournald Driver = "journald"
	DriverFile     Driver = "file"
	DriverWindows  Driver = "windows"
	DriverNone     Driver = "none"
)

// Format represents the log format type
type Format string

const (
	FormatJSON Format = "json"
	FormatText Format = "text"
)

// Config holds the logging configuration
type Config struct {
	Driver Driver `yaml:"driver"`
	Format Format `yaml:"format"`
	Level  string `yaml:"level"`
	File   string `yaml:"file"` // Used when driver is "file"
}

// DefaultConfig returns the default logging configuration
func DefaultConfig() Config {
	return Config{
		Driver: DriverFile,
		Format: FormatJSON,
		Level:  "info",
		File:   "periphery.log",
	}
}

// Initialize sets up the global logger based on the configuration
// Returns a cleanup function that should be called on shutdown
func Initialize(cfg Config) (func(), error) {
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		return nil, fmt.Errorf("invalid log level %q: %w", cfg.Level, err)
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	var encoder zapcore.Encoder
	switch cfg.Format {
	case FormatJSON:
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	case FormatText:
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	default:
		return nil, fmt.Errorf("unsupported log format: %s", cfg.Format)
	}

	writer, closer, err := getWriter(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create log writer: %w", err)
	}

	core := zapcore.NewCore(encoder, writer, level)
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	zap.ReplaceGlobals(logger)

	cleanup := func() {
		_ = logger.Sync()
		if closer != nil {
			_ = closer()
		}
	}

	return cleanup, nil
}

// getWriter returns the appropriate WriteSyncer based on the driver configuration
func getWriter(cfg Config) (zapcore.WriteSyncer, func() error, error) {
	switch cfg.Driver {
	case DriverNone:
		return zapcore.AddSync(&noopWriter{}), nil, nil

	case DriverFile:
		if cfg.File == "" {
			return nil, nil, fmt.Errorf("file path is required for file driver")
		}
		file, err := os.OpenFile(cfg.File, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to open log file %s: %w", cfg.File, err)
		}
		return zapcore.AddSync(file), file.Close, nil

	case DriverSyslog:
		return getSyslogWriter(cfg)

	case DriverJournald:
		return getJournaldWriter(cfg)

	case DriverWindows:
		return getWindowsWriter(cfg)

	default:
		return nil, nil, fmt.Errorf("unsupported log driver: %s", cfg.Driver)
	}
}

// noopWriter is a writer that discards all writes
type noopWriter struct{}

func (nw *noopWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}
