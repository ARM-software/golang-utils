/*
 * Copyright (C) 2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package frontmatter

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	core "github.com/ARM-software/golang-utils/utils/serialisation/frontmatter"
)

func TestDocumentParserForwarding(t *testing.T) {
	t.Parallel()

	p, err := NewDocumentParser(core.NewFormat(core.WithStart("---"), core.WithEnd("---")))
	require.NoError(t, err)
	p = p.Configure(WithNormaliseContent(true), WithDiscardIntroductoryComments(true))
	input := "<!-- intro -->\n---\ntitle: Test\n---\n\r\n Body\rNext\r\n"
	ctx := context.Background()

	metadata, content, err := p.Extract(ctx, strings.NewReader(input))
	require.NoError(t, err)
	assert.Equal(t, "title: Test\n", string(metadata))
	assert.Equal(t, "Body\nNext", string(content))

	metadataReader, contentReader, err := p.Parse(ctx, strings.NewReader(input))
	require.NoError(t, err)
	metadata, err = io.ReadAll(metadataReader)
	require.NoError(t, err)
	content, err = io.ReadAll(contentReader)
	require.NoError(t, err)
	assert.Equal(t, "title: Test\n", string(metadata))
	assert.Equal(t, "Body\nNext", string(content))

	_, _, err = p.Extract(ctx, strings.NewReader("---\ntitle: Test\n"))
	errortest.AssertError(t, err, commonerrors.ErrNotFound)
	_, _, err = p.Parse(ctx, strings.NewReader("---\ntitle: Test\n"))
	errortest.AssertError(t, err, commonerrors.ErrNotFound)
	_, err = NewDocumentParser()
	errortest.AssertError(t, err, commonerrors.ErrUndefined)
}
