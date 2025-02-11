package aesrsa

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/filesystem"
)

// ParsePEMBlock will parse the first PEM block found within path
func ParsePEMBlock(path string) (block *pem.Block, err error) {
	if path == "" {
		err = fmt.Errorf("%w: no certificate provided", commonerrors.ErrUndefined)
		return
	}

	if !filesystem.Exists(path) {
		err = fmt.Errorf("%w: could not find certificate at '%v'", commonerrors.ErrNotFound, path)
		return
	}

	certBytes, err := filesystem.ReadFile(path)
	if err != nil {
		err = fmt.Errorf("%w: failed to read certificate: %v", commonerrors.ErrUnexpected, err.Error())
		return
	}

	block, _ = pem.Decode(certBytes)
	if block == nil {
		err = fmt.Errorf("%w: failed to decode PEM block from certificate", commonerrors.ErrUnexpected)
		return
	}

	return
}

// DecryptHybridAESRSAEncryptedPayload takes a path to an RSA private key and uses it to decode the AES key in a
// hybrid encoded payload. This AES key is then used to decode the actual payload contents. Information of the use
// of hybrid AES RSA encryption can be found here https://www.ijrar.org/papers/IJRAR23B1852.pdf
func DecryptHybridAESRSAEncryptedPayload(privateKeyPath string, payload *HybridAESRSAEncryptedPayload) (decrypted []byte, err error) {
	block, err := ParsePEMBlock(privateKeyPath)
	if err != nil {
		err = fmt.Errorf("%w: could not parse PEM block from '%v': %v", commonerrors.ErrUnexpected, privateKeyPath, err.Error())
		return
	}

	if block == nil {
		err = fmt.Errorf("%w: block was empty", commonerrors.ErrEmpty)
		return
	}

	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		err = fmt.Errorf("%w: could not parse private key at '%v' %v", commonerrors.ErrUnexpected, privateKeyPath, err.Error())
		return
	}

	ciphertext, encryptedKey, nonce, err := DecodeHybridAESRSAEncryptedPayload(payload)
	if err != nil {
		err = fmt.Errorf("%w: could not decode payload: %v", commonerrors.ErrUnexpected, err.Error())
		return
	}

	aesKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, priv, encryptedKey, []byte(""))
	if err != nil {
		err = fmt.Errorf("%w: could not decrypt private key %v", commonerrors.ErrUnexpected, err.Error())
		return
	}

	blockCipher, err := aes.NewCipher(aesKey)
	if err != nil {
		err = fmt.Errorf("%w: could not create new cipher %v", commonerrors.ErrUnexpected, err.Error())
		return
	}

	aesGCM, err := cipher.NewGCM(blockCipher)
	if err != nil {
		err = fmt.Errorf("%w: could not create new GCM %v", commonerrors.ErrUnexpected, err.Error())
		return
	}

	decrypted, err = aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		err = fmt.Errorf("%w: could not open ciphertext %v", commonerrors.ErrUnexpected, err.Error())
		return
	}

	return
}

// EncryptHybridAESRSAEncryptedPayload takes a path to a valid x509 certificate for key encypherment and uses it to
// encode a payload using hybrid RSA AES encryption where an AES key is used to encrypt the content in payload and the
// AES key is encrypted using RSA encryption. AES encryption is used to encode the payload itself as it is faster than
// RSA for larger payloads. RSA is used to encrypt the relatively small AES key and allows asymmetric encryption
// whilst also being fast. More information can be found at https://www.ijrar.org/papers/IJRAR23B1852.pdf
func EncryptHybridAESRSAEncryptedPayload(certPath string, payload []byte) (encrypted *HybridAESRSAEncryptedPayload, err error) {
	block, err := ParsePEMBlock(certPath)
	if err != nil {
		err = fmt.Errorf("%w: could not parse PEM block from '%v': %v", commonerrors.ErrUnexpected, certPath, err.Error())
		return
	}

	if block == nil {
		err = fmt.Errorf("%w: block was empty", commonerrors.ErrEmpty)
		return
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		err = fmt.Errorf("%w: failed parsing certificate: %v", commonerrors.ErrUnexpected, err.Error())
		return
	}

	rsaPub, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		err = fmt.Errorf("%w: public key in certificate is not RSA", commonerrors.ErrInvalid)
		return
	}

	aesKey := make([]byte, 32)
	_, err = rand.Read(aesKey)
	if err != nil {
		err = fmt.Errorf("%w: failed generating AES key: %v", commonerrors.ErrUnexpected, err.Error())
		return
	}

	blockCipher, err := aes.NewCipher(aesKey)
	if err != nil {
		err = fmt.Errorf("%w: failed to create cipher: %v", commonerrors.ErrUnexpected, err.Error())
		return
	}

	aesGCM, err := cipher.NewGCM(blockCipher)
	if err != nil {
		err = fmt.Errorf("%w: failed creating GCM: %v", commonerrors.ErrUnexpected, err.Error())
		return
	}

	nonce := make([]byte, aesGCM.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		err = fmt.Errorf("%w: failed to generate nonce: %v", commonerrors.ErrUnexpected, err.Error())
		return
	}

	ciphertext := aesGCM.Seal(nil, nonce, payload, nil)
	encryptedAESKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, rsaPub, aesKey, []byte{})
	if err != nil {
		err = fmt.Errorf("%w: could not complete RSA encryption: %v", commonerrors.ErrUnexpected, err.Error())
		return
	}

	encrypted = EncodeHybridAESRSAEncryptedPayload(ciphertext, encryptedAESKey, nonce)
	return
}
