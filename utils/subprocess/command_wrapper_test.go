package subprocess

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/ARM-software/golang-utils/utils/logs"
	"github.com/ARM-software/golang-utils/utils/platform"
)

func TestCmdRun(t *testing.T) {
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
			var cmd *command
			loggers, err := logs.CreateStdLogger("Test")
			require.Nil(t, err)
			if platform.IsWindows() {
				cmd = newCommand(loggers, test.cmdWindows, test.argWindows...)
			} else {
				cmd = newCommand(loggers, test.cmdOther, test.argOther...)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			wrapper := cmd.GetCmd(ctx)
			err = wrapper.Run()
			require.Nil(t, err)
			err = wrapper.Run()
			require.NotNil(t, err)
			cmd.Reset()
			wrapper = cmd.GetCmd(ctx)
			err = wrapper.Run()
			require.Nil(t, err)
		})
	}
}

func TestCmdStartStop(t *testing.T) {
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
			argWindows: []string{"SLEEP 4"},
			cmdOther:   "sleep",
			argOther:   []string{"4"},
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			defer goleak.VerifyNone(t)
			var cmd *command
			loggers, err := logs.CreateStdLogger("Test")
			require.Nil(t, err)

			if platform.IsWindows() {
				cmd = newCommand(loggers, test.cmdWindows, test.argWindows...)
			} else {
				cmd = newCommand(loggers, test.cmdOther, test.argOther...)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			wrapper := cmd.GetCmd(ctx)
			err = wrapper.Start()
			require.Nil(t, err)
			pid, err := wrapper.Pid()
			require.Nil(t, err)
			assert.NotZero(t, pid)
			err = wrapper.Start()
			require.NotNil(t, err)
			err = wrapper.Stop()
			require.Nil(t, err)
		})
	}
}
