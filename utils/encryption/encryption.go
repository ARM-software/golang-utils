package encryption

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/nacl/box"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

const KeySize = 32

var (
	keySizeError = fmt.Errorf("%w: recipient key has invalid size (expected %d bytes)", commonerrors.ErrInvalid, KeySize)
)

// GenerateKeyPair generates a asymmetric key pair suitable for use with encryption utilities. Works with [NaCl box](https://nacl.cr.yp.to/box.html.)
func GenerateKeyPair() (base64EncodedPublicKey, base64EncodedPrivateKey string, err error) {
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		err = fmt.Errorf("%w: could not generate keys: %v", commonerrors.ErrUnexpected, err.Error())
		return
	}

	base64EncodedPublicKey = base64.StdEncoding.EncodeToString((*pub)[:])
	base64EncodedPrivateKey = base64.StdEncoding.EncodeToString((*priv)[:])
	return
}

// EncryptWithPublicKey encrypts small messages using a 32-byte public key (See https://libsodium.gitbook.io/doc/public-key_cryptography/sealed_boxes)
func EncryptWithPublicKey(base64EncodedPublicKey string, message string) (encryptedBase64Message string, err error) {
	decodedPublicKey, err := base64.StdEncoding.DecodeString(base64EncodedPublicKey)
	if err != nil {
		err = base64DecodingError(err)
		return
	}
	if len(decodedPublicKey) != KeySize {
		err = keySizeError
		return
	}

	recipientKey := [KeySize]byte{}
	copy(recipientKey[:], decodedPublicKey)

	secretBytes := []byte(message)

	encryptedBytes, err := box.SealAnonymous([]byte{}, secretBytes, &recipientKey, rand.Reader)
	if err != nil {
		err = fmt.Errorf("%w: box.SealAnonymous failed with error: %v", commonerrors.ErrUnexpected, err.Error())
		return
	}

	encryptedBase64Message = base64.StdEncoding.EncodeToString(encryptedBytes)

	return
}

func base64DecodingError(err error) error {
	return fmt.Errorf("%w: base64.StdEncoding.DecodeString was unable to decode string: %v", commonerrors.ErrInvalid, err.Error())
}

// DecryptWithKeyPair decrypts small base64 encoded messages
func DecryptWithKeyPair(base64EncodedPublicKey, base64EncodedPrivateKey, base64EncodedEncryptedMessage string) (decryptedMessage string, err error) {
	decodedPublicKey, err := base64.StdEncoding.DecodeString(base64EncodedPublicKey)
	if err != nil {
		err = base64DecodingError(err)
		return
	}
	if len(decodedPublicKey) != KeySize {
		err = keySizeError
		return
	}
	decodedPrivateKey, err := base64.StdEncoding.DecodeString(base64EncodedPrivateKey)
	if err != nil {
		err = base64DecodingError(err)
		return
	}
	if len(decodedPrivateKey) != KeySize {
		err = keySizeError
		return
	}

	decodedMessage, err := base64.StdEncoding.DecodeString(base64EncodedEncryptedMessage)
	if err != nil {
		err = base64DecodingError(err)
		return
	}

	publicKey := [KeySize]byte{}
	copy(publicKey[:], decodedPublicKey)

	privateKey := [KeySize]byte{}
	copy(privateKey[:], decodedPrivateKey)

	message, ok := box.OpenAnonymous([]byte{}, decodedMessage, &publicKey, &privateKey)
	if !ok {
		err = fmt.Errorf("%w: message could not be decrypted", commonerrors.ErrInvalid)
		return
	}
	decryptedMessage = string(message)
	return
}
