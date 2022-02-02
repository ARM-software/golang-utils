/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package filesystem

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/hashing"
	"github.com/ARM-software/golang-utils/utils/idgen"
	"github.com/ARM-software/golang-utils/utils/platform"
)

func TestExists(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			tmpFile, err := fs.TempFileInTempDir("test-exist-")
			require.Nil(t, err)

			err = tmpFile.Close()
			require.Nil(t, err)

			fileName := tmpFile.Name()
			defer func() { _ = fs.Rm(fileName) }()
			assert.True(t, fs.Exists(tmpFile.Name()))

			randomFile, err := idgen.GenerateUUID4()
			require.Nil(t, err)
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
			require.Nil(t, err)
			defer func() { _ = fs.Rm(tmpDir) }()

			filePath := fmt.Sprintf("%v%v%v", tmpDir, string(fs.PathSeparator()), "test.txt")
			// According to documentation, if the file does not exist, and the O_CREATE flag
			//// is passed, it is created with mode perm (before umask)
			mode := 0600
			file, err := fs.OpenFile(filePath, os.O_RDWR|os.O_CREATE, os.FileMode(mode))
			require.Nil(t, err)
			defer func() { _ = file.Close() }()

			_, err = io.WriteString(file, "initial")
			require.Nil(t, err)
			err = file.Close()
			require.Nil(t, err)

			assert.True(t, fs.Exists(filePath))

			testFileMode(t, fs, filePath, mode)

			ignoredMode := 0322
			file, err = fs.OpenFile(filePath, os.O_WRONLY|os.O_APPEND, os.FileMode(ignoredMode))
			require.Nil(t, err)
			defer func() { _ = file.Close() }()

			_, err = io.WriteString(file, "|append")
			require.Nil(t, err)
			err = file.Close()
			require.Nil(t, err)
			testFileMode(t, fs, filePath, mode)

			ignoredMode = 0400
			file, err = fs.OpenFile(filePath, os.O_RDONLY, os.FileMode(ignoredMode))
			require.Nil(t, err)
			defer func() { _ = file.Close() }()

			contents, err := ioutil.ReadAll(file)
			require.Nil(t, err)
			expectedContents := "initial|append"
			assert.Equal(t, expectedContents, string(contents))
			err = file.Close()
			require.Nil(t, err)
			testFileMode(t, fs, filePath, mode)

			ignoredMode = 0664
			file, err = fs.OpenFile(filePath, os.O_RDWR|os.O_TRUNC, os.FileMode(ignoredMode))
			require.Nil(t, err)
			defer func() { _ = file.Close() }()

			testFileMode(t, fs, filePath, mode)

			contents, err = ioutil.ReadAll(file)
			require.Nil(t, err)
			err = file.Close()
			require.Nil(t, err)
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
	require.Nil(t, err)
	defer func() { _ = tmpFile1.Close() }()
	defer func() { _ = fs.Rm(tmpFile1.Name()) }()
	assert.False(t, IsFileHandleUnset(tmpFile1.Fd()))
	err = tmpFile1.Close()
	require.Nil(t, err)
	_ = fs.Rm(tmpFile1.Name())

	// Test for in memory FS
	fs = NewFs(InMemoryFS)
	tmpFile2, err := fs.TempFileInTempDir("test-filehandle-*.txt")
	require.Nil(t, err)
	defer func() { _ = tmpFile2.Close() }()
	defer func() { _ = fs.Rm(tmpFile2.Name()) }()
	assert.True(t, IsFileHandleUnset(tmpFile2.Fd()))
	err = tmpFile2.Close()
	require.Nil(t, err)
	_ = fs.Rm(tmpFile2.Name())
}

func TestCreate(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			tmpDir, err := fs.TempDirInTempDir("test-create-")
			require.Nil(t, err)
			defer func() { _ = fs.Rm(tmpDir) }()

			txt := "This is a test sentence!!!"
			filePath := fmt.Sprintf("%v%v%v", tmpDir, string(fs.PathSeparator()), "test.txt")
			err = fs.WriteFile(filePath, []byte(txt), 0755)
			require.Nil(t, err)

			assert.True(t, fs.Exists(filePath))

			bytes, err := fs.ReadFile(filePath)
			require.Nil(t, err)

			assert.Equal(t, txt, string(bytes))
			_ = fs.Rm(tmpDir)
		})
	}
}

