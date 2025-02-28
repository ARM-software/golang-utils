/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package filesystem

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/idgen"
	"github.com/ARM-software/golang-utils/utils/platform"
	"github.com/ARM-software/golang-utils/utils/reflection"
	"github.com/ARM-software/golang-utils/utils/safeio"
	sizeUnits "github.com/ARM-software/golang-utils/utils/units/size"
)

func TestCurrentDirectory(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			current, err := fs.CurrentDirectory()
			require.NoError(t, err)
			assert.NotEmpty(t, current)
		})
	}
}

func TestExists(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			tmpFile, err := fs.TouchTempFileInTempDir("test-exist-")
			require.NoError(t, err)

			defer func() { _ = fs.Rm(tmpFile) }()
			assert.True(t, fs.Exists(tmpFile))

			randomFile, err := idgen.GenerateUUID4()
			require.NoError(t, err)
			assert.False(t, fs.Exists(randomFile))
			_ = fs.Rm(tmpFile)
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
			file, err := fs.OpenFile(filePath, os.O_RDWR|os.O_CREATE, os.FileMode(mode)) //nolint:gosec // this is a test and mode is 0600
			require.NoError(t, err)
			defer func() { _ = file.Close() }()

			_, err = safeio.WriteString(context.TODO(), file, "initial")
			require.NoError(t, err)
			err = file.Close()
			require.NoError(t, err)

			assert.True(t, fs.Exists(filePath))

			testFileMode(t, fs, filePath, mode)

			ignoredMode := 0322
			file, err = fs.OpenFile(filePath, os.O_WRONLY|os.O_APPEND, os.FileMode(ignoredMode)) //nolint:gosec // this is a test and mode is 0322
			require.NoError(t, err)
			defer func() { _ = file.Close() }()

			_, err = safeio.WriteString(context.TODO(), file, "|append")
			require.NoError(t, err)
			err = file.Close()
			require.NoError(t, err)
			testFileMode(t, fs, filePath, mode)

			ignoredMode = 0400
			file, err = fs.OpenFile(filePath, os.O_RDONLY, os.FileMode(ignoredMode)) //nolint:gosec // this is a test and mode is 0400
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
			file, err = fs.OpenFile(filePath, os.O_RDWR|os.O_TRUNC, os.FileMode(ignoredMode)) //nolint:gosec // this is a test and mode is 0664
			require.NoError(t, err)
			defer func() { _ = file.Close() }()

			testFileMode(t, fs, filePath, mode)

			contents, err = safeio.ReadAll(context.TODO(), file)
			require.Error(t, err)
			errortest.AssertError(t, err, commonerrors.ErrEmpty)
			err = file.Close()
			require.NoError(t, err)
			assert.Equal(t, "", string(contents))
			_ = fs.Rm(filePath)
			_ = fs.Rm(tmpDir)
		})
	}
}

func TestTouch(t *testing.T) {
	// Similar to https://github.com/m-spitfire/go-touch/blob/master/touch_test.go
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			tmpDir, err := fs.TempDirInTempDir("test-touch-")
			require.NoError(t, err)
			defer func() { _ = fs.Rm(tmpDir) }()
			tmpFile, err := fs.TouchTempFile(tmpDir, "test-touch*.test")
			require.NoError(t, err)
			assert.True(t, fs.Exists(tmpFile))
			empty, _ := fs.IsEmpty(tmpFile)
			assert.True(t, empty)
			timesFile1, err := fs.StatTimes(tmpFile)
			require.NoError(t, err)

			tmpFile2 := filepath.Join(tmpDir, "testtouch2.test")
			err = fs.Touch(tmpFile2)
			require.NoError(t, err)
			empty, err = fs.IsEmpty(tmpFile2)
			require.NoError(t, err)
			assert.True(t, empty)
			timesFile2, err := fs.StatTimes(tmpFile2)
			require.NoError(t, err)

			err = fs.Touch(tmpFile)
			require.NoError(t, err)
			err = fs.Touch(tmpFile2)
			require.NoError(t, err)

			timesFile12, err := fs.StatTimes(tmpFile)
			require.NoError(t, err)
			timesFile22, err := fs.StatTimes(tmpFile2)
			require.NoError(t, err)

			assert.NotEqual(t, timesFile12.ModTime(), timesFile1.ModTime())
			assert.NotEqual(t, timesFile2.ModTime(), timesFile22.ModTime())

			t.Run("empty path", func(t *testing.T) {
				err = fs.Touch("                 ")
				require.Error(t, err)
				errortest.AssertError(t, err, commonerrors.ErrUndefined)

			})
			t.Run("directory path", func(t *testing.T) {
				tmpDir2 := filepath.Join(tmpDir, "testfolder") + string(fs.PathSeparator())
				err = fs.Touch(tmpDir2)
				require.NoError(t, err)
				empty, err = fs.IsEmpty(tmpDir2)
				require.NoError(t, err)
				assert.True(t, empty)
				isDir, err := fs.IsDir(tmpDir2)
				require.NoError(t, err)
				assert.True(t, isDir)
			})
			t.Run("permissions-temp", func(t *testing.T) {
				tmpOpenFile, err := fs.TouchTempFile(tmpDir, "test-touch-perm*.test")
				require.NoError(t, err)
				assert.True(t, fs.Exists(tmpOpenFile))
				f, err := fs.GenericOpen(tmpOpenFile)
				require.NoError(t, err)
				require.NoError(t, f.Close())
				_, err = fs.ReadFile(tmpOpenFile)
				errortest.AssertError(t, err, nil, commonerrors.ErrEmpty)
				err = fs.WriteFile(tmpOpenFile, []byte(faker.Sentence()), 0644)
				require.NoError(t, err)
			})
			t.Run("permissions", func(t *testing.T) {
				tmpOpenFile := filepath.Join(tmpDir, "test-touch-perm-direct.test")
				err := fs.Touch(tmpOpenFile)
				require.NoError(t, err)
				assert.True(t, fs.Exists(tmpOpenFile))
				f, err := fs.GenericOpen(tmpOpenFile)
				require.NoError(t, err)
				require.NoError(t, f.Close())
				_, err = fs.ReadFile(tmpOpenFile)
				errortest.AssertError(t, err, nil, commonerrors.ErrEmpty)
				err = fs.WriteFile(tmpOpenFile, []byte(faker.Sentence()), 0644)
				require.NoError(t, err)
			})
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
			errortest.AssertError(t, err, commonerrors.ErrCancelled)
			if fs.Exists(filePath) {
				empty, err := fs.IsEmpty(filePath)
				require.NoError(t, err)

				bytes, err := fs.ReadFile(filePath)
				if empty {
					require.Error(t, err)
					errortest.AssertError(t, err, commonerrors.ErrEmpty)
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
			errortest.AssertError(t, err, commonerrors.ErrCancelled)
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
			tmpFile, err := fs.TouchTempFile(tmpDir, "test-*.txt")
			require.NoError(t, err)

			require.True(t, fs.Exists(tmpFile))
			for _, mode := range []int{0666, 0777, 0555, 0766, 0444, 0644} {
				t.Run(fmt.Sprintf("%v", mode), func(t *testing.T) {
					err = fs.Chmod(tmpFile, os.FileMode(mode)) //nolint:gosec // this is a test and mode can be seen above to abide by the conversion rules
					if err != nil {
						_ = fs.Chmod(tmpFile, os.FileMode(mode)) //nolint:gosec // this is a test and mode can be seen above to abide by the conversion rules
					}
					require.NoError(t, err)
					testFileMode(t, fs, tmpFile, mode)
				})
			}
			_ = fs.Rm(tmpDir)
		})
	}
}

