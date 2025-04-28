package sharedcache

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/filesystem"
	"github.com/ARM-software/golang-utils/utils/filesystem/filesystemtest"
)

func TestNothingInCacheWorkflow(t *testing.T) { // Single fetch with no file previously cached
	for c := range CacheTypes {
		cacheType := CacheTypes[c]
		for i := range filesystem.FileSystemTypes {
			fsType := filesystem.FileSystemTypes[i]
			testName := fmt.Sprintf("%v_for_fs_%v_and_cache_%v", t.Name(), fsType, cacheType)
			t.Run(testName, func(t *testing.T) {
				t.Parallel()
				ctx := context.Background()
				fs := filesystem.NewFs(fsType)
				// set up temp remote directory
				tmpRemoteDir, err := fs.TempDirInTempDir(fmt.Sprintf("test-%v-remote", testName))
				require.NoError(t, err)
				defer func() { _ = fs.Rm(tmpRemoteDir) }()

				tmpDestDir, err := fs.TempDirInTempDir(fmt.Sprintf("test-%v-local", testName))
				require.NoError(t, err)
				defer func() { _ = fs.Rm(tmpDestDir) }()

				remoteCache, err := NewCache(cacheType, fs, &Configuration{
					RemoteStoragePath: tmpRemoteDir,
					Timeout:           time.Second,
				})
				require.NoError(t, err)
				count, err := remoteCache.EntriesCount(context.TODO())
				require.NoError(t, err)
				assert.Zero(t, count)

				key := remoteCache.GenerateKey("test", "cache", fmt.Sprintf("%v", cacheType))

				err = remoteCache.Fetch(ctx, key, tmpDestDir)
				require.NotNil(t, err)
				assert.True(t, commonerrors.Any(err, commonerrors.ErrNotFound, commonerrors.ErrEmpty))
			})
		}
	}
}

func TestSimpleCacheWorkflow(t *testing.T) { // Simple store, followed by fetch
	for c := range CacheTypes {
		cacheType := CacheTypes[c]
		for i := range filesystem.FileSystemTypes {
			fsType := filesystem.FileSystemTypes[i]
			if fsType == filesystem.InMemoryFS && cacheType == CacheMutable {
				// FIXME There is an error with lock unlock when using in memory fs
				continue
			}
			testName := fmt.Sprintf("%v_for_fs_%v_and_cache_%v", t.Name(), fsType, cacheType)
			t.Run(testName, func(t *testing.T) {
				t.Parallel()
				ctx := context.Background()
				fs := filesystem.NewFs(fsType)
				// set up temp remote directory
				tmpRemoteDir, err := fs.TempDirInTempDir(fmt.Sprintf("test-%v-remote", testName))
				require.NoError(t, err)
				defer func() { _ = fs.Rm(tmpRemoteDir) }()

				tmpSrcDir, err := fs.TempDirInTempDir(fmt.Sprintf("test-%v-local", testName))
				require.NoError(t, err)
				defer func() { _ = fs.Rm(tmpSrcDir) }()

				tmpDestDir, err := fs.TempDirInTempDir(fmt.Sprintf("test-%v-dest", testName))
				require.NoError(t, err)
				defer func() { _ = fs.Rm(tmpDestDir) }()
				items, err := fs.Ls(tmpDestDir)
				require.NoError(t, err)
				require.Empty(t, items)

				tree := filesystemtest.CreateTestFileTree(t, fs, tmpSrcDir, time.Now(), time.Now())
				expectedTree, err := fs.ConvertToRelativePath(tmpSrcDir, tree...)
				require.NoError(t, err)

				remoteCache, err := NewCache(cacheType, fs, &Configuration{
					RemoteStoragePath: tmpRemoteDir,
					Timeout:           time.Second,
				})
				require.NoError(t, err)
				count, err := remoteCache.EntriesCount(context.TODO())
				require.NoError(t, err)
				assert.Zero(t, count)

				key := remoteCache.GenerateKey("test", "cache", fmt.Sprintf("%v", cacheType))
				err = remoteCache.Store(ctx, key, tmpSrcDir)
				require.NoError(t, err)

				// check remote directory isn't empty
				count, err = remoteCache.EntriesCount(context.TODO())
				require.NoError(t, err)
				assert.Equal(t, int64(1), count)

				err = remoteCache.Fetch(ctx, key, tmpDestDir)
				require.NoError(t, err)
				items, err = fs.Ls(tmpDestDir)
				require.NoError(t, err)
				require.NotEmpty(t, items)
				var content []string
				err = fs.ListDirTree(tmpDestDir, &content)
				require.NoError(t, err)
				actualTree, err := fs.ConvertToRelativePath(tmpDestDir, content...)
				require.NoError(t, err)

				sort.Strings(expectedTree)
				sort.Strings(actualTree)
				require.Equal(t, expectedTree, actualTree)
			})
		}
	}
}

