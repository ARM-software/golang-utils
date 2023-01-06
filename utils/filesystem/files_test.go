/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package filesystem

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/bxcodec/faker/v3"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/idgen"
	"github.com/ARM-software/golang-utils/utils/platform"
	"github.com/ARM-software/golang-utils/utils/reflection"
	"github.com/ARM-software/golang-utils/utils/safeio"
)

func TestExists(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			tmpFile, err := fs.TempFileInTempDir("test-exist-")
			require.NoError(t, err)

			err = tmpFile.Close()
			require.NoError(t, err)

			fileName := tmpFile.Name()
			defer func() { _ = fs.Rm(fileName) }()
			assert.True(t, fs.Exists(tmpFile.Name()))

			randomFile, err := idgen.GenerateUUID4()
			require.NoError(t, err)
			assert.False(t, fs.Exists(randomFile))
			_ = fs.Rm(fileName)
		})
	}
}

func TestOpen(t *testing.T) {
	// Similar to https://github.com/spf13/afero/blob/787d034dfe70e44075ccc060d346146ef53270ad/afero_test.go#L79
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			tmpDir, err := fs.TempDirInTempDir("test-open-")
			require.NoError(t, err)
			defer func() { _ = fs.Rm(tmpDir) }()

			filePath := fmt.Sprintf("%v%v%v", tmpDir, string(fs.PathSeparator()), "test.txt")
			// According to documentation, if the file does not exist, and the O_CREATE flag
			//// is passed, it is created with mode perm (before umask)
			mode := 0600
			file, err := fs.OpenFile(filePath, os.O_RDWR|os.O_CREATE, os.FileMode(mode))
			require.NoError(t, err)
			defer func() { _ = file.Close() }()

			_, err = safeio.WriteString(context.TODO(), file, "initial")
			require.NoError(t, err)
			err = file.Close()
			require.NoError(t, err)

			assert.True(t, fs.Exists(filePath))

			testFileMode(t, fs, filePath, mode)

			ignoredMode := 0322
			file, err = fs.OpenFile(filePath, os.O_WRONLY|os.O_APPEND, os.FileMode(ignoredMode))
			require.NoError(t, err)
			defer func() { _ = file.Close() }()

			_, err = safeio.WriteString(context.TODO(), file, "|append")
			require.NoError(t, err)
			err = file.Close()
			require.NoError(t, err)
			testFileMode(t, fs, filePath, mode)

			ignoredMode = 0400
			file, err = fs.OpenFile(filePath, os.O_RDONLY, os.FileMode(ignoredMode))
			require.NoError(t, err)
			defer func() { _ = file.Close() }()

			contents, err := safeio.ReadAll(context.TODO(), file)
			require.NoError(t, err)
			expectedContents := "initial|append"
			assert.Equal(t, expectedContents, string(contents))
			err = file.Close()
			require.NoError(t, err)
			testFileMode(t, fs, filePath, mode)

			ignoredMode = 0664
			file, err = fs.OpenFile(filePath, os.O_RDWR|os.O_TRUNC, os.FileMode(ignoredMode))
			require.NoError(t, err)
			defer func() { _ = file.Close() }()

			testFileMode(t, fs, filePath, mode)

			contents, err = safeio.ReadAll(context.TODO(), file)
			require.NoError(t, err)
			err = file.Close()
			require.NoError(t, err)
			assert.Equal(t, "", string(contents))
			_ = fs.Rm(filePath)
			_ = fs.Rm(tmpDir)
		})
	}
}

func TestFileHandle(t *testing.T) {
	// Test for standard FS
	fs := NewFs(StandardFS)
	tmpFile1, err := fs.TempFileInTempDir("test-filehandle-*.txt")
	require.NoError(t, err)
	defer func() { _ = tmpFile1.Close() }()
	defer func() { _ = fs.Rm(tmpFile1.Name()) }()
	assert.False(t, IsFileHandleUnset(tmpFile1.Fd()))
	err = tmpFile1.Close()
	require.NoError(t, err)
	_ = fs.Rm(tmpFile1.Name())

	// Test for in memory FS
	fs = NewFs(InMemoryFS)
	tmpFile2, err := fs.TempFileInTempDir("test-filehandle-*.txt")
	require.NoError(t, err)
	defer func() { _ = tmpFile2.Close() }()
	defer func() { _ = fs.Rm(tmpFile2.Name()) }()
	assert.True(t, IsFileHandleUnset(tmpFile2.Fd()))
	err = tmpFile2.Close()
	require.NoError(t, err)
	_ = fs.Rm(tmpFile2.Name())
}

func TestCreate(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			tmpDir, err := fs.TempDirInTempDir("test-create-")
			require.NoError(t, err)
			defer func() { _ = fs.Rm(tmpDir) }()

			txt := "This is a test sentence!!!"
			filePath := fmt.Sprintf("%v%v%v", tmpDir, string(fs.PathSeparator()), "test.txt")
			err = fs.WriteFile(filePath, []byte(txt), 0755)
			require.NoError(t, err)

			assert.True(t, fs.Exists(filePath))

			bytes, err := fs.ReadFile(filePath)
			require.NoError(t, err)

			assert.Equal(t, txt, string(bytes))
			_ = fs.Rm(tmpDir)
		})
	}
}

func TestCancelledWrite(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			tmpDir, err := fs.TempDirInTempDir("test-cancel-write-")
			require.NoError(t, err)
			defer func() { _ = fs.Rm(tmpDir) }()

			txt := faker.Sentence()
			filePath := fmt.Sprintf("%v%v%v", tmpDir, string(fs.PathSeparator()), "test.txt")
			ctx, cancel := context.WithCancel(context.TODO())
			cancel()
			err = fs.WriteFileWithContext(ctx, filePath, []byte(txt), 0755)
			require.Error(t, err)
			assert.True(t, commonerrors.Any(err, commonerrors.ErrCancelled))
			if fs.Exists(filePath) {
				empty, err := fs.IsEmpty(filePath)
				require.NoError(t, err)

				bytes, err := fs.ReadFile(filePath)
				if empty {
					require.Error(t, err)
					assert.True(t, commonerrors.Any(err, commonerrors.ErrEmpty))
				} else {
					require.NoError(t, err)
					assert.NotEqual(t, txt, string(bytes))
				}
			}
			require.NoError(t, fs.Rm(tmpDir))
		})
	}
}

