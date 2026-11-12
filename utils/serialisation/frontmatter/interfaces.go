/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package frontmatter

import (
	"context"
	"io"
)

//go:generate go tool mockgen -destination=../../mocks/serialisation/mock_$GOPACKAGE.go -package=mocks github.com/ARM-software/golang-utils/utils/serialisation/$GOPACKAGE IFrontMatterParser
//go:generate go tool mockgen -destination=../mocks/mock_$GOPACKAGE.go -package=mocks github.com/ARM-software/golang-utils/utils/serialisation/$GOPACKAGE IFrontMatterParser

// IFrontMatterParser extracts front matter content from a reader.
//
// Implementations detect one of the configured front matter formats at the
// start of the stream and return only the extracted front matter content.
type IFrontMatterParser interface {
	// Configure applies a parser option to the front matter parser.
	Configure(...ParserOption) IFrontMatterParser

	// Parse extracts the front matter from r and returns a reader containing only
	// the extracted front matter content.
	Parse(context.Context, io.Reader) (frontmatter io.Reader, err error)

	// Extract extracts the front matter from r and returns the extracted front
	// matter content as raw bytes.
	Extract(context.Context, io.Reader) (frontmatter []byte, err error)
}

// IDocumentParser extracts front matter and the remaining document content.
// It is separate from IFrontMatterParser so existing implementations remain valid.
type IDocumentParser interface {
	// Configure applies options to the document parser.
	Configure(...ParserOption) IDocumentParser

	// Parse returns independent readers for metadata and content. Content preserves
	// introductory comments and whitespace unless the corresponding options are set.
	// The input must remain available until the content reader has been consumed.
	Parse(context.Context, io.Reader) (frontmatter io.Reader, content io.Reader, err error)

	// Extract reads metadata and content into memory. Empty parts are valid.
	Extract(context.Context, io.Reader) (frontmatter []byte, content []byte, err error)
}
