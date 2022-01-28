package iconv

import (
	"bufio"
	"bytes"
	"fmt"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
	"io"
	"io/ioutil"
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

func (t *Converter) ConvertString(input string) (transformedStr string, err error) {
	res, err := t.ConvertBytes([]byte(input))
	if err != nil {
		return
	}
	transformedStr = string(res)
	return
}

func (t *Converter) ConvertBytes(input []byte) ([]byte, error) {
	reader := t.Convert(bytes.NewReader(input))
	return ioutil.ReadAll(reader)
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
