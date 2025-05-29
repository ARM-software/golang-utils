/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package iconv provides utilities to convert characters into different charsets
package iconv

import (
	"context"
	"io"
)

//go:generate go tool mockgen -destination=../../mocks/mock_$GOPACKAGE.go -package=mocks github.com/ARM-software/golang-utils/utils/charset/$GOPACKAGE ICharsetConverter

type ICharsetConverter interface {
	// ConvertString converts the charset of an input string
	ConvertString(input string) (string, error)

	// ConvertStringWithContext converts the charset of an input string
	ConvertStringWithContext(ctx context.Context, input string) (string, error)

	// ConvertBytes converts the charset of an input byte array
	ConvertBytes(input []byte) ([]byte, error)

	// ConvertBytesWithContext converts the charset of an input byte array
	ConvertBytesWithContext(ctx context.Context, input []byte) ([]byte, error)

	// Convert converts the charset of a reader
	Convert(reader io.Reader) io.Reader

	// String describes the conversion
	String() string
}
