package logs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
	"go.uber.org/goleak"
)

func TestFIFOLoggerLineIterator(t *testing.T) {
	defer goleak.VerifyNone(t)
	t.Run("logger tests", func(t *testing.T) {
		loggers, err := NewFIFOLogger()
		require.NoError(t, err)
		defer func() { _ = loggers.Close() }()
		testLog(t, loggers)
	})
	t.Run("read lines", func(t *testing.T) {
		loggers, err := NewFIFOLogger()
		require.NoError(t, err)
		defer func() { _ = loggers.Close() }()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		loggers.LogError("Test\r err\n")
		loggers.Log("\rTest1\r")
		count := 0

		var b strings.Builder
		for line := range loggers.LineIterator(ctx) {
			_, err := b.WriteString(line + "\n")
			require.NoError(t, err)
			count++
		}

		assert.Equal(t, "Test err\n\nTest1\n", b.String())
		assert.Equal(t, 3, count) // log error added a line
	})
}

func TestPlainFIFOLoggerLineIterator(t *testing.T) {
	defer goleak.VerifyNone(t)
	t.Run("logger tests", func(t *testing.T) {
		loggers, err := NewPlainFIFOLogger()
		require.NoError(t, err)
		defer func() { _ = loggers.Close() }()
		testLog(t, loggers)
	})
	t.Run("read lines", func(t *testing.T) {
		loggers, err := NewPlainFIFOLogger()
		require.NoError(t, err)
		defer func() { _ = loggers.Close() }()
		go func() {
			time.Sleep(500 * time.Millisecond)
			loggers.LogError("Test err")
			loggers.Log("")
			time.Sleep(100 * time.Millisecond)
			loggers.Log("Test1")
			loggers.Log("\n\n\n")
			time.Sleep(200 * time.Millisecond)
			loggers.Log("Test2\n")
		}()

		count := 0
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		var b strings.Builder
		for line := range loggers.LineIterator(ctx) {
			_, err := b.WriteString(line + "\n")
			require.NoError(t, err)
			count++
		}

		assert.Equal(t, "Test errTest1\n\n\nTest2\n", b.String())
		assert.Equal(t, 4, count)
	})
}

func Test_iterateOverLines(t *testing.T) {
	endIncompleteLine := faker.Word()
	testLines := fmt.Sprintf("%v\n%v", strings.ReplaceAll(faker.Paragraph(), " ", "/r/n"), endIncompleteLine)
	buf := bytes.NewBufferString(testLines)
	numberOfLines := strings.Count(testLines, "\n")
	lineCounter := 0
	yield := func(string) bool {
		lineCounter++
		return true
	}
	t.Run("cancelled", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(context.Background())
		cancel()
		errortest.AssertError(t, iterateOverLines(cancelCtx, buf, yield), commonerrors.ErrCancelled)
		assert.Zero(t, lineCounter)
	})
	t.Run("success", func(t *testing.T) {
		err := iterateOverLines(context.Background(), buf, yield)
		require.NoError(t, err)
		assert.Equal(t, numberOfLines, lineCounter)
		assert.Equal(t, len(endIncompleteLine), buf.Len())
		line, err := buf.ReadString(newLine)
		require.Error(t, err)
		assert.Equal(t, io.EOF, err)
		assert.Equal(t, endIncompleteLine, line)
	})
}

func Test_IterateOverLines(t *testing.T) {
	lastIncompleteLine := faker.Sentence()
	overallLines := []string{fmt.Sprintf("%v\n%v", faker.Word(), faker.Word()), fmt.Sprintf("%v\n%v", strings.ReplaceAll(faker.Sentence(), " ", "\r"), faker.Name()), fmt.Sprintf("%v\n%v\n%v", faker.DomainName(), faker.IPv4(), lastIncompleteLine)}
	expectedLines := strings.Split(strings.ReplaceAll(strings.TrimSuffix(strings.Join(overallLines, ""), "\n"+lastIncompleteLine), "\r", ""), "\n")
	index := 0
	nextLine := func(fCtx context.Context) ([]byte, error) {
		err := parallelisation.DetermineContextError(fCtx)
		if err != nil {
			return nil, err
		}
		if index >= len(overallLines) {
			return nil, nil
		}
		b := []byte(overallLines[index])
		index++
		return b, nil
	}
	lineCounter := 0
	var readLines []string
	yield := func(l string) bool {
		lineCounter++
		readLines = append(readLines, l)
		return true
	}
	cctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := IterateOverLines(cctx, nextLine, yield)
	errortest.AssertError(t, err, commonerrors.ErrTimeout, commonerrors.ErrCancelled)
	assert.Equal(t, 4, lineCounter)
	assert.EqualValues(t, expectedLines, readLines)
}
