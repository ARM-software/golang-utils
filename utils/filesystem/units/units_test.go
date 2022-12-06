package units

import (
	"testing"

	"github.com/ARM-software/golang-utils/utils/filesystem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// If it works for M and K it will probaby work for the larger ones as they are just multiplying by 1000/1024

var testCases = []float64{
	1 * B, 12 * B, 125 * B, 988 * B, 67 * B,
	1 * KB, 102 * KB, 12.5 * KB, 98 * KB, 679 * KB,
	1 * MB, 77 * MB, 5 * MB, 188 * MB, 617 * MB,
	1011 * KiB, 2 * KiB, 1000 * KiB, 2000 * KiB, 1111 * KiB,
	27 * MiB, 2 * MiB, 76 * MiB, 22 * MiB, 0.7 * MiB,
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
