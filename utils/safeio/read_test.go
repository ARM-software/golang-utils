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

func TestReadAll(t *testing.T) {
	var buf bytes.Buffer
	text := faker.Sentence()
	n, err := WriteString(context.Background(), &buf, text)
	require.NoError(t, err)
	require.NotZero(t, n)
	assert.Equal(t, len(text), n)
	assert.Equal(t, text, buf.String())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	rbytes, err := ReadAll(ctx, &buf)
	require.NoError(t, err)
	assert.NotEmpty(t, rbytes)
	assert.Equal(t, text, string(rbytes))

	buf.Reset()
	n, err = WriteString(context.Background(), &buf, text)
	require.NoError(t, err)
	require.NotZero(t, n)
	assert.Equal(t, len(text), n)

	cancel()
	rbytes, err = ReadAll(ctx, &buf)
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrCancelled)
	assert.Empty(t, rbytes)
}

func TestReadAllEmpty(t *testing.T) {
	var buf bytes.Buffer
	rbytes, err := ReadAll(context.Background(), &buf)
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrEmpty)
	assert.Empty(t, rbytes)
}

func TestReadAtMost(t *testing.T) {
	var buf bytes.Buffer
	text := faker.Sentence()
	n, err := WriteString(context.Background(), &buf, text)
	require.NoError(t, err)
	require.NotZero(t, n)
	assert.Equal(t, len(text), n)
	assert.Equal(t, text, buf.String())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	rbytes, err := ReadAtMost(ctx, &buf, int64(len(text)), -1)
	require.NoError(t, err)
	assert.NotEmpty(t, rbytes)
	assert.Equal(t, text, string(rbytes))

	buf.Reset()
	n, err = WriteString(context.Background(), &buf, text)
	require.NoError(t, err)
	require.NotZero(t, n)
	assert.Equal(t, len(text), n)

	rbytes, err = ReadAtMost(ctx, &buf, int64(len(text)-2), -1)
	require.NoError(t, err)
	assert.NotEmpty(t, rbytes)
	assert.Equal(t, len(text)-2, len(rbytes))

	buf.Reset()
	n, err = WriteString(context.Background(), &buf, text)
	require.NoError(t, err)
	require.NotZero(t, n)
	assert.Equal(t, len(text), n)

	cancel()
	rbytes, err = ReadAtMost(ctx, &buf, int64(len(text)), -1)
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrCancelled)
	assert.Empty(t, rbytes)
}

func TestNewByteReader(t *testing.T) {
	text := faker.Sentence()
	ctx, cancel := context.WithCancel(context.TODO())
	result, err := ReadAll(context.TODO(), NewByteReader(ctx, []byte(text)))
	require.NoError(t, err)
	assert.Equal(t, text, string(result))

	cancel()
	result, err = ReadAll(context.TODO(), NewByteReader(ctx, []byte(text)))
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrCancelled)
	assert.Empty(t, result)
}
