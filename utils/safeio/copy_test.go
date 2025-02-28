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

func TestCopyDataWithContext(t *testing.T) {
	var buf1, buf2 bytes.Buffer
	text := faker.Sentence()
	n, err := WriteString(context.Background(), &buf1, text)
	require.NoError(t, err)
	require.NotZero(t, n)
	assert.Equal(t, len(text), n)
	n2, err := CopyDataWithContext(context.Background(), &buf1, &buf2)
	require.NoError(t, err)
	require.NotZero(t, n2)
	assert.Equal(t, int64(len(text)), n2)
	assert.Equal(t, text, buf2.String())

	ctx, cancel := context.WithCancel(context.Background())
	buf1.Reset()
	buf2.Reset()
	n, err = WriteString(context.Background(), &buf1, text)
	require.NoError(t, err)
	require.NotZero(t, n)
	assert.Equal(t, len(text), n)

	cancel()
	n2, err = CopyDataWithContext(ctx, &buf1, &buf2)
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrCancelled)
	assert.Zero(t, n2)
	assert.Empty(t, buf2.String())
}

func TestCopyNWithContext(t *testing.T) {
	var buf1, buf2 bytes.Buffer
	text := faker.Sentence()
	n, err := WriteString(context.Background(), &buf1, text)
	require.NoError(t, err)
	require.NotZero(t, n)
	assert.Equal(t, len(text), n)
	n2, err := CopyNWithContext(context.Background(), &buf1, &buf2, int64(len(text)))
	require.NoError(t, err)
	require.NotZero(t, n2)
	assert.Equal(t, int64(len(text)), n2)
	assert.Equal(t, text, buf2.String())

	ctx, cancel := context.WithCancel(context.Background())
	buf1.Reset()
	buf2.Reset()
	n, err = WriteString(context.Background(), &buf1, text)
	require.NoError(t, err)
	require.NotZero(t, n)
	assert.Equal(t, len(text), n)

	cancel()
	n2, err = CopyNWithContext(ctx, &buf1, &buf2, int64(len(text)))
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrCancelled)
	assert.Zero(t, n2)
	assert.Empty(t, buf2.String())

	n, err = WriteString(context.Background(), &buf1, text)
	require.NoError(t, err)
	require.NotZero(t, n)
	assert.Equal(t, len(text), n)
	n2, err = CopyNWithContext(context.Background(), &buf1, &buf2, int64(len(text)-1))
	require.NoError(t, err)
	require.NotZero(t, n2)
	assert.Equal(t, int64(len(text)-1), n2)
}