func TestChmod(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			tmpDir, err := fs.TempDirInTempDir("test-chmod-")
			require.Nil(t, err)
			defer func() { _ = fs.Rm(tmpDir) }()

			filePath := fmt.Sprintf("%v%v%v", tmpDir, string(fs.PathSeparator()), "test.txt")
			file, err := fs.CreateFile(filePath)
			require.Nil(t, err)
			defer func() { _ = file.Close() }()
			err = file.Close()
			require.Nil(t, err)
			require.True(t, fs.Exists(filePath))
			for _, mode := range []int{0666, 0777, 0555, 0766, 0444, 0644} {
				err = fs.Chmod(filePath, os.FileMode(mode))
				if err != nil {
					_ = fs.Chmod(filePath, os.FileMode(mode))
				}
				require.Nil(t, err)
				testFileMode(t, fs, filePath, mode)
			}
			_ = fs.Rm(tmpDir)
		})
	}
}

func createTestFileTree(t *testing.T, fs FS, testDir, basePath string, withLinks bool, fileModTime time.Time, fileAccessTime time.Time) []string {
	err := fs.MkDir(testDir)
	require.Nil(t, err)

	var sLinks []string
	rand.Seed(time.Now().UnixMilli())                                        //nolint:gosec //causes G404: Use of weak random number generator (math/rand instead of crypto/rand) (gosec), So disable gosec
	for i := 0; i < int(math.Max(float64(1), float64(rand.Intn(10)))); i++ { //nolint:gosec //causes G404: Use of weak random number generator (math/rand instead of crypto/rand) (gosec), So disable gosec
		c := fmt.Sprint("test", i+1)
		path := filepath.Join(testDir, c)

		err := fs.MkDir(path)
		require.Nil(t, err)

		for j := 0; j < int(math.Max(float64(1), float64(rand.Intn(10)))); j++ { //nolint:gosec //causes G404: Use of weak random number generator (math/rand instead of crypto/rand) (gosec), So disable gosec
			c := fmt.Sprint("test", j+1)
			path := filepath.Join(path, c)

			err := fs.MkDir(path)
			require.Nil(t, err)

			if withLinks {
				if len(sLinks) > 0 {
					c1 := fmt.Sprint("link", j+1)
					c2 := filepath.Join(path, c1)
					err = fs.Symlink(sLinks[0], c2)
					require.Nil(t, err)
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
				require.Nil(t, err)
			}
		}
	}
	var tree []string
	err = fs.ListDirTree(testDir, &tree)
	require.Nil(t, err)

	//unifying timestamps
	for _, path := range tree {
		err = fs.Chtimes(path, fileAccessTime, fileModTime)
		require.Nil(t, err)
	}

	return tree
}

func TestConvertPaths(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)

			// set up temp directory
			tmpDir, err := fs.TempDirInTempDir("temp")
			require.Nil(t, err)
			defer func() { _ = fs.Rm(tmpDir) }()

			// create a directory for the test
			tree := createTestFileTree(t, fs, tmpDir, "", false, time.Now(), time.Now())
			relTree, err := fs.ConvertToRelativePath(tmpDir, tree)
			require.Nil(t, err)
			absTree, err := fs.ConvertToAbsolutePath(tmpDir, relTree)
			require.Nil(t, err)

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

func TestZip(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)

			for i := 0; i < 10; i++ {
				func() {
					// create a directory for the test
					tmpDir, err := fs.TempDirInTempDir("temp")
					require.Nil(t, err)
					defer func() { _ = fs.Rm(tmpDir) }()

					testDir := filepath.Join(tmpDir, "test") // remember to read tmpdir at beginning
					zipfile := filepath.Join(tmpDir, "test.zip")
					outDir := filepath.Join(tmpDir, "output")

					// create a file tree for the test
					// Regarding timestamp preservation, the following link should be read as it gives some insight about how zip tools work or behave
					// https://blog.joshlemon.com.au/dfir-for-compressed-files/
					// The bottom line though is that the zip specification stipulates that file timestamp is stored using MS-DOS format which has a 2-second precision.
					// see Section 4.4.6 of the spec https://pkware.cachefly.net/webdocs/casestudies/APPNOTE.TXT
					// As a result, the built-in timestamp resolution of files in a ZIP archive is only two seconds and so, file timestamps will not be fully preserved when a zip/unzip is performed.
					// Making the FS think the tree was made 3 seconds ago.
					tree := createTestFileTree(t, fs, testDir, "", false, time.Now().Add(-3*time.Second), time.Now())

					// zip the directory into the zipfile
					err = fs.Zip(testDir, zipfile)
					require.Nil(t, err)

					// unzip
					tree2, err := fs.Unzip(zipfile, outDir)
					require.Nil(t, err)

					// Check no files were lost in the zip/unzip process.
					relativeSrcTree, err := fs.ConvertToRelativePath(testDir, tree)
					require.Nil(t, err)
					relativeResultTree, err := fs.ConvertToRelativePath(outDir, tree2)
					require.Nil(t, err)
					sort.Strings(relativeSrcTree)
					sort.Strings(relativeResultTree)
					require.Equal(t, relativeSrcTree, relativeResultTree)

					hasher, err := NewFileHash(hashing.HashXXHash)
					require.Nil(t, err)

					// check for size, timestamp, hash preservation
					for _, path := range relativeSrcTree {
						srcFilePath := filepath.Join(testDir, path)
						fileinfoSrc, err := fs.Lstat(srcFilePath)
						require.Nil(t, err)
						resultFilePath := filepath.Join(outDir, path)
						fileinfoResult, err := fs.Lstat(resultFilePath)
						require.Nil(t, err)
						//TODO handle links separately
						if IsSymLink(fileinfoSrc) {
							continue
						}
						// Check sizes
						assert.Equal(t, fileinfoSrc.Size(), fileinfoResult.Size())

						// Check file timestamps
						// FIXME understand why the timestamp is not preserved when using the FS in memory
						if fs.GetType() != InMemoryFS {
							// Allowing some tolerance in case of time rounding or truncation happening (https://golang.org/pkg/os/#Chtimes) + the 2-second time resolution of zip
							// see comment above
							assert.True(t, math.Abs(fileinfoSrc.ModTime().Sub(fileinfoResult.ModTime()).Seconds()) <= 2)

							fileTimesSrc, err := fs.StatTimes(filepath.Join(testDir, path))
							require.Nil(t, err)
							fileTimeResult, err := fs.StatTimes(filepath.Join(outDir, path))
							require.Nil(t, err)
							assert.True(t, math.Abs(fileTimesSrc.ModTime().Sub(fileTimeResult.ModTime()).Seconds()) <= 2)
						}

						// perform hash comparison
						if IsRegularFile(fileinfoSrc) {
							hashSrc, err := hasher.CalculateFile(fs, srcFilePath)
							require.Nil(t, err)
							hashResult, err := hasher.CalculateFile(fs, resultFilePath)
							require.Nil(t, err)
							assert.Equal(t, hashSrc, hashResult)
						}
					}
					err = fs.Rm(tmpDir)
					require.Nil(t, err)
				}()
			}
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
			require.Nil(t, err)
			defer func() { _ = fs.Rm(tmpDir) }()

			txt := "This is a test sentence!!!"
			filePath := fmt.Sprintf("%v%v%v", tmpDir, string(fs.PathSeparator()), "test.txt")
			err = fs.WriteFile(filePath, []byte(txt), 0755)
			require.Nil(t, err)

			symlink := filepath.Join(tmpDir, "symlink-tofile")
			hardlink := filepath.Join(tmpDir, "hardlink-tofile")

			err = fs.Symlink(filePath, symlink)
			if errors.Is(err, commonerrors.ErrNotImplemented) || errors.Is(err, afero.ErrNoSymlink) {
				return
			}
			require.Nil(t, err)

			err = fs.Link(filePath, hardlink)
			require.Nil(t, err)

			assert.True(t, fs.Exists(symlink))
			assert.True(t, fs.Exists(hardlink))

			isLink, err := fs.IsLink(symlink)
			require.Nil(t, err)
			assert.True(t, isLink)

			isFile, err := fs.IsFile(symlink)
			require.Nil(t, err)
			assert.True(t, isFile)

			isLink, err = fs.IsLink(hardlink)
			require.Nil(t, err)
			assert.False(t, isLink)

			isFile, err = fs.IsFile(hardlink)
			require.Nil(t, err)
			assert.True(t, isFile)

			link, err := fs.Readlink(symlink)
			require.Nil(t, err)
			assert.Equal(t, filePath, link)

			link, err = fs.Readlink(hardlink)
			require.NotNil(t, err)
			assert.Equal(t, "", link)

			bytes, err := fs.ReadFile(symlink)
			require.Nil(t, err)
			assert.Equal(t, txt, string(bytes))

			bytes, err = fs.ReadFile(hardlink)
			require.Nil(t, err)
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
			require.Nil(t, err)

			err = tmpFile.Close()
			require.Nil(t, err)
			defer func() { _ = fs.Rm(tmpFile.Name()) }()

			fileName := tmpFile.Name()
			defer func() { _ = fs.Rm(fileName) }()
			assert.True(t, fs.Exists(tmpFile.Name()))
			fileTimes, err := fs.StatTimes(fileName)
			require.Nil(t, err)
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
			require.Nil(t, err)
			defer func() { _ = fs.Rm(tmpDir) }()

			empty, err := fs.IsEmpty(tmpDir)
			require.Nil(t, err)
			assert.True(t, empty)

			tmpFile, err := fs.TempFile(tmpDir, "test-isempty-*.test")
			require.Nil(t, err)
			_ = tmpFile.Close()
			defer func() { _ = fs.Rm(tmpFile.Name()) }()

			isFileEmpty, err := fs.IsEmpty(tmpFile.Name())
			require.Nil(t, err)
			assert.True(t, isFileEmpty)

			checkNotEmpty(t, fs, tmpDir)

			testDir := filepath.Join(tmpDir, "testDir")
			err = fs.MkDir(testDir)
			require.Nil(t, err)
			defer func() { _ = fs.Rm(testDir) }()

			checkNotEmpty(t, fs, tmpDir)

			err = fs.Rm(testDir)
			require.Nil(t, err)

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
			require.Nil(t, err)
			defer func() { _ = fs.Rm(tmpDir) }()

			tmpFile, err := fs.TempFile(tmpDir, "test-isempty-*.test")
			require.Nil(t, err)
			_ = tmpFile.Close()
			defer func() { _ = fs.Rm(tmpFile.Name()) }()

			timesDirOrig, err := fs.StatTimes(tmpDir)
			require.Nil(t, err)
			timesFileOrig, err := fs.StatTimes(tmpFile.Name())
			require.Nil(t, err)
			newTimeA := time.Now().Add(-time.Hour)
			newTimeM := time.Now().Add(-30 * time.Minute)
			err = fs.Chtimes(tmpDir, newTimeA, newTimeM)
			require.Nil(t, err)
			err = fs.Chtimes(tmpFile.Name(), newTimeA, newTimeM)
			require.Nil(t, err)

			timesDirMod, err := fs.StatTimes(tmpDir)
			require.Nil(t, err)
			timesFileMod, err := fs.StatTimes(tmpFile.Name())
			require.Nil(t, err)

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
			require.Nil(t, err)
		})
	}
}

