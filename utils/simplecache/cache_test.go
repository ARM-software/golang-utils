package simplecache

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/filesystem"
	"github.com/ARM-software/golang-utils/utils/internal/testutils"
	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestCache_Add(t *testing.T) {
	tests := []struct {
		name    string
		srcFs   filesystem.FilesystemType
		cacheFs filesystem.FilesystemType
	}{
		{
			name:    "Add_content_to_cache_same_filesystem",
			srcFs:   filesystem.StandardFS,
			cacheFs: filesystem.StandardFS,
		},
		{
			name:    "Add_content_to_cache_different_filesystem",
			srcFs:   filesystem.StandardFS,
			cacheFs: filesystem.InMemoryFS,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer goleak.VerifyNone(t)

			srcFs := filesystem.NewFs(test.srcFs)
			tmpSrcDir, err := srcFs.TempDirInTempDir("test-cache-src")
			require.NoError(t, err)
			defer func() { _ = srcFs.Rm(tmpSrcDir) }()

			_, err = testutils.CreateTestFileTree(srcFs, tmpSrcDir, time.Now(), time.Now())
			require.NoError(t, err)
			srcContent, err := srcFs.Ls(tmpSrcDir)
			require.NoError(t, err)

			cacheFs := filesystem.NewFs(test.cacheFs)
			tmpCacheDir, err := cacheFs.TempDirInTempDir("test-cache")
			require.NoError(t, err)
			defer func() { _ = cacheFs.Rm(tmpCacheDir) }()

			ctx := context.Background()
			cache, err := NewDefaultSimpleCache(ctx, cacheFs, tmpCacheDir)
			require.NoError(t, err)

			for _, path := range srcContent {
				absoluteSrcPath := filesystem.FilePathJoin(srcFs, tmpSrcDir, path)
				err := cache.Add(ctx, faker.UUIDDigit(), srcFs, absoluteSrcPath)
				require.NoError(t, err)
			}

			var cacheTree []string
			err = cacheFs.ListDirTree(tmpCacheDir, &cacheTree)
			require.NoError(t, err)

			var srcTree []string
			err = srcFs.ListDirTree(tmpSrcDir, &srcTree)
			require.NoError(t, err)

			require.Equal(t, len(srcTree), len(cacheTree), "Cache Dir and Src Dir have differnet number of content")

			srcRelTree, err := srcFs.ConvertToRelativePath(tmpSrcDir, srcTree...)
			require.NoError(t, err)

			cacheRelTree, err := cacheFs.ConvertToRelativePath(tmpCacheDir, cacheTree...)
			require.NoError(t, err)

			slices.Sort(srcRelTree)
			slices.Sort(cacheRelTree)
			require.Equal(t, srcRelTree, cacheRelTree, "Cache Dir and Src Dir have differnet contents")

			err = cache.Close(ctx)
			require.NoError(t, err)
		})
	}

	t.Run("Add_invalid_path", func(t *testing.T) {
		defer goleak.VerifyNone(t)

		fs := filesystem.NewFs(filesystem.StandardFS)
		tmpCacheDir, err := fs.TempDirInTempDir("test-cache")
		require.NoError(t, err)
		defer func() { _ = fs.Rm(tmpCacheDir) }()

		ctx := context.Background()
		cache, err := NewDefaultSimpleCache(ctx, fs, tmpCacheDir)
		require.NoError(t, err)

		err = cache.Add(ctx, faker.UUIDDigit(), fs, "/does/not/exist")
		errortest.AssertError(t, err, commonerrors.ErrNotFound)

		err = cache.Close(ctx)
		require.NoError(t, err)
	})

	t.Run("Add_duplicate_entry", func(t *testing.T) {
		defer goleak.VerifyNone(t)

		fs := filesystem.NewFs(filesystem.StandardFS)
		tmpCacheDir, err := fs.TempDirInTempDir("test-cache")
		require.NoError(t, err)
		defer func() { _ = fs.Rm(tmpCacheDir) }()

		ctx := context.Background()
		cache, err := NewDefaultSimpleCache(ctx, fs, tmpCacheDir)
		require.NoError(t, err)

		tmpTestPath, err := fs.TouchTempFileInTempDir(faker.Word())
		require.NoError(t, err)

		id := faker.UUIDDigit()
		err = cache.Add(ctx, id, fs, tmpTestPath)
		require.NoError(t, err)

		err = cache.Add(ctx, id, fs, tmpTestPath)
		errortest.AssertError(t, err, commonerrors.ErrExists)

		err = cache.Close(ctx)
		require.NoError(t, err)
	})
}

