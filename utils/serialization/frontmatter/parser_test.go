package frontmatter

import (
	"context"
	"embed"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/safeio"
)

var (
	//go:embed testdata/*.md
	exampleFiles embed.FS
)

func TestParseYAMLFrontMatter(t *testing.T) {
	p, err := NewParserWithOptions(
		WithStart("---"),
		WithEnd("---"),
	)
	require.NoError(t, err)

	frontMatter, err := p.Parse(context.Background(), strings.NewReader("---\ntitle: Test\ncount: 2\n---\nbody\n"))
	require.NoError(t, err)

	content, err := safeio.ReadAll(context.Background(), frontMatter)
	require.NoError(t, err)
	assert.Equal(t, "title: Test\ncount: 2\n", string(content))
}

func TestParseTOMLFrontMatter(t *testing.T) {
	p, err := NewParserWithOptions(
		WithStart("+++"),
		WithEnd("+++"),
	)
	require.NoError(t, err)

	frontMatter, err := p.Parse(context.Background(), strings.NewReader("+++\ntitle = \"Test\"\ncount = 2\n+++\nbody\n"))
	require.NoError(t, err)

	content, err := safeio.ReadAll(context.Background(), frontMatter)
	require.NoError(t, err)
	assert.Equal(t, "title = \"Test\"\ncount = 2\n", string(content))
}

func TestParseJSONFrontMatterWithDelimiters(t *testing.T) {
	p, err := NewParserWithOptions(
		WithStart("{"),
		WithEnd("}"),
		WithUnmarshalDelimiters(),
		WithRequiresNewLine(),
	)
	require.NoError(t, err)

	frontMatter, err := p.Parse(context.Background(), strings.NewReader("{\n  \"title\": \"Test\"\n}\n\nbody\n"))
	require.NoError(t, err)

	content, err := safeio.ReadAll(context.Background(), frontMatter)
	require.NoError(t, err)
	assert.Equal(t, "{\n  \"title\": \"Test\"\n}\n", string(content))
}

func TestParseCRLFInspiredBySimplematterAndPositFrontmatter(t *testing.T) {
	// Sources:
	// - https://github.com/remcohaszing/simplematter/blob/main/test/test.ts
	// - https://github.com/posit-dev/frontmatter/blob/main/tests/testthat/test-whitespace.R
	p, err := NewParserWithOptions(
		WithStart("---"),
		WithEnd("---"),
	)
	require.NoError(t, err)

	frontMatter, err := p.Parse(context.Background(), strings.NewReader("---\r\ntitle: Test\r\n---\r\nBody\r\n"))
	require.NoError(t, err)

	content, err := safeio.ReadAll(context.Background(), frontMatter)
	require.NoError(t, err)
	assert.Equal(t, "title: Test\r\n", string(content))
}

func TestParseEmptyFrontMatterInspiredBySimplematterAndLpil(t *testing.T) {
	// Sources:
	// - https://github.com/remcohaszing/simplematter/blob/main/test/test.ts
	// - https://github.com/lpil/frontmatter/blob/main/test/frontmatter_test.gleam
	p, err := NewParserWithOptions(
		WithStart("---"),
		WithEnd("---"),
	)
	require.NoError(t, err)

	frontMatter, err := p.Parse(context.Background(), strings.NewReader("---\n---\nHello!"))
	require.NoError(t, err)

	_, err = safeio.ReadAll(context.Background(), frontMatter)
	errortest.AssertError(t, err, commonerrors.ErrEmpty)
}

func TestParseNotFound(t *testing.T) {
	p, err := NewParserWithOptions(
		WithStart("---"),
		WithEnd("---"),
	)
	require.NoError(t, err)

	_, err = p.Parse(context.Background(), strings.NewReader("body only\n"))
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrNotFound)
}

func TestParseYAMLFrontMatterWithoutTrailingNewline(t *testing.T) {
	p, err := NewParserWithOptions(
		WithStart("---"),
		WithEnd("---"),
	)
	require.NoError(t, err)

	frontMatter, err := p.Parse(context.Background(), strings.NewReader("---\ntitle: Test\ncount: 2\n---"))
	require.NoError(t, err)

	content, err := safeio.ReadAll(context.Background(), frontMatter)
	require.NoError(t, err)
	assert.Equal(t, "title: Test\ncount: 2\n", string(content))
}

