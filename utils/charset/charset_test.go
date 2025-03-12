/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package charset

import (
	"context"
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/encoding/unicode"

	"github.com/ARM-software/golang-utils/utils/charset/charsetaliases"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/safeio"
)

func selectRandomUnsupportedCharset() string {
	random := rand.New(rand.NewSource(time.Now().Unix())) //nolint:gosec //causes G404: Use of weak random number generator (math/rand instead of crypto/rand) (gosec), So disable gosec as this is just for
	keys := reflect.ValueOf(charsetaliases.KnownUnsupportedIconvEncodings).MapKeys()
	return keys[random.Intn(len(keys))].Interface().(string) //nolint:gosec //causes G404: Use of weak random number generator (math/rand instead of crypto/rand) (gosec), So disable gosec
}

func selectRandomSupportedCharset() string {
	random := rand.New(rand.NewSource(time.Now().Unix())) //nolint:gosec //causes G404: Use of weak random number generator (math/rand instead of crypto/rand) (gosec), So disable gosec as this is just for
	// aliases := charsetaliases.ICUCharsetAliases
	aliases := []string{"csUTF8", "utf8", "iso-ir-138", "ISO_8859-8", "ISO-8859-8", "hebrew", "csISOLatinHebrew", "iso-ir-6", "ANSI_X3.4-1968", "ANSI_X3.4-1986", "ISO_646.irv:1991", "ISO646-US", "US-ASCII", "us", "IBM367", "cp367", "csASCII"}
	return aliases[random.Intn(len(aliases))] //nolint:gosec //causes G404: Use of weak random number generator (math/rand instead of crypto/rand) (gosec), So disable gosec
}

func TestIconv(t *testing.T) {
	// values are determined from http://www.skandissystems.com/en_us/charset.htm
	tests := []struct {
		src          string
		expected     string
		fromEncoding string
		toEncoding   string
	}{
		{"花间一壶酒，独酌无相亲。", "\xbb\xa8\xbc\xe4\xd2\xbb\xba\xf8\xbe\xc6\xa3\xac\xb6\xc0\xd7\xc3" +
			"\xce\xde\xcf\xe0\xc7\xd7\xa1\xa3", "utf8", "GBK"},
		{"A\u3000\u554a\u4e02\u4e90\u72dc\u7349\u02ca\u2588Z€", "A\xa1\xa1\xb0\xa1\x81\x40\x81\x80\xaa\x40\xaa\x80\xa8\x40\xa8\x80Z\x80", "Utf-8", "Gbk"},
		{"花间一壶酒，独酌无相亲。", "\xbb\xa8\xbc\xe4\xd2\xbb\xba\xf8\xbe\xc6\xa3\xac\xb6\xc0\xd7\xc3" +
			"\xce\xde\xcf\xe0\xc7\xd7\xa1\xa3", "utf8", "gb18030"},
		{"\u0081\u00de\u00df\u00e0\u00e1\u00e2\u00e3\uffff\U00010000", "\x81\x30\x81\x31\x81\x30\x89\x37\x81\x30\x89\x38\xa8\xa4\xa8\xa2" +
			"\x81\x30\x89\x39\x81\x30\x8a\x30\x84\x31\xa4\x39\x90\x30\x81\x30", "utf-8", "GB18030"},
		{"漢字", "\xba\x7e\xa6\x72", "UTF-8", "big5"},
		{"こんにちは、Pythonプログラミング", "\xa4\xb3\xa4\xf3\xa4\xcb\xa4\xc1\xa4\xcf\xa1\xa2Python\xa5\xd7\xa5\xed\xa5\xb0\xa5\xe9\xa5\xdf\xa5\xf3\xa5\xb0", "Utf-8", "euc-JP"},
		{"a\xfe\xfeb", "a\ufffdb", "gbk", "UTF8"},
		{"\x80", "€", "GB18030", "UTF8"},
		{"\xba\x7e\xa6\x72", "漢字", "Big5", "utf-8"},
		{"\xa4\xb3\xa4\xf3\xa4\xcb\xa4\xc1\xa4\xcf\xa1\xa2Python\xa5\xd7\xa5\xed\xa5\xb0\xa5\xe9\xa5\xdf\xa5\xf3\xa5\xb0", "こんにちは、Pythonプログラミング", "EUC-JP", "utf-8"},
		{"\xba\x7e\xa6\x72", "\xb4\xc1\xbb\xfa", "Big5", "EUC-JP"},
		{"Hello world!", "Hello world!", "UTF8", "ISO8859-1"},
		{"Hello world!", "Hello world!", "ISO-8859-1", "UTF8"},
		{"Gar\xe7on !", "Garçon !", "windows-1252", "UTF8"},
		{"Garçon !", "Gar\xe7on !", "utf-8", "WINDOWS-1252"},
	}
	for i := range tests {
		test := tests[i]
		t.Run(fmt.Sprintf("iconv_from_%v_to_%v", test.fromEncoding, test.toEncoding), func(t *testing.T) {

			dstStr, err := IconvStringFromLabels(test.src, test.fromEncoding, test.toEncoding)
			require.NoError(t, err)
			require.Equal(t, test.expected, dstStr)

			dstBytes, err := IconvBytesFromLabels([]byte(test.src), test.fromEncoding, test.toEncoding)
			require.NoError(t, err)
			require.Equal(t, test.expected, string(dstBytes))

			dst, err := IconvFromLabels(strings.NewReader(test.src), test.fromEncoding, test.toEncoding)
			require.NoError(t, err)
			bytes, err := safeio.ReadAll(context.TODO(), dst)
			require.NoError(t, err)
			require.Equal(t, test.expected, string(bytes))
		})
	}
}

