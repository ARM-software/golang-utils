package iconv

import (
	"io"
)

//go:generate mockgen -destination=../mocks/mock_$GOPACKAGE.go -package=mocks github.com/ARM-software/golang-utils/utils/$GOPACKAGE ICharsetConverter

type ICharsetConverter interface {
	// ConvertString converts the charset of an input string
	ConvertString(input string) (string, error)

	// ConvertBytes converts the charset of an input byte array
	ConvertBytes(input []byte) ([]byte, error)

	// Convert converts the charset of a reader
	Convert(reader io.Reader) io.Reader

	// String describes the conversion
	String() string
}
