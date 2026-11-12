/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package frontmatter

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	core "github.com/ARM-software/golang-utils/utils/serialisation/frontmatter"
)

func TestExtractDocument(t *testing.T) {
	for _, test := range []struct {
		name, input, metadata, body string
	}{
		{"Hugo lookahead", "{\n\"title\": \"Test\"\n}\n\nBody  text\n", "{\n\"title\": \"Test\"\n}\n", "\nBody  text\n"},
		{"Hexo", ";;;\n{\"title\": \"Test\"}\n;;;\nBody", "{\"title\": \"Test\"}\n", "Body"},
		{"tagged", "---json\n{\"title\": \"Test\"}\n---\nBody", "{\"title\": \"Test\"}\n", "Body"},
		{"preamble and CRLF", "\xef\xbb\xbf<!-- intro -->\r\n{\r\n}\r\n\r\nBody\r\n", "{\r\n}\r\n", "\xef\xbb\xbf<!-- intro -->\r\n\r\nBody\r\n"},
		{"absent", "<!-- intro -->\nBody\r\n", "", "<!-- intro -->\nBody\r\n"},
		{"empty metadata", "---json\n---\nBody", "", "Body"},
		{"empty document", "", "", ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			metadata, body, err := ExtractDocument(context.Background(), strings.NewReader(test.input))
			require.NoError(t, err)
			assert.Equal(t, test.metadata, string(metadata))
			assert.Equal(t, test.body, string(body))
		})
	}

	metadata, body, err := ExtractDocument(context.Background(), strings.NewReader("<!-- intro -->\n;;;\n{}\n;;;\n\r\n Body\r\n"),
		core.WithDiscardIntroductoryComments(true), core.WithNormaliseContent(true))
	require.NoError(t, err)
	assert.Equal(t, "{}\n", string(metadata))
	assert.Equal(t, "Body", string(body))

	_, _, err = ExtractDocument(context.Background(), strings.NewReader("---json\n{}\n"))
	errortest.AssertError(t, err, commonerrors.ErrNotFound)
}

func TestExtractContent(t *testing.T) {
	const input = "<!-- intro -->\n---json\n{invalid}\n---\n\r\n Body\r\n"
	body, err := ExtractContent(context.Background(), strings.NewReader(input),
		core.WithDiscardIntroductoryComments(true), core.WithNormaliseContent(true))
	require.NoError(t, err)
	assert.Equal(t, "Body", string(body))

	const unclosed = "---json\n{}\n"
	_, _, documentErr := ExtractDocument(context.Background(), strings.NewReader(unclosed))
	_, err = ExtractContent(context.Background(), strings.NewReader(unclosed))
	errortest.AssertError(t, documentErr, commonerrors.ErrNotFound)
	errortest.AssertError(t, err, commonerrors.ErrNotFound)
}
