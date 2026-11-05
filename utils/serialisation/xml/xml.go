/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package xml

import (
	"bytes"
	"encoding/xml"

	"golang.org/x/net/html/charset"
)

var xmlHeaderCommentPrefixes = [][]byte{
	[]byte("Copyright (C)"),
	[]byte("SPDX-License-Identifier:"),
}

var (
	xmlCommentStart = []byte("<!--")
	xmlCommentEnd   = []byte("-->")
)

// UnmarshallXML unmarshals XML bytes into value.
//
// It exists instead of using encoding/xml.Unmarshal directly because the
// decoder is configured with a charset reader, which allows non-UTF-8 XML
// encodings to be handled.
func UnmarshallXML(data []byte, value any) error {
	data = StripLeadingXMLHeaderComment(data)

	reader := bytes.NewReader(data)
	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = charset.NewReaderLabel
	return decoder.Decode(&value)
}

// StripLeadingXMLHeaderComment removes a generated leading XML comment block
// when it contains the repository copyright or SPDX header markers.
//
// This is intended for XML-like fixtures or serialised content where an
// automated header insertion step prepends a licence comment before the real
// document content. Comments that do not contain the recognised header markers
// are preserved unchanged.
func StripLeadingXMLHeaderComment(data []byte) []byte {
	commentStart := bytes.Index(data, xmlCommentStart)
	if commentStart == -1 {
		return data
	}

	commentEnd := bytes.Index(data[commentStart:], xmlCommentEnd)
	if commentEnd == -1 {
		return data
	}
	commentEnd += commentStart + len(xmlCommentEnd)

	comment := data[commentStart:commentEnd]
	hasHeaderPrefix := false
	for _, prefix := range xmlHeaderCommentPrefixes {
		if bytes.Contains(comment, prefix) {
			hasHeaderPrefix = true
			break
		}
	}
	if !hasHeaderPrefix {
		return data
	}

	trimmedCommentStart := commentStart
	for trimmedCommentStart > 0 {
		switch data[trimmedCommentStart-1] {
		case ' ', '\t', '\r', '\n':
			trimmedCommentStart--
		default:
			goto trimTrailing
		}
	}

trimTrailing:
	trimmedCommentEnd := commentEnd
	for trimmedCommentEnd < len(data) {
		switch data[trimmedCommentEnd] {
		case ' ', '\t', '\r', '\n':
			trimmedCommentEnd++
		default:
			goto rebuild
		}
	}

rebuild:
	cleaned := make([]byte, 0, len(data)-(trimmedCommentEnd-trimmedCommentStart))
	cleaned = append(cleaned, data[:trimmedCommentStart]...)
	cleaned = append(cleaned, data[trimmedCommentEnd:]...)
	return cleaned
}
