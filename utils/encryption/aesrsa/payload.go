package aesrsa

import (
	"encoding/base64"
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/config"
)

type HybridAESRSAEncryptedPayload struct {
	// Ciphertext contains the encrypted contents
	CipherText string `json:"cipher_text" yaml:"cipher_text" mapstructure:"cipher_text"`
	// EncryptedKey contains the encryped AES key used to encrypt the data
	EncryptedKey string `json:"encrypted_key" yaml:"encrypted_key" mapstructure:"encrypted_key"`
	// Nonce used for encryption is required during decryption
	Nonce string `json:"nonce" yaml:"nonce" mapstructure:"nonce"`
}

func (p *HybridAESRSAEncryptedPayload) Validate() (err error) {
	err = config.ValidateEmbedded(p)
	if err != nil {
		return err
	}

	return validation.ValidateStruct(p,
		validation.Field(&p.CipherText, validation.Required),
		validation.Field(&p.EncryptedKey, validation.Required),
		validation.Field(&p.Nonce, validation.Required),
	)
}

func DecodeHybridAESRSAEncryptedPayload(p *HybridAESRSAEncryptedPayload) (cipher, key, nonce []byte, err error) {
	key, err = base64.StdEncoding.DecodeString(p.EncryptedKey)
	if err != nil {
		err = fmt.Errorf("%w: could not decode base64 encoded encrypted key %v", commonerrors.ErrUnexpected, err.Error())
		return
	}

	nonce, err = base64.StdEncoding.DecodeString(p.Nonce)
	if err != nil {
		err = fmt.Errorf("%w: could not decode base64 encoded nonce %v", commonerrors.ErrUnexpected, err.Error())
		return
	}

	cipher, err = base64.StdEncoding.DecodeString(p.CipherText)
	if err != nil {
		err = fmt.Errorf("%w: could not decode base64 encoded ciphertext %v", commonerrors.ErrUnexpected, err.Error())
		return
	}

	return
}

func EncodeHybridAESRSAEncryptedPayload(cipher, key, nonce []byte) (p *HybridAESRSAEncryptedPayload) {
	return &HybridAESRSAEncryptedPayload{
		EncryptedKey: base64.StdEncoding.EncodeToString(key),
		Nonce:        base64.StdEncoding.EncodeToString(nonce),
		CipherText:   base64.StdEncoding.EncodeToString(cipher),
	}
}
