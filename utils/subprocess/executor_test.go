/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package subprocess

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/logs"
	"github.com/ARM-software/golang-utils/utils/logs/logstest"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
	"github.com/ARM-software/golang-utils/utils/platform"
)

type testIO struct {
	in  io.Reader
	out *bytes.Buffer
	err *bytes.Buffer
}

func newTestIO() *testIO {
	return &testIO{
		in:  strings.NewReader(""),
		out: &bytes.Buffer{},
		err: &bytes.Buffer{},
	}
}

func (t *testIO) Register(context.Context) (io.Reader, io.Writer, io.Writer) {
	return t.in, t.out, t.err
}

type execFunc func(ctx context.Context, l logs.Loggers, cmd string, args ...string) error
type startFunc func(ctx context.Context, l logs.Loggers, cmd string, args ...string) (*Subprocess, error)

func newDefaultExecutor(t *testing.T) execFunc {
	t.Helper()
	return func(ctx context.Context, l logs.Loggers, cmd string, args ...string) error {
		return Execute(ctx, l, "", "", "", cmd, args...)
	}
}

func newCustomIOExecutor(t *testing.T, customIO *testIO) execFunc {
	t.Helper()
	return func(ctx context.Context, l logs.Loggers, cmd string, args ...string) (err error) {
		p := &Subprocess{}
		err = p.SetupWithEnvironmentWithCustomIO(ctx, l, customIO, nil, "", "", "", cmd, args...)
		if err != nil {
			return
		}
		return p.Execute()
	}
}

func newDefaultStarter(t *testing.T) startFunc {
	t.Helper()
	return func(ctx context.Context, l logs.Loggers, cmd string, args ...string) (*Subprocess, error) {
		return Start(ctx, l, "", "", "", cmd, args...)
	}
}

func newCustomIOStarter(t *testing.T, customIO *testIO) startFunc {
	t.Helper()
	return func(ctx context.Context, l logs.Loggers, cmd string, args ...string) (p *Subprocess, err error) {
		p = &Subprocess{}
		err = p.SetupWithEnvironmentWithCustomIO(ctx, l, customIO, nil, "", "", "", cmd, args...)
		if err != nil {
			return
		}
		err = p.Start()
		return
	}
}

func TestExecuteEmptyLines(t *testing.T) {
	t.Skip("would need to be reinstated when fixed")
	defer goleak.VerifyNone(t)
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
			random, err := faker.RandomInt(0, 25, 1)
			require.NoError(t, err)
			for i := 0; i < random[0]; i++ {
				out += faker.Sentence()
				randomJ, err := faker.RandomInt(0, 10, 1)
				require.NoError(t, err)
				if randomJ[0] > 5 {
					out += platform.LineSeparator()
				}
			}
			return
		}(),
	}

	newline := "\n"
	if platform.IsWindows() {
		newline = "\r\n"
	}

	edgeCases := []string{ // both these would mess with the regex
		newline, // just a '\n'
		"",      // empty string
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
		t.Run(fmt.Sprintf("test #%v", i), func(t *testing.T) {
			test := tests[i]

			for j := range test.Inputs {
				loggers, err := logs.NewStringLogger("Test") // clean log between each test
				require.NoError(t, err)
				if platform.IsWindows() {
					err = Execute(context.Background(), loggers, "", "", "", "cmd", "/c", "echo", test.Inputs[j])
				} else {
					err = Execute(context.Background(), loggers, "", "", "", "echo", test.Inputs[j])
				}
				require.NoError(t, err)

				contents := loggers.GetLogContent()
				require.NotZero(t, contents)

				actualLines := strings.Split(contents, "\n")
				expectedLines := strings.Split(test.ExpectedOutputs[j], "\n")
				fmt.Println("A:::::: ", actualLines)
				fmt.Println("B:::::: ", expectedLines)
				t.Run(fmt.Sprintf("%v", expectedLines), func(t *testing.T) {
					require.Len(t, actualLines, len(expectedLines)+3-i) // length of test string without ' ' + the two logs saying it is starting and complete + empty line at start (remove i to account for the blank line)
					for k, line := range actualLines[1 : len(actualLines)-2] {
						assert.Contains(t, line, expectedLines[k]) // if the newlines were removed then these would line up
					}
				})
			}
		})
	}
}

