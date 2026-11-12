/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package frontmatter

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
)

func TestDocumentParser(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name     string
		input    string
		metadata string
		content  string
	}{
		{name: "exact body", input: "---\ntitle: Test\n---\n\n  Body  text\t\r\nlast", metadata: "title: Test\n", content: "\n  Body  text\t\r\nlast"},
		{name: "CRLF", input: "---\r\ntitle: Test\r\n---\r\n\r\nBody\r\n", metadata: "title: Test\r\n", content: "\r\nBody\r\n"},
		{name: "no body or final newline", input: "---\ntitle: Test\n---", metadata: "title: Test\n"},
		{name: "empty metadata", input: "---\n---\nBody", content: "Body"},
		{name: "empty parts", input: "---\n---"},
		{name: "BOM", input: "\xef\xbb\xbf---\ntitle: Test\n---\nBody", metadata: "title: Test\n", content: "\xef\xbb\xbfBody"},
		{name: "indented repeated BOM", input: " \t\xef\xbb\xbf\xef\xbb\xbf---\ntitle: Test\n---\nBody", metadata: "title: Test\n", content: "\xef\xbb\xbf\xef\xbb\xbfBody"},
		{name: "preamble", input: "\xef\xbb\xbf\n \t\n<!-- intro -->\n---\ntitle: Test\n---\nBody", metadata: "title: Test\n", content: "\xef\xbb\xbf\n \t\n<!-- intro -->\nBody"},
		{name: "shared opening fence", input: "---\ntitle: Test\n...\nBody", metadata: "title: Test\n", content: "Body"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			p, err := NewDocumentParser(NewFormat(WithStart("---"), WithEnd("---")), NewFormat(WithStart("---"), WithEnd("...")))
			require.NoError(t, err)

			metadata, content, err := p.Extract(context.Background(), iotest.OneByteReader(strings.NewReader(test.input)))
			require.NoError(t, err)
			assert.Equal(t, test.metadata, string(metadata))
			assert.Equal(t, test.content, string(content))
		})
	}

	t.Run("JSON lookahead", func(t *testing.T) {
		t.Parallel()

		p, err := NewDocumentParser(NewFormat(WithStart("{"), WithEnd("}"), IncludeDelimitersWhenUnmarshalling(), WithRequiresNewLine()))
		require.NoError(t, err)

		metadata, content, err := p.Extract(context.Background(), strings.NewReader("{\nvalue\n}\nnot blank\n}\n \t\r\nBody"))
		require.NoError(t, err)
		assert.Equal(t, "{\nvalue\n}\nnot blank\n}\n", string(metadata))
		assert.Equal(t, " \t\r\nBody", string(content))

		p, err = NewDocumentParser(NewFormat(WithStart("{"), WithEnd("}"), WithRequiresNewLine()))
		require.NoError(t, err)
		metadata, content, err = p.Extract(context.Background(), strings.NewReader("{\nvalue\n}\n\nBody"))
		require.NoError(t, err)
		assert.Equal(t, "value\n", string(metadata))
		assert.Equal(t, "\nBody", string(content))
	})
}

func TestDocumentParserReaders(t *testing.T) {
	t.Parallel()

	p, err := NewDocumentParser(NewFormat(WithStart("+++"), WithEnd("+++")))
	require.NoError(t, err)
	p.Configure(WithBufferCapacity(1))
	body := strings.Repeat("Body  text\r\n", 2000)
	metadataText := strings.Repeat("metadata\n", 2000)
	src := strings.NewReader(fmt.Sprintf("+++\n%s+++\n%s", metadataText, body))

	metadata, content, err := p.Parse(context.Background(), src)
	require.NoError(t, err)
	assert.Positive(t, src.Len(), "Parse must not eagerly consume the body")
	assert.Equal(t, body, string(readDocumentPart(t, content)))
	assert.Equal(t, metadataText, string(readDocumentPart(t, metadata)), "metadata remains readable after the body")
}

func readDocumentPart(t *testing.T, r io.Reader) []byte {
	t.Helper()

	data, err := io.ReadAll(r)
	require.NoError(t, err)
	return data
}