func TestCleanDir(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			tmpDir, err := fs.TempDirInTempDir("test-cleandir-")
			require.Nil(t, err)
			defer func() { _ = fs.Rm(tmpDir) }()

			empty, err := fs.IsEmpty(tmpDir)
			require.Nil(t, err)
			assert.True(t, empty)

			tmpFile, err := fs.TempFile(tmpDir, "test-cleandir-*.test")
			require.Nil(t, err)
			err = tmpFile.Close()
			require.Nil(t, err)

			checkNotEmpty(t, fs, tmpDir)

			testDir := filepath.Join(tmpDir, "testDir")
			err = fs.MkDir(testDir)
			require.Nil(t, err)

			checkNotEmpty(t, fs, tmpDir)

			err = fs.CleanDir(tmpDir)
			require.Nil(t, err)
			assert.True(t, fs.Exists(tmpDir))

			empty, err = fs.IsEmpty(tmpDir)
			require.Nil(t, err)
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
			require.Nil(t, err)
			defer func() { _ = fs.Rm(tmpDir) }()

			tmpDir1, err := fs.TempDir(tmpDir, "test-ls-")
			require.Nil(t, err)
			tmpFile, err := fs.TempFile(tmpDir, "test-ls-*.test")
			require.Nil(t, err)

			err = tmpFile.Close()
			require.Nil(t, err)

			fileName := tmpFile.Name()

			files, err := fs.Ls(tmpDir)
			require.Nil(t, err)
			assert.Len(t, files, 2)
			_, found := collection.Find(&files, filepath.Base(fileName))
			assert.True(t, found)
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
			tmpDir, err := fs.TempDirInTempDir("temp")
			require.Nil(t, err)
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
			require.Nil(t, err)

			// Sort lists so that equal works better
			sort.Strings(tree)
			sort.Strings(walkList)

			// check if equal
			require.Equal(t, walkList, tree)
		})
	}
}