func TestStartStop(t *testing.T) {
	currentDir, err := os.Getwd()
	require.NoError(t, err)
	tests := []struct {
		name       string
		cmdWindows string
		argWindows []string
		cmdOther   string
		argOther   []string
	}{
		{
			name:       "ShortProcess",
			cmdWindows: "cmd",
			argWindows: []string{"/c", "dir", currentDir},
			cmdOther:   "ls",
			argOther:   []string{"-l", currentDir},
		},
		{
			name:       "LongProcess",
			cmdWindows: "cmd",
			argWindows: []string{"/c", fmt.Sprintf("ping -n 2 -w %v localhost > nul", time.Second.Milliseconds())}, // See https://stackoverflow.com/a/79268314/45375
			cmdOther:   "sleep",
			argOther:   []string{"1"},
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			defer goleak.VerifyNone(t)
			loggers, err := logs.NewLogrLogger(logstest.NewTestLogger(t), "test")
			require.NoError(t, err)

			var p *Subprocess
			if platform.IsWindows() {
				p, err = New(context.Background(), loggers, "", "", "", test.cmdWindows, test.argWindows...)
			} else {
				p, err = New(context.Background(), loggers, "", "", "", test.cmdOther, test.argOther...)
			}
			require.NoError(t, err)
			require.NotNil(t, p)
			assert.False(t, p.IsOn())
			err = p.Start()
			require.NoError(t, err)
			assert.True(t, p.IsOn())

			// Checking idempotence
			err = p.Start()
			require.NoError(t, err)
			err = p.Check()
			require.NoError(t, err)

			time.Sleep(200 * time.Millisecond)
			err = p.Restart()
			require.NoError(t, err)
			assert.True(t, p.IsOn())
			err = p.Stop()
			require.NoError(t, err)
			assert.False(t, p.IsOn())
			// Checking idempotence
			err = p.Stop()
			require.NoError(t, err)
			time.Sleep(100 * time.Millisecond)
			err = p.Execute()
			require.NoError(t, err)
		})
	}
}

func TestStartInterrupt(t *testing.T) {
	currentDir, err := os.Getwd()
	require.NoError(t, err)
	tests := []struct {
		name       string
		cmdWindows string
		argWindows []string
		cmdOther   string
		argOther   []string
	}{
		{
			name:       "ShortProcess",
			cmdWindows: "cmd",
			argWindows: []string{"/c", "dir", currentDir},
			cmdOther:   "ls",
			argOther:   []string{"-l", currentDir},
		},
		{
			name:       "LongProcess",
			cmdWindows: "cmd",
			argWindows: []string{"/c", fmt.Sprintf("ping -n 2 -w %v localhost > nul", time.Second.Milliseconds())}, // See https://stackoverflow.com/a/79268314/45375
			cmdOther:   "sleep",
			argOther:   []string{"1"},
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			defer goleak.VerifyNone(t)
			loggers, err := logs.NewLogrLogger(logstest.NewTestLogger(t), "test")
			require.NoError(t, err)

			var p *Subprocess
			if platform.IsWindows() {
				p, err = New(context.Background(), loggers, "", "", "", test.cmdWindows, test.argWindows...)
			} else {
				p, err = New(context.Background(), loggers, "", "", "", test.cmdOther, test.argOther...)
			}
			require.NoError(t, err)
			require.NotNil(t, p)
			assert.False(t, p.IsOn())
			err = p.Start()
			require.NoError(t, err)
			assert.True(t, p.IsOn())

			// Checking idempotence
			err = p.Start()
			require.NoError(t, err)
			err = p.Check()
			require.NoError(t, err)

			time.Sleep(200 * time.Millisecond)
			err = p.Restart()
			require.NoError(t, err)
			assert.True(t, p.IsOn())
			err = p.Interrupt(context.Background())
			require.NoError(t, err)
			assert.False(t, p.IsOn())
			// Checking idempotence
			err = p.Interrupt(context.Background())
			require.NoError(t, err)
			time.Sleep(100 * time.Millisecond)
			err = p.Execute()
			require.NoError(t, err)
		})
	}
}
func TestExecute(t *testing.T) {
	currentDir, err := os.Getwd()
	require.NoError(t, err)

	tests := []struct {
		name       string
		cmdWindows string
		argWindows []string
		cmdOther   string
		argOther   []string
		expectIO   bool
	}{
		{
			name:       "ShortProcess",
			cmdWindows: "cmd",
			argWindows: []string{"/c", "dir", currentDir},
			cmdOther:   "ls",
			argOther:   []string{"-l", currentDir},
			expectIO:   true,
		},
		{
			name:       "LongProcess",
			cmdWindows: "cmd",
			argWindows: []string{"/c", fmt.Sprintf("ping -n 2 -w %v localhost > nul", time.Second.Milliseconds())}, // See https://stackoverflow.com/a/79268314/45375
			cmdOther:   "sleep",
			argOther:   []string{"1"},
			expectIO:   false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer goleak.VerifyNone(t)

			customIO := newTestIO()
			executors := []struct {
				name string
				run  execFunc
				io   *testIO
			}{
				{"normal", newDefaultExecutor(t), nil},
				{"with IO", newCustomIOExecutor(t, customIO), customIO},
			}

			for _, executor := range executors {
				t.Run(executor.name, func(t *testing.T) {
					var loggers logs.Loggers = &logs.GenericLoggers{}
					err := loggers.Check()
					assert.Error(t, err)

					err = executor.run(context.Background(), loggers, "ls")
					assert.Error(t, err)

					loggers, err = logs.NewLogrLogger(logstest.NewTestLogger(t), "test")
					require.NoError(t, err)

					if platform.IsWindows() {
						err = executor.run(context.Background(), loggers, test.cmdWindows, test.argWindows...)
					} else {
						err = executor.run(context.Background(), loggers, test.cmdOther, test.argOther...)
					}
					require.NoError(t, err)

					if executor.io != nil && test.expectIO {
						assert.NotZero(t, executor.io.out.Len()+executor.io.err.Len()) // expect some output
					}
				})
			}
		})
	}
}

