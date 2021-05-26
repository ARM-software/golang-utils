package subprocess

import (
	"context"
	"math/rand"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/bxcodec/faker/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARMmbed/golang-utils/utils/logs"
	"github.com/ARMmbed/golang-utils/utils/platform"
)

func TestExecute(t *testing.T) {
	var loggers logs.Loggers = &logs.GenericLoggers{}
	err := loggers.Check()
	assert.NotNil(t, err)

	err = Execute(context.Background(), loggers, "", "", "", "ls")
	assert.NotNil(t, err)

	loggers, err = logs.CreateStdLogger("Test")
	require.Nil(t, err)
	if platform.IsWindows() {
		err = Execute(context.Background(), loggers, "", "", "", "cmd", "/c", "dir")
	} else {
		err = Execute(context.Background(), loggers, "", "", "", "ls", "-l")
	}
	require.Nil(t, err)
}

func TestExecuteEmptyLines(t *testing.T) {

	multilineEchos := []string{ // Some weird lines with contents and empty lines to be filtered
		`hello

world
test 1

#####



`,
		" ",
		faker.Word(),
		faker.Paragraph(),
		faker.Sentence(),
		func() (out string) { // funky random paragraph with plenty of random newlines
			randI := rand.Intn(25)
			for i := 0; i < randI; i++ {
				out += faker.Sentence()
				if rand.Intn(10) > 5 {
					out += "\n"
				}
			}
			return
		}(),
	}

	edgeCases := []string{ // both these would mess with the regex
		`
`, // just a '\n'
		"", // empty string
	}

	var cleanedLines []string
	for _, multilineEcho := range multilineEchos {
		cleanedMultiline := regexp.MustCompile(`[\t\r\n]+`).ReplaceAllString(strings.TrimSpace(multilineEcho), "\n")
		cleanedLines = append(cleanedLines, cleanedMultiline)
	}

	tests := []struct {
		Inputs          []string
		ExpectedOutputs []string
	}{
		{ // Normal tests
			multilineEchos,
			cleanedLines,
		},
		{ // Edge cases where the line will be deleted (these don't cause the logger to print a blank line)
			edgeCases,
			[]string{
				"",
				"",
			},
		},
	}

	for i := range tests {
		for j, testInput := range tests[i].Inputs {
			loggers, err := logs.CreateStringLogger("Test") // clean log between each test
			require.Nil(t, err)

			err = Execute(context.Background(), loggers, "", "", "", "echo", testInput)
			require.Nil(t, err)

			contents := loggers.GetLogContent()
			require.NotZero(t, contents)

			actualLines := strings.Split(contents, "\n")
			expectedLines := strings.Split(tests[i].ExpectedOutputs[j], "\n")
			require.Len(t, actualLines, len(expectedLines)+3-i) // length of test string without ' ' + the two logs saying it is starting and complete + empty line at start (remove i to account for the blank line)

			for k, line := range actualLines[1 : len(actualLines)-2] {
				b := strings.Contains(line, expectedLines[k]) // if the newlines were removed then these would line up
				require.True(t, b)
			}
		}
	}
}

func TestShortLivedSubprocess(t *testing.T) {
	var loggers logs.Loggers = &logs.GenericLoggers{}
	err := loggers.Check()
	assert.NotNil(t, err)

	_, err = New(context.Background(), loggers, "", "", "", "ls")
	assert.NotNil(t, err)

	loggers, err = logs.CreateStdLogger("Test")
	require.Nil(t, err)
	var p *Subprocess
	if platform.IsWindows() {
		p, err = New(context.Background(), loggers, "", "", "", "cmd", "dir")
	} else {
		p, err = New(context.Background(), loggers, "", "", "", "ls", "-l")
	}
	require.Nil(t, err)
	defer func() { _ = p.Stop() }()
	_testSubprocess(t, p)
}