func TestChmodRecursive(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			tmpDir, err := fs.TempDirInTempDir("test-chmod-recursive-")
			require.NoError(t, err)
			defer func() { _ = fs.Rm(tmpDir) }()
			tmpFile, err := fs.TouchTempFile(tmpDir, "test-*.txt")
			require.NoError(t, err)

			tmpDir2, err := fs.TempDir(tmpDir, "test-chmod-recursive-")
			require.NoError(t, err)

			tmpFile2, err := fs.TouchTempFile(tmpDir2, "test-*.txt")
			require.NoError(t, err)

			require.True(t, fs.Exists(tmpFile))
			for _, mode := range []int{0666, 0777, 0555, 0766, 0444, 0644} {
				t.Run(fmt.Sprintf("%v", mode), func(t *testing.T) {
					err = fs.ChmodRecursively(context.TODO(), tmpFile2, os.FileMode(mode)) //nolint:gosec // this is a test and mode can be seen above to abide by the conversion rules
					if err == nil {
						require.NoError(t, err)
						testFileMode(t, fs, tmpFile2, mode)

						err = fs.ChmodRecursively(context.TODO(), tmpDir, os.FileMode(mode)) //nolint:gosec // this is a test and mode can be seen above to abide by the conversion rules
						if err == nil {
							require.NoError(t, err)
							testFileMode(t, fs, tmpFile, mode)
							testFileMode(t, fs, tmpFile2, mode)
						} else {
							errortest.AssertErrorDescription(t, err, "permission denied")
						}

					} else {
						errortest.AssertErrorDescription(t, err, "permission denied")
					}

				})

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

			tmpFile, err := fs.TouchTempFile(tmpDir, "test-*.txt")
			require.NoError(t, err)

			require.True(t, fs.Exists(tmpFile))
			uID, gID, err := fs.FetchOwners(tmpFile)
			if err != nil {
				errortest.AssertError(t, err, commonerrors.ErrNotImplemented, commonerrors.ErrUnsupported)
			} else {
				err = fs.Chown(tmpFile, uID, gID)
				if err != nil {
					errortest.AssertError(t, err, commonerrors.ErrNotImplemented, commonerrors.ErrUnsupported)
				} else {
					newUID, newGID, err := fs.FetchOwners(tmpFile)
					require.NoError(t, err)
					assert.Equal(t, uID, newUID)
					assert.Equal(t, gID, newGID)
				}
			}
			_ = fs.Rm(tmpDir)
		})
	}
}

func TestChangeOwnership(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			tmpDir, err := fs.TempDirInTempDir("test-chown-changeownership-")
			require.NoError(t, err)
			defer func() { _ = fs.Rm(tmpDir) }()

			tmpFile, err := fs.TouchTempFile(tmpDir, "test-*.txt")
			require.NoError(t, err)
			require.True(t, fs.Exists(tmpFile))
			owner, err := fs.FetchFileOwner(tmpFile)
			if err != nil {
				errortest.AssertError(t, err, commonerrors.ErrNotImplemented, commonerrors.ErrUnsupported, commonerrors.ErrNotFound)
			} else {
				require.NotNil(t, owner)
				err = fs.ChangeOwnership(tmpFile, owner)
				if err != nil {
					errortest.AssertError(t, err, commonerrors.ErrNotImplemented, commonerrors.ErrUnsupported)
				} else {
					newUID, newGID, err := fs.FetchOwners(tmpFile)
					require.NoError(t, err)
					assert.Equal(t, owner.Uid, strconv.Itoa(newUID))
					assert.Equal(t, owner.Gid, strconv.Itoa(newGID))
				}
			}
			_ = fs.Rm(tmpDir)
		})
	}
}

