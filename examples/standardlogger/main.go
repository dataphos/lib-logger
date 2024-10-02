package main

import (
	"sync"

	"github.com/dataphos/lib-logger/logger"
	"github.com/dataphos/lib-logger/standardlogger"
)

var log logger.Log
var labels logger.Labels

func init() {
	// labels identify log messages
	labels = logger.Labels{"product": "Persistor", "id": "client0", "remove": "me"}
	labels.Del("remove").Add(logger.L{"license": "enterprise"})

	// log is constructed with Labels
	log = standardlogger.New(labels)
}

func IWillPanic(i int) {
	if i == 0 {
		// panic can be called with log.Panicw which allows you
		// to add error code and other fields you think are
		// important
		log.Panicw("Panicking!", 1000, logger.F{"i": i})
	}
	IWillPanic(i - 1)
}

func giveMeTheLog(log logger.Log) {
	log.Info("Log passed")
}

func main() {
	defer log.Close()
	defer log.PanicLogger()

	// ======================================
	// LOGGING FUNCTIONS WITHOUT EXTRA FIELDS
	// ======================================
	log.Info("Info")
	log.Warn("Warning")
	// Error, Fatal, Panic require error code
	log.Error("Error", 0)
	// log.Fatal("Error", 0) // uncomment to write log and perform os.Exit(1)
	// log.Panic("Panic", 0)

	// ======================================
	// LOGGING FUNCTIONS WITH EXTRA FIELDS
	// ======================================
	log.Infow("Info", logger.Fields{"userId": 55, "objId": 64})
	log.Warnw("Warn", logger.Fields{"objId": 43})
	log.Errorw("Error", 0, logger.Fields{"objId": 43})
	// log.Fatalw("Fatal", 0, logger.Fields{"userId": 55})
	// log.Panicw("Panic", 0, logger.Fields{"reqId": 22})

	// ======================================
	// LOGGING FUNCTIONS WITH EXTRA FIELDS
	// SHORT SYNTAX
	// logger.F is an alias of logger.Fields
	// ======================================
	log.Infow("Info", logger.F{"userId": 55, "objId": 64})
	log.Warnw("Warn", logger.F{"objId": 43})
	log.Errorw("Error", 0, logger.F{"objId": 43})
	// log.Fatalw("Fatal", 0, logger.F{"userId": 55})
	// log.Panicw("Panic", 0, logger.F{"reqId": 22})

	// PanicLogger logs any panic that it recovers
	// and re-emits the recovered panic.
	// This means it does not interfere with the normal
	// execution flow, but only observes.
	// uncomment this to trigger deferred PanicLogger
	// IWillPanic(5)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	defer wg.Wait()
	go func() {
		defer wg.Done()

		defer func() {
			// PanicLogger recovers and re-emits panic
			// so that it does not interfere with the normal
			// execution flow.
			recover()
		}()

		// ==========================
		// CHILD LABELS
		// ==========================
		// To create independent child Labels,
		// call Clone() on the parent Labels object.
		routineLabels := labels.Clone()

		// Add fields
		// logger.L is an alias of logger.Labels
		routineLabels.Add(logger.L{"component": "routine", "remove": "me", "and": "me"})

		// Delete keys
		routineLabels.Del("remove", "and")

		// Chaining is possible (.Clone().Add(...).Del(...) etc.)
		routineLabels2 := labels.Clone().
			Add(logger.L{"component": "routine", "remove": "me", "and": "me"}).
			Del("remove", "and")
		_ = routineLabels2

		routineLog := standardlogger.New(routineLabels)
		defer routineLog.PanicLogger()

		// logging in goroutine
		routineLog.Info("Info from goroutine")

		// panic can also be called directly which only
		// allows you to pass an error message
		panic("HELP!")
	}()

	// to pass a log, use logger.Log interface
	giveMeTheLog(log)

	// Creating logger with log level
	logger := standardlogger.New(labels, standardlogger.WithLogLevel(logger.LevelInfo))
	logger.Info("Info")
	logger.Warn("Warn")
	logger.Error("Error", 0)
}
