/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package frontmatter

import (
	"bufio"
	"bytes"
	"context"
	"io"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
	"github.com/ARM-software/golang-utils/utils/safeio"
)

type documentParser struct {
	*parser
}

// NewDocumentParser constructs a parser returning both metadata and content using
// the supplied formats. Unlike NewParser, absence of front matter is not an error:
// the entire document is returned as content. Unclosed front matter remains an error.
func NewDocumentParser(formats ...*Format) (p IDocumentParser, err error) {
	metadataParser, err := NewParser(formats...)
	if err != nil {
		return
	}

	p = &documentParser{parser: metadataParser.(*parser)}
	return
}

// Configure applies parser options without changing metadata extraction behaviour.
func (p *documentParser) Configure(options ...ParserOption) IDocumentParser {
	p.parser.Configure(options...)
	return p
}

// Parse separates metadata from document content. The content reader streams from
// the input unless normalisation is enabled, which reads the content into memory.
// Introductory comments, BOM and whitespace are preserved before the remaining body
// by default. The returned readers may be consumed independently, in either order.
func (p *documentParser) Parse(ctx context.Context, r io.Reader) (frontmatter io.Reader, content io.Reader, err error) {
	if err = parallelisation.DetermineContextError(ctx); err != nil {
		return
	}

	if r == nil {
		err = commonerrors.UndefinedVariable("reader")
		return
	}

	e := &extractor{
		formats: p.formats,
		reader:  bufio.NewReader(safeio.NewContextualReader(ctx, r)),
		output:  bytes.NewBuffer(make([]byte, 0, p.options.BufferCapacity)),
	}

	metadata, err := e.extract()
	if err != nil {
		if !e.noFrontmatter {
			return
		}

		err = nil
	}

	buffered := e.output.Bytes()
	preamble := buffered
	var lookahead []byte

	if !e.noFrontmatter {
		preamble = buffered[:e.blockStart]
		bomStart, bomEnd := leadingBOM(buffered[e.blockStart:])
		if bomEnd > bomStart {
			preamble = append(bytes.Clone(preamble), buffered[e.blockStart+bomStart:e.blockStart+bomEnd]...)
		}

		lookahead = buffered[e.bodyStart:]
	}

	if p.options.DiscardIntroductoryComments && len(e.comments) != 0 {
		var retained []byte
		start := 0

		for _, comment := range e.comments {
			retained = append(retained, preamble[start:comment.start]...)
			start = comment.end
		}

		preamble = append(retained, preamble[start:]...)
	}

	// The extractor can consume a lookahead line and bufio can read further ahead.
	// Replay both before continuing from the buffered reader, not the original input.
	content = safeio.NewContextualReader(ctx, io.MultiReader(bytes.NewReader(preamble), bytes.NewReader(lookahead), e.reader))
	if p.options.NormaliseContent {
		body, subErr := io.ReadAll(content)
		if subErr != nil {
			err = safeio.ConvertIOError(subErr)
			return
		}

		body = bytes.ReplaceAll(body, []byte("\r\n"), []byte("\n"))
		body = bytes.ReplaceAll(body, []byte("\r"), []byte("\n"))
		content = safeio.NewByteReader(ctx, bytes.TrimSpace(body))
	}

	frontmatter = safeio.NewByteReader(ctx, metadata)
	return
}

// Extract materialises both document parts. Empty metadata and content are valid.
func (p *documentParser) Extract(ctx context.Context, r io.Reader) (frontmatter []byte, content []byte, err error) {
	if err = parallelisation.DetermineContextError(ctx); err != nil {
		return
	}

	metadataReader, contentReader, err := p.Parse(ctx, r)
	if err != nil {
		return
	}

	frontmatter, err = io.ReadAll(metadataReader)
	if err != nil {
		err = safeio.ConvertIOError(err)
		return
	}

	content, err = io.ReadAll(contentReader)
	err = safeio.ConvertIOError(err)
	return
}
