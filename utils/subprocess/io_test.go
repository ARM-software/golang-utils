package subprocess

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultIO(t *testing.T) {
	io := NewDefaultIO()
	in, out, errs := io.Register(context.Background())
	assert.Equal(t, os.Stdin, in)
	assert.Equal(t, os.Stdout, out)
	assert.Equal(t, os.Stderr, errs)
}
