// Package encryption defines utilities with regards to cryptography.
package encryption

import (
	"encoding/json"
	"fmt"
)

//go:generate go tool mockgen -destination=../mocks/mock_$GOPACKAGE.go -package=mocks github.com/ARM-software/golang-utils/utils/$GOPACKAGE IKeyPair

// IKeyPair defines an asymmetric key pair for cryptography.
type IKeyPair interface {
	fmt.Stringer
	fmt.GoStringer
	json.Marshaler
	// GetPublicKey returns the public key (base64 encoded)
	GetPublicKey() string
	// GetPrivateKey returns the private key (base64 encoded)
	GetPrivateKey() string
}
