package signing

import "github.com/ARM-software/golang-utils/utils/encryption"

type ICodeSigner interface {
	encryption.IKeyPair
	// Sign will sign a message and return a signature
	Sign(message []byte) (signature []byte, err error)
	// Verify will take a message and a signature and verify whether the signature is a valid signature of the message based on the signers public key
	Verify(message, signature []byte) (ok bool, err error)
	// GenerateSignature will sign a message but return a base64 encoded signature for ease of use
	GenerateSignature(message []byte) (signatureBase64 string, err error)
	// Verify will take a message and a base64 encoded signature and verify whether the signature is a valid signature of the message based on the signers public key
	VerifySignature(message []byte, signatureBase64 string) (ok bool, err error)
}
