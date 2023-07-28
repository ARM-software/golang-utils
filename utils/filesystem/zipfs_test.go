package filesystem

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
)

const zipTestFileContent = "Test names:\r\nGeorge\r\nGeoffrey\r\nGonzo"

func Test_zipFs_Exists(t *testing.T) {
	fs, zipFile, err := NewZipFileSystemFromStandardFileSystem(filepath.Join("testdata", "testunzip.zip"), NoLimits())
	require.NoError(t, err)
	defer func() {
		if zipFile != nil {
			_ = zipFile.Close()
		}
	}()
	require.NotNil(t, zipFile)
	assert.False(t, fs.Exists(faker.DomainName()))
	// FIXME: enable when issue in afero is fixed
	// assert.True(t, fs.Exists(string(filepath.Separator)))
	// assert.True(t, fs.Exists("/"))
	assert.True(t, fs.Exists("testunzip/test.txt"))
	assert.True(t, fs.Exists("testunzip/child.zip"))
	assert.True(t, fs.Exists("testunzip/ไป ไหน มา.txt"))
	require.NoError(t, zipFile.Close())
}

func Test_zipFs_Exists2(t *testing.T) {
	// using afero test zip file
	fs, zipFile, err := NewZipFileSystemFromStandardFileSystem(filepath.Join("testdata", "t.zip"), NoLimits())
	require.NoError(t, err)
	defer func() {
		if zipFile != nil {
			_ = zipFile.Close()
		}
	}()
	require.NotNil(t, zipFile)
	assert.False(t, fs.Exists(faker.DomainName()))
	assert.True(t, fs.Exists(string(filepath.Separator)))
	assert.True(t, fs.Exists("/"))

	assert.True(t, fs.Exists("testDir1"))
	assert.True(t, fs.Exists("testFile"))
	assert.True(t, fs.Exists("testDir1/testFile"))
	require.NoError(t, zipFile.Close())
}

func Test_zipFS_FileInfo(t *testing.T) {
	fs, zipFile, err := NewZipFileSystemFromStandardFileSystem(filepath.Join("testdata", "testunzip.zip"), NoLimits())
	require.NoError(t, err)
	defer func() {
		if zipFile != nil {
			_ = zipFile.Close()
		}
	}()
	require.NotNil(t, zipFile)
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

	require.NoError(t, zipFile.Close())
}

func Test_zipFs_Browsing(t *testing.T) {
	fs, zipFile, err := NewZipFileSystemFromStandardFileSystem(filepath.Join("testdata", "testunzip.zip"), NoLimits())
	require.NoError(t, err)
	defer func() {
		if zipFile != nil {
			_ = zipFile.Close()
		}
	}()
	require.NotNil(t, zipFile)
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
	require.NoError(t, zipFile.Close())
}

func Test_zipFs_Browsing2(t *testing.T) {
	zipFs, zipFile, err := NewZipFileSystemFromStandardFileSystem(filepath.Join("testdata", "t.zip"), NoLimits())
	require.NoError(t, err)
	defer func() {
		if zipFile != nil {
			_ = zipFile.Close()
		}
	}()
	require.NotNil(t, zipFile)
	var wFunc = func(path string, info fs.FileInfo, err error) error {
		fmt.Println(path)
		return nil
	}
	require.NoError(t, zipFs.Walk("/", wFunc))
	require.NoError(t, zipFile.Close())
}

func Test_zipFs_LS(t *testing.T) {
	fs, zipFile, err := NewZipFileSystemFromStandardFileSystem(filepath.Join("testdata", "t.zip"), NoLimits())
	require.NoError(t, err)
	defer func() {
		if zipFile != nil {
			_ = zipFile.Close()
		}
	}()
	require.NotNil(t, zipFile)

	files, err := fs.Ls("/")
	require.NoError(t, err)
	assert.NotZero(t, files)
	assert.Contains(t, files, "testFile")

	files, err = fs.Ls("sub/")
	require.NoError(t, err)
	assert.NotZero(t, files)
	assert.Contains(t, files, "testDir2")
	require.NoError(t, zipFile.Close())
}

func Test_zipFs_itemType(t *testing.T) {
	fs, zipFile, err := NewZipFileSystemFromStandardFileSystem(filepath.Join("testdata", "testunzip.zip"), NoLimits())
	require.NoError(t, err)
	defer func() {
		if zipFile != nil {
			_ = zipFile.Close()
		}
	}()
	require.NotNil(t, zipFile)

	isFile, err := fs.IsFile("unzip")
	require.NoError(t, err)
	assert.False(t, isFile)
	// FIXME: Enable when issue in afero is fixed
	// isDir, err := fs.IsDir("unzip")
	// require.NoError(t, err)
	// assert.True(t, isDir)
	isFile, err = fs.IsFile("testunzip/test.txt")
	require.NoError(t, err)
	assert.True(t, isFile)
	isDir, err := fs.IsDir("testunzip/test.txt")
	require.NoError(t, err)
	assert.False(t, isDir)
	require.NoError(t, zipFile.Close())
}

func Test_zipFs_itemType2(t *testing.T) {
	fs, zipFile, err := NewZipFileSystemFromStandardFileSystem(filepath.Join("testdata", "t.zip"), NoLimits())
	require.NoError(t, err)
	defer func() {
		if zipFile != nil {
			_ = zipFile.Close()
		}
	}()
	require.NotNil(t, zipFile)

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
	require.NoError(t, zipFile.Close())
}

func Test_zipFs_Read(t *testing.T) {
	fs, zipFile, err := NewZipFileSystemFromStandardFileSystem(filepath.Join("testdata", "testunzip.zip"), NoLimits())
	require.NoError(t, err)
	defer func() {
		if zipFile != nil {
			_ = zipFile.Close()
		}
	}()
	require.NotNil(t, zipFile)

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

	require.NoError(t, zipFile.Close())
}

func Test_zipFs_not_supported(t *testing.T) {
	fs, zipFile, err := NewZipFileSystemFromStandardFileSystem(filepath.Join("testdata", "testunzip.zip"), NoLimits())
	require.NoError(t, err)
	defer func() {
		if zipFile != nil {
			_ = zipFile.Close()
		}
	}()
	require.NotNil(t, zipFile)

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

	require.NoError(t, zipFile.Close())

}
