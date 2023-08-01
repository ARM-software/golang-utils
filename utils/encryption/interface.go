// Package encryption defines utilities with regards to cryptography.
package encryption

//go:generate mockgen -destination=../mocks/mock_$GOPACKAGE.go -package=mocks github.com/ARM-software/golang-utils/utils/$GOPACKAGE IKeyPair

// IKeyPair defines an asymmetric key pair for cryptography.
type IKeyPair interface {
	// GetPublicKey returns the public key (base64 encoded)
	GetPublicKey() string
	// GetPrivateKey returns the private key (base64 encoded)
	GetPrivateKey() string
}
