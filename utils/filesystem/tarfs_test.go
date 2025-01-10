package filesystem

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
)

const tarTestFileContent = "Test names:\r\nGeorge\r\nGeoffrey\r\nGonzo"

var (
	aferoTarTestContentTree = []string{
		string(globalFileSystem.PathSeparator()),
		filepath.Join("/", "sub"),
		filepath.Join("/", "sub", "testDir2"),
		filepath.Join("/", "sub", "testDir2", "testFile"),
		filepath.Join("/", "testDir1"),
		filepath.Join("/", "testDir1", "testFile"),
		filepath.Join("/", "testFile"),
	}
)

func Test_tarFs_Close(t *testing.T) {
	fs, zipFile, err := NewTarFileSystemFromStandardFileSystem(filepath.Join("testdata", "testuntar.tar"), NoLimits())
	require.NoError(t, err)
	defer func() {
		if zipFile != nil {
			_ = zipFile.Close()
		}
	}()
	require.NotNil(t, zipFile)
	_, err = fs.Stat("testuntar/test.txt")
	assert.NoError(t, err)
	require.NoError(t, fs.Close())
	_, err = fs.Stat("testuntar/test.txt")
	errortest.AssertErrorDescription(t, err, "closed")
	require.NoError(t, fs.Close())
	require.NoError(t, fs.Close())
}

func Test_tarFs_Exists(t *testing.T) {
	fs, _, err := NewTarFileSystemFromStandardFileSystem(filepath.Join("testdata", "testuntar.tar"), NoLimits())
	require.NoError(t, err)
	defer func() { _ = fs.Close() }()

	assert.False(t, fs.Exists(faker.DomainName()))
	// FIXME: enable when issue in afero is fixed https://github.com/spf13/afero/issues/395
	// assert.True(t, fs.Exists(string(filepath.Separator)))
	// assert.True(t, fs.Exists("/"))
	assert.True(t, fs.Exists("testuntar/test.txt"))
	assert.True(t, fs.Exists("testuntar/child.zip"))
	assert.True(t, fs.Exists("testuntar/ไป ไหน มา.txt"))
	require.NoError(t, fs.Close())
}

func Test_tarFs_Exists_usingAferoTestTar(t *testing.T) {
	// using afero test zip file
	fs, _, err := NewTarFileSystemFromStandardFileSystem(filepath.Join("testdata", "t.tar"), NoLimits())
	require.NoError(t, err)
	defer func() { _ = fs.Close() }()

	assert.False(t, fs.Exists(faker.DomainName()))
	assert.True(t, fs.Exists(string(filepath.Separator)))
	assert.True(t, fs.Exists("/"))

	assert.True(t, fs.Exists("testDir1"))
	assert.True(t, fs.Exists("testFile"))
	assert.True(t, fs.Exists("testDir1/testFile"))
	require.NoError(t, fs.Close())
}

func Test_tarFs_FileInfo(t *testing.T) {
	fs, _, err := NewTarFileSystemFromStandardFileSystem(filepath.Join("testdata", "testuntar.tar"), NoLimits())
	require.NoError(t, err)
	defer func() { _ = fs.Close() }()

	zfile, err := fs.Stat("/")
	require.NoError(t, err)
	assert.Equal(t, string(filepath.Separator), zfile.Name())
	assert.True(t, zfile.IsDir())
	assert.Zero(t, zfile.Size())

	zfile, err = fs.Stat("testuntar/test.txt")
	require.NoError(t, err)
	assert.Equal(t, "test.txt", zfile.Name())
	assert.False(t, zfile.IsDir())
	assert.NotZero(t, zfile.Size())

	require.NoError(t, fs.Close())
}

func Test_tarFs_Browsing(t *testing.T) {
	fs, _, err := NewTarFileSystemFromStandardFileSystem(filepath.Join("testdata", "testuntar.tar"), NoLimits())
	require.NoError(t, err)
	defer func() { _ = fs.Close() }()

	empty, err := fs.IsEmpty(faker.DomainName())
	require.NoError(t, err)
	assert.True(t, empty)
	empty, err = fs.IsEmpty("testuntar/test.txt")
	require.NoError(t, err)
	assert.False(t, empty)
	empty, err = fs.IsEmpty("testuntar/child.zip")
	require.NoError(t, err)
	assert.False(t, empty)
	empty, err = fs.IsEmpty("testuntar/ไป ไหน มา.txt")
	require.NoError(t, err)
	assert.True(t, empty)
	require.NoError(t, fs.Close())
}