func TestSimpleCacheWorkflow_WithExcludedFilesystemItems(t *testing.T) { // Simple store, followed by fetch
	for c := range CacheTypes {
		cacheType := CacheTypes[c]
		for i := range filesystem.FileSystemTypes {
			fsType := filesystem.FileSystemTypes[i]
			if fsType == filesystem.InMemoryFS && cacheType == CacheMutable {
				// FIXME There is an error with lock unlock when using in memory fs
				continue
			}
			testName := fmt.Sprintf("%v_for_fs_%v_and_cache_%v", t.Name(), fsType, cacheType)
			t.Run(testName, func(t *testing.T) {
				t.Parallel()
				ctx := context.Background()
				fs := filesystem.NewFs(fsType)
				// set up temp remote directory
				tmpRemoteDir, err := fs.TempDirInTempDir(fmt.Sprintf("test-%v-remote", testName))
				require.NoError(t, err)
				defer func() { _ = fs.Rm(tmpRemoteDir) }()

				// Add random folders in the cache
				_, err = fs.TempDir(tmpRemoteDir, ".snapshot-to-ignore")
				require.NoError(t, err)
				_, err = fs.TempDir(tmpRemoteDir, "ignore-folder.snapshot-to")
				require.NoError(t, err)
				_, err = fs.TempDir(tmpRemoteDir, ".exclude-folder")
				require.NoError(t, err)
				_, err = fs.TempDir(tmpRemoteDir, "another-folder-to-exclude")
				require.NoError(t, err)
				f, err := fs.TempFile(tmpRemoteDir, ".ignore-file.*.test")
				require.NoError(t, err)
				require.NoError(t, f.Close())
				f, err = fs.TempFile(tmpRemoteDir, "another-file-to-exclude.*.test")
				require.NoError(t, err)
				require.NoError(t, f.Close())

				tmpSrcDir, err := fs.TempDirInTempDir(fmt.Sprintf("test-%v-local", testName))
				require.NoError(t, err)
				defer func() { _ = fs.Rm(tmpSrcDir) }()

				tmpDestDir, err := fs.TempDirInTempDir(fmt.Sprintf("test-%v-dest", testName))
				require.NoError(t, err)
				defer func() { _ = fs.Rm(tmpDestDir) }()
				items, err := fs.Ls(tmpDestDir)
				require.NoError(t, err)
				require.Empty(t, items)

				tree := filesystemtest.CreateTestFileTree(t, fs, tmpSrcDir, time.Now(), time.Now())
				expectedTree, err := fs.ConvertToRelativePath(tmpSrcDir, tree...)
				require.NoError(t, err)

				remoteCache, err := NewCache(cacheType, fs, &Configuration{
					RemoteStoragePath: tmpRemoteDir,
					Timeout:           time.Second,
				})
				require.NoError(t, err)
				count, err := remoteCache.EntriesCount(context.TODO())
				require.NoError(t, err)
				assert.Equal(t, int64(6), count)

				remoteCache, err = NewCache(cacheType, fs, &Configuration{
					RemoteStoragePath:       tmpRemoteDir,
					Timeout:                 time.Second,
					FilesystemItemsToIgnore: ".*exclude.*,.*ignore.*",
				})
				require.NoError(t, err)
				count, err = remoteCache.EntriesCount(context.TODO())
				require.NoError(t, err)
				assert.Zero(t, count)

				key := remoteCache.GenerateKey("test", "cache", fmt.Sprintf("%v", cacheType))
				err = remoteCache.Store(ctx, key, tmpSrcDir)
				require.NoError(t, err)

				// check remote directory isn't empty
				count, err = remoteCache.EntriesCount(context.TODO())
				require.NoError(t, err)
				assert.Equal(t, int64(1), count)

				err = remoteCache.Fetch(ctx, key, tmpDestDir)
				require.NoError(t, err)
				items, err = fs.Ls(tmpDestDir)
				require.NoError(t, err)
				require.NotEmpty(t, items)
				var content []string
				err = fs.ListDirTree(tmpDestDir, &content)
				require.NoError(t, err)
				actualTree, err := fs.ConvertToRelativePath(tmpDestDir, content...)
				require.NoError(t, err)

				sort.Strings(expectedTree)
				sort.Strings(actualTree)
				require.Equal(t, expectedTree, actualTree)
			})
		}
	}
}

