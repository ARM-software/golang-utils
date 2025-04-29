package filecache

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/filesystem"
	"github.com/ARM-software/golang-utils/utils/filesystem/filesystemtest"
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

			_ = filesystemtest.CreateTestFileTree(t, srcFs, tmpSrcDir, time.Now(), time.Now())
			srcContent, err := srcFs.Ls(tmpSrcDir)
			require.NoError(t, err)

			cacheFs := filesystem.NewFs(test.cacheFs)
			tmpCacheDir, err := cacheFs.TempDirInTempDir("test-cache")
			require.NoError(t, err)
			defer func() { _ = cacheFs.Rm(tmpCacheDir) }()

			ctx := context.Background()
			config := DefaultFileCacheConfig()
			config.CachePath = tmpCacheDir
			cache, err := NewFsFileCache(ctx, srcFs, cacheFs, tmpSrcDir, config)
			require.NoError(t, err)

			for _, path := range srcContent {
				err := cache.Store(ctx, path)
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

			err = cache.Close()
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
		config := DefaultFileCacheConfig()
		config.CachePath = tmpCacheDir
		cache, err := NewFsFileCache(ctx, fs, fs, "/does/not/exist", config)
		require.NoError(t, err)

		err = cache.Store(ctx, faker.UUIDDigit())
		errortest.AssertError(t, err, commonerrors.ErrNotFound)

		err = cache.Close()
		require.NoError(t, err)
	})

	t.Run("Add_duplicate_entry", func(t *testing.T) {
		defer goleak.VerifyNone(t)

		fs := filesystem.NewFs(filesystem.StandardFS)
		tmpCacheDir, err := fs.TempDirInTempDir("test-cache")
		require.NoError(t, err)
		defer func() { _ = fs.Rm(tmpCacheDir) }()

		tmpTestPath, err := fs.TouchTempFileInTempDir(faker.Word())
		require.NoError(t, err)
		tmpTestDir := fs.TempDirectory()
		tmpTestBase := filesystem.FilePathBase(fs, tmpTestPath)

		ctx := context.Background()
		config := DefaultFileCacheConfig()
		config.CachePath = tmpCacheDir
		cache, err := NewFsFileCache(ctx, fs, fs, tmpTestDir, config)
		require.NoError(t, err)

		require.NoError(t, err)

		err = cache.Store(ctx, tmpTestBase)
		require.NoError(t, err)

		err = cache.Store(ctx, tmpTestBase)
		errortest.AssertError(t, err, commonerrors.ErrExists)

		err = cache.Close()
		require.NoError(t, err)
	})
}

