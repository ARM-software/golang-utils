/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package subprocess

import (
	"context"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/bxcodec/faker/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/ARM-software/golang-utils/utils/logs"
	"github.com/ARM-software/golang-utils/utils/logs/logstest"
	"github.com/ARM-software/golang-utils/utils/platform"
)

func TestExecuteEmptyLines(t *testing.T) {
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
			randI := rand.Intn(25)       //nolint:gosec //causes G404: Use of weak random number generator (math/rand instead of crypto/rand) (gosec), So disable gosec
			for i := 0; i < randI; i++ { //nolint:gosec //causes G404: Use of weak random number generator (math/rand instead of crypto/rand) (gosec), So disable gosec
				out += faker.Sentence()
				if rand.Intn(10) > 5 { //nolint:gosec //causes G404: Use of weak random number generator (math/rand instead of crypto/rand) (gosec), So disable gosec
					out += platform.LineSeparator()
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
			loggers, err := logs.NewStringLogger("Test") // clean log between each test
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

func TestStartStop(t *testing.T) {
	currentDir, err := os.Getwd()
	require.Nil(t, err)
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
			argWindows: []string{"dir", currentDir},
			cmdOther:   "ls",
			argOther:   []string{"-l", currentDir},
		},
		{
			name:       "LongProcess",
			cmdWindows: "cmd",
			argWindows: []string{"SLEEP 1"},
			cmdOther:   "sleep",
			argOther:   []string{"1"},
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			defer goleak.VerifyNone(t)
			loggers, err := logs.NewLogrLogger(logstest.NewTestLogger(t), "test")
			require.Nil(t, err)

			var p *Subprocess
			if platform.IsWindows() {
				p, err = New(context.Background(), loggers, "", "", "", test.cmdWindows, test.argWindows...)
			} else {
				p, err = New(context.Background(), loggers, "", "", "", test.cmdOther, test.argOther...)
			}
			require.Nil(t, err)
			require.NotNil(t, p)
			assert.False(t, p.IsOn())
			err = p.Start()
			require.Nil(t, err)
			assert.True(t, p.IsOn())

			// Checking idempotence
			err = p.Start()
			require.Nil(t, err)
			err = p.Check()
			require.Nil(t, err)

			time.Sleep(200 * time.Millisecond)
			err = p.Restart()
			require.Nil(t, err)
			assert.True(t, p.IsOn())
			err = p.Stop()
			require.Nil(t, err)
			assert.False(t, p.IsOn())
			// Checking idempotence
			err = p.Stop()
			require.Nil(t, err)
			time.Sleep(100 * time.Millisecond)
			err = p.Execute()
			require.Nil(t, err)
		})
	}
}

func TestExecute(t *testing.T) {
	currentDir, err := os.Getwd()
	require.Nil(t, err)
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
			argWindows: []string{"dir", currentDir},
			cmdOther:   "ls",
			argOther:   []string{"-l", currentDir},
		},
		{
			name:       "LongProcess",
			cmdWindows: "cmd",
			argWindows: []string{"SLEEP 1"},
			cmdOther:   "sleep",
			argOther:   []string{"1"},
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			defer goleak.VerifyNone(t)
			var loggers logs.Loggers = &logs.GenericLoggers{}
			err := loggers.Check()
			assert.NotNil(t, err)

			err = Execute(context.Background(), loggers, "", "", "", "ls")
			assert.NotNil(t, err)

			loggers, err = logs.NewLogrLogger(logstest.NewTestLogger(t), "test")
			require.Nil(t, err)
			if platform.IsWindows() {
				err = Execute(context.Background(), loggers, "", "", "", test.cmdWindows, test.argWindows...)
			} else {
				err = Execute(context.Background(), loggers, "", "", "", test.cmdOther, test.argOther...)
			}
			require.Nil(t, err)
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
			argWindows: []string{"SLEEP 4"},
			cmdOther:   "sleep",
			argOther:   []string{"4"},
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			defer goleak.VerifyNone(t)
			loggers, err := logs.NewLogrLogger(logstest.NewTestLogger(t), "test")
			require.Nil(t, err)
			cancellableCtx, cancelFunc := context.WithCancel(context.Background())

			var p *Subprocess
			if platform.IsWindows() {
				p, err = New(cancellableCtx, loggers, "", "", "", test.cmdWindows, test.argWindows...)
			} else {
				p, err = New(cancellableCtx, loggers, "", "", "", test.cmdOther, test.argOther...)
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
			argWindows: []string{"SLEEP 4"},
			cmdOther:   "sleep",
			argOther:   []string{"4"},
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			defer goleak.VerifyNone(t)
			loggers, err := logs.NewLogrLogger(logstest.NewTestLogger(t), "test")
			require.Nil(t, err)
			ctx, cancelFunc := context.WithCancel(context.Background())
			var p *Subprocess
			if platform.IsWindows() {
				p, err = New(ctx, loggers, "", "", "", test.cmdWindows, test.argWindows...)
			} else {
				p, err = New(ctx, loggers, "", "", "", test.cmdOther, test.argOther...)
			}
			require.Nil(t, err)
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
			time.Sleep(200 * time.Millisecond)
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
			argWindows: []string{"SLEEP 4"},
			cmdOther:   "sleep",
			argOther:   []string{"4"},
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			defer goleak.VerifyNone(t)
			loggers, err := logs.NewLogrLogger(logstest.NewTestLogger(t), "test")
			require.Nil(t, err)
			ctx := context.Background()
			var p *Subprocess
			if platform.IsWindows() {
				p, err = New(ctx, loggers, "", "", "", test.cmdWindows, test.argWindows...)
			} else {
				p, err = New(ctx, loggers, "", "", "", test.cmdOther, test.argOther...)
			}
			require.Nil(t, err)
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
			p.Cancel()
			// checking idempotence.
			p.Cancel()
			time.Sleep(200 * time.Millisecond)
			assert.False(t, p.IsOn())
		})
	}
}
