package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

var (
	logLevelMap = map[string]zapcore.Level{
		"debug": zapcore.DebugLevel,
		"info":  zapcore.InfoLevel,
		"warn":  zapcore.WarnLevel,
		"error": zapcore.ErrorLevel,
		"fatal": zapcore.FatalLevel,
		"panic": zapcore.PanicLevel,
	}
)

type ZapLogger struct {
	*zap.Logger
}

func MustNewLogger(logLevel string) *ZapLogger {
	zl, err := NewLogger(logLevel)
	if err != nil {
		panic(err)
	}
	return zl
}

func NewLogger(logLevel string) (*ZapLogger, error) {
	level, exists := logLevelMap[logLevel]
	if !exists {
		return &ZapLogger{zap.NewNop()}, nil
	}

	config := zap.Config{
		Level:             zap.NewAtomicLevelAt(level),
		Development:       false,
		DisableCaller:     false,
		DisableStacktrace: true,
		Sampling:          nil,
		Encoding:          "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:    "timestamp",
			MessageKey: "msg",
			EncodeTime: zapcore.ISO8601TimeEncoder,
		},
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
		InitialFields:    map[string]any{"pid": os.Getpid()},
	}

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	// Skip reporting the wrapper as log caller
	logger = logger.WithOptions(zap.AddCallerSkip(1))

	return &ZapLogger{logger}, nil
}