func TestGarbageCollection(t *testing.T) {
	for _, fsType := range []int{StandardFS} {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			tmpDir, err := fs.TempDirInTempDir("test-gc-")
			require.Nil(t, err)
			defer func() { _ = fs.Rm(tmpDir) }()

			tmpDir1, err := fs.TempDir(tmpDir, "test-gc-")
			require.Nil(t, err)
			tmpDir2, err := fs.TempDir(tmpDir, "test-gc-")
			require.Nil(t, err)
			tmpFile, err := fs.TempFile(tmpDir, "test-gc-*.test")
			require.Nil(t, err)

			err = tmpFile.Close()
			require.Nil(t, err)
			tmpFile1, err := fs.TempFile(tmpDir1, "test-gc-*.test")
			require.Nil(t, err)
			err = tmpFile1.Close()
			require.Nil(t, err)

			ctime := time.Now()
			time.Sleep(500 * time.Millisecond)

			tmpDir3, err := fs.TempDir(tmpDir, "test-gc-")
			require.Nil(t, err)
			tmpFile2, err := fs.TempFile(tmpDir3, "test-gc-*.test")
			require.Nil(t, err)
			err = tmpFile2.Close()
			require.Nil(t, err)
			tmpFile3, err := fs.TempFile(tmpDir2, "test-gc-*.test")
			require.Nil(t, err)
			err = tmpFile3.Close()
			require.Nil(t, err)

			files, err := fs.Ls(tmpDir)
			require.Nil(t, err)
			assert.Len(t, files, 4)

			elapsedTime := time.Since(ctime)

			err = fs.GarbageCollect(tmpDir, elapsedTime)
			require.Nil(t, err)

			files, err = fs.Ls(tmpDir)
			require.Nil(t, err)
			assert.Len(t, files, 2)
			_, found := collection.Find(&files, filepath.Base(tmpDir2))
			assert.True(t, found)
			_, found = collection.Find(&files, filepath.Base(tmpDir3))
			assert.True(t, found)
			files, err = fs.Ls(tmpDir3)
			require.Nil(t, err)
			assert.Len(t, files, 1)
			_, found = collection.Find(&files, filepath.Base(tmpFile2.Name()))
			assert.True(t, found)
			_ = fs.Rm(tmpDir)
		})
	}
}

