package log

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tevino/abool"
)

// concept
/*
- Logging function:
  - check if file-based levelling enabled
    - if yes, check if level is active on this file
  - check if level is active
  - send data to backend via big buffered channel
- Backend:
  - wait until there is time for writing logs
  - write logs
  - configurable if logged to folder (buffer + rollingFileAppender) and/or console
  - console: log everything above INFO to stderr
- Channel overbuffering protection:
  - if buffer is full, trigger write
- Anti-Importing-Loop:
  - everything imports logging
  - logging is configured by main module and is supplied access to configuration and taskmanager
*/

type severity uint32

type logLine struct {
	msg       string
	trace     *ContextTracer
	level     severity
	timestamp time.Time
	file      string
	line      int
}

func (ll *logLine) Equal(ol *logLine) bool {
	switch {
	case ll.msg != ol.msg:
		return false
	case ll.trace != nil || ol.trace != nil:
		return false
	case ll.file != ol.file:
		return false
	case ll.line != ol.line:
		return false
	case ll.level != ol.level:
		return false
	}
	return true
}

const (
	TraceLevel    severity = 1
	DebugLevel    severity = 2
	InfoLevel     severity = 3
	WarningLevel  severity = 4
	ErrorLevel    severity = 5
	CriticalLevel severity = 6
)

var (
	logBuffer             chan *logLine
	forceEmptyingOfBuffer chan bool

	logLevelInt = uint32(3)
	logLevel    = &logLevelInt

	pkgLevelsActive = abool.NewBool(false)
	pkgLevels       = make(map[string]severity)
	pkgLevelsLock   sync.Mutex

	logsWaiting     = make(chan bool, 1)
	logsWaitingFlag = abool.NewBool(false)

	shutdownSignal    = make(chan struct{}, 0)
	shutdownWaitGroup sync.WaitGroup

	initializing  = abool.NewBool(false)
	started       = abool.NewBool(false)
	startedSignal = make(chan struct{}, 0)

	testErrors = abool.NewBool(false)
)

func SetPkgLevels(levels map[string]severity) {
	pkgLevelsLock.Lock()
	pkgLevels = levels
	pkgLevelsLock.Unlock()
	pkgLevelsActive.Set()
}

func UnSetPkgLevels() {
	pkgLevelsActive.UnSet()
}

func SetLogLevel(level severity) {
	atomic.StoreUint32(logLevel, uint32(level))
}

func ParseLevel(level string) severity {
	switch strings.ToLower(level) {
	case "trace":
		return 1
	case "debug":
		return 2
	case "info":
		return 3
	case "warning":
		return 4
	case "error":
		return 5
	case "critical":
		return 6
	}
	return 0
}

func Start() (err error) {

	if !initializing.SetToIf(false, true) {
		return nil
	}

	logBuffer = make(chan *logLine, 8192)
	forceEmptyingOfBuffer = make(chan bool, 4)

	initialLogLevel := ParseLevel(logLevelFlag)
	if initialLogLevel > 0 {
		atomic.StoreUint32(logLevel, uint32(initialLogLevel))
	} else {
		err = fmt.Errorf("log warning: invalid log level \"%s\", falling back to level info", logLevelFlag)
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
	}

	// get and set file loglevels
	pkgLogLevels := pkgLogLevelsFlag
	if len(pkgLogLevels) > 0 {
		newPkgLevels := make(map[string]severity)
		for _, pair := range strings.Split(pkgLogLevels, ",") {
			splitted := strings.Split(pair, "=")
			if len(splitted) != 2 {
				err = fmt.Errorf("log warning: invalid file log level \"%s\", ignoring", pair)
				fmt.Fprintf(os.Stderr, "%s\n", err.Error())
				break
			}
			fileLevel := ParseLevel(splitted[1])
			if fileLevel == 0 {
				err = fmt.Errorf("log warning: invalid file log level \"%s\", ignoring", pair)
				fmt.Fprintf(os.Stderr, "%s\n", err.Error())
				break
			}
			newPkgLevels[splitted[0]] = fileLevel
		}
		SetPkgLevels(newPkgLevels)
	}

	startWriter()

	started.Set()
	close(startedSignal)

	return err
}

// Shutdown writes remaining log lines and then stops the logger.
func Shutdown() {
	close(shutdownSignal)
	shutdownWaitGroup.Wait()
}
