package standardlogger_test

import (
	"bytes"
	"io"
	"os"
	"regexp"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/dataphos/lib-logger/logger"
	"github.com/dataphos/lib-logger/standardlogger"
)

func TestGetLabelsAsZapFields(t *testing.T) {
	tests := []struct {
		name           string
		labels         logger.Labels
		expectedFields []zap.Field
	}{
		{"key0", logger.Labels{"key0": "val0"}, []zap.Field{zap.String("key0", "val0")}},
		{
			"key0,key1",
			logger.Labels{"key0": "val0", "key1": "val1"},
			[]zap.Field{zap.String("key0", "val0"), zap.String("key1", "val1")},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			results := standardlogger.GetLabelsAsZapFields(test.labels)

			for _, field := range test.expectedFields {
				found := false
				for _, res := range results {
					if field.Equals(res) {
						found = true

						break
					}
				}
				if !found {
					t.Errorf("getLabelsAsZapFields(%v)=%v, %v not found.", test.labels, results, field)
				}
			}
		})
	}
}

func TestGetLabelsKeys(t *testing.T) {
	tests := []struct {
		name    string
		labels  logger.Labels
		keysSet map[string]interface{}
	}{
		{"key0", logger.Labels{"key0": "val0"}, map[string]interface{}{"key0": true}},
		{"key0,key1", logger.Labels{"key0": "val0", "key1": "val1"}, map[string]interface{}{"key0": true, "key1": true}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			keys := standardlogger.GetLabelsKeys(test.labels)

			if len(keys) != len(test.keysSet) {
				t.Errorf("getLabelsKeys(%v)=%v wrong number of keys, %d expected.", test.labels, keys, len(test.keysSet))
			}

			for key := range test.keysSet {
				found := false
				for _, v := range keys {
					if v == key {
						found = true

						break
					}
				}

				if !found {
					t.Errorf("getLabelsKeys(%v)=%v, %s expected.", test.labels, keys, key)
				}
			}
		})
	}
}

func TestGetLoggerFieldsAsZapFields(t *testing.T) {
	tests := []struct {
		name           string
		fields         logger.Fields
		expectedFields []zap.Field
	}{
		{"key0", logger.Fields{"key0": "val0"}, []zap.Field{zap.String("key0", "val0")}},
		{
			"key0,key1",
			logger.Fields{"key0": "val0", "key1": 5},
			[]zap.Field{zap.String("key0", "val0"), zap.Int("key1", 5)},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			results := standardlogger.GetLoggerFieldsAsZapFields(test.fields)

			for _, field := range test.expectedFields {
				found := false
				for _, res := range results {
					if field.Equals(res) {
						found = true

						break
					}
				}
				if !found {
					t.Errorf("getLoggerFieldsAsZapFields(%v)=%v, %v not found.", test.fields, results, field)
				}
			}
		})
	}
}

func TestNew(t *testing.T) {
	labels := logger.Labels{"product": "Persistor"}
	standardlogger.New(labels)
}

func TestNew_ImplementsLoggerLog(t *testing.T) {
	labels := logger.Labels{"product": "Persistor"}
	// let the compiler do the check.
	log := standardlogger.New(labels)
	_ = log
}

func TestNew_TagsPresent(t *testing.T) {
	stderrBck := os.Stdout

	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	os.Stdout = write

	outC := make(chan string)

	go func() {
		var buf bytes.Buffer

		io.Copy(&buf, read)

		outC <- buf.String()
	}()

	log := standardlogger.New(logger.Labels{"label0": "val0"})
	log.Info("Info msg")

	write.Close()

	os.Stdout = stderrBck
	out := <-outC

	const wantSubstring = "\"tags\":[\"label0\"]"
	if !strings.Contains(out, wantSubstring) {
		t.Errorf("Tags not present, want %s.", wantSubstring)
	}
}

func TestNew_RFC3339NanoTimeFormat(t *testing.T) {
	stderrBck := os.Stdout

	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	os.Stdout = write

	outC := make(chan []byte)

	go func() {
		var buf bytes.Buffer

		io.Copy(&buf, read)

		outC <- buf.Bytes()
	}()

	log := standardlogger.New(logger.Labels{"label0": "val0"})
	log.Info("Info msg")

	write.Close()

	os.Stdout = stderrBck
	out := <-outC

	// https://regex101.com/r/qH0sU7/1
	re := regexp.MustCompile(`((?:(\d{4}-\d{2}-\d{2})T(\d{2}:\d{2}:\d{2}(?:\.\d+)?))(Z|[\+-]\d{2}:\d{2})?)`)

	if re.Find(out) == nil {
		t.Errorf("Timestamp in RFC3339Nano not present.")
	}
}

