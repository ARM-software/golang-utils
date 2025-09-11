package subprocess

import (
	"context"
	"io"
	"os"
	"sync"

	"github.com/ARM-software/golang-utils/utils/logs"
)

//go:generate go tool mockgen -destination=../mocks/mock_$GOPACKAGE.go -package=mocks github.com/ARM-software/golang-utils/utils/$GOPACKAGE ICommandIO

// ICommandIO allows you to set the stdin, stdout, and stderr that will be used in a subprocess. A context can be injected for context aware readers and writers
type ICommandIO interface {
	// Register creates new readers and writers based on the constructor methods in the ICommandIO implementation. If the constructors are not specified then it will default to os.Stdin, os.Stdout, and os.Stderr
	Register(context.Context) (in io.Reader, out, errs io.Writer)
}

type commandIO struct {
	newInFunc    func(context.Context) io.Reader
	newOutFunc   func(context.Context) io.Writer
	newErrorFunc func(context.Context) io.Writer
	mu           sync.Mutex
}

func NewIO(
	newInFunc func(context.Context) io.Reader,
	newOutFunc func(context.Context) io.Writer,
	newErrorFunc func(context.Context) io.Writer,
) ICommandIO {
	return &commandIO{
		mu:           sync.Mutex{},
		newInFunc:    newInFunc,
		newOutFunc:   newOutFunc,
		newErrorFunc: newErrorFunc,
	}
}

func NewIOFromLoggers(loggers logs.Loggers) ICommandIO {
	return NewIO(
		nil,
		func(ctx context.Context) io.Writer { return newOutStreamer(ctx, loggers) },
		func(ctx context.Context) io.Writer { return newErrLogStreamer(ctx, loggers) },
	)
}

func NewDefaultIO() ICommandIO {
	return NewIO(nil, nil, nil)
}

func (c *commandIO) Register(ctx context.Context) (in io.Reader, out, errs io.Writer) {
	c.mu.Lock()
	defer c.mu.Unlock()
	in, out, errs = os.Stdin, os.Stdout, os.Stderr
	if c.newInFunc != nil {
		in = c.newInFunc(ctx)
	}
	if c.newOutFunc != nil {
		out = c.newOutFunc(ctx)
	}
	if c.newErrorFunc != nil {
		errs = c.newErrorFunc(ctx)
	}
	return
}
