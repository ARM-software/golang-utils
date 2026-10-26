/*
 * Copyright (C) 2020-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package serialization //nolint:misspell

import (
	serialisation "github.com/ARM-software/golang-utils/utils/serialisation"
	serialisationxml "github.com/ARM-software/golang-utils/utils/serialisation/xml" //nolint:misspell
)

// UnmarshallXML unmarshals XML bytes into value.
//
// This compatibility wrapper preserves the historical serialisation package API
// while delegating the actual implementation to the dedicated xml subproject.
func UnmarshallXML(data []byte, value any) error {
	return serialisation.UnmarshallXML(data, value)
}

// StripLeadingXMLHeaderComment removes a generated leading XML comment block
// when it contains the repository copyright or SPDX header markers.
//
// This is intended for XML-like fixtures or serialised content where an
// automated header insertion step prepends a licence comment before the real
// document content. Comments that do not contain the recognised header markers
// are preserved unchanged.
func StripLeadingXMLHeaderComment(data []byte) []byte {
	return serialisationxml.StripLeadingXMLHeaderComment(data)
}
