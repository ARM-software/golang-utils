package logs

import (
	"bytes"
	"context"
	"fmt"
	"iter"
	"strings"
	"time"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/diodes"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
)

const (
	newLine    = '\n'
	bufferSize = 10000
)

type loggerAlerter struct {
	log Loggers
}

func (l *loggerAlerter) Alert(missed int) {
	if l.log != nil {
		l.log.LogError(fmt.Sprintf("Logger dropped %d messages", missed))
	}
}

func newLoggerAlerter(logs Loggers) diodes.Alerter {
	return &loggerAlerter{log: logs}
}

func newFIFODiode(ctx context.Context, ringBufferSize int, pollingPeriod time.Duration, droppedMessagesLogger Loggers) *fifoDiode {
	dCtx, cancel := context.WithCancel(ctx)
	cancelStore := parallelisation.NewCancelFunctionsStore()
	cancelStore.RegisterCancelFunction(cancel)
	return &fifoDiode{
		d:           diodes.NewPoller(diodes.NewManyToOne(ringBufferSize, newLoggerAlerter(droppedMessagesLogger)), diodes.WithPollingInterval(pollingPeriod), diodes.WithPollingContext(dCtx)),
		cancelStore: cancelStore,
	}
}

type fifoDiode struct {
	d           *diodes.Poller
	cancelStore *parallelisation.CancelFunctionStore
}

func (d *fifoDiode) Set(data []byte) {
	d.d.Set(diodes.GenericDataType(&data))
}

func (d *fifoDiode) Close() error {
	d.cancelStore.Cancel()
	return nil
}

// LineIterator returns an iterator over lines. It should only be called within the context of the same goroutine.
func (d *fifoDiode) LineIterator(ctx context.Context) iter.Seq[string] {
	return func(yield func(string) bool) {
		err := IterateOverLines(ctx, func(fCtx context.Context) (b []byte, err error) {
			err = parallelisation.DetermineContextError(fCtx)
			if err != nil {
				return
			}
			data, has := d.d.TryNext()
			if has {
				b = *(*[]byte)(data)
				return
			}
			if d.d.IsDone() {
				err = commonerrors.ErrEOF
				return
			}
			return
		}, yield)
		if err != nil {
			return
		}
	}
}

func cleanseLine(line string) string {
	return strings.TrimSuffix(strings.ReplaceAll(line, "\r", ""), string(newLine))
}

func iterateOverLines(ctx context.Context, b *bytes.Buffer, yield func(string) bool) (err error) {
	for {
		subErr := parallelisation.DetermineContextError(ctx)
		if subErr != nil {
			err = subErr
			return
		}
		line, foundErr := b.ReadString(newLine)
		if foundErr == nil {
			if !yield(line) {
				err = commonerrors.ErrEOF
				return
			}
		} else {
			b.Reset()
			_, subErr = b.Write([]byte(line))
			if subErr != nil {
				err = subErr
				return
			}
			return
		}
	}
}

func IterateOverLines(ctx context.Context, fetchNext func(fCtx context.Context) ([]byte, error), yield func(string) bool) (err error) {
	extendedYield := func(s string) bool {
		return yield(cleanseLine(s))
	}
	b := bytes.NewBuffer(make([]byte, 0, 512))
	for {
		subErr := parallelisation.DetermineContextError(ctx)
		if subErr != nil {
			err = subErr
			return
		}
		nextBuf, subErr := fetchNext(ctx)
		if subErr != nil {
			err = subErr
			return
		}
		if len(nextBuf) == 0 {
			parallelisation.SleepWithContext(ctx, 10*time.Millisecond)
			continue
		}
		_, subErr = b.Write(nextBuf)
		if subErr != nil {
			err = subErr
			return
		}
		subErr = iterateOverLines(ctx, b, extendedYield)
		if subErr != nil {
			err = subErr
			return
		}
	}
}

type FIFOLoggers struct {
	d       *fifoDiode
	newline bool
}

func (l *FIFOLoggers) SetLogSource(_ string) error {
	return nil
}

func (l *FIFOLoggers) SetLoggerSource(_ string) error {
	return nil
}

func (l *FIFOLoggers) Log(output ...any) {
	l.log(output...)
}

func (l *FIFOLoggers) LogError(err ...any) {
	l.log(err...)
}

func (l *FIFOLoggers) log(args ...any) {
	b := bytes.NewBufferString(fmt.Sprint(args...))
	if l.newline {
		_, _ = b.Write([]byte{newLine})
	}
	l.d.Set(b.Bytes())
}

func (l *FIFOLoggers) Check() error {
	if l.d == nil {
		return commonerrors.UndefinedVariable("FIFO diode")
	}
	return nil
}

// LineIterator returns an iterator over lines. It should only be called within the context of the same goroutine.
func (l *FIFOLoggers) LineIterator(ctx context.Context) iter.Seq[string] {
	return l.d.LineIterator(ctx)
}

// Close closes the logger
func (l *FIFOLoggers) Close() (err error) {
	return l.d.Close()
}

// NewFIFOLogger creates a logger to a bytes buffer.
// All messages (whether they are output or error) are merged together.
// Once messages have been accessed they are gone
func NewFIFOLogger() (loggers *FIFOLoggers, err error) {
	loggers, err = newDefaultFIFOLogger(true)
	return
}

// NewPlainFIFOLogger creates a logger to a bytes buffer with no extra flag, prefix or tag, just the logged text.
// All messages (whether they are output or error) are merged together.
// Once messages have been accessed they are gone
func NewPlainFIFOLogger() (loggers *FIFOLoggers, err error) {
	loggers, err = newDefaultFIFOLogger(false)
	return
}

func newDefaultFIFOLogger(addNewLine bool) (loggers *FIFOLoggers, err error) {
	l, err := NewNoopLogger("FIFO")
	if err != nil {
		return
	}
	return NewFIFOLoggerWithBuffer(addNewLine, bufferSize, 50*time.Millisecond, l)
}

func NewFIFOLoggerWithBuffer(addNewLine bool, ringBufferSize int, pollingPeriod time.Duration, droppedMessageLogger Loggers) (loggers *FIFOLoggers, err error) {
	loggers = &FIFOLoggers{
		d:       newFIFODiode(context.Background(), ringBufferSize, pollingPeriod, droppedMessageLogger),
		newline: addNewLine,
	}
	return
}
