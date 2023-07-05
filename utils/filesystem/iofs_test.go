package filesystem

import (
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_IoFS_Exists(t *testing.T) {
	ifs, err := NewEmbedFileSystem(&testContent)
	require.NoError(t, err)

	ioFs, err := ConvertToIOFilesystem(ifs)
	require.NoError(t, err)

	fs, err := ConvertFromIOFilesystem(ioFs)
	require.NoError(t, err)

	assert.False(t, fs.Exists(faker.DomainName()))
	assert.True(t, fs.Exists("testdata"))
	assert.True(t, fs.Exists("testdata/embed"))
	assert.True(t, fs.Exists("testdata/embed/level1/test.txt"))
}
