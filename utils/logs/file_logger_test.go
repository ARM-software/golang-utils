package logs

import (
	"testing"

	"github.com/ARMmbed/build-service-common/common/filesystem"

	"github.com/stretchr/testify/require"
)

func TestFileLogger(t *testing.T) {
	file, err := filesystem.TempFileInTempDir("test-filelog-*.log")
	require.Nil(t, err)

	err = file.Close()
	require.Nil(t, err)
	defer func() { _ = filesystem.Rm(file.Name()) }()

	loggers, err := CreateFileLogger(file.Name(), "Test")
	require.Nil(t, err)

	_testLog(t, loggers)
	err = filesystem.Rm(file.Name())
	require.Nil(t, err)
}
