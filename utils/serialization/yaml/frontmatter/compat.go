/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package frontmatter

import (
	"context"
	"io"

	core "github.com/ARM-software/golang-utils/utils/serialisation/frontmatter"
	serialisation "github.com/ARM-software/golang-utils/utils/serialisation/yaml/frontmatter"
)

// NewYAMLFrontmatterFormat returns the YAML front matter format used by Hugo.
func NewYAMLFrontmatterFormat() *core.Format { return serialisation.NewYAMLFrontmatterFormat() }

// NewYAMLDocumentEndFrontmatterFormat uses the YAML document end marker.
func NewYAMLDocumentEndFrontmatterFormat() *core.Format {
	return serialisation.NewYAMLDocumentEndFrontmatterFormat()
}

// NewYAML2FrontmatterFormat uses `---yaml` as the opening delimiter.
func NewYAML2FrontmatterFormat() *core.Format { return serialisation.NewYAML2FrontmatterFormat() }

// ExtractYAMLFrontmatter extracts YAML front matter from a reader.
func ExtractYAMLFrontmatter(ctx context.Context, r io.Reader) ([]byte, error) {
	return serialisation.ExtractYAMLFrontmatter(ctx, r)
}

// UnmarshallYAMLFrontMatter extracts and unmarshals YAML front matter.
func UnmarshallYAMLFrontMatter(ctx context.Context, r io.Reader, v any) error {
	return serialisation.UnmarshallYAMLFrontMatter(ctx, r, v)
}
