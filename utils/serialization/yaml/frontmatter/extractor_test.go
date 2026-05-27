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
	"github.com/ARM-software/golang-utils/utils/reflection"
)

var (
	//go:embed testdata/*.md
	exampleFiles embed.FS
)

func TestExtractYAMLFrontmatterInspiredBySimplematter(t *testing.T) {
	// Source: https://github.com/remcohaszing/simplematter/blob/main/test/test.ts
	content, err := ExtractYAMLFrontmatter(context.Background(), strings.NewReader("---\nhello: yaml\n---\nRest of document\n"))
	require.NoError(t, err)
	assert.Equal(t, "hello: yaml\n", string(content))
}

func TestExtractYAMLFrontmatterCRLFInspiredBySimplematterAndPositFrontmatter(t *testing.T) {
	// Sources:
	// - https://github.com/remcohaszing/simplematter/blob/main/test/test.ts
	// - https://github.com/posit-dev/frontmatter/blob/main/tests/testthat/test-whitespace.R
	content, err := ExtractYAMLFrontmatter(context.Background(), strings.NewReader("---\r\ntitle: Test\r\n---\r\nBody\r\n"))
	require.NoError(t, err)
	assert.Equal(t, "title: Test\r\n", string(content))
}

func TestUnmarshallYAMLFrontMatterInspiredByAdrgFrontmatter(t *testing.T) {
	// Source: https://github.com/adrg/frontmatter/blob/master/frontmatter_test.go
	type matter struct {
		Name string `yaml:"name"`
	}

	var v matter
	err := UnmarshallYAMLFrontMatter(context.Background(), strings.NewReader("---\nname: frontmatter\n---\nrest of the file\n"), &v)
	require.NoError(t, err)
	assert.Equal(t, "frontmatter", v.Name)
}

func TestExtractYAMLFrontmatterNotFoundInspiredByLpil(t *testing.T) {
	// Source: https://github.com/lpil/frontmatter/blob/main/test/frontmatter_test.gleam
	_, err := ExtractYAMLFrontmatter(context.Background(), strings.NewReader("Hello, Joe!"))
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrNotFound)
}

func TestExtractYAMLFrontmatterExampleFilesInspiredByJxsonFrontMatter(t *testing.T) {
	// Sources:
	// - https://github.com/jxson/front-matter/blob/master/examples/complex-yaml.md
	tests := []struct {
		name            string
		fileName        string
		expectedContain string
	}{
		{
			name:            "complex yaml",
			fileName:        "complex-yaml.md",
			expectedContain: "pets:",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data, err := exampleFiles.ReadFile("testdata/" + test.fileName)
			require.NoError(t, err)

			content, err := ExtractYAMLFrontmatter(context.Background(), strings.NewReader(string(data)))
			require.NoError(t, err)
			assert.Contains(t, string(content), test.expectedContain)
		})
	}
}

func TestUnmarshallYAMLFrontMatterExampleFileInspiredByJxsonFrontMatter(t *testing.T) {
	// Source: https://github.com/jxson/front-matter/blob/master/examples/complex-yaml.md
	type matter struct {
		Title string   `json:"title"`
		Name  string   `json:"name"`
		Age   string   `json:"age"`
		Pets  []string `json:"pets"`
	}

	data, err := exampleFiles.ReadFile("testdata/complex-yaml.md")
	require.NoError(t, err)

	var v matter
	err = UnmarshallYAMLFrontMatter(context.Background(), strings.NewReader(string(data)), &v)
	require.NoError(t, err)
	assert.Equal(t, "This is a title!", v.Title)
	assert.Equal(t, "Derek Worthen", v.Name)
	assert.Equal(t, "young", v.Age)
	assert.ElementsMatch(t, []string{"cat", "dog", "bat"}, v.Pets)
}

type ModelCard struct {
	Tags        []string `yaml:"tags" json:"tags"`
	Datasets    string   `yaml:"datasets" json:"datasets"`
	PipelineTag string   `yaml:"pipeline_tag" json:"pipeline_tag"`
	Licences    []string `yaml:"license" json:"license"`
	LicenceName string   `yaml:"license_name" json:"license_name"`
	LicenceURL  string   `yaml:"license_link" json:"license_link"`
	LibraryName string   `yaml:"library_name" json:"library_name"`
}

func TestParseModelCard_WhenReadmeHasFrontMatter_ShouldReturnCard(t *testing.T) {
	var card ModelCard
	err := UnmarshallYAMLFrontMatter(context.Background(), strings.NewReader("---\ntags:\n  - text-generation\n  - llama\npipeline_tag: text-generation\nlicense: [mit, apache-2.0]\nlicense_name: MIT\nlicense_link: https://example.com/license\nlibrary_name: transformers\ndatasets: dataset/name\n---\n# Model\n"), &card)
	require.NoError(t, err)

	assert.ElementsMatch(t, []string{"llama", "text-generation"}, card.Tags)
	assert.Equal(t, "dataset/name", card.Datasets)
	assert.Equal(t, "text-generation", card.PipelineTag)
	assert.ElementsMatch(t, []string{"apache-2.0", "mit"}, card.Licences)
	assert.Equal(t, "MIT", card.LicenceName)
	assert.Equal(t, "https://example.com/license", card.LicenceURL)
	assert.Equal(t, "transformers", card.LibraryName)
}

func TestUnmarshallYAMLFrontMatter_WhenReadmeHasNoFrontMatter_ShouldReturnNotFound(t *testing.T) {
	var card ModelCard
	err := UnmarshallYAMLFrontMatter(context.Background(), strings.NewReader("# Model\n"), &card)
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrNotFound)
	assert.True(t, reflection.IsEmpty(card))
}

func TestUnmarshallYAMLFrontMatter_WhenFrontMatterIsInvalid_ShouldReturnInvalid(t *testing.T) {
	var card ModelCard
	err := UnmarshallYAMLFrontMatter(context.Background(), strings.NewReader("---\nlicense: [\n---\nbody"), &card)
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrMarshalling)
}