func TestStart(t *testing.T) {

	currentDir, err := os.Getwd()
	require.NoError(t, err)

	tests := []struct {
		name       string
		cmdWindows string
		argWindows []string
		cmdOther   string
		argOther   []string
		expectIO   bool
	}{
		{
			name:       "ShortProcess",
			cmdWindows: "cmd",
			argWindows: []string{"/c", "dir", currentDir},
			cmdOther:   "ls",
			argOther:   []string{"-l", currentDir},
			expectIO:   true,
		},
		{
			name:       "LongProcess",
			cmdWindows: "cmd",
			argWindows: []string{"/c", fmt.Sprintf("ping -n 2 -w %v localhost > nul", time.Second.Milliseconds())}, // See https://stackoverflow.com/a/79268314/45375
			cmdOther:   "sleep",
			argOther:   []string{"1"},
			expectIO:   false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer goleak.VerifyNone(t)

			customIO := newTestIO()
			executors := []struct {
				name  string
				start startFunc
				io    *testIO
			}{
				{"normal", newDefaultStarter(t), nil},
				{"with IO", newCustomIOStarter(t, customIO), customIO},
			}

			for _, executor := range executors {
				t.Run(executor.name, func(t *testing.T) {
					var loggers logs.Loggers = &logs.GenericLoggers{}
					err := loggers.Check()
					assert.Error(t, err)

					_, err = executor.start(context.Background(), loggers, "ls")
					assert.Error(t, err)

					loggers, err = logs.NewLogrLogger(logstest.NewTestLogger(t), "test")
					require.NoError(t, err)

					var p *Subprocess
					if platform.IsWindows() {
						p, err = executor.start(context.Background(), loggers, test.cmdWindows, test.argWindows...)
					} else {
						p, err = executor.start(context.Background(), loggers, test.cmdOther, test.argOther...)
					}
					require.NoError(t, err)
					require.NotNil(t, p)
					defer func() { _ = p.Stop() }()
					require.NoError(t, p.Wait(context.Background()))
					p.Cancel()
				})
			}
		})
	}
}