func TestParseTrailingWhitespaceOnFenceInspiredByPositFrontmatter(t *testing.T) {
	// Source: https://github.com/posit-dev/frontmatter/blob/main/tests/testthat/test-whitespace.R
	p, err := NewParserWithOptions(
		WithStart("---"),
		WithEnd("---"),
	)
	require.NoError(t, err)

	frontMatter, err := p.Parse(context.Background(), strings.NewReader("---    \ntitle: Test\n---      \nBody"))
	require.NoError(t, err)

	content, err := safeio.ReadAll(context.Background(), frontMatter)
	require.NoError(t, err)
	assert.Equal(t, "title: Test\n", string(content))
}

func TestWithFormatOptions(t *testing.T) {
	format := NewFormat(
		WithStart("---json"),
		WithEnd("---"),
		WithUnmarshalDelimiters(),
		WithRequiresNewLine(),
	)
	require.NotNil(t, format)
	assert.Equal(t, "---json", format.Start)
	assert.Equal(t, "---", format.End)
	assert.True(t, format.UnmarshalDelims)
	assert.True(t, format.RequiresNewLine)
}

func TestParseCustomFormatInspiredByAdrgFrontmatter(t *testing.T) {
	// Source: https://github.com/adrg/frontmatter/blob/master/frontmatter_test.go
	p, err := NewParserWithOptions(
		WithStart("..."),
		WithEnd("..."),
	)
	require.NoError(t, err)

	frontMatter, err := p.Parse(context.Background(), strings.NewReader("...\nname: custom\ntags:\n  - go\n...\nbody\n"))
	require.NoError(t, err)

	content, err := safeio.ReadAll(context.Background(), frontMatter)
	require.NoError(t, err)
	assert.Equal(t, "name: custom\ntags:\n  - go\n", string(content))
}

func TestParseOnlyOpeningFenceInspiredByLpilAndSimplematter(t *testing.T) {
	// Sources:
	// - https://github.com/lpil/frontmatter/blob/main/test/frontmatter_test.gleam
	// - https://github.com/remcohaszing/simplematter/blob/main/test/test.ts
	p, err := NewParserWithOptions(
		WithStart("---"),
		WithEnd("---"),
	)
	require.NoError(t, err)

	_, err = p.Parse(context.Background(), strings.NewReader("---\nHello, Joe!"))
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrNotFound)
}

func TestParseWithMultipleFormats(t *testing.T) {
	yamlFormat := NewFormat(WithStart("---"), WithEnd("---"))
	tomlFormat := NewFormat(WithStart("+++"), WithEnd("+++"))

	p, err := NewParser(yamlFormat, tomlFormat)
	require.NoError(t, err)

	frontMatter, err := p.Parse(context.Background(), strings.NewReader("+++\ntitle = \"Test\"\n+++\nbody\n"))
	require.NoError(t, err)

	content, err := safeio.ReadAll(context.Background(), frontMatter)
	require.NoError(t, err)
	assert.Equal(t, "title = \"Test\"\n", string(content))
}

func TestParseYAMLExampleFilesInspiredByJxsonFrontMatter(t *testing.T) {
	// Sources:
	// - https://github.com/jxson/front-matter/blob/master/examples/yaml-seperator.md
	// - https://github.com/jxson/front-matter/blob/master/examples/complex-yaml.md
	tests := []struct {
		name            string
		fileName        string
		start           string
		end             string
		expectedContain string
	}{
		{
			name:            "yaml separator",
			fileName:        "yaml-seperator.md",
			start:           "= yaml =",
			end:             "= yaml =",
			expectedContain: "title: I couldn't think of a better name",
		},
		{
			name:            "complex yaml",
			fileName:        "complex-yaml.md",
			start:           "---",
			end:             "---",
			expectedContain: "pets:",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data, err := exampleFiles.ReadFile("testdata/" + test.fileName)
			require.NoError(t, err)

			p, err := NewParserWithOptions(
				WithStart(test.start),
				WithEnd(test.end),
			)
			require.NoError(t, err)

			frontMatter, err := p.Parse(context.Background(), strings.NewReader(string(data)))
			require.NoError(t, err)

			content, err := safeio.ReadAll(context.Background(), frontMatter)
			require.NoError(t, err)
			assert.Contains(t, string(content), test.expectedContain)
		})
	}
}
