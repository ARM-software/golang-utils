package logs

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"sync"
)

type FIFOWriter struct {
	io.WriteCloser
	mu   sync.RWMutex
	Logs bytes.Buffer
}

func (w *FIFOWriter) Write(p []byte) (n int, err error) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	w.Logs.Write(p)
	return
}

func (w *FIFOWriter) Close() (err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.Logs.Reset()
	return
}

func (w *FIFOWriter) Read() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	n := w.Logs.Len()
	if n == 0 {
		return ""
	}
	bytes := w.Logs.Next(n)
	return string(bytes)
}

type FIFOLoggers struct {
	GenericLoggers
	LogWriter FIFOWriter
}

func (l *FIFOLoggers) Check() error {
	return l.GenericLoggers.Check()
}

func (l *FIFOLoggers) Read() string {
	return l.LogWriter.Read()
}

// Close closes the logger
func (l *FIFOLoggers) Close() (err error) {
	err = l.LogWriter.Close()
	if err != nil {
		return
	}
	err = l.GenericLoggers.Close()
	return
}

// NewFIFOLogger creates a logger to a bytes buffer.
// All messages (whether they are output or error) are merged together.
// Once messages have been accessed they are gone
func NewFIFOLogger(loggerSource string) (loggers *FIFOLoggers, err error) {
	loggers = &FIFOLoggers{
		LogWriter: FIFOWriter{},
	}
	loggers.GenericLoggers = GenericLoggers{
		Output: log.New(&loggers.LogWriter, fmt.Sprintf("[%v] Output: ", loggerSource), log.LstdFlags),
		Error:  log.New(&loggers.LogWriter, fmt.Sprintf("[%v] Error: ", loggerSource), log.LstdFlags),
	}
	return
}

// NewPlainFIFOLogger creates a logger to a bytes buffer with no extra flag, prefix or tag, just the logged text.
// All messages (whether they are output or error) are merged together.
// Once messages have been accessed they are gone
func NewPlainFIFOLogger() (loggers *FIFOLoggers, err error) {
	loggers = &FIFOLoggers{
		LogWriter: FIFOWriter{},
	}
	loggers.GenericLoggers = GenericLoggers{
		Output: log.New(&loggers.LogWriter, "", 0),
		Error:  log.New(&loggers.LogWriter, "", 0),
	}
	return
}
