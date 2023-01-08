package sizetest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/filesystem"
	"github.com/ARM-software/golang-utils/utils/units/size"
)

// If it works for M and K it will most likely work for the larger ones as they are just multiplying by 1000/1024

var testCases = []float64{
	1 * size.B, 12 * size.B, 125 * size.B, 988 * size.B, 67 * size.B,
	1 * size.KB, 102 * size.KB, 12.5 * size.KB, 98 * size.KB, 679 * size.KB,
	1 * size.MB, 77 * size.MB, 5 * size.MB, 188 * size.MB, 617 * size.MB,
	1011 * size.KiB, 2 * size.KiB, 1000 * size.KiB, 2000 * size.KiB, 1111 * size.KiB,
	27 * size.MiB, 2 * size.MiB, 76 * size.MiB, 22 * size.MiB, 0.7 * size.MiB,
}

func TestGetFileSize(t *testing.T) {
	fs := filesystem.NewFs(filesystem.InMemoryFS)

	for i := range testCases {
		tmpFile, err := fs.TempFileInTempDir("test-filesize-")
		require.NoError(t, err)
		data := make([]byte, int64(testCases[i]))
		_, _ = tmpFile.Write(data)

		err = tmpFile.Close()
		require.NoError(t, err)

		fileName := tmpFile.Name()
		defer func() { _ = fs.Rm(fileName) }()

		size, err := fs.GetFileSize(tmpFile.Name())
		require.NoError(t, err)
		assert.Equal(t, int64(testCases[i]), size)
	}
}
