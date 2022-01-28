package charset

import (
	"bufio"
	"fmt"
	"github.com/ARM-software/golang-utils/utils/charset/iconv"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"
	"io"
	"unicode/utf8"

	"github.com/gogs/chardet"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

// DetectTextEncoding returns best guess of encoding of given content.
func DetectTextEncoding(content []byte) (encoding.Encoding, string, error) {
	if utf8.Valid(content) {
		return LookupCharset("UTF-8")
	}

	result, err := chardet.NewTextDetector().DetectBest(content)
	if err != nil {
		return nil, "", err
	}

	return LookupCharset(result.Charset)
}

// DetectTextEncodingFromReader returns best guess of encoding of given reader content. Looks at the first 1024 bytes in the same way as https://pkg.go.dev/golang.org/x/net/html/charset#DetermineEncoding
func DetectTextEncodingFromReader(reader io.Reader) (encoding.Encoding, string, error) {
	bytes, err := bufio.NewReader(reader).Peek(1024)
	if err != nil {
		return nil, "", err
	}
	return DetectTextEncoding(bytes)
}

// LookupCharset returns the encoding with the specified charsetLabel, and its canonical
// name. Matching is case-insensitive and ignores
// leading and trailing whitespace.
func LookupCharset(charsetLabel string) (charsetEnc encoding.Encoding, charsetName string, err error) {
	charsetEnc, err = htmlindex.Get(charsetLabel)
	var tempErr error
	if err != nil {
		otherLabel, tempErr := getEncodingMapping().GetCanonicalName(charsetLabel)
		if tempErr == nil {
			charsetEnc, tempErr = htmlindex.Get(otherLabel)
		}
	}
	err=tempErr
	if err != nil {
		err = fmt.Errorf("%w: charset [%v] is not supported by go: %v", commonerrors.ErrUnsupported, charsetLabel, err.Error())
		return
	}
	charsetName, err = htmlindex.Name(charsetEnc)
	return
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
