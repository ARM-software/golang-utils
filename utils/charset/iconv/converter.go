/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package iconv

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/ARM-software/golang-utils/utils/safeio"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// This is really similar to https://github.com/mushroomsir/iconv

func NewConverter(fromEncoding encoding.Encoding, toEncoding encoding.Encoding) ICharsetConverter {
	return &Converter{
		fromEncoding: fromEncoding,
		toEncoding:   toEncoding,
	}
}

type Converter struct {
	fromEncoding encoding.Encoding
	toEncoding   encoding.Encoding
}

func (t *Converter) ConvertStringWithContext(ctx context.Context, input string) (transformedStr string, err error) {
	res, err := t.ConvertBytesWithContext(ctx, []byte(input))
	if err != nil {
		return
	}
	transformedStr = string(res)
	return
}

func (t *Converter) ConvertBytesWithContext(ctx context.Context, input []byte) ([]byte, error) {
	reader := t.Convert(bytes.NewReader(input))
	return safeio.ReadAll(ctx, reader)
}

func (t *Converter) ConvertString(input string) (string, error) {
	return t.ConvertStringWithContext(context.Background(), input)
}

func (t *Converter) ConvertBytes(input []byte) ([]byte, error) {
	return t.ConvertBytesWithContext(context.Background(), input)
}

func (t *Converter) Convert(reader io.Reader) io.Reader {
	return t.convert(reader)
}

func (t *Converter) String() string {
	return fmt.Sprintf("%v to %v", t.fromEncoding, t.toEncoding)
}

func (t *Converter) convert(reader io.Reader) (resReader io.Reader) {
	if t.fromEncoding == unicode.UTF8 {
		resReader = bufio.NewReader(reader)
	} else {
		resReader = transform.NewReader(reader, t.fromEncoding.NewDecoder())
	}
	if t.toEncoding != unicode.UTF8 {
		resReader = transform.NewReader(resReader, t.toEncoding.NewEncoder())
	}
	return
}
