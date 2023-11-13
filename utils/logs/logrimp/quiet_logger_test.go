package logrimp

import (
	"os"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/go-logr/logr"
	"github.com/hashicorp/go-hclog"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/exp/slog"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

func TestQuietLoggerImplementations(t *testing.T) {
	zl, err := zap.NewDevelopment()
	require.NoError(t, err)
	tests := []struct {
		Logger logr.Logger
		name   string
	}{
		{
			Logger: NewNoopLogger(),
			name:   "NoOp",
		},
		{
			Logger: NewStdOutLogr(),
			name:   "Standard Output",
		},
		{
			Logger: NewZapLogger(zl),
			name:   "Zap",
		},
		{
			Logger: NewHclogLogger(hclog.New(nil)),
			name:   "HClog",
		},
		{
			Logger: NewLogrusLogger(logrus.New()),
			name:   "Logrus",
		},
		{
			Logger: NewSlogLogger(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))),
			name:   "slog",
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			logger := NewQuietLogger(test.Logger)
			logger.WithName(faker.Name()).WithValues("foo", "bar").Info(faker.Sentence())
			logger.Error(commonerrors.ErrUnexpected, faker.Sentence(), faker.Word(), faker.Name())
		},
		)
	}

}
