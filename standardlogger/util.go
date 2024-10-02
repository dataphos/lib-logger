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

package standardlogger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/dataphos/lib-logger/logger"
)

func getLevelAsZapLevel(lvl logger.Level) zapcore.Level {
	var zapLogLevel zapcore.Level

	switch lvl {
	case logger.LevelInfo:
		zapLogLevel = zap.InfoLevel
	case logger.LevelWarn:
		zapLogLevel = zap.WarnLevel
	case logger.LevelError:
		zapLogLevel = zap.ErrorLevel
	case logger.LevelPanic:
		zapLogLevel = zap.PanicLevel
	case logger.LevelFatal:
		zapLogLevel = zap.FatalLevel
	default:
		zapLogLevel = zap.InfoLevel
	}

	return zapLogLevel
}

func GetLabelsAsZapFields(labels logger.Labels) []zap.Field {
	fields := make([]zap.Field, len(labels))
	i := 0

	for k, v := range labels {
		fields[i] = zap.String(k, v)
		i++
	}

	return fields
}

func GetLabelsKeys(labels logger.Labels) []string {
	keys := make([]string, len(labels))
	i := 0

	for k := range labels {
		keys[i] = k
		i++
	}

	return keys
}

func GetLoggerFieldsAsZapFields(loggerFields logger.Fields) []zap.Field {
	fields := make([]zap.Field, len(loggerFields))
	i := 0

	for k, v := range loggerFields {
		fields[i] = zap.Any(k, v)
		i++
	}

	return fields
}
