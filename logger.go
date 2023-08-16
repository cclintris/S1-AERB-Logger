package logger

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	. "gitlab-smartgaia.sercomm.com/s1util/logger/buffer"
	. "gitlab-smartgaia.sercomm.com/s1util/logger/buffer/util"
)

// LogOptions ...
type LogOptions uint16

const (
	OPT_ALL_ENABLED       LogOptions = 0xFFFF
	OPT_ALL_DISABLED      LogOptions = 0x0000
	OPT_DEFAULT           LogOptions = 0x0003
	OPT_HAS_REPORT_CALLER LogOptions = 0x0001
	OPT_HAS_SHORT_CALLER  LogOptions = 0x0002

	FILE     string = "file"
	FUNCTION string = "func"
	RESOURCE string = "res"
	CATEGORY string = "cat"

	BUFFER_MODE string = "BUFFER_MODE"
	PLAIN_MODE  string = "PLAIN_MODE"
)

// Logger struct
type Logger struct {
	logrus.Logger

	Options LogOptions
	/*
		Resources       map[string]string
		ResourcesString string
	*/
	Resources *Resources
	Category  string
	Buffer    *RingBuffer
	Mode      string
}

// Log struct
type Log struct {
	File     string       `json:"file"`
	Function string       `json:"func"`
	Level    logrus.Level `json:"level"`
	Message  string       `json:"msg"`
	Resource interface{}  `json:"res"`
	Time     time.Time    `json:"time"`
}

type Resources struct {
	typeMap    map[string]string // resource type map
	printedStr string
}

// LoggerHook ...
type LoggerHook struct {
	logrus.Hook

	Logger *Logger
}

// LoggerHookBuffer ...
type LoggerHookBuffer struct {
	logrus.Hook

	Logger *Logger
}

// LoggerHookFlush ...
type LoggerHookFlush struct {
	logrus.Hook

	Logger *Logger
}

// LoggerHookPlain ...
type LoggerHookPlain struct {
	logrus.Hook

	Logger *Logger
}

// Implement Singleton pattern
var (
	once    sync.Once
	logger  *Logger
	defOpts LogOptions = OPT_DEFAULT
)

// Hookname String
var (
	LOGGER_HOOK        = "logger.LoggerHook"
	LOGGER_HOOK_BUFFER = "logger.LoggerHookBuffer"
	LOGGER_HOOK_FLUSH  = "logger.LoggerHookFlush"
	LOGGER_HOOK_PLAIN  = "logger.LoggerHookPlain"
)

// New is a function to obtain a singleton instance of Logger.
func New() *Logger {
	return new(OPT_DEFAULT)
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

	// Set initial hooks.
	_logger.Hooks.Add(LoggerHook{Logger: _logger})
	_logger.Hooks.Add(LoggerHookBuffer{Logger: _logger})
	_logger.Hooks.Add(LoggerHookFlush{Logger: _logger})
	_logger.Hooks.Add(LoggerHookPlain{Logger: _logger})

	if _logger.Options&OPT_HAS_REPORT_CALLER > 0 {
		_logger.SetReportCaller(true)
	}

	// Set log level.
	// Set to Debug strictly. All behavior will be controlled by hooks instead of third party specification.
	_logger.SetLevel(logrus.DebugLevel)

	// initialize resource
	_logger.Resources = &Resources{}
	_logger.Resources.Clear()

	/* TODO: use goroutine id or thread id is better.
	// Generate logId
	logId := GenerateRunId() //
	instance.WithFields(logrus.Fields{"logId": logId})
	*/

	// initialize buffer
	_logger.Buffer = &RingBuffer{}

	dbs, err := ParseUnit(os.Getenv("DEFAULT_BUFFER_SIZE"))
	if err != nil {
		dbs, _ = ParseUnit("1 MB")
	}

	mbs, err := ParseUnit(os.Getenv("MAXIMUM_BUFFER_SIZE"))
	if err != nil {
		mbs, _ = ParseUnit("5 MB")
	}

	extCoef, err := ParseUnit((os.Getenv("EXTEND_COEFFICIENT")))
	if err != nil {
		extCoef, _ = ParseUnit("2 MB")
	}

	_logger.Buffer.Init(dbs, mbs, extCoef)

	// set initial logger mode
	if _logger.Mode != BUFFER_MODE {
		_logger.Mode = BUFFER_MODE
	}

	// disable logrus ability by default
	_logger.disable()

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

// //////////////////////////////////////////////////////////////////////////////
// Resources
// //////////////////////////////////////////////////////////////////////////////
// Clear clear all resource types.
func (r *Resources) Clear() *Resources {
	r.typeMap = make(map[string]string)
	r.printedStr = ""
	return r
}

// parseResource parse resource
func (r *Resources) parseResource(resource string) (string, string) {
	idx := strings.Index(resource, ":")
	if idx == -1 {
		return "X", resource // X: means unknown resource type.
	} else {
		return string(resource[:idx]), string(resource[idx+1:])
	}
}

// createKeyValuePairs convert map to string and separated by ','.
func (r *Resources) createKeyValuePairs(m map[string]string) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	b := &bytes.Buffer{}
	for _, k := range keys {
		if b.Len() > 0 {
			fmt.Fprintf(b, ", ")
		}
		fmt.Fprintf(b, "%s:%s", k, m[k])
	}
	return b.String()
}

// Set set resource type.
func (r *Resources) Set(resource string) *Resources {
	t, id := r.parseResource(resource)
	r.typeMap[t] = id
	r.printedStr = r.createKeyValuePairs(r.typeMap)
	return r
}

// Unset unset specific resource type.
func (r *Resources) Unset(resourceType string) *Resources {
	delete(r.typeMap, resourceType)
	r.printedStr = r.createKeyValuePairs(r.typeMap)
	return r
}