func TestComplexCacheWorkflow(t *testing.T) { // Multiple Store action. The fetch should return the latest files stored in cache.
	for c := range CacheTypes {
		cacheType := CacheTypes[c]
		for i := range filesystem.FileSystemTypes {
			fsType := filesystem.FileSystemTypes[i]
			if fsType == filesystem.InMemoryFS && cacheType == CacheMutable {
				// FIXME There is an error with locks and in-memory fs (see details in ILock)
				continue
			}
			testName := fmt.Sprintf("%v_for_fs_%v_and_cache_%v", t.Name(), fsType, cacheType)
			t.Run(testName, func(t *testing.T) {
				t.Parallel()
				ctx := context.Background()
				fs := filesystem.NewFs(fsType)
				// set up temp remote directory
				tmpRemoteDir, err := fs.TempDirInTempDir(fmt.Sprintf("test-%v-remote", testName))
				require.NoError(t, err)
				defer func() { _ = fs.Rm(tmpRemoteDir) }()

				tmpSrcDir, err := fs.TempDirInTempDir(fmt.Sprintf("test-%v-local", testName))
				require.NoError(t, err)
				defer func() { _ = fs.Rm(tmpSrcDir) }()

				remoteCache, err := NewCache(cacheType, fs, &Configuration{
					RemoteStoragePath: tmpRemoteDir,
					Timeout:           time.Second,
				})
				require.NoError(t, err)
				count, err := remoteCache.EntriesCount(context.TODO())
				require.NoError(t, err)
				assert.Zero(t, count)

				var expectedTree []string
				key := remoteCache.GenerateKey("test", "cache", fmt.Sprintf("%v", cacheType))

				for i := 0; i < 10; i++ {
					err = fs.CleanDir(tmpSrcDir)
					require.NoError(t, err)
					tree := filesystemtest.CreateTestFileTree(t, fs, tmpSrcDir, time.Now(), time.Now())
					expectedTree, err = fs.ConvertToRelativePath(tmpSrcDir, tree...)
					require.NoError(t, err)
					err = remoteCache.Store(ctx, key, tmpSrcDir)
					require.NoError(t, err)
					time.Sleep(5 * time.Millisecond)
				}

				// check remote directory isn't empty
				count, err = remoteCache.EntriesCount(context.TODO())
				require.NoError(t, err)
				assert.Equal(t, int64(1), count)

				// Cleaning up src directory
				err = fs.CleanDir(tmpSrcDir)
				require.NoError(t, err)
				items, err := fs.Ls(tmpSrcDir)
				require.NoError(t, err)
				require.Empty(t, items)

				err = remoteCache.Fetch(ctx, key, tmpSrcDir)
				require.NoError(t, err)
				items, err = fs.Ls(tmpSrcDir)
				require.NoError(t, err)
				require.NotEmpty(t, items)
				var content []string
				err = fs.ListDirTree(tmpSrcDir, &content)
				require.NoError(t, err)
				actualTree, err := fs.ConvertToRelativePath(tmpSrcDir, content...)
				require.NoError(t, err)

				sort.Strings(expectedTree)
				sort.Strings(actualTree)
				require.Equal(t, expectedTree, actualTree)
			})
		}
	}
}

