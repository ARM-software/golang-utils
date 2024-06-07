package filesystem

import (
	"io"
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

const zipTestFileContent = "Test names:\r\nGeorge\r\nGeoffrey\r\nGonzo"

var (
	aferoTestZipContentTree = []string{
		string(globalFileSystem.PathSeparator()),
		filepath.Join("/", "sub"),
		filepath.Join("/", "sub", "testDir2"),
		filepath.Join("/", "sub", "testDir2", "testFile"),
		filepath.Join("/", "testDir1"),
		filepath.Join("/", "testDir1", "testFile"),
		filepath.Join("/", "testFile"),
	}
)

func Test_zipFs_Close(t *testing.T) {
	fs, zipFile, err := NewZipFileSystemFromStandardFileSystem(filepath.Join("testdata", "testunzip.zip"), NoLimits())
	require.NoError(t, err)
	defer func() {
		if zipFile != nil {
			_ = zipFile.Close()
		}
	}()
	require.NotNil(t, zipFile)
	_, err = fs.Stat("testunzip/test.txt")
	assert.NoError(t, err)
	require.NoError(t, fs.Close())
	_, err = fs.Stat("testunzip/test.txt")
	errortest.AssertErrorDescription(t, err, "closed")
	require.NoError(t, fs.Close())
	require.NoError(t, fs.Close())
}

func Test_zipFs_Exists(t *testing.T) {
	fs, _, err := NewZipFileSystemFromStandardFileSystem(filepath.Join("testdata", "testunzip.zip"), NoLimits())
	require.NoError(t, err)
	defer func() { _ = fs.Close() }()

	assert.False(t, fs.Exists(faker.DomainName()))
	// FIXME: enable when issue in afero is fixed https://github.com/spf13/afero/issues/395
	// assert.True(t, fs.Exists(string(filepath.Separator)))
	// assert.True(t, fs.Exists("/"))
	assert.True(t, fs.Exists("testunzip/test.txt"))
	assert.True(t, fs.Exists("testunzip/child.zip"))
	assert.True(t, fs.Exists("testunzip/ไป ไหน มา.txt"))
	require.NoError(t, fs.Close())
}

func Test_zipFs_Exists_usingAferoTestZip(t *testing.T) {
	// using afero test zip file
	fs, _, err := NewZipFileSystemFromStandardFileSystem(filepath.Join("testdata", "t.zip"), NoLimits())
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

func Test_zipFS_FileInfo(t *testing.T) {
	fs, _, err := NewZipFileSystemFromStandardFileSystem(filepath.Join("testdata", "testunzip.zip"), NoLimits())
	require.NoError(t, err)
	defer func() { _ = fs.Close() }()

	zfile, err := fs.Stat("/")
	require.NoError(t, err)
	assert.Equal(t, string(filepath.Separator), zfile.Name())
	assert.True(t, zfile.IsDir())
	assert.Zero(t, zfile.Size())

	zfile, err = fs.Stat("testunzip/test.txt")
	require.NoError(t, err)
	assert.Equal(t, "test.txt", zfile.Name())
	assert.False(t, zfile.IsDir())
	assert.NotZero(t, zfile.Size())

	require.NoError(t, fs.Close())
}

func Test_zipFs_Browsing(t *testing.T) {
	fs, _, err := NewZipFileSystemFromStandardFileSystem(filepath.Join("testdata", "testunzip.zip"), NoLimits())
	require.NoError(t, err)
	defer func() { _ = fs.Close() }()

	empty, err := fs.IsEmpty(faker.DomainName())
	require.NoError(t, err)
	assert.True(t, empty)
	empty, err = fs.IsEmpty("testunzip/test.txt")
	require.NoError(t, err)
	assert.False(t, empty)
	empty, err = fs.IsEmpty("testunzip/child.zip")
	require.NoError(t, err)
	assert.False(t, empty)
	empty, err = fs.IsEmpty("testunzip/ไป ไหน มา.txt")
	require.NoError(t, err)
	assert.True(t, empty)
	require.NoError(t, fs.Close())
}

func Test_zipFs_Browsing_usingAferoTestZip(t *testing.T) {
	zipFs, _, err := NewZipFileSystemFromStandardFileSystem(filepath.Join("testdata", "t.zip"), NoLimits())
	require.NoError(t, err)
	defer func() { _ = zipFs.Close() }()

	// Warning: this assumes the walk function is executed in the same goroutine and not concurrently.
	// If not, this list should be created with some thread access protection in place.
	pathList := []string{}

	var wFunc = func(path string, info fs.FileInfo, err error) error {
		pathList = append(pathList, path)
		return nil
	}

	require.NoError(t, zipFs.Walk("/", wFunc))
	require.NoError(t, zipFs.Close())

	sort.Strings(pathList)
	sort.Strings(aferoTestZipContentTree)
	assert.Equal(t, aferoTestZipContentTree, pathList)
}

func Test_zipFs_LS(t *testing.T) {
	fs, _, err := NewZipFileSystemFromStandardFileSystem(filepath.Join("testdata", "t.zip"), NoLimits())
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

func Test_zipFs_itemType(t *testing.T) {
	fs, _, err := NewZipFileSystemFromStandardFileSystem(filepath.Join("testdata", "testunzip.zip"), NoLimits())
	require.NoError(t, err)
	defer func() { _ = fs.Close() }()

	isFile, err := fs.IsFile("unzip")
	require.NoError(t, err)
	assert.False(t, isFile)
	// FIXME: Enable when issue in afero is fixed https://github.com/spf13/afero/issues/395
	// isDir, err := fs.IsDir("unzip")
	// require.NoError(t, err)
	// assert.True(t, isDir)
	isFile, err = fs.IsFile("testunzip/test.txt")
	require.NoError(t, err)
	assert.True(t, isFile)
	isDir, err := fs.IsDir("testunzip/test.txt")
	require.NoError(t, err)
	assert.False(t, isDir)
	require.NoError(t, fs.Close())
}

func Test_zipFs_itemType_usingAferoTestZip(t *testing.T) {
	fs, _, err := NewZipFileSystemFromStandardFileSystem(filepath.Join("testdata", "t.zip"), NoLimits())
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

func Test_zipFs_Read(t *testing.T) {
	fs, _, err := NewZipFileSystemFromStandardFileSystem(filepath.Join("testdata", "testunzip.zip"), NoLimits())
	require.NoError(t, err)
	defer func() { _ = fs.Close() }()

	t.Run("using file opening", func(t *testing.T) {
		f, err := fs.GenericOpen("testunzip/test.txt")
		require.NoError(t, err)
		defer func() { _ = f.Close() }()
		c, err := io.ReadAll(f)
		require.NoError(t, err)
		assert.Equal(t, zipTestFileContent, string(c))
		require.NoError(t, f.Close())
	})

	t.Run("using file opening 2", func(t *testing.T) {
		f, err := fs.OpenFile("testunzip/test.txt", os.O_RDONLY, os.FileMode(0600))
		require.NoError(t, err)
		defer func() { _ = f.Close() }()
		c, err := io.ReadAll(f)
		require.NoError(t, err)
		assert.Equal(t, zipTestFileContent, string(c))
		require.NoError(t, f.Close())
	})

	t.Run("using file read", func(t *testing.T) {
		c, err := fs.ReadFile("testunzip/test.txt")
		require.NoError(t, err)
		assert.Equal(t, zipTestFileContent, string(c))
	})

	require.NoError(t, fs.Close())
}

func Test_zipFs_not_supported(t *testing.T) {
	fs, _, err := NewZipFileSystemFromStandardFileSystem(filepath.Join("testdata", "testunzip.zip"), NoLimits())
	require.NoError(t, err)
	defer func() { _ = fs.Close() }()

	_, err = fs.TempDir("testdata", "aaaa")
	errortest.AssertErrorDescription(t, err, "operation not permitted")

	f, err := fs.OpenFile("testunzip/test.txt", os.O_RDWR, os.FileMode(0600))
	defer func() {
		if f != nil {
			_ = f.Close()
		}
	}()
	require.Error(t, err)
	errortest.AssertErrorDescription(t, err, "operation not permitted")

	err = fs.Chmod("testunzip/test.txt", os.FileMode(0600))
	require.Error(t, err)
	errortest.AssertErrorDescription(t, err, "operation not permitted")

	require.NoError(t, fs.Close())

}
