package safeio

import (
	"context"
	"io"

	"github.com/dolmen-go/contextio"
)

// WriteString writes a string to dst similarly to io.WriteString but with context control to stop when asked to.
func WriteString(ctx context.Context, dst io.Writer, s string) (n int, err error) {
	n, err = io.WriteString(contextio.NewWriter(ctx, dst), s)
	err = convertIOError(err)
	return
}
