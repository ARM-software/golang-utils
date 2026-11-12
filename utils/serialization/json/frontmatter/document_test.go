/*
 * Copyright (C) 2026 Arm Limited or its affiliates and Contributors. All rights reserved.
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

func TestExtractDocumentForwarding(t *testing.T) {
	t.Parallel()

	input := "<!-- intro -->\n---json\n{\"title\": \"Test\"}\n---\n\r\n Body\rNext\r\n"
	metadata, content, err := ExtractDocument(context.Background(), strings.NewReader(input), core.WithNormaliseContent(true), core.WithDiscardIntroductoryComments(true))
	require.NoError(t, err)
	assert.Equal(t, "{\"title\": \"Test\"}\n", string(metadata))
	assert.Equal(t, "Body\nNext", string(content))

	_, _, err = ExtractDocument(context.Background(), strings.NewReader("---json\n{}\n"))
	errortest.AssertError(t, err, commonerrors.ErrNotFound)
}

func TestExtractDocumentWithDecodeForwarding(t *testing.T) {
	t.Parallel()

	v := struct {
		Title string `json:"title"`
	}{}
	input := "<!-- intro -->\n---json\n{\"title\": \"Test\"}\n---\n\r\n Body\rNext\r\n"
	content, err := ExtractDocumentWithDecode(context.Background(), strings.NewReader(input), &v, core.WithNormaliseContent(true), core.WithDiscardIntroductoryComments(true))
	require.NoError(t, err)
	assert.Equal(t, "Test", v.Title)
	assert.Equal(t, "Body\nNext", string(content))

	content, err = ExtractDocumentWithDecode(context.Background(), strings.NewReader("---json\n{invalid}\n---\nBody"), &v)
	errortest.AssertError(t, err, commonerrors.ErrUnexpected)
	assert.Equal(t, "Body", string(content))
}

func TestExtractContentForwarding(t *testing.T) {
	t.Parallel()

	input := "<!-- intro -->\n---json\n{invalid}\n---\n\r\n Body\rNext\r\n"
	content, err := ExtractContent(context.Background(), strings.NewReader(input), core.WithNormaliseContent(true), core.WithDiscardIntroductoryComments(true))
	require.NoError(t, err)
	assert.Equal(t, "Body\nNext", string(content))

	_, err = ExtractContent(context.Background(), strings.NewReader("---json\n{}\n"))
	errortest.AssertError(t, err, commonerrors.ErrNotFound)
}
