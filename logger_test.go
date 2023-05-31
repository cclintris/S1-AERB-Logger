package logger_test

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	gojsonq "github.com/thedevsaddam/gojsonq/v2"

	"gitlab-smartgaia.sercomm.com/s1util/logger"
	s1logger "gitlab-smartgaia.sercomm.com/s1util/logger"
)

const (
	DeviceResource string = "D:3C62F006E1D1-2110DMM000018"
	UserResource   string = "U:001b1607-ca91-4929-8287-ac9eb1aca221"
	RegionResource string = "R:85abbc66-0cfe-47bd-a4eb-d6feca985567"
	Category       string = "MyCategory"
)

var buf bytes.Buffer
var logLevel string = "debug"
var expectedResources string = ""

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags

	// Initialize s1 logger.
	os.Setenv("LOG_LEVEL", logLevel)

	s1logger.New().SetOutput(&buf)
	//s1logger.NewWithOptions(s1logger.OPT_DEAFULT).SetOutput(&buf)

	os.Exit(m.Run())
}

func TestNewLogger(t *testing.T) {

	logger := s1logger.New()
	logger.SetResource(RegionResource).SetResource(UserResource).SetResource(DeviceResource).SetCategory(Category)
	logger.UnsetResource("R")
	expectedResources = fmt.Sprintf("%s, %s", DeviceResource, UserResource)

	var jq *gojsonq.JSONQ

	// test TRACE level
	testTrace(t, "log w/ resource, TRACE level")

	// test DEBUG level
	testDebug(t, "log w/ resource, DEBUG level")

	// test INFO level
	testInfo(t, "log w/ resource, INFO level")

	subfun(t)

	// go routine
	done := make(chan struct{})
	go goroutine(t, done)
	<-done

	logger.ClearAll()

	buf.Reset()
	logger.Infof("log w/o resource")
	jq = gojsonq.New().JSONString(buf.String())
	assert.NotEmpty(t, jq.Reset().Find(s1logger.FILE))
	assert.NotEmpty(t, jq.Reset().Find(s1logger.FUNCTION))
	assert.Nil(t, jq.Reset().Find(s1logger.RESOURCE))
	assert.Nil(t, jq.Reset().Find(s1logger.CATEGORY))
}

func testTrace(t *testing.T, msg string) {
	logger := s1logger.New()

	// test DEBUG level
	buf.Reset()
	logger.Trace(msg)
	jq := gojsonq.New().JSONString(buf.String())
	if l, _ := logrus.ParseLevel(logLevel); l >= logrus.TraceLevel {
		assert.NotEmpty(t, jq.Reset().Find(s1logger.FILE))
		assert.NotEmpty(t, jq.Reset().Find(s1logger.FUNCTION))
		assert.Equal(t, expectedResources, jq.Reset().Find(s1logger.RESOURCE))
		assert.Equal(t, Category, jq.Reset().Find(s1logger.CATEGORY))
	} else {
		assert.Empty(t, buf.String())
	}
}

func testDebug(t *testing.T, msg string) {
	logger := s1logger.New()

	// test DEBUG level
	buf.Reset()
	logger.Debug(msg)
	jq := gojsonq.New().JSONString(buf.String())
	if l, _ := logrus.ParseLevel(logLevel); l >= logrus.DebugLevel {
		assert.NotEmpty(t, jq.Reset().Find(s1logger.FILE))
		assert.NotEmpty(t, jq.Reset().Find(s1logger.FUNCTION))
		assert.Equal(t, expectedResources, jq.Reset().Find(s1logger.RESOURCE))
		assert.Equal(t, Category, jq.Reset().Find(s1logger.CATEGORY))
	} else {
		assert.Empty(t, buf.String())
	}
}

func testInfo(t *testing.T, msg string) {
	logger := s1logger.New()

	// test INFO level
	buf.Reset()
	logger.Info(msg)
	jq := gojsonq.New().JSONString(buf.String())
	if l, _ := logrus.ParseLevel(logLevel); l >= logrus.InfoLevel {
		assert.NotEmpty(t, jq.Reset().Find(s1logger.FILE))
		assert.NotEmpty(t, jq.Reset().Find(s1logger.FUNCTION))
		assert.Equal(t, expectedResources, jq.Reset().Find(s1logger.RESOURCE))
		assert.Equal(t, Category, jq.Reset().Find(s1logger.CATEGORY))
	} else {
		assert.Empty(t, buf.String())
	}
}