func TestCancelledRead(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			tmpDir, err := fs.TempDirInTempDir("test-cancel-read-")
			require.NoError(t, err)
			defer func() { _ = fs.Rm(tmpDir) }()

			txt := faker.Sentence()
			filePath := fmt.Sprintf("%v%v%v", tmpDir, string(fs.PathSeparator()), "test.txt")
			ctx, cancel := context.WithCancel(context.TODO())
			cancel()
			err = fs.WriteFile(filePath, []byte(txt), 0755)
			require.NoError(t, err)

			assert.True(t, fs.Exists(filePath))

			bytes, err := fs.ReadFile(filePath)
			require.NoError(t, err)

			assert.Equal(t, txt, string(bytes))

			bytes, err = fs.ReadFileWithContext(ctx, filePath)
			require.Error(t, err)
			assert.True(t, commonerrors.Any(err, commonerrors.ErrCancelled))
			assert.NotEqual(t, txt, string(bytes))

			require.NoError(t, fs.Rm(tmpDir))
		})
	}
}

func TestChmod(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			tmpDir, err := fs.TempDirInTempDir("test-chmod-")
			require.NoError(t, err)
			defer func() { _ = fs.Rm(tmpDir) }()

			filePath := fmt.Sprintf("%v%v%v", tmpDir, string(fs.PathSeparator()), "test.txt")
			file, err := fs.CreateFile(filePath)
			require.NoError(t, err)
			defer func() { _ = file.Close() }()
			err = file.Close()
			require.NoError(t, err)
			require.True(t, fs.Exists(filePath))
			for _, mode := range []int{0666, 0777, 0555, 0766, 0444, 0644} {
				err = fs.Chmod(filePath, os.FileMode(mode))
				if err != nil {
					_ = fs.Chmod(filePath, os.FileMode(mode))
				}
				require.NoError(t, err)
				testFileMode(t, fs, filePath, mode)
			}
			_ = fs.Rm(tmpDir)
		})
	}
}

func TestChown(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			tmpDir, err := fs.TempDirInTempDir("test-chown-")
			require.NoError(t, err)
			defer func() { _ = fs.Rm(tmpDir) }()

			filePath := fmt.Sprintf("%v%v%v", tmpDir, string(fs.PathSeparator()), "test.txt")
			file, err := fs.CreateFile(filePath)
			require.NoError(t, err)
			defer func() { _ = file.Close() }()
			err = file.Close()
			require.NoError(t, err)
			require.True(t, fs.Exists(filePath))
			uID, gID, err := fs.FetchOwners(filePath)
			if err != nil {
				assert.True(t, commonerrors.Any(err, commonerrors.ErrNotImplemented, commonerrors.ErrUnsupported))
			} else {
				err = fs.Chown(filePath, uID, gID)
				if err != nil {
					assert.True(t, commonerrors.Any(err, commonerrors.ErrNotImplemented, commonerrors.ErrUnsupported))
				} else {
					newUID, newGID, err := fs.FetchOwners(filePath)
					require.NoError(t, err)
					assert.Equal(t, uID, newUID)
					assert.Equal(t, gID, newGID)
				}
			}
			_ = fs.Rm(tmpDir)
		})
	}
}

func createTestFileTree(t *testing.T, fs FS, testDir, basePath string, withLinks bool, fileModTime time.Time, fileAccessTime time.Time) []string {
	err := fs.MkDir(testDir)
	require.NoError(t, err)

	var sLinks []string
	rand.Seed(time.Now().UnixMilli())                                        //nolint:gosec //causes G404: Use of weak random number generator (math/rand instead of crypto/rand) (gosec), So disable gosec
	for i := 0; i < int(math.Max(float64(1), float64(rand.Intn(10)))); i++ { //nolint:gosec //causes G404: Use of weak random number generator (math/rand instead of crypto/rand) (gosec), So disable gosec
		c := fmt.Sprint("test", i+1)
		path := filepath.Join(testDir, c)

		err := fs.MkDir(path)
		require.NoError(t, err)

		for j := 0; j < int(math.Max(float64(1), float64(rand.Intn(10)))); j++ { //nolint:gosec //causes G404: Use of weak random number generator (math/rand instead of crypto/rand) (gosec), So disable gosec
			c := fmt.Sprint("test", j+1)
			path := filepath.Join(path, c)

			err := fs.MkDir(path)
			require.NoError(t, err)

			if withLinks {
				if len(sLinks) > 0 {
					c1 := fmt.Sprint("link", j+1)
					c2 := filepath.Join(path, c1)
					err = fs.Symlink(sLinks[0], c2)
					require.NoError(t, err)
					if len(sLinks) > 1 {
						sLinks = sLinks[1:]
					} else {
						sLinks = nil
					}
				}
			}

			for k := 0; k < int(math.Max(float64(1), float64(rand.Intn(10)))); k++ { //nolint:gosec //causes G404: Use of weak random number generator (math/rand instead of crypto/rand) (gosec), So disable gosec
				c := fmt.Sprint("test", k+1, ".txt")
				finalPath := filepath.Join(path, c)

				// pick a couple of files to make symlinks (1 in 10)
				r := rand.Intn(10) //nolint:gosec //causes G404: Use of weak random number generator (math/rand instead of crypto/rand) (gosec), So disable gosec
				if r == 4 {
					fPath := filepath.Join(basePath, path, c)
					sLinks = append(sLinks, fPath)
				}

				s := fmt.Sprint("file ", i+1, j+1, k+1)
				err = fs.WriteFile(finalPath, []byte(s), 0755)
				require.NoError(t, err)
			}
		}
	}
	var tree []string
	err = fs.ListDirTree(testDir, &tree)
	require.NoError(t, err)

	// unifying timestamps
	for _, path := range tree {
		err = fs.Chtimes(path, fileAccessTime, fileModTime)
		require.NoError(t, err)
	}

	return tree
}

