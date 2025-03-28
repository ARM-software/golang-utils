package aesrsa

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/filesystem"
)

// ParsePEMBlock will parse the first PEM block found within path
func ParsePEMBlock(path string) (block *pem.Block, err error) {
	if path == "" {
		err = commonerrors.New(commonerrors.ErrUndefined, "no certificate provided")
		return
	}

	if !filesystem.Exists(path) {
		err = commonerrors.Newf(commonerrors.ErrNotFound, "could not find certificate at '%v'", path)
		return
	}

	certBytes, err := filesystem.ReadFile(path)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrUnexpected, err, "failed to read certificate")
		return
	}

	block, _ = pem.Decode(certBytes)
	if block == nil {
		err = commonerrors.New(commonerrors.ErrUnexpected, "failed to decode PEM block from certificate")
		return
	}

	return
}

// DecryptHybridAESRSAEncryptedPayloadFromPrivateKey takes a path to an RSA private key and uses it to decode the AES
// key in a hybrid encoded payload. This AES key is then used to decode the actual payload contents. Information of
// the use of hybrid AES RSA encryption can be found here https://www.ijrar.org/papers/IJRAR23B1852.pdf
func DecryptHybridAESRSAEncryptedPayloadFromPrivateKey(privateKeyPath string, payload *HybridAESRSAEncryptedPayload) (decrypted []byte, err error) {
	block, err := ParsePEMBlock(privateKeyPath)
	if err != nil {
		err = commonerrors.WrapErrorf(commonerrors.ErrUnexpected, err, "could not parse PEM block from '%v'", privateKeyPath)
		return
	}

	if block == nil {
		err = commonerrors.New(commonerrors.ErrEmpty, "block was empty")
		return
	}

	return DecryptHybridAESRSAEncryptedPayloadFromBytes(block.Bytes, payload)
}

// DecryptHybridAESRSAEncryptedPayloadFromBytes takes a path to an RSA private key PEM file and uses it to
// decode the AES key in a hybrid encoded payload. This AES key is then used to decode the actual payload contents.
// Information of the use of hybrid AES RSA encryption can be found here https://www.ijrar.org/papers/IJRAR23B1852.pdf
func DecryptHybridAESRSAEncryptedPayloadFromBytes(block []byte, payload *HybridAESRSAEncryptedPayload) (decrypted []byte, err error) {
	if payload == nil {
		err = commonerrors.New(commonerrors.ErrUndefined, "payload must not be nil")
		return
	}

	priv, err := x509.ParsePKCS1PrivateKey(block)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrUnexpected, err, "could not parse private key")
		return
	}

	ciphertext, encryptedKey, nonce, err := DecodeHybridAESRSAEncryptedPayload(payload)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrUnexpected, err, "could not decode payload")
		return
	}

	aesKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, priv, encryptedKey, []byte{})
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrUnexpected, err, "could not decrypt private key")
		return
	}

	blockCipher, err := aes.NewCipher(aesKey)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrUnexpected, err, "could not create new cipher")
		return
	}

	aesGCM, err := cipher.NewGCM(blockCipher)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrUnexpected, err, "could not create new GCM")
		return
	}

	decrypted, err = aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrUnexpected, err, "could not open ciphertext")
		return
	}

	return
}

func encryptWithRSAKey(rsaPub *rsa.PublicKey, payload []byte) (encrypted *HybridAESRSAEncryptedPayload, err error) {
	if rsaPub == nil {
		err = commonerrors.New(commonerrors.ErrUndefined, "rsa public key is undefined")
		return
	}

	aesKey := make([]byte, 32)
	_, err = rand.Read(aesKey)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrUnexpected, err, "failed generating AES key")
		return
	}

	blockCipher, err := aes.NewCipher(aesKey)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrUnexpected, err, "failed to create cipher")
		return
	}

	aesGCM, err := cipher.NewGCM(blockCipher)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrUnexpected, err, "failed creating GCM")
		return
	}

	nonce := make([]byte, aesGCM.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrUnexpected, err, "failed to generate nonce")
		return
	}

	ciphertext := aesGCM.Seal(nil, nonce, payload, nil)
	encryptedAESKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, rsaPub, aesKey, []byte{})
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrUnexpected, err, "could not complete RSA encryption")
		return
	}

	encrypted = EncodeHybridAESRSAEncryptedPayload(ciphertext, encryptedAESKey, nonce)
	return
}