func TestChangeOwnershipRecursive(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			tmpDir, err := fs.TempDirInTempDir("test-chown-changeownership-")
			require.NoError(t, err)
			defer func() { _ = fs.Rm(tmpDir) }()

			tmpDir2, err := fs.TempDir(tmpDir, "test-chown-changeownership-")
			require.NoError(t, err)

			tmpFile, err := fs.TouchTempFile(tmpDir, "test-*.txt")
			require.NoError(t, err)
			require.True(t, fs.Exists(tmpFile))

			tmpFile2, err := fs.TouchTempFile(tmpDir2, "test-*.txt")
			require.NoError(t, err)
			require.True(t, fs.Exists(tmpFile))

			owner, err := fs.FetchFileOwner(tmpFile)
			if err != nil {
				errortest.AssertError(t, err, commonerrors.ErrNotImplemented, commonerrors.ErrUnsupported, commonerrors.ErrNotFound)
			} else {
				require.NotNil(t, owner)
				err = fs.ChangeOwnershipRecursively(context.TODO(), tmpDir, owner)
				if err != nil {
					errortest.AssertError(t, err, commonerrors.ErrNotImplemented, commonerrors.ErrUnsupported)
				} else {
					newUID, newGID, err := fs.FetchOwners(tmpFile)
					require.NoError(t, err)
					assert.Equal(t, owner.Uid, strconv.Itoa(newUID))
					assert.Equal(t, owner.Gid, strconv.Itoa(newGID))
					newUID, newGID, err = fs.FetchOwners(tmpFile2)
					require.NoError(t, err)
					assert.Equal(t, owner.Uid, strconv.Itoa(newUID))
					assert.Equal(t, owner.Gid, strconv.Itoa(newGID))
				}
			}
			t.Run("empty user", func(t *testing.T) {
				err = fs.ChangeOwnershipRecursively(context.TODO(), tmpFile, nil)
				require.Error(t, err)
				errortest.AssertError(t, err, commonerrors.ErrUndefined)
			})
			t.Run("cancelled context", func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.TODO())
				cancel()
				err = fs.ChangeOwnershipRecursively(ctx, tmpFile, nil)
				require.Error(t, err)
				errortest.AssertError(t, err, commonerrors.ErrCancelled, commonerrors.ErrTimeout)
			})
			_ = fs.Rm(tmpDir)
		})
	}
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
			tree := GenerateTestFileTree(t, fs, tmpDir, "", false, time.Now(), time.Now())
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

			txt := fmt.Sprintf("This is a test sentence!!! %v", faker.Sentence())
			tmpFile, err := fs.TouchTempFile(tmpDir, "test-*.txt")
			require.NoError(t, err)
			err = fs.WriteFile(tmpFile, []byte(txt), 0755)
			require.NoError(t, err)

			symlink := filepath.Join(tmpDir, "symlink-tofile")
			hardlink := filepath.Join(tmpDir, "hardlink-tofile")

			err = fs.Symlink(tmpFile, symlink)
			if errors.Is(err, commonerrors.ErrNotImplemented) || errors.Is(err, afero.ErrNoSymlink) {
				return
			}
			require.NoError(t, err)

			err = fs.Link(tmpFile, hardlink)
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
			assert.Equal(t, tmpFile, link)

			link, err = fs.Readlink(hardlink)
			require.Error(t, err)
			assert.Empty(t, link)

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
			tmpFile, err := fs.TouchTempFileInTempDir("test-file-times-*.txt")
			require.NoError(t, err)
			defer func() { _ = fs.Rm(tmpFile) }()

			assert.True(t, fs.Exists(tmpFile))
			fileTimes, err := fs.StatTimes(tmpFile)
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
			_ = fs.Rm(tmpFile)
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

			tmpFile, err := fs.TouchTempFile(tmpDir, "test-isempty-*.test")
			require.NoError(t, err)
			defer func() { _ = fs.Rm(tmpFile) }()

			isFileEmpty, err := fs.IsEmpty(tmpFile)
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

			tmpFile, err := fs.TouchTempFile(tmpDir, "test-isempty-*.test")
			require.NoError(t, err)
			defer func() { _ = fs.Rm(tmpFile) }()

			timesDirOrig, err := fs.StatTimes(tmpDir)
			require.NoError(t, err)
			timesFileOrig, err := fs.StatTimes(tmpFile)
			require.NoError(t, err)
			newTimeA := time.Now().Add(-time.Hour)
			newTimeM := time.Now().Add(-30 * time.Minute)
			err = fs.Chtimes(tmpDir, newTimeA, newTimeM)
			require.NoError(t, err)
			err = fs.Chtimes(tmpFile, newTimeA, newTimeM)
			require.NoError(t, err)

			timesDirMod, err := fs.StatTimes(tmpDir)
			require.NoError(t, err)
			timesFileMod, err := fs.StatTimes(tmpFile)
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

			_, err = fs.TouchTempFile(tmpDir, "test-cleandir-*.test")
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

			_, err = fs.TouchTempFile(tmpDir, "test-cleandir-*.test")
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

			_, err = fs.TouchTempFile(tmpDir, "test-rm-*.test")
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
func TestRemoveWithPrivileges(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			tmpDir, err := fs.TempDirInTempDir("test-rm-with-privileges-")
			require.NoError(t, err)
			defer func() { _ = fs.Rm(tmpDir) }()

			empty, err := fs.IsEmpty(tmpDir)
			require.NoError(t, err)
			assert.True(t, empty)

			tmpFile, err := fs.TouchTempFile(tmpDir, "test-rm-*.test")
			require.NoError(t, err)

			dirToRemove1, err := fs.TempDir(tmpDir, "testDirToRemove")
			require.NoError(t, err)

			dirToRemove2, err := fs.TempDir(tmpDir, "testDirToRemove")
			require.NoError(t, err)

			checkNotEmpty(t, fs, tmpDir)

			// TODO: add user and change file and folder ownership

			err = fs.RemoveWithPrivileges(context.TODO(), tmpFile)
			require.NoError(t, err)

			err = fs.RemoveWithPrivileges(context.TODO(), dirToRemove1)
			require.NoError(t, err)

			err = fs.RemoveWithPrivileges(context.TODO(), dirToRemove2)
			require.NoError(t, err)

			assert.True(t, fs.Exists(tmpDir))

			empty, err = fs.IsEmpty(tmpDir)
			require.NoError(t, err)
			assert.True(t, empty)

			require.NoError(t, fs.Rm(tmpDir))
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
			tmpFile, err := fs.TouchTempFile(tmpDir, "test-ls-*.test")
			require.NoError(t, err)

			files, err := fs.Ls(tmpDir)
			require.NoError(t, err)
			assert.Len(t, files, 2)
			_, found := collection.Find(&files, filepath.Base(tmpFile))
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
			tmpFile, err := fs.TouchTempFile(tmpDir, "test-ls-*.test")
			require.NoError(t, err)

			files, err := fs.Ls(tmpDir)
			require.NoError(t, err)
			assert.Len(t, files, 2)

			files, err = fs.LsWithExclusionPatterns(tmpDir, ".*[.]test")
			require.NoError(t, err)
			assert.Len(t, files, 1)

			_, found := collection.Find(&files, filepath.Base(tmpFile))
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
			tree := GenerateTestFileTree(t, fs, testDir, "", false, time.Now(), time.Now())
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
			tree := GenerateTestFileTree(t, fs, testDir, "", false, time.Now(), time.Now())
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
			_, err = fs.TouchTempFile(tmpDir, "test-gc-*.test")
			require.NoError(t, err)

			_, err = fs.TouchTempFile(tmpDir1, "test-gc-*.test")
			require.NoError(t, err)

			ctime := time.Now()
			time.Sleep(500 * time.Millisecond)

			tmpDir3, err := fs.TempDir(tmpDir, "test-gc-")
			require.NoError(t, err)
			tmpFile, err := fs.TouchTempFile(tmpDir3, "test-gc-*.test")
			require.NoError(t, err)
			_, err = fs.TouchTempFile(tmpDir2, "test-gc-*.test")
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
			_, found = collection.Find(&files, filepath.Base(tmpFile))
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

			_, err = fs.TouchTempFile(level1, "test-findall-*.test1")
			require.NoError(t, err)

			_, err = fs.TouchTempFile(level2, "test-findall-*.test2")
			require.NoError(t, err)

			_, err = fs.TouchTempFile(level3, "test-findall-*.test3")
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

func TestTempFile(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			tmpDir, err := fs.TempDirInTempDir("test-copy-to-file-")
			require.NoError(t, err)
			defer func() { _ = fs.Rm(tmpDir) }()

			f, err := fs.TempFile(tmpDir, "test-tmp-file-*.txt")
			require.NoError(t, err)
			defer func() { require.NoError(t, f.Close()) }()
			content := faker.Paragraph()
			p, err := f.Write([]byte(content))
			require.NoError(t, err)
			assert.Equal(t, p, len(content))
			require.NoError(t, f.Close())

			actual, err := fs.ReadFile(f.Name())
			require.NoError(t, err)
			assert.Equal(t, content, string(actual))
		})
	}
}

