package logger_test

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	s1logger "gitlab-smartgaia.sercomm.com/s1util/logger"
	. "gitlab-smartgaia.sercomm.com/s1util/logger/buffer"
	. "gitlab-smartgaia.sercomm.com/s1util/logger/buffer/constant"
)

const (
	DeviceResource string = "D:3C62F006E1D1-2110DMM000018"
	UserResource   string = "U:001b1607-ca91-4929-8287-ac9eb1aca221"
	RegionResource string = "R:85abbc66-0cfe-47bd-a4eb-d6feca985567"
	Category       string = "MyCategory"
)

var (
	logLevel          string = "error"
	logger            s1logger.Logger
	expectedCategory  string = ""
	expectedResources string = ""
)

func makeMsg(logLevel string) string {
	return fmt.Sprintf("log w/ resource, %s level", logLevel)
}

func setup() {
	fmt.Println("[logger_test]: enter setup")

	// Set environment variables
	os.Setenv("LOG_LEVEL", logLevel)
	os.Setenv("DEFAULT_BUFFER_SIZE", "1 KB")
	os.Setenv("MAXIMUM_BUFFER_SIZE", "2 KB")
	os.Setenv("EXTEND_COEFFICIENT", "1 KB")

	// Initialize s1 logger
	logger = *s1logger.New()
	// logger.SetLevel(logrus.DebugLevel)
	logger.SetResource(RegionResource).SetResource(UserResource).SetResource(DeviceResource).SetCategory(Category)
	logger.UnsetResource("R")

	// Expected value setup
	expectedResources = fmt.Sprintf("%s, %s", DeviceResource, UserResource)
	expectedCategory = Category

	fmt.Println("[logger_test]: leave setup")
}

func teardown() {
	fmt.Println("[logger_test]: enter teardown")

	// Clean up s1 logger
	logger.ClearAll()

	fmt.Println("[logger_test]: leave teardown")
}

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags

	// setup
	setup()

	// execute
	retCode := m.Run()

	// teardown
	teardown()

	os.Exit(retCode)
}

func TestNew(t *testing.T) {
	buf := logger.Buffer

	assert.False(t, buf.IsFull())
	assert.True(t, buf.IsEmpty())
	assert.Equal(t, int(math.Round(KB)), buf.Capacity())
	assert.Equal(t, 0, buf.Length())
	assert.Equal(t, 0, buf.VirtualLength())

	data := make([]byte, 4)
	n, err := buf.Read(data)
	assert.Equal(t, 0, n)
	assert.Error(t, ErrIsEmpty, err)
}

func TestTrace(t *testing.T) {

	// test TRACE level
	msg := makeMsg("TRACE")
	logger.Trace(msg)

	buf := logger.Buffer

	assert.False(t, buf.IsFull())
	assert.True(t, buf.IsEmpty())
	assert.Equal(t, int(math.Round(KB)), buf.Capacity())
	assert.Equal(t, 0, buf.Length())
	assert.Equal(t, 0, buf.VirtualLength())
}

func TestDebug(t *testing.T) {

	// test TRACE level
	msg := makeMsg("DEBUG")
	logger.Debug(msg)

	buf := logger.Buffer

	assert.False(t, buf.IsFull())
	assert.False(t, buf.IsEmpty())
	assert.Equal(t, int(math.Round(KB)), buf.Capacity())
	assert.Equal(t, 224, buf.Length())
	assert.Equal(t, 224, buf.VirtualLength())
}

func TestFatal(t *testing.T) {

	// test Fatal level
	msgFatal := makeMsg("FATAL")

	logger.Fatal(msgFatal)

	buf := logger.Buffer

	assert.True(t, buf.IsEmpty())
	assert.False(t, buf.IsFull())
}

func TestError(t *testing.T) {

	// test ERROR level
	msgError1 := makeMsg("ERROR1")
	msgDebug1 := makeMsg("DEBUG1")
	msgDebug2 := makeMsg("DEBUG2")
	msgDebug3 := makeMsg("DEBUG3")

	buf := logger.Buffer

	assert.True(t, buf.IsEmpty())

	logger.Debug(msgDebug1)
	assert.Equal(t, 225, buf.Length())
	assert.Equal(t, 225, buf.VirtualLength())

	logger.Debug(msgDebug2)
	assert.Equal(t, 450, buf.Length())
	assert.Equal(t, 450, buf.VirtualLength())

	logger.Debug(msgDebug3)
	assert.Equal(t, 675, buf.Length())
	assert.Equal(t, 675, buf.VirtualLength())

	assert.False(t, buf.IsEmpty())

	logger.Error(msgError1)

	assert.True(t, buf.IsEmpty())
	assert.Equal(t, 0, buf.Length())
	assert.Equal(t, 0, buf.VirtualLength())
}

func TestMultiError(t *testing.T) {

	// test multi ERROR level
	msgError1 := makeMsg("ERROR1")
	msgError2 := makeMsg("ERROR2")
	msgDebug1 := makeMsg("DEBUG1")
	msgDebug2 := makeMsg("DEBUG2")
	msgDebug3 := makeMsg("DEBUG3")
	msgInfo1 := makeMsg("INFO1")
	msgInfo2 := makeMsg("INFO2")

	buf := logger.Buffer

	assert.True(t, buf.IsEmpty())

	logger.Debug(msgDebug1)

	logger.Debug(msgDebug2)

	logger.Debug(msgDebug3)

	assert.False(t, buf.IsEmpty())

	logger.Error(msgError1)

	assert.True(t, buf.IsEmpty())

	logger.Info(msgInfo1)

	logger.Info(msgInfo2)

	assert.True(t, buf.IsEmpty())

	logger.Error(msgError2)

	assert.True(t, buf.IsEmpty())
}

func TestEmptyError(t *testing.T) {

	// test ERROR behavior when buffer is empty
	msgError := makeMsg("ERROR")

	buf := logger.Buffer

	assert.True(t, buf.IsEmpty())

	logger.Error(msgError)

	assert.True(t, buf.IsEmpty())
}

func TestFull_OverFlowBuffer(t *testing.T) {

	// test ERROR behavior when buffer is full
	buf := logger.Buffer

	for i := 1; i <= 20; i++ {
		msgDebug := makeMsg(strconv.Itoa(i))
		logger.Debug(msgDebug)
	}

	msgErr := makeMsg("ERROR")
	logger.Error(msgErr)

	assert.True(t, buf.IsEmpty())

	logger.Info("After Error")
}