func TestConvertPaths(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)

			// set up temp directory
			tmpDir, err := fs.TempDirInTempDir("temp")
			require.NoError(t, err)
			defer func() { _ = fs.Rm(tmpDir) }()

			// create a directory for the test
			tree := createTestFileTree(t, fs, tmpDir, "", false, time.Now(), time.Now())
			relTree, err := fs.ConvertToRelativePath(tmpDir, tree...)
			require.NoError(t, err)
			absTree, err := fs.ConvertToAbsolutePath(tmpDir, relTree...)
			require.NoError(t, err)

			// sort so the list of directories can be evaluated more easily
			sort.Strings(tree)
			sort.Strings(relTree)
			sort.Strings(absTree)

			// check if equal.
			require.Equal(t, absTree, tree)
			require.NotEqual(t, relTree, tree)
		})
	}
}

func TestLink(t *testing.T) {
	if platform.IsWindows() {
		fmt.Println("In order to run TestLink on Windows, Developer mode must be enabled: https://github.com/golang/go/pull/24307")
	}
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			tmpDir, err := fs.TempDirInTempDir("test-link-")
			require.NoError(t, err)
			defer func() { _ = fs.Rm(tmpDir) }()

			txt := "This is a test sentence!!!"
			filePath := fmt.Sprintf("%v%v%v", tmpDir, string(fs.PathSeparator()), "test.txt")
			err = fs.WriteFile(filePath, []byte(txt), 0755)
			require.NoError(t, err)

			symlink := filepath.Join(tmpDir, "symlink-tofile")
			hardlink := filepath.Join(tmpDir, "hardlink-tofile")

			err = fs.Symlink(filePath, symlink)
			if errors.Is(err, commonerrors.ErrNotImplemented) || errors.Is(err, afero.ErrNoSymlink) {
				return
			}
			require.NoError(t, err)

			err = fs.Link(filePath, hardlink)
			require.NoError(t, err)

			assert.True(t, fs.Exists(symlink))
			assert.True(t, fs.Exists(hardlink))

			isLink, err := fs.IsLink(symlink)
			require.NoError(t, err)
			assert.True(t, isLink)

			isFile, err := fs.IsFile(symlink)
			require.NoError(t, err)
			assert.True(t, isFile)

			isLink, err = fs.IsLink(hardlink)
			require.NoError(t, err)
			assert.False(t, isLink)

			isFile, err = fs.IsFile(hardlink)
			require.NoError(t, err)
			assert.True(t, isFile)

			link, err := fs.Readlink(symlink)
			require.NoError(t, err)
			assert.Equal(t, filePath, link)

			link, err = fs.Readlink(hardlink)
			require.NotNil(t, err)
			assert.Equal(t, "", link)

			bytes, err := fs.ReadFile(symlink)
			require.NoError(t, err)
			assert.Equal(t, txt, string(bytes))

			bytes, err = fs.ReadFile(hardlink)
			require.NoError(t, err)
			assert.Equal(t, txt, string(bytes))
			_ = fs.Rm(tmpDir)
		})
	}
}

func TestStatTimes(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			tmpFile, err := fs.TempFileInTempDir("test-times-")
			require.NoError(t, err)

			err = tmpFile.Close()
			require.NoError(t, err)
			defer func() { _ = fs.Rm(tmpFile.Name()) }()

			fileName := tmpFile.Name()
			defer func() { _ = fs.Rm(fileName) }()
			assert.True(t, fs.Exists(tmpFile.Name()))
			fileTimes, err := fs.StatTimes(fileName)
			require.NoError(t, err)
			assert.NotZero(t, fileTimes)
			assert.NotZero(t, fileTimes.AccessTime())
			assert.NotZero(t, fileTimes.ModTime())
			if fileTimes.HasBirthTime() {
				assert.NotZero(t, fileTimes.BirthTime())
			}
			if fileTimes.HasChangeTime() {
				assert.NotZero(t, fileTimes.ChangeTime())
			}
			_ = fs.Rm(tmpFile.Name())
		})
	}
}

func TestIsEmpty(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			tmpDir, err := fs.TempDirInTempDir("test-isempty-")
			require.NoError(t, err)
			defer func() { _ = fs.Rm(tmpDir) }()

			empty, err := fs.IsEmpty(tmpDir)
			require.NoError(t, err)
			assert.True(t, empty)

			tmpFile, err := fs.TempFile(tmpDir, "test-isempty-*.test")
			require.NoError(t, err)
			_ = tmpFile.Close()
			defer func() { _ = fs.Rm(tmpFile.Name()) }()

			isFileEmpty, err := fs.IsEmpty(tmpFile.Name())
			require.NoError(t, err)
			assert.True(t, isFileEmpty)

			checkNotEmpty(t, fs, tmpDir)

			testDir := filepath.Join(tmpDir, "testDir")
			err = fs.MkDir(testDir)
			require.NoError(t, err)
			defer func() { _ = fs.Rm(testDir) }()

			checkNotEmpty(t, fs, tmpDir)

			err = fs.Rm(testDir)
			require.NoError(t, err)

			checkNotEmpty(t, fs, tmpDir)
			_ = fs.Rm(tmpDir)
		})
	}
}

func TestChtimes(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			tmpDir, err := fs.TempDirInTempDir("test-isempty-")
			require.NoError(t, err)
			defer func() { _ = fs.Rm(tmpDir) }()

			tmpFile, err := fs.TempFile(tmpDir, "test-isempty-*.test")
			require.NoError(t, err)
			_ = tmpFile.Close()
			defer func() { _ = fs.Rm(tmpFile.Name()) }()

			timesDirOrig, err := fs.StatTimes(tmpDir)
			require.NoError(t, err)
			timesFileOrig, err := fs.StatTimes(tmpFile.Name())
			require.NoError(t, err)
			newTimeA := time.Now().Add(-time.Hour)
			newTimeM := time.Now().Add(-30 * time.Minute)
			err = fs.Chtimes(tmpDir, newTimeA, newTimeM)
			require.NoError(t, err)
			err = fs.Chtimes(tmpFile.Name(), newTimeA, newTimeM)
			require.NoError(t, err)

			timesDirMod, err := fs.StatTimes(tmpDir)
			require.NoError(t, err)
			timesFileMod, err := fs.StatTimes(tmpFile.Name())
			require.NoError(t, err)

			assert.NotEqual(t, timesDirOrig, timesDirMod)
			assert.True(t, newTimeM.Equal(timesDirMod.ModTime()))
			if timesDirMod.HasAccessTime() {
				assert.True(t, newTimeA.Equal(timesDirMod.AccessTime()))
			}

			assert.NotEqual(t, timesFileOrig, timesFileMod)
			assert.True(t, newTimeM.Equal(timesFileMod.ModTime()))
			if timesFileMod.HasAccessTime() {
				assert.True(t, newTimeA.Equal(timesFileMod.AccessTime()))
			}
			err = fs.Rm(tmpDir)
			require.NoError(t, err)
		})
	}
}