func subfun(t *testing.T) {
	logger := s1logger.New()

	buf.Reset()
	logger.Info("log  w/ resource in subfun")
	jq := gojsonq.New().JSONString(buf.String())
	if l, _ := logrus.ParseLevel(logLevel); l >= logrus.InfoLevel {
		assert.NotEmpty(t, jq.Reset().Find(s1logger.FILE))
		assert.NotEmpty(t, jq.Reset().Find(s1logger.FUNCTION))
		assert.Equal(t, expectedResources, jq.Reset().Find(s1logger.RESOURCE))
		assert.Equal(t, Category, jq.Reset().Find(s1logger.CATEGORY))
	} else {
		assert.Empty(t, buf.String())
	}
}

func goroutine(t *testing.T, done chan<- struct{}) {
	logger := s1logger.New()

	defer close(done)

	buf.Reset()
	logger.Info("log  w/ resource in goroutine")
	jq := gojsonq.New().JSONString(buf.String())
	if l, _ := logrus.ParseLevel(logLevel); l >= logrus.InfoLevel {
		assert.NotEmpty(t, jq.Reset().Find(s1logger.FILE))
		assert.NotEmpty(t, jq.Reset().Find(s1logger.FUNCTION))
		assert.Equal(t, expectedResources, jq.Reset().Find(s1logger.RESOURCE))
		assert.Equal(t, Category, jq.Reset().Find(s1logger.CATEGORY))
	} else {
		assert.Empty(t, buf.String())
	}
}

func TestNewAlwaysLogger(t *testing.T) {
	var _buf bytes.Buffer

	// _logger is always new logger and set it to INFO
	os.Setenv("LOG_LEVEL", "info")
	_logger := s1logger.NewAlways(logger.OPT_DEAFULT)
	_logger.SetOutput(&_buf)

	// logger is singleton logger.
	logger := s1logger.New()
	logger.SetResource(UserResource).SetResource(DeviceResource).SetCategory(Category)
	expectedResources = fmt.Sprintf("%s, %s", DeviceResource, UserResource)

	// test TRACE level
	testTrace(t, "log w/ resource, TRACE level")

	// test DEBUG level
	testDebug(t, "log w/ resource, DEBUG level")

	// test INFO level
	testInfo(t, "log w/ resource, INFO level")

	// test _logger w/ DEBUG level
	_buf.Reset()
	_logger.Debug("always new logger, DEBUG level")
	assert.Empty(t, _buf.String())
}

func TestResources(t *testing.T) {

	logger := s1logger.New()
	logger.SetResource(UserResource).SetResource(DeviceResource).SetCategory(Category)
	expectedResources = fmt.Sprintf("%s, %s", DeviceResource, UserResource)

	var jq *gojsonq.JSONQ

	// test TRACE level
	testTrace(t, "log w/ resource, TRACE level")

	// test DEBUG level
	testDebug(t, "log w/ resource, DEBUG level")

	// test INFO level
	testInfo(t, "log w/ resource, INFO level")

	subfun(t)

	// go routine
	done := make(chan struct{})
	go goroutine(t, done)
	<-done

	logger.ClearAll()

	buf.Reset()
	logger.Infof("log w/o resource")
	jq = gojsonq.New().JSONString(buf.String())
	assert.NotEmpty(t, jq.Reset().Find(s1logger.FILE))
	assert.NotEmpty(t, jq.Reset().Find(s1logger.FUNCTION))
	assert.Nil(t, jq.Reset().Find(s1logger.RESOURCE))
	assert.Nil(t, jq.Reset().Find(s1logger.CATEGORY))
}

func TestEmptyResources(t *testing.T) {

	logger := s1logger.New()
	logger.SetResource(DeviceResource).UnsetResource("D")
	var jq *gojsonq.JSONQ

	buf.Reset()
	logger.Infof("log w/o resource")
	jq = gojsonq.New().JSONString(buf.String())
	assert.NotEmpty(t, jq.Reset().Find(s1logger.FILE))
	assert.NotEmpty(t, jq.Reset().Find(s1logger.FUNCTION))
	assert.Nil(t, jq.Reset().Find(s1logger.RESOURCE))
	assert.Nil(t, jq.Reset().Find(s1logger.CATEGORY))
}