func TestCopyToFile(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			tmpDir, err := fs.TempDirInTempDir("test-copy-to-file-")
			require.NoError(t, err)
			defer func() { _ = fs.Rm(tmpDir) }()

			empty, err := fs.IsEmpty(tmpDir)
			require.NoError(t, err)
			assert.True(t, empty)

			tmpFile, err := fs.TouchTempFile(tmpDir, "test-copy-to-file-*.test")
			require.NoError(t, err)

			checkNotEmpty(t, fs, tmpDir)

			empty, err = fs.IsEmpty(tmpFile)
			require.NoError(t, err)
			require.True(t, empty, "created file must be empty")

			destinationFiles := []string{faker.Word(), faker.DomainName(), "." + faker.Word(), faker.Word() + "." + faker.Word(), "." + faker.Word() + "." + faker.Word()}
			for i := range destinationFiles {
				destinationFile := destinationFiles[i]
				// This should result to the creation of a new file named after destination and with the content of the source
				t.Run(fmt.Sprintf("copy to file to a non existing file [%v]", destinationFile), func(t *testing.T) {
					dest := filepath.Join(tmpDir, destinationFile)
					assert.False(t, fs.Exists(dest))
					err = fs.CopyToFile(tmpFile, dest)
					require.NoError(t, err)
					assert.True(t, fs.Exists(dest))
					isFile, err := fs.IsFile(dest)
					require.NoError(t, err)
					assert.True(t, isFile)
				})
			}

			for i := range destinationFiles {
				destinationFile := destinationFiles[i]
				// any existing file should be overwritten
				t.Run(fmt.Sprintf("copy to file to an existing file [%v]", destinationFile), func(t *testing.T) {
					dest := filepath.Join(tmpDir, destinationFile)
					err := fs.WriteFile(dest, []byte("this is a test "+faker.Sentence()), os.ModePerm)
					require.NoError(t, err)
					assert.True(t, fs.Exists(dest))
					empty, err = fs.IsEmpty(dest)
					require.NoError(t, err)
					assert.False(t, empty)
					err = fs.CopyToFile(tmpFile, dest)
					require.NoError(t, err)
					assert.True(t, fs.Exists(dest))
					isFile, err := fs.IsFile(dest)
					require.NoError(t, err)
					assert.True(t, isFile)
					empty, err = fs.IsEmpty(dest)
					require.NoError(t, err)
					assert.True(t, empty, "existing file must be overwritten during copy")
					require.NoError(t, fs.Rm(dest))
				})
			}
			destinationFolders := []string{faker.DomainName(), faker.DomainName(), "." + faker.Word(), faker.Word() + "." + faker.Word(), "." + faker.Word() + "." + faker.Word()}
			pathSeparators := []string{"/", string(fs.PathSeparator())}
			for i := range destinationFolders {
				for j := range pathSeparators {
					sep := pathSeparators[j]
					destinationFolder := destinationFolders[i] + sep
					// if the destination ends with a path separator then it must be understood as a folder
					t.Run(fmt.Sprintf("copy to file to a non existing folder [%v] should fail", destinationFolder), func(t *testing.T) {
						dest := filepath.Join(tmpDir, destinationFolder) + sep
						require.NoError(t, fs.Rm(dest))
						assert.False(t, fs.Exists(dest))
						err = fs.CopyToFile(tmpFile, dest)
						require.Error(t, err)
						assert.False(t, fs.Exists(dest))
					})
				}
			}

			destinationFolders = []string{faker.DomainName(), faker.DomainName() + string(fs.PathSeparator()), "." + faker.Word(), faker.Word() + "." + faker.Word(), "." + faker.Word() + "." + faker.Word()}
			for i := range destinationFolders {
				destinationFolder := destinationFolders[i]
				t.Run(fmt.Sprintf("copy to file to an existing folder [%v] should fail", destinationFolder), func(t *testing.T) {
					dest := filepath.Join(tmpDir, destinationFolder)
					err = fs.MkDir(dest)
					assert.True(t, fs.Exists(dest))
					isDir, err := fs.IsDir(dest)
					require.NoError(t, err)
					assert.True(t, isDir)
					err = fs.CopyToFile(tmpFile, dest)
					require.Error(t, err)
				})
			}
			t.Run(fmt.Sprintf("copy to file a folder [%v] should fail", tmpDir), func(t *testing.T) {
				err = fs.CopyToFile(tmpDir, faker.Word())
				require.Error(t, err)
				errortest.AssertError(t, err, commonerrors.ErrInvalid)
			})

			_ = fs.Rm(tmpDir)
		})
	}
}

func TestCopyToDirectory(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			tmpDir, err := fs.TempDirInTempDir("test-copy-to-dir-")
			require.NoError(t, err)
			defer func() { _ = fs.Rm(tmpDir) }()

			empty, err := fs.IsEmpty(tmpDir)
			require.NoError(t, err)
			assert.True(t, empty)

			tmpFile, err := fs.TouchTempFile(tmpDir, "test-copy-*.test")
			require.NoError(t, err)

			checkNotEmpty(t, fs, tmpDir)

			empty, err = fs.IsEmpty(tmpFile)
			require.NoError(t, err)
			require.True(t, empty, "created file must be empty")

			destinationFolders := []string{faker.DomainName(), faker.DomainName(), "." + faker.Word(), faker.Word() + "." + faker.Word(), "." + faker.Word() + "." + faker.Word()}
			pathSeparators := []string{"/", string(fs.PathSeparator())}
			for i := range destinationFolders {
				for j := range pathSeparators {
					sep := pathSeparators[j]
					destinationFolder := destinationFolders[i] + sep
					// if the destination ends with a path separator then it must be understood as a folder
					t.Run(fmt.Sprintf("copy file to a non existing folder [%v]", destinationFolder), func(t *testing.T) {
						dest := filepath.Join(tmpDir, destinationFolder) + sep
						require.NoError(t, fs.Rm(dest))
						assert.False(t, fs.Exists(dest))
						err = fs.CopyToDirectory(tmpFile, dest)
						require.NoError(t, err)
						assert.True(t, fs.Exists(dest))
						isDir, err := fs.IsDir(dest)
						require.NoError(t, err)
						assert.True(t, isDir)
						destFile := filepath.Join(dest, filepath.Base(tmpFile))
						assert.True(t, fs.Exists(destFile))
						isFile, err := fs.IsFile(destFile)
						require.NoError(t, err)
						assert.True(t, isFile)
					})
				}
			}
			// if the destination is an existing folder, then the file must be copied to a new file in the destination folder named after the source file.
			destinationFolders = []string{faker.DomainName(), faker.DomainName() + string(fs.PathSeparator()), "." + faker.Word(), faker.Word() + "." + faker.Word(), "." + faker.Word() + "." + faker.Word()}
			for i := range destinationFolders {
				destinationFolder := destinationFolders[i]
				t.Run(fmt.Sprintf("copy file to an existing folder [%v]", destinationFolder), func(t *testing.T) {
					dest := filepath.Join(tmpDir, destinationFolder)
					err = fs.MkDir(dest)
					assert.True(t, fs.Exists(dest))
					isDir, err := fs.IsDir(dest)
					require.NoError(t, err)
					assert.True(t, isDir)
					err = fs.CopyToDirectory(tmpFile, dest)
					require.NoError(t, err)
					destFile := filepath.Join(dest, filepath.Base(tmpFile))
					assert.True(t, fs.Exists(destFile))
					isFile, err := fs.IsFile(destFile)
					require.NoError(t, err)
					assert.True(t, isFile)
				})
			}

			_ = fs.Rm(tmpDir)
		})
	}
}

