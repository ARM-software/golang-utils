package filecache

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"golang.org/x/sync/errgroup"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/filesystem"
	"github.com/ARM-software/golang-utils/utils/filesystem/filesystemtest"
	"github.com/ARM-software/golang-utils/utils/hashing"
)

func TestFileCache_Add(t *testing.T) {
	tests := []struct {
		name    string
		srcFs   filesystem.FilesystemType
		cacheFs filesystem.FilesystemType
	}{
		{
			name:    "Add content to cache same filesystem",
			srcFs:   filesystem.StandardFS,
			cacheFs: filesystem.StandardFS,
		},
		{
			name:    "Add content to cache different filesystem",
			srcFs:   filesystem.StandardFS,
			cacheFs: filesystem.InMemoryFS,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer goleak.VerifyNone(t)

			srcFs := filesystem.NewFs(test.srcFs)
			tmpSrcDir := filesystem.FilePathJoin(srcFs, t.TempDir(), "test-cache-src")
			require.NoError(t, srcFs.MkDir(tmpSrcDir))

			_ = filesystemtest.CreateTestFileTree(t, srcFs, tmpSrcDir, time.Now(), time.Now())
			srcContent, err := srcFs.Ls(tmpSrcDir)
			require.NoError(t, err)

			cacheFs := filesystem.NewFs(test.cacheFs)
			tmpCacheDir := filesystem.FilePathJoin(cacheFs, t.TempDir(), "test-cache")
			require.NoError(t, cacheFs.MkDir(tmpCacheDir))

			ctx := context.Background()
			config := DefaultFileCacheConfig()
			config.CachePath = tmpCacheDir
			cache, err := NewFsFileCache(ctx, srcFs, cacheFs, tmpSrcDir, config)
			require.NoError(t, err)

			defer func() {
				err = cache.Close(ctx)
				require.NoError(t, err)
			}()

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

			require.Equal(t, len(srcTree), len(cacheTree), "Cache Dir and Src Dir have a different number of content")

			srcRelTree, err := srcFs.ConvertToRelativePath(tmpSrcDir, srcTree...)
			require.NoError(t, err)

			cacheRelTree, err := cacheFs.ConvertToRelativePath(tmpCacheDir, cacheTree...)
			require.NoError(t, err)

			slices.Sort(srcRelTree)
			slices.Sort(cacheRelTree)
			require.Equal(t, srcRelTree, cacheRelTree, "Cache Dir and Src Dir have different contents")
		})
	}

	t.Run("Add invalid path", func(t *testing.T) {
		defer goleak.VerifyNone(t)

		fs := filesystem.GetGlobalFileSystem()
		tmpCacheDir := filesystem.FilePathJoin(fs, t.TempDir(), "test-cache")
		require.NoError(t, fs.MkDir(tmpCacheDir))

		ctx := context.Background()
		config := DefaultFileCacheConfig()
		config.CachePath = tmpCacheDir
		cache, err := NewFsFileCache(ctx, fs, fs, "/does/not/exist", config)
		require.NoError(t, err)

		defer func() {
			err = cache.Close(ctx)
			require.NoError(t, err)
		}()

		err = cache.Store(ctx, faker.UUIDDigit())
		errortest.AssertError(t, err, commonerrors.ErrNotFound)
	})

	t.Run("Add duplicate entry", func(t *testing.T) {
		defer goleak.VerifyNone(t)

		fs := filesystem.GetGlobalFileSystem()
		tmpCacheDir := filesystem.FilePathJoin(fs, t.TempDir(), "test-cache")
		require.NoError(t, fs.MkDir(tmpCacheDir))

		tmpTestPath, err := fs.TouchTempFileInTempDir(faker.Word())
		require.NoError(t, err)
		tmpTestDir := fs.TempDirectory()
		tmpTestBase := filesystem.FilePathBase(fs, tmpTestPath)

		ctx := context.Background()
		config := DefaultFileCacheConfig()
		config.CachePath = tmpCacheDir
		cache, err := NewFsFileCache(ctx, fs, fs, tmpTestDir, config)
		require.NoError(t, err)

		defer func() {
			err = cache.Close(ctx)
			require.NoError(t, err)
		}()

		err = cache.Store(ctx, tmpTestBase)
		require.NoError(t, err)

		err = cache.Store(ctx, tmpTestBase)
		errortest.AssertError(t, err, commonerrors.ErrExists)
	})
}