func TestIconvWithUnsupportedCharset(t *testing.T) {
	input := "花间一壶酒，独酌无相亲。"
	tests := []struct {
		fromEncoding string
		toEncoding   string
	}{
		{
			fromEncoding: selectRandomSupportedCharset(),
			toEncoding:   selectRandomUnsupportedCharset(),
		},
		{
			fromEncoding: selectRandomUnsupportedCharset(),
			toEncoding:   selectRandomUnsupportedCharset(),
		},
		{
			fromEncoding: selectRandomUnsupportedCharset(),
			toEncoding:   selectRandomSupportedCharset(),
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(fmt.Sprintf("iconv_from_%v_to_%v", test.fromEncoding, test.toEncoding), func(t *testing.T) {

			dstStr, err := IconvStringFromLabels(input, test.fromEncoding, test.toEncoding)
			require.Error(t, err)
			assert.Empty(t, dstStr)

			dstBytes, err := IconvBytesFromLabels([]byte(input), test.fromEncoding, test.toEncoding)
			require.Error(t, err)
			assert.Empty(t, dstBytes)

			_, err = IconvFromLabels(strings.NewReader(input), test.fromEncoding, test.toEncoding)
			require.Error(t, err)
		})
	}
}

func TestLookUp(t *testing.T) {
	tests := []struct {
		list []string
		desc string
	}{
		{charsetaliases.IconvCharsetAliases, "Aliases from iconv built-in list"},
		{charsetaliases.ICUCharsetAliases, "Aliases from ICU/IANA list"},
		{charsetaliases.ICUCharsetNames, "Names from ICU/IANA list"},
	}
	for i := range tests {
		test := tests[i]
		t.Run(test.desc, func(t *testing.T) {
			for i := range test.list {
				charset := test.list[i]
				encoding, name, err := LookupCharset(charset)
				if err == nil {
					assert.NotEmpty(t, name)
					assert.NotNil(t, encoding)
				} else {
					if commonerrors.Any(err, commonerrors.ErrUnsupported) {
						assert.Error(t, err)
						errortest.AssertError(t, err, commonerrors.ErrUnsupported)
						assert.Empty(t, name)
						assert.Nil(t, encoding)
					} else {
						_, ok := charsetaliases.KnownUnsupportedIconvEncodings[charset]
						assert.True(t, ok, fmt.Sprintf("charset  [%v] is not known nor is considered as unsupported", charset))
					}
				}
			}
		})
	}
}

func TestDetectEncoding(t *testing.T) {
	// values are determined from http://www.skandissystems.com/en_us/charset.htm
	tests := []struct {
		src                  string
		expectedEncodingName string
		expectedEncoding     encoding.Encoding
	}{
		{"花间一壶酒，独酌无相亲。", "utf-8", unicode.UTF8},
		{"\xa1\xa1\xb0\xa1\x81\x40\x81\x80\xaa\x40\xaa\x80\xa8\x40\xa8\x80\x80" +
			"\x82\xb1\x82\xf1\x82\xc9\x82\xbf\x82\xcd\x81\x43\x90\xa2\x8a\x45\x81\x49\x20\x8e\x84\x82\xcc\x96\xbc" +
			"\x91\x4f\x82\xcd\x20\x53\x70\x69\x65\x67\x65\x6c\x20\x82\xc5\x82\xb7\x81\x42\x20", "shift_jis", japanese.ShiftJIS},
		{"\x81\x30\x81\x31\x81\x30\x89\x37\x81\x30\x89\x38\xa8\xa4\xa8\xa2" +
			"\x81\x30\x89\x39\x81\x30\x8a\x30\x84\x31\xa4\x39\x90\x30\x81\x30", "gb18030", simplifiedchinese.GB18030},
		{"\u0081\u00de\u00df\u00e0\u00e1\u00e2\u00e3\uffff\U00010000", "utf-8", unicode.UTF8},
		{"こんにちは、Pythonプログラミング", "utf-8", unicode.UTF8},
		{"漢字", "utf-8", unicode.UTF8},
		{"\xa4\xb3\xa4\xf3\xa4\xcb\xa4\xc1\xa4\xcf\xa1\xa2Python\xa5\xd7\xa5\xed\xa5\xb0\xa5\xe9\xa5\xdf\xa5\xf3\xa5\xb0", "euc-jp", japanese.EUCJP},
		{"\xba\x7e\xa6\x72" +
			"\x8a\x30\x84\x31\xa4\x39\x90\x30\x81\x30\x81\x30\x89\x37\x81\x30\x89\x38\xa8", "gb18030", simplifiedchinese.GB18030},
		{"\xa4\xa4\xae\xc9\xb9\x71\xa4\x6c\xb3\xf8\xa1\x47\xb4\xa3\xa8\xd1\xa4\xa4\xb0\xea\xae\xc9\xb3\xf8\xa1\x42" +
			"\xa4\x75\xb0\xd3\xae\xc9\xb3\xf8\xa1\x42\xa9\xf4\xb3\xf8\xb3\xcc\xb8\xd4\xba\xc9\xaa\xba\xb7\x73\xbb\x44\xb8" +
			"\xea\xb0\x54\xa1\x41\xac\x46\xaa\x76\xa1\x42\xb0\x5d\xb8\x67\xa1\x42\xaa\xd1\xa5\xab\xa1\x42\xaa\xc0\xb7\x7c" +
			"\xa1\x42\xb0\xea\xbb\xda\xa1\x42\xa6\x61\xa4\xe8\xa1\x42\xac\xec\xa7\xde\xa1\x42\xae\x54\xbc\xd6\xa1\x42\xc5\xe9" +
			"\xa8\x7c\xa1\x42\xc3\xc0\xa4\xe5\xa1\x42\xa5\xcd\xac\xa1\xa1\x42\xae\xc8\xb9\x43\xa1\x42\xbc\x76\xad\xb5\xb8\x60" +
			"\xa5\xd8\xa1\x42\xb3\xa1\xb8\xa8\xae\xe6\xb5\x4c\xa9\xd2\xa4\xa3\xa5\x5d", "big5", traditionalchinese.Big5},
		{"\xb9\xde\xc0\xba\x20\xb8\xde\xc0\xcf\xc7\xd4\x20\xba\xb8\xb1\xe2" +
			"\xbd\xce\xc0\xcc\xbf\xf9\xb5\xe5\x20\xb9\xc2\xc1\xf7", "euc-kr", korean.EUCKR},
		{"Garçon !", "utf-8", unicode.UTF8},
		{"Gar\xe7on !", "windows-1252", charmap.Windows1252},
		{"\x4e\x6f\x75\x73\x20\x76\x6f\x75\x73\x20\x74\x72\x61\x6e\x73\x6d\x65\x74\x74\x72\x6f" +
			"\x6e\x73\x20\x6c\x65\x73\x20\x69\x6e\x66\x6f\x72\x6d\x61\x74\x69\x6f\x6e\x73\x20\x64\x65\x6d\x61\x6e\x64" +
			"\xe9\x65\x73\x20\x64\x61\x6e\x73\x20\x6c\x65\x73\x20\x6d\x65\x69\x6c\x6c\x65\x75\x72\x73\x20\x64\xe9\x6c" +
			"\x61\x69\x73\x2e\x20\x43\x65", "windows-1252", charmap.Windows1252},
	}
	for i := range tests {
		test := tests[i]
		t.Run(fmt.Sprintf("detect_encoding_%v", test.expectedEncodingName), func(t *testing.T) {
			encoding, charsetName, err := DetectTextEncoding([]byte(test.src))
			require.NoError(t, err)
			assert.Equal(t, test.expectedEncoding, encoding)
			assert.Equal(t, test.expectedEncodingName, charsetName)

			encoding, charsetName, err = DetectTextEncodingFromReader(strings.NewReader(test.src))
			require.NoError(t, err)
			assert.Equal(t, test.expectedEncoding, encoding)
			assert.Equal(t, test.expectedEncodingName, charsetName)
		})
	}
}

func TestDetectEncodingError(t *testing.T) {
	input := "\x80" // Input is too small for detecting its charset.
	encoding, charsetName, err := DetectTextEncoding([]byte(input))
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrNotFound)
	assert.Empty(t, charsetName)
	assert.Empty(t, encoding)
}
