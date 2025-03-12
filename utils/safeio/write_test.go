package safeio

import (
	"bytes"
	"context"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
)

func TestWriteString(t *testing.T) {
	var buf bytes.Buffer
	text := faker.Sentence()
	n, err := WriteString(context.Background(), &buf, text)
	require.NoError(t, err)
	assert.Equal(t, len(text), n)
	assert.Equal(t, text, buf.String())
	buf.Reset()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cancel()
	n, err = WriteString(ctx, &buf, text)
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrCancelled)
	assert.Zero(t, n)
	assert.Empty(t, buf.String())
}