func TestExcludes(t *testing.T) {
	t.Parallel() // marks TLog as capable of running in parallel with other tests
	var listOfPaths = []string{
		"some/path", "somepath", ".snapshot", ".snapshot/path", "test/.snapshot/some/path", ".snapshot123", ".snapshot123/path", "test/.snapshot123/some-path", "test/.snapshot123/some/path",
	}
	tests := []struct {
		inputlist       []string
		exclusions      []string
		expectedResults []string
	}{
		{
			inputlist:       listOfPaths,
			exclusions:      []string{},
			expectedResults: listOfPaths,
		},
		{
			inputlist:       listOfPaths,
			exclusions:      []string{"noexclusion"},
			expectedResults: listOfPaths,
		},
		{
			inputlist:       []string{},
			exclusions:      []string{"any"},
			expectedResults: []string{},
		},
		{
			inputlist:       listOfPaths,
			exclusions:      []string{""},
			expectedResults: listOfPaths,
		},
		{
			inputlist:       listOfPaths,
			exclusions:      []string{"some.*"},
			expectedResults: []string{".snapshot", ".snapshot/path", ".snapshot123", ".snapshot123/path"},
		},
		{
			inputlist:       listOfPaths,
			exclusions:      []string{".*path"},
			expectedResults: []string{".snapshot", ".snapshot123"},
		},
		{
			inputlist:       listOfPaths,
			exclusions:      []string{"[.]snapshot.*"},
			expectedResults: []string{"some/path", "somepath"},
		},
		{
			inputlist:       listOfPaths,
			exclusions:      []string{"[.]snapshot.*/.*"},
			expectedResults: []string{"some/path", "somepath", ".snapshot", ".snapshot123"},
		},
		{
			inputlist:       listOfPaths,
			exclusions:      []string{"[.]snapshot.*", ".*path"},
			expectedResults: []string{},
		},
	}
	for i, tt := range tests {
		tt := tt // NOTE: https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			t.Parallel() // marks each test case as capable of running in parallel with each other
			actualList, err := ExcludeAll(tt.inputlist, tt.exclusions...)
			require.Nil(t, err)
			assert.Equal(t, tt.expectedResults, actualList)
		})
	}

}

