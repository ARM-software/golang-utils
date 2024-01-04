package filesystem

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
)

func TestFilepathStem(t *testing.T) {
	t.Run("given a filename with extension, it strips extension", func(t *testing.T) {
		assert.Equal(t, "foo", FilepathStem("foo.bar"))
		assert.Equal(t, "library.tar", FilepathStem("library.tar.gz"))
		assert.Equal(t, "cool", FilepathStem("cool"))
	})

	t.Run("given a filepath, it returns only file name", func(t *testing.T) {
		fp := filepath.Join("super", "foo", "bar.baz")
		assert.Equal(t, "bar", FilepathStem(fp))
		fp = filepath.Join("nice", "file", "path")
		assert.Equal(t, "path", FilepathStem(fp))
	})
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