func (r *Resources) String() string {
	return r.printedStr
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

// SetResource set resource.
func (l *Logger) SetResource(resource string) *Logger {
	l.Resources.Set(resource)
	return l
}

// UnsetResource set resource.
func (l *Logger) UnsetResource(resourceType string) *Logger {
	l.Resources.Unset(resourceType)
	return l
}

// ClearResource clear resource.
func (l *Logger) ClearResource() *Logger {
	l.Resources.Clear()
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

// ClearAll clear all extra fields and clears buffered logs.
func (l *Logger) ClearAll() *Logger {
	l.ClearResource()
	l.ClearCategory()
	l.Mode = BUFFER_MODE
	l.Buffer.Reset()
	l.recover()
	return l
}

// Disable logrus.
func (l *Logger) disable() {
	if l.Out == io.Discard {
		return
	}
	l.SetOutput(io.Discard)
}

// Restore logrus.
func (l *Logger) recover() {
	if l.Out == os.Stderr {
		return
	}
	l.SetOutput(os.Stderr)
}

// Wrap and construct ringlog given logrus entry
func (l *Logger) logWrapper(entry *logrus.Entry) *Log {
	function, file := l.callerPrettyfier(entry.Caller)

	return &Log{
		Message:  entry.Message,
		Level:    entry.Level,
		Time:     entry.Time,
		Function: function,
		File:     file,
		Resource: entry.Data[RESOURCE],
	}
}

// Replace logrus hooks
func (l *Logger) updateHooks(removeHooks []string, newHooks []logrus.Hook) {
	updatedHooks := make(logrus.LevelHooks)

	for level, hooks := range l.Logger.Hooks {
		for _, h := range hooks {
			hookName := fmt.Sprintf("%T", h)
			if !matchHookName(hookName, removeHooks) {
				updatedHooks[level] = append(updatedHooks[level], h)
			}
		}
	}

	l.ReplaceHooks(updatedHooks)

	for _, newHook := range newHooks {
		l.AddHook(newHook)
	}
}

// Check if a hook matches a given hookname
func matchHookName(hookName string, matcher []string) bool {
	for _, h := range matcher {
		if h == hookName {
			return true
		}
	}
	return false
}

// Restore logrus hooks
func (l *Logger) restoreHooks(hooks logrus.LevelHooks) {
	_ = l.ReplaceHooks(hooks)
}

////////////////////////////////////////////////////////////////////////////////
// LoggerHook
////////////////////////////////////////////////////////////////////////////////

// Levels for LoggerHook ...
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

// Fire to modify entry.Data.
func (h LoggerHook) Fire(entry *logrus.Entry) error {
	// fmt.Println("[logrus hook]: enter LoggerHook")

	if len(h.Logger.Resources.String()) > 0 {
		entry.Data[RESOURCE] = h.Logger.Resources.String()
	}
	if len(h.Logger.Category) > 0 {
		entry.Data[CATEGORY] = h.Logger.Category
	}
	return nil
}

// Levels for LoggerHookBuffer ...
func (hBuffer LoggerHookBuffer) Levels() []logrus.Level {

	levels := []logrus.Level{
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
		logrus.TraceLevel,
	}
	return levels
}

// Fire to buffer logs
func (hBuffer LoggerHookBuffer) Fire(entry *logrus.Entry) error {

	if hBuffer.Logger.Mode != BUFFER_MODE {
		return nil
	}

	// fmt.Println("[logrus hook]: enter LoggerHookBuffer")

	// buffer logs
	log := hBuffer.Logger.logWrapper(entry)
	jLog, err := json.Marshal(log)
	if err != nil {
		return err
	}

	// buffer length of log to be written in little endian
	sLog := string(jLog)
	l := len(sLog)
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(l))

	n, err := hBuffer.Logger.Buffer.Write(buf)
	if n != 4 || err != nil {
		return err
	}

	// buffer actual log
	n, err = hBuffer.Logger.Buffer.Write([]byte(sLog))
	if n != l || err != nil {
		return err
	}

	return nil
}

// Levels for LoggerHookFlush ...
func (hFlush LoggerHookFlush) Levels() []logrus.Level {

	levels := []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
	}
	return levels
}

// Fire to flush out logs
func (hFlush LoggerHookFlush) Fire(entry *logrus.Entry) error {

	if hFlush.Logger.Mode != BUFFER_MODE {
		return nil
	}

	// fmt.Println("[logrus hook]: enter LoggerHookFlush")

	// flush all logs from buffer
	for !hFlush.Logger.Buffer.IsEmpty() {
		buf := make([]byte, 4)
		n, err := hFlush.Logger.Buffer.Read(buf)
		if n != 4 || err != nil {
			return err
		}

		l := int(binary.LittleEndian.Uint32(buf))
		buf = make([]byte, l)
		n, err = hFlush.Logger.Buffer.Read(buf)
		if n != l || err != nil {
			return err
		}

		stdLog := bytes.NewBuffer(buf).String()
		fmt.Println(stdLog)
	}

	hFlush.Logger.Mode = PLAIN_MODE

	return nil
}

// Levels for LoggerHookPlain ...
func (hPlain LoggerHookPlain) Levels() []logrus.Level {

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

// Fire to console log to standard output
func (hPlain LoggerHookPlain) Fire(entry *logrus.Entry) error {

	if hPlain.Logger.Mode != PLAIN_MODE {
		return nil
	}

	// fmt.Println("[logrus hook]: enter LoggerHookPlain")

	log := hPlain.Logger.logWrapper(entry)
	jLog, err := json.Marshal(log)
	if err != nil {
		return err
	}
	fmt.Println(string(jLog))
	return nil
}