func TestExecuteWithCustomIO_Stderr(t *testing.T) {
	if platform.IsWindows() {
		t.Skip("Uses bash and redirection so can't run on Windows")
	}
	defer goleak.VerifyNone(t)

	loggers, err := logs.NewLogrLogger(logstest.NewTestLogger(t), "test")
	require.NoError(t, err)

	customIO := newTestIO()
	run := newCustomIOExecutor(t, customIO)

	msg := "hello adrien"
	err = run(context.Background(), loggers, "bash", "-c", fmt.Sprintf("echo %s 1>&2", msg))
	require.NoError(t, err)

	require.Empty(t, customIO.out.String()) // should be no stdout
	require.Equal(t, fmt.Sprintln(msg), customIO.err.String())
}

func TestOutput(t *testing.T) {
	loggers, err := logs.NewLogrLogger(logstest.NewTestLogger(t), "testOutput")
	require.NoError(t, err)
	currentDir, err := os.Getwd()
	require.NoError(t, err)
	tests := []struct {
		name         string
		cmdWindows   string
		argWindows   []string
		cmdOther     string
		argOther     []string
		expectOutput bool
		runCount     int
	}{
		{
			name:         "ShortProcess",
			cmdWindows:   "cmd",
			argWindows:   []string{"/c", "dir", currentDir},
			cmdOther:     "ls",
			argOther:     []string{"-l", currentDir},
			expectOutput: true,
			runCount:     1,
		},
		{
			name:       "LongProcess",
			cmdWindows: "cmd",
			argWindows: []string{"/c", fmt.Sprintf("ping -n 2 -w %v localhost > nul", time.Second.Milliseconds())}, // See https://stackoverflow.com/a/79268314/45375
			cmdOther:   "sleep",
			argOther:   []string{"1"},
			runCount:   1,
		},
		{
			name:         "BothStdOutAndStdErr",
			cmdOther:     "./testdata/echo_stdout_and_stderr.sh",
			argOther:     []string{"foo"},
			expectOutput: true,
			runCount:     5,
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			defer goleak.VerifyNone(t)
			var output string
			for i := 0; i < test.runCount; i++ {
				if platform.IsWindows() {
					if test.cmdWindows == "" {
						t.Skip("Not suitable for Windows")
					} else {
						output, err = Output(context.Background(), loggers, test.cmdWindows, test.argWindows...)
					}
				} else {
					output, err = Output(context.Background(), loggers, test.cmdOther, test.argOther...)
				}
				require.NoError(t, err)
				if test.expectOutput {
					assert.NotEmpty(t, output)
				}
			}
		})
	}
}

func TestCancelledSubprocess(t *testing.T) {
	tests := []struct {
		name       string
		cmdWindows string
		argWindows []string
		cmdOther   string
		argOther   []string
	}{
		{
			name:       "LongProcess",
			cmdWindows: "cmd",
			argWindows: []string{"/c", fmt.Sprintf("ping -n 2 -w %v localhost > nul", (10 * time.Second).Milliseconds())}, // See https://stackoverflow.com/a/79268314/45375
			cmdOther:   "sleep",
			argOther:   []string{"4"},
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			defer goleak.VerifyNone(t)
			loggers, err := logs.NewLogrLogger(logstest.NewTestLogger(t), "test")
			require.NoError(t, err)
			cancellableCtx, cancelFunc := context.WithCancel(context.Background())

			var p *Subprocess
			if platform.IsWindows() {
				p, err = New(cancellableCtx, loggers, "", "", "", test.cmdWindows, test.argWindows...)
			} else {
				p, err = New(cancellableCtx, loggers, "", "", "", test.cmdOther, test.argOther...)
			}
			require.NoError(t, err)
			defer func() { _ = p.Stop() }()

			assert.False(t, p.IsOn())
			err = p.Start()
			require.NoError(t, err)
			assert.True(t, p.IsOn())
			time.Sleep(10 * time.Millisecond)
			cancelFunc()
			cancelCtx, cancelFunc := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancelFunc()
			require.NoError(t, parallelisation.WaitUntil(cancelCtx, func(ctx2 context.Context) (done bool, err error) {
				err = parallelisation.DetermineContextError(ctx2)
				if err != nil {
					return
				}
				done = !p.IsOn()
				return
			}, 50*time.Millisecond))
			assert.False(t, p.IsOn())
		})
	}
}

