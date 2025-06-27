package filesystem

import (
	"context"
	"fmt"
	"math"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/hashing"
	"github.com/ARM-software/golang-utils/utils/units/multiplication"
	"github.com/ARM-software/golang-utils/utils/units/size"
)

func TestZip(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)

			for i := 0; i < 10; i++ {
				func() {
					// create a directory for the test
					tmpDir, err := fs.TempDirInTempDir("temp")
					require.NoError(t, err)
					defer func() { _ = fs.Rm(tmpDir) }()

					testDir := FilePathJoin(fs, tmpDir, "test") // remember to read tmpdir at beginning
					zipfile := FilePathJoin(fs, tmpDir, "test.zip")
					outDir := FilePathJoin(fs, tmpDir, "output")

					// create a file tree for the test
					// Regarding timestamp preservation, the following link should be read as it gives some insight about how zip tools work or behave
					// https://blog.joshlemon.com.au/dfir-for-compressed-files/
					// The bottom line though is that the zip specification stipulates that file timestamp is stored using MS-DOS format which has a 2-second precision.
					// see Section 4.4.6 of the spec https://pkware.cachefly.net/webdocs/casestudies/APPNOTE.TXT
					// As a result, the built-in timestamp resolution of files in a ZIP archive is only two seconds and so, file timestamps will not be fully preserved when a zip/unzip is performed.
					// Making the FS think the tree was made 3 seconds ago.
					tree := GenerateTestFileTree(t, fs, testDir, "", false, time.Now().Add(-3*time.Second), time.Now())

					// zip the directory into the zipfile
					err = fs.Zip(testDir, zipfile)
					require.NoError(t, err)

					// unzip
					tree2, err := fs.Unzip(zipfile, outDir)
					require.NoError(t, err)

					// Check no files were lost in the zip/unzip process.
					relativeSrcTree, err := fs.ConvertToRelativePath(testDir, tree...)
					require.NoError(t, err)
					relativeResultTree, err := fs.ConvertToRelativePath(outDir, tree2...)
					require.NoError(t, err)
					sort.Strings(relativeSrcTree)
					sort.Strings(relativeResultTree)
					require.Equal(t, relativeSrcTree, relativeResultTree)

					hasher, err := NewFileHash(hashing.HashXXHash)
					require.NoError(t, err)

					// check for size, timestamp, hash preservation
					for _, path := range relativeSrcTree {
						srcFilePath := FilePathJoin(fs, testDir, path)
						fileinfoSrc, err := fs.Lstat(srcFilePath)
						require.NoError(t, err)
						resultFilePath := FilePathJoin(fs, outDir, path)
						fileinfoResult, err := fs.Lstat(resultFilePath)
						require.NoError(t, err)
						// TODO handle links separately
						if IsSymLink(fileinfoSrc) {
							continue
						}
						// Check sizes
						assert.Equal(t, fileinfoSrc.Size(), fileinfoResult.Size())

						// Check file timestamps
						// FIXME understand why the timestamp is not preserved when using the FS in memory
						// https://github.com/spf13/afero/issues/297
						if fs.GetType() != InMemoryFS {
							// Allowing some tolerance in case of time rounding or truncation happening (https://golang.org/pkg/os/#Chtimes) + the 2-second time resolution of zip
							// see comment above
							assert.True(t, math.Abs(fileinfoSrc.ModTime().Sub(fileinfoResult.ModTime()).Seconds()) <= 2)

							fileTimesSrc, err := fs.StatTimes(FilePathJoin(fs, testDir, path))
							require.NoError(t, err)
							fileTimeResult, err := fs.StatTimes(FilePathJoin(fs, outDir, path))
							require.NoError(t, err)
							assert.True(t, math.Abs(fileTimesSrc.ModTime().Sub(fileTimeResult.ModTime()).Seconds()) <= 2)
						}

						// perform hash comparison
						if IsRegularFile(fileinfoSrc) {
							hashSrc, err := hasher.CalculateFile(fs, srcFilePath)
							require.NoError(t, err)
							hashResult, err := hasher.CalculateFile(fs, resultFilePath)
							require.NoError(t, err)
							assert.Equal(t, hashSrc, hashResult)
						}
					}
					err = fs.Rm(tmpDir)
					require.NoError(t, err)
				}()
			}
		})
	}
}

