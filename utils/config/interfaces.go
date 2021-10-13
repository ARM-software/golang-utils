package config

//go:generate mockgen -destination=../mocks/mock_$GOPACKAGE.go -package=mocks github.com/ARM-software/golang-utils/utils/$GOPACKAGE IServiceConfiguration

type IServiceConfiguration interface {
	// Validates configuration entries.
	Validate() error
}
