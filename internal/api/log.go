package api

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

var apiLog *logrus.Logger

func SetLogger(logger *logrus.Logger) {
	if logger == nil {
		panic(fmt.Errorf("logger must not be nil"))
	}
	apiLog = logger
}
