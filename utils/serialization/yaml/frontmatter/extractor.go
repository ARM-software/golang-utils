package frontmatter

import (
	"context"
	"io"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/serialization/frontmatter" //nolint:misspell
	"github.com/ARM-software/golang-utils/utils/serialization/yaml"        //nolint:misspell
)

// NewYAMLFrontmatterFormat returns the YAML front matter format used by Hugo.
//
// Source: https://gohugo.io/content-management/front-matter/
func NewYAMLFrontmatterFormat() *frontmatter.Format {
	return frontmatter.NewFormat(frontmatter.WithStart("---"), frontmatter.WithEnd("---"))
}

// NewYAMLDocumentEndFrontmatterFormat returns the YAML front matter format that
// uses the YAML document end marker.
//
// Source: https://yaml.org/spec/1.2.2/#912-document-markers
func NewYAMLDocumentEndFrontmatterFormat() *frontmatter.Format {
	return frontmatter.NewFormat(frontmatter.WithStart("---"), frontmatter.WithEnd("..."))
}

// NewYAML2FrontmatterFormat returns the alternative YAML front matter format
// using `---yaml` as the opening delimiter.
//
// Source: https://github.com/adrg/frontmatter/tree/master
func NewYAML2FrontmatterFormat() *frontmatter.Format {
	return frontmatter.NewFormat(frontmatter.WithStart("---yaml"), frontmatter.WithEnd("---"))
}

// ExtractYAMLFrontmatter extracts YAML front matter content from r.
//
// The extractor tries the known YAML front matter formats in order and returns
// only the extracted front matter bytes.
func ExtractYAMLFrontmatter(ctx context.Context, r io.Reader) (content []byte, err error) {
	p, err := frontmatter.NewParser(NewYAMLFrontmatterFormat(), NewYAMLDocumentEndFrontmatterFormat(), NewYAML2FrontmatterFormat())
	if err != nil {
		err = commonerrors.DescribeCircumstance(err, "cannot create a YAML frontmatter parser")
		return
	}
	content, err = p.Extract(ctx, r)
	if err != nil {
		err = commonerrors.DescribeCircumstance(err, "failed extracting the YAML frontmatter")
	}

	return
}

// UnmarshallYAMLFrontMatter extracts YAML front matter from r and unmarshals it
// into v.
func UnmarshallYAMLFrontMatter(ctx context.Context, r io.Reader, v any) (err error) {
	content, err := ExtractYAMLFrontmatter(ctx, r)
	if err != nil {
		return
	}
	err = yaml.UnmarshallWithContext(ctx, content, v)
	if err != nil {
		err = commonerrors.DescribeCircumstance(err, "failed unmarshalling the YAML frontmatter")
	}
	return
}
