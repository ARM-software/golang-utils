/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package frontmatter

import (
	"context"
	"io"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
	"github.com/ARM-software/golang-utils/utils/serialisation/frontmatter"            //nolint:misspell
	jsonserialisation "github.com/ARM-software/golang-utils/utils/serialisation/json" //nolint:misspell
)

// NewJSONFrontmatterFormat returns the JSON front matter format used by Hugo.
//
// Source: https://gohugo.io/content-management/front-matter/
func NewJSONFrontmatterFormat() *frontmatter.Format {
	return frontmatter.NewFormat(frontmatter.WithStart("{"), frontmatter.WithEnd("}"), frontmatter.WithRequiresNewLine(), frontmatter.IncludeDelimitersWhenUnmarshalling())
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

// ExtractDocument extracts JSON front matter and the remaining document from r.
// Options configure the document parser; content is preserved by default.
func ExtractDocument(ctx context.Context, r io.Reader, options ...frontmatter.ParserOption) (metadata []byte, content []byte, err error) {
	if err = parallelisation.DetermineContextError(ctx); err != nil {
		return
	}

	p, err := frontmatter.NewDocumentParser(NewJSONFrontmatterFormat(), NewJSONFrontmatterFormat2(), NewJSONFrontmatterFormat3())
	if err != nil {
		err = commonerrors.DescribeCircumstance(err, "cannot create a JSON document parser")
		return
	}

	metadata, content, err = p.Configure(options...).Extract(ctx, r)
	if err != nil {
		err = commonerrors.DescribeCircumstance(err, "failed extracting the JSON document")
	}

	return
}

// ExtractContent returns document content without the JSON front matter block.
// It forwards options and errors from ExtractDocument without decoding metadata.
func ExtractContent(ctx context.Context, r io.Reader, options ...frontmatter.ParserOption) (content []byte, err error) {
	_, content, err = ExtractDocument(ctx, r, options...)
	return
}

// ExtractDocumentWithDecode decodes JSON front matter into v and returns the
// document content. Empty metadata leaves v unchanged. Non-empty metadata requires
// a destination accepted by the JSON decoder. Content is returned even if decoding
// fails; extraction errors prevent decoding.
func ExtractDocumentWithDecode(ctx context.Context, r io.Reader, v any, options ...frontmatter.ParserOption) (content []byte, err error) {
	if err = parallelisation.DetermineContextError(ctx); err != nil {
		return
	}

	metadata, content, err := ExtractDocument(ctx, r, options...)
	if err != nil || len(metadata) == 0 {
		return
	}

	err = jsonserialisation.UnmarshallWithContext(ctx, metadata, v)
	if err != nil {
		err = commonerrors.DescribeCircumstance(err, "failed decoding the JSON frontmatter")
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
	err = jsonserialisation.UnmarshallWithContext(ctx, content, v)
	if err != nil {
		err = commonerrors.DescribeCircumstance(err, "failed unmarshalling the JSON frontmatter")
	}
	return
}