type MockZapLogger struct {
	Synced bool
}

func (s *MockZapLogger) Sync() error {
	s.Synced = true

	return nil
}

func (s *MockZapLogger) With(...zap.Field) *zap.Logger {
	return nil
}

func (s *MockZapLogger) Info(string, ...zap.Field) {
}

func (s *MockZapLogger) Warn(string, ...zap.Field) {
}

func (s *MockZapLogger) Error(string, ...zap.Field) {
}

func (s *MockZapLogger) Fatal(string, ...zap.Field) {
}

func (s *MockZapLogger) Panic(string, ...zap.Field) {
}

func (s *MockZapLogger) Core() zapcore.Core {
	return nil
}

func TestStandardLog_Flush(t *testing.T) {
	zapLogger := &MockZapLogger{Synced: false}

	log := standardlogger.StandardLog{
		ZapLogger: zapLogger,
	}

	log.Flush()

	if !zapLogger.Synced {
		t.Error("Logger not synced.")
	}
}

func TestStandardLog_Close(t *testing.T) {
	zapLogger := &MockZapLogger{Synced: false}

	log := standardlogger.StandardLog{
		ZapLogger: zapLogger,
	}

	log.Close()

	if !zapLogger.Synced {
		t.Error("Logger not synced.")
	}
}

func setupLogger() (*standardlogger.StandardLog, *observer.ObservedLogs) {
	core, logs := observer.New(zap.InfoLevel)

	return &standardlogger.StandardLog{
		ZapLogger: zap.New(core),
	}, logs
}

func TestStandardLog_Infow(t *testing.T) {
	log, logs := setupLogger()
	log.Infow("Info msg", logger.Fields{"product": "Persistor", "license": "enterprise"})

	expectedMsg := "Info msg"
	expectedFields := []zap.Field{zap.String("product", "Persistor"), zap.String("license", "enterprise")}

	entry := logs.All()[0]
	if entry.Message != expectedMsg {
		t.Errorf("Wrong message, want %s.", expectedMsg)
	}

	for _, field := range expectedFields {
		found := false

		for _, v := range entry.Context {
			if field.Equals(v) {
				found = true

				break
			}
		}

		if !found {
			t.Errorf("Field not found, want %v.", field)
		}
	}
}

func TestStandardLog_Warnw(t *testing.T) {
	log, logs := setupLogger()
	log.Warnw("Warn msg", logger.Fields{"product": "Persistor", "license": "enterprise"})

	expectedMsg := "Warn msg"
	expectedFields := []zap.Field{zap.String("product", "Persistor"), zap.String("license", "enterprise")}

	entry := logs.All()[0]

	if entry.Message != expectedMsg {
		t.Errorf("Wrong message, want %s.", expectedMsg)
	}

	for _, field := range expectedFields {
		found := false

		for _, v := range entry.Context {
			if field.Equals(v) {
				found = true

				break
			}
		}

		if !found {
			t.Errorf("Field not found, want %v.", field)
		}
	}
}

func TestStandardLog_Errorw(t *testing.T) {
	log, logs := setupLogger()
	log.Errorw("ERR", 0, logger.Fields{"product": "Persistor", "license": "enterprise"})

	expectedMsg := "ERR"
	expectedFields := []zap.Field{
		zap.String("product", "Persistor"), zap.String("license", "enterprise"),
		zap.Uint64("code", 0),
	}

	entry := logs.All()[0]

	if entry.Message != expectedMsg {
		t.Errorf("Wrong message, want %s.", expectedMsg)
	}

	for _, field := range expectedFields {
		found := false

		for _, v := range entry.Context {
			if field.Equals(v) {
				found = true

				break
			}
		}

		if !found {
			t.Errorf("Field not found, want %v.", field)
		}
	}
}

func TestStandardLog_Fatalw(t *testing.T) {
	// untestable
	// see go/pkg/mod/go.uber.org/zap@v1.16.0/logger.go:287
	// log.Fatal always invokes os.Exit().
}