func TestZipWithExclusion(t *testing.T) {
	for _, fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(fsType)

			for i := 0; i < 10; i++ {
				func() {
					// create a directory for the test
					tmpDir, err := fs.TempDirInTempDir("temp")
					require.NoError(t, err)
					defer func() { _ = fs.Rm(tmpDir) }()

					testDir := FilePathJoin(fs, tmpDir, "test") // remember to read tmpdir at beginning
					zipfile := FilePathJoin(fs, tmpDir, "test.zip")
					outDir := FilePathJoin(fs, tmpDir, "output")

					// create a file tree for the test
					tree := GenerateTestFileTree(t, fs, testDir, "", false, time.Now().Add(-3*time.Second), time.Now())

					exclusionPatterns := []string{".*test2.*", ".*test3.*"}

					cleansedTree, err := fs.ExcludeAll(tree, exclusionPatterns...)
					require.NoError(t, err)

					// zip the directory into the zipfile ignoring some paths.
					err = fs.ZipWithContextAndLimitsAndExclusionPatterns(context.TODO(), testDir, zipfile, DefaultLimits(), exclusionPatterns...)
					require.NoError(t, err)

					// unzip
					tree2, err := fs.Unzip(zipfile, outDir)
					require.NoError(t, err)

					// Check no files were lost in the zip/unzip process.
					relativeSrcTree, err := fs.ConvertToRelativePath(testDir, cleansedTree...)
					require.NoError(t, err)
					relativeResultTree, err := fs.ConvertToRelativePath(outDir, tree2...)
					require.NoError(t, err)
					sort.Strings(relativeSrcTree)
					sort.Strings(relativeResultTree)
					require.Equal(t, relativeSrcTree, relativeResultTree)

					err = fs.Rm(tmpDir)
					require.NoError(t, err)
				}()
			}
		})
	}
}

func Test_IsZip(t *testing.T) {
	fs := NewFs(StandardFS)

	testInDir := "testdata"
	tests := []struct {
		testFile string
		isZip    bool
	}{
		{
			testFile: "",
			isZip:    false,
		},
		{
			testFile: FilePathJoin(fs, testInDir, "5MB.zip"),
			isZip:    false,
		},
		{
			testFile: FilePathJoin(fs, testInDir, "1KB.bin"),
			isZip:    false,
		},
		{
			testFile: FilePathJoin(fs, testInDir, "1KB.gz"),
			isZip:    false,
		},
		{
			testFile: FilePathJoin(fs, testInDir, "unknownfile.zip"),
			isZip:    true, // File does not exist but has a ZIP name
		},
		{
			testFile: FilePathJoin(fs, testInDir, "unknownfile.zipx"),
			isZip:    true, // File does not exist but has a ZIP name
		},
		{
			testFile: FilePathJoin(fs, testInDir, "invalidzipfile.zip"),
			isZip:    false,
		},
		{
			testFile: FilePathJoin(fs, testInDir, "testunzip2.7z"),
			isZip:    false,
		},
		{
			testFile: FilePathJoin(fs, testInDir, "testunzip.zip"),
			isZip:    true,
		},
		{
			testFile: FilePathJoin(fs, testInDir, "42.zip"),
			isZip:    true,
		},
		{
			testFile: FilePathJoin(fs, testInDir, "zip-bomb.zip"),
			isZip:    true,
		},
		{
			testFile: FilePathJoin(fs, testInDir, "zbsm.zip"),
			isZip:    true,
		},
		{
			testFile: FilePathJoin(fs, testInDir, "zipwithnonutf8.zip"),
			isZip:    true,
		},
		{
			testFile: FilePathJoin(fs, testInDir, "zipwithnonutf8filenames2.zip"),
			isZip:    true,
		},
		{
			testFile: FilePathJoin(fs, testInDir, "k64f.pack"),
			isZip:    true,
		},
		{
			testFile: FilePathJoin(fs, testInDir, "10.zip"),
			isZip:    true,
		},
		{
			testFile: FilePathJoin(fs, testInDir, "child.zip"),
			isZip:    true,
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(fmt.Sprintf("#%v (%v)", i, filepath.Base(test.testFile)), func(t *testing.T) {
			assert.Equal(t, test.isZip, fs.IsZip(test.testFile))
		})
	}
}

func TestUnzip_Limits(t *testing.T) {
	fs := NewFs(StandardFS)

	testInDir := "testdata"
	testFile := "validlargezipfile"
	srcPath := FilePathJoin(fs, testInDir, testFile+".zip")
	destPath, err := fs.TempDirInTempDir("unzip-limits-")
	require.NoError(t, err)
	defer func() { _ = fs.Rm(destPath) }()
	limits := NewLimits(int64(size.GiB), uint64(size.KiB), multiplication.Mega, 1, true) // Total size limited to 10 Kb

	empty, err := fs.IsEmpty(destPath)
	assert.NoError(t, err)
	assert.True(t, empty)
	_, err = fs.Unzip(srcPath, destPath)
	assert.NoError(t, err)
	empty, err = fs.IsEmpty(destPath)
	assert.NoError(t, err)
	assert.False(t, empty)

	err = fs.CleanDirWithContext(context.Background(), destPath)
	require.NoError(t, err)
	empty, err = fs.IsEmpty(destPath)
	assert.NoError(t, err)
	assert.True(t, empty)

	contextWithTimeOut, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancel()
	_, err = fs.UnzipWithContext(contextWithTimeOut, srcPath, destPath)
	assert.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrTimeout)

	err = fs.CleanDirWithContext(context.Background(), destPath)
	require.NoError(t, err)
	empty, err = fs.IsEmpty(destPath)
	assert.NoError(t, err)
	assert.True(t, empty)

	_, err = fs.UnzipWithContextAndLimits(context.Background(), srcPath, destPath, limits)
	assert.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrTooLarge)
}