func TestFileCache_Fetch(t *testing.T) {
	t.Run("Fetch file", func(t *testing.T) {
		defer goleak.VerifyNone(t)

		fs := filesystem.GetGlobalFileSystem()
		tmpCacheDir := filesystem.FilePathJoin(fs, t.TempDir(), "test-cache")
		require.NoError(t, fs.MkDir(tmpCacheDir))

		tmpfilePath, err := fs.TouchTempFileInTempDir(faker.Word())
		require.NoError(t, err)
		tmpfileDir := fs.TempDirectory()
		tmpfileBase := filesystem.FilePathBase(fs, tmpfilePath)

		ctx := context.Background()
		config := DefaultFileCacheConfig()
		config.CachePath = tmpCacheDir
		cache, err := NewFsFileCache(ctx, fs, fs, tmpfileDir, config)
		require.NoError(t, err)

		defer func() {
			err = cache.Close(ctx)
			require.NoError(t, err)
		}()

		err = cache.Store(ctx, tmpfileBase)
		require.NoError(t, err)

		err = fs.Rm(tmpfilePath)
		require.NoError(t, err)
		require.False(t, filesystem.Exists(tmpfilePath), "could not remove the src file")

		err = cache.Fetch(ctx, tmpfileBase, fs, tmpfilePath)
		require.NoError(t, err)

		require.True(t, filesystem.Exists(tmpfilePath), "cache did not fetch the file")
	})

	t.Run("Fetch file non-existent", func(t *testing.T) {
		defer goleak.VerifyNone(t)

		fs := filesystem.GetGlobalFileSystem()
		tmpCacheDir := filesystem.FilePathJoin(fs, t.TempDir(), "test-cache")
		require.NoError(t, fs.MkDir(tmpCacheDir))

		tmpSrcDir := filesystem.FilePathJoin(fs, t.TempDir(), "test-src")
		require.NoError(t, fs.MkDir(tmpSrcDir))

		ctx := context.Background()
		config := DefaultFileCacheConfig()
		config.CachePath = tmpCacheDir
		cache, err := NewFsFileCache(ctx, fs, fs, tmpSrcDir, config)
		require.NoError(t, err)

		defer func() {
			err = cache.Close(ctx)
			require.NoError(t, err)
		}()

		err = cache.Fetch(ctx, faker.UUIDDigit(), fs, tmpSrcDir)
		errortest.AssertError(t, err, commonerrors.ErrNotFound)
	})

	t.Run("Fetch file overwrite", func(t *testing.T) {
		defer goleak.VerifyNone(t)

		fs := filesystem.GetGlobalFileSystem()
		tmpDir := t.TempDir()
		tmpCacheDir := filesystem.FilePathJoin(fs, tmpDir, "test-cache")
		require.NoError(t, fs.MkDir(tmpCacheDir))

		tmpfilePath := filesystem.FilePathJoin(fs, tmpDir, faker.Word())
		tmpFile, err := fs.CreateFile(tmpfilePath)
		require.NoError(t, err)
		tmpfileBase := filesystem.FilePathBase(fs, tmpfilePath)

		ctx := context.Background()
		config := DefaultFileCacheConfig()
		config.CachePath = tmpCacheDir
		cache, err := NewFsFileCache(ctx, fs, fs, tmpDir, config)
		require.NoError(t, err)

		defer func() {
			err = cache.Close(ctx)
			require.NoError(t, err)
		}()

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
	})
}