func TestComplexCacheWorkflowWithCleanCache(t *testing.T) { // Multiple Store action. The fetch should return the latest files stored in cache. A clean entry is performed after the multiple stores
	for c := range CacheTypes {
		cacheType := CacheTypes[c]
		for i := range filesystem.FileSystemTypes {
			fsType := filesystem.FileSystemTypes[i]
			if fsType == filesystem.InMemoryFS {
				// FIXME There is an error with locks and in-memory fs (see details in ILock)
				continue
			}
			testName := fmt.Sprintf("%v_for_fs_%v_and_cache_%v", t.Name(), fsType, cacheType)
			t.Run(testName, func(t *testing.T) {
				t.Parallel()
				ctx := context.Background()
				fs := filesystem.NewFs(fsType)
				// set up temp remote directory
				tmpRemoteDir, err := fs.TempDirInTempDir(fmt.Sprintf("test-%v-remote", testName))
				require.NoError(t, err)
				defer func() { _ = fs.Rm(tmpRemoteDir) }()

				tmpSrcDir, err := fs.TempDirInTempDir(fmt.Sprintf("test-%v-local", testName))
				require.NoError(t, err)
				defer func() { _ = fs.Rm(tmpSrcDir) }()

				remoteCache, err := NewCache(cacheType, fs, &Configuration{
					RemoteStoragePath: tmpRemoteDir,
					Timeout:           time.Second,
				})
				require.NoError(t, err)
				count, err := remoteCache.EntriesCount(context.TODO())
				require.NoError(t, err)
				assert.Zero(t, count)

				var expectedTree []string
				key := remoteCache.GenerateKey("test", "cache", fmt.Sprintf("%v", cacheType))

				for i := 0; i < 10; i++ {
					err = fs.CleanDir(tmpSrcDir)
					require.NoError(t, err)
					tree := filesystemtest.CreateTestFileTree(t, fs, tmpSrcDir, time.Now(), time.Now())
					expectedTree, err = fs.ConvertToRelativePath(tmpSrcDir, tree...)
					require.NoError(t, err)
					err = remoteCache.Store(ctx, key, tmpSrcDir)
					require.NoError(t, err)
				}
				err = remoteCache.CleanEntry(ctx, key)
				require.NoError(t, err)

				// check remote directory isn't empty
				count, err = remoteCache.EntriesCount(context.TODO())
				require.NoError(t, err)
				assert.Equal(t, int64(1), count)

				// Cleaning up src directory
				err = fs.CleanDir(tmpSrcDir)
				require.NoError(t, err)
				items, err := fs.Ls(tmpSrcDir)
				require.NoError(t, err)
				require.Empty(t, items)

				err = remoteCache.Fetch(ctx, key, tmpSrcDir)
				require.NoError(t, err)
				items, err = fs.Ls(tmpSrcDir)
				require.NoError(t, err)
				require.NotEmpty(t, items)
				var content []string
				err = fs.ListDirTree(tmpSrcDir, &content)
				require.NoError(t, err)
				actualTree, err := fs.ConvertToRelativePath(tmpSrcDir, content...)
				require.NoError(t, err)

				sort.Strings(expectedTree)
				sort.Strings(actualTree)
				require.Equal(t, expectedTree, actualTree)
			})
		}
	}
}

