package filesystemtest

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/filesystem"
	"github.com/ARM-software/golang-utils/utils/idgen"
)

// CreateTestFileTree generates a randomised tree of directories and files under the specified `testDir` in the provided filesystem and returns a slice of all created paths.
func CreateTestFileTree(t *testing.T, fs filesystem.FS, testDir string, fileModTime time.Time, fileAccessTime time.Time) []string {
	t.Helper()
	// This can be fixed for testing.
	random := rand.New(rand.NewSource(time.Now().Unix())) //nolint:gosec //causes G404: Use of weak random number generator (math/rand instead of crypto/rand) (gosec), So disable gosec as this is just for
	err := fs.MkDir(testDir)
	require.NoError(t, err)

	randI := random.Intn(5) //nolint:gosec //causes G404: Use of weak random number generator (math/rand instead of crypto/rand) (gosec), So disable gosec
	if randI == 0 {
		randI = 1
	}
	for i := 0; i < randI; i++ {
		c := fmt.Sprintf("test%v", i+1)
		path := filepath.Join(testDir, c)

		err = fs.MkDir(path)
		require.NoError(t, err)

		for j := 0; j < random.Intn(5); j++ { //nolint:gosec //causes G404: Use of weak random number generator (math/rand instead of crypto/rand) (gosec), So disable gosec
			uuid, err := idgen.GenerateUUID4()
			if err != nil {
				uuid = "uuid"
			}
			c := fmt.Sprintf("test-%v-%v", uuid, j+1)
			path := filepath.Join(path, c)

			err = fs.MkDir(path)
			require.NoError(t, err)

			for k := 0; k < random.Intn(5); k++ { //nolint:gosec //causes G404: Use of weak random number generator (math/rand instead of crypto/rand) (gosec), So disable gosec
				uuid, err = idgen.GenerateUUID4()
				if err != nil {
					uuid = "uuid"
				}
				c := fmt.Sprintf("test-%v-%v%v", uuid, k+1, ".txt")
				finalPath := filepath.Join(path, c)

				s := fmt.Sprintf("file-%v-%v%v%v ", uuid, i+1, j+1, k+1)
				err = fs.WriteFile(finalPath, []byte(s), 0755)
				require.NoError(t, err)
			}
		}
	}
	var tree []string
	err = fs.ListDirTree(testDir, &tree)
	require.NoError(t, err)

	// unifying timestamps
	for _, path := range tree {
		err = fs.Chtimes(path, fileAccessTime, fileModTime)
		require.NoError(t, err)
	}

	return tree
}

// GenerateTestFile generates a file and writes zero-filled blocks until it reaches desiredSizeInBytes.
func GenerateTestFile(t *testing.T, fs filesystem.FS, filePath string, desiredSizeInBytes int, blockSizeInBytes int) {
	t.Helper()

	file, err := fs.CreateFile(filePath)
	require.NoError(t, err)

	defer func() {
		err = file.Close()
		require.NoError(t, err)
	}()

	data := make([]byte, blockSizeInBytes)
	var currentSize int
	for currentSize < desiredSizeInBytes {
		size, err := file.Write(data)
		require.NoError(t, err)

		currentSize += size
	}
}
