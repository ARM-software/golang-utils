package frontmatter

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
)

func TestExtractJSONFrontmatterHugoInspiredByAdrgFrontmatter(t *testing.T) {
	// Source: https://github.com/adrg/frontmatter/blob/master/frontmatter_test.go
	content, err := ExtractJSONFrontmatter(context.Background(), strings.NewReader("{\n  \"name\": \"frontmatter\"\n}\n\nrest of the file\n"))
	require.NoError(t, err)
	assert.Equal(t, "{\n  \"name\": \"frontmatter\"\n}\n", string(content))
}

func TestExtractJSONFrontmatterHexoInspiredByAdrgFrontmatter(t *testing.T) {
	// Source: https://github.com/adrg/frontmatter/blob/master/frontmatter_test.go
	content, err := ExtractJSONFrontmatter(context.Background(), strings.NewReader(";;;\n{\n  \"name\": \"frontmatter\"\n}\n;;;\nrest of the file\n"))
	require.NoError(t, err)
	assert.Equal(t, "{\n  \"name\": \"frontmatter\"\n}\n", string(content))
}

func TestUnmarshallJSONFrontMatterInspiredByAdrgFrontmatter(t *testing.T) {
	// Source: https://github.com/adrg/frontmatter/blob/master/frontmatter_test.go
	type matter struct {
		Name string `json:"name"`
	}

	var v matter
	err := UnmarshallJSONFrontMatter(context.Background(), strings.NewReader("---json\n{\n  \"name\": \"frontmatter\"\n}\n---\nrest of the file\n"), &v)
	require.NoError(t, err)
	assert.Equal(t, "frontmatter", v.Name)
}

func TestExtractJSONFrontmatterNotFoundInspiredBySimplematter(t *testing.T) {
	// Source: https://github.com/remcohaszing/simplematter/blob/main/test/test.ts
	_, err := ExtractJSONFrontmatter(context.Background(), strings.NewReader("Rest of document\n"))
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrNotFound)
}