func TestCleanDir(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			tmpDir, err := fs.TempDirInTempDir("test-cleandir-")
			require.NoError(t, err)
			defer func() { _ = fs.Rm(tmpDir) }()

			empty, err := fs.IsEmpty(tmpDir)
			require.NoError(t, err)
			assert.True(t, empty)

			tmpFile, err := fs.TempFile(tmpDir, "test-cleandir-*.test")
			require.NoError(t, err)
			err = tmpFile.Close()
			require.NoError(t, err)

			checkNotEmpty(t, fs, tmpDir)

			testDir := filepath.Join(tmpDir, "testDir")
			err = fs.MkDir(testDir)
			require.NoError(t, err)

			checkNotEmpty(t, fs, tmpDir)

			err = fs.CleanDir(tmpDir)
			require.NoError(t, err)
			assert.True(t, fs.Exists(tmpDir))

			empty, err = fs.IsEmpty(tmpDir)
			require.NoError(t, err)
			assert.True(t, empty)
			_ = fs.Rm(tmpDir)
		})
	}
}

func TestCleanDirWithExclusion(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			tmpDir, err := fs.TempDirInTempDir("test-cleandir-with-exclusion-")
			require.NoError(t, err)
			defer func() { _ = fs.Rm(tmpDir) }()

			empty, err := fs.IsEmpty(tmpDir)
			require.NoError(t, err)
			assert.True(t, empty)

			tmpFile, err := fs.TempFile(tmpDir, "test-cleandir-*.test")
			require.NoError(t, err)
			err = tmpFile.Close()
			require.NoError(t, err)

			checkNotEmpty(t, fs, tmpDir)

			testDir := filepath.Join(tmpDir, "testDirToIgnorePlease")
			err = fs.MkDir(testDir)
			require.NoError(t, err)

			err = fs.MkDir(filepath.Join(tmpDir, "testDirToRemove"))
			require.NoError(t, err)

			checkNotEmpty(t, fs, tmpDir)

			err = fs.CleanDirWithContextAndExclusionPatterns(context.TODO(), tmpDir, ".*ToIgnore.*")
			require.NoError(t, err)
			assert.True(t, fs.Exists(tmpDir))

			empty, err = fs.IsEmpty(tmpDir)
			require.NoError(t, err)
			assert.False(t, empty)

			var tree []string
			err = fs.ListDirTree(tmpDir, &tree)
			require.NoError(t, err)
			require.Len(t, tree, 1)
			assert.Equal(t, "testDirToIgnorePlease", filepath.Base(tree[0]))
			_ = fs.Rm(tmpDir)
		})
	}
}
func TestRemoveWithExclusion(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			tmpDir, err := fs.TempDirInTempDir("test-rm-with-exclusion-")
			require.NoError(t, err)
			defer func() { _ = fs.Rm(tmpDir) }()

			empty, err := fs.IsEmpty(tmpDir)
			require.NoError(t, err)
			assert.True(t, empty)

			tmpFile, err := fs.TempFile(tmpDir, "test-rm-*.test")
			require.NoError(t, err)
			err = tmpFile.Close()
			require.NoError(t, err)

			checkNotEmpty(t, fs, tmpDir)

			testDir := filepath.Join(tmpDir, "testDirToIgnorePlease")
			err = fs.MkDir(testDir)
			require.NoError(t, err)

			err = fs.MkDir(filepath.Join(tmpDir, "testDirToRemove"))
			require.NoError(t, err)

			checkNotEmpty(t, fs, tmpDir)

			err = fs.RemoveWithContextAndExclusionPatterns(context.TODO(), tmpDir, ".*ToIgnore.*")
			require.NoError(t, err)
			assert.True(t, fs.Exists(tmpDir))

			empty, err = fs.IsEmpty(tmpDir)
			require.NoError(t, err)
			assert.False(t, empty)

			var tree []string
			err = fs.ListDirTree(tmpDir, &tree)
			require.NoError(t, err)
			require.Len(t, tree, 1)
			assert.Equal(t, "testDirToIgnorePlease", filepath.Base(tree[0]))

			err = fs.RemoveWithContextAndExclusionPatterns(context.TODO(), tmpDir, "test-rm-with-exclusion-.*")
			require.NoError(t, err)
			assert.True(t, fs.Exists(tmpDir))
			empty, err = fs.IsEmpty(tmpDir)
			require.NoError(t, err)
			assert.True(t, empty)

			_ = fs.Rm(tmpDir)
		})
	}
}

func TestLs(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			tmpDir, err := fs.TempDirInTempDir("test-ls-")
			require.NoError(t, err)
			defer func() { _ = fs.Rm(tmpDir) }()

			tmpDir1, err := fs.TempDir(tmpDir, "test-ls-")
			require.NoError(t, err)
			tmpFile, err := fs.TempFile(tmpDir, "test-ls-*.test")
			require.NoError(t, err)

			err = tmpFile.Close()
			require.NoError(t, err)

			fileName := tmpFile.Name()

			files, err := fs.Ls(tmpDir)
			require.NoError(t, err)
			assert.Len(t, files, 2)
			_, found := collection.Find(&files, filepath.Base(fileName))
			assert.True(t, found)
			_, found = collection.Find(&files, filepath.Base(tmpDir1))
			assert.True(t, found)
			_ = fs.Rm(tmpDir)
		})
	}
}