func TestCache_Restore(t *testing.T) {
	t.Run("Restore_file", func(t *testing.T) {
		defer goleak.VerifyNone(t)

		fs := filesystem.NewFs(filesystem.StandardFS)
		tmpCacheDir, err := fs.TempDirInTempDir("test-cache")
		require.NoError(t, err)
		defer func() { _ = fs.Rm(tmpCacheDir) }()

		ctx := context.Background()
		cache, err := NewDefaultSimpleCache(ctx, fs, tmpCacheDir)
		require.NoError(t, err)

		tmpfilePath, err := fs.TouchTempFileInTempDir(faker.Word())
		require.NoError(t, err)

		id := faker.UUIDDigit()
		err = cache.Add(ctx, id, fs, tmpfilePath)
		require.NoError(t, err)

		err = fs.Rm(tmpfilePath)
		require.NoError(t, err)

		err = cache.Restore(ctx, id)
		require.NoError(t, err)

		require.True(t, filesystem.Exists(tmpfilePath), "cache did not restore the file")
		err = cache.Close(ctx)
		require.NoError(t, err)
	})

	t.Run("Restore_file_non_existent", func(t *testing.T) {
		defer goleak.VerifyNone(t)

		fs := filesystem.NewFs(filesystem.StandardFS)
		tmpCacheDir, err := fs.TempDirInTempDir("test-cache")
		require.NoError(t, err)
		defer func() { _ = fs.Rm(tmpCacheDir) }()

		ctx := context.Background()
		cache, err := NewDefaultSimpleCache(ctx, fs, tmpCacheDir)
		require.NoError(t, err)

		err = cache.Restore(ctx, faker.UUIDDigit())
		errortest.AssertError(t, err, commonerrors.ErrNotFound)

		err = cache.Close(ctx)
		require.NoError(t, err)
	})

	t.Run("Restore_file_overwrite", func(t *testing.T) {
		defer goleak.VerifyNone(t)

		fs := filesystem.NewFs(filesystem.StandardFS)
		tmpCacheDir, err := fs.TempDirInTempDir("test-cache")
		require.NoError(t, err)
		defer func() { _ = fs.Rm(tmpCacheDir) }()

		ctx := context.Background()
		cache, err := NewDefaultSimpleCache(ctx, fs, tmpCacheDir)
		require.NoError(t, err)

		tmpFile, err := fs.TempFileInTempDir(faker.Word())
		require.NoError(t, err)
		tmpfilePath := tmpFile.Name()
		defer func() { _ = fs.Rm(tmpfilePath) }()

		originalFileContent := []byte("Original")
		_, err = tmpFile.Write(originalFileContent)
		require.NoError(t, err)

		id := faker.UUIDDigit()
		err = cache.Add(ctx, id, fs, tmpfilePath)
		require.NoError(t, err)

		NewFileContent := []byte("New")
		_, err = tmpFile.Write(NewFileContent)
		require.NoError(t, err)

		err = cache.Restore(ctx, id)
		require.NoError(t, err)

		cacheFileContent, err := fs.ReadFile(tmpfilePath)
		require.NoError(t, err)
		require.Equal(t, originalFileContent, cacheFileContent)

		err = cache.Close(ctx)
		require.NoError(t, err)
	})
}

func TestCache_Remove(t *testing.T) {
	t.Run("Remove_file", func(t *testing.T) {
		defer goleak.VerifyNone(t)

		fs := filesystem.NewFs(filesystem.StandardFS)
		tmpCacheDir, err := fs.TempDirInTempDir("test-cache")
		require.NoError(t, err)
		defer func() { _ = fs.Rm(tmpCacheDir) }()

		ctx := context.Background()
		cache, err := NewDefaultSimpleCache(ctx, fs, tmpCacheDir)
		require.NoError(t, err)

		tmpfilePath, err := fs.TouchTempFileInTempDir(faker.Word())
		require.NoError(t, err)
		defer func() { _ = fs.Rm(tmpfilePath) }()

		id := faker.UUIDDigit()
		err = cache.Add(ctx, id, fs, tmpfilePath)
		require.NoError(t, err)

		err = cache.Remove(ctx, id)
		require.NoError(t, err)

		cacheExists, err := cache.Contains(ctx, id)
		require.NoError(t, err)
		require.False(t, cacheExists, "cache still has an entry after removing")

		cacheFilePath := filesystem.FilePathJoin(fs, tmpCacheDir, filesystem.FilePathBase(fs, tmpfilePath))
		require.False(t, filesystem.Exists(cacheFilePath), "cache still has the file after removing")

		err = cache.Close(ctx)
		require.NoError(t, err)
	})

	t.Run("Remove_file_non_existent", func(t *testing.T) {
		defer goleak.VerifyNone(t)

		fs := filesystem.NewFs(filesystem.StandardFS)
		tmpCacheDir, err := fs.TempDirInTempDir("test-cache")
		require.NoError(t, err)
		defer func() { _ = fs.Rm(tmpCacheDir) }()

		ctx := context.Background()
		cache, err := NewDefaultSimpleCache(ctx, fs, tmpCacheDir)
		require.NoError(t, err)

		err = cache.Remove(ctx, faker.UUIDDigit())
		errortest.AssertError(t, err, commonerrors.ErrNotFound)

		err = cache.Close(ctx)
		require.NoError(t, err)
	})
}