func TestFileCache_Remove(t *testing.T) {
	t.Run("Remove file", func(t *testing.T) {
		defer goleak.VerifyNone(t)

		fs := filesystem.GetGlobalFileSystem()
		tmpDir := t.TempDir()
		tmpCacheDir := filesystem.FilePathJoin(fs, tmpDir, "test-cache")
		require.NoError(t, fs.MkDir(tmpCacheDir))

		tmpfilePath := filesystem.FilePathJoin(fs, tmpDir, faker.Word())
		require.NoError(t, fs.Touch(tmpfilePath))
		tmpTestBase := filesystem.FilePathBase(fs, tmpfilePath)

		ctx := context.Background()
		config := DefaultFileCacheConfig()
		config.CachePath = tmpCacheDir
		cache, err := NewFsFileCache(ctx, fs, fs, tmpDir, config)
		require.NoError(t, err)

		defer func() {
			err = cache.Close(ctx)
			require.NoError(t, err)
		}()

		err = cache.Store(ctx, tmpTestBase)
		require.NoError(t, err)

		err = cache.Evict(ctx, tmpTestBase)
		require.NoError(t, err)

		cacheExists, err := cache.Has(ctx, tmpTestBase)
		require.NoError(t, err)
		require.False(t, cacheExists, "cache still has an entry after removing")

		cacheFilePath := filesystem.FilePathJoin(fs, tmpCacheDir, filesystem.FilePathBase(fs, tmpfilePath))
		require.False(t, filesystem.Exists(cacheFilePath), "cache still has the file after removing")
	})

	t.Run("Remove file non-existent", func(t *testing.T) {
		defer goleak.VerifyNone(t)

		fs := filesystem.GetGlobalFileSystem()
		tmpCacheDir := filesystem.FilePathJoin(fs, t.TempDir(), "test-cache")
		require.NoError(t, fs.MkDir(tmpCacheDir))

		ctx := context.Background()
		config := DefaultFileCacheConfig()
		config.CachePath = tmpCacheDir
		cache, err := NewFsFileCache(ctx, fs, fs, tmpCacheDir, config)
		require.NoError(t, err)

		defer func() {
			err = cache.Close(ctx)
			require.NoError(t, err)
		}()

		id := faker.UUIDDigit()
		err = cache.Evict(ctx, id)
		require.NoError(t, err)
		require.False(t, filesystem.Exists(id), "cache still has the file after removing")
	})
}

