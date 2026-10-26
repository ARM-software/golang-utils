/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package serialisation //nolint:misspell

import serialisationxml "github.com/ARM-software/golang-utils/utils/serialisation/xml" //nolint:misspell

// UnmarshallXML unmarshals XML bytes into value.
//
// This convenience wrapper delegates to the XML implementation in the
// serialisation/xml package.
func UnmarshallXML(data []byte, value any) error {
	return serialisationxml.UnmarshallXML(data, value)
}
