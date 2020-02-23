package logger

import (
	"github.com/sirupsen/logrus"
	"math/rand"
	"os"
	"time"
)

// New is a function to obtain a singleton instance of Logger.
func New() *logrus.Logger {
	instance := logrus.New()

	// Set log format.
	instance.SetFormatter(&logrus.JSONFormatter{})

	// Set log level.
	l, err := logrus.ParseLevel(os.Getenv("LOG_LEVEL"))
	if nil != err {
		l = logrus.InfoLevel
	}
	instance.SetLevel(l)

	/* TODO: use goroutine id or thread id is better.
	// Generate logId
	logId := GenerateRunId() //
	instance.WithFields(logrus.Fields{"logId": logId})
	*/

	return instance
}

func GenerateRunId() string {
	rand.Seed(time.Now().UnixNano())
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, 10)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