func TestDocumentParserWithoutFrontmatter(t *testing.T) {
	t.Parallel()

	for _, input := range []string{"", "Body\r\nlast", " \t\r\n\u2003", "\xef\xbb\xbf<!-- intro -->\n\nBody", "<!-- unterminated\nBody", "Body\n---\ntitle: later\n---\n"} {
		t.Run(input, func(t *testing.T) {
			t.Parallel()

			p, err := NewDocumentParser(NewFormat(WithStart("---"), WithEnd("---")))
			require.NoError(t, err)

			metadata, content, err := p.Extract(context.Background(), strings.NewReader(input))
			require.NoError(t, err)
			assert.Empty(t, metadata)
			assert.Equal(t, input, string(content))
		})
	}
}

func TestDocumentParserNormaliseContent(t *testing.T) {
	t.Parallel()

	assert.False(t, DefaultParserOptions().NormaliseContent)
	for _, test := range []struct {
		name    string
		input   string
		content string
		options []ParserOption
	}{
		{name: "Unicode whitespace and line endings", input: "---\r\n title: Test \r\n---\r\n\u2003\r\nFirst  line\t \r\n  second\rthird\n\u00a0", content: "First  line\t \n  second\nthird", options: []ParserOption{WithNormaliseContent(true)}},
		{name: "without metadata", input: "\u2003 First\u2003second\r\n", content: "First\u2003second", options: []ParserOption{WithNormaliseContent(true)}},
		{name: "whitespace only", input: " \t\r\n\u2003", options: []ParserOption{WithNormaliseContent(true)}},
		{name: "empty", options: []ParserOption{WithNormaliseContent(true)}},
		{name: "disabled again", input: " \r\nBody\r\n", content: " \r\nBody\r\n", options: []ParserOption{WithNormaliseContent(true), WithNormaliseContent(false)}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			p, err := NewDocumentParser(NewFormat(WithStart("---"), WithEnd("---")))
			require.NoError(t, err)
			metadata, content, err := p.Configure(test.options...).Extract(context.Background(), strings.NewReader(test.input))
			require.NoError(t, err)
			assert.Equal(t, test.content, string(content))
			if strings.HasPrefix(test.input, "---") {
				assert.Equal(t, " title: Test \r\n", string(metadata))
			}
		})
	}
}

func TestDocumentParserIntroductoryComments(t *testing.T) {
	t.Parallel()

	assert.False(t, DefaultParserOptions().DiscardIntroductoryComments)
	for _, test := range []struct {
		name      string
		input     string
		preserved string
		discarded string
	}{
		{name: "single line", input: "<!-- intro -->\n---\ntitle: Test\n---\nBody", preserved: "<!-- intro -->\nBody", discarded: "Body"},
		{name: "multiline and multiple blocks", input: "<!--\nintro\n-->\r\n\n<!-- second -->\n---\ntitle: Test\n---\nBody", preserved: "<!--\nintro\n-->\r\n\n<!-- second -->\nBody", discarded: "\nBody"},
		{name: "indented BOM and whitespace", input: " \t\xef\xbb\xbf\xef\xbb\xbf<!-- intro -->\n \t\n---\ntitle: Test\n---\nBody", preserved: " \t\xef\xbb\xbf\xef\xbb\xbf<!-- intro -->\n \t\nBody", discarded: "\xef\xbb\xbf\xef\xbb\xbf \t\nBody"},
		{name: "without metadata", input: "<!-- intro -->\r\n\nBody", preserved: "<!-- intro -->\r\n\nBody", discarded: "\nBody"},
		{name: "comment only at EOF", input: "<!-- intro -->", preserved: "<!-- intro -->"},
		{name: "comments after frontmatter remain", input: "<!-- intro -->\n---\ntitle: Test\n---\n<!-- body -->\nBody\n<!-- later -->", preserved: "<!-- intro -->\n<!-- body -->\nBody\n<!-- later -->", discarded: "<!-- body -->\nBody\n<!-- later -->"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			p, err := NewDocumentParser(NewFormat(WithStart("---"), WithEnd("---")))
			require.NoError(t, err)
			metadata, content, err := p.Extract(context.Background(), strings.NewReader(test.input))
			require.NoError(t, err)
			assert.Equal(t, test.preserved, string(content))

			strippedMetadata, content, err := p.Configure(WithDiscardIntroductoryComments(true)).Extract(context.Background(), strings.NewReader(test.input))
			require.NoError(t, err)
			assert.Equal(t, metadata, strippedMetadata)
			assert.Equal(t, test.discarded, string(content))
		})
	}

	p, err := NewDocumentParser(NewFormat(WithStart("---"), WithEnd("---")))
	require.NoError(t, err)
	input := " \r\n<!-- intro -->\r\n\r\n Body\rNext\r\n"
	_, content, err := p.Configure(WithDiscardIntroductoryComments(true), WithNormaliseContent(true)).Extract(context.Background(), strings.NewReader(input))
	require.NoError(t, err)
	assert.Equal(t, "Body\nNext", string(content))

	_, content, err = p.Configure(WithDiscardIntroductoryComments(false), WithNormaliseContent(false)).Extract(context.Background(), strings.NewReader(input))
	require.NoError(t, err)
	assert.Equal(t, input, string(content))
}