func TestLsWithExclusion(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			tmpDir, err := fs.TempDirInTempDir("test-ls-with-exclusion-")
			require.NoError(t, err)
			defer func() { _ = fs.Rm(tmpDir) }()

			tmpDir1, err := fs.TempDir(tmpDir, "test-ls-")
			require.NoError(t, err)
			tmpFile, err := fs.TempFile(tmpDir, "test-ls-*.test")
			require.NoError(t, err)

			err = tmpFile.Close()
			require.NoError(t, err)

			fileName := tmpFile.Name()

			files, err := fs.Ls(tmpDir)
			require.NoError(t, err)
			assert.Len(t, files, 2)

			files, err = fs.LsWithExclusionPatterns(tmpDir, ".*[.]test")
			require.NoError(t, err)
			assert.Len(t, files, 1)

			_, found := collection.Find(&files, filepath.Base(fileName))
			assert.False(t, found)
			_, found = collection.Find(&files, filepath.Base(tmpDir1))
			assert.True(t, found)
			_ = fs.Rm(tmpDir)
		})
	}
}

func TestWalk(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)

			// set up temp directory
			tmpDir, err := fs.TempDirInTempDir("test_walk")
			require.NoError(t, err)
			defer func() { _ = fs.Rm(tmpDir) }()

			testDir := filepath.Join(tmpDir, "test")
			_ = fs.Rm(testDir)

			// create a directory for the test
			tree := createTestFileTree(t, fs, testDir, "", false, time.Now(), time.Now())
			tree = append(tree, testDir) // Walk requires root too

			var walkList []string
			err = fs.Walk(testDir, func(path string, info os.FileInfo, err error) error {
				walkList = append(walkList, path)
				return nil
			})
			require.NoError(t, err)

			// Sort lists so that equal works better
			sort.Strings(tree)
			sort.Strings(walkList)

			// check if equal
			require.Equal(t, walkList, tree)
		})
	}
}

func TestWalkWithExclusions(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)

			// set up temp directory
			tmpDir, err := fs.TempDirInTempDir("test_walk_with_exclusion")
			require.NoError(t, err)
			defer func() { _ = fs.Rm(tmpDir) }()

			testDir := filepath.Join(tmpDir, "tree")
			_ = fs.Rm(testDir)

			// create a directory for the test
			tree := createTestFileTree(t, fs, testDir, "", false, time.Now(), time.Now())
			tree = append(tree, testDir) // Walk requires root too

			exclusionPatterns := ".*test2.*"
			var walkList []string
			walkFunc := func(path string, info os.FileInfo, err error) error {
				walkList = append(walkList, path)
				return nil
			}
			err = fs.WalkWithContextAndExclusionPatterns(context.TODO(), testDir, walkFunc, exclusionPatterns)
			require.NoError(t, err)

			cleansedTree, err := ExcludeAll(tree, exclusionPatterns)
			require.NoError(t, err)
			assert.GreaterOrEqual(t, len(tree), len(cleansedTree))

			// Sort lists so that equal works better
			sort.Strings(cleansedTree)
			sort.Strings(walkList)

			// check if equal
			require.Equal(t, walkList, cleansedTree)

			walkList = nil
			err = fs.WalkWithContextAndExclusionPatterns(context.TODO(), testDir, walkFunc, ".*tree.*")
			require.NoError(t, err)
			assert.Empty(t, walkList)
		})
	}
}

func TestGarbageCollection(t *testing.T) {
	for _, fsType := range []FilesystemType{StandardFS} {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			tmpDir, err := fs.TempDirInTempDir("test-gc-")
			require.NoError(t, err)
			defer func() { _ = fs.Rm(tmpDir) }()

			tmpDir1, err := fs.TempDir(tmpDir, "test-gc-")
			require.NoError(t, err)
			tmpDir2, err := fs.TempDir(tmpDir, "test-gc-")
			require.NoError(t, err)
			tmpFile, err := fs.TempFile(tmpDir, "test-gc-*.test")
			require.NoError(t, err)

			err = tmpFile.Close()
			require.NoError(t, err)
			tmpFile1, err := fs.TempFile(tmpDir1, "test-gc-*.test")
			require.NoError(t, err)
			err = tmpFile1.Close()
			require.NoError(t, err)

			ctime := time.Now()
			time.Sleep(500 * time.Millisecond)

			tmpDir3, err := fs.TempDir(tmpDir, "test-gc-")
			require.NoError(t, err)
			tmpFile2, err := fs.TempFile(tmpDir3, "test-gc-*.test")
			require.NoError(t, err)
			err = tmpFile2.Close()
			require.NoError(t, err)
			tmpFile3, err := fs.TempFile(tmpDir2, "test-gc-*.test")
			require.NoError(t, err)
			err = tmpFile3.Close()
			require.NoError(t, err)

			files, err := fs.Ls(tmpDir)
			require.NoError(t, err)
			assert.Len(t, files, 4)

			elapsedTime := time.Since(ctime)

			err = fs.GarbageCollect(tmpDir, elapsedTime)
			require.NoError(t, err)

			files, err = fs.Ls(tmpDir)
			require.NoError(t, err)
			assert.Len(t, files, 2)
			_, found := collection.Find(&files, filepath.Base(tmpDir2))
			assert.True(t, found)
			_, found = collection.Find(&files, filepath.Base(tmpDir3))
			assert.True(t, found)
			files, err = fs.Ls(tmpDir3)
			require.NoError(t, err)
			assert.Len(t, files, 1)
			_, found = collection.Find(&files, filepath.Base(tmpFile2.Name()))
			assert.True(t, found)
			_ = fs.Rm(tmpDir)
		})
	}
}

