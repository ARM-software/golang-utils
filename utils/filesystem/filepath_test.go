package filesystem

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