type terminalDocumentReader struct {
	data string
	err  error
}

func (r *terminalDocumentReader) Read(p []byte) (n int, err error) {
	n = copy(p, r.data)
	r.data = r.data[n:]
	if len(r.data) == 0 {
		err = r.err
		r.err = io.EOF
	}

	return
}

func TestDocumentParserErrors(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name      string
		input     io.Reader
		format    *Format
		normalise bool
		expected  error
	}{
		{name: "nil reader", expected: commonerrors.ErrUndefined},
		{name: "unclosed", input: strings.NewReader("---\ntitle: Test\n"), expected: commonerrors.ErrNotFound},
		{name: "source not-found is not absent metadata", input: iotest.ErrReader(commonerrors.ErrNotFound), expected: commonerrors.ErrNotFound},
		{name: "comment error", input: io.MultiReader(strings.NewReader("<!--\n"), iotest.ErrReader(commonerrors.ErrNotFound)), expected: commonerrors.ErrNotFound},
		{name: "source error with bytes", input: &terminalDocumentReader{data: "partial", err: commonerrors.ErrEmpty}, expected: commonerrors.ErrEmpty},
		{name: "truncated detection", input: iotest.ErrReader(io.ErrUnexpectedEOF), expected: commonerrors.ErrEOF},
		{name: "truncated comment", input: io.MultiReader(strings.NewReader("<!--\n"), iotest.ErrReader(io.ErrUnexpectedEOF)), expected: commonerrors.ErrEOF},
		{name: "truncated metadata", input: io.MultiReader(strings.NewReader("---\n"), &terminalDocumentReader{data: "partial", err: io.ErrUnexpectedEOF}), expected: commonerrors.ErrEOF},
		{name: "truncated lookahead", input: io.MultiReader(strings.NewReader("{\nvalue\n}\n"), iotest.ErrReader(io.ErrUnexpectedEOF)), format: NewFormat(WithStart("{"), WithEnd("}"), WithRequiresNewLine()), expected: commonerrors.ErrEOF},
		{name: "unclosed JSON", input: strings.NewReader("{\nvalue\n}\nBody"), format: NewFormat(WithStart("{"), WithEnd("}"), WithRequiresNewLine()), expected: commonerrors.ErrNotFound},
		{name: "body error", input: io.MultiReader(strings.NewReader("---\ntitle: Test\n---\n"), &terminalDocumentReader{data: "partial body", err: commonerrors.ErrUnexpected}), expected: commonerrors.ErrUnexpected},
		{name: "normalised body error", input: io.MultiReader(strings.NewReader("Body\n"), iotest.ErrReader(commonerrors.ErrUnexpected)), normalise: true, expected: commonerrors.ErrUnexpected},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			format := test.format
			if format == nil {
				format = NewFormat(WithStart("---"), WithEnd("---"))
			}

			p, err := NewDocumentParser(format)
			require.NoError(t, err)
			_, content, err := p.Configure(WithNormaliseContent(test.normalise)).Extract(context.Background(), test.input)
			errortest.RequireError(t, err, test.expected)
			assert.Nil(t, content)
		})
	}

	t.Run("streaming body error is deferred", func(t *testing.T) {
		t.Parallel()

		p, err := NewDocumentParser(NewFormat(WithStart("---"), WithEnd("---")))
		require.NoError(t, err)
		_, content, err := p.Parse(context.Background(), io.MultiReader(strings.NewReader("---\ntitle: Test\n---\n"), iotest.ErrReader(commonerrors.ErrUnexpected)))
		require.NoError(t, err)
		_, err = io.ReadAll(content)
		errortest.RequireError(t, err, commonerrors.ErrUnexpected)
	})

	t.Run("cancellation", func(t *testing.T) {
		t.Parallel()

		p, err := NewDocumentParser(NewFormat(WithStart("---"), WithEnd("---")))
		require.NoError(t, err)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		metadata, content, err := p.Parse(ctx, strings.NewReader("---\ntitle: Test\n---\nBody"))
		require.NoError(t, err)
		cancel()
		_, err = io.ReadAll(metadata)
		errortest.RequireError(t, err, context.Canceled)
		_, err = io.ReadAll(content)
		errortest.RequireError(t, err, context.Canceled)
		_, _, err = p.Extract(ctx, nil)
		errortest.RequireError(t, err, commonerrors.ErrCancelled)
		_, _, err = p.Parse(ctx, nil)
		errortest.RequireError(t, err, commonerrors.ErrCancelled)
	})
}

