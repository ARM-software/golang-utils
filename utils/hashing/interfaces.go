package hashing

import "io"

//go:generate mockgen -destination=../mocks/mock_$GOPACKAGE.go -package=mocks github.com/ARM-software/golang-utils/utils/$GOPACKAGE IHash

type IHash interface {
	Calculate(reader io.Reader) (string, error)
	GetType() string
}
