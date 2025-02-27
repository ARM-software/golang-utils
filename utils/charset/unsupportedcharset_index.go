/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package charset

import (
	"strings"

	"golang.org/x/text/encoding"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

var (
	unsupportedCharsets = []string{"UTF-7-IMAP"}
)

// GetUnsupported gets valid IANA charset encoding we know are not supported by golang but not reported as such.
func GetUnsupported(name string) (encoding.Encoding, error) {
	for i := range unsupportedCharsets {
		if strings.EqualFold(unsupportedCharsets[i], name) {
			return nil, nil
		}
	}
	return nil, commonerrors.New(commonerrors.ErrInvalid, "invalid encoding name")
}
