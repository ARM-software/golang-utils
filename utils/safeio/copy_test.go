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
	"github.com/ARM-software/golang-utils/utils/safecast"
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
	assert.Equal(t, safecast.ToInt64(len(text)), n2)
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
	n2, err := CopyNWithContext(context.Background(), &buf1, &buf2, safecast.ToInt64(len(text)))
	require.NoError(t, err)
	require.NotZero(t, n2)
	assert.Equal(t, safecast.ToInt64(len(text)), n2)
	assert.Equal(t, text, buf2.String())

	ctx, cancel := context.WithCancel(context.Background())
	buf1.Reset()
	buf2.Reset()
	n, err = WriteString(context.Background(), &buf1, text)
	require.NoError(t, err)
	require.NotZero(t, n)
	assert.Equal(t, len(text), n)

	cancel()
	n2, err = CopyNWithContext(ctx, &buf1, &buf2, safecast.ToInt64(len(text)))
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrCancelled)
	assert.Zero(t, n2)
	assert.Empty(t, buf2.String())

	n, err = WriteString(context.Background(), &buf1, text)
	require.NoError(t, err)
	require.NotZero(t, n)
	assert.Equal(t, len(text), n)
	n2, err = CopyNWithContext(context.Background(), &buf1, &buf2, safecast.ToInt64(len(text)-1))
	require.NoError(t, err)
	require.NotZero(t, n2)
	assert.Equal(t, safecast.ToInt64(len(text)-1), n2)
}

func TestCat(t *testing.T) {
	var buf1, buf2, buf3 bytes.Buffer
	text1 := faker.Sentence()
	text2 := faker.Paragraph()
	n, err := WriteString(context.Background(), &buf1, text1)
	require.NoError(t, err)
	require.NotZero(t, n)
	assert.Equal(t, len(text1), n)
	n, err = WriteString(context.Background(), &buf2, text2)
	require.NoError(t, err)
	require.NotZero(t, n)
	assert.Equal(t, len(text2), n)
	n3, err := Cat(context.Background(), &buf3, &buf1, &buf2)
	require.NoError(t, err)
	require.NotZero(t, n3)
	assert.Equal(t, safecast.ToInt64(len(text1)+len(text2)), n3)
	assert.Equal(t, text1+text2, buf3.String())

	ctx, cancel := context.WithCancel(context.Background())
	buf1.Reset()
	buf2.Reset()
	buf3.Reset()
	n, err = WriteString(context.Background(), &buf1, text1)
	require.NoError(t, err)
	require.NotZero(t, n)
	assert.Equal(t, len(text1), n)
	n, err = WriteString(context.Background(), &buf2, text2)
	require.NoError(t, err)
	require.NotZero(t, n)
	assert.Equal(t, len(text2), n)

	cancel()
	n3, err = Cat(ctx, &buf3, &buf1, &buf2)
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrCancelled)
	assert.Zero(t, n3)
	assert.Empty(t, buf3.String())
}

func TestCatN(t *testing.T) {
	var buf1, buf2, buf3 bytes.Buffer
	text1 := faker.Sentence()
	text2 := faker.Paragraph()
	n, err := WriteString(context.Background(), &buf1, text1)
	require.NoError(t, err)
	require.NotZero(t, n)
	assert.Equal(t, len(text1), n)
	n, err = WriteString(context.Background(), &buf2, text2)
	require.NoError(t, err)
	require.NotZero(t, n)
	assert.Equal(t, len(text2), n)
	n3, err := CatN(context.Background(), &buf3, safecast.ToInt64(len(text1)+len(text2)), &buf1, &buf2)
	require.NoError(t, err)
	require.NotZero(t, n3)
	assert.Equal(t, safecast.ToInt64(len(text1)+len(text2)), n3)
	assert.Equal(t, text1+text2, buf3.String())

	ctx, cancel := context.WithCancel(context.Background())
	buf1.Reset()
	buf2.Reset()
	buf3.Reset()
	n, err = WriteString(context.Background(), &buf1, text1)
	require.NoError(t, err)
	require.NotZero(t, n)
	assert.Equal(t, len(text1), n)
	n, err = WriteString(context.Background(), &buf2, text2)
	require.NoError(t, err)
	require.NotZero(t, n)
	assert.Equal(t, len(text2), n)

	cancel()
	n3, err = CatN(ctx, &buf3, safecast.ToInt64(len(text1)+len(text2)), &buf1, &buf2)
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrCancelled)
	assert.Zero(t, n3)
	assert.Empty(t, buf3.String())

	n3, err = CatN(context.Background(), &buf3, safecast.ToInt64(len(text1)+1), &buf1, &buf2)
	require.NoError(t, err)
	require.NotZero(t, n3)
	assert.Equal(t, safecast.ToInt64(len(text1)+1), n3)
}
