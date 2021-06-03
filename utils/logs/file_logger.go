package logs

import (
	"errors"
	"fmt"
	"log"
	"os"
)

type FileLoggers struct {
	GenericLoggers
	File *os.File
}

func (l *FileLoggers) Check() (err error) {
	err = l.GenericLoggers.Check()
	if err != nil {
		return
	}
	if l.File == nil {
		err = errors.New("undefined log file")
	}
	return
}

// Closes the logger
func (l *FileLoggers) Close() (err error) {
	if l.File != nil {
		err = l.File.Close()
	}
	if err != nil {
		return
	}
	err = l.GenericLoggers.Close()
	return
}

// Creates a logger to a file
func CreateFileLogger(logFile string, loggerSource string) (loggers Loggers, err error) {
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	loggers = &FileLoggers{
		GenericLoggers: GenericLoggers{
			Output: log.New(file, fmt.Sprintf("[%v] Output: ", loggerSource), log.LstdFlags),
			Error:  log.New(file, fmt.Sprintf("[%v] Error: ", loggerSource), log.LstdFlags),
		},
		File: file,
	}
	return
}