func TestFileCache_Close(t *testing.T) {
	defer goleak.VerifyNone(t)

	fs := filesystem.GetGlobalFileSystem()
	tmpSrcDir := filesystem.FilePathJoin(fs, t.TempDir(), "test-cache-src")
	require.NoError(t, fs.MkDir(tmpSrcDir))

	_ = filesystemtest.CreateTestFileTree(t, fs, tmpSrcDir, time.Now(), time.Now())
	srcContent, err := fs.Ls(tmpSrcDir)
	require.NoError(t, err)

	tmpCacheDir := filesystem.FilePathJoin(fs, t.TempDir(), "test-cache")
	require.NoError(t, fs.MkDir(tmpCacheDir))

	ctx := context.Background()
	config := DefaultFileCacheConfig()
	config.CachePath = tmpCacheDir
	cache, err := NewFsFileCache(ctx, fs, fs, tmpSrcDir, config)
	require.NoError(t, err)

	for _, path := range srcContent {
		err := cache.Store(ctx, path)
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

	err = cache.Store(ctx, tmpSrcDir)
	errortest.AssertError(t, err, commonerrors.ErrConflict)

	err = cache.Evict(ctx, faker.UUIDDigit())
	errortest.AssertError(t, err, commonerrors.ErrConflict)

	_, err = cache.Has(ctx, faker.UUIDDigit())
	errortest.AssertError(t, err, commonerrors.ErrConflict)

	err = cache.Fetch(ctx, faker.UUIDDigit(), fs, tmpSrcDir)
	errortest.AssertError(t, err, commonerrors.ErrConflict)

	err = cache.Close(ctx)
	require.NoError(t, err)
}

func TestFileCache_GarbageCollection(t *testing.T) {
	defer goleak.VerifyNone(t)

	fs := filesystem.GetGlobalFileSystem()
	tmpCacheDir := filesystem.FilePathJoin(fs, t.TempDir(), "test-cache")
	require.NoError(t, fs.MkDir(tmpCacheDir))

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

	defer func() {
		err = cache.Close(ctx)
		require.NoError(t, err)
	}()

	id := faker.UUIDDigit()
	err = cache.Store(ctx, tmpfileBase)
	require.NoError(t, err)

	time.Sleep(3 * time.Second)

	cacheExists, err := cache.Has(ctx, id)
	require.NoError(t, err)
	require.False(t, cacheExists, "cache still has an entry after removing")

	cacheFilePath := filesystem.FilePathJoin(fs, tmpCacheDir, filesystem.FilePathBase(fs, tmpfilePath))
	require.False(t, filesystem.Exists(cacheFilePath), "cache still has the file after removing")
}

func TestFileCache_Cancel(t *testing.T) {
	t.Run("Cancel after cache creation", func(t *testing.T) {
		defer goleak.VerifyNone(t)

		fs := filesystem.GetGlobalFileSystem()
		tmpCacheDir := filesystem.FilePathJoin(fs, t.TempDir(), "test-cache")
		require.NoError(t, fs.MkDir(tmpCacheDir))

		tmpfilePath := filesystem.FilePathJoin(fs, t.TempDir(), faker.Word())
		require.NoError(t, fs.Touch(tmpfilePath))

		ctx, cancel := context.WithCancel(context.Background())
		config := DefaultFileCacheConfig()
		config.CachePath = tmpCacheDir
		_, err := NewFsFileCache(ctx, fs, fs, t.TempDir(), config)
		require.NoError(t, err)

		cancel()
		errortest.AssertError(t, ctx.Err(), context.Canceled)
	})

	t.Run("Cancel during add", func(t *testing.T) {
		defer goleak.VerifyNone(t)

		fs := filesystem.GetGlobalFileSystem()
		tmpSrcDir := filesystem.FilePathJoin(fs, t.TempDir(), "test-cache-src")
		require.NoError(t, fs.MkDir(tmpSrcDir))

		_ = filesystemtest.CreateTestFileTree(t, fs, tmpSrcDir, time.Now(), time.Now())
		srcContent, err := fs.Ls(tmpSrcDir)
		require.NoError(t, err)

		tmpCacheDir := filesystem.FilePathJoin(fs, t.TempDir(), "test-cache")
		require.NoError(t, fs.MkDir(tmpCacheDir))

		ctx, cancel := context.WithCancel(context.Background())
		config := DefaultFileCacheConfig()
		config.CachePath = tmpCacheDir
		cache, err := NewFsFileCache(ctx, fs, fs, tmpSrcDir, config)
		require.NoError(t, err)

		go func() {
			for _, path := range srcContent {
				_ = cache.Store(ctx, path)
			}
		}()

		cancel()
		errortest.AssertError(t, ctx.Err(), context.Canceled)
	})

	t.Run("Cancel during close", func(t *testing.T) {
		defer goleak.VerifyNone(t)

		fs := filesystem.GetGlobalFileSystem()
		tmpSrcDir := filesystem.FilePathJoin(fs, t.TempDir(), "test-cache-src")
		require.NoError(t, fs.MkDir(tmpSrcDir))

		_ = filesystemtest.CreateTestFileTree(t, fs, tmpSrcDir, time.Now(), time.Now())
		srcContent, err := fs.Ls(tmpSrcDir)
		require.NoError(t, err)

		tmpCacheDir := filesystem.FilePathJoin(fs, t.TempDir(), "test-cache")
		require.NoError(t, fs.MkDir(tmpCacheDir))

		ctx, cancel := context.WithCancel(context.Background())
		config := DefaultFileCacheConfig()
		config.CachePath = tmpCacheDir
		cache, err := NewFsFileCache(ctx, fs, fs, tmpSrcDir, config)
		require.NoError(t, err)

		for _, path := range srcContent {
			err := cache.Store(ctx, path)
			require.NoError(t, err)
		}

		go func() { _ = cache.Close(ctx) }()

		cancel()
		errortest.AssertError(t, ctx.Err(), context.Canceled)
	})
}

func TestFileCache_Concurent_Caches(t *testing.T) {
	t.Run("Caches", func(t *testing.T) {
		defer goleak.VerifyNone(t)

		var wg sync.WaitGroup
		numCaches := 50
		for i := 0; i < numCaches; i++ {
			wg.Add(1)

			go func() {
				defer wg.Done()

				fs := filesystem.GetGlobalFileSystem()
				tmpSrcDir := filesystem.FilePathJoin(fs, t.TempDir(), "test-cache-src")
				require.NoError(t, fs.MkDir(tmpSrcDir))

				_ = filesystemtest.CreateTestFileTree(t, fs, tmpSrcDir, time.Now(), time.Now())
				srcContent, err := fs.Ls(tmpSrcDir)
				require.NoError(t, err)

				tmpCacheDir := filesystem.FilePathJoin(fs, t.TempDir(), "test-cache")
				require.NoError(t, fs.MkDir(tmpCacheDir))

				ctx := context.Background()
				config := DefaultFileCacheConfig()
				config.CachePath = tmpCacheDir
				cache, err := NewFsFileCache(ctx, fs, fs, tmpSrcDir, config)
				require.NoError(t, err)

				defer func() {
					err = cache.Close(ctx)
					require.NoError(t, err)
				}()

				for _, path := range srcContent {
					err := cache.Store(ctx, path)
					require.NoError(t, err)
				}

				var cacheTree []string
				err = fs.ListDirTree(tmpCacheDir, &cacheTree)
				require.NoError(t, err)

				var srcTree []string
				err = fs.ListDirTree(tmpSrcDir, &srcTree)
				require.NoError(t, err)

				require.Equal(t, len(srcTree), len(cacheTree), "Cache Dir and Src Dir have a different number of content")

				srcRelTree, err := fs.ConvertToRelativePath(tmpSrcDir, srcTree...)
				require.NoError(t, err)

				cacheRelTree, err := fs.ConvertToRelativePath(tmpCacheDir, cacheTree...)
				require.NoError(t, err)

				slices.Sort(srcRelTree)
				slices.Sort(cacheRelTree)
				require.Equal(t, srcRelTree, cacheRelTree, "Cache Dir and Src Dir have different contents")
			}()
		}

		wg.Wait()
	})
	t.Run("Cache operations", func(t *testing.T) {
		defer goleak.VerifyNone(t)

		fs := filesystem.GetGlobalFileSystem()
		tmpSrcDir := filesystem.FilePathJoin(fs, t.TempDir(), "test-cache-src")
		require.NoError(t, fs.MkDir(tmpSrcDir))

		_ = filesystemtest.CreateTestFileTree(t, fs, tmpSrcDir, time.Now(), time.Now())
		evictSrcContent, err := fs.Ls(tmpSrcDir)
		require.NoError(t, err)

		evictSuffix := "-evict"
		for i, path := range evictSrcContent {
			newBaseName := path + evictSuffix
			newPath := filesystem.FilePathJoin(fs, tmpSrcDir, newBaseName)
			oldPath := filesystem.FilePathJoin(fs, tmpSrcDir, path)
			err := fs.Move(oldPath, newPath)
			require.NoError(t, err)
			evictSrcContent[i] = newBaseName
		}

		_ = filesystemtest.CreateTestFileTree(t, fs, tmpSrcDir, time.Now(), time.Now())
		storeSrcContent, err := fs.LsWithExclusionPatterns(tmpSrcDir, fmt.Sprintf(".*%s", evictSuffix))
		require.NoError(t, err)

		tmpCacheDir := filesystem.FilePathJoin(fs, t.TempDir(), "test-cache")
		require.NoError(t, fs.MkDir(tmpCacheDir))

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		config := DefaultFileCacheConfig()
		config.CachePath = tmpCacheDir
		cache, err := NewFsFileCache(ctx, fs, fs, tmpSrcDir, config)
		require.NoError(t, err)

		defer func() {
			err = cache.Close(ctx)
			require.NoError(t, err)
			cancel()
		}()

		for _, path := range evictSrcContent {
			err := cache.Store(ctx, path)
			require.NoError(t, err)
		}

		g, gCtx := errgroup.WithContext(ctx)
		g.Go(func() error {
			for _, path := range storeSrcContent {
				if err := cache.Store(gCtx, path); err != nil {
					return err
				}
			}

			return nil
		})

		g.Go(func() error {
			for _, path := range evictSrcContent {
				if err := cache.Evict(gCtx, path); err != nil {
					return err
				}
			}

			return nil
		})

		err = g.Wait()
		require.NoError(t, err)

		var cacheTree []string
		err = fs.ListDirTree(tmpCacheDir, &cacheTree)
		require.NoError(t, err)

		var srcTree []string
		err = fs.ListDirTreeWithContextAndExclusionPatterns(ctx, tmpSrcDir, &srcTree, fmt.Sprintf(".*%s", evictSuffix))
		require.NoError(t, err)

		require.Equal(t, len(srcTree), len(cacheTree), "Cache Dir and Src Dir have a different number of content")

		srcRelTree, err := fs.ConvertToRelativePath(tmpSrcDir, srcTree...)
		require.NoError(t, err)

		cacheRelTree, err := fs.ConvertToRelativePath(tmpCacheDir, cacheTree...)
		require.NoError(t, err)

		slices.Sort(srcRelTree)
		slices.Sort(cacheRelTree)
		require.Equal(t, srcRelTree, cacheRelTree, "Cache Dir and Src Dir have different contents")

	})

	t.Run("Restore Evict", func(t *testing.T) {
		defer goleak.VerifyNone(t)

		fs := filesystem.GetGlobalFileSystem()
		tmpCacheDir := filesystem.FilePathJoin(fs, t.TempDir(), "test-cache")
		require.NoError(t, fs.MkDir(tmpCacheDir))

		tmpSrcDir := filesystem.FilePathJoin(fs, t.TempDir(), "test-cache-src")
		require.NoError(t, fs.MkDir(tmpSrcDir))

		testFileName := "500MB.bin"
		tmpTestFilePath := filesystem.FilePathJoin(fs, tmpSrcDir, testFileName)
		filesystemtest.GenerateTestFile(t, fs, tmpTestFilePath, 500*1024*1024, 1024)
		hasher, err := filesystem.NewFileHash(hashing.HashSha256)
		require.NoError(t, err)
		expectedHash, err := hasher.CalculateFile(fs, tmpTestFilePath)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		config := DefaultFileCacheConfig()
		config.CachePath = tmpCacheDir
		cache, err := NewFsFileCache(ctx, fs, fs, tmpSrcDir, config)
		require.NoError(t, err)

		defer func() {
			err = cache.Close(ctx)
			require.NoError(t, err)
			cancel()
		}()

		err = cache.Store(ctx, testFileName)
		require.NoError(t, err)

		err = fs.Rm(tmpTestFilePath)
		require.NoError(t, err)
		require.False(t, filesystem.Exists(tmpTestFilePath), "could not remove the src file")

		g, gCtx := errgroup.WithContext(ctx)
		g.Go(func() error {
			if err := cache.Fetch(gCtx, testFileName, fs, tmpTestFilePath); err != nil {
				return err
			}

			if !filesystem.Exists(tmpTestFilePath) {
				return commonerrors.New(commonerrors.ErrUnexpected, "cache did not restore the file")
			}

			return nil
		})

		g.Go(func() error {
			time.Sleep(100 * time.Millisecond)
			if err := cache.Evict(gCtx, testFileName); err != nil {
				return err
			}

			exists, err := cache.Has(gCtx, testFileName)
			if err != nil {
				return err
			}

			if exists {
				return commonerrors.New(commonerrors.ErrUnexpected, "cache did not evict the entry")
			}

			return nil
		})

		err = g.Wait()
		require.NoError(t, err)

		actualHash, err := hasher.CalculateFile(fs, tmpTestFilePath)
		require.NoError(t, err)

		require.Equal(t, actualHash, expectedHash, "the copied file got corrupted")
	})
}