func TestFindAll(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			tmpDir, err := fs.TempDirInTempDir("test-findall-")
			require.Nil(t, err)
			defer func() { _ = fs.Rm(tmpDir) }()

			empty, err := fs.IsEmpty(tmpDir)
			require.Nil(t, err)
			assert.True(t, empty)

			level1 := filepath.Join(tmpDir, "level1")
			err = fs.MkDir(level1)
			require.Nil(t, err)

			level2 := filepath.Join(level1, "level2")
			err = fs.MkDir(level2)
			require.Nil(t, err)

			level3 := filepath.Join(level2, "level3")
			err = fs.MkDir(level3)
			require.Nil(t, err)

			checkNotEmpty(t, fs, tmpDir)

			tmpFile1, err := fs.TempFile(level1, "test-findall-*.test1")
			require.Nil(t, err)
			err = tmpFile1.Close()
			require.Nil(t, err)

			tmpFile2, err := fs.TempFile(level2, "test-findall-*.test2")
			require.Nil(t, err)
			err = tmpFile2.Close()
			require.Nil(t, err)

			tmpFile3, err := fs.TempFile(level3, "test-findall-*.test3")
			require.Nil(t, err)
			err = tmpFile3.Close()
			require.Nil(t, err)

			list, err := fs.FindAll(tmpDir, ".test1", "test3")
			require.Nil(t, err)
			assert.Equal(t, 2, len(list))

			for _, file := range list {
				ext := filepath.Ext(file)
				assert.Equal(t, ".test", ext[0:len(ext)-1])
				assert.True(t, ext[len(ext)-1] == '1' || ext[len(ext)-1] == '3')
			}
			err = fs.Rm(tmpDir)
			require.Nil(t, err)
		})
	}
}

func TestCopy(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			tmpDir, err := fs.TempDirInTempDir("test-copy-")
			require.Nil(t, err)
			defer func() { _ = fs.Rm(tmpDir) }()

			empty, err := fs.IsEmpty(tmpDir)
			require.Nil(t, err)
			assert.True(t, empty)

			tmpFile, err := fs.TempFile(tmpDir, "test-copy-*.test")
			require.Nil(t, err)
			err = tmpFile.Close()
			require.Nil(t, err)

			checkNotEmpty(t, fs, tmpDir)
			checkCopy(t, fs, tmpFile.Name(), filepath.Join(tmpDir, "newDir"))
			_ = fs.Rm(tmpDir)
		})
	}
}

func TestCopyFolder(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			tmpDir, err := fs.TempDirInTempDir("test-src-copy-")
			require.Nil(t, err)

			parentDir, err := fs.TempDir(tmpDir, "parentDir-")
			require.Nil(t, err)

			childDir, err := fs.TempDir(parentDir, "childDir-")
			require.Nil(t, err)

			_, err = fs.TempDir(childDir, "gcDir-")
			require.Nil(t, err)

			testDir, err := fs.TempDirInTempDir("test-dest-dir-")
			require.Nil(t, err)
			defer func() {
				_ = fs.Rm(tmpDir)
				_ = fs.Rm(testDir)
			}()

			checkNotEmpty(t, fs, parentDir)
			checkCopyDir(t, fs, parentDir, filepath.Join(testDir, "newDir"))
		})
	}
}

func TestMove(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			tmpDir, err := fs.TempDirInTempDir("test-move-")
			require.Nil(t, err)
			defer func() { _ = fs.Rm(tmpDir) }()

			empty, err := fs.IsEmpty(tmpDir)
			require.Nil(t, err)
			assert.True(t, empty)

			tmpFile, err := fs.TempFile(tmpDir, "test-move-*.test")
			require.Nil(t, err)
			err = tmpFile.Close()
			require.Nil(t, err)

			testDir := filepath.Join(tmpDir, "testDir")
			err = fs.MkDir(testDir)
			require.Nil(t, err)

			tmpFile2, err := fs.TempFile(testDir, "test-move-*.test")
			require.Nil(t, err)
			err = tmpFile2.Close()
			require.Nil(t, err)

			testDir2 := filepath.Join(testDir, "testDir2")
			err = fs.MkDir(testDir2)
			require.Nil(t, err)
			checkNotEmpty(t, fs, tmpDir)
			err = fs.Move(tmpFile.Name(), tmpFile.Name())
			require.Nil(t, err)
			err = fs.Move(testDir, testDir)
			require.Nil(t, err)
			checkMove(t, fs, tmpFile.Name(), filepath.Join(tmpDir, "test.test"))
			checkMove(t, fs, testDir, filepath.Join(tmpDir, "testDir3"))
			_ = fs.Rm(tmpDir)
		})
	}
}

