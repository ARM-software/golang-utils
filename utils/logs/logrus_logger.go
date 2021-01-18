package logs

import (
	"fmt"
	"log"

	"github.com/sirupsen/logrus"
)

// Creates a logger to logrus logger (https://github.com/Sirupsen/logrus)
func NewLogrusLogger(logrusL *logrus.Logger, loggerSource string) (loggers Loggers, err error) {
	loggers = &GenericLoggers{
		Output: log.New(logrusL.WriterLevel(logrus.InfoLevel), fmt.Sprintf("[%v] ", loggerSource), log.LstdFlags),
		Error:  log.New(logrusL.WriterLevel(logrus.ErrorLevel), fmt.Sprintf("[%v] ", loggerSource), log.LstdFlags),
	}
	return
}
