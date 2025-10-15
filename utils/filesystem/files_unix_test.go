//go:build !windows
// +build !windows

package filesystem

import (
	"fmt"
	"net"
	"path/filepath"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"
)

func TestIsUnixFile(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprint(fsType), func(t *testing.T) {
			fs := NewFs(fsType)

			tmpDir := t.TempDir()

			t.Run("normal file", func(t *testing.T) {
				filePath := filepath.Join(tmpDir, faker.Word())
				err := fs.Touch(filePath)
				require.NoError(t, err)
				b, err := fs.IsFile(filePath)
				require.NoError(t, err)
				assert.True(t, b)
			})

			t.Run("special file", func(t *testing.T) {
				if fsType == InMemoryFS {
					t.Skip("In-memory file system won't have hardware devices or special files")
				}

				b, err := fs.IsFile("/dev/null")
				require.NoError(t, err)
				assert.True(t, b)

				fifoPath := filepath.Join(tmpDir, faker.Word())
				require.NoError(t, err)
				defer func() { _ = fs.Rm(fifoPath) }()
				err = unix.Mkfifo(fifoPath, 0666)
				require.NoError(t, err)
				b, err = fs.IsFile(fifoPath)
				require.NoError(t, err)
				assert.True(t, b)
				err = fs.Rm(fifoPath)
				require.NoError(t, err)

				socketPath := filepath.Join(tmpDir, faker.Word())
				require.NoError(t, err)
				defer func() { _ = fs.Rm(socketPath) }()
				l, err := net.Listen("unix", socketPath)
				require.NoError(t, err)
				defer func() { _ = l.Close() }()
				b, err = fs.IsFile(socketPath)
				require.NoError(t, err)
				assert.True(t, b)
				err = fs.Rm(socketPath)
				require.NoError(t, err)
			})
		})
	}
}
