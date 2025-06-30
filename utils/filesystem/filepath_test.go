package filesystem

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/platform"
)

func TestFilePathFromTo(t *testing.T) {
	t.Run("given a path, toSlash works as filepath", func(t *testing.T) {
		fp := filepath.Join("super", "foo", "bar.baz")
		assert.Equal(t, filepath.ToSlash(fp), FilePathToSlash(GetGlobalFileSystem(), fp))
		assert.Equal(t, filepath.ToSlash(fp), FilePathToSlash(NewTestFilesystem(t, '/'), filepath.ToSlash(fp)))
	})

	t.Run("given a path with slashes, it is correctly converted", func(t *testing.T) {
		fp := path.Join("super", "foo", "bar.baz")
		assert.Equal(t, filepath.FromSlash(fp), FilePathFromSlash(GetGlobalFileSystem(), fp))
		assert.Equal(t, fp, FilePathFromSlash(NewTestFilesystem(t, '/'), fp))
	})

	t.Run("given a path, FilePathToPlatformPathSeparator converts to platform path", func(t *testing.T) {
		fp := filepath.Join("super", "foo", "bar.baz")
		assert.Equal(t, fp, FilePathToPlatformPathSeparator(NewTestFilesystem(t, '/'), filepath.ToSlash(fp)))
		fp2 := strings.ReplaceAll(filepath.ToSlash(fp), "/", `\`)
		assert.Equal(t, fp, FilePathToPlatformPathSeparator(NewTestFilesystem(t, '\\'), fp2))
		assert.Equal(t, fp2, FilePathToPlatformPathSeparator(NewTestFilesystem(t, '/'), fp2))
	})

	t.Run("given a path with slashes, it is correctly converted to filesystem separator", func(t *testing.T) {
		fp := filepath.Join("super", "foo", "bar.baz")
		assert.Equal(t, fp, FilePathFromPlatformPathSeparator(NewTestFilesystem(t, platform.PathSeparator), fp))
		fp2 := path.Join("super", "foo", "bar.baz")
		assert.Equal(t, fp2, FilePathFromPlatformPathSeparator(NewTestFilesystem(t, '/'), fp))
		assert.Equal(t, filepath.ToSlash(fp2), FilePathFromPlatformPathSeparator(NewTestFilesystem(t, '/'), fp2))
	})

}

func TestFilepathStem(t *testing.T) {
	t.Run("given a filename with extension, it strips extension", func(t *testing.T) {
		assert.Equal(t, "foo", FilepathStem("foo.bar"))
		assert.Equal(t, "foo", FilePathStemOnFilesystem(NewTestFilesystem(t, platform.PathSeparator), "foo.bar"))
		assert.Equal(t, "library.tar", FilepathStem("library.tar.gz"))
		assert.Equal(t, "library.tar", FilePathStemOnFilesystem(NewTestFilesystem(t, '/'), "library.tar.gz"))
		assert.Equal(t, "cool", FilepathStem("cool"))
		assert.Equal(t, "cool", FilePathStemOnFilesystem(NewTestFilesystem(t, '\\'), "cool"))
	})

	t.Run("given a filepath, it returns only file name", func(t *testing.T) {
		fp := filepath.Join("super", "foo", "bar.baz")
		assert.Equal(t, "bar", FilepathStem(fp))
		fp = path.Join("super", "foo", "bar.baz")
		assert.Equal(t, "bar", FilePathStemOnFilesystem(NewTestFilesystem(t, '/'), fp))
		fp = filepath.Join("nice", "file", "path")
		assert.Equal(t, "path", FilepathStem(fp))
		fp = path.Join("nice", "file", "path")
		assert.Equal(t, "path", FilePathStemOnFilesystem(NewTestFilesystem(t, '/'), fp))
	})
	assert.Empty(t, FilePathStemOnFilesystem(nil, faker.Sentence()))

}

func TestFilepathParents(t *testing.T) {
	type PathTest struct {
		path            string
		expectedParents []string
	}
	tests := []PathTest{
		{},
		{
			path:            "                             ",
			expectedParents: nil,
		},
		{
			path:            "/",
			expectedParents: nil,
		},
		{
			path:            ".",
			expectedParents: nil,
		},
		{
			path:            "./",
			expectedParents: nil,
		},
		{
			path:            "./blah",
			expectedParents: nil,
		},
		{
			path:            filepath.Join("a", "great", "fake", "path", "blah"),
			expectedParents: []string{"a", filepath.Join("a", "great"), filepath.Join("a", "great", "fake"), filepath.Join("a", "great", "fake", "path")},
		},
		{
			path:            "/foo/bar/setup.py",
			expectedParents: []string{`foo`, filepath.Join(`foo`, `bar`)},
		},
	}

	if platform.IsWindows() {
		tests = append(tests, PathTest{
			path:            "C:/foo/bar/setup.py",
			expectedParents: []string{"C:", filepath.Join(`C:`, `\foo`), filepath.Join(`C:`, `\foo`, `bar`)},
		})
	} else {
		tests = append(tests, PathTest{
			path:            "C:/foo/bar/setup.py",
			expectedParents: []string{"C:", filepath.Join(`C:`, `foo`), filepath.Join(`C:`, `foo`, `bar`)},
		})
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			assert.ElementsMatch(t, tt.expectedParents, FilepathParents(tt.path))
		})
	}
}

func TestFileTreeDepth(t *testing.T) {
	random := fmt.Sprintf("%v %v %v", faker.Name(), faker.Name(), faker.Name())
	complexRandom := fmt.Sprintf("%v&#~@£*-_()^+!%v %v", faker.Name(), faker.Name(), faker.Name())
	tests := []struct {
		root          string
		file          string
		expectedDepth int64
	}{
		{},
		{
			root:          faker.Word(),
			file:          "",
			expectedDepth: 0,
		},
		{
			root:          "",
			file:          fmt.Sprintf(".%v%v", string(PathSeparator()), random),
			expectedDepth: 0,
		},
		{
			root:          "",
			file:          fmt.Sprintf(".%v%v", string(PathSeparator()), complexRandom),
			expectedDepth: 0,
		},
		{
			root:          "",
			file:          fmt.Sprintf(".%v%v/%v", string(PathSeparator()), random, random),
			expectedDepth: 1,
		},
		{
			root:          "",
			file:          fmt.Sprintf(".%v%v%v%v", string(PathSeparator()), random, string(PathSeparator()), random),
			expectedDepth: 1,
		},
		{
			root:          fmt.Sprintf("./%v", random),
			file:          fmt.Sprintf("./%v/%v", random, complexRandom),
			expectedDepth: 0,
		},
		{
			root:          fmt.Sprintf("./%v", complexRandom),
			file:          fmt.Sprintf("./%v/%v", random, complexRandom),
			expectedDepth: 2,
		},
		{
			root:          fmt.Sprintf("./%v", complexRandom),
			file:          fmt.Sprintf("./%v/%v", complexRandom, random),
			expectedDepth: 0,
		},
		{
			root:          fmt.Sprintf("./%v", complexRandom),
			file:          fmt.Sprintf("./%v/%v/%v/%v/%v/%v/%v", complexRandom, random, random, random, random, random, random),
			expectedDepth: 5,
		},
		{
			root:          fmt.Sprintf(".%v%v", string(PathSeparator()), complexRandom),
			file:          fmt.Sprintf(".%v%v%v%v%v%v%v%v", string(PathSeparator()), complexRandom, string(PathSeparator()), random, string(PathSeparator()), random, string(PathSeparator()), random),
			expectedDepth: 2,
		},
	}

	for fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(FileSystemTypes[fsType])
			for i := range tests {
				test := tests[i]
				t.Run(fmt.Sprintf("#%v %v", i, FilepathStem(test.file)), func(t *testing.T) {
					depth, err := FileTreeDepth(fs, test.root, test.file)
					require.NoError(t, err)
					assert.Equal(t, test.expectedDepth, depth)
				})
			}
		})
	}
}

func TestEndsWithPathSeparator(t *testing.T) {
	for fsType := range FileSystemTypes {
		t.Run(fmt.Sprintf("%v_for_fs_%v", t.Name(), fsType), func(t *testing.T) {
			fs := NewFs(FileSystemTypes[fsType])

			assert.True(t, EndsWithPathSeparator(fs, "test fsdfs .fsdffs /"))
			assert.False(t, EndsWithPathSeparator(fs, "test fsdfs .fsdffs "))
			assert.False(t, EndsWithPathSeparator(fs, filepath.Join(faker.DomainName(), "test fsdfs .fsdffs ")))
			assert.True(t, EndsWithPathSeparator(fs, filepath.Join(faker.DomainName(), "test fsdfs .fsdffs ")+"/"))
			assert.False(t, EndsWithPathSeparator(fs, filepath.Join(faker.DomainName(), "test fsdfs .fsdffs /")), "join should trim the tailing separator")
			assert.True(t, EndsWithPathSeparator(fs, "test fsdfs .fsdffs "+string(fs.PathSeparator())))
			assert.True(t, EndsWithPathSeparator(fs, filepath.Join(faker.DomainName(), "test fsdfs .fsdffs ")+string(fs.PathSeparator())))
			assert.False(t, EndsWithPathSeparator(fs, filepath.Join(faker.DomainName(), "test fsdfs .fsdffs "+string(fs.PathSeparator()))), "join should trim the tailing separator")
		})
	}
}

func TestNewPathExistRule(t *testing.T) {
	t.Run("disable", func(t *testing.T) {
		err := NewOSPathExistRule(false).Validate(faker.URL())
		require.NoError(t, err)
	})
	t.Run("happy existing path", func(t *testing.T) {
		require.NoError(t, NewOSPathExistRule(true).Validate(TempDirectory()))
		testDir, err := TempDirInTempDir("test-path-rule-")
		require.NoError(t, err)
		defer func() { _ = Rm(testDir) }()
		require.NoError(t, NewOSPathExistRule(true).Validate(testDir))
		testFile, err := TouchTempFile(testDir, "test-file*.test")
		require.NoError(t, err)
		require.NoError(t, NewOSPathExistRule(true).Validate(testFile))
	})
	t.Run("non-existent path but valid", func(t *testing.T) {
		err := NewOSPathExistRule(true).Validate(strings.ReplaceAll(faker.Sentence(), " ", "/"))
		require.Error(t, err)
		errortest.AssertError(t, err, commonerrors.ErrNotFound)
		err = NewOSPathValidationRule(true).Validate(strings.ReplaceAll(faker.Sentence(), " ", "/"))
		require.NoError(t, err)
		err = NewOSPathExistRule(true).Validate(faker.URL())
		require.Error(t, err)
		errortest.AssertError(t, err, commonerrors.ErrNotFound)
		err = NewOSPathValidationRule(true).Validate(faker.URL())
		require.NoError(t, err)
	})

	t.Run("invalid paths", func(t *testing.T) {
		tests := []struct {
			entry         any
			expectedError []error
		}{
			{
				entry:         nil,
				expectedError: []error{commonerrors.ErrUndefined, commonerrors.ErrInvalid},
			},
			{
				entry:         "                  ",
				expectedError: []error{commonerrors.ErrUndefined, commonerrors.ErrInvalid},
			},
			{
				entry:         123,
				expectedError: []error{commonerrors.ErrInvalid},
			},
			{
				entry:         fmt.Sprintf("%v\n%v\n%v", faker.Paragraph(), faker.Paragraph(), faker.Sentence()),
				expectedError: []error{commonerrors.ErrInvalid},
			},
		}
		for i := range tests {
			test := tests[i]
			t.Run(fmt.Sprintf("%v", test.entry), func(t *testing.T) {
				err := NewOSPathValidationRule(true).Validate(test.entry)
				require.Error(t, err)
				errortest.AssertError(t, err, test.expectedError...)
				err = NewOSPathExistRule(true).Validate(test.entry)
				require.Error(t, err)
				errortest.AssertError(t, err, test.expectedError...)
			})
		}

	})

}

func TestFilePathJoin(t *testing.T) {
	embedFS, err := NewEmbedFileSystem(&testContent)
	require.NoError(t, err)
	tests := []struct {
		fs              FS
		elements        []string
		expectedWindows string
		expectedLinux   string
	}{
		{
			fs:              nil,
			elements:        []string{"test1", "test2", "..", "c"},
			expectedWindows: "",
			expectedLinux:   "",
		},
		{
			fs:              NewStandardFileSystem(),
			elements:        []string{"test1", "c"},
			expectedWindows: "test1\\c",
			expectedLinux:   "test1/c",
		},
		{
			fs:              NewStandardFileSystem(),
			elements:        []string{},
			expectedWindows: "",
			expectedLinux:   "",
		},
		{
			fs:              embedFS,
			elements:        []string{"test1", "c"},
			expectedWindows: "test1/c",
			expectedLinux:   "test1/c",
		},
		{
			fs:              embedFS,
			elements:        []string{""},
			expectedWindows: "",
			expectedLinux:   "",
		},
		{
			fs:              NewStandardFileSystem(),
			elements:        []string{""},
			expectedWindows: "",
			expectedLinux:   "",
		},
		{
			fs:              NewStandardFileSystem(),
			elements:        []string{"a"},
			expectedWindows: "a",
			expectedLinux:   "a",
		},
		{
			fs:              NewStandardFileSystem(),
			elements:        []string{"a", ""},
			expectedWindows: "a",
			expectedLinux:   "a",
		},
		{
			fs:              embedFS,
			elements:        []string{"a", ""},
			expectedWindows: "a",
			expectedLinux:   "a",
		},
		{
			fs:              NewStandardFileSystem(),
			elements:        []string{"", "a"},
			expectedWindows: "a",
			expectedLinux:   "a",
		},
		{
			fs:              embedFS,
			elements:        []string{"", "a"},
			expectedWindows: "a",
			expectedLinux:   "a",
		},
		{
			fs:              NewStandardFileSystem(),
			elements:        []string{"", ""},
			expectedWindows: "",
			expectedLinux:   "",
		},
		{
			fs:              embedFS,
			elements:        []string{"", ""},
			expectedWindows: "",
			expectedLinux:   "",
		},
		{
			fs:              NewTestFilesystem(t, '/'),
			elements:        []string{"test1", "c"},
			expectedWindows: "test1/c",
			expectedLinux:   "test1/c",
		},
		{
			fs:              NewTestFilesystem(t, '\\'),
			elements:        []string{"test1", "c"},
			expectedWindows: `test1\c`,
			expectedLinux:   `test1\c`,
		},
		{
			fs:              NewStandardFileSystem(),
			elements:        []string{"test1", "test2", "..", "c"},
			expectedWindows: `test1\c`,
			expectedLinux:   "test1/c",
		},
		{
			fs:              NewTestFilesystem(t, '/'),
			elements:        []string{"test1", "test2", "..", "c"},
			expectedWindows: "test1/c",
			expectedLinux:   "test1/c",
		},
		{
			fs:              NewTestFilesystem(t, '\\'),
			elements:        []string{"test1", "test2", "..", "c"},
			expectedWindows: `test1\c`,
			expectedLinux:   `test1\c`,
		},
		{
			fs:              NewStandardFileSystem(),
			elements:        []string{"test1", "test2", "..", "\\c"},
			expectedWindows: "test1\\c",
			expectedLinux:   "test1/\\c",
		},
		{
			fs:              NewStandardFileSystem(),
			elements:        []string{"test1", "test2", "..", "/c"},
			expectedWindows: "test1\\c",
			expectedLinux:   "test1/c",
		},
		{
			fs:              embedFS,
			elements:        []string{"test1", "test2", "..", "c"},
			expectedWindows: "test1/c",
			expectedLinux:   "test1/c",
		},
		{
			fs:              embedFS,
			elements:        []string{"test1", "test2", "..", "/c"},
			expectedWindows: "test1/c",
			expectedLinux:   "test1/c",
		},
		{
			fs:              embedFS,
			elements:        []string{"test1", "test2", "..", "\\c"},
			expectedWindows: "test1/\\c",
			expectedLinux:   "test1/\\c",
		},
	}
	for i := range tests {
		test := tests[i]
		fsType := ""
		if test.fs != nil {
			fsType = fmt.Sprintf("%v", test.fs.GetType())
		}
		t.Run(fmt.Sprintf("%v %v %v", fsType, i, test.elements), func(t *testing.T) {
			join := FilePathJoin(test.fs, test.elements...)
			if platform.IsWindows() {
				assert.Equal(t, test.expectedWindows, join)
			} else {
				assert.Equal(t, test.expectedLinux, join)
			}
		})
	}
}

func TestFilePathClean(t *testing.T) {
	tests := []struct {
		fs              FS
		path            string
		expectedWindows string
		expectedLinux   string
	}{
		{
			fs:              nil,
			path:            path.Join("test1", "c"),
			expectedWindows: "",
			expectedLinux:   "",
		},
		{
			fs:              NewStandardFileSystem(),
			path:            ".",
			expectedWindows: ".",
			expectedLinux:   ".",
		},
		{
			fs:              NewTestFilesystem(t, '/'),
			path:            ".",
			expectedWindows: ".",
			expectedLinux:   ".",
		},
		{
			fs:              NewTestFilesystem(t, '/'),
			path:            "a",
			expectedWindows: "a",
			expectedLinux:   "a",
		},
		{
			fs:              NewTestFilesystem(t, '\\'),
			path:            "a",
			expectedWindows: "a",
			expectedLinux:   "a",
		},
		{
			fs:              NewStandardFileSystem(),
			path:            filepath.Join("test1", "c"),
			expectedWindows: "test1\\c",
			expectedLinux:   "test1/c",
		},
		{
			fs:              NewStandardFileSystem(),
			path:            filepath.Join("test1", "test2", "..", "c"),
			expectedWindows: "test1\\c",
			expectedLinux:   "test1/c",
		},
		{
			fs:              NewTestFilesystem(t, '/'),
			path:            path.Join("test1", "c"),
			expectedWindows: "test1/c",
			expectedLinux:   "test1/c",
		},
		{
			fs:              NewTestFilesystem(t, '/'),
			path:            `test1\c`,
			expectedWindows: `test1\c`,
			expectedLinux:   `test1\c`,
		},
		{
			fs:              NewTestFilesystem(t, '\\'),
			path:            `test1\c`,
			expectedWindows: `test1\c`,
			expectedLinux:   `test1\c`,
		},
		{
			fs:              NewTestFilesystem(t, '\\'),
			path:            `test1/c`,
			expectedWindows: `test1\c`,
			expectedLinux:   `test1\c`,
		},
		{
			fs:              NewTestFilesystem(t, '/'),
			path:            "test1/test2/../c",
			expectedWindows: "test1/c",
			expectedLinux:   "test1/c",
		},
		{
			fs:              NewTestFilesystem(t, '/'),
			path:            `test1\test2\..\c`,
			expectedWindows: `test1\test2\..\c`,
			expectedLinux:   `test1\test2\..\c`,
		},
		{
			fs:              NewTestFilesystem(t, '\\'),
			path:            "test1/test2/../c",
			expectedWindows: `test1\c`,
			expectedLinux:   `test1\c`,
		},
		{
			fs:              NewTestFilesystem(t, '\\'),
			path:            `test1\test2\..\c`,
			expectedWindows: `test1\c`,
			expectedLinux:   `test1\c`,
		},
	}
	for i := range tests {
		test := tests[i]
		fsType := ""
		if test.fs != nil {
			fsType = fmt.Sprintf("%v", test.fs.GetType())
		}
		t.Run(fmt.Sprintf("%v %v %v", fsType, i, test.path), func(t *testing.T) {
			cleaned := FilePathClean(test.fs, test.path)
			if platform.IsWindows() {
				assert.Equal(t, test.expectedWindows, cleaned)
			} else {
				assert.Equal(t, test.expectedLinux, cleaned)
			}
		})
	}
}

func TestFilePathBase(t *testing.T) {
	embedFS, err := NewEmbedFileSystem(&testContent)
	require.NoError(t, err)
	tests := []struct {
		fs           FS
		pathLinux    string
		pathWindows  string
		expectedBase string
	}{
		{
			fs:           nil,
			pathLinux:    "test1/test2",
			pathWindows:  "test1\\test2",
			expectedBase: "",
		},
		{
			fs:           NewStandardFileSystem(),
			pathLinux:    "",
			pathWindows:  "",
			expectedBase: ".",
		},
		{
			fs:           embedFS,
			pathLinux:    "",
			pathWindows:  ".",
			expectedBase: ".",
		},
		{
			fs:           NewStandardFileSystem(),
			pathLinux:    "/test1/test2/file.ext",
			pathWindows:  "C:\\test1\\test2\\file.ext",
			expectedBase: "file.ext",
		},
		{
			fs:           NewStandardFileSystem(),
			pathLinux:    "./test1/test2/",
			pathWindows:  ".\\test1\\test2\\",
			expectedBase: "test2",
		},
		{
			fs:           embedFS,
			pathLinux:    "./test1/test2/",
			pathWindows:  "./test1/test2/",
			expectedBase: "test2",
		},
		{
			fs:           embedFS,
			pathLinux:    ".\\test1\\test2\\",
			pathWindows:  ".\\test1\\test2\\",
			expectedBase: ".\\test1\\test2\\",
		},
	}
	for i := range tests {
		test := tests[i]
		fsType := ""
		if test.fs != nil {
			fsType = fmt.Sprintf("%v", test.fs.GetType())
		}

		var testPath string
		if platform.IsWindows() {
			testPath = test.pathWindows
		} else {
			testPath = test.pathLinux
		}

		t.Run(fmt.Sprintf("%v %v", fsType, testPath), func(t *testing.T) {
			assert.Equal(t, test.expectedBase, FilePathBase(test.fs, testPath))
		})
	}
}

func TestFilePathDir(t *testing.T) {
	embedFS, err := NewEmbedFileSystem(&testContent)
	require.NoError(t, err)
	tests := []struct {
		fs          FS
		pathLinux   string
		pathWindows string
		expectedDir string
	}{
		{
			fs:          nil,
			pathLinux:   "test1/test2",
			pathWindows: "test1\\test2",
			expectedDir: "",
		},
		{
			fs:          NewStandardFileSystem(),
			pathLinux:   "",
			pathWindows: "",
			expectedDir: ".",
		},
		{
			fs:          embedFS,
			pathLinux:   "",
			pathWindows: "",
			expectedDir: ".",
		},
		{
			fs:          NewStandardFileSystem(),
			pathLinux:   ".",
			pathWindows: ".",
			expectedDir: ".",
		},
		{
			fs:          embedFS,
			pathLinux:   ".",
			pathWindows: ".",
			expectedDir: ".",
		},
		{
			fs:          NewStandardFileSystem(),
			pathLinux:   "abc",
			pathWindows: "abc",
			expectedDir: ".",
		},
		{
			fs:          embedFS,
			pathLinux:   "abc",
			pathWindows: "abc",
			expectedDir: ".",
		},
		{
			fs:          embedFS,
			pathLinux:   "./.",
			pathWindows: `.\.`,
			expectedDir: ".",
		},
		{
			fs:          embedFS,
			pathLinux:   "./test/test",
			pathWindows: "./test/test",
			expectedDir: "test",
		},
		{
			fs:          embedFS,
			pathLinux:   "./test/////test",
			pathWindows: "./test/////test",
			expectedDir: "test",
		},
		{
			fs:          NewStandardFileSystem(),
			pathLinux:   "./test/test",
			pathWindows: "./test/test",
			expectedDir: "test",
		},
		{
			fs:          NewStandardFileSystem(),
			pathLinux:   "./test////////////test",
			pathWindows: "./test////////////test",
			expectedDir: "test",
		},
		{
			fs:          embedFS,
			pathLinux:   `.\test\test`,
			pathWindows: `.\test\test`,
			expectedDir: ".",
		},
		{
			fs:          NewTestFilesystem(t, '\\'),
			pathLinux:   `.\test`,
			pathWindows: `.\test`,
			expectedDir: ".",
		},
		{
			fs:          NewTestFilesystem(t, '\\'),
			pathLinux:   `.\\\\\\\\test`,
			pathWindows: `.\\\\\\\\test`,
			expectedDir: ".",
		},
	}
	for i := range tests {
		test := tests[i]
		fsType := ""
		if test.fs != nil {
			fsType = fmt.Sprintf("%v", test.fs.GetType())
		}

		var testPath string
		if platform.IsWindows() {
			testPath = test.pathWindows
		} else {
			testPath = test.pathLinux
		}

		t.Run(fmt.Sprintf("%v %v %v", fsType, i, testPath), func(t *testing.T) {
			assert.Equal(t, test.expectedDir, FilePathDir(test.fs, testPath))
		})
	}
}

func TestFilePathIsAbs(t *testing.T) {
	embedFS, err := NewEmbedFileSystem(&testContent)
	require.NoError(t, err)
	tests := []struct {
		fs           FS
		pathLinux    string
		pathWindows  string
		isAbs        bool
		specialLinux bool
	}{
		{
			fs:          nil,
			pathLinux:   "test1/test2",
			pathWindows: "test1\\test2",
			isAbs:       false,
		},
		{
			fs:          NewStandardFileSystem(),
			pathLinux:   "",
			pathWindows: "",
			isAbs:       false,
		},
		{
			fs:          embedFS,
			pathLinux:   "",
			pathWindows: "",
			isAbs:       false,
		},
		{
			fs:          NewStandardFileSystem(),
			pathLinux:   ".",
			pathWindows: ".",
			isAbs:       false,
		},
		{
			fs:          NewStandardFileSystem(),
			pathLinux:   "/",
			pathWindows: `C:\\`,
			isAbs:       true,
		},
		{
			fs:          NewStandardFileSystem(),
			pathLinux:   "/usr/bin/gcc",
			pathWindows: `C:\\usr\bin\gcc`,
			isAbs:       true,
		},
		{
			fs:          NewStandardFileSystem(),
			pathLinux:   "/usr/../bin/gcc",
			pathWindows: `C:\\usr\..\bin\gcc`,
			isAbs:       true,
		},
		{
			fs:          embedFS,
			pathLinux:   ".",
			pathWindows: ".",
			isAbs:       false,
		},
		{
			fs:          embedFS,
			pathLinux:   "/usr/../bin/gcc",
			pathWindows: "/usr/../bin/gcc",
			isAbs:       true,
		},
		{
			fs:           NewTestFilesystem(t, '\\'),
			pathLinux:    "/usr/../bin/gcc",
			pathWindows:  "/usr/../bin/gcc",
			isAbs:        false,
			specialLinux: true,
		},
		{
			fs:          NewStandardFileSystem(),
			pathLinux:   "..",
			pathWindows: "..",
			isAbs:       false,
		},
		{
			fs:          embedFS,
			pathLinux:   "..",
			pathWindows: "..",
			isAbs:       false,
		},
	}
	for i := range tests {
		test := tests[i]
		fsType := ""
		if test.fs != nil {
			fsType = fmt.Sprintf("%v", test.fs.GetType())
		}

		var testPath string
		if platform.IsWindows() {
			testPath = test.pathWindows
		} else {
			testPath = test.pathLinux
		}

		t.Run(fmt.Sprintf("%v %v %v", fsType, i, testPath), func(t *testing.T) {
			if test.specialLinux {
				if platform.IsWindows() {
					assert.Equal(t, test.isAbs, FilePathIsAbs(test.fs, testPath))
				}
			} else {
				assert.Equal(t, test.isAbs, FilePathIsAbs(test.fs, testPath))
			}

		})
	}
}

func TestFilePathExt(t *testing.T) {
	embedFS, err := NewEmbedFileSystem(&testContent)
	require.NoError(t, err)
	tests := []struct {
		fs          FS
		pathLinux   string
		pathWindows string
		ext         string
	}{
		{
			fs:          nil,
			pathLinux:   "test1/test2.test",
			pathWindows: "test1\\test2.test",
			ext:         "",
		},
		{
			fs:          NewStandardFileSystem(),
			pathLinux:   "test1/test2.test",
			pathWindows: "test1\\test2.test",
			ext:         ".test",
		},
		{
			fs:          embedFS,
			pathLinux:   "test1/test2.test",
			pathWindows: "test1/test2.test",
			ext:         ".test",
		},
		{
			fs:          embedFS,
			pathLinux:   `test1\test2.test`,
			pathWindows: `test1\test2.test`,
			ext:         ".test",
		},
		{
			fs:          embedFS,
			pathLinux:   "test1/test2",
			pathWindows: "test1/test2",
			ext:         "",
		},
		{
			fs:          embedFS,
			pathLinux:   `test1\test2`,
			pathWindows: `test1\test2`,
			ext:         "",
		},
	}
	for i := range tests {
		test := tests[i]
		fsType := ""
		if test.fs != nil {
			fsType = fmt.Sprintf("%v", test.fs.GetType())
		}

		var testPath string
		if platform.IsWindows() {
			testPath = test.pathWindows
		} else {
			testPath = test.pathLinux
		}

		t.Run(fmt.Sprintf("%v %v %v", fsType, i, testPath), func(t *testing.T) {
			assert.Equal(t, test.ext, FilePathExt(test.fs, testPath))
		})
	}
}

func TestFilePathSplit(t *testing.T) {
	embedFS, err := NewEmbedFileSystem(&testContent)
	require.NoError(t, err)
	tests := []struct {
		fs          FS
		pathLinux   string
		pathWindows string
		dir         string
		file        string
	}{
		{
			fs:          nil,
			pathLinux:   "test1/test2.test",
			pathWindows: "test1\\test2.test",
			dir:         "",
			file:        "",
		},
		{
			fs:          NewStandardFileSystem(),
			pathLinux:   "test1/test2.test",
			pathWindows: `test1\test2.test`,
			dir:         FilePathToPlatformPathSeparator(NewTestFilesystem(t, '\\'), `test1\`),
			file:        "test2.test",
		},
		{
			fs:          embedFS,
			pathLinux:   "test1/test2.test",
			pathWindows: "test1/test2.test",
			dir:         "test1/",
			file:        "test2.test",
		},
		{
			fs:          embedFS,
			pathLinux:   `test1\test2.test`,
			pathWindows: `test1\test2.test`,
			dir:         "",
			file:        `test1\test2.test`,
		},
	}
	for i := range tests {
		test := tests[i]
		fsType := ""
		if test.fs != nil {
			fsType = fmt.Sprintf("%v", test.fs.GetType())
		}

		var testPath string
		if platform.IsWindows() {
			testPath = test.pathWindows
		} else {
			testPath = test.pathLinux
		}

		t.Run(fmt.Sprintf("%v %v %v", fsType, i, testPath), func(t *testing.T) {
			dir, file := FilePathSplit(test.fs, testPath)
			assert.Equal(t, test.dir, dir)
			assert.Equal(t, test.file, file)
		})
	}
}

func TestFilePathVolume(t *testing.T) {
	if !platform.IsWindows() {
		t.Skip("volume is a windows concept")
	}
	assert.Empty(t, FilePathVolumeName(nil, "C:"))
	assert.Equal(t, "C:", FilePathVolumeName(NewTestFilesystem(t, '\\'), "C:"))
	assert.Equal(t, "C:", FilePathVolumeName(NewTestFilesystem(t, '\\'), "C://"))
	assert.Equal(t, "C:", FilePathVolumeName(NewTestFilesystem(t, '/'), "C://"))
}