func TestCache_Close(t *testing.T) {
	defer goleak.VerifyNone(t)

	fs := filesystem.NewFs(filesystem.StandardFS)
	tmpSrcDir, err := fs.TempDirInTempDir("test-cache-src")
	require.NoError(t, err)
	defer func() { _ = fs.Rm(tmpSrcDir) }()

	_, err = testutils.CreateTestFileTree(fs, tmpSrcDir, time.Now(), time.Now())
	require.NoError(t, err)
	srcContent, err := fs.Ls(tmpSrcDir)
	require.NoError(t, err)

	tmpCacheDir, err := fs.TempDirInTempDir("test-cache")
	require.NoError(t, err)
	defer func() { _ = fs.Rm(tmpCacheDir) }()

	ctx := context.Background()
	cache, err := NewDefaultSimpleCache(ctx, fs, tmpCacheDir)
	require.NoError(t, err)

	for _, path := range srcContent {
		absoluteSrcPath := filesystem.FilePathJoin(fs, tmpSrcDir, path)
		err := cache.Add(ctx, faker.UUIDDigit(), fs, absoluteSrcPath)
		require.NoError(t, err)
	}

	err = cache.Close(ctx)
	require.NoError(t, err)

	var srcTree []string
	err = fs.ListDirTree(tmpSrcDir, &srcTree)
	require.NoError(t, err)
	srcRelTree, err := fs.ConvertToRelativePath(tmpSrcDir, srcTree...)
	require.NoError(t, err)

	for _, relSrcPath := range srcRelTree {
		absoluteCachePath := filesystem.FilePathJoin(fs, tmpCacheDir, relSrcPath)
		require.False(t, filesystem.Exists(absoluteCachePath), "cache did not delete all files after closing")
	}

	err = cache.Add(ctx, faker.UUIDDigit(), fs, faker.Word())
	errortest.AssertError(t, err, commonerrors.ErrForbidden)

	err = cache.Remove(ctx, faker.UUIDDigit())
	errortest.AssertError(t, err, commonerrors.ErrForbidden)

	_, err = cache.Contains(ctx, faker.UUIDDigit())
	errortest.AssertError(t, err, commonerrors.ErrForbidden)

	err = cache.Restore(ctx, faker.UUIDDigit())
	errortest.AssertError(t, err, commonerrors.ErrForbidden)

	err = cache.Close(ctx)
	errortest.AssertError(t, err, commonerrors.ErrForbidden)
}

func TestCache_GarbageCollection(t *testing.T) {
	defer goleak.VerifyNone(t)

	fs := filesystem.NewFs(filesystem.StandardFS)
	tmpCacheDir, err := fs.TempDirInTempDir("test-cache")
	require.NoError(t, err)
	defer func() { _ = fs.Rm(tmpCacheDir) }()

	ctx := context.Background()
	config := &Config{
		cachefs:                 fs,
		cachePath:               tmpCacheDir,
		garbageCollectionPeriod: 1 * time.Second,
		ttl:                     2 * time.Second,
	}

	cache, err := NewSimpleCache(ctx, config)
	require.NoError(t, err)

	tmpfilePath, err := fs.TouchTempFileInTempDir(faker.Word())
	require.NoError(t, err)

	id := faker.UUIDDigit()
	err = cache.Add(ctx, id, fs, tmpfilePath)
	require.NoError(t, err)

	time.Sleep(3 * time.Second)

	cacheExists, err := cache.Contains(ctx, id)
	require.NoError(t, err)
	require.False(t, cacheExists, "cache still has an entry after removing")

	cacheFilePath := filesystem.FilePathJoin(fs, tmpCacheDir, filesystem.FilePathBase(fs, tmpfilePath))
	require.False(t, filesystem.Exists(cacheFilePath), "cache still has the file after removing")

	err = cache.Close(ctx)
	require.NoError(t, err)
}
