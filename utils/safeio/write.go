package safeio

import (
	"context"
	"io"

	"github.com/dolmen-go/contextio"
)

// WriteString writes a string to dst similarly to io.WriteString but with context control to stop when asked to.
func WriteString(ctx context.Context, dst io.Writer, s string) (n int, err error) {
	n, err = io.WriteString(ContextualWriter(ctx, dst), s)
	err = ConvertIOError(err)
	return
}

// ContextualWriter returns a writer which is context aware.
// Context state is checked BEFORE every Write.
func ContextualWriter(ctx context.Context, writer io.Writer) io.Writer {
	return &contextualCopier{w: contextio.NewWriter(ctx, writer)}
}

type contextualCopier struct {
	w io.Writer
}

func (w *contextualCopier) Write(p []byte) (n int, err error) {
	n, err = w.w.Write(p)
	err = ConvertIOError(err)
	return
}

func (w *contextualCopier) ReadFrom(r io.Reader) (int64, error) {
	if reader, ok := w.w.(io.ReaderFrom); ok {
		return safeReadFrom(reader, r)
	}
	return safeCopy(w.w, r, io.Copy)
}