func TestCancelledSubprocess2(t *testing.T) {
	tests := []struct {
		name       string
		cmdWindows string
		argWindows []string
		cmdOther   string
		argOther   []string
	}{
		{
			name:       "LongProcess",
			cmdWindows: "cmd",
			argWindows: []string{"/c", fmt.Sprintf("ping -n 2 -w %v localhost > nul", (4 * time.Second).Milliseconds())}, // See https://stackoverflow.com/a/79268314/45375
			cmdOther:   "sleep",
			argOther:   []string{"10"},
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			defer goleak.VerifyNone(t)
			loggers, err := logs.NewLogrLogger(logstest.NewTestLogger(t), "test")
			require.NoError(t, err)
			ctx, cancelFunc := context.WithCancel(context.Background())
			var p *Subprocess
			if platform.IsWindows() {
				p, err = New(ctx, loggers, "", "", "", test.cmdWindows, test.argWindows...)
			} else {
				p, err = New(ctx, loggers, "", "", "", test.cmdOther, test.argOther...)
			}
			require.NoError(t, err)
			defer func() { _ = p.Stop() }()

			ready := make(chan bool)
			go func(proc *Subprocess) {
				ready <- true
				_ = proc.Execute()
			}(p)
			<-ready
			time.Sleep(10 * time.Millisecond)
			assert.True(t, p.IsOn())
			time.Sleep(10 * time.Millisecond)
			cancelFunc()
			cancelCtx, cancelFunc := context.WithTimeout(context.Background(), time.Second)
			defer cancelFunc()
			require.NoError(t, parallelisation.WaitUntil(cancelCtx, func(ctx2 context.Context) (done bool, err error) {
				err = parallelisation.DetermineContextError(ctx2)
				if err != nil {
					return
				}
				done = !p.IsOn()
				return
			}, 50*time.Millisecond))
			assert.False(t, p.IsOn())
		})
	}
}

func TestCancelledSubprocess3(t *testing.T) {
	tests := []struct {
		name       string
		cmdWindows string
		argWindows []string
		cmdOther   string
		argOther   []string
	}{
		{
			name:       "LongProcess",
			cmdWindows: "cmd",
			argWindows: []string{"/c", fmt.Sprintf("ping -n 2 -w %v localhost > nul", (4 * time.Second).Milliseconds())}, // See https://stackoverflow.com/a/79268314/45375
			cmdOther:   "sleep",
			argOther:   []string{"4"},
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			defer goleak.VerifyNone(t)
			loggers, err := logs.NewLogrLogger(logstest.NewTestLogger(t), "test")
			require.NoError(t, err)
			ctx := context.Background()
			var p *Subprocess
			if platform.IsWindows() {
				p, err = New(ctx, loggers, "", "", "", test.cmdWindows, test.argWindows...)
			} else {
				p, err = New(ctx, loggers, "", "", "", test.cmdOther, test.argOther...)
			}
			require.NoError(t, err)
			defer func() { _ = p.Stop() }()

			ready := make(chan bool)
			go func(proc *Subprocess) {
				ready <- true
				_ = proc.Execute()
			}(p)
			<-ready
			cancelCtx, cancelFunc := context.WithTimeout(ctx, time.Second)
			defer cancelFunc()
			require.NoError(t, parallelisation.WaitUntil(cancelCtx, func(ctx2 context.Context) (done bool, err error) {
				err = parallelisation.DetermineContextError(ctx2)
				if err != nil {
					return
				}
				done = p.IsOn()
				return
			}, 50*time.Millisecond))
			assert.True(t, p.IsOn())
			time.Sleep(10 * time.Millisecond)
			p.Cancel()
			// checking idempotence.
			p.Cancel()
			cancelCtx, cancelFunc = context.WithTimeout(context.Background(), 2*time.Second)
			defer cancelFunc()
			require.NoError(t, parallelisation.WaitUntil(cancelCtx, func(ctx2 context.Context) (done bool, err error) {
				err = parallelisation.DetermineContextError(ctx2)
				if err != nil {
					return
				}
				done = !p.IsOn()
				return
			}, 50*time.Millisecond))
			assert.False(t, p.IsOn())
		})
	}
}

