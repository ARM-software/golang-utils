package frontmatter

import (
	"context"
	"io"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/serialization/frontmatter"            //nolint:misspell
	jsonserialization "github.com/ARM-software/golang-utils/utils/serialization/json" //nolint:misspell
)

// NewJSONFrontmatterFormat returns the JSON front matter format used by Hugo.
//
// Source: https://gohugo.io/content-management/front-matter/
func NewJSONFrontmatterFormat() *frontmatter.Format {
	return frontmatter.NewFormat(frontmatter.WithStart("{"), frontmatter.WithEnd("}"), frontmatter.WithRequiresNewLine(), frontmatter.WithUnmarshalDelimiters())
}

// NewJSONFrontmatterFormat2 returns the JSON front matter format used by Hexo.
//
// Source: https://hexo.io/docs/front-matter
func NewJSONFrontmatterFormat2() *frontmatter.Format {
	return frontmatter.NewFormat(frontmatter.WithStart(";;;"), frontmatter.WithEnd(";;;"))
}

// NewJSONFrontmatterFormat3 returns the JSON front matter format used by the
// Deno standard front matter package.
//
// Source: https://jsr.io/@std/front-matter
func NewJSONFrontmatterFormat3() *frontmatter.Format {
	return frontmatter.NewFormat(frontmatter.WithStart("---json"), frontmatter.WithEnd("---"))
}

// ExtractJSONFrontmatter extracts JSON front matter content from r.
//
// The extractor tries the known JSON front matter formats in order and returns
// only the extracted front matter bytes.
func ExtractJSONFrontmatter(ctx context.Context, r io.Reader) (content []byte, err error) {
	p, err := frontmatter.NewParser(NewJSONFrontmatterFormat(), NewJSONFrontmatterFormat2(), NewJSONFrontmatterFormat3())
	if err != nil {
		err = commonerrors.DescribeCircumstance(err, "cannot create a JSON frontmatter parser")
		return
	}
	content, err = p.Extract(ctx, r)
	if err != nil {
		err = commonerrors.DescribeCircumstance(err, "failed extracting the JSON frontmatter")
	}

	return
}

// UnmarshallJSONFrontMatter extracts JSON front matter from r and unmarshals it
// into v.
func UnmarshallJSONFrontMatter(ctx context.Context, r io.Reader, v any) (err error) {
	content, err := ExtractJSONFrontmatter(ctx, r)
	if err != nil {
		return
	}
	err = jsonserialization.UnmarshallWithContext(ctx, content, v)
	if err != nil {
		err = commonerrors.DescribeCircumstance(err, "failed unmarshalling the JSON frontmatter")
	}
	return
}