func TestErrorCodeIsPresentInErrorw(t *testing.T) {
	stderrBck := os.Stderr

	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	os.Stderr = write

	outC := make(chan string)

	go func() {
		var buf bytes.Buffer

		io.Copy(&buf, read)

		outC <- buf.String()
	}()

	labels := logger.Labels{"key0": "val0"}
	log := standardlogger.New(labels)

	log.Errorw("Error msg", 1000, logger.F{})

	write.Close()

	os.Stderr = stderrBck
	out := <-outC

	const wantSubstring = "\"code\":1000"
	if !strings.Contains(out, wantSubstring) {
		t.Errorf("Error code missing, want substring '%s'.", wantSubstring)
	}
}

func TestErrorCodeIsPresentInPanicw(t *testing.T) {
	defer func() {
		if r := recover(); r != nil { //nolint:varnamelen //short variable makes sense here
			panicData, ok := r.(*standardlogger.PanicContainer)
			if !ok {
				t.Error("Panic without PanicContainer.")
			} else {
				if panicData.Code != 1000 {
					t.Error("Wrong error code, 1000 expected.")
				}
			}
		}
	}()

	labels := logger.Labels{"key0": "val0"}
	log := standardlogger.New(labels)

	log.Panicw("Error msg", 1000, logger.F{})
}

func TestLabelsPresentInInfow(t *testing.T) {
	stderrBck := os.Stdout

	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	os.Stdout = write

	outC := make(chan string)

	go func() {
		var buf bytes.Buffer

		io.Copy(&buf, read)

		outC <- buf.String()
	}()

	labels := logger.Labels{"key0": "val0"}
	log := standardlogger.New(labels)

	log.Infow("Info msg", logger.Fields{})

	write.Close()

	os.Stdout = stderrBck
	out := <-outC

	const wantSubstring = "\"key0\":\"val0\""
	if !strings.Contains(out, wantSubstring) {
		t.Errorf("Labels missing, want substring '%s'.", wantSubstring)
	}
}

func TestLabelsPresentInWarnw(t *testing.T) {
	stderrBck := os.Stdout

	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	os.Stdout = write

	outC := make(chan string)

	go func() {
		var buf bytes.Buffer

		io.Copy(&buf, read)

		outC <- buf.String()
	}()

	labels := logger.Labels{"key0": "val0"}
	log := standardlogger.New(labels)
	log.Warnw("Warn msg", logger.Fields{})

	write.Close()

	os.Stdout = stderrBck
	out := <-outC

	const wantSubstring = "\"key0\":\"val0\""
	if !strings.Contains(out, wantSubstring) {
		t.Errorf("Labels missing, want substring '%s'.", wantSubstring)
	}
}

func TestLabelsPresentInErrorw(t *testing.T) {
	stderrBck := os.Stderr

	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	os.Stderr = write

	outC := make(chan string)

	go func() {
		var buf bytes.Buffer

		io.Copy(&buf, read)

		outC <- buf.String()
	}()

	labels := logger.Labels{"key0": "val0"}
	log := standardlogger.New(labels)
	log.Errorw("Error msg", 1000, logger.Fields{})

	write.Close()

	os.Stderr = stderrBck
	out := <-outC

	const wantSubstring = "\"key0\":\"val0\""
	if !strings.Contains(out, wantSubstring) {
		t.Errorf("Labels missing, want substring '%s'.", wantSubstring)
	}
}

func TestPanicLogger(t *testing.T) {
	stderrBck := os.Stderr

	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	os.Stderr = write

	outC := make(chan string)

	go func() {
		var buf bytes.Buffer

		io.Copy(&buf, read)

		outC <- buf.String()
	}()

	defer func() {
		// panic expected.
		recover()

		write.Close()

		os.Stderr = stderrBck
		out := <-outC

		const wantSubstringCode = "\"code\":1000"
		if !strings.Contains(out, wantSubstringCode) {
			t.Errorf("Error code missing, want substring '%s'.", wantSubstringCode)
		}

		const wantSubstringLabels = "\"key0\":\"val0\""
		if !strings.Contains(out, wantSubstringLabels) {
			t.Errorf("Labels missing, want substring '%s'.", wantSubstringLabels)
		}

		const wantSubstringMsg = "\"msg\":\"PANIC!\""
		if !strings.Contains(out, wantSubstringMsg) {
			t.Errorf("Msg missing, want substring '%s'.", wantSubstringMsg)
		}
	}()

	labels := logger.Labels{"key0": "val0"}
	log := standardlogger.New(labels)

	defer log.PanicLogger()
	log.Panicw("PANIC!", 1000, logger.Fields{})
}

