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
	SetInput(context.Context) io.Reader
	SetOutput(context.Context) io.Writer
	SetError(context.Context) io.Writer
}

type commandIO struct {
	input        io.Reader
	output       io.Writer
	error        io.Writer
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
		func(context.Context) io.Reader { return os.Stdin },
		func(ctx context.Context) io.Writer { return newOutStreamer(ctx, loggers) },
		func(ctx context.Context) io.Writer { return newErrLogStreamer(ctx, loggers) },
	)
}

func NewDefaultIO() ICommandIO {
	return NewIO(nil, nil, nil)
}

func (c *commandIO) SetInput(ctx context.Context) io.Reader {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.input = os.Stdin
	if c.newInFunc != nil {
		c.input = c.newInFunc(ctx)
	}
	return c.input
}

func (c *commandIO) SetOutput(ctx context.Context) io.Writer {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.output = os.Stdout
	if c.newOutFunc != nil {
		c.output = c.newOutFunc(ctx)
	}
	return c.output
}

func (c *commandIO) SetError(ctx context.Context) io.Writer {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.error = os.Stderr
	if c.newErrorFunc != nil {
		c.error = c.newErrorFunc(ctx)
	}
	return c.error
}