func TestUnzip_ZipBomb(t *testing.T) {
	// See description of ZIP bombs https://en.wikipedia.org/wiki/Zip_bomb
	// Until protection is part of Go https://github.com/golang/go/issues/33026 & https://github.com/golang/go/issues/33036
	tests := []struct {
		testFile string
	}{
		{
			testFile: "42", // See https://unforgettable.dk/
		},
		{
			testFile: "zbsm", // See https://www.bamsoftware.com/hacks/zipbomb/
		},
		{
			testFile: "zip-bomb-nested-large", // lots of large files
		},
		{
			testFile: "zip-bomb-nested-small", // empty file zipped and that zip also zipped (repeated a few times)
		},
		{
			testFile: "zip-bomb", // 4.5 exabytes with 10 layers
		},
	}

	fs := NewFs(StandardFS)
	testInDir := "testdata"
	destPath, err := fs.TempDirInTempDir("unzip-limits-")
	require.NoError(t, err)
	defer func() { _ = fs.Rm(destPath) }()
	limits := NewLimits(int64(size.GiB), uint64(size.MiB), multiplication.Mega, 3, true) // Total size limited to 1 Mb with a max nesting depth of 3

	empty, err := fs.IsEmpty(destPath)
	assert.NoError(t, err)
	assert.True(t, empty)

	for i := range tests {
		test := tests[i]
		t.Run(test.testFile, func(t *testing.T) {
			srcPath := FilePathJoin(fs, testInDir, test.testFile+".zip")

			_, err = fs.UnzipWithContextAndLimits(context.Background(), srcPath, destPath, limits)
			assert.Error(t, err)
			errortest.AssertError(t, err, commonerrors.ErrUnsupported, commonerrors.ErrTooLarge)
		})
	}

}

func TestUnzip(t *testing.T) {
	fs := NewFs(StandardFS)

	testInDir := "testdata"
	testFile := "testunzip"
	srcPath := FilePathJoin(fs, testInDir, testFile+".zip")
	destPath, err := fs.TempDirInTempDir("unzip")
	require.NoError(t, err)
	defer func() { _ = fs.Rm(destPath) }()
	outPath := FilePathJoin(fs, destPath, testFile)
	expectedfileList := []string{
		FilePathJoin(fs, outPath, "readme.txt"),
		FilePathJoin(fs, outPath, "test.txt"),
		FilePathJoin(fs, outPath, "todo.txt"),
		FilePathJoin(fs, outPath, "child.zip"),
		FilePathJoin(fs, outPath, "L'irrǸsolution est toujours une marque de faiblesse.txt"),
		FilePathJoin(fs, outPath, "ไป ไหน มา.txt"),
	}
	sort.Strings(expectedfileList)

	/* Check Unzip */
	fileList, err := fs.Unzip(srcPath, destPath)

	sort.Strings(fileList)
	assert.NoError(t, err)
	assert.Equal(t, len(fileList), len(expectedfileList))
	assert.Equal(t, expectedfileList, fileList)

}