func TestFindAll(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			tmpDir, err := fs.TempDirInTempDir("test-findall-")
			require.NoError(t, err)
			defer func() { _ = fs.Rm(tmpDir) }()

			empty, err := fs.IsEmpty(tmpDir)
			require.NoError(t, err)
			assert.True(t, empty)

			level1 := filepath.Join(tmpDir, "level1")
			err = fs.MkDir(level1)
			require.NoError(t, err)

			level2 := filepath.Join(level1, "level2")
			err = fs.MkDir(level2)
			require.NoError(t, err)

			level3 := filepath.Join(level2, "level3")
			err = fs.MkDir(level3)
			require.NoError(t, err)

			checkNotEmpty(t, fs, tmpDir)

			tmpFile1, err := fs.TempFile(level1, "test-findall-*.test1")
			require.NoError(t, err)
			err = tmpFile1.Close()
			require.NoError(t, err)

			tmpFile2, err := fs.TempFile(level2, "test-findall-*.test2")
			require.NoError(t, err)
			err = tmpFile2.Close()
			require.NoError(t, err)

			tmpFile3, err := fs.TempFile(level3, "test-findall-*.test3")
			require.NoError(t, err)
			err = tmpFile3.Close()
			require.NoError(t, err)

			var tree []string
			require.NoError(t, fs.ListDirTreeWithContextAndExclusionPatterns(context.TODO(), tmpDir, &tree, ".*[.]test[^13]"))
			require.NotEmpty(t, tree)
			assert.Len(t, tree, 5)

			list, err := fs.FindAll(tmpDir, ".test1", "test3")
			require.NoError(t, err)
			assert.Len(t, list, 2)

			for _, file := range list {
				ext := filepath.Ext(file)
				assert.Equal(t, ".test", ext[0:len(ext)-1])
				assert.True(t, ext[len(ext)-1] == '1' || ext[len(ext)-1] == '3')
			}
			err = fs.Rm(tmpDir)
			require.NoError(t, err)
		})
	}
}

func TestCopy(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			tmpDir, err := fs.TempDirInTempDir("test-copy-")
			require.NoError(t, err)
			defer func() { _ = fs.Rm(tmpDir) }()

			empty, err := fs.IsEmpty(tmpDir)
			require.NoError(t, err)
			assert.True(t, empty)

			tmpFile, err := fs.TempFile(tmpDir, "test-copy-*.test")
			require.NoError(t, err)
			err = tmpFile.Close()
			require.NoError(t, err)

			checkNotEmpty(t, fs, tmpDir)
			checkCopy(t, fs, tmpFile.Name(), filepath.Join(tmpDir, "newDir"))
			_ = fs.Rm(tmpDir)
		})
	}
}

func TestCopyWithExclusion(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			tmpDir, err := fs.TempDirInTempDir("test-copy-with-exclusion-")
			require.NoError(t, err)
			defer func() { _ = fs.Rm(tmpDir) }()

			empty, err := fs.IsEmpty(tmpDir)
			require.NoError(t, err)
			assert.True(t, empty)

			tmpFile, err := fs.TempFile(tmpDir, "test-copy-*.test")
			require.NoError(t, err)
			err = tmpFile.Close()
			require.NoError(t, err)

			checkNotEmpty(t, fs, tmpDir)
			checkCopy(t, fs, tmpFile.Name(), filepath.Join(tmpDir, "newDir"), "test-copy-with-exclusion-.*")
			_ = fs.Rm(tmpDir)
		})
	}
}

func TestCopyFolder(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			tmpDir, err := fs.TempDirInTempDir("test-src-copy-")
			require.NoError(t, err)

			parentDir, err := fs.TempDir(tmpDir, "parentDir-")
			require.NoError(t, err)

			childDir, err := fs.TempDir(parentDir, "childDir-")
			require.NoError(t, err)

			_, err = fs.TempDir(childDir, "gcDir-")
			require.NoError(t, err)

			testDir, err := fs.TempDirInTempDir("test-dest-dir-")
			require.NoError(t, err)
			defer func() {
				_ = fs.Rm(tmpDir)
				_ = fs.Rm(testDir)
			}()

			checkNotEmpty(t, fs, parentDir)
			checkCopyDir(t, fs, parentDir, filepath.Join(testDir, "newDir"))
		})
	}
}

func TestCopyFolderWithExclusion(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			tmpDir, err := fs.TempDirInTempDir("test-src-copy-with-exclusion-")
			require.NoError(t, err)
			defer func() {
				_ = fs.Rm(tmpDir)
			}()

			parentDir, err := fs.TempDir(tmpDir, "parentDir-")
			require.NoError(t, err)

			childDir, err := fs.TempDir(parentDir, "childDir-")
			require.NoError(t, err)

			_, err = fs.TempDir(childDir, "gcDir-")
			require.NoError(t, err)

			testDir, err := fs.TempDirInTempDir("test-dest-dir-")
			require.NoError(t, err)
			defer func() {
				_ = fs.Rm(testDir)
			}()

			checkNotEmpty(t, fs, parentDir)
			checkCopyDir(t, fs, parentDir, filepath.Join(testDir, "newDir"), "childDir-.*")
		})
	}
}

func TestMove(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			tmpDir, err := fs.TempDirInTempDir("test-move-")
			require.NoError(t, err)
			defer func() { _ = fs.Rm(tmpDir) }()

			empty, err := fs.IsEmpty(tmpDir)
			require.NoError(t, err)
			assert.True(t, empty)

			tmpFile, err := fs.TempFile(tmpDir, "test-move-*.test")
			require.NoError(t, err)
			err = tmpFile.Close()
			require.NoError(t, err)

			testDir := filepath.Join(tmpDir, "testDir")
			err = fs.MkDir(testDir)
			require.NoError(t, err)

			tmpFile2, err := fs.TempFile(testDir, "test-move-*.test")
			require.NoError(t, err)
			err = tmpFile2.Close()
			require.NoError(t, err)

			testDir2 := filepath.Join(testDir, "testDir2")
			err = fs.MkDir(testDir2)
			require.NoError(t, err)
			checkNotEmpty(t, fs, tmpDir)
			err = fs.Move(tmpFile.Name(), tmpFile.Name())
			require.NoError(t, err)
			err = fs.Move(testDir, testDir)
			require.NoError(t, err)
			checkMove(t, fs, tmpFile.Name(), filepath.Join(tmpDir, "test.test"))
			checkMove(t, fs, testDir, filepath.Join(tmpDir, "testDir3"))
			_ = fs.Rm(tmpDir)
		})
	}
}

