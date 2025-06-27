package filesystem

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// GenerateTestFileTree generates a file tree for testing purposes and returns a list of all the files and filesystem items created.
// testDir corresponds to the folder where the tree is created
// basePath corresponds to the base path for symlinks
// fileModTime, fileAccessTime are for specifying created files ch times
func GenerateTestFileTree(t *testing.T, fs FS, testDir, basePath string, withLinks bool, fileModTime time.Time, fileAccessTime time.Time) []string {
	t.Helper()
	err := fs.MkDir(testDir)
	require.NoError(t, err)

	var sLinks []string
	random := rand.New(rand.NewSource(time.Now().Unix()))                      //nolint:gosec //causes G404: Use of weak random number generator (math/rand instead of crypto/rand) (gosec), So disable gosec
	for i := 0; i < int(math.Max(float64(1), float64(random.Intn(10)))); i++ { //nolint:gosec //causes G404: Use of weak random number generator (math/rand instead of crypto/rand) (gosec), So disable gosec
		c := fmt.Sprint("test", i+1)
		path := FilePathJoin(fs, testDir, c)

		err := fs.MkDir(path)
		require.NoError(t, err)

		for j := 0; j < int(math.Max(float64(1), float64(random.Intn(10)))); j++ { //nolint:gosec //causes G404: Use of weak random number generator (math/rand instead of crypto/rand) (gosec), So disable gosec
			c := fmt.Sprint("test", j+1)
			path := FilePathJoin(fs, path, c)

			err := fs.MkDir(path)
			require.NoError(t, err)

			if withLinks {
				if len(sLinks) > 0 {
					c1 := fmt.Sprint("link", j+1)
					c2 := FilePathJoin(fs, path, c1)
					err = fs.Symlink(sLinks[0], c2)
					require.NoError(t, err)
					if len(sLinks) > 1 {
						sLinks = sLinks[1:]
					} else {
						sLinks = nil
					}
				}
			}

			for k := 0; k < int(math.Max(float64(1), float64(random.Intn(10)))); k++ { //nolint:gosec //causes G404: Use of weak random number generator (math/rand instead of crypto/rand) (gosec), So disable gosec
				c := fmt.Sprint("test", k+1, ".txt")
				finalPath := FilePathJoin(fs, path, c)

				// pick a couple of files to make symlinks (1 in 10)
				r := random.Intn(10) //nolint:gosec //causes G404: Use of weak random number generator (math/rand instead of crypto/rand) (gosec), So disable gosec
				if r == 4 {
					fPath := FilePathJoin(fs, basePath, path, c)
					sLinks = append(sLinks, fPath)
				}

				s := fmt.Sprint("file ", i+1, j+1, k+1)
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

func NewTestFilesystem(t *testing.T, pathSeparator rune) FS {
	t.Helper()
	return NewVirtualFileSystemWithPathSeparator(afero.NewMemMapFs(), InMemoryFS, IdentityPathConverterFunc, pathSeparator)
}

func NewTestOSFilesystem(t *testing.T, pathSeparator rune) FS {
	t.Helper()
	return NewVirtualFileSystemWithPathSeparator(afero.NewOsFs(), StandardFS, IdentityPathConverterFunc, pathSeparator)
}
