package logs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"iter"
	"sync"
	"time"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
)

type FIFOWriter struct {
	io.WriteCloser
	source string
	mu     sync.RWMutex
	Logs   bytes.Buffer
}

func (w *FIFOWriter) SetSource(source string) (err error) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	w.source = source
	return
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

func (w *FIFOWriter) ReadLines(ctx context.Context) iter.Seq[string] {
	return func(yield func(string) bool) {
		var partial []byte
		for {
			if err := parallelisation.DetermineContextError(ctx); err != nil {
				return
			}

			buf := func() []byte {
				w.mu.Lock()
				defer w.mu.Unlock()
				defer w.Logs.Reset()
				tmp := w.Logs.Bytes()
				buf := make([]byte, len(tmp))
				copy(buf, tmp)
				return buf
			}()

			if len(buf) == 0 {
				if err := parallelisation.DetermineContextError(ctx); err != nil {
					if len(partial) > 0 {
						yield(string(partial))
					}
					return
				}

				parallelisation.SleepWithContext(ctx, 50*time.Millisecond)
				continue
			}

			if len(partial) > 0 {
				buf = append(partial, buf...)
				partial = nil
			}

			for {
				idx := bytes.IndexByte(buf, '\n')
				if idx < 0 {
					break
				}
				line := buf[:idx]

				if len(line) > 0 && line[len(line)-1] == '\r' {
					line = line[:len(line)-1]
				}
				buf = buf[idx+1:]
				if len(line) == 0 {
					continue
				}

				if !yield(string(line)) {
					return
				}
			}

			if len(buf) > 0 {
				partial = buf
			}
		}
	}
}

type FIFOLoggers struct {
	Output    WriterWithSource
	Error     WriterWithSource
	LogWriter *FIFOWriter
	newline   bool
}

func (l *FIFOLoggers) SetLogSource(source string) error {
	err := l.Check()
	if err != nil {
		return err
	}
	return l.Output.SetSource(source)
}

func (l *FIFOLoggers) SetLoggerSource(source string) error {
	err := l.Check()
	if err != nil {
		return err
	}
	return l.Output.SetSource(source)
}

func (l *FIFOLoggers) Log(output ...interface{}) {
	_, _ = l.Output.Write([]byte(fmt.Sprint(output...)))
	if l.newline {
		_, _ = l.Output.Write([]byte("\n"))
	}
}

func (l *FIFOLoggers) LogError(err ...interface{}) {
	_, _ = l.Error.Write([]byte(fmt.Sprint(err...)))
	if l.newline {
		_, _ = l.Output.Write([]byte("\n"))
	}
}

func (l *FIFOLoggers) Check() error {
	if l.Error == nil || l.Output == nil {
		return commonerrors.ErrNoLogger
	}
	return nil
}

func (l *FIFOLoggers) Read() string {
	return l.LogWriter.Read()
}

func (l *FIFOLoggers) ReadLines(ctx context.Context) iter.Seq[string] {
	return l.LogWriter.ReadLines(ctx)
}

// Close closes the logger
func (l *FIFOLoggers) Close() (err error) {
	return l.LogWriter.Close()
}

// NewFIFOLogger creates a logger to a bytes buffer.
// All messages (whether they are output or error) are merged together.
// Once messages have been accessed they are gone
func NewFIFOLogger() (loggers *FIFOLoggers, err error) {
	l, err := NewNoopLogger("Noop Logger")
	if err != nil {
		return
	}
	logWriter := &FIFOWriter{}

	loggers = &FIFOLoggers{
		LogWriter: logWriter,
		newline:   true,
		Output:    NewDiodeWriterForSlowWriter(logWriter, 10000, 50*time.Millisecond, l),
		Error:     NewDiodeWriterForSlowWriter(logWriter, 10000, 50*time.Millisecond, l),
	}
	return
}

// NewPlainFIFOLogger creates a logger to a bytes buffer with no extra flag, prefix or tag, just the logged text.
// All messages (whether they are output or error) are merged together.
// Once messages have been accessed they are gone
func NewPlainFIFOLogger() (loggers *FIFOLoggers, err error) {
	l, err := NewNoopLogger("Noop Logger")
	if err != nil {
		return
	}

	logWriter := &FIFOWriter{}
	loggers = &FIFOLoggers{
		LogWriter: logWriter,
		Output:    NewDiodeWriterForSlowWriter(logWriter, 10000, 50*time.Millisecond, l),
		Error:     NewDiodeWriterForSlowWriter(logWriter, 10000, 50*time.Millisecond, l),
	}
	return
}