func TestReadFile(t *testing.T) {
	fs := NewFs(InMemoryFS)
	tmpFile, err := fs.TempFileInTempDir("test-readfile-")
	require.NoError(t, err)

	byteInput := []byte("Here is a Test string....")
	_, err = tmpFile.Write(byteInput)
	require.NoError(t, err)

	err = tmpFile.Close()
	require.NoError(t, err)

	fileName := tmpFile.Name()
	defer func() { _ = fs.Rm(fileName) }()

	byteOut, err := fs.ReadFile(tmpFile.Name())
	require.NoError(t, err)
	assert.Equal(t, byteOut, byteInput)

	_, err = fs.ReadFile("unknown_file")
	assert.NotNil(t, err)
}

func TestGetFileSize(t *testing.T) {
	fs := NewFs(InMemoryFS)
	tmpFile, err := fs.TempFileInTempDir("test-filesize-")
	require.NoError(t, err)

	for indx := 0; indx < 50; indx++ {
		_, _ = tmpFile.WriteString(" Here is a Test string....")
	}

	err = tmpFile.Close()
	require.NoError(t, err)

	fileName := tmpFile.Name()
	defer func() { _ = fs.Rm(fileName) }()

	size, err := fs.GetFileSize(tmpFile.Name())
	require.NoError(t, err)
	assert.Equal(t, int64(1300), size)

	_, err = fs.GetFileSize("Unknown-File")
	assert.NotNil(t, err)
}

func TestSubDirectories(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			tmpDir, err := fs.TempDirInTempDir("test-subdirectories-")
			require.NoError(t, err)

			// Test empty directory
			dirlist, err := fs.SubDirectories(tmpDir)
			assert.Empty(t, dirlist)
			assert.NoError(t, err)
			empty, err := fs.IsEmpty(tmpDir)
			assert.NoError(t, err)
			assert.True(t, empty)

			// Test directory with subdirectories
			testInput := filepath.Join(tmpDir, "ARM")
			require.NoError(t, fs.MkDir(filepath.Join(testInput, ".CMSIS")))
			require.NoError(t, fs.MkDir(filepath.Join(testInput, ".git")))
			require.NoError(t, fs.MkDir(filepath.Join(testInput, "1Test")))
			require.NoError(t, fs.MkDir(filepath.Join(testInput, "Test.Test")))
			require.NoError(t, fs.MkDir(filepath.Join(testInput, "CMSIS")))
			require.NoError(t, fs.MkDir(filepath.Join(testInput, "Tools")))
			file, err := fs.CreateFile(filepath.Join(testInput, "testfile.ini"))
			require.NoError(t, err)
			require.NoError(t, file.Close())

			dirlist, err = fs.SubDirectories(testInput)
			assert.NoError(t, err)
			assert.Len(t, dirlist, 4)
			sort.Strings(dirlist)

			expected := []string{"1Test", "CMSIS", "Test.Test", "Tools"}
			assert.Equal(t, expected, dirlist)
			require.NoError(t, fs.Rm(testInput))
		})
	}
}

func TestSubDirectoriesWithExclusion(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			tmpDir, err := fs.TempDirInTempDir("test-subdirectories-with-exclusion-")
			require.NoError(t, err)

			// Test directory with subdirectories
			testInput := filepath.Join(tmpDir, "ARM")
			require.NoError(t, fs.MkDir(filepath.Join(testInput, ".CMSIS")))
			require.NoError(t, fs.MkDir(filepath.Join(testInput, ".git")))
			require.NoError(t, fs.MkDir(filepath.Join(testInput, "CMSIS")))
			require.NoError(t, fs.MkDir(filepath.Join(testInput, "Tools")))
			file, err := fs.CreateFile(filepath.Join(testInput, "testfile.ini"))
			require.NoError(t, err)
			require.NoError(t, file.Close())

			dirlist, err := fs.SubDirectoriesWithContextAndExclusionPatterns(context.TODO(), testInput, ".*CMSIS.*")
			assert.NoError(t, err)
			assert.Len(t, dirlist, 2)
			sort.Strings(dirlist)

			expected := []string{".git", "Tools"}
			assert.Equal(t, expected, dirlist)

			require.NoError(t, fs.Rm(testInput))
		})
	}
}

func TestListDirTree(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			testDir, err := fs.TempDirInTempDir("test-list-tree")
			require.NoError(t, err)
			defer func() { _ = fs.Rm(testDir) }()

			empty, err := fs.IsEmpty(testDir)
			require.NoError(t, err)
			assert.True(t, empty)

			parentDirPath, err := fs.TempDir(testDir, "parentDir-")
			require.NoError(t, err)

			testFile, err := fs.TempFile(testDir, "test-file-*.test")
			require.NoError(t, err)
			err = testFile.Close()
			require.NoError(t, err)

			childDirPath, err := fs.TempDir(parentDirPath, "childDir-")
			require.NoError(t, err)

			gcDirPath, err := fs.TempDir(childDirPath, "gcDir-")
			require.NoError(t, err)

			checkNotEmpty(t, fs, testDir)

			var list []string
			err = fs.ListDirTree(testDir, &list)
			assert.Len(t, list, 4)
			require.NoError(t, err)

			parentDir := filepath.Base(parentDirPath)
			childDir := filepath.Base(childDirPath)
			gcDir := filepath.Base(gcDirPath)
			testFileName := filepath.Base(testFile.Name())

			var fileList []string
			var expectedList = []string{
				filepath.Join(string(fs.PathSeparator()), parentDir),
				filepath.Join(string(fs.PathSeparator()), parentDir, childDir),
				filepath.Join(string(fs.PathSeparator()), parentDir, childDir, gcDir),
				filepath.Join(string(fs.PathSeparator()), testFileName)}

			for _, item := range list {
				fileList = append(fileList, strings.ReplaceAll(item, filepath.Dir(parentDirPath), ""))
			}

			sort.Strings(fileList)
			sort.Strings(expectedList)
			assert.True(t, reflect.DeepEqual(fileList, expectedList))
		})
	}
}

