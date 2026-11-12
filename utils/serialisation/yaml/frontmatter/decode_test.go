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
	core "github.com/ARM-software/golang-utils/utils/serialisation/frontmatter"
)

func TestExtractDocumentWithDecode(t *testing.T) {
	ctx := context.Background()
	var card ModelCard
	body, err := ExtractDocumentWithDecode(ctx, strings.NewReader("<!-- intro -->\n---\ndatasets: example\n---\n\r\n Body\r\n"), &card,
		core.WithDiscardIntroductoryComments(true), core.WithNormaliseContent(true))
	require.NoError(t, err)
	assert.Equal(t, "example", card.Datasets)
	assert.Equal(t, "Body", string(body))

	var values map[string]any
	body, err = ExtractDocumentWithDecode(ctx, strings.NewReader("---\ntitle: Test\ncount: 2\n---\n\r\n Body  text\r\n"), &values)
	require.NoError(t, err)
	assert.Len(t, values, 2)
	assert.Equal(t, "Test", values["title"])
	assert.Equal(t, float64(2), values["count"])
	assert.Equal(t, "\r\n Body  text\r\n", string(body))

	for _, input := range []string{"Body", "---\n---\nBody"} {
		card := ModelCard{Datasets: "original"}
		body, err := ExtractDocumentWithDecode(ctx, strings.NewReader(input), &card)
		require.NoError(t, err)
		assert.Equal(t, "original", card.Datasets)
		assert.Equal(t, "Body", string(body))

		_, err = ExtractDocumentWithDecode(ctx, strings.NewReader(input), nil)
		require.NoError(t, err, "empty metadata must skip destination validation")
	}

	for _, test := range []struct {
		name, input string
		v           any
		expectedErr error
	}{
		{"invalid YAML", "---\ntitle: [invalid\n---\nBody", &ModelCard{}, commonerrors.ErrMarshalling},
		{"invalid destination", "---\ntitle: Test\n---\nBody", nil, commonerrors.ErrUnexpected},
	} {
		t.Run(test.name, func(t *testing.T) {
			body, err := ExtractDocumentWithDecode(ctx, strings.NewReader(test.input), test.v)
			errortest.AssertError(t, err, test.expectedErr)
			assert.Equal(t, "Body", string(body))
		})
	}

	card = ModelCard{Datasets: "original"}
	src := io.MultiReader(strings.NewReader("---\ndatasets: changed\n---\npartial body"), iotest.ErrReader(io.ErrUnexpectedEOF))
	body, err = ExtractDocumentWithDecode(ctx, src, &card)
	errortest.AssertError(t, err, commonerrors.ErrEOF)
	assert.Equal(t, "partial body", string(body))
	assert.Equal(t, "original", card.Datasets, "extraction failure must prevent decoding")

	ctx, cancel := context.WithCancel(ctx)
	cancel()
	_, err = ExtractDocumentWithDecode(ctx, strings.NewReader("---\ndatasets: changed\n---\nBody"), &card)
	errortest.AssertError(t, err, commonerrors.ErrCancelled)
	assert.Equal(t, "original", card.Datasets)
}

func ExampleExtractDocumentWithDecode() {
	var metadata struct {
		Title string `json:"title"`
	}

	body, err := ExtractDocumentWithDecode(context.Background(), strings.NewReader("---\ntitle: Example\n---\nDocument body.\n"), &metadata)
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