func TestCache_Fetch(t *testing.T) {
	t.Run("Fetch_file", func(t *testing.T) {
		defer goleak.VerifyNone(t)

		fs := filesystem.NewFs(filesystem.StandardFS)
		tmpCacheDir, err := fs.TempDirInTempDir("test-cache")
		require.NoError(t, err)
		defer func() { _ = fs.Rm(tmpCacheDir) }()

		tmpfilePath, err := fs.TouchTempFileInTempDir(faker.Word())
		require.NoError(t, err)
		tmpfileDir := fs.TempDirectory()
		tmpfileBase := filesystem.FilePathBase(fs, tmpfilePath)

		ctx := context.Background()
		config := DefaultFileCacheConfig()
		config.CachePath = tmpCacheDir
		cache, err := NewFsFileCache(ctx, fs, fs, tmpfileDir, config)
		require.NoError(t, err)

		err = cache.Store(ctx, tmpfileBase)
		require.NoError(t, err)

		err = fs.Rm(tmpfilePath)
		require.NoError(t, err)

		err = cache.Fetch(ctx, tmpfileBase, fs, tmpfilePath)
		require.NoError(t, err)

		require.True(t, filesystem.Exists(tmpfilePath), "cache did not fetch the file")
		err = cache.Close()
		require.NoError(t, err)
	})

	t.Run("Fetch_file_non_existent", func(t *testing.T) {
		defer goleak.VerifyNone(t)

		fs := filesystem.NewFs(filesystem.StandardFS)
		tmpCacheDir, err := fs.TempDirInTempDir("test-cache")
		require.NoError(t, err)
		defer func() { _ = fs.Rm(tmpCacheDir) }()

		tmpSrcDir, err := fs.TempDirInTempDir("test-src")
		require.NoError(t, err)
		defer func() { _ = fs.Rm(tmpSrcDir) }()

		ctx := context.Background()
		config := DefaultFileCacheConfig()
		config.CachePath = tmpCacheDir
		cache, err := NewFsFileCache(ctx, fs, fs, tmpSrcDir, config)
		require.NoError(t, err)

		err = cache.Fetch(ctx, faker.UUIDDigit(), fs, tmpSrcDir)
		errortest.AssertError(t, err, commonerrors.ErrNotFound)

		err = cache.Close()
		require.NoError(t, err)
	})

	t.Run("Fetch_file_overwrite", func(t *testing.T) {
		defer goleak.VerifyNone(t)

		fs := filesystem.NewFs(filesystem.StandardFS)
		tmpCacheDir, err := fs.TempDirInTempDir("test-cache")
		require.NoError(t, err)
		defer func() { _ = fs.Rm(tmpCacheDir) }()

		tmpFile, err := fs.TempFileInTempDir(faker.Word())
		require.NoError(t, err)
		tmpfilePath := tmpFile.Name()
		tmpfileDir := fs.TempDirectory()
		tmpfileBase := filesystem.FilePathBase(fs, tmpfilePath)
		defer func() { _ = fs.Rm(tmpfilePath) }()

		ctx := context.Background()
		config := DefaultFileCacheConfig()
		config.CachePath = tmpCacheDir
		cache, err := NewFsFileCache(ctx, fs, fs, tmpfileDir, config)
		require.NoError(t, err)

		originalFileContent := []byte("Original")
		_, err = tmpFile.Write(originalFileContent)
		require.NoError(t, err)

		err = cache.Store(ctx, tmpfileBase)
		require.NoError(t, err)

		NewFileContent := []byte("New")
		_, err = tmpFile.Write(NewFileContent)
		require.NoError(t, err)

		err = cache.Fetch(ctx, tmpfileBase, fs, tmpfilePath)
		require.NoError(t, err)

		cacheFileContent, err := fs.ReadFile(tmpfilePath)
		require.NoError(t, err)
		require.Equal(t, originalFileContent, cacheFileContent)

		err = cache.Close()
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

		tmpfilePath, err := fs.TouchTempFileInTempDir(faker.Word())
		require.NoError(t, err)
		defer func() { _ = fs.Rm(tmpfilePath) }()
		tmpTestDir := fs.TempDirectory()
		tmpTestBase := filesystem.FilePathBase(fs, tmpfilePath)

		ctx := context.Background()
		config := DefaultFileCacheConfig()
		config.CachePath = tmpCacheDir
		cache, err := NewFsFileCache(ctx, fs, fs, tmpTestDir, config)
		require.NoError(t, err)

		err = cache.Store(ctx, tmpTestBase)
		require.NoError(t, err)

		err = cache.Evict(ctx, tmpTestBase)
		require.NoError(t, err)

		cacheExists, err := cache.Has(ctx, tmpTestBase)
		require.NoError(t, err)
		require.False(t, cacheExists, "cache still has an entry after removing")

		cacheFilePath := filesystem.FilePathJoin(fs, tmpCacheDir, filesystem.FilePathBase(fs, tmpfilePath))
		require.False(t, filesystem.Exists(cacheFilePath), "cache still has the file after removing")

		err = cache.Close()
		require.NoError(t, err)
	})

	t.Run("Remove_file_non_existent", func(t *testing.T) {
		defer goleak.VerifyNone(t)

		fs := filesystem.NewFs(filesystem.StandardFS)
		tmpCacheDir, err := fs.TempDirInTempDir("test-cache")
		require.NoError(t, err)
		defer func() { _ = fs.Rm(tmpCacheDir) }()

		ctx := context.Background()
		config := DefaultFileCacheConfig()
		config.CachePath = tmpCacheDir
		cache, err := NewFsFileCache(ctx, fs, fs, tmpCacheDir, config)
		require.NoError(t, err)

		id := faker.UUIDDigit()
		err = cache.Evict(ctx, id)
		require.NoError(t, err)
		require.False(t, filesystem.Exists(id), "cache still has the file after removing")

		err = cache.Close()
		require.NoError(t, err)
	})
}

