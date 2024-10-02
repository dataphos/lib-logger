// Copyright 2024 Syntio Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
