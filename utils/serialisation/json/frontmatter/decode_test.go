/*
 * Copyright (C) 2026 Arm Limited or its affiliates and Contributors. All rights reserved.
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
	core "github.com/ARM-software/golang-utils/utils/serialisation/frontmatter"
)

func TestExtractDocumentWithDecode(t *testing.T) {
	ctx := context.Background()
	var metadata struct {
		Title string `json:"title"`
	}
	body, err := ExtractDocumentWithDecode(ctx, strings.NewReader("<!-- intro -->\n---json\n{\"title\": \"Test\"}\n---\n\r\n Body\r\n"), &metadata,
		core.WithDiscardIntroductoryComments(true), core.WithNormaliseContent(true))
	require.NoError(t, err)
	assert.Equal(t, "Test", metadata.Title)
	assert.Equal(t, "Body", string(body))

	var values map[string]any
	body, err = ExtractDocumentWithDecode(ctx, strings.NewReader("{\n\"title\": \"Test\", \"count\": 2\n}\n\n Body  text\r\n"), &values)
	require.NoError(t, err)
	assert.Len(t, values, 2)
	assert.Equal(t, "Test", values["title"])
	assert.Equal(t, float64(2), values["count"])
	assert.Equal(t, "\n Body  text\r\n", string(body))

	for _, input := range []string{"Body", ";;;\n;;;\nBody"} {
		metadata.Title = "original"
		body, err := ExtractDocumentWithDecode(ctx, strings.NewReader(input), &metadata)
		require.NoError(t, err)
		assert.Equal(t, "original", metadata.Title)
		assert.Equal(t, "Body", string(body))

		_, err = ExtractDocumentWithDecode(ctx, strings.NewReader(input), nil)
		require.NoError(t, err, "empty metadata must skip destination validation")
	}

	for _, test := range []struct {
		name, input string
		v           any
		expectedErr error
	}{
		{"invalid JSON", "---json\n{invalid}\n---\nBody", &metadata, commonerrors.ErrUnexpected},
		{"invalid destination", "---json\n{}\n---\nBody", nil, commonerrors.ErrUnexpected},
	} {
		t.Run(test.name, func(t *testing.T) {
			body, err := ExtractDocumentWithDecode(ctx, strings.NewReader(test.input), test.v)
			errortest.AssertError(t, err, test.expectedErr)
			assert.Equal(t, "Body", string(body))
		})
	}

	metadata.Title = "original"
	src := io.MultiReader(strings.NewReader("---json\n{\"title\": \"changed\"}\n---\npartial body"), iotest.ErrReader(io.ErrUnexpectedEOF))
	body, err = ExtractDocumentWithDecode(ctx, src, &metadata)
	errortest.AssertError(t, err, commonerrors.ErrEOF)
	assert.Equal(t, "partial body", string(body))
	assert.Equal(t, "original", metadata.Title, "extraction failure must prevent decoding")

	ctx, cancel := context.WithCancel(ctx)
	cancel()
	_, err = ExtractDocumentWithDecode(ctx, strings.NewReader("---json\n{\"title\": \"changed\"}\n---\nBody"), &metadata)
	errortest.AssertError(t, err, commonerrors.ErrCancelled)
	assert.Equal(t, "original", metadata.Title)
}

func ExampleExtractDocumentWithDecode() {
	var metadata struct {
		Title string `json:"title"`
	}

	body, err := ExtractDocumentWithDecode(context.Background(), strings.NewReader("---json\n{\"title\": \"Example\"}\n---\nDocument body.\n"), &metadata)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(metadata.Title)
	fmt.Print(string(body))
	// Output:
	// Example
	// Document body.
}