// EncryptHybridAESRSAEncryptedPayloadFromPublickKeyBytes takes a public key for key encypherment and uses it to
// encode a payload using hybrid RSA AES encryption where an AES key is used to encrypt the content in payload and the
// AES key is encrypted using RSA encryption. AES encryption is used to encode the payload itself as it is faster than
// RSA for larger payloads. RSA is used to encrypt the relatively small AES key and allows asymmetric encryption
// whilst also being fast. More information can be found at https://www.ijrar.org/papers/IJRAR23B1852.pdf
func EncryptHybridAESRSAEncryptedPayloadFromPublickKeyBytes(publicKeyBytes []byte, payload []byte) (encrypted *HybridAESRSAEncryptedPayload, err error) {
	block, _ := pem.Decode(publicKeyBytes)
	if block == nil {
		err = commonerrors.New(commonerrors.ErrUnexpected, "failed to decode PEM block from certificate")
		return
	}

	if block == nil {
		err = commonerrors.New(commonerrors.ErrEmpty, "block was empty")
		return
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		pkcs1PublicKey, subErr := x509.ParsePKCS1PublicKey(block.Bytes)
		if subErr != nil {
			err = commonerrors.Newf(commonerrors.ErrUnexpected, "failed to parse public key as PKIX or PKCS1: %v, %v", err, subErr)
			return
		}

		return encryptWithRSAKey(pkcs1PublicKey, payload)
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		err = commonerrors.New(commonerrors.ErrUnexpected, "could not parse PKIX pem block as RSA key")
		return
	}

	return encryptWithRSAKey(rsaPub, payload)
}

// EncryptHybridAESRSAEncryptedPayloadFromBytes takes an x509 certificate for key encypherment and uses it to
// encode a payload using hybrid RSA AES encryption where an AES key is used to encrypt the content in payload and the
// AES key is encrypted using RSA encryption. AES encryption is used to encode the payload itself as it is faster than
// RSA for larger payloads. RSA is used to encrypt the relatively small AES key and allows asymmetric encryption
// whilst also being fast. More information can be found at https://www.ijrar.org/papers/IJRAR23B1852.pdf
func EncryptHybridAESRSAEncryptedPayloadFromBytes(block []byte, payload []byte) (encrypted *HybridAESRSAEncryptedPayload, err error) {
	cert, err := x509.ParseCertificate(block)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrUnexpected, err, "failed parsing certificate")
		return
	}

	rsaPub, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		err = commonerrors.New(commonerrors.ErrInvalid, "public key in certificate is not RSA")
		return
	}

	return encryptWithRSAKey(rsaPub, payload)
}

// EncryptHybridAESRSAEncryptedPayloadFromCertificate takes a path to a valid x509 certificate for key encypherment
// and uses it to encode a payload using hybrid RSA AES encryption where an AES key is used to encrypt the content in
// payload and the AES key is encrypted using RSA encryption. AES encryption is used to encode the payload itself as
// it is faster than RSA for larger payloads. RSA is used to encrypt the relatively small AES key and allows asymmetric
// encryption whilst also being fast. More information can be found at https://www.ijrar.org/papers/IJRAR23B1852.pdf
func EncryptHybridAESRSAEncryptedPayloadFromCertificate(certPath string, payload []byte) (encrypted *HybridAESRSAEncryptedPayload, err error) {
	block, err := ParsePEMBlock(certPath)
	if err != nil {
		err = commonerrors.WrapErrorf(commonerrors.ErrUnexpected, err, "could not parse PEM block from '%v'", certPath)
		return
	}

	if block == nil {
		err = commonerrors.New(commonerrors.ErrEmpty, "block was empty")
		return
	}

	return EncryptHybridAESRSAEncryptedPayloadFromBytes(block.Bytes, payload)
}
