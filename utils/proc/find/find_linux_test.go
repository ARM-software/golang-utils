//go:build linux

package find

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/logs"
	"github.com/ARM-software/golang-utils/utils/logs/logstest"
	"github.com/ARM-software/golang-utils/utils/subprocess"
)

func TestFind(t *testing.T) {
	for _, test := range []struct {
		name      string
		processes int
	}{
		{
			name:      "One process",
			processes: 1,
		},
		{
			name:      "Many processes",
			processes: 10,
		},
		{
			name:      "No process",
			processes: 0,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			processString := faker.Sentence()
			for range test.processes {
				l, err := logs.NewLogrLogger(logstest.NewStdTestLogger(), test.name)
				require.NoError(t, err)

				cmd, err := subprocess.New(ctx, l, "start", "success", "failed", "sh", "-c", fmt.Sprintf("sleep 10 ; echo '%v'", processString))
				require.NoError(t, err)

				err = cmd.Start()
				require.NoError(t, err)
			}

			processes, err := FindProcessByName(ctx, processString)
			assert.NoError(t, err)
			assert.Len(t, processes, test.processes)

			// stopping processes shows they were parsed correctly
			for _, process := range processes {
				err = process.Terminate(ctx)
				require.NoError(t, err)
			}
			processes, err = FindProcessByName(ctx, processString)
			require.NoError(t, err)
			assert.Empty(t, processes)
		})
	}

	t.Run("Cancel context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		processString := faker.Sentence()

		l, err := logs.NewLogrLogger(logstest.NewStdTestLogger(), "context cancelled")
		require.NoError(t, err)

		cmd, err := subprocess.New(ctx, l, "start", "success", "failed", "sh", "-c", fmt.Sprintf("sleep 10 ; echo '%v'", processString))
		require.NoError(t, err)

		err = cmd.Start()
		require.NoError(t, err)
		cancel()

		processes, err := FindProcessByName(ctx, processString)
		errortest.AssertError(t, err, commonerrors.ErrCancelled)
		assert.Empty(t, processes)
	})

}