// TestCopyFile verifies `Copy` has a similar behaviour to `cp -r` when copying files
// e.g. filesystem.Copy("source.txt","does-not-exist") and filesystem.Copy("source.txt",".does-not-exist") both create files with the content of source.txt
func TestCopyFile(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			tmpDir, err := fs.TempDirInTempDir("test-copy-file-")
			require.NoError(t, err)
			defer func() { _ = fs.Rm(tmpDir) }()

			empty, err := fs.IsEmpty(tmpDir)
			require.NoError(t, err)
			assert.True(t, empty)

			tmpFile, err := fs.TouchTempFile(tmpDir, "test-copy-*.test")
			require.NoError(t, err)

			checkNotEmpty(t, fs, tmpDir)

			empty, err = fs.IsEmpty(tmpFile)
			require.NoError(t, err)
			require.True(t, empty, "created file must be empty")

			destinationFiles := []string{faker.Word(), faker.DomainName(), "." + faker.Word(), faker.Word() + "." + faker.Word(), "." + faker.Word() + "." + faker.Word()}
			for i := range destinationFiles {
				destinationFile := destinationFiles[i]
				// This should result to the creation of a new file named after destination and with the content of the source
				t.Run(fmt.Sprintf("copy file to a non existing file [%v]", destinationFile), func(t *testing.T) {
					dest := filepath.Join(tmpDir, destinationFile)
					assert.False(t, fs.Exists(dest))
					err = fs.Copy(tmpFile, dest)
					require.NoError(t, err)
					assert.True(t, fs.Exists(dest))
					isFile, err := fs.IsFile(dest)
					require.NoError(t, err)
					assert.True(t, isFile)
				})
			}

			for i := range destinationFiles {
				destinationFile := destinationFiles[i]
				// any existing file should be overwritten
				t.Run(fmt.Sprintf("copy file to an existing file [%v]", destinationFile), func(t *testing.T) {
					dest := filepath.Join(tmpDir, destinationFile)
					err := fs.WriteFile(dest, []byte("this is a test "+faker.Sentence()), os.ModePerm)
					require.NoError(t, err)
					assert.True(t, fs.Exists(dest))
					empty, err = fs.IsEmpty(dest)
					require.NoError(t, err)
					assert.False(t, empty)
					err = fs.Copy(tmpFile, dest)
					require.NoError(t, err)
					assert.True(t, fs.Exists(dest))
					isFile, err := fs.IsFile(dest)
					require.NoError(t, err)
					assert.True(t, isFile)
					empty, err = fs.IsEmpty(dest)
					require.NoError(t, err)
					assert.True(t, empty, "existing file must be overwritten during copy")
					require.NoError(t, fs.Rm(dest))
				})
			}
			destinationFolders := []string{faker.DomainName(), faker.DomainName(), "." + faker.Word(), faker.Word() + "." + faker.Word(), "." + faker.Word() + "." + faker.Word()}
			pathSeparators := []string{"/", string(fs.PathSeparator())}
			for i := range destinationFolders {
				for j := range pathSeparators {
					sep := pathSeparators[j]
					destinationFolder := destinationFolders[i] + sep
					// if the destination ends with a path separator then it must be understood as a folder
					t.Run(fmt.Sprintf("copy file to a non existing folder [%v]", destinationFolder), func(t *testing.T) {
						dest := filepath.Join(tmpDir, destinationFolder) + sep
						require.NoError(t, fs.Rm(dest))
						assert.False(t, fs.Exists(dest))
						err = fs.Copy(tmpFile, dest)
						require.NoError(t, err)
						assert.True(t, fs.Exists(dest))
						isDir, err := fs.IsDir(dest)
						require.NoError(t, err)
						assert.True(t, isDir)
						destFile := filepath.Join(dest, filepath.Base(tmpFile))
						assert.True(t, fs.Exists(destFile))
						isFile, err := fs.IsFile(destFile)
						require.NoError(t, err)
						assert.True(t, isFile)
					})
				}
			}
			// if the destination is an existing folder, then the file must be copied to a new file in the destination folder named after the source file.
			destinationFolders = []string{faker.DomainName(), faker.DomainName() + string(fs.PathSeparator()), "." + faker.Word(), faker.Word() + "." + faker.Word(), "." + faker.Word() + "." + faker.Word()}
			for i := range destinationFolders {
				destinationFolder := destinationFolders[i]
				t.Run(fmt.Sprintf("copy file to an existing folder [%v]", destinationFolder), func(t *testing.T) {
					dest := filepath.Join(tmpDir, destinationFolder)
					err = fs.MkDir(dest)
					assert.True(t, fs.Exists(dest))
					isDir, err := fs.IsDir(dest)
					require.NoError(t, err)
					assert.True(t, isDir)
					err = fs.Copy(tmpFile, dest)
					require.NoError(t, err)
					destFile := filepath.Join(dest, filepath.Base(tmpFile))
					assert.True(t, fs.Exists(destFile))
					isFile, err := fs.IsFile(destFile)
					require.NoError(t, err)
					assert.True(t, isFile)
				})
			}

			_ = fs.Rm(tmpDir)
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

			tmpFile, err := fs.TouchTempFile(tmpDir, "test-copy-*.test")
			require.NoError(t, err)

			checkNotEmpty(t, fs, tmpDir)
			dests := []string{
				filepath.Join(tmpDir, "does-not-exist"),
				filepath.Join(tmpDir, "does-not-exist") + string(fs.PathSeparator()),
				filepath.Join(tmpDir, ".does-not-exist") + string(fs.PathSeparator()),
				filepath.Join(tmpDir, "does-not-exist") + "/",
				filepath.Join(tmpDir, ".does-not-exist") + "/",
				filepath.Join(tmpDir, ".does-not-exist"),
				filepath.Join(tmpDir, "does-not-exist.file"),
				filepath.Join(tmpDir, ".does-not-exist.file"),
				filepath.Join(tmpDir, "does-not-exist", "does-not-exist"),
				filepath.Join(tmpDir, "does-not-exist", ".does-not-exist"),
				filepath.Join(tmpDir, "does-not-exist", "does-not-exist.file"),
				filepath.Join(tmpDir, "does-not-exist", ".does-not-exist.file"),
			}
			for i := range dests {
				dest := dests[i]
				t.Run(dest, func(t *testing.T) {
					checkCopy(t, fs, tmpFile, dest)
				})
			}
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

			tmpFile, err := fs.TouchTempFile(tmpDir, "test-copy-*.test")
			require.NoError(t, err)

			checkNotEmpty(t, fs, tmpDir)

			dests := []string{
				filepath.Join(tmpDir, "test-copy-with-exclusion-test"),
				filepath.Join(tmpDir, "does-not-exist", "test-copy-with-exclusion-test"),
				filepath.Join(tmpDir, ".does-not-exist", "test-copy-with-exclusion-test"),
				filepath.Join(tmpDir, "does-not-exist.file", "test-copy-with-exclusion-test"),
				filepath.Join(tmpDir, ".does-not-exist.file", "test-copy-with-exclusion-test"),
			}
			for i := range dests {
				dest := dests[i]
				t.Run(dest, func(t *testing.T) {
					checkCopy(t, fs, tmpFile, dest, "test-copy-with-exclusion-.*")
				})
			}

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
			checkCopyDir(t, fs, parentDir, filepath.Join(testDir, "does-not-exist"))
			checkCopyDir(t, fs, parentDir, filepath.Join(testDir, ".does-not-exist"))
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

func TestCopyTreeWithExistingDestination(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			tmpDir, err := fs.TempDirInTempDir("test-copy-tree-with-existing-destination-")
			require.NoError(t, err)
			defer func() {
				_ = fs.Rm(tmpDir)
			}()
			srcFolderName := "src"
			src := filepath.Join(tmpDir, srcFolderName)
			dest := filepath.Join(tmpDir, "dest")
			err = fs.MkDir(dest)
			require.NoError(t, err)

			srcFileList := GenerateTestFileTree(t, fs, src, "", false, time.Now(), time.Now())
			srcFileList, err = fs.ConvertToRelativePath(tmpDir, srcFileList...)
			require.NoError(t, err)

			err = fs.Copy(src, dest)
			require.NoError(t, err)

			var destFileList []string
			err = fs.ListDirTree(dest, &destFileList)
			require.NoError(t, err)
			destFileList, err = fs.ConvertToRelativePath(dest, destFileList...)
			require.NoError(t, err)

			// the srcFolderName is present in the destination
			srcFileList = append(srcFileList, srcFolderName)

			assert.ElementsMatch(t, srcFileList, destFileList, "all items should have been copied and under the srcFolderName path")
		})
	}
}

