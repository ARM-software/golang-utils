package subprocess

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultIO(t *testing.T) {
	io := NewDefaultIO()
	assert.Equal(t, os.Stdin, io.SetInput(context.Background()))
	assert.Equal(t, os.Stdout, io.SetOutput(context.Background()))
	assert.Equal(t, os.Stderr, io.SetError(context.Background()))
}
