package logger

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// LogOptions ...
type LogOptions uint16

const (
	OPT_ALL_ENABLED       LogOptions = 0xFFFF
	OPT_ALL_DISABLED      LogOptions = 0x0000
	OPT_DEAFULT           LogOptions = 0x0003
	OPT_HAS_REPORT_CALLER LogOptions = 0x0001
	OPT_HAS_SHORT_CALLER  LogOptions = 0x0002

	FILE     string = "file"
	FUNCTION string = "func"
	RESOURCE string = "res"
	CATEGORY string = "cat"
)

// Logger struct
type Logger struct {
	logrus.Logger

	Options  LogOptions
	Resource map[string]string
	Category string
}

// LoggerHook ...
type LoggerHook struct {
	logrus.Hook

	Logger *Logger
}

// Implement Singleton pattern
var (
	once    sync.Once
	logger  *Logger
	defOpts LogOptions = OPT_DEAFULT
)

// New is a function to obtain a singleton instance of Logger.
func New() *Logger {
	return new(OPT_DEAFULT)
}

// NewWithOptions is a function to obtain and initialized a singleton instance of Logger with options.
func NewWithOptions(options LogOptions) *Logger {
	return new(options)
}

// NewAlways is a function to obtain a total new instance of Logger with options.
func NewAlways(options LogOptions) *Logger {
	return newAlways(options)
}

func new(_options LogOptions) *Logger {
	// Execute code block once.
	once.Do(func() {
		logger = newAlways(_options)
	})
	return logger
}

func newAlways(_options LogOptions) *Logger {
	_logger := &Logger{
		Logger: logrus.Logger{
			Out:          os.Stderr,
			Hooks:        make(logrus.LevelHooks),
			Level:        logrus.InfoLevel,
			ExitFunc:     os.Exit,
			ReportCaller: false,
		},
		Options: _options,
	}
	// Set json format.
	_logger.SetFormatter(&logrus.JSONFormatter{
		CallerPrettyfier: _logger.callerPrettyfier,
	})

	// Set Hookers
	_logger.Hooks.Add(LoggerHook{Logger: _logger})

	//
	if _logger.Options&OPT_HAS_REPORT_CALLER > 0 {
		_logger.SetReportCaller(true)
	}

	// Set log level.
	l, err := logrus.ParseLevel(os.Getenv("LOG_LEVEL"))
	if nil != err {
		l = logrus.InfoLevel
	}
	_logger.SetLevel(l)

	// initialize resource
	_logger.Resource = make(map[string]string)

	/* TODO: use goroutine id or thread id is better.
	// Generate logId
	logId := GenerateRunId() //
	instance.WithFields(logrus.Fields{"logId": logId})
	*/

	return _logger
}

// GenerateRunId ...
func GenerateRunId() string {
	rand.Seed(time.Now().UnixNano())
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, 10)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

////////////////////////////////////////////////////////////////////////////////
// Logger
////////////////////////////////////////////////////////////////////////////////

// callerPrettyfier
func (l *Logger) callerPrettyfier(caller *runtime.Frame) (function string, file string) {
	file = caller.File
	function = caller.Function

	if l.Options&OPT_HAS_SHORT_CALLER > 0 {
		if i := strings.LastIndex(caller.File, "/"); i >= 0 {
			file = caller.File[i+1:]
		}
		file = fmt.Sprintf("%s:%d", file, caller.Line)

		if i := strings.LastIndex(caller.Function, "."); i >= 0 {
			function = caller.Function[i+1:]
		}
	}
	return function, file
}

// parse resource
func (l *Logger) parseResource(resource string) (string, string) {
	idx := strings.Index(resource, ":")
	if idx == -1 {
		return "X", resource // X: means unknown resource type.
	} else {
		return string(resource[:idx]), string(resource[idx+1:])
	}
}

// convert map to string and separated by ','.
func (l *Logger) createKeyValuePairs(m map[string]string) string {
	b := &bytes.Buffer{}
	for key, value := range m {
		if b.Len() > 0 {
			fmt.Fprintf(b, ", ")
		}
		fmt.Fprintf(b, "%s:%s", key, value)
	}
	return b.String()
}

// SetResource set resource.
func (l *Logger) SetResource(resource string) *Logger {
	t, id := l.parseResource(resource)
	l.Resource[t] = id
	return l
}

// ClearResource clear resource.
func (l *Logger) ClearResource() *Logger {
	l.Resource = make(map[string]string)
	return l
}

// SetCategory set category.
func (l *Logger) SetCategory(category string) *Logger {
	l.Category = category
	return l
}

// ClearCategory clear category.
func (l *Logger) ClearCategory() *Logger {
	l.Category = ""
	return l
}

// ClearAll clear all extra fields.
func (l *Logger) ClearAll() *Logger {
	l.ClearResource()
	l.ClearCategory()
	return l
}

////////////////////////////////////////////////////////////////////////////////
// LoggerHook
////////////////////////////////////////////////////////////////////////////////

// Levels ...
func (h LoggerHook) Levels() []logrus.Level {

	levels := []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
		logrus.TraceLevel,
	}
	return levels
}

// Fire The place to modify entry.Data.
func (h LoggerHook) Fire(entry *logrus.Entry) error {
	if len(h.Logger.Resource) > 0 {
		entry.Data[RESOURCE] = h.Logger.createKeyValuePairs(h.Logger.Resource)
	}
	if len(h.Logger.Category) > 0 {
		entry.Data[CATEGORY] = h.Logger.Category
	}
	return nil
}
