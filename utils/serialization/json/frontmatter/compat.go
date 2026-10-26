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

// UnmarshallJSONFrontMatter extracts and unmarshals JSON front matter.
func UnmarshallJSONFrontMatter(ctx context.Context, r io.Reader, v any) error {
	return serialisation.UnmarshallJSONFrontMatter(ctx, r, v)
}
