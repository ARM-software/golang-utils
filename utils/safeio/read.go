package safeio

import (
	"bytes"
	"context"
	"io"

	"github.com/dolmen-go/contextio"
	"go.uber.org/atomic"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
)

// ReadAll reads the whole content of src similarly to io.ReadAll but with context control to stop when asked to.
func ReadAll(ctx context.Context, src io.Reader) ([]byte, error) {
	return ReadAtMost(ctx, src, -1, -1)
}

// ReadAtMost reads the content of src and at most max bytes. It provides a functionality close to io.ReadAtLeast but with a different goal.
// if bufferCapacity is not set i.e. set to a negative value, it will be set by default to max
// if max is set to a negative value, the entirety of the reader will be read
func ReadAtMost(ctx context.Context, src io.Reader, max int64, bufferCapacity int64) (content []byte, err error) {
	if bufferCapacity < 0 {
		if max < 0 {
			bufferCapacity = bytes.MinRead
		} else {
			bufferCapacity = max
		}
	}
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}

	buf := bytes.NewBuffer(make([]byte, 0, bufferCapacity))
	// If the buffer overflows, we will get bytes.ErrTooLarge.
	// Return that as an error. Any other panic remains.
	defer func() {
		e := recover()
		if e == nil {
			return
		}
		if panicErr, ok := e.(error); ok && (panicErr == bytes.ErrTooLarge || commonerrors.Any(panicErr, commonerrors.ErrTooLarge, bytes.ErrTooLarge)) {
			err = commonerrors.WrapError(commonerrors.ErrTooLarge, panicErr, "")
		} else {
			panic(e)
		}
	}()
	var reader io.Reader
	if max >= 0 {
		reader = io.LimitReader(src, max)
	} else {
		reader = src
	}
	safeBuf := NewContextualReaderFrom(ctx, buf)
	read, err := safeBuf.ReadFrom(NewContextualReader(ctx, reader))
	err = ConvertIOError(err)
	if err != nil {
		return
	}
	if read == int64(0) {
		err = commonerrors.New(commonerrors.ErrEmpty, "no bytes were read")
	}
	content = buf.Bytes()
	return
}

// NewByteReader return a byte reader which is context aware.
func NewByteReader(ctx context.Context, someBytes []byte) io.Reader {
	return NewContextualReader(ctx, bytes.NewReader(someBytes))
}

// NewContextualReader returns a reader which is context aware.
// Context state is checked BEFORE every Read.
func NewContextualReader(ctx context.Context, reader io.Reader) io.Reader {
	return contextio.NewReader(ctx, reader)
}

type safeReadCloser struct {
	reader io.Reader // use reader to ensure idempotency since you can't call close on the reader itself, only via the wrapper
	close  parallelisation.CloseFunc
	closed *atomic.Bool
}

func (r safeReadCloser) Read(p []byte) (int, error) {
	return r.reader.Read(p)
}

func (r safeReadCloser) Close() error {
	if r.closed.Swap(true) {
		return nil
	}

	return r.close()
}

// NewContextualReadCloser returns a readcloser which is context aware.
// Context state is checked during the read and close is called if the context is cancelled
// This allows for readers that block on syscalls to be stopped via a context
func NewContextualReadCloser(ctx context.Context, reader io.ReadCloser) io.ReadCloser {
	stop := context.AfterFunc(ctx, func() { _ = reader.Close() })

	r := safeReadCloser{
		reader: contextio.NewReader(ctx, reader),
		close: func() error {
			_ = stop()
			return nil
		},
		closed: atomic.NewBool(false),
	}

	return r
}

func NewContextualMultipleReader(ctx context.Context, reader ...io.Reader) io.Reader {
	readers := make([]io.Reader, len(reader))
	for i := range reader {
		readers[i] = NewContextualReader(ctx, reader[i])
	}
	return io.MultiReader(readers...)
}

// NewContextualReaderFrom returns a io.ReaderFrom which is context aware.
// Context state is checked BEFORE every Read, Write, Copy.
func NewContextualReaderFrom(ctx context.Context, reader io.ReaderFrom) io.ReaderFrom {
	return &contextualReaderFrom{r: reader, ctx: ctx}
}

type contextualReaderFrom struct {
	r   io.ReaderFrom
	ctx context.Context
}

func (c *contextualReaderFrom) ReadFrom(r io.Reader) (n int64, err error) {
	return safeReadFrom(c.r, NewContextualReader(c.ctx, r))
}

func safeReadFrom(rr io.ReaderFrom, r io.Reader) (n int64, err error) {
	n, err = rr.ReadFrom(r)
	err = ConvertIOError(err)
	return
}
