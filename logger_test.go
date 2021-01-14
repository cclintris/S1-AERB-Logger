package logger_test

import (
	"bytes"
	"os"
	"strconv"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	gojsonq "github.com/thedevsaddam/gojsonq/v2"

	s1logger "gitlab-smartgaia.sercomm.com/s1util/logger"
)

const (
	Resource string = "D:123456"
	Category string = "MyCategory"
)

var buf bytes.Buffer

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags

	// Initialize s1 logger.
	os.Setenv("LOG_LEVEL", strconv.Itoa(int(logrus.InfoLevel)))
	s1logger.NewWithOptions(s1logger.OPT_DEAFULT).SetOutput(&buf)

	os.Exit(m.Run())
}

func TestLogger(t *testing.T) {

	logger := s1logger.New()
	logger.SetResource(Resource).SetCategory(Category)

	buf.Reset()
	logger.Infof("log w/ resource")
	jq := gojsonq.New().JSONString(buf.String())
	// fmt.Println(jq.String())
	assert.NotEmpty(t, jq.Reset().Find(s1logger.FILE))
	assert.NotEmpty(t, jq.Reset().Find(s1logger.FUNCTION))
	assert.Equal(t, jq.Reset().Find(s1logger.RESOURCE), Resource)
	assert.Equal(t, jq.Reset().Find(s1logger.CATEGORY), Category)

	subfun(t)

	// go routine
	done := make(chan struct{})
	go goroutine(t, done)
	<-done

	logger.ClearAll()

	buf.Reset()
	logger.Infof("log w/o resource")
	jq = gojsonq.New().JSONString(buf.String())
	// fmt.Println(jq.String())
	assert.NotEmpty(t, jq.Reset().Find(s1logger.FILE))
	assert.NotEmpty(t, jq.Reset().Find(s1logger.FUNCTION))
	assert.Nil(t, jq.Reset().Find(s1logger.RESOURCE))
	assert.Nil(t, jq.Reset().Find(s1logger.CATEGORY))
}

func subfun(t *testing.T) {
	logger := s1logger.New()

	buf.Reset()
	logger.Infof("log  w/ resource in subfun")
	jq := gojsonq.New().JSONString(buf.String())
	// fmt.Println(jq.String())
	assert.NotEmpty(t, jq.Reset().Find(s1logger.FILE))
	assert.NotEmpty(t, jq.Reset().Find(s1logger.FUNCTION))
	assert.Equal(t, jq.Reset().Find(s1logger.RESOURCE), Resource)
	assert.Equal(t, jq.Reset().Find(s1logger.CATEGORY), Category)
}

func goroutine(t *testing.T, done chan<- struct{}) {
	logger := s1logger.New()

	defer close(done)

	buf.Reset()
	logger.Infof("log  w/ resource in goroutine")
	jq := gojsonq.New().JSONString(buf.String())
	// fmt.Println(jq.String())
	assert.NotEmpty(t, jq.Reset().Find(s1logger.FILE))
	assert.NotEmpty(t, jq.Reset().Find(s1logger.FUNCTION))
	assert.Equal(t, jq.Reset().Find(s1logger.RESOURCE), Resource)
	assert.Equal(t, jq.Reset().Find(s1logger.CATEGORY), Category)
}
