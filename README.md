# lib-logger
*Common Logging Library*

This repository contains the `Log` interface and `standardlogger` implementation
that logs to stdout.

## Import package
```golang
import (
	"github.com/dataphos/lib-logger/logger"
	"github.com/dataphos/lib-logger/standardlogger"
)
```

## How To Use

### Labels
To use the logger you first have to define Labels.
Labels are a map-like object containing key-value string pairs sent with every log entry.

Labels creation:
```golang
// only *strings* are allowed as keys and values
labels := logger.Labels{"product": "Persistor", "clientId": "client0", "license": "enterprise"}

// logger.L is an alias of logger.Labels
// this should be used in the Add method (see below)
labels := logger.L{"product": "Persistor", "clientId": "client0", "license": "enterprise"}
```
This creates `"product"="Persistor", "clientId"="client0", "license"="enterprise"` key-value pairs.

Labels also allow you to add and delete keys:
```golang
labels := logger.Labels{"product": "Persistor", "clientId": "client0", "license": "enterprise"}
labels.Del("clientId", "license")
labels.Add(logger.L{"product": "SchemaRegistry", "cluster": "C1"}) // overrides product and adds cluster
```

When defining multiple loggers it is useful to have one master `Labels` object
and many children that add additional information:
```golang
mainLabels := logger.Labels{"product": "Persistor", "clientId": "client0"}

// subscriber component
// create an independent copy of the mainLabels and add component=subscriber key-value pair
subscriberLabels := mainLabels.Clone()
subscriberLabels.Add(logger.L{"component": "subscriber"})

// publisher component
// chaining is allowed
publisherLabels := mainLabels.Clone().Add(logger.L{"component": "publisher"})
```
**Note:** without a call to `Clone()`, child `Labels` is not independent of 
the parent `Labels`, and calling `Add()` or `Del()` will modify the original `Labels`
instance.

### How To Log
Create an instance of the `standardlogger` and pass `Labels` instance as 
the parameter:
```golang
log := standardlogger.New(labels)
```

To log a message, use logging functions:
```golang
// ======================================
// LOGGING FUNCTIONS WITHOUT EXTRA FIELDS
// ======================================
log.Info("Info")
log.Warn("Warning")
// Error, Fatal, Panic require error code
log.Error("Error", 0)
log.Fatal("Error", 0)
log.Panic("Panic", 0)

// ======================================
// LOGGING FUNCTIONS WITH EXTRA FIELDS
// SHORT SYNTAX
// logger.F is an alias of logger.Fields
// ======================================
log.Infow("Info", logger.F{"userId": 55, "objId": 64})
log.Warnw("Warn", logger.F{"objId": 43})
log.Errorw("Error", 0, logger.F{"objId": 43})
log.Fatalw("Fatal", 0, logger.F{"userId": 55})
log.Panicw("Panic", 0, logger.F{"reqId": 22})
```

### How To Pass a `log` Object
To pass a `log`, use the `logger.Log` interface:
```golang
func giveMeTheLog(log logger.Log) {
    log.Info("Log passed")
}
```

### How To Log Panics
To properly log panics, make a deferred call to `log.PanicLogger()` at the
beginning of every goroutine. Possible deferred recovery function should be 
deferred **before** the PanicLogger. 

The main goroutine begins at `func main()` so
that is where you should make a deferred call to `log.PanicLogger()`.
```golang
func main() {
    defer func() {
        if r:=recover(); r!=nil {
            // recover panicking goroutine	
        }	
    }()
    labels := logger.Labels{"product": "Persistor"}
    log := standardlogger.New(labels)	
    defer log.PanicLogger()
    
    // panic can be called directly
    panic("PANIC!")
    
    // it can also be called with the log.Panicw method.
    // this allows you to set error code and 
    // additional information
    log.Panicw("Panic", 1000, logger.F{"reqId": 22})
    
    // goroutine
    go func() {
        routineLabels := labels.Clone().
        	Add(logger.L{"component": "routine", "remove": "me", "and": "me"}).
        	Del("remove", "and")
        
    	routineLog := standardlogger.New(routineLabels)
    	defer routineLog.PanicLogger()
    	
    	// code...
    }()
}
```
### How To Log With Log Level

Create an instance of the `standardlogger`, pass `Lables` instance and to set desired log level use `WithLogLevel(level)` 
You can set level to `logger.LevelInfo`, `logger.LevelWarn`, `logger.LevelError`, `logger.LevelPanic` and `logger.LevelFatal`. Logs of a lower level won't be printed to stdout.
Here is an example of how to do this:

```golang
logger := standardlogger.New(labels, standardlogger.WithLogLevel(logger.LevelInfo))
```
You can log a message using logging functions as shown before.

# Testing
Standard logger has a `NewForTesting` constructor that keeps logged records in memory:
```golang
log, logs := NewForTesting(logger.Labels{"product": "Persistor"})
```
where `log` is an instance of the `standardlogger` and `logs` is an instance of 
[zaptest observer](go.uber.org/zap/zaptest/observer)'s `ObservedLogs`.

`ObservedLogs` holds all logged records as a list of `LoggedEntry` instances.
They in turn hold the underlying `Entry` and the `Context`:
```golang
type LoggedEntry struct {
    zapcore.Entry
    Context []zapcore.Field
}
```
Important fields in the `zapcore.Entry` are `Message` and `Level`.

The following example shows basic usage of the `NewForTesting` in a unit test:
```golang
func TestNewForTesting(t *testing.T) {
	log, logs := NewForTesting(logger.Labels{})
	log.Info("Info")

	expectedMsg := "Info"
	entry := logs.All()[0]
	if entry.Message != expectedMsg {
		t.Errorf("Wrong message, want %s.", expectedMsg)
	}
}
```

This example demonstrates how to test when logging with extra fields:
```golang
func TestNewForTesting_WithLabelsOnWarn(t *testing.T) {
	log, logs := NewForTesting(
		logger.Labels{
			"product": "Persistor",
			"license": "enterprise",
		})
	log.Warn("Warn")

	expectedFields := []zap.Field{
		zap.String("product", "Persistor"),
		zap.String("license", "enterprise"),
	}
	expectedMsg := "Warn"

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
```

More examples can be found in `standardlogger/standardlogger_test.go`.

## Full Example
Full working example can be found at `example/main.go`

## Important
Standard logger is a thin wrapper around Uber `zap` logging library: https://github.com/uber-go/zap.

### Sampling
Sampling is **enabled** to protect from a flood of errors.
Logs are **dropped intentionally** by zap when sampling is enabled. 
Sampling will cause **repeated logs within a second** to be sampled. 
Read more at: https://github.com/uber-go/zap/blob/master/FAQ.md#why-are-some-of-my-logs-missing

## Creating a New Logger
To create a new logger, implement the `Log` interface that can be
found in the `logger` package.
