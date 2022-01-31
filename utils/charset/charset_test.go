package charset

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/charset/charsetaliases"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

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
			bytes, err := ioutil.ReadAll(dst)
			require.NoError(t, err)
			require.Equal(t, test.expected, string(bytes))
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
					assert.True(t, commonerrors.Any(err, commonerrors.ErrUnsupported), fmt.Sprintf("charset with alias [%v] could not be found", charset))
					assert.Empty(t, name)
					assert.Nil(t, encoding)
				}
			}
		})
	}
}
