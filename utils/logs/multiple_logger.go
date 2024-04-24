package logs

import (
	"fmt"
	"sync"

	"github.com/go-logr/logr"
	"golang.org/x/sync/errgroup"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

type MultipleLogger struct {
	mu           sync.RWMutex
	loggers      []Loggers
	loggerSource string
}

func (c *MultipleLogger) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	g := new(errgroup.Group)
	for i := range c.loggers {
		g.Go(c.loggers[i].Close)
	}
	return g.Wait()
}

func (c *MultipleLogger) Check() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	g := new(errgroup.Group)
	for i := range c.loggers {
		g.Go(c.loggers[i].Check)
	}
	return g.Wait()
}

func (c *MultipleLogger) SetLogSource(source string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	var err error
	for i := range c.loggers {
		err = c.loggers[i].SetLogSource(source)
	}
	return err
}

func (c *MultipleLogger) SetLoggerSource(source string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.setLoggerSource(source)
}

func (c *MultipleLogger) setLoggerSource(source string) error {
	var err error
	for i := range c.loggers {
		err = c.loggers[i].SetLoggerSource(source)
		if err != nil {
			return err
		}
	}

	c.loggerSource = source
	return nil
}

func (c *MultipleLogger) Log(output ...interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i := range c.loggers {
		c.loggers[i].Log(output...)
	}
}

func (c *MultipleLogger) LogError(err ...interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i := range c.loggers {
		c.loggers[i].LogError(err...)
	}
}

func (c *MultipleLogger) GetLoggerSource() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.loggerSource
}

func (c *MultipleLogger) AppendLogger(l ...logr.Logger) error {
	for i := range l {
		logger, err := NewLogrLogger(l[i], c.GetLoggerSource())
		if err != nil {
			return err
		}
		err = c.Append(logger)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *MultipleLogger) Append(l ...Loggers) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.loggers = append(c.loggers, l...)
	return nil
}

type MultipleLoggerWithLoggerSource struct {
	MultipleLogger
}

func (c *MultipleLoggerWithLoggerSource) Append(l ...Loggers) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.loggers = append(c.loggers, l...)
	return c.setLoggerSource(c.loggerSource)
}

// NewMultipleLoggers returns a logger which abstracts and internally manages a list of loggers.
// if no default loggers are provided, the logger will be set to print to the standard output.
func NewMultipleLoggers(loggerSource string, loggersList ...Loggers) (l IMultipleLoggers, err error) {
	if loggerSource == "" {
		err = commonerrors.ErrNoLoggerSource
		return
	}
	l = &MultipleLoggerWithLoggerSource{}
	err = l.SetLoggerSource(loggerSource)
	if err != nil {
		return
	}

	list := loggersList
	if len(list) == 0 {
		std, err := NewStdLogger(loggerSource)
		if err != nil {
			return nil, err
		}
		list = []Loggers{std}
	}
	err = l.Append(list...)
	return
}

// NewCombinedLoggers returns a logger which logs to a list of logger. If list is empty, it will error.
func NewCombinedLoggers(loggersList ...Loggers) (l IMultipleLoggers, err error) {
	if len(loggersList) == 0 {
		err = commonerrors.ErrNoLogger
		return
	}
	l = &MultipleLogger{}
	err = l.Append(loggersList...)
	return
}
