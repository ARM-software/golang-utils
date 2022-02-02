package iconv

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/encoding/unicode"
)

func TestConvertBytes(t *testing.T) {

	// values are determined from http://www.skandissystems.com/en_us/charset.htm
	tests := []struct {
		src       string
		expected  string
		converter ICharsetConverter
	}{
		{"花间一壶酒，独酌无相亲。", "\xbb\xa8\xbc\xe4\xd2\xbb\xba\xf8\xbe\xc6\xa3\xac\xb6\xc0\xd7\xc3" +
			"\xce\xde\xcf\xe0\xc7\xd7\xa1\xa3", NewConverter(unicode.UTF8, simplifiedchinese.GBK)},
		{"A\u3000\u554a\u4e02\u4e90\u72dc\u7349\u02ca\u2588Z€", "A\xa1\xa1\xb0\xa1\x81\x40\x81\x80\xaa\x40\xaa\x80\xa8\x40\xa8\x80Z\x80", NewConverter(unicode.UTF8, simplifiedchinese.GBK)},
		{"花间一壶酒，独酌无相亲。", "\xbb\xa8\xbc\xe4\xd2\xbb\xba\xf8\xbe\xc6\xa3\xac\xb6\xc0\xd7\xc3" +
			"\xce\xde\xcf\xe0\xc7\xd7\xa1\xa3", NewConverter(unicode.UTF8, simplifiedchinese.GB18030)},
		{"\u0081\u00de\u00df\u00e0\u00e1\u00e2\u00e3\uffff\U00010000", "\x81\x30\x81\x31\x81\x30\x89\x37\x81\x30\x89\x38\xa8\xa4\xa8\xa2" +
			"\x81\x30\x89\x39\x81\x30\x8a\x30\x84\x31\xa4\x39\x90\x30\x81\x30", NewConverter(unicode.UTF8, simplifiedchinese.GB18030)},
		{"漢字", "\xba\x7e\xa6\x72", NewConverter(unicode.UTF8, traditionalchinese.Big5)},
		{"こんにちは、Pythonプログラミング", "\xa4\xb3\xa4\xf3\xa4\xcb\xa4\xc1\xa4\xcf\xa1\xa2Python\xa5\xd7\xa5\xed\xa5\xb0\xa5\xe9\xa5\xdf\xa5\xf3\xa5\xb0", NewConverter(unicode.UTF8, japanese.EUCJP)},
		{"a\xfe\xfeb", "a\ufffdb", NewConverter(simplifiedchinese.GBK, unicode.UTF8)},
		{"\x80", "€", NewConverter(simplifiedchinese.GB18030, unicode.UTF8)},
		{"\xba\x7e\xa6\x72", "漢字", NewConverter(traditionalchinese.Big5, unicode.UTF8)},
		{"\xa4\xb3\xa4\xf3\xa4\xcb\xa4\xc1\xa4\xcf\xa1\xa2Python\xa5\xd7\xa5\xed\xa5\xb0\xa5\xe9\xa5\xdf\xa5\xf3\xa5\xb0", "こんにちは、Pythonプログラミング", NewConverter(japanese.EUCJP, unicode.UTF8)},
		{"\xba\x7e\xa6\x72", "\xb4\xc1\xbb\xfa", NewConverter(traditionalchinese.Big5, japanese.EUCJP)},
		{"Hello world!", "Hello world!", NewConverter(unicode.UTF8, charmap.ISO8859_1)},
		{"Hello world!", "Hello world!", NewConverter(charmap.ISO8859_1, unicode.UTF8)},
		{"Gar\xe7on !", "Garçon !", NewConverter(charmap.Windows1252, unicode.UTF8)},
		{"Garçon !", "Gar\xe7on !", NewConverter(unicode.UTF8, charmap.Windows1252)},
	}
	for i := range tests {
		test := tests[i]
		t.Run(fmt.Sprintf("conversion_%v", test.converter.String()), func(t *testing.T) {

			dstStr, err := test.converter.ConvertString(test.src)
			require.NoError(t, err)
			require.Equal(t, test.expected, dstStr)

			dstBytes, err := test.converter.ConvertBytes([]byte(test.src))
			require.NoError(t, err)
			require.Equal(t, test.expected, string(dstBytes))

			dst := test.converter.Convert(strings.NewReader(test.src))
			bytes, err := ioutil.ReadAll(dst)
			require.NoError(t, err)
			require.Equal(t, test.expected, string(bytes))
		})
	}
}
