package safeio

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

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

func TestSafeCopyDataWithContext(t *testing.T) {
	defer goleak.VerifyNone(t)
	var buf1, buf2 bytes.Buffer
	text := faker.Sentence()
	n, err := WriteString(context.Background(), &buf1, text)
	require.NoError(t, err)
	require.NotZero(t, n)
	assert.Equal(t, len(text), n)
	rc := io.NopCloser(bytes.NewReader(buf1.Bytes())) // make it an io.ReadCloser
	n2, err := SafeCopyDataWithContext(context.Background(), rc, &buf2)
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
	rc = io.NopCloser(bytes.NewReader(buf1.Bytes()))
	n2, err = SafeCopyDataWithContext(ctx, rc, &buf2)
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrCancelled)
	assert.Zero(t, n2)
	assert.Empty(t, buf2.String())

	r, w, err := os.Pipe()
	require.NoError(t, err)
	defer func() { _ = w.Close() }()
	ctx2, unblock := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		_, errCopy := SafeCopyDataWithContext(ctx2, r, io.Discard)
		_ = r.Close()
		_ = errCopy
		close(done)
	}()

	time.Sleep(50 * time.Millisecond) // let it enter read(2) https://man7.org/linux/man-pages/man2/read.2.html
	unblock()

	select {
	case <-done:
		// Expected case: unblocked
	case <-time.After(2 * time.Second):
		assert.FailNow(t, "context cancel should have unblocked copy")
	}
}

func TestSafeCopyNWithContext(t *testing.T) {
	defer goleak.VerifyNone(t)
	var buf1, buf2 bytes.Buffer
	text := faker.Sentence()
	n, err := WriteString(context.Background(), &buf1, text)
	require.NoError(t, err)
	require.NotZero(t, n)
	assert.Equal(t, len(text), n)
	rc := io.NopCloser(bytes.NewReader(buf1.Bytes()))
	n2, err := SafeCopyNWithContext(context.Background(), rc, &buf2, safecast.ToInt64(len(text)))
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
	rc = io.NopCloser(bytes.NewReader(buf1.Bytes()))
	n2, err = SafeCopyNWithContext(ctx, rc, &buf2, safecast.ToInt64(len(text)))
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrCancelled)
	assert.Zero(t, n2)
	assert.Empty(t, buf2.String())

	buf1.Reset()
	buf2.Reset()
	n, err = WriteString(context.Background(), &buf1, text)
	require.NoError(t, err)
	require.NotZero(t, n)
	rc = io.NopCloser(bytes.NewReader(buf1.Bytes()))

	wantN := safecast.ToInt64(len(text) - 1)
	n2, err = SafeCopyNWithContext(context.Background(), rc, &buf2, wantN)
	require.NoError(t, err)
	require.NotZero(t, n2)
	assert.Equal(t, wantN, n2)
	assert.Equal(t, text[:len(text)-1], buf2.String())

	r, w, err := os.Pipe()
	require.NoError(t, err)
	defer func() { _ = w.Close() }()
	ctx2, unblock := context.WithCancel(context.Background())
	done := make(chan struct{})
	var (
		copied  int64
		copyErr error
	)

	go func() {
		copied, copyErr = SafeCopyNWithContext(ctx2, r, io.Discard, 1024) // nothing to read means it blocks
		_ = r.Close()
		close(done)
	}()

	time.Sleep(50 * time.Millisecond) // let it enter read(2) https://man7.org/linux/man-pages/man2/read.2.html
	unblock()

	select {
	case <-done:
		errortest.AssertError(t, copyErr, commonerrors.ErrCancelled)
		assert.Zero(t, copied)
	case <-time.After(2 * time.Second):
		assert.FailNow(t, "context cancel should have unblocked copy")
	}
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