func TestOutputWithEnvironment(t *testing.T) {
	if platform.IsWindows() {
		t.Skip("access denied")
	}
	defer goleak.VerifyNone(t)
	logger, err := logs.NewLogrLogger(logstest.NewTestLogger(t), "test")
	require.NoError(t, err)
	t.Run("happy", func(t *testing.T) {
		output, err := OutputWithEnvironment(context.Background(), logger, nil, "du", "-h")
		require.NoError(t, err)
		assert.NotEmpty(t, output)
		fmt.Println(output)
	})
	t.Run("happy with output", func(t *testing.T) {
		testString := fmt.Sprintf("'This is a test %v!'", faker.Sentence())
		output, err := OutputWithEnvironment(context.Background(), logger, nil, "echo", testString)
		require.NoError(t, err)
		assert.NotEmpty(t, output)
		assert.Equal(t, testString, strings.TrimSpace(output))
	})
	t.Run("happy with output in stderr", func(t *testing.T) {
		testString := fmt.Sprintf("This is a test %v!", faker.Sentence())
		output, err := OutputWithEnvironment(context.Background(), logger, nil, "bash", "-c", fmt.Sprintf("echo %v 1>&2", testString))
		require.NoError(t, err)
		assert.NotEmpty(t, output)
		assert.Equal(t, testString, strings.TrimSpace(output))
	})
	t.Run("environment", func(t *testing.T) {
		testString := fmt.Sprintf("This is a test %v!", faker.Sentence())
		output, err := OutputWithEnvironment(context.Background(), logger, []string{fmt.Sprintf("TEST_ENV=%v", testString)}, "env")
		require.NoError(t, err)
		assert.NotEmpty(t, output)
		fmt.Println(output)

	})
	t.Run("environment 2", func(t *testing.T) {
		testString := fmt.Sprintf("This is a test %v!", faker.Sentence())
		output, err := OutputWithEnvironment(context.Background(), logger, []string{fmt.Sprintf("TEST_ENV=%v", testString)}, "bash", "-c", "echo ${TEST_ENV}")
		require.NoError(t, err)
		assert.NotEmpty(t, output)
		assert.Equal(t, testString, strings.TrimSpace(output))
	})
}

func TestWait(t *testing.T) {
	t.Run("Valid subprocess returns no error", func(t *testing.T) {
		var cmd *exec.Cmd
		if platform.IsWindows() {
			// See https://stackoverflow.com/a/79268314/45375
			cmd = exec.Command("cmd", "/c", fmt.Sprintf("ping -n 2 -w %v localhost > nul", (time.Second).Milliseconds())) //nolint:gosec // Causes G204: Subprocess launched with a potential tainted input or cmd arguments
		} else {
			cmd = exec.Command("sh", "-c", "sleep 1")
		}
		defer func() { _ = CleanKillOfCommand(context.Background(), cmd) }()
		require.NoError(t, cmd.Start())

		p := &Subprocess{
			command: &command{
				cmdWrapper: cmdWrapper{
					cmd: cmd,
				},
			},
		}

		assert.NoError(t, p.Wait(context.Background()))
	})

	t.Run("Invalid subprocess returns expected error", func(t *testing.T) {
		p := &Subprocess{command: nil}
		err := p.Wait(context.Background())

		require.Error(t, err)
		errortest.AssertError(t, err, commonerrors.ErrConflict)
	})

	t.Run("Cancelled context returns error", func(t *testing.T) {
		var cmd *exec.Cmd
		if platform.IsWindows() {
			// See https://stackoverflow.com/a/79268314/45375
			cmd = exec.Command("cmd", "/c", fmt.Sprintf("ping -n 2 -w %v localhost > nul", (10*time.Second).Milliseconds())) //nolint:gosec // Causes G204: Subprocess launched with a potential tainted input or cmd arguments
		} else {
			cmd = exec.Command("sh", "-c", "sleep 10")
		}
		defer func() { _ = CleanKillOfCommand(context.Background(), cmd) }()
		require.NoError(t, cmd.Start())

		p := &Subprocess{
			command: &command{
				cmdWrapper: cmdWrapper{
					cmd: cmd,
				},
			},
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		errortest.AssertError(t, p.Wait(ctx), commonerrors.ErrCancelled)
	})
}