func TestPanicLoggerPlainPanic(t *testing.T) {
	stderrBck := os.Stderr

	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	os.Stderr = write

	outC := make(chan string)

	go func() {
		var buf bytes.Buffer

		io.Copy(&buf, read)

		outC <- buf.String()
	}()

	defer func() {
		// panic expected.
		recover()

		write.Close()

		os.Stderr = stderrBck
		out := <-outC

		const wantSubstringLabels = "\"key0\":\"val0\""
		if !strings.Contains(out, wantSubstringLabels) {
			t.Errorf("Labels missing, want substring '%s'.", wantSubstringLabels)
		}

		const wantSubstringMsg = "\"msg\":\"PANIC!\""
		if !strings.Contains(out, wantSubstringMsg) {
			t.Errorf("Msg missing, want substring '%s'.", wantSubstringMsg)
		}
	}()

	labels := logger.Labels{"key0": "val0"}
	log := standardlogger.New(labels)

	defer log.PanicLogger()
	panic("PANIC!")
}

func TestStandardLog_Info(t *testing.T) {
	stderrBck := os.Stderr

	_, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	os.Stderr = write

	defer func() {
		write.Close()

		os.Stderr = stderrBck
	}()

	labels := logger.Labels{"key0": "val0"}
	log := standardlogger.New(labels)

	log.Info("INFO")
}

func TestStandardLog_Warn(t *testing.T) {
	stderrBck := os.Stderr

	_, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	os.Stderr = write

	defer func() {
		write.Close()

		os.Stderr = stderrBck
	}()

	labels := logger.Labels{"key0": "val0"}
	log := standardlogger.New(labels)

	log.Warn("WARN")
}

func TestStandardLog_Error(t *testing.T) {
	stderrBck := os.Stderr

	_, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	os.Stderr = write

	defer func() {
		write.Close()

		os.Stderr = stderrBck
	}()

	labels := logger.Labels{"key0": "val0"}
	log := standardlogger.New(labels)

	log.Error("ERROR", 0)
}

func TestStandardLog_Panic(t *testing.T) {
	stderrBck := os.Stderr

	_, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	os.Stderr = write

	defer func() {
		write.Close()

		os.Stderr = stderrBck
	}()

	labels := logger.Labels{"key0": "val0"}
	log := standardlogger.New(labels)

	defer func() {
		recover()
	}()

	log.Panic("PANIC", 0)

	t.Error("Panic expected.")
}

func TestStandardLog_InfoCaller(t *testing.T) {
	stderrBck := os.Stdout

	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	os.Stdout = write

	outC := make(chan string)

	go func() {
		var buf bytes.Buffer

		io.Copy(&buf, read)

		outC <- buf.String()
	}()

	labels := logger.Labels{"key0": "val0"}
	log := standardlogger.New(labels)

	log.Info("Info msg")

	write.Close()

	os.Stdout = stderrBck
	out := <-outC

	const wantSubstring = "\"caller\":\"standardlogger/standardlogger_test.go"
	if !strings.Contains(out, wantSubstring) {
		t.Errorf("Wrong caller, want substring '%s'.", wantSubstring)
	}
}

func TestStandardLog_WarnCaller(t *testing.T) {
	stderrBck := os.Stdout

	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	os.Stdout = write

	outC := make(chan string)

	go func() {
		var buf bytes.Buffer

		io.Copy(&buf, read)

		outC <- buf.String()
	}()

	labels := logger.Labels{"key0": "val0"}
	log := standardlogger.New(labels)
	log.Warn("Warn msg")

	write.Close()

	os.Stdout = stderrBck
	out := <-outC

	const wantSubstring = "\"caller\":\"standardlogger/standardlogger_test.go"
	if !strings.Contains(out, wantSubstring) {
		t.Errorf("Wrong caller, want substring '%s'.", wantSubstring)
	}
}

func TestStandardLog_ErrorCaller(t *testing.T) {
	stderrBck := os.Stderr

	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	os.Stderr = write

	outC := make(chan string)

	go func() {
		var buf bytes.Buffer

		io.Copy(&buf, read)

		outC <- buf.String()
	}()

	labels := logger.Labels{"key0": "val0"}
	log := standardlogger.New(labels)

	log.Error("Error msg", 0)

	write.Close()

	os.Stderr = stderrBck
	out := <-outC

	const wantSubstring = "\"caller\":\"standardlogger/standardlogger_test.go"
	if !strings.Contains(out, wantSubstring) {
		t.Errorf("Wrong caller, want substring '%s'.", wantSubstring)
	}
}

