package logs

import (
	"io/ioutil"
	"log"
)

func NewNoopLogger(loggerSource string) (loggers Loggers, err error) {
	loggers = &GenericLoggers{
		Output: log.New(ioutil.Discard, "", 0),
		Error:  log.New(ioutil.Discard, "", 0),
	}
	return
}
