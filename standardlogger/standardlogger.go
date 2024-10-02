// Package standardlogger provides a zap-based logging implementation with structured, leveled logging, including support for panic handling and configurable log levels through the StandardLog type and associated functions.
package standardlogger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/dataphos/lib-logger/logger"
)

type StandardLog struct {
	ZapLogger zapLogger
}

type zapLogger interface {
	Sync() error
	With(fields ...zap.Field) *zap.Logger
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	Panic(msg string, fields ...zap.Field)
	Core() zapcore.Core
}

type PanicContainer struct {
	msg    string
	Code   uint64
	fields logger.Fields
}

type Option func(*loggerSettings)

type loggerSettings struct {
	logLevel logger.Level
}

var defaultSettings = loggerSettings{
	logLevel: logger.LevelInfo,
}

// WithLogLevel returns Option that sets the desired log level.
func WithLogLevel(logLevel logger.Level) Option {
	return func(ls *loggerSettings) {
		ls.logLevel = logLevel
	}
}

func New(labels logger.Labels, opts ...Option) logger.Log {
	settings := defaultSettings

	for _, opt := range opts {
		opt(&settings)
	}

	zapLogLevel := getLevelAsZapLevel(settings.logLevel)

	// Enabled returns true if the given level is at or above this level.
	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel && zapLogLevel.Enabled(lvl)
	})

	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl < zapcore.ErrorLevel && zapLogLevel.Enabled(lvl)
	})

	consoleDebugging := zapcore.Lock(os.Stdout)
	consoleErrors := zapcore.Lock(os.Stderr)

	// Set timestamp to be in RFC3339Nano format.
	// This format is easily human-readable, unlike unix timestamp.
	// Fluent Bit can parse this format without custom scripting.
	conf := zap.NewProductionEncoderConfig()
	conf.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	consoleEncoder := zapcore.NewJSONEncoder(conf)

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, consoleErrors, highPriority),
		zapcore.NewCore(consoleEncoder, consoleDebugging, lowPriority),
	)

	// From a zapcore.Core, it's easy to construct a Logger.
	zapLogger := zap.New(core, zap.AddCallerSkip(1),
		zap.Fields(GetLabelsAsZapFields(labels)...),
		zap.Fields(zap.Strings("tags", GetLabelsKeys(labels))),
		zap.AddCaller(),
		zap.AddStacktrace(zap.ErrorLevel),
	)

	return &StandardLog{
		ZapLogger: zapLogger,
	}
}

func (l *StandardLog) Infow(msg string, fields logger.Fields) {
	l.ZapLogger.Info(msg, GetLoggerFieldsAsZapFields(fields)...)
}

func (l *StandardLog) Info(msg string) {
	l.ZapLogger.Info(msg)
}

func (l *StandardLog) Warnw(msg string, fields logger.Fields) {
	l.ZapLogger.Warn(msg, GetLoggerFieldsAsZapFields(fields)...)
}

func (l *StandardLog) Warn(msg string) {
	l.ZapLogger.Warn(msg)
}

func (l *StandardLog) Errorw(msg string, code uint64, fields logger.Fields) {
	l.ZapLogger.With(zap.Uint64("code", code)).Error(msg, GetLoggerFieldsAsZapFields(fields)...)
}

func (l *StandardLog) Error(msg string, code uint64) {
	l.ZapLogger.With(zap.Uint64("code", code)).Error(msg)
}

func (l *StandardLog) Fatalw(msg string, code uint64, fields logger.Fields) {
	l.ZapLogger.With(zap.Uint64("code", code)).Fatal(msg, GetLoggerFieldsAsZapFields(fields)...)
}

func (l *StandardLog) Fatal(msg string, code uint64) {
	l.ZapLogger.With(zap.Uint64("code", code)).Fatal(msg)
}

func (l *StandardLog) Panicw(msg string, code uint64, fields logger.Fields) {
	panicData := &PanicContainer{msg: msg, Code: code, fields: fields}
	panic(panicData)
}

func (l *StandardLog) Panic(msg string, code uint64) {
	l.Panicw(msg, code, logger.Fields{})
}

func (l *StandardLog) PanicLogger() {
	if r := recover(); r != nil { //nolint:varnamelen //short variable makes sense here
		panicData, ok := r.(*PanicContainer)
		if ok {
			fields := panicData.fields
			l.ZapLogger.With(zap.Uint64("code", panicData.Code)).Panic(panicData.msg, GetLoggerFieldsAsZapFields(fields)...)
		} else {
			l.ZapLogger.
				Panic(fmt.Sprint(r))
		}
	}
}

func (l *StandardLog) Close() {
	l.Flush()
}

func (l *StandardLog) Flush() {
	l.ZapLogger.Sync() //nolint:errcheck,gosec //no error handling here
}

func GetCore(l *StandardLog) zapcore.Core {
	return l.ZapLogger.Core()
}