func TestReadFile(t *testing.T) {
	fs := NewFs(InMemoryFS)
	tmpFile, err := fs.TempFileInTempDir("test-readfile-")
	require.Nil(t, err)

	byteInput := []byte("Here is a Test string....")
	_, err = tmpFile.Write(byteInput)
	require.Nil(t, err)

	err = tmpFile.Close()
	require.Nil(t, err)

	fileName := tmpFile.Name()
	defer func() { _ = fs.Rm(fileName) }()

	byteOut, err := fs.ReadFile(tmpFile.Name())
	require.Nil(t, err)
	assert.Equal(t, byteOut, byteInput)

	_, err = fs.ReadFile("unknown_file")
	assert.NotNil(t, err)
}

func TestGetFileSize(t *testing.T) {
	fs := NewFs(InMemoryFS)
	tmpFile, err := fs.TempFileInTempDir("test-filesize-")
	require.Nil(t, err)

	for indx := 0; indx < 50; indx++ {
		_, _ = tmpFile.WriteString(" Here is a Test string....")
	}

	err = tmpFile.Close()
	require.Nil(t, err)

	fileName := tmpFile.Name()
	defer func() { _ = fs.Rm(fileName) }()

	size, err := fs.GetFileSize(tmpFile.Name())
	require.Nil(t, err)
	assert.Equal(t, int64(1300), size)

	_, err = fs.GetFileSize("Unknown-File")
	assert.NotNil(t, err)
}

func TestUnzip(t *testing.T) {
	fs := NewFs(StandardFS)

	testInDir := "testdata"
	testFile := "testunzip"
	srcPath := filepath.Join(testInDir, testFile+".zip")
	destPath, err := fs.TempDirInTempDir("unzip")
	require.Nil(t, err)
	defer func() { _ = fs.Rm(destPath) }()
	outPath := filepath.Join(destPath, testFile)
	expectedfileList := []string{
		filepath.Join(outPath, "readme.txt"),
		filepath.Join(outPath, "test.txt"),
		filepath.Join(outPath, "todo.txt"),
		filepath.Join(outPath, "child.zip"),
	}
	sort.Strings(expectedfileList)

	/* Check Unzip */
	fileList, err := fs.Unzip(srcPath, destPath)

	sort.Strings(fileList)
	assert.Nil(t, err)
	assert.Equal(t, len(fileList), 4)
	assert.Equal(t, expectedfileList, fileList)

	/* Source zip file not available */
	srcPath = filepath.Join(testInDir, "unknownfile.zip")
	_, err = fs.Unzip(srcPath, destPath)
	assert.NotNil(t, err)

	/* Invalid source path */
	srcPath = filepath.Join(testInDir, "invalidzipfile.zip")
	_, err = fs.Unzip(srcPath, destPath)
	assert.NotNil(t, err)
}

func TestInvalidUnzip(t *testing.T) {
	fs := NewFs(StandardFS)
	// Testing zip file attached to https://github.com/golang/go/issues/10741
	testInDir := "testdata"
	testFile := "zipwithnonutf8.zip"
	srcPath := filepath.Join(testInDir, testFile)
	destPath, err := fs.TempDirInTempDir("unzip")
	require.Nil(t, err)
	defer func() { _ = fs.Rm(destPath) }()
	/* Check Unzip */
	_, err = fs.Unzip(srcPath, destPath)
	require.NotNil(t, err)
	assert.True(t, commonerrors.Any(err, commonerrors.ErrInvalid))
}

func TestSubDirectories(t *testing.T) {
	fs := NewFs(StandardFS)
	testDir := "testdata"

	// Test empty directory
	dirlist, err := fs.SubDirectories(testDir)
	assert.Nil(t, dirlist)
	assert.Nil(t, err)

	// Test directory with subdirectories
	testInput := filepath.Join(testDir, "ARM")
	_ = fs.MkDir(filepath.Join(testInput, "CMSIS"))
	_ = fs.MkDir(filepath.Join(testInput, "Tools"))
	file, _ := fs.CreateFile(filepath.Join(testInput, "testfile.ini"))
	defer func() {
		_ = file.Close()
		_ = fs.Rm(testInput)
	}()

	dirlist, err = fs.SubDirectories(testInput)
	expected := []string{"CMSIS", "Tools"}
	assert.Equal(t, expected, dirlist)
	assert.Nil(t, err)
}

