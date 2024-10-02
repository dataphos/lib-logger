// Package logger defines an interface for logging with various levels and methods for structured logging, including Info, Warn, Error, Fatal, and Panic, as well as options for attaching fields and handling panics, along with methods for flushing and closing the logger.
package logger

type Fields map[string]interface{}

type F = Fields

type Log interface {
	Info(msg string)

	Infow(msg string, fields Fields)

	Warn(msg string)

	Warnw(msg string, fields Fields)

	Error(msg string, code uint64)

	Errorw(msg string, code uint64, fields Fields)

	Fatal(msg string, code uint64)

	Fatalw(msg string, code uint64, fields Fields)

	Panic(msg string, code uint64)

	Panicw(msg string, code uint64, fields Fields)

	PanicLogger()

	Flush()

	Close()
}

type Level int8

const (
	LevelInfo = iota
	LevelWarn
	LevelError
	LevelPanic
	LevelFatal
)
