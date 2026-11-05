/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package xml

import serialisationxml "github.com/ARM-software/golang-utils/utils/serialisation/xml"

// UnmarshallXML unmarshals XML bytes into value.
//
// It exists instead of using encoding/xml.Unmarshal directly because the
// decoder is configured with a charset reader, which allows non-UTF-8 XML
// encodings to be handled.
func UnmarshallXML(data []byte, value any) error {
	return serialisationxml.UnmarshallXML(data, value)
}