func TestCopyTreeWithMissingDestination(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			tmpDir, err := fs.TempDirInTempDir("test-copy-tree-with-missing-destination-")
			require.NoError(t, err)
			defer func() {
				_ = fs.Rm(tmpDir)
			}()
			srcFolderName := "src"
			src := filepath.Join(tmpDir, srcFolderName)
			dest := filepath.Join(tmpDir, "dest")

			srcFileList := GenerateTestFileTree(t, fs, src, "", false, time.Now(), time.Now())
			srcFileList, err = fs.ConvertToRelativePath(src, srcFileList...)
			require.NoError(t, err)

			err = fs.Copy(src, dest)
			require.NoError(t, err)

			var destFileList []string
			err = fs.ListDirTree(dest, &destFileList)
			require.NoError(t, err)
			destFileList, err = fs.ConvertToRelativePath(dest, destFileList...)
			require.NoError(t, err)

			// the src folder must not be present in the destination
			assert.ElementsMatch(t, srcFileList, destFileList, "all items should have been copied right under the destination root. srcFolderName should not be present in destination")
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

			tmpFile, err := fs.TouchTempFile(tmpDir, "test-move-*.test")
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
			err = fs.Move(tmpFile, tmpFile)
			require.NoError(t, err)
			err = fs.Move(testDir, testDir)
			require.NoError(t, err)
			checkMove(t, fs, tmpFile, filepath.Join(tmpDir, "test.test"))
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

	byteOut, err := fs.ReadFile(fileName)
	require.NoError(t, err)
	assert.Equal(t, byteOut, byteInput)

	_, err = fs.ReadFile("unknown_file")
	assert.Error(t, err)
}

func TestGetFileSize(t *testing.T) {
	fs := NewFs(InMemoryFS)

	tmpFile, err := fs.TouchTempFileInTempDir("test-filesize-*.test")
	require.NoError(t, err)
	defer func() { _ = fs.Rm(tmpFile) }()

	builder := strings.Builder{}
	for indx := 0; indx < 50; indx++ {
		builder.WriteString(" Here is a Test string....")
	}

	err = fs.WriteFile(tmpFile, []byte(builder.String()), os.FileMode(0777))
	require.NoError(t, err)

	size, err := fs.GetFileSize(tmpFile)
	require.NoError(t, err)
	assert.Equal(t, int64(1.3*sizeUnits.KB), size)

	_, err = fs.GetFileSize("Unknown-File")
	assert.Error(t, err)
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
			_, err = fs.TouchTempFile(testInput, "test-subdir-*.ini")
			require.NoError(t, err)

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
			_, err = fs.TouchTempFile(testInput, "test-subdirectory-*.ini")
			require.NoError(t, err)

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

			tmpFile, err := fs.TouchTempFile(testDir, "test-listdir-*.test")
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
			testFileName := filepath.Base(tmpFile)

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

			_, err = fs.TouchTempFile(testDir, "test-listdir-*.test")
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
	require.NoError(t, fs.MkDir(dest))
	assert.True(t, fs.Exists(dest))
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
		t.Run(fmt.Sprintf("checking compliance with [cp %v %v]", filepath.Base(oldFile), filepath.Base(dest)), func(t *testing.T) {
			isSrcFile, err := fs.IsFile(oldFile)
			require.NoError(t, err)
			isDestFile, err := fs.IsFile(dest)
			require.NoError(t, err)
			if isSrcFile {
				if isDestFile {
					empty2, err := fs.IsEmpty(dest)
					require.NoError(t, err)
					assert.Equal(t, empty, empty2, "content of destination file should be the same as source")
				} else {
					file := filepath.Join(filepath.Clean(dest), filepath.Base(oldFile))
					require.True(t, fs.Exists(file), "destination folder should be created and the source file should be a child")
					empty2, err := fs.IsEmpty(file)
					require.NoError(t, err)
					assert.Equal(t, empty, empty2, "content of destination file should be the same as source")
				}

			} else {
				assert.False(t, isDestFile, "destination should be a directory like the source")
			}
		})
	} else {
		t.Run("checking path exclusions", func(t *testing.T) {
			if IsPathExcludedFromPatterns(dest, fs.PathSeparator(), exclusionPattern...) || IsPathExcludedFromPatterns(oldFile, fs.PathSeparator(), exclusionPattern...) {
				assert.False(t, fs.Exists(dest))
			} else {
				assert.True(t, fs.Exists(dest))
			}
		})
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

func TestConvertToOSFile(t *testing.T) {
	temp := t.TempDir()
	file, err := TempFile(temp, "test-file-convert-.test")
	require.NoError(t, err)
	defer func() { _ = file.Close() }()
	osFile := ConvertToOSFile(file)
	require.NotNil(t, osFile)
	text := fmt.Sprintf("test some file data; %v", faker.Paragraph())
	_, err = safeio.CopyDataWithContext(context.TODO(), bytes.NewReader([]byte(text)), osFile)
	require.NoError(t, err)
	require.NoError(t, osFile.Close())
	require.Error(t, file.Close(), "file must already be closed")
	rFile, err := GenericOpen(file.Name())
	require.NoError(t, err)
	defer func() { _ = rFile.Close() }()
	osFile = ConvertToOSFile(rFile)
	defer func() { _ = osFile.Close() }()
	content, err := io.ReadAll(osFile)
	require.NoError(t, err)
	assert.Equal(t, text, string(content))
	require.NoError(t, osFile.Close())
	require.Error(t, rFile.Close(), "file must already be closed")
}

func TestLsRecursive(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run("Test LsRecursive includes directories with available files and directories", func(t *testing.T) {
			fs := NewFs(fsType)

			rootDir := t.TempDir()
			err := generateTestTree(fs, rootDir)
			require.NoError(t, err)

			result, err := fs.LsRecursive(context.Background(), rootDir, true)
			require.NoError(t, err)

			expectedFiles := []string{
				filepath.Join(rootDir, "dir2", "dir3", "dir4", "test1.txt"),
				filepath.Join(rootDir, "dir2", "dir3", "dir4", "test2.o"),
				filepath.Join(rootDir, "dir2", "dir3", "dir4", "test3.h"),
				filepath.Join(rootDir, "dir2", "dir5", "test4.txt"),
				filepath.Join(rootDir, "dir2", "dir5", "test5.jar"),

				rootDir,
				filepath.Join(rootDir, "dir2"),
				filepath.Join(rootDir, "dir2", "dir3"),
				filepath.Join(rootDir, "dir2", "dir3", "dir4"),
				filepath.Join(rootDir, "dir2", "dir5"),
			}

			assert.ElementsMatch(t, expectedFiles, result)
		})

		t.Run("Test LsRecursive without including directories with available files and directories", func(t *testing.T) {
			fs := NewFs(fsType)

			rootDir := t.TempDir()
			err := generateTestTree(fs, rootDir)
			require.NoError(t, err)

			result, err := fs.LsRecursive(context.Background(), rootDir, false)
			require.NoError(t, err)

			expectedFiles := []string{
				filepath.Join(rootDir, "dir2", "dir3", "dir4", "test1.txt"),
				filepath.Join(rootDir, "dir2", "dir3", "dir4", "test2.o"),
				filepath.Join(rootDir, "dir2", "dir3", "dir4", "test3.h"),
				filepath.Join(rootDir, "dir2", "dir5", "test4.txt"),
				filepath.Join(rootDir, "dir2", "dir5", "test5.jar"),
			}

			assert.ElementsMatch(t, expectedFiles, result)
		})

		t.Run("Test LsRecursive with canceled context", func(t *testing.T) {
			fs := NewFs(fsType)

			rootDir := t.TempDir()
			err := generateTestTree(fs, rootDir)
			require.NoError(t, err)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			result, err := fs.LsRecursive(ctx, rootDir, true)
			errortest.AssertError(t, err, commonerrors.ErrCancelled, commonerrors.ErrTimeout)
			assert.Empty(t, result, "Expected no results when context is canceled")
		})

		t.Run("Test LsRecursive with non-existent directory", func(t *testing.T) {
			fs := NewFs(fsType)

			nonExistentDir := filepath.Join(t.TempDir(), "non_existent_dir")

			result, err := fs.LsRecursive(context.Background(), nonExistentDir, true)
			errortest.AssertError(t, err, os.ErrNotExist, commonerrors.ErrNotFound)
			assert.Empty(t, result, "Expected no results when directory does not exist")
		})
	}
}

