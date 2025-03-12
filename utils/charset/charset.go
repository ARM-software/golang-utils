/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package charset

import (
	"bufio"
	"io"
	"unicode/utf8"

	"github.com/gogs/chardet"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"
	"golang.org/x/text/encoding/ianaindex"

	"github.com/ARM-software/golang-utils/utils/charset/iconv"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

// DetectTextEncoding returns best guess of encoding of given content.
func DetectTextEncoding(content []byte) (encoding.Encoding, string, error) {
	if utf8.Valid(content) {
		return LookupCharset("UTF-8")
	}

	result, err := chardet.NewTextDetector().DetectBest(content)
	if err != nil {
		return nil, "", commonerrors.WrapError(commonerrors.ErrNotFound, err, "")
	}

	return LookupCharset(result.Charset)
}

// DetectTextEncodingFromReader returns best guess of encoding of given reader content. Looks at the first 1024 bytes in the same way as https://pkg.go.dev/golang.org/x/net/html/charset#DetermineEncoding
func DetectTextEncodingFromReader(reader io.Reader) (encoding.Encoding, string, error) {
	bytes, err := bufio.NewReader(reader).Peek(1024)
	if !commonerrors.Any(err, nil, io.EOF, commonerrors.ErrEOF) {
		return nil, "", err
	}
	return DetectTextEncoding(bytes)
}

// LookupCharset returns the encoding with the specified charsetLabel, and its canonical
// name. Matching is case-insensitive and ignores
// leading and trailing whitespace.
func LookupCharset(charsetLabel string) (charsetEnc encoding.Encoding, charsetName string, err error) {
	charsetEnc, err = findCharsetEncoding(charsetLabel)
	if err != nil {
		if commonerrors.Any(err, commonerrors.ErrUnsupported) {
			err = commonerrors.WrapErrorf(commonerrors.ErrUnsupported, err, "charset [%v] is not supported by go", charsetLabel)
		} else {
			err = commonerrors.WrapErrorf(commonerrors.ErrInvalid, err, "charset [%v] is invalid", charsetLabel)
		}
		return
	}
	charsetName, err = htmlindex.Name(charsetEnc)
	if err == nil {
		return
	}
	charsetName, err = ianaindex.IANA.Name(charsetEnc)
	return
}

func findCharsetEncoding(charsetLabel string) (charsetEnc encoding.Encoding, err error) {
	// Check in http://www.w3.org/TR/encoding
	charsetEnc, err = findCharsetEncodingInAnIndex(htmlindex.Get, charsetLabel)
	if commonerrors.Any(err, nil, commonerrors.ErrUnsupported) {
		return
	}
	// Look at this index https://www.iana.org/assignments/character-sets/character-sets.xhtml
	charsetEnc, err = findCharsetEncodingInAnIndex(ianaindex.IANA.Encoding, charsetLabel)
	if commonerrors.Any(err, nil, commonerrors.ErrUnsupported) {
		return
	}
	// Look at the list of known unsupported charsets
	charsetEnc, err = findCharsetEncodingInAnIndex(GetUnsupported, charsetLabel)
	return
}

func findCharsetEncodingInAnIndex(indexSearch func(string) (encoding.Encoding, error), charsetLabel string) (charsetEnc encoding.Encoding, err error) {
	charsetEnc, err = checkEncodingSupport(indexSearch(charsetLabel))
	if commonerrors.Any(err, nil, commonerrors.ErrUnsupported) {
		return
	}
	otherLabel, err := getEncodingMapping().GetCanonicalName(charsetLabel)
	if err != nil {
		return
	}
	charsetEnc, err = checkEncodingSupport(indexSearch(otherLabel))
	return
}

func checkEncodingSupport(charsetEnc encoding.Encoding, err error) (encoding.Encoding, error) {
	// according to index documentation, if the error is nil but the encoding as well, then the encoding should be considered as unsupported by the language
	newErr := err
	if err == nil {
		if charsetEnc == nil {
			newErr = commonerrors.New(commonerrors.ErrUnsupported, "unsupported charset encoding")
		}
	}
	return charsetEnc, newErr
}

// IconvString converts string from one text encoding charset to another.
func IconvString(input string, fromEncoding encoding.Encoding, toEncoding encoding.Encoding) (string, error) {
	return iconv.NewConverter(fromEncoding, toEncoding).ConvertString(input)
}

// IconvStringFromLabels is similar to IconvString but uses labels.
func IconvStringFromLabels(input string, fromEncodingLabel string, toEncodingLabel string) (transformedText string, err error) {
	fromEncoding, _, err := LookupCharset(fromEncodingLabel)
	if err != nil {
		return
	}
	toEncoding, _, err := LookupCharset(toEncodingLabel)
	if err != nil {
		return
	}
	transformedText, err = IconvString(input, fromEncoding, toEncoding)
	return
}

// IconvBytes converts bytes from one text encoding charset to another.
func IconvBytes(input []byte, fromEncoding encoding.Encoding, toEncoding encoding.Encoding) ([]byte, error) {
	return iconv.NewConverter(fromEncoding, toEncoding).ConvertBytes(input)
}

// IconvBytesFromLabels is similar to IconvBytes but uses labels.
func IconvBytesFromLabels(input []byte, fromEncodingLabel string, toEncodingLabel string) (transformedBytes []byte, err error) {
	fromEncoding, _, err := LookupCharset(fromEncodingLabel)
	if err != nil {
		return
	}
	toEncoding, _, err := LookupCharset(toEncodingLabel)
	if err != nil {
		return
	}
	transformedBytes, err = IconvBytes(input, fromEncoding, toEncoding)
	return
}

// Iconv converts from any supported text encodings to any other, through Unicode conversion.
// Similar to https://www.gnu.org/software/libiconv/ but using pure go as opposed to many go libraries
func Iconv(reader io.Reader, fromEncoding encoding.Encoding, toEncoding encoding.Encoding) io.Reader {
	return iconv.NewConverter(fromEncoding, toEncoding).Convert(reader)
}

// IconvFromLabels is similar to Iconv but uses labels.
func IconvFromLabels(reader io.Reader, fromEncodingLabel string, toEncodingLabel string) (transformedReader io.Reader, err error) {
	fromEncoding, _, err := LookupCharset(fromEncodingLabel)
	if err != nil {
		return
	}
	toEncoding, _, err := LookupCharset(toEncodingLabel)
	if err != nil {
		return
	}
	transformedReader = Iconv(reader, fromEncoding, toEncoding)
	return
}