func TestListDirTreeWithExclusion(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			testDir, err := fs.TempDirInTempDir("test-list-tree-with-exclusion")
			require.NoError(t, err)
			defer func() { _ = fs.Rm(testDir) }()

			empty, err := fs.IsEmpty(testDir)
			require.NoError(t, err)
			assert.True(t, empty)

			parentDirPath, err := fs.TempDir(testDir, "parentDir-")
			require.NoError(t, err)

			testFile, err := fs.TempFile(testDir, "test-file-*.test")
			require.NoError(t, err)
			err = testFile.Close()
			require.NoError(t, err)

			childDirPath, err := fs.TempDir(parentDirPath, "childDir-")
			require.NoError(t, err)

			_, err = fs.TempDir(childDirPath, "gcDir-")
			require.NoError(t, err)
			_, err = fs.TempDir(childDirPath, "gcDir-1234-")
			require.NoError(t, err)

			checkNotEmpty(t, fs, testDir)

			var list []string
			err = fs.ListDirTree(testDir, &list)
			assert.Len(t, list, 5)
			require.NoError(t, err)
			list = nil
			err = fs.ListDirTreeWithContextAndExclusionPatterns(context.TODO(), testDir, &list, ".*[.]test", "gcDir-.*")
			assert.Len(t, list, 2)
			require.NoError(t, err)

			parentDir := filepath.Base(parentDirPath)
			childDir := filepath.Base(childDirPath)

			var fileList []string
			var expectedList = []string{
				filepath.Join(string(fs.PathSeparator()), parentDir),
				filepath.Join(string(fs.PathSeparator()), parentDir, childDir),
			}

			for _, item := range list {
				fileList = append(fileList, strings.ReplaceAll(item, filepath.Dir(parentDirPath), ""))
			}

			sort.Strings(fileList)
			sort.Strings(expectedList)
			assert.True(t, reflect.DeepEqual(fileList, expectedList))
		})
	}
}

func checkCopyDir(t *testing.T, fs FS, src string, dest string, exclusionPattern ...string) {
	assert.True(t, fs.Exists(src))
	assert.False(t, fs.Exists(dest))
	var err error
	if reflection.IsEmpty(exclusionPattern) {
		err = fs.Copy(src, dest)
	} else {
		err = fs.CopyWithContextAndExclusionPatterns(context.TODO(), src, dest, exclusionPattern...)
	}
	require.NoError(t, err)

	defer func() { _ = fs.Rm(dest) }()
	assert.True(t, fs.Exists(src))
	assert.True(t, fs.Exists(dest))

	var srcFiles []string
	var destFiles []string

	err = fs.ListDirTreeWithContextAndExclusionPatterns(context.TODO(), src, &srcFiles, exclusionPattern...)
	require.NoError(t, err)

	destPath := filepath.Join(dest, filepath.Base(src))
	err = fs.ListDirTreeWithContextAndExclusionPatterns(context.TODO(), destPath, &destFiles, exclusionPattern...)
	require.NoError(t, err)

	var srcContent []string
	var destContent []string

	for _, item := range srcFiles {
		srcContent = append(srcContent, strings.ReplaceAll(item, filepath.Dir(src), ""))
	}
	for _, item := range destFiles {
		destContent = append(destContent, strings.ReplaceAll(item, filepath.Dir(destPath), ""))
	}

	sort.Strings(srcContent)
	sort.Strings(destContent)
	assert.True(t, reflect.DeepEqual(srcContent, destContent))
}

func checkCopy(t *testing.T, fs FS, oldFile string, dest string, exclusionPattern ...string) {

	assert.True(t, fs.Exists(oldFile))
	assert.False(t, fs.Exists(dest))

	empty, err := fs.IsEmpty(oldFile)
	require.NoError(t, err)

	if reflection.IsEmpty(exclusionPattern) {
		err = fs.Copy(oldFile, dest)
	} else {
		err = fs.CopyWithContextAndExclusionPatterns(context.TODO(), oldFile, dest, exclusionPattern...)
	}
	require.NoError(t, err)
	defer func() { _ = fs.Rm(dest) }()

	assert.True(t, fs.Exists(oldFile))
	if reflection.IsEmpty(exclusionPattern) {
		assert.True(t, fs.Exists(dest))

		empty2, err := fs.IsEmpty(filepath.Join(dest, filepath.Base(oldFile)))
		require.NoError(t, err)
		assert.Equal(t, empty, empty2)
	} else {
		if IsPathExcludedFromPatterns(dest, fs.PathSeparator(), exclusionPattern...) || IsPathExcludedFromPatterns(oldFile, fs.PathSeparator(), exclusionPattern...) {
			assert.False(t, fs.Exists(dest))
		} else {
			assert.True(t, fs.Exists(dest))
		}
	}
	require.NoError(t, fs.Rm(dest))
}

func checkMove(t *testing.T, fs FS, oldFile string, newFile string) {

	assert.True(t, fs.Exists(oldFile))
	assert.False(t, fs.Exists(newFile))

	empty, err := fs.IsEmpty(oldFile)
	require.NoError(t, err)

	err = fs.Move(oldFile, newFile)
	require.NoError(t, err)
	defer func() { _ = fs.Rm(newFile) }()
	assert.False(t, fs.Exists(oldFile))
	assert.True(t, fs.Exists(newFile))

	empty2, err := fs.IsEmpty(newFile)
	require.NoError(t, err)
	assert.Equal(t, empty, empty2)
}

func checkNotEmpty(t *testing.T, fs FS, tmpDir string) {
	empty, err := fs.IsEmpty(tmpDir)
	require.NoError(t, err)
	assert.False(t, empty)
}

func testFileMode(t *testing.T, fs FS, filePath string, mode int) {
	fi, err := fs.Lstat(filePath)
	require.NoError(t, err)
	if platform.IsWindows() {
		// Only user's permissions matter. Execution rights are not really considered.
		userExpectedPermission := mode & 0600
		userActualPermission := int(fi.Mode().Perm()) & 0600
		assert.Equal(t, userExpectedPermission, userActualPermission)
	} else {
		assert.Equal(t, mode, int(fi.Mode().Perm()))
	}
}
