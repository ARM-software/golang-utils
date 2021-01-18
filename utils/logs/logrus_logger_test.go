package logs

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestLogrusLogger(t *testing.T) {
	loggers, err := NewLogrusLogger(logrus.StandardLogger(), "Test")
	require.Nil(t, err)
	_testLog(t, loggers)
}