func TestRemoveEntry(t *testing.T) { // A store followed by a remove entry followed by a fetch
	for c := range CacheTypes {
		cacheType := CacheTypes[c]
		for i := range filesystem.FileSystemTypes {
			fsType := filesystem.FileSystemTypes[i]
			testName := fmt.Sprintf("%v_for_fs_%v_and_cache_%v", t.Name(), fsType, cacheType)
			t.Run(testName, func(t *testing.T) {
				t.Parallel()
				ctx := context.Background()
				fs := filesystem.NewFs(fsType)
				// set up temp remote directory
				tmpRemoteDir, err := fs.TempDirInTempDir(fmt.Sprintf("test-%v-remote", testName))
				require.NoError(t, err)
				defer func() { _ = fs.Rm(tmpRemoteDir) }()

				tmpSrcDir, err := fs.TempDirInTempDir(fmt.Sprintf("test-%v-local", testName))
				require.NoError(t, err)
				defer func() { _ = fs.Rm(tmpSrcDir) }()

				remoteCache, err := NewCache(cacheType, fs, &Configuration{
					RemoteStoragePath: tmpRemoteDir,
					Timeout:           time.Second,
				})
				require.NoError(t, err)
				count, err := remoteCache.EntriesCount(context.TODO())
				require.NoError(t, err)
				assert.Zero(t, count)

				key := remoteCache.GenerateKey("test", "cache", fmt.Sprintf("%v", cacheType))

				_ = filesystemtest.CreateTestFileTree(t, fs, tmpSrcDir, time.Now(), time.Now())
				err = remoteCache.Store(ctx, key, tmpSrcDir)
				require.NoError(t, err)

				// check remote directory isn't empty
				count, err = remoteCache.EntriesCount(context.TODO())
				require.NoError(t, err)
				assert.Equal(t, int64(1), count)

				// Cleaning up src directory
				err = fs.CleanDir(tmpSrcDir)
				require.NoError(t, err)
				items, err := fs.Ls(tmpSrcDir)
				require.NoError(t, err)
				require.Empty(t, items)

				err = remoteCache.RemoveEntry(ctx, key)
				require.NoError(t, err)
				count, err = remoteCache.EntriesCount(context.TODO())
				require.NoError(t, err)
				assert.Zero(t, count)

				err = remoteCache.Fetch(ctx, key, tmpSrcDir)
				require.NotNil(t, err)
				assert.True(t, commonerrors.Any(err, commonerrors.ErrNotFound, commonerrors.ErrEmpty))
			})
		}
	}
}

func TestEntryAge(t *testing.T) { // A store followed by a remove entry followed by a fetch
	for c := range CacheTypes {
		cacheType := CacheTypes[c]
		for i := range filesystem.FileSystemTypes {
			fsType := filesystem.FileSystemTypes[i]
			testName := fmt.Sprintf("%v_for_fs_%v_and_cache_%v", t.Name(), fsType, cacheType)
			t.Run(testName, func(t *testing.T) {
				t.Parallel()
				random := rand.New(rand.NewSource(time.Now().Unix())) //nolint:gosec //causes G404: Use of weak random number generator (math/rand instead of crypto/rand) (gosec), So disable gosec as this is just for
				ctx := context.Background()
				fs := filesystem.NewFs(fsType)
				// set up temp remote directory
				tmpRemoteDir, err := fs.TempDirInTempDir(fmt.Sprintf("test-%v-remote", testName))
				require.NoError(t, err)
				defer func() { _ = fs.Rm(tmpRemoteDir) }()

				tmpSrcDir, err := fs.TempDirInTempDir(fmt.Sprintf("test-%v-local", testName))
				require.NoError(t, err)
				defer func() { _ = fs.Rm(tmpSrcDir) }()

				remoteCache, err := NewCache(cacheType, fs, &Configuration{
					RemoteStoragePath: tmpRemoteDir,
					Timeout:           time.Second,
				})
				require.NoError(t, err)
				count, err := remoteCache.EntriesCount(context.TODO())
				require.NoError(t, err)
				assert.Zero(t, count)

				key := remoteCache.GenerateKey("test", "cache", fmt.Sprintf("%v", cacheType))

				_ = filesystemtest.CreateTestFileTree(t, fs, tmpSrcDir, time.Now(), time.Now())
				err = remoteCache.Store(ctx, key, tmpSrcDir)
				require.NoError(t, err)

				// check remote directory isn't empty
				count, err = remoteCache.EntriesCount(context.TODO())
				require.NoError(t, err)
				assert.Equal(t, int64(1), count)

				// Check entry age
				age, err := remoteCache.GetEntryAge(context.TODO(), key)
				require.NoError(t, err)
				assert.True(t, age < 5*time.Second)

				testAge := time.Duration(random.Intn(24)) * time.Hour //nolint:gosec //causes G404: Use of weak random number generator (math/rand instead of crypto/rand) (gosec), So disable gosec
				err = remoteCache.SetEntryAge(context.TODO(), key, testAge)
				require.NoError(t, err)
				age, err = remoteCache.GetEntryAge(context.TODO(), key)
				require.NoError(t, err)
				assert.Equal(t, int64(testAge.Seconds()), int64(age.Seconds()))

				err = remoteCache.RemoveEntry(ctx, key)
				require.NoError(t, err)
			})
		}
	}
}
