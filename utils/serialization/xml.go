/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package serialization //nolint:misspell

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

// UnmarshallXML was introduced instead
// of using xml.Unmarshal() as this only supports UTF8
// But it's been noticed that UnmarshalXml doesn't support UTF16
func UnmarshallXML(data []byte, value any) error {
	data = StripLeadingXMLHeaderComment(data)

	// Read the XML file and create an in-memory model constructed from the
	// elements in the data
	reader := bytes.NewReader(data)
	decoder := xml.NewDecoder(reader)

	decoder.CharsetReader = charset.NewReaderLabel
	return decoder.Decode(&value)
}

// StripLeadingXMLHeaderComment removes a generated leading XML comment block
// when it contains the repository copyright or SPDX header markers.
//
// This is intended for XML-like fixtures or serialized content where an
// automated header insertion step prepends a license comment before the real
// document content. Comments that do not contain the recognized header markers
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
