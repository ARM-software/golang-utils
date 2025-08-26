// Package safeio provides functions similar to utilities in built-in io package but with safety nets.
package safeio

import (
	"context"
	"io"

	"github.com/ARM-software/golang-utils/utils/parallelisation"
)

// CopyDataWithContext copies from src to dst similarly to io.Copy but with context control to stop when asked.
func CopyDataWithContext(ctx context.Context, src io.Reader, dst io.Writer) (copied int64, err error) {
	return copyDataWithContext(ctx, src, dst, io.Copy)
}

// CopyNWithContext copies n bytes from src to dst similarly to io.CopyN but with context control to stop when asked.
func CopyNWithContext(ctx context.Context, src io.Reader, dst io.Writer, n int64) (copied int64, err error) {
	return copyDataWithContext(ctx, src, dst, func(dst io.Writer, src io.Reader) (int64, error) { return io.CopyN(dst, src, n) })
}

// CatN concatenates n bytes from multiple sources to dst. It is intended to provide functionality quite similar to `cat` posix command but with context control.
func CatN(ctx context.Context, dst io.Writer, n int64, src ...io.Reader) (copied int64, err error) {
	return CopyNWithContext(ctx, NewContextualMultipleReader(ctx, src...), dst, n)
}

// Cat concatenates bytes from multiple sources to dst. It is intended to provide functionality quite similar to `cat` posix command but with context control.
func Cat(ctx context.Context, dst io.Writer, src ...io.Reader) (copied int64, err error) {
	return CopyDataWithContext(ctx, NewContextualMultipleReader(ctx, src...), dst)
}

func copyDataWithContext(ctx context.Context, src io.Reader, dst io.Writer, copyFunc func(io.Writer, io.Reader) (int64, error)) (copied int64, err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	copied, err = safeCopy(ContextualWriter(ctx, dst), NewContextualReader(ctx, src), copyFunc)
	return
}

func safeCopy(w io.Writer, r io.Reader, iocopyFunc func(io.Writer, io.Reader) (int64, error)) (int64, error) {
	copied, err := iocopyFunc(w, r)
	err = ConvertIOError(err)
	return copied, err
}
