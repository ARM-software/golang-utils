package filesystem

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
)

//go:embed *
var testContent embed.FS

const testFile1Content = "this is a text file with some content"

func Test_embedFS_Exists(t *testing.T) {
	fs, err := NewEmbedFileSystem(&testContent)
	require.NoError(t, err)
	assert.False(t, fs.Exists(faker.DomainName()))
	assert.True(t, fs.Exists("testdata"))
	assert.True(t, fs.Exists("testdata/embed"))
	assert.True(t, fs.Exists("testdata/embed/level1/test.txt"))
}

func Test_embedFS_Browsing(t *testing.T) {
	fs, err := NewEmbedFileSystem(&testContent)
	require.NoError(t, err)
	empty, err := fs.IsEmpty(faker.DomainName())
	require.NoError(t, err)
	assert.True(t, empty)
	empty, err = fs.IsEmpty("testdata")
	require.NoError(t, err)
	assert.False(t, empty)
	empty, err = fs.IsEmpty("testdata/embed")
	require.NoError(t, err)
	assert.False(t, empty)
	empty, err = fs.IsEmpty("testdata/embed/level1")
	require.NoError(t, err)
	assert.False(t, empty)
}

func Test_embedFS_Browsing2(t *testing.T) {
	efs, err := NewEmbedFileSystem(&testContent)
	require.NoError(t, err)
	var wFunc = func(path string, info fs.FileInfo, err error) error {
		fmt.Println(path)
		return nil
	}
	require.NoError(t, efs.Walk("testdata/embed", wFunc))
}

func Test_embedFS_LS(t *testing.T) {
	efs, err := NewEmbedFileSystem(&testContent)
	require.NoError(t, err)

	files, err := efs.Ls("testdata")
	require.NoError(t, err)
	assert.NotZero(t, files)
	assert.Contains(t, files, "embed")

	files, err = efs.Ls("testdata/embed")
	require.NoError(t, err)
	assert.NotZero(t, files)
	assert.Contains(t, files, "level1")
	assert.Contains(t, files, "test.txt")
}

func Test_embedFS_itemType(t *testing.T) {
	efs, err := NewEmbedFileSystem(&testContent)
	require.NoError(t, err)
	isFile, err := efs.IsFile("testdata")
	require.NoError(t, err)
	assert.False(t, isFile)
	isDir, err := efs.IsDir("testdata")
	require.NoError(t, err)
	assert.True(t, isDir)
	isFile, err = efs.IsFile("testdata/embed/level1/test.txt")
	require.NoError(t, err)
	assert.True(t, isFile)
	isDir, err = efs.IsDir("testdata/embed/level1/test.txt")
	require.NoError(t, err)
	assert.False(t, isDir)
}

func Test_embedFS_Read(t *testing.T) {
	efs, err := NewEmbedFileSystem(&testContent)
	require.NoError(t, err)

	t.Run("embed read", func(t *testing.T) {
		c, err := testContent.ReadFile("testdata/embed/test.txt")
		require.NoError(t, err)
		assert.Contains(t, string(c), testFile1Content)
	})

	t.Run("using file opening", func(t *testing.T) {
		f, err := efs.GenericOpen("testdata/embed/test.txt")
		require.NoError(t, err)
		defer func() { _ = f.Close() }()
		c, err := io.ReadAll(f)
		require.NoError(t, err)
		assert.Contains(t, string(c), testFile1Content)
		require.NoError(t, f.Close())
	})

	t.Run("using file opening 2", func(t *testing.T) {
		f, err := efs.OpenFile("testdata/embed/test.txt", os.O_RDONLY, os.FileMode(0600))
		require.NoError(t, err)
		defer func() { _ = f.Close() }()
		c, err := io.ReadAll(f)
		require.NoError(t, err)
		assert.Contains(t, string(c), testFile1Content)
		require.NoError(t, f.Close())
	})

	t.Run("using file read", func(t *testing.T) {
		c, err := efs.ReadFile("testdata/embed/test.txt")
		require.NoError(t, err)
		assert.Contains(t, string(c), testFile1Content)
	})

}

func Test_embed_not_supported(t *testing.T) {
	efs, err := NewEmbedFileSystem(nil)
	errortest.RequireError(t, err, commonerrors.ErrUndefined)
	assert.Nil(t, efs)

	efs, err = NewEmbedFileSystem(&testContent)
	require.NoError(t, err)

	_, err = efs.TempDir("testdata", "aaaa")
	errortest.AssertErrorDescription(t, err, "operation not permitted")

	f, err := efs.OpenFile("testdata/embed/test.txt", os.O_RDWR, os.FileMode(0600))
	defer func() {
		if f != nil {
			_ = f.Close()
		}
	}()
	require.Error(t, err)
	errortest.AssertErrorDescription(t, err, "operation not permitted")

	err = efs.Chmod("testdata/embed/test.txt", os.FileMode(0600))
	require.Error(t, err)
	errortest.AssertErrorDescription(t, err, "operation not permitted")
}
