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
	"github.com/ARM-software/golang-utils/utils/serialisation/frontmatter" //nolint:misspell
	"github.com/ARM-software/golang-utils/utils/serialisation/yaml"        //nolint:misspell
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

// ExtractDocument extracts YAML front matter and the remaining document from r.
// Options configure the document parser; content is preserved by default.
func ExtractDocument(ctx context.Context, r io.Reader, options ...frontmatter.ParserOption) (metadata []byte, content []byte, err error) {
	if err = parallelisation.DetermineContextError(ctx); err != nil {
		return
	}

	p, err := frontmatter.NewDocumentParser(NewYAMLFrontmatterFormat(), NewYAMLDocumentEndFrontmatterFormat(), NewYAML2FrontmatterFormat())
	if err != nil {
		err = commonerrors.DescribeCircumstance(err, "cannot create a YAML document parser")
		return
	}

	metadata, content, err = p.Configure(options...).Extract(ctx, r)
	if err != nil {
		err = commonerrors.DescribeCircumstance(err, "failed extracting the YAML document")
	}

	return
}

// ExtractContent returns document content without the YAML front matter block.
// It forwards options and errors from ExtractDocument without decoding metadata.
func ExtractContent(ctx context.Context, r io.Reader, options ...frontmatter.ParserOption) (content []byte, err error) {
	_, content, err = ExtractDocument(ctx, r, options...)
	return
}

// ExtractDocumentWithDecode decodes YAML front matter into v and returns the
// document content. Empty metadata leaves v unchanged. Non-empty metadata requires
// a destination accepted by the YAML decoder. Content is returned even if decoding
// fails; extraction errors prevent decoding.
func ExtractDocumentWithDecode(ctx context.Context, r io.Reader, v any, options ...frontmatter.ParserOption) (content []byte, err error) {
	if err = parallelisation.DetermineContextError(ctx); err != nil {
		return
	}

	metadata, content, err := ExtractDocument(ctx, r, options...)
	if err != nil || len(metadata) == 0 {
		return
	}

	err = yaml.UnmarshallWithContext(ctx, metadata, v)
	if err != nil {
		err = commonerrors.DescribeCircumstance(err, "failed decoding the YAML frontmatter")
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
