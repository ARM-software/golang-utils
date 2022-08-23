package sharedcache

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/filesystem"
)

func TestGenerateKey(t *testing.T) {
	builderID := "test-builder"
	projectPath := "project-path"

	// https://asecuritysite.com/encryption/xxHash
	// xxhash("") = ef46db3751d8e999
	// xxhash("ef46db3751d8e999" + "project-path") = 9eef8a610714c80f
	// xxhash("9eef8a610714c80f" + "test-builder") = 8dee155db58cff5d
	expectedHashValue := "8dee155db58cff5d"

	// check hash is as expected
	testHashValue := GenerateKey(projectPath, builderID)

	require.Equal(t, testHashValue, expectedHashValue)
}

func TestHashWithHashFileExists(t *testing.T) {
	for _, fsType := range filesystem.FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			t.Parallel()
			fs := filesystem.NewFs(fsType)

			// set up temp remote directory
			tmpRemoteDir, err := fs.TempDirInTempDir("test-hashFile")
			require.NoError(t, err)
			defer func() { _ = fs.Rm(tmpRemoteDir) }()

			testFile := filepath.Join(tmpRemoteDir, "test")

			// Add test file containing fake hash
			hashFilePath := fmt.Sprintf("%v%v", testFile, hashFileDescriptor)
			err = fs.WriteFile(hashFilePath, []byte("testtesttesttest"), 0755) // 16 chars long
			require.NoError(t, err)

			// check that it uses the hash in the hash file
			hash, err := getHash(context.TODO(), fs, testFile, false)
			require.NoError(t, err)
			require.Equal(t, "testtesttesttest", hash)

		})
	}
}

func TestHashWithHashFileNotExist(t *testing.T) {
	for _, fsType := range filesystem.FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			t.Parallel()
			fs := filesystem.NewFs(fsType)

			// set up temp remote directory
			tmpRemoteDir, err := fs.TempDirInTempDir("test-hashFile")
			require.NoError(t, err)
			defer func() { _ = fs.Rm(tmpRemoteDir) }()

			// paths for test file and eventual hash file
			testFile := filepath.Join(tmpRemoteDir, "test")
			hashFilePath := fmt.Sprintf("%v%v", testFile, hashFileDescriptor)
			err = fs.Rm(hashFilePath)
			require.NoError(t, err)

			// Add test file
			err = fs.WriteFile(testFile, []byte("test"), 0755)
			require.NoError(t, err)

			// check has is as expected
			hash, err := getHash(context.TODO(), fs, testFile, false)
			require.NoError(t, err)
			require.Equal(t, "4fdcca5ddb678139", hash) // hash for file = "test" containing "test"

		})
	}
}

func TestHashWithHashFileForceUpdate(t *testing.T) {
	for _, fsType := range filesystem.FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			t.Parallel()
			fs := filesystem.NewFs(fsType)

			// set up temp remote directory
			tmpRemoteDir, err := fs.TempDirInTempDir("test-hashFile")
			require.NoError(t, err)
			defer func() { _ = fs.Rm(tmpRemoteDir) }()

			testFile := filepath.Join(tmpRemoteDir, "test")

			// Add test file containing fake hash
			err = fs.WriteFile(testFile, []byte("test"), 0755) // 16 chars long
			require.NoError(t, err)

			// check that it uses correct hash not the fake hash
			hash, err := getHash(context.TODO(), fs, testFile, true)
			require.NoError(t, err)
			require.Equal(t, "4fdcca5ddb678139", hash) // hash for file = "test" containing "test"
		})
	}
}

func testTransfert(t *testing.T, ctx context.Context, fs filesystem.FS, folder1, dest1, dest2 string) {
	testFile := filepath.Join(folder1, "test-tranfer1")

	testContent := "test"
	// create file
	err := fs.WriteFile(testFile, []byte(testContent), 0755) // 16 chars long
	require.NoError(t, err)

	destFile, err := TransferFiles(ctx, fs, dest2, testFile)
	require.NoError(t, err)
	err = fs.Rm(testFile)
	require.NoError(t, err)

	testFile, err = TransferFiles(ctx, fs, dest1, destFile)
	require.NoError(t, err)
	content, err := fs.ReadFile(testFile)
	require.NoError(t, err)

	assert.Equal(t, testContent, string(content))
}

func TestTransferWithExistingDestination(t *testing.T) {
	for _, fsType := range filesystem.FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			fs := filesystem.NewFs(fsType)

			folder1, err := fs.TempDirInTempDir("test-tranfer1")
			require.NoError(t, err)
			defer func() { _ = fs.Rm(folder1) }()
			folder2, err := fs.TempDirInTempDir("test-tranfer2")
			require.NoError(t, err)
			defer func() { _ = fs.Rm(folder2) }()

			testTransfert(t, ctx, fs, folder1, folder1, folder2)
		})
	}
}

func TestTransferWithDestinationFiles(t *testing.T) {
	for _, fsType := range filesystem.FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			ctx := context.Background()
			fs := filesystem.NewFs(fsType)

			folder1, err := fs.TempDirInTempDir("test-tranfer1")
			require.NoError(t, err)
			defer func() { _ = fs.Rm(folder1) }()
			folder2, err := fs.TempDirInTempDir("test-tranfer2")
			require.NoError(t, err)
			defer func() { _ = fs.Rm(folder2) }()
			dest1 := filepath.Join(folder1, "test_dest_tranfer1.txt")
			dest2 := filepath.Join(folder2, "test_dest_tranfer2.txt")

			testTransfert(t, ctx, fs, folder1, dest1, dest2)
		})
	}
}

func TestTransferWithNonExistentDestinationFolders(t *testing.T) {
	for _, fsType := range filesystem.FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			ctx := context.Background()
			fs := filesystem.NewFs(fsType)

			folder1, err := fs.TempDirInTempDir("test-tranfer1")
			require.NoError(t, err)
			defer func() { _ = fs.Rm(folder1) }()
			folder2, err := fs.TempDirInTempDir("test-tranfer2")
			require.NoError(t, err)
			defer func() { _ = fs.Rm(folder2) }()
			dest1 := filepath.Join(folder1, "test_dest_tranfer1_dir/")
			dest2 := filepath.Join(folder2, "test_dest_tranfer2_dir/")

			testTransfert(t, ctx, fs, folder1, dest1, dest2)
		})
	}
}