func TestListDirTree(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)
			testDir, err := fs.TempDirInTempDir("test-list-Dir")
			require.Nil(t, err)
			defer func() { _ = fs.Rm(testDir) }()

			empty, err := fs.IsEmpty(testDir)
			require.Nil(t, err)
			assert.True(t, empty)

			parentDirPath, err := fs.TempDir(testDir, "parentDir-")
			require.Nil(t, err)

			testFile, err := fs.TempFile(testDir, "test-file-*.test")
			require.Nil(t, err)
			err = testFile.Close()
			require.Nil(t, err)

			childDirPath, err := fs.TempDir(parentDirPath, "childDir-")
			require.Nil(t, err)

			gcDirPath, err := fs.TempDir(childDirPath, "gcDir-")
			require.Nil(t, err)

			checkNotEmpty(t, fs, testDir)

			list := []string{}
			err = fs.ListDirTree(testDir, &list)
			assert.Equal(t, 4, len(list))
			require.Nil(t, err)

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
				fileList = append(fileList, strings.Replace(item, filepath.Dir(parentDirPath), "", -1))
			}

			sort.Strings(fileList)
			sort.Strings(expectedList)
			assert.True(t, reflect.DeepEqual(fileList, expectedList))
		})
	}
}

func checkCopyDir(t *testing.T, fs FS, src string, dest string) {
	assert.True(t, fs.Exists(src))
	assert.False(t, fs.Exists(dest))

	err := fs.Copy(src, dest)
	require.Nil(t, err)

	defer func() { _ = fs.Rm(dest) }()
	assert.True(t, fs.Exists(src))
	assert.True(t, fs.Exists(dest))

	srcFiles := []string{}
	destFiles := []string{}

	err = fs.ListDirTree(src, &srcFiles)
	require.Nil(t, err)

	destPath := filepath.Join(dest, filepath.Base(src))
	err = fs.ListDirTree(destPath, &destFiles)
	require.Nil(t, err)

	var srcContent []string
	var destContent []string

	for _, item := range srcFiles {
		srcContent = append(srcContent, strings.Replace(item, filepath.Dir(src), "", -1))
	}
	for _, item := range destFiles {
		destContent = append(destContent, strings.Replace(item, filepath.Dir(destPath), "", -1))
	}

	sort.Strings(srcContent)
	sort.Strings(destContent)
	assert.True(t, reflect.DeepEqual(srcContent, destContent))
}

func checkCopy(t *testing.T, fs FS, oldFile string, dest string) {

	assert.True(t, fs.Exists(oldFile))
	assert.False(t, fs.Exists(dest))

	empty, err := fs.IsEmpty(oldFile)
	require.Nil(t, err)

	err = fs.Copy(oldFile, dest)
	require.Nil(t, err)

	defer func() { _ = fs.Rm(dest) }()
	assert.True(t, fs.Exists(oldFile))
	assert.True(t, fs.Exists(dest))

	empty2, err := fs.IsEmpty(filepath.Join(dest, filepath.Base(oldFile)))
	require.Nil(t, err)
	assert.Equal(t, empty, empty2)
}

func checkMove(t *testing.T, fs FS, oldFile string, newFile string) {

	assert.True(t, fs.Exists(oldFile))
	assert.False(t, fs.Exists(newFile))

	empty, err := fs.IsEmpty(oldFile)
	require.Nil(t, err)

	err = fs.Move(oldFile, newFile)
	require.Nil(t, err)
	defer func() { _ = fs.Rm(newFile) }()
	assert.False(t, fs.Exists(oldFile))
	assert.True(t, fs.Exists(newFile))

	empty2, err := fs.IsEmpty(newFile)
	require.Nil(t, err)
	assert.Equal(t, empty, empty2)
}

func checkNotEmpty(t *testing.T, fs FS, tmpDir string) {
	empty, err := fs.IsEmpty(tmpDir)
	require.Nil(t, err)
	assert.False(t, empty)
}

func testFileMode(t *testing.T, fs FS, filePath string, mode int) {
	fi, err := fs.Lstat(filePath)
	require.Nil(t, err)
	if platform.IsWindows() {
		// Only user's permissions matter. Execution rights are not really considered.
		userExpectedPermission := mode & 0600
		userActualPermission := int(fi.Mode().Perm()) & 0600
		assert.Equal(t, userExpectedPermission, userActualPermission)
	} else {
		assert.Equal(t, mode, int(fi.Mode().Perm()))
	}
}