func TestDocumentParserConstructor(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name     string
		formats  []*Format
		expected error
	}{
		{name: "no formats", expected: commonerrors.ErrUndefined},
		{name: "nil format", formats: []*Format{nil}, expected: commonerrors.ErrUndefined},
		{name: "missing opening", formats: []*Format{{End: "---"}}, expected: commonerrors.ErrInvalid},
		{name: "missing closing", formats: []*Format{{Start: "---"}}, expected: commonerrors.ErrInvalid},
	} {
		t.Run(test.name, func(t *testing.T) {
			p, err := NewDocumentParser(test.formats...)
			errortest.RequireError(t, err, test.expected)
			assert.Nil(t, p)
		})
	}
}

type legacyDocumentTestParser struct{}

var _ IFrontMatterParser = legacyDocumentTestParser{}

func (p legacyDocumentTestParser) Configure(...ParserOption) IFrontMatterParser { return p }

func (legacyDocumentTestParser) Parse(context.Context, io.Reader) (io.Reader, error) {
	return strings.NewReader("legacy"), nil
}

func (legacyDocumentTestParser) Extract(context.Context, io.Reader) ([]byte, error) {
	return []byte("legacy"), nil
}

func TestFrontMatterParserCompatibility(t *testing.T) {
	t.Parallel()

	p, err := NewParser(NewFormat(WithStart("---"), WithEnd("---")))
	require.NoError(t, err)
	p.Configure(WithNormaliseContent(true), WithDiscardIntroductoryComments(true))
	input := "<!-- intro -->\n---\r\n title: Test \r\n---\r\nBody"
	metadata, err := p.Parse(context.Background(), strings.NewReader(input))
	require.NoError(t, err)
	assert.Equal(t, " title: Test \r\n", string(readDocumentPart(t, metadata)))
	bytes, err := p.Extract(context.Background(), strings.NewReader(input))
	require.NoError(t, err)
	assert.Equal(t, " title: Test \r\n", string(bytes))
	_, err = p.Extract(context.Background(), strings.NewReader("---\n---\nBody"))
	errortest.RequireError(t, err, commonerrors.ErrEmpty)
	_, err = p.Extract(context.Background(), strings.NewReader("Body"))
	errortest.RequireError(t, err, commonerrors.ErrNotFound)
}
