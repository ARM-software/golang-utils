/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package serialization //nolint:misspell

import serializationxml "github.com/ARM-software/golang-utils/utils/serialization/xml" //nolint:misspell

// UnmarshallXML unmarshals XML bytes into value.
//
// This compatibility wrapper preserves the historical serialization package API
// while delegating the actual implementation to the dedicated xml subproject.
func UnmarshallXML(data []byte, value any) error {
	return serializationxml.UnmarshallXML(data, value)
}
