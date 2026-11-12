/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package frontmatter

import (
	"context"
	"io"

	core "github.com/ARM-software/golang-utils/utils/serialisation/frontmatter"
	serialisation "github.com/ARM-software/golang-utils/utils/serialisation/json/frontmatter"
)

// NewJSONFrontmatterFormat returns the JSON front matter format used by Hugo.
func NewJSONFrontmatterFormat() *core.Format { return serialisation.NewJSONFrontmatterFormat() }

// NewJSONFrontmatterFormat2 returns the JSON front matter format used by Hexo.
func NewJSONFrontmatterFormat2() *core.Format { return serialisation.NewJSONFrontmatterFormat2() }

// NewJSONFrontmatterFormat3 returns the JSON front matter format used by Deno.
func NewJSONFrontmatterFormat3() *core.Format { return serialisation.NewJSONFrontmatterFormat3() }

// ExtractJSONFrontmatter extracts JSON front matter from a reader.
func ExtractJSONFrontmatter(ctx context.Context, r io.Reader) ([]byte, error) {
	return serialisation.ExtractJSONFrontmatter(ctx, r)
}

// ExtractDocument extracts JSON front matter and the remaining document from a reader.
func ExtractDocument(ctx context.Context, r io.Reader, options ...core.ParserOption) ([]byte, []byte, error) {
	return serialisation.ExtractDocument(ctx, r, options...)
}

// ExtractContent returns document content without the JSON front matter block.
func ExtractContent(ctx context.Context, r io.Reader, options ...core.ParserOption) ([]byte, error) {
	return serialisation.ExtractContent(ctx, r, options...)
}

// ExtractDocumentWithDecode decodes front matter into v and returns document content.
func ExtractDocumentWithDecode(ctx context.Context, r io.Reader, v any, options ...core.ParserOption) ([]byte, error) {
	return serialisation.ExtractDocumentWithDecode(ctx, r, v, options...)
}

// UnmarshallJSONFrontMatter extracts and unmarshals JSON front matter.
func UnmarshallJSONFrontMatter(ctx context.Context, r io.Reader, v any) error {
	return serialisation.UnmarshallJSONFrontMatter(ctx, r, v)
}