func TestLsRecursiveWithExclusionPatterns(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run("Test LsRecursiveWithExtensionPatterns includes directories with exclusion patterns", func(t *testing.T) {
			fs := NewFs(fsType)

			rootDir := t.TempDir()
			err := generateTestTree(fs, rootDir)
			require.NoError(t, err)

			exclusionPatterns := []string{
				globToRegex("*.o"),
				globToRegex("*.jar"),
			}

			result, err := fs.LsRecursiveWithExclusionPatterns(context.Background(), rootDir, true, exclusionPatterns...)
			require.NoError(t, err)
			expectedFiles := []string{
				filepath.Join(rootDir, "dir2", "dir3", "dir4", "test1.txt"),
				filepath.Join(rootDir, "dir2", "dir3", "dir4", "test3.h"),
				filepath.Join(rootDir, "dir2", "dir5", "test4.txt"),

				rootDir,
				filepath.Join(rootDir, "dir2"),
				filepath.Join(rootDir, "dir2", "dir3"),
				filepath.Join(rootDir, "dir2", "dir3", "dir4"),
				filepath.Join(rootDir, "dir2", "dir5"),
			}

			assert.ElementsMatch(t, expectedFiles, result)
		})

		t.Run("Test LsRecursiveWithExtensionPatterns without including directories with exclusion patterns", func(t *testing.T) {
			fs := NewFs(fsType)

			rootDir := t.TempDir()
			err := generateTestTree(fs, rootDir)
			require.NoError(t, err)

			exclusionPatterns := []string{
				globToRegex("*.o"),
				globToRegex("*.jar"),
			}

			result, err := fs.LsRecursiveWithExclusionPatterns(context.Background(), rootDir, false, exclusionPatterns...)
			require.NoError(t, err)
			expectedFiles := []string{
				filepath.Join(rootDir, "dir2", "dir3", "dir4", "test1.txt"),
				filepath.Join(rootDir, "dir2", "dir3", "dir4", "test3.h"),
				filepath.Join(rootDir, "dir2", "dir5", "test4.txt"),
			}

			assert.ElementsMatch(t, expectedFiles, result)
		})

		t.Run("Test LsRecursiveWithExtensionPatterns with canceled context", func(t *testing.T) {
			fs := NewFs(fsType)

			rootDir := t.TempDir()
			err := generateTestTree(fs, rootDir)
			require.NoError(t, err)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			result, err := fs.LsRecursiveWithExclusionPatterns(ctx, rootDir, true)
			errortest.AssertError(t, err, commonerrors.ErrCancelled, commonerrors.ErrTimeout)
			assert.Empty(t, result, "Expected no results when context is canceled")
		})

		t.Run("Test LsRecursiveWithExtensionPatterns with non-existent directory", func(t *testing.T) {
			fs := NewFs(fsType)

			nonExistentDir := filepath.Join(t.TempDir(), "non_existent_dir")

			result, err := fs.LsRecursiveWithExclusionPatterns(context.Background(), nonExistentDir, true)
			errortest.AssertError(t, err, os.ErrNotExist, commonerrors.ErrNotFound)
			assert.Empty(t, result, "Expected no results when directory does not exist")
		})
	}
}

func TestLsRecursiveWithExclusionPatternsAndLimits(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run("Test LsRecursiveWithExtensionPatternsAndLimits includes directories with enough max depth and enough max file count", func(t *testing.T) {
			fs := NewFs(fsType)

			rootDir := t.TempDir()
			err := generateTestTree(fs, rootDir)
			require.NoError(t, err)

			limits := &Limits{MaxDepth: 4, MaxFileCount: 10, Recursive: true}
			results, err := fs.LsRecursiveWithExclusionPatternsAndLimits(context.Background(), rootDir, limits, true)
			require.NoError(t, err)
			expectedFiles := []string{
				filepath.Join(rootDir, "dir2", "dir3", "dir4", "test1.txt"),
				filepath.Join(rootDir, "dir2", "dir3", "dir4", "test2.o"),
				filepath.Join(rootDir, "dir2", "dir3", "dir4", "test3.h"),
				filepath.Join(rootDir, "dir2", "dir5", "test4.txt"),
				filepath.Join(rootDir, "dir2", "dir5", "test5.jar"),

				rootDir,
				filepath.Join(rootDir, "dir2"),
				filepath.Join(rootDir, "dir2", "dir3"),
				filepath.Join(rootDir, "dir2", "dir3", "dir4"),
				filepath.Join(rootDir, "dir2", "dir5"),
			}

			assert.ElementsMatch(t, expectedFiles, results)
		})

		t.Run("Test LsRecursiveWithExtensionPatternsAndLimits without including directories with enough max depth and enough max file count", func(t *testing.T) {
			fs := NewFs(fsType)

			rootDir := t.TempDir()
			err := generateTestTree(fs, rootDir)
			require.NoError(t, err)

			limits := &Limits{MaxDepth: 4, MaxFileCount: 10, Recursive: true}
			results, err := fs.LsRecursiveWithExclusionPatternsAndLimits(context.Background(), rootDir, limits, false)
			require.NoError(t, err)
			expectedFiles := []string{
				filepath.Join(rootDir, "dir2", "dir3", "dir4", "test1.txt"),
				filepath.Join(rootDir, "dir2", "dir3", "dir4", "test2.o"),
				filepath.Join(rootDir, "dir2", "dir3", "dir4", "test3.h"),
				filepath.Join(rootDir, "dir2", "dir5", "test4.txt"),
				filepath.Join(rootDir, "dir2", "dir5", "test5.jar"),
			}

			assert.ElementsMatch(t, expectedFiles, results)
		})

		t.Run("Test LsRecursiveWithExtensionPatternsAndLimits without including directories with enough max depth and enough max file count and exclusion patterns", func(t *testing.T) {
			fs := NewFs(fsType)

			rootDir := t.TempDir()
			err := generateTestTree(fs, rootDir)
			require.NoError(t, err)

			exclusionPatterns := []string{
				globToRegex("*.o"),
				globToRegex("*.jar"),
			}

			limits := &Limits{MaxDepth: 4, MaxFileCount: 5, Recursive: true}
			results, err := fs.LsRecursiveWithExclusionPatternsAndLimits(context.Background(), rootDir, limits, false, exclusionPatterns...)
			require.NoError(t, err)
			expectedFiles := []string{
				filepath.Join(rootDir, "dir2", "dir3", "dir4", "test1.txt"),
				filepath.Join(rootDir, "dir2", "dir3", "dir4", "test3.h"),
				filepath.Join(rootDir, "dir2", "dir5", "test4.txt"),
			}

			assert.ElementsMatch(t, expectedFiles, results)
		})

		t.Run("Test LsRecursiveWithExtensionPatternsAndLimits without including directories with NOT enough max depth and enough max file count", func(t *testing.T) {
			fs := NewFs(fsType)

			rootDir := t.TempDir()
			err := generateTestTree(fs, rootDir)
			require.NoError(t, err)

			limits := &Limits{MaxDepth: 3, MaxFileCount: 5, Recursive: true}
			results, err := fs.LsRecursiveWithExclusionPatternsAndLimits(context.Background(), rootDir, limits, false)
			expectedFiles := []string{
				filepath.Join(rootDir, "dir2", "dir5", "test4.txt"),
				filepath.Join(rootDir, "dir2", "dir5", "test5.jar"),
			}

			assert.ElementsMatch(t, expectedFiles, results)
			assert.NoError(t, err)
		})

		t.Run("Test LsRecursiveWithExtensionPatternsAndLimits without including directories with enough max depth and NOT enough max file count", func(t *testing.T) {
			fs := NewFs(fsType)

			rootDir := t.TempDir()
			err := generateTestTree(fs, rootDir)
			require.NoError(t, err)

			limits := &Limits{MaxDepth: 4, MaxFileCount: 2, Recursive: true}
			results, err := fs.LsRecursiveWithExclusionPatternsAndLimits(context.Background(), rootDir, limits, false)
			expectedFilesSize := 2

			assert.Equal(t, expectedFilesSize, len(results))
			errortest.AssertError(t, err, commonerrors.ErrTooLarge)
		})

		t.Run("Test LsRecursiveWithExtensionPatternsAndLimits with nil limits", func(t *testing.T) {
			fs := NewFs(fsType)

			rootDir := t.TempDir()
			err := generateTestTree(fs, rootDir)
			require.NoError(t, err)

			results, err := fs.LsRecursiveWithExclusionPatternsAndLimits(context.Background(), rootDir, nil, true)

			errortest.AssertError(t, err, commonerrors.ErrUndefined)
			assert.Empty(t, results)
		})

		t.Run("Test LsRecursiveWithExtensionPatternsAndLimits with canceled context", func(t *testing.T) {
			fs := NewFs(fsType)

			rootDir := t.TempDir()
			err := generateTestTree(fs, rootDir)
			require.NoError(t, err)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			limits := &Limits{MaxDepth: 4, MaxFileCount: 5, Recursive: true}
			result, err := fs.LsRecursiveWithExclusionPatternsAndLimits(ctx, rootDir, limits, true)
			errortest.AssertError(t, err, commonerrors.ErrCancelled, commonerrors.ErrTimeout)
			assert.Empty(t, result, "Expected no results when context is canceled")
		})

		t.Run("Test LsRecursiveWithExtensionPatternsAndLimits with non-existent directory", func(t *testing.T) {
			fs := NewFs(fsType)

			nonExistentDir := filepath.Join(t.TempDir(), "non_existent_dir")

			limits := &Limits{MaxDepth: 4, MaxFileCount: 5, Recursive: true}
			result, err := fs.LsRecursiveWithExclusionPatternsAndLimits(context.Background(), nonExistentDir, limits, true)
			errortest.AssertError(t, err, os.ErrNotExist, commonerrors.ErrNotFound)
			assert.Empty(t, result, "Expected no results when directory does not exist")
		})
	}
}