func TestLongerLivedSubprocess(t *testing.T) {
	var loggers, err = logs.CreateStdLogger("Test")
	require.Nil(t, err)

	var p *Subprocess
	if platform.IsWindows() {
		p, err = New(context.Background(), loggers, "", "", "", "cmd", "SLEEP 4")
	} else {
		p, err = New(context.Background(), loggers, "", "", "", "sleep", "4")
	}
	require.Nil(t, err)
	defer func() { _ = p.Stop() }()
	_testSubprocess(t, p)
}

func _testSubprocess(t *testing.T, p *Subprocess) {
	assert.False(t, p.IsOn())
	err := p.Start()
	require.Nil(t, err)
	assert.True(t, p.IsOn())

	//Checking idempotence
	err = p.Start()
	require.Nil(t, err)
	err = p.Check()
	require.Nil(t, err)

	time.Sleep(time.Duration(200) * time.Millisecond)

	err = p.Restart()
	require.Nil(t, err)
	assert.True(t, p.IsOn())

	err = p.Stop()
	require.Nil(t, err)
	assert.False(t, p.IsOn())

	//Checking idempotence
	err = p.Stop()
	require.Nil(t, err)

	err = p.Execute()
	assert.NotNil(t, err)
}

func TestCancelledLongerLivedSubprocess(t *testing.T) {
	var loggers, err = logs.CreateStdLogger("Test")
	require.Nil(t, err)
	cancellableCtx, cancelFunc := context.WithCancel(context.Background())

	var p *Subprocess
	if platform.IsWindows() {
		p, err = New(cancellableCtx, loggers, "", "", "", "cmd", "SLEEP 4")
	} else {
		p, err = New(cancellableCtx, loggers, "", "", "", "sleep", "4")
	}
	require.Nil(t, err)
	defer func() { _ = p.Stop() }()

	assert.False(t, p.IsOn())
	err = p.Start()
	require.Nil(t, err)
	assert.True(t, p.IsOn())
	time.Sleep(10 * time.Millisecond)
	cancelFunc()
	time.Sleep(200 * time.Millisecond)
	assert.False(t, p.IsOn())
}

func TestCancelledLongerLivedSubprocess2(t *testing.T) {
	var loggers, err = logs.CreateStdLogger("Test")
	require.Nil(t, err)
	ctx := context.Background()
	var p *Subprocess
	if platform.IsWindows() {
		p, err = New(ctx, loggers, "", "", "", "cmd", "SLEEP 4")
	} else {
		p, err = New(ctx, loggers, "", "", "", "sleep", "4")
	}
	require.Nil(t, err)
	defer func() { _ = p.Stop() }()

	ready := make(chan bool)
	go func() {
		ready <- true
		_ = p.Execute()
	}()
	<-ready
	time.Sleep(10 * time.Millisecond)
	assert.True(t, p.IsOn())
	time.Sleep(10 * time.Millisecond)
	p.Cancel()
	time.Sleep(200 * time.Millisecond)
	assert.False(t, p.IsOn())
}

func TestCancelledLongerLivedSubprocess3(t *testing.T) {
	var loggers, err = logs.CreateStdLogger("Test")
	require.Nil(t, err)
	ctx, cancelFunc := context.WithCancel(context.Background())
	var p *Subprocess
	if platform.IsWindows() {
		p, err = New(ctx, loggers, "", "", "", "cmd", "SLEEP 4")
	} else {
		p, err = New(ctx, loggers, "", "", "", "sleep", "4")
	}
	require.Nil(t, err)
	defer func() { _ = p.Stop() }()

	ready := make(chan bool)
	go func() {
		ready <- true
		_ = p.Execute()
	}()
	<-ready
	time.Sleep(10 * time.Millisecond)
	assert.True(t, p.IsOn())
	time.Sleep(10 * time.Millisecond)
	cancelFunc()
	time.Sleep(200 * time.Millisecond)
	assert.False(t, p.IsOn())
}
