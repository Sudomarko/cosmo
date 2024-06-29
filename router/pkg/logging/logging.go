package logging

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	requestIDField = "reqId"
)

type RequestIDKey struct{}

func New(prettyLogging bool, logFile *os.File, debug bool, level zapcore.Level) *zap.Logger {
	if logFile == nil {
		return newZapLogger(zapcore.AddSync(os.Stdout), prettyLogging, debug, level)
	} else {
		return newMultiWriterZapLogger(zapcore.AddSync(os.Stdout), zapcore.AddSync(logFile), prettyLogging, debug, level)
	}
}

func zapBaseEncoderConfig() zapcore.EncoderConfig {
	ec := zap.NewProductionEncoderConfig()
	ec.EncodeDuration = zapcore.SecondsDurationEncoder
	ec.TimeKey = "time"
	return ec
}

func ZapJsonEncoder() zapcore.Encoder {
	ec := zapBaseEncoderConfig()
	ec.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		nanos := t.UnixNano()
		millis := int64(math.Trunc(float64(nanos) / float64(time.Millisecond)))
		enc.AppendInt64(millis)
	}
	return zapcore.NewJSONEncoder(ec)
}

func zapConsoleEncoder() zapcore.Encoder {
	ec := zapBaseEncoderConfig()
	ec.ConsoleSeparator = " "
	ec.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05 PM")
	ec.EncodeLevel = zapcore.CapitalColorLevelEncoder
	return zapcore.NewConsoleEncoder(ec)
}

func attachBaseFields(logger *zap.Logger) *zap.Logger {
	host, err := os.Hostname()
	if err != nil {
		host = "unknown"
	}

	logger = logger.With(
		zap.String("hostname", host),
		zap.Int("pid", os.Getpid()),
	)

	return logger
}

func newZapLogger(syncer zapcore.WriteSyncer, prettyLogging bool, debug bool, level zapcore.Level) *zap.Logger {
	var encoder zapcore.Encoder
	var zapOpts []zap.Option

	if prettyLogging {
		encoder = zapConsoleEncoder()
	} else {
		encoder = ZapJsonEncoder()
	}

	if debug {
		zapOpts = append(zapOpts, zap.AddCaller())
	}

	zapOpts = append(zapOpts, zap.AddStacktrace(zap.ErrorLevel))

	zapLogger := zap.New(zapcore.NewCore(
		encoder,
		syncer,
		level,
	), zapOpts...)

	if prettyLogging {
		return zapLogger
	}

	zapLogger = attachBaseFields(zapLogger)

	return zapLogger
}

func newMultiWriterZapLogger(consoleSyncer zapcore.WriteSyncer, fileSyncer zapcore.WriteSyncer, prettyLogging bool, debug bool, level zapcore.Level) *zap.Logger {
	var consoleEncoder zapcore.Encoder
	var consoleCore zapcore.Core
	var fileEncoder zapcore.Encoder
	var fileCore zapcore.Core
	var zapOpts []zap.Option

	if prettyLogging {
		consoleEncoder = zapConsoleEncoder()
	} else {
		consoleEncoder = ZapJsonEncoder()
	}

	fileEncoder = ZapJsonEncoder()

	if debug {
		zapOpts = append(zapOpts, zap.AddCaller())
	}

	zapOpts = append(zapOpts, zap.AddStacktrace(zap.ErrorLevel))

	consoleCore = zapcore.NewCore(
		fileEncoder,
		fileSyncer,
		level,
	)

	fileCore = zapcore.NewCore(
		consoleEncoder,
		consoleSyncer,
		level,
	)

	// Combine the cores
	combinedCore := zapcore.NewTee(consoleCore, fileCore)

	zapLogger := zap.New(combinedCore, zapOpts...)

	zapLogger = attachBaseFields(zapLogger)

	return zapLogger
}

func ZapLogLevelFromString(logLevel string) (zapcore.Level, error) {
	switch strings.ToUpper(logLevel) {
	case "DEBUG":
		return zap.DebugLevel, nil
	case "INFO":
		return zap.InfoLevel, nil
	case "WARNING":
		return zap.WarnLevel, nil
	case "ERROR":
		return zap.ErrorLevel, nil
	case "FATAL":
		return zap.FatalLevel, nil
	case "PANIC":
		return zap.PanicLevel, nil
	default:
		return -1, fmt.Errorf("unknown log level: %s", logLevel)
	}
}

func NewLogFile(destination string) (*os.File, error) {
	if destination == "" {
		return nil, nil
	}

	// Get the current file's directory
	_, filePath, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("unable to get caller information")
	}
	currentDir := filepath.Dir(filePath)

	file, err := os.OpenFile(filepath.Join(currentDir, "../..", destination, "router_log.json"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	return file, err
}

func WithRequestID(reqID string) zap.Field {
	return zap.String(requestIDField, reqID)
}