func TestLsRecursiveFromOpenedDirectory(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run("Test LsRecursiveFromOpenedDirectory includes directories with available files and directories", func(t *testing.T) {
			fs := NewFs(fsType)

			rootDir := t.TempDir()
			err := generateTestTree(fs, rootDir)
			require.NoError(t, err)

			f, err := fs.GenericOpen(rootDir)
			require.NoError(t, err)
			defer func() { _ = f.Close() }()

			results, err := fs.LsRecursiveFromOpenedDirectory(context.Background(), f, true)
			assert.NoError(t, err)
			expectedFiles := []string{
				filepath.Join(rootDir, "dir2", "dir3", "dir4", "test1.txt"),
				filepath.Join(rootDir, "dir2", "dir3", "dir4", "test2.o"),
				filepath.Join(rootDir, "dir2", "dir3", "dir4", "test3.h"),
				filepath.Join(rootDir, "dir2", "dir5", "test4.txt"),
				filepath.Join(rootDir, "dir2", "dir5", "test5.jar"),

				rootDir,
				filepath.Join(rootDir, "dir2"),
				filepath.Join(rootDir, "dir2", "dir3"),
				filepath.Join(rootDir, "dir2", "dir3", "dir4"),
				filepath.Join(rootDir, "dir2", "dir5"),
			}
			assert.ElementsMatch(t, expectedFiles, results)
		})

		t.Run("Test LsRecursiveFromOpenedDirectory without including directories with available files and directories", func(t *testing.T) {
			fs := NewFs(fsType)

			rootDir := t.TempDir()
			err := generateTestTree(fs, rootDir)
			require.NoError(t, err)

			f, err := fs.GenericOpen(rootDir)
			require.NoError(t, err)
			defer func() { _ = f.Close() }()

			results, err := fs.LsRecursiveFromOpenedDirectory(context.Background(), f, false)
			assert.NoError(t, err)
			expectedFiles := []string{
				filepath.Join(rootDir, "dir2", "dir3", "dir4", "test1.txt"),
				filepath.Join(rootDir, "dir2", "dir3", "dir4", "test2.o"),
				filepath.Join(rootDir, "dir2", "dir3", "dir4", "test3.h"),
				filepath.Join(rootDir, "dir2", "dir5", "test4.txt"),
				filepath.Join(rootDir, "dir2", "dir5", "test5.jar"),
			}
			assert.ElementsMatch(t, expectedFiles, results)
		})

		t.Run("Test LsRecursiveFromOpenedDirectory with canceled context", func(t *testing.T) {
			fs := NewFs(fsType)

			rootDir := t.TempDir()
			err := generateTestTree(fs, rootDir)
			require.NoError(t, err)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			f, err := fs.GenericOpen(rootDir)
			require.NoError(t, err)
			defer func() { _ = f.Close() }()

			result, err := fs.LsRecursiveFromOpenedDirectory(ctx, f, true)
			errortest.AssertError(t, err, commonerrors.ErrCancelled, commonerrors.ErrTimeout)
			assert.Empty(t, result, "Expected no results when context is canceled")
		})

		t.Run("Test LsRecursiveFromOpenedDirectory with non-existent directory", func(t *testing.T) {
			fs := NewFs(fsType)

			result, err := fs.LsRecursiveFromOpenedDirectory(context.Background(), nil, true)
			errortest.AssertError(t, err, commonerrors.ErrUndefined)
			assert.Empty(t, result, "Expected no results when directory does not exist")
		})
	}
}

func generateTestTree(fs FS, rootDir string) error {
	dirs := []string{
		filepath.Join(rootDir, "dir2", "dir3", "dir4"),
		filepath.Join(rootDir, "dir2", "dir5"),
	}
	files := []string{
		filepath.Join(rootDir, "dir2", "dir3", "dir4", "test1.txt"),
		filepath.Join(rootDir, "dir2", "dir3", "dir4", "test2.o"),
		filepath.Join(rootDir, "dir2", "dir3", "dir4", "test3.h"),
		filepath.Join(rootDir, "dir2", "dir5", "test4.txt"),
		filepath.Join(rootDir, "dir2", "dir5", "test5.jar"),
	}

	for _, dir := range dirs {
		if err := fs.MkDirAll(dir, 0755); err != nil {
			return err
		}
	}
	for _, file := range files {
		err := fs.WriteFile(file, []byte("test content"), 0644)
		if err != nil {
			return err
		}
	}
	return nil
}

func globToRegex(glob string) string {
	glob = strings.ReplaceAll(glob, ".", "\\.")
	glob = strings.ReplaceAll(glob, "*", ".*")
	return "^" + glob + "$"
}