func TestCache_Close(t *testing.T) {
	defer goleak.VerifyNone(t)

	fs := filesystem.NewFs(filesystem.StandardFS)
	tmpSrcDir, err := fs.TempDirInTempDir("test-cache-src")
	require.NoError(t, err)
	defer func() { _ = fs.Rm(tmpSrcDir) }()

	_ = filesystemtest.CreateTestFileTree(t, fs, tmpSrcDir, time.Now(), time.Now())
	srcContent, err := fs.Ls(tmpSrcDir)
	require.NoError(t, err)

	tmpCacheDir, err := fs.TempDirInTempDir("test-cache")
	require.NoError(t, err)
	defer func() { _ = fs.Rm(tmpCacheDir) }()

	ctx := context.Background()
	config := DefaultFileCacheConfig()
	config.CachePath = tmpCacheDir
	cache, err := NewFsFileCache(ctx, fs, fs, tmpSrcDir, config)
	require.NoError(t, err)

	for _, path := range srcContent {
		err := cache.Store(ctx, path)
		require.NoError(t, err)
	}

	err = cache.Close()
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

	err = cache.Store(ctx, tmpSrcDir)
	errortest.AssertError(t, err, commonerrors.ErrConflict)

	err = cache.Evict(ctx, faker.UUIDDigit())
	errortest.AssertError(t, err, commonerrors.ErrConflict)

	_, err = cache.Has(ctx, faker.UUIDDigit())
	errortest.AssertError(t, err, commonerrors.ErrConflict)

	err = cache.Fetch(ctx, faker.UUIDDigit(), fs, tmpSrcDir)
	errortest.AssertError(t, err, commonerrors.ErrConflict)

	err = cache.Close()
	require.NoError(t, err)
}

func TestCache_GarbageCollection(t *testing.T) {
	defer goleak.VerifyNone(t)

	fs := filesystem.NewFs(filesystem.StandardFS)
	tmpCacheDir, err := fs.TempDirInTempDir("test-cache")
	require.NoError(t, err)
	defer func() { _ = fs.Rm(tmpCacheDir) }()

	tmpfilePath, err := fs.TouchTempFileInTempDir(faker.Word())
	require.NoError(t, err)
	tmpfileDir := fs.TempDirectory()
	tmpfileBase := filesystem.FilePathBase(fs, tmpfilePath)

	ctx := context.Background()
	config := &FileCacheConfig{
		CachePath:               tmpCacheDir,
		GarbageCollectionPeriod: 1 * time.Second,
		TTL:                     2 * time.Second,
	}

	cache, err := NewFsFileCache(ctx, fs, fs, tmpfileDir, config)
	require.NoError(t, err)

	id := faker.UUIDDigit()
	err = cache.Store(ctx, tmpfileBase)
	require.NoError(t, err)

	time.Sleep(3 * time.Second)

	cacheExists, err := cache.Has(ctx, id)
	require.NoError(t, err)
	require.False(t, cacheExists, "cache still has an entry after removing")

	cacheFilePath := filesystem.FilePathJoin(fs, tmpCacheDir, filesystem.FilePathBase(fs, tmpfilePath))
	require.False(t, filesystem.Exists(cacheFilePath), "cache still has the file after removing")

	err = cache.Close()
	require.NoError(t, err)
}