func setupLoggerWithLevel(core zapcore.Core) (*standardlogger.StandardLog, *observer.ObservedLogs) {
	coreObserver, logs := observer.New(zapcore.LevelOf(core))

	return &standardlogger.StandardLog{
		ZapLogger: zap.New(zapcore.NewTee(coreObserver, core)),
	}, logs
}

func TestLogLevelInfo(t *testing.T) {
	stderrBck := os.Stderr

	_, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	os.Stderr = write

	defer func() {
		write.Close()

		os.Stderr = stderrBck
	}()

	labels := logger.Labels{"product": "Persistor"}
	log := standardlogger.New(labels, standardlogger.WithLogLevel(logger.LevelInfo))
	core := standardlogger.GetCore(log.(*standardlogger.StandardLog)) //nolint:forcetypeassert //not necessary in tests.

	logger, logs := setupLoggerWithLevel(core)

	logger.Info("INFO")
	logger.Warn("WARN")
	logger.Error("ERROR", 0)

	expectedMsg := "INFO"

	entry := logs.All()[0]

	if entry.Message != expectedMsg {
		t.Errorf("Wrong message, want %s.", expectedMsg)
	}

	expectedNumberOfLogs := 3
	if logs.Len() != expectedNumberOfLogs {
		t.Errorf("Wrong number of logs, want %d.", expectedNumberOfLogs)
	}
}

func TestLogLevelWarn(t *testing.T) {
	stderrBck := os.Stderr

	_, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	os.Stderr = write

	defer func() {
		write.Close()

		os.Stderr = stderrBck
	}()

	labels := logger.Labels{"product": "Persistor"}
	log := standardlogger.New(labels, standardlogger.WithLogLevel(logger.LevelWarn))
	core := standardlogger.GetCore(log.(*standardlogger.StandardLog)) //nolint:forcetypeassert //not necessary in tests.
	logger, logs := setupLoggerWithLevel(core)

	logger.Info("INFO")
	logger.Warn("WARN")
	logger.Error("ERROR", 0)

	expectedMsg := "WARN"

	entry := logs.All()[0]

	if entry.Message != expectedMsg {
		t.Errorf("Wrong message, want %s.", expectedMsg)
	}

	expectedNumberOfLogs := 2
	if logs.Len() != expectedNumberOfLogs {
		t.Errorf("Wrong number of logs, want %d.", expectedNumberOfLogs)
	}
}

func TestLogLevelError(t *testing.T) {
	stderrBck := os.Stderr

	_, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	os.Stderr = write

	defer func() {
		write.Close()

		os.Stderr = stderrBck
	}()

	labels := logger.Labels{"product": "Persistor"}
	log := standardlogger.New(labels, standardlogger.WithLogLevel(logger.LevelError))
	core := standardlogger.GetCore(log.(*standardlogger.StandardLog)) //nolint:forcetypeassert //not necessary in tests.
	logger, logs := setupLoggerWithLevel(core)

	logger.Info("INFO")
	logger.Warn("WARN")
	logger.Error("ERROR", 0)

	expectedMsg := "ERROR"

	entry := logs.All()[0]

	if entry.Message != expectedMsg {
		t.Errorf("Wrong message, want %s.", expectedMsg)
	}

	expectedNumberOfLogs := 1
	if logs.Len() != expectedNumberOfLogs {
		t.Errorf("Wrong number of logs, want %d.", expectedNumberOfLogs)
	}
}

func TestLogLevelPanic(t *testing.T) {
	stderrBck := os.Stderr

	_, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	os.Stderr = write

	defer func() {
		recover()
		write.Close()

		os.Stderr = stderrBck
	}()

	labels := logger.Labels{"product": "Persistor"}
	log := standardlogger.New(labels, standardlogger.WithLogLevel(logger.LevelPanic))
	core := standardlogger.GetCore(log.(*standardlogger.StandardLog)) //nolint:forcetypeassert //not necessary in tests.
	logger, logs := setupLoggerWithLevel(core)                        //nolint:staticcheck //wrongly flagged

	logger.Info("INFO")
	logger.Warn("WARN")
	logger.Error("ERROR", 0)
	logger.Panic("PANIC", 0)

	expectedMsg := "PANIC"

	entry := logs.All()[0]
	if entry.Message != expectedMsg {
		t.Errorf("Wrong message, want %s.", expectedMsg)
	}

	expectedNumberOfLogs := 1
	if logs.Len() != expectedNumberOfLogs {
		t.Errorf("Wrong number of logs, want %d.", expectedNumberOfLogs)
	}
}
