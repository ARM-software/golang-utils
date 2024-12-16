/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package serialization //nolint:misspell

import (
	"bytes"
	"encoding/xml"

	"golang.org/x/net/html/charset"
)

// UnmarshallXML was introduced instead
// of using xml.Unmarshal() as this only supports UTF8
// But it's been noticed that UnmarshalXml doesn't support UTF16
func UnmarshallXML(data []byte, value interface{}) error {
	// Read the XML file and create an in-memory model constructed from the
	// elements in the data
	reader := bytes.NewReader(data)
	decoder := xml.NewDecoder(reader)

	decoder.CharsetReader = charset.NewReaderLabel
	return decoder.Decode(&value)
}
