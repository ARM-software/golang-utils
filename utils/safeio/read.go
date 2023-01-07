package safeio

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/dolmen-go/contextio"

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
		if panicErr, ok := e.(error); ok && panicErr == bytes.ErrTooLarge {
			err = fmt.Errorf("%w: %v", commonerrors.ErrTooLarge, panicErr.Error())
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
	read, err := buf.ReadFrom(contextio.NewReader(ctx, reader))
	err = convertIOError(err)
	if err != nil {
		return
	}
	if read == int64(0) {
		err = fmt.Errorf("%w: no bytes were read", commonerrors.ErrEmpty)
	}
	content = buf.Bytes()
	return
}