func Test_tarFs_Browsing_usingAferoTestTar(t *testing.T) {
	tarFs, _, err := NewTarFileSystemFromStandardFileSystem(filepath.Join("testdata", "t.tar"), NoLimits())
	require.NoError(t, err)
	defer func() { _ = tarFs.Close() }()

	// Warning: this assumes the walk function is executed in the same goroutine and not concurrently.
	// If not, this list should be created with some thread access protection in place.
	pathList := []string{}

	var wFunc = func(path string, info fs.FileInfo, err error) error {
		pathList = append(pathList, path)
		return nil
	}

	require.NoError(t, tarFs.Walk("/", wFunc))
	require.NoError(t, tarFs.Close())

	sort.Strings(pathList)
	sort.Strings(aferoTarTestContentTree)
	assert.Equal(t, aferoTarTestContentTree, pathList)
}

func Test_tarFs_LS(t *testing.T) {
	fs, _, err := NewTarFileSystemFromStandardFileSystem(filepath.Join("testdata", "t.tar"), NoLimits())
	require.NoError(t, err)
	defer func() { _ = fs.Close() }()

	files, err := fs.Ls("/")
	require.NoError(t, err)
	assert.NotZero(t, files)
	assert.Contains(t, files, "testFile")

	files, err = fs.Ls("sub/")
	require.NoError(t, err)
	assert.NotZero(t, files)
	assert.Contains(t, files, "testDir2")
	require.NoError(t, fs.Close())
}

func Test_tarFs_itemType(t *testing.T) {
	fs, _, err := NewTarFileSystemFromStandardFileSystem(filepath.Join("testdata", "testuntar.tar"), NoLimits())
	require.NoError(t, err)
	defer func() { _ = fs.Close() }()

	isFile, err := fs.IsFile("unzip")
	require.NoError(t, err)
	assert.False(t, isFile)
	// FIXME: Enable when issue in afero is fixed https://github.com/spf13/afero/issues/395
	// isDir, err := fs.IsDir("unzip")
	// require.NoError(t, err)
	// assert.True(t, isDir)
	isFile, err = fs.IsFile("testuntar/test.txt")
	require.NoError(t, err)
	assert.True(t, isFile)
	isDir, err := fs.IsDir("testuntar/test.txt")
	require.NoError(t, err)
	assert.False(t, isDir)
	require.NoError(t, fs.Close())
}

func Test_tarFs_itemType_usingAferoTestTar(t *testing.T) {
	fs, _, err := NewTarFileSystemFromStandardFileSystem(filepath.Join("testdata", "t.tar"), NoLimits())
	require.NoError(t, err)
	defer func() { _ = fs.Close() }()

	isFile, err := fs.IsFile("testDir1")
	require.NoError(t, err)
	assert.False(t, isFile)
	isDir, err := fs.IsDir("testDir1")
	require.NoError(t, err)
	assert.True(t, isDir)
	isFile, err = fs.IsFile("testDir1/testFile")
	require.NoError(t, err)
	assert.True(t, isFile)
	isDir, err = fs.IsDir("testDir1/testFile")
	require.NoError(t, err)
	assert.False(t, isDir)
	require.NoError(t, fs.Close())
}

func Test_tarFs_Read(t *testing.T) {
	fs, _, err := NewTarFileSystemFromStandardFileSystem(filepath.Join("testdata", "testuntar.tar"), NoLimits())
	require.NoError(t, err)
	defer func() { _ = fs.Close() }()

	c, err := fs.ReadFile("testuntar/test.txt")
	require.NoError(t, err)
	assert.Equal(t, tarTestFileContent, string(c))

	require.NoError(t, fs.Close())
}

func Test_tarFs_not_supported(t *testing.T) {
	fs, _, err := NewTarFileSystemFromStandardFileSystem(filepath.Join("testdata", "testuntar.tar"), NoLimits())
	require.NoError(t, err)
	defer func() { _ = fs.Close() }()

	_, err = fs.TempDir("testdata", "aaaa")
	errortest.AssertErrorDescription(t, err, "operation not permitted")

	f, err := fs.OpenFile("testuntar/test.txt", os.O_RDWR, os.FileMode(0600))
	defer func() {
		if f != nil {
			_ = f.Close()
		}
	}()
	require.Error(t, err)
	errortest.AssertErrorDescription(t, err, "operation not permitted")

	err = fs.Chmod("testuntar/test.txt", os.FileMode(0600))
	require.Error(t, err)
	errortest.AssertErrorDescription(t, err, "operation not permitted")

	require.NoError(t, fs.Close())

}
