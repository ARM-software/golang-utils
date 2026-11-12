/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package frontmatter

import serialisation "github.com/ARM-software/golang-utils/utils/serialisation/frontmatter"

// Format describes a front matter block.
type Format = serialisation.Format

// FormatOption configures a front matter format.
type FormatOption = serialisation.FormatOption

// ParserOptions configures the front matter parser.
type ParserOptions = serialisation.ParserOptions

// ParserOption configures the front matter parser.
type ParserOption = serialisation.ParserOption

// IFrontMatterParser extracts front matter from a reader.
type IFrontMatterParser = serialisation.IFrontMatterParser

// IDocumentParser extracts front matter and document content from a reader.
type IDocumentParser = serialisation.IDocumentParser

// DefaultFormat returns a format with the package defaults.
func DefaultFormat() *Format { return serialisation.DefaultFormat() }

// NewFormat constructs a format from the supplied options.
func NewFormat(options ...FormatOption) *Format { return serialisation.NewFormat(options...) }

// WithStart sets the opening delimiter.
func WithStart(start string) FormatOption { return serialisation.WithStart(start) }

// WithEnd sets the closing delimiter.
func WithEnd(end string) FormatOption { return serialisation.WithEnd(end) }

// IncludeDelimitersWhenUnmarshalling includes the delimiters in the extracted content.
func IncludeDelimitersWhenUnmarshalling() FormatOption {
	return serialisation.IncludeDelimitersWhenUnmarshalling()
}

// WithRequiresNewLine requires an empty line after the closing delimiter.
func WithRequiresNewLine() FormatOption { return serialisation.WithRequiresNewLine() }

// DefaultParserOptions returns the default parser options.
func DefaultParserOptions() *ParserOptions { return serialisation.DefaultParserOptions() }

// NewParserOptions constructs parser options from the supplied options.
func NewParserOptions(options ...ParserOption) *ParserOptions {
	return serialisation.NewParserOptions(options...)
}

// WithBufferCapacity configures the parser's buffer capacity.
func WithBufferCapacity(capacity int) ParserOption { return serialisation.WithBufferCapacity(capacity) }

// WithNormaliseContent configures document body whitespace and line-ending normalisation.
func WithNormaliseContent(normalise bool) ParserOption {
	return serialisation.WithNormaliseContent(normalise)
}

// WithDiscardIntroductoryComments configures removal of introductory HTML comments.
func WithDiscardIntroductoryComments(discard bool) ParserOption {
	return serialisation.WithDiscardIntroductoryComments(discard)
}

// NewDocumentParser creates a document parser for the supplied formats.
func NewDocumentParser(formats ...*Format) (IDocumentParser, error) {
	return serialisation.NewDocumentParser(formats...)
}

// NewParser creates a parser for the supplied formats.
func NewParser(formats ...*Format) (IFrontMatterParser, error) {
	return serialisation.NewParser(formats...)
}

// NewParserWithFormatOptions creates a parser with the supplied format options.
func NewParserWithFormatOptions(options ...FormatOption) (IFrontMatterParser, error) {
	return serialisation.NewParserWithFormatOptions(options...)
}