func TestUnzip_NonRecursive(t *testing.T) {
	fs := NewFs(StandardFS)

	testInDir := "testdata"
	destPath, err := fs.TempDirInTempDir("test-unzip-recursive-")
	require.NoError(t, err)
	defer func() { _ = fs.Rm(destPath) }()

	tests := []struct {
		zipFile           string
		expectedfileList  []string
		expectedTopFolder bool
	}{
		{
			zipFile: FilePathJoin(fs, testInDir, "testunzip.zip"),
			expectedfileList: []string{
				"readme.txt",
				"test.txt",
				"todo.txt",
				"L'irrǸsolution est toujours une marque de faiblesse.txt",
				"ไป ไหน มา.txt",
				"child.zip",
			},
			expectedTopFolder: true,
		},
		{
			zipFile: FilePathJoin(fs, testInDir, "testunzip2.zip"),
			expectedfileList: []string{
				"test1.txt",
				"test2.txt",
				"testunzip.zip",
				"child.zip",
			},
			expectedTopFolder: false,
		},
		{
			zipFile: FilePathJoin(fs, testInDir, "testunzip3.zip"),
			expectedfileList: []string{
				"testunzip2.7z",
				"testunzip2.zip",
				"testunzip.zip",
				"test1.txt",
				"test2.txt",
			},
			expectedTopFolder: false,
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(fmt.Sprintf("#%v %v", i, FilepathStem(test.zipFile)), func(t *testing.T) {
			var outPath string
			if test.expectedTopFolder {
				outPath = FilePathJoin(fs, destPath, FilepathStem(test.zipFile))
			} else {
				outPath = destPath
			}
			expectedfileList, err := fs.ConvertToAbsolutePath(outPath, test.expectedfileList...)
			require.NoError(t, err)
			sort.Strings(expectedfileList)

			/* Check Unzip */
			fileList, err := fs.UnzipWithContextAndLimits(context.TODO(), test.zipFile, destPath, DefaultNonRecursiveZipLimits())
			require.NoError(t, err)
			sort.Strings(fileList)
			assert.Equal(t, len(fileList), len(expectedfileList))
			assert.Equal(t, expectedfileList, fileList)
		})
	}
}

func TestUnzip_Recursive(t *testing.T) {
	fs := NewFs(StandardFS)

	testInDir := "testdata"
	destPath, err := fs.TempDirInTempDir("test-unzip-recursive-")
	require.NoError(t, err)
	defer func() { _ = fs.Rm(destPath) }()

	tests := []struct {
		zipFile           string
		expectedfileList  []string
		expectedTopFolder bool
		expectedError     error
	}{
		{
			zipFile: FilePathJoin(fs, testInDir, "testunzip.zip"),
			expectedfileList: []string{
				"readme.txt",
				"test.txt",
				"todo.txt",
				"L'irrǸsolution est toujours une marque de faiblesse.txt",
				"ไป ไหน มา.txt",
				FilePathJoin(fs, "child", "readme.txt"),
				FilePathJoin(fs, "child", "test.txt"),
				FilePathJoin(fs, "child", "todo.txt"),
			},
			expectedTopFolder: true,
		},
		{
			zipFile: FilePathJoin(fs, testInDir, "testunzip2.zip"),
			expectedfileList: []string{
				"test1.txt",
				"test2.txt",
				FilePathJoin(fs, "child", "readme.txt"),
				FilePathJoin(fs, "child", "test.txt"),
				FilePathJoin(fs, "child", "todo.txt"),
				FilePathJoin(fs, "testunzip", "testunzip", "readme.txt"),
				FilePathJoin(fs, "testunzip", "testunzip", "test.txt"),
				FilePathJoin(fs, "testunzip", "testunzip", "todo.txt"),
				FilePathJoin(fs, "testunzip", "testunzip", "L'irrǸsolution est toujours une marque de faiblesse.txt"),
				FilePathJoin(fs, "testunzip", "testunzip", "ไป ไหน มา.txt"),
				FilePathJoin(fs, "testunzip", "testunzip", "child", "readme.txt"),
				FilePathJoin(fs, "testunzip", "testunzip", "child", "test.txt"),
				FilePathJoin(fs, "testunzip", "testunzip", "child", "todo.txt"),
			},
			expectedTopFolder: false,
		},
		{
			zipFile: FilePathJoin(fs, testInDir, "testunzip3.zip"),
			expectedfileList: []string{
				"test1.txt",
				"test2.txt",
				"testunzip2.7z",
				FilePathJoin(fs, "testunzip2", "testunzip2"),
				FilePathJoin(fs, "testunzip2", "testunzip2", "test1.txt"),
				FilePathJoin(fs, "testunzip2", "testunzip2", "test2.txt"),
				FilePathJoin(fs, "testunzip2", "testunzip2", "testunzip", "testunzip", "L'irrǸsolution est toujours une marque de faiblesse.txt"),
				FilePathJoin(fs, "testunzip2", "testunzip2", "testunzip", "testunzip", "child", "readme.txt"),
				FilePathJoin(fs, "testunzip2", "testunzip2", "testunzip", "testunzip", "child", "test.txt"),
				FilePathJoin(fs, "testunzip2", "testunzip2", "testunzip", "testunzip", "child", "todo.txt"),
				FilePathJoin(fs, "testunzip2", "testunzip2", "testunzip", "testunzip", "readme.txt"),
				FilePathJoin(fs, "testunzip2", "testunzip2", "testunzip", "testunzip", "test.txt"),
				FilePathJoin(fs, "testunzip2", "testunzip2", "testunzip", "testunzip", "todo.txt"),
				FilePathJoin(fs, "testunzip2", "testunzip2", "testunzip", "testunzip", "ไป ไหน มา.txt"),
				FilePathJoin(fs, "testunzip", "testunzip", "L'irrǸsolution est toujours une marque de faiblesse.txt"),
				FilePathJoin(fs, "testunzip", "testunzip", "child", "readme.txt"),
				FilePathJoin(fs, "testunzip", "testunzip", "child", "test.txt"),
				FilePathJoin(fs, "testunzip", "testunzip", "child", "todo.txt"),
				FilePathJoin(fs, "testunzip", "testunzip", "readme.txt"),
				FilePathJoin(fs, "testunzip", "testunzip", "test.txt"),
				FilePathJoin(fs, "testunzip", "testunzip", "todo.txt"),
				FilePathJoin(fs, "testunzip", "testunzip", "ไป ไหน มา.txt")},
			expectedTopFolder: false,
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(fmt.Sprintf("#%v %v", i, FilepathStem(test.zipFile)), func(t *testing.T) {
			var outPath string
			if test.expectedTopFolder {
				outPath = FilePathJoin(fs, destPath, FilepathStem(test.zipFile))
			} else {
				outPath = destPath
			}
			expectedfileList, err := fs.ConvertToAbsolutePath(outPath, test.expectedfileList...)
			require.NoError(t, err)
			sort.Strings(expectedfileList)

			/* Check Unzip */
			fileList, err := fs.UnzipWithContextAndLimits(context.TODO(), test.zipFile, destPath, DefaultZipLimits())
			if test.expectedError == nil {
				require.NoError(t, err)
				sort.Strings(fileList)
				assert.Equal(t, len(fileList), len(expectedfileList))
				assert.Equal(t, expectedfileList, fileList)
			} else {
				require.Error(t, err)
				errortest.AssertError(t, err, test.expectedError)
			}
			require.NoError(t, fs.CleanDir(destPath))
		})
	}
}

func TestUnzip_Failures(t *testing.T) {
	fs := NewFs(StandardFS)

	testInDir := "testdata"
	destPath, err := fs.TempDirInTempDir("unzip")
	require.NoError(t, err)
	defer func() { _ = fs.Rm(destPath) }()

	tests := []struct {
		zipFile       string
		expectedError error
	}{
		{
			zipFile:       FilePathJoin(fs, testInDir, "unknownfile.zip"),
			expectedError: commonerrors.ErrNotFound,
		},
		{
			zipFile:       FilePathJoin(fs, testInDir, "invalidzipfile.zip"),
			expectedError: commonerrors.ErrInvalid,
		},
		{
			zipFile:       FilePathJoin(fs, testInDir, "testunzip2.7z"),
			expectedError: commonerrors.ErrInvalid,
		},
		{
			zipFile:       FilePathJoin(fs, testInDir, "5MB.zip"),
			expectedError: commonerrors.ErrInvalid,
		},
		{
			zipFile:       FilePathJoin(fs, testInDir, "1KB.gz"),
			expectedError: commonerrors.ErrInvalid,
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(fmt.Sprintf("#%v", i), func(t *testing.T) {
			_, err = fs.Unzip(test.zipFile, destPath)
			require.Error(t, err)
			errortest.AssertError(t, err, test.expectedError)
		})
	}
}

func TestUnzip_DepthLimit(t *testing.T) {
	fs := NewFs(StandardFS)

	testInDir := "testdata"
	destPath, err := fs.TempDirInTempDir("test-unzip-depth-limit-")
	require.NoError(t, err)
	defer func() { _ = fs.Rm(destPath) }()

	tests := []struct {
		zipFile       string
		expectedDepth int64
		expectedError error
	}{
		{
			zipFile:       FilePathJoin(fs, testInDir, "testunzip.zip"),
			expectedDepth: 2,
		},
		{
			zipFile:       FilePathJoin(fs, testInDir, "testunzip2.zip"),
			expectedDepth: 3,
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(fmt.Sprintf("#%v %v", i, FilepathStem(test.zipFile)), func(t *testing.T) {

			_, err := fs.UnzipWithContextAndLimits(context.TODO(), test.zipFile, destPath, DefaultZipLimits())
			assert.NoError(t, err)
			require.NoError(t, fs.CleanDir(destPath))
			_, err = fs.UnzipWithContextAndLimits(context.TODO(), test.zipFile, destPath, RecursiveZipLimits(test.expectedDepth))
			assert.NoError(t, err)
			require.NoError(t, fs.CleanDir(destPath))
			_, err = fs.UnzipWithContextAndLimits(context.TODO(), test.zipFile, destPath, RecursiveZipLimits(test.expectedDepth-1))
			assert.Error(t, err)
			errortest.AssertError(t, err, commonerrors.ErrTooLarge)
			require.NoError(t, fs.CleanDir(destPath))
			_, err = fs.UnzipWithContextAndLimits(context.TODO(), test.zipFile, destPath, RecursiveZipLimits(0))
			assert.Error(t, err)
			errortest.AssertError(t, err, commonerrors.ErrTooLarge)
			require.NoError(t, fs.CleanDir(destPath))
		})
	}
}

func TestUnzipWithNonUTF8Filenames(t *testing.T) {
	fs := NewFs(StandardFS)
	// Testing zip file attached to https://github.com/golang/go/issues/10741
	testInDir := "testdata"
	tests := []struct {
		zipFile       string
		expectedFiles []string
		expectedError error
	}{
		{
			zipFile: "zipwithnonutf8.zip",
			expectedFiles: []string{
				"La douceur du miel ne console pas de la piq�re de l'abeille.txt",
				"\x83T\x83\x93\x83v\x83\x8b.txt",
			},
			expectedError: nil,
		},
		{
			zipFile: "zipwithnonutf8filenames2.zip",
			expectedFiles: []string{"examples",
				FilePathJoin(fs, "examples", "AN-32013 FT32F0XX\xb2\xce\xca\xfd.pdf"),
				FilePathJoin(fs, "examples", "BAT32G133_Packʹ\xd3\xc3˵\xc3\xf7.pdf"),
				FilePathJoin(fs, "examples", "OpenAtomFoundation_TencentOS-tiny_ \xcc\xdaѶ\xce\xef\xc1\xaa\xcd\xf8\xd6ն˲\xd9\xd7\xf7ϵͳ.html"),
				FilePathJoin(fs, "examples", "hello_world.c"),
				FilePathJoin(fs, "examples", "main.c"),
			},
			expectedError: nil,
		},
		// TODO create a zip file with non supported encoding
		//
		//	{
		//		zipFile:       ,
		//		expectedFiles: nil,
		//		expectedError: commonerrors.ErrInvalid,
		//	},
	}
	for i := range tests {
		test := tests[i]
		t.Run(fmt.Sprintf("zipfile_%v", test.zipFile), func(t *testing.T) {
			srcPath := FilePathJoin(fs, testInDir, test.zipFile)
			destPath, err := fs.TempDirInTempDir("unzip")
			require.NoError(t, err)
			defer func() { _ = fs.Rm(destPath) }()
			/* Check Unzip */
			fileList, err := fs.Unzip(srcPath, destPath)
			if test.expectedError != nil {
				require.Error(t, err)
				errortest.AssertError(t, err, test.expectedError)
				assert.Empty(t, fileList)
			} else {
				require.NoError(t, err)
				sort.Strings(fileList)
				var expectedFiles []string
				for j := range test.expectedFiles {
					expectedFiles = append(expectedFiles, FilePathJoin(fs, destPath, test.expectedFiles[j]))
				}
				sort.Strings(expectedFiles)
				assert.NoError(t, err)

				assert.Equal(t, len(fileList), len(test.expectedFiles))
				assert.Equal(t, expectedFiles, fileList)
			}
			_ = fs.Rm(destPath)
		})
	}

}

func TestUnzipFileCountLimit(t *testing.T) {
	fs := NewFs(StandardFS)

	testInDir := "testdata"
	limits := NewLimits(int64(size.GiB), uint64(10*size.GiB), multiplication.Deka, multiplication.Deka, true)

	t.Run("unzip file above file count limit", func(t *testing.T) {
		testFile := "abovefilecountlimitzip"
		srcPath := FilePathJoin(fs, testInDir, testFile+".zip")

		destPath, err := fs.TempDirInTempDir("unzip-limits-")
		assert.NoError(t, err)
		defer func() {
			_ = fs.Rm(destPath)
		}()

		_, err = fs.UnzipWithContextAndLimits(context.TODO(), srcPath, destPath, limits)
		errortest.AssertError(t, err, commonerrors.ErrTooLarge)
	})

	t.Run("unzip file below file count limit", func(t *testing.T) {
		testFile := "belowfilecountlimitzip"
		srcPath := FilePathJoin(fs, testInDir, testFile+".zip")

		destPath, err := fs.TempDirInTempDir("unzip-limits-")
		assert.NoError(t, err)

		defer func() {
			if tempErr := fs.Rm(destPath); tempErr != nil {
				err = tempErr
			}
		}()

		_, err = fs.UnzipWithContextAndLimits(context.TODO(), srcPath, destPath, limits)
		assert.NoError(t, err)
	})
}

func testSanitiseZipExtractPath(t *testing.T, filePath string) {
	fs := NewStandardFileSystem()
	dst := FilePathJoin(fs, faker.Word(), faker.Name(), faker.UUIDHyphenated(), faker.Name())
	rootDepth, err := FileTreeDepth(fs, "", dst)
	require.NoError(t, err)
	dest, subErr := sanitiseZipExtractPath(fs, filePath, dst)
	require.NotEmpty(t, dest)
	destDepth, err := FileTreeDepth(fs, "", dest)
	if rootDepth > destDepth || !strings.Contains(dest, dst) {
		errortest.RequireError(t, subErr, commonerrors.ErrMalicious)
	} else {
		require.NoError(t, err)
	}
}

func FuzzSanitiseZipExtractPath(f *testing.F) {
	f.Add("..")
	f.Fuzz(testSanitiseZipExtractPath)
}

func TestZipBetweenFilesystemWithDifferentFilePathSeparators(t *testing.T) {
	fs1 := NewTestOSFilesystem(t, '\\')
	fs2 := NewTestOSFilesystem(t, '/')
	ctx := context.Background()
	tests := []struct {
		fsSrc  FS
		fsDest FS
	}{
		{fs1, fs1},
		{fs2, fs2},
		{fs1, fs2},
		{fs2, fs1},
	}
	for k := range tests {
		fsSrc := tests[k].fsSrc
		fsDest := tests[k].fsDest
		t.Run(fmt.Sprintf("%v_from_fs_with_sep_%v_to_fs_with_sep_%v", t.Name(), string(fsSrc.PathSeparator()), string(fsDest.PathSeparator())), func(t *testing.T) {
			for i := 0; i < 10; i++ {
				func() {
					// create a directory for the test
					tmpDirFs1, err := fsSrc.TempDirInTempDir("zip")
					require.NoError(t, err)
					defer func() { _ = fsSrc.Rm(tmpDirFs1) }()
					tmpDirFs2, err := fsDest.TempDirInTempDir("unzip")
					require.NoError(t, err)
					defer func() { _ = fsDest.Rm(tmpDirFs2) }()

					testDirSrc := FilePathJoin(fsSrc, tmpDirFs1, "test")
					srcZipfile := FilePathJoin(fsSrc, testDirSrc, "test.zip")
					testDirDest := FilePathJoin(fsDest, tmpDirFs2, "test")
					outDir := FilePathJoin(fsDest, testDirDest, "output")
					destZipFile := FilePathJoin(fsDest, testDirDest, "test.zip")

					tree := GenerateTestFileTree(t, fsSrc, testDirSrc, "", false, time.Now(), time.Now())
					tree = append(tree, srcZipfile)
					// zip the directory into the zipfile
					err = fsSrc.Zip(testDirSrc, srcZipfile)
					require.NoError(t, err)
					require.True(t, fsSrc.Exists(srcZipfile))

					require.NoError(t, CopyBetweenFS(ctx, fsSrc, srcZipfile, fsDest, destZipFile))

					// unzip
					tree2, err := fsDest.Unzip(destZipFile, outDir)
					require.NoError(t, err)

					// Check no files were lost in the zip/unzip process.
					relativeSrcTree, err := fsSrc.ConvertToRelativePath(testDirSrc, tree...)
					require.NoError(t, err)
					relativeSrcTree = FilePathsToSlash(fsSrc, relativeSrcTree...)
					relativeResultTree, err := fsDest.ConvertToRelativePath(outDir, tree2...)
					require.NoError(t, err)
					relativeResultTree = FilePathsToSlash(fsDest, relativeResultTree...)
					sort.Strings(relativeSrcTree)
					sort.Strings(relativeResultTree)
					require.Equal(t, relativeSrcTree, relativeResultTree)

					hasher, err := NewFileHash(hashing.HashXXHash)
					require.NoError(t, err)

					// check for size, timestamp, hash preservation
					for _, path := range relativeSrcTree {
						if FilePathExt(fsSrc, FilePathFromSlash(fsSrc, path)) == ".zip" {
							continue
						}
						srcFilePath := FilePathJoin(fsSrc, testDirSrc, FilePathFromSlash(fsSrc, path))
						require.True(t, fsDest.Exists(srcFilePath))
						fileinfoSrc, err := fsSrc.Lstat(srcFilePath)
						require.NoError(t, err)
						resultFilePath := FilePathJoin(fsDest, outDir, FilePathFromSlash(fsDest, path))
						require.True(t, fsDest.Exists(resultFilePath))
						fileinfoResult, err := fsDest.Lstat(resultFilePath)
						require.NoError(t, err)
						// TODO handle links separately
						if IsSymLink(fileinfoSrc) {
							continue
						}
						// Check sizes
						assert.Equal(t, fileinfoSrc.Size(), fileinfoResult.Size())

						// perform hash comparison
						if IsRegularFile(fileinfoSrc) {
							hashSrc, err := hasher.CalculateFile(fsSrc, srcFilePath)
							require.NoError(t, err)
							hashResult, err := hasher.CalculateFile(fsDest, resultFilePath)
							require.NoError(t, err)
							assert.Equal(t, hashSrc, hashResult)
						}
					}
					err = fsSrc.Rm(tmpDirFs1)
					require.NoError(t, err)
					err = fsDest.Rm(tmpDirFs2)
					require.NoError(t, err)
				}()
			}
		})
	}

}
