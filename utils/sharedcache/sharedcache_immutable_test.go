package sharedcache

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/filesystem"
	"github.com/ARM-software/golang-utils/utils/internal/testutils"
)

// listCompleteFilesByModTime returns a slice of filenames in a directory sorted by modification time
func TestListCompleteFilesByModTime(t *testing.T) {
	for i := range filesystem.FileSystemTypes {
		fsType := filesystem.FileSystemTypes[i]
		testName := fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType)
		t.Run(testName, func(t *testing.T) {
			fs := filesystem.NewFs(fsType)

			// Set up temp remote directory
			tmpRemoteDir, err := fs.TempDirInTempDir(fmt.Sprintf("test-%v-remote", testName))
			require.NoError(t, err)
			defer func() { _ = fs.Rm(tmpRemoteDir) }()

			// Create file1
			file1 := fmt.Sprintf("%v-1.zip", faker.Word())
			path1 := filepath.Join(tmpRemoteDir, file1)
			err = fs.WriteFile(path1, []byte("thing"), 0755)
			require.NoError(t, err)
			time.Sleep(50 * time.Millisecond)

			// Create file2
			file2 := fmt.Sprintf("%v-2.zip", faker.Word())
			path2 := filepath.Join(tmpRemoteDir, file2)
			err = fs.WriteFile(path2, []byte("stuff"), 0755)
			require.NoError(t, err)
			time.Sleep(50 * time.Millisecond)

			// run listCompleteFilesByModTime
			sorted, err := listCompleteFilesByModTime(context.TODO(), fs, tmpRemoteDir)
			require.NoError(t, err)

			// Assert that file2 was written after file1
			assert.Equal(t, file2, sorted[0])
			assert.Equal(t, file1, sorted[1])

			// Update file1
			err = fs.WriteFile(path1, []byte("something"), 0755)
			require.NoError(t, err)
			time.Sleep(50 * time.Millisecond)

			// run listCompleteFilesByModTime again
			sorted, err = listCompleteFilesByModTime(context.TODO(), fs, tmpRemoteDir)
			require.NoError(t, err)

			// Assert that now file1 was written after file2
			assert.Equal(t, file1, sorted[0])
			assert.Equal(t, file2, sorted[1])

		})
	}
}

func TestStoreImmutableCache(t *testing.T) {
	for i := range filesystem.FileSystemTypes {
		fsType := filesystem.FileSystemTypes[i]
		testName := fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType)
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

			_, err = testutils.CreateTestFileTree(fs, tmpSrcDir, time.Now(), time.Now())
			require.NoError(t, err)

			remoteCache, err := NewSharedImmutableCacheRepository(&Configuration{
				RemoteStoragePath: tmpRemoteDir,
			}, fs)
			require.NoError(t, err)
			count, err := remoteCache.EntriesCount(context.TODO())
			require.NoError(t, err)
			assert.Zero(t, count)

			key := remoteCache.GenerateKey("test", "cache")
			for i := 0; i < 5; i++ {
				// store src
				err = remoteCache.Store(ctx, key, tmpSrcDir)
				require.NoError(t, err)
				// sleep so files have different mod times
				time.Sleep(50 * time.Millisecond)
			}

			// check remote directory isn't empty
			count, err = remoteCache.EntriesCount(context.TODO())
			require.NoError(t, err)
			assert.Equal(t, int64(1), count)

			entryPath := remoteCache.getCacheEntryPath(key)
			files, err := fs.Ls(entryPath)
			require.NoError(t, err)
			require.NotEmpty(t, files)
			assert.Len(t, files, 5+5) // 5 zip files + 5 hash files
		})
	}
}

func TestCleanEntryImmutableCache(t *testing.T) {
	for i := range filesystem.FileSystemTypes {
		fsType := filesystem.FileSystemTypes[i]
		testName := fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType)
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

			_, err = testutils.CreateTestFileTree(fs, tmpSrcDir, time.Now(), time.Now())
			require.NoError(t, err)

			remoteCache, err := NewSharedImmutableCacheRepository(&Configuration{
				RemoteStoragePath: tmpRemoteDir,
			}, fs)
			require.NoError(t, err)
			count, err := remoteCache.EntriesCount(context.TODO())
			require.NoError(t, err)
			assert.Zero(t, count)

			key := remoteCache.GenerateKey("test", "cache")
			for i := 0; i < 5; i++ {
				// store src
				err = remoteCache.Store(ctx, key, tmpSrcDir)
				require.NoError(t, err)
				// sleep so files have different mod times
				time.Sleep(50 * time.Millisecond)
			}

			// create a fake part file
			entryPath := remoteCache.getCacheEntryPath(key)
			partFileName := remoteCache.generateCachedPackageName()
			file, err := fs.CreateFile(filepath.Join(entryPath, partFileName))
			require.NoError(t, err)
			_ = file.Close()

			// check remote directory isn't empty
			count, err = remoteCache.EntriesCount(context.TODO())
			require.NoError(t, err)
			assert.Equal(t, int64(1), count)
			files, err := fs.Ls(entryPath)
			require.NoError(t, err)
			require.NotEmpty(t, files)
			assert.Len(t, files, 5+5+1) // 5 zip files + 5 hash files + 1 part file

			err = remoteCache.CleanEntry(ctx, key)
			require.NoError(t, err)

			// check remote directory isn't empty
			count, err = remoteCache.EntriesCount(context.TODO())
			require.NoError(t, err)
			assert.Equal(t, int64(1), count)
			entryPath = remoteCache.getCacheEntryPath(key)
			files, err = fs.Ls(entryPath)
			require.NoError(t, err)
			require.NotEmpty(t, files)
			assert.Len(t, files, 2+1) // 1 zip file + 1 hash file + 1 part file
			assert.Contains(t, files, partFileName)
		})
	}
}
