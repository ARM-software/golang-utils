package signing

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

type Ed25519Signer struct {
	Public  ed25519.PublicKey  `json:"public"`
	private ed25519.PrivateKey `json:"-"`
	seed    []byte             `json:"-"`
}

func (k *Ed25519Signer) String() string {
	return fmt.Sprintf("{Public: %v}", k.GetPublicKey())
}

func (k *Ed25519Signer) GoString() string {
	return fmt.Sprintf("KeyPair(%q)", k.String())
}

func (k *Ed25519Signer) MarshalJSON() (jsonBytes []byte, err error) {
	return []byte(fmt.Sprintf("{\"public\":%q}", k.GetPublicKey())), nil
}

func (k *Ed25519Signer) GetPublicKey() string {
	return base64.StdEncoding.EncodeToString(k.Public)
}

func (k *Ed25519Signer) GetPrivateKey() string {
	return base64.StdEncoding.EncodeToString(k.private)
}

func (k *Ed25519Signer) Sign(message []byte) (signature []byte, err error) {
	if len(k.private) == 0 {
		err = commonerrors.New(commonerrors.ErrUndefined, "missing private key")
		return
	}
	if len(k.private) != ed25519.PrivateKeySize {
		err = commonerrors.Newf(commonerrors.ErrInvalid, "invalid private key length %v", len(k.private))
		return
	}

	signature, err = k.private.Sign(nil, message, &ed25519.Options{})
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrUnexpected, err, "error occurred whilst signing")
		return
	}

	return
}

func (k *Ed25519Signer) GenerateSignature(message []byte) (signatureBase64 string, err error) {
	signature, err := k.Sign(message)
	if err != nil {
		return
	}
	signatureBase64 = base64.StdEncoding.EncodeToString(signature)
	return
}

func (k *Ed25519Signer) Verify(message, signature []byte) (ok bool, err error) {
	if len(k.Public) == 0 {
		err = commonerrors.New(commonerrors.ErrUndefined, "missing public key")
		return
	}
	if len(k.Public) != ed25519.PublicKeySize {
		err = commonerrors.Newf(commonerrors.ErrInvalid, "invalid public key length %v", len(k.Public))
		return
	}

	ok = ed25519.Verify(k.Public, message, signature)
	return
}

func (k *Ed25519Signer) VerifySignature(message []byte, signatureBase64 string) (ok bool, err error) {
	signature, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil {
		return
	}

	ok, err = k.Verify(message, signature)
	return
}

// NewEd25519Signer will create a Ed25519Signer that can both sign new messages as well as verify them
func NewEd25519Signer(privateKey ed25519.PrivateKey) (signer *Ed25519Signer, err error) {
	if privateKey == nil {
		err = commonerrors.New(commonerrors.ErrUndefined, "privateKey must be defined")
		return
	}
	if len(privateKey) != ed25519.PrivateKeySize {
		err = commonerrors.Newf(commonerrors.ErrInvalid, "private key must have length %v, it has length %v", ed25519.PrivateKeySize, len(privateKey))
		return
	}

	publicKey, ok := privateKey.Public().(ed25519.PublicKey)
	if !ok {
		err = commonerrors.New(commonerrors.ErrUnexpected, "could not extract public key from private key")
		return
	}

	signer = &Ed25519Signer{
		Public:  publicKey,
		private: privateKey,
		seed:    privateKey.Seed(),
	}

	return
}

// NewEd25519Verifier will create a Ed25519Signer with only a public key meaning it can only verify messages
func NewEd25519Verifier(publicKey ed25519.PublicKey) (signer *Ed25519Signer, err error) {
	if publicKey == nil {
		err = commonerrors.New(commonerrors.ErrUndefined, "publicKey must be defined")
		return
	}
	if len(publicKey) != ed25519.PublicKeySize {
		err = commonerrors.Newf(commonerrors.ErrInvalid, "public key must have length %v, it has length %v", ed25519.PrivateKeySize, len(publicKey))
		return
	}

	signer = &Ed25519Signer{
		Public: publicKey,
	}

	return
}

// NewEd25519SignerFromBase64 will create a Ed25519Signer that can both sign new messages as well as verify them
// It will take a private key encoded as base64
func NewEd25519SignerFromBase64(privateKeyB64 string) (signer *Ed25519Signer, err error) {
	privateKey, err := base64.StdEncoding.DecodeString(privateKeyB64)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrInvalid, err, "could not decode private key from base64")
		return
	}

	return NewEd25519Signer(privateKey)
}

// NewEd25519VerifierFromBase64 will create a Ed25519Signer with only a public key meaning it can only verify messages
// It will take a public key encoded as base64
func NewEd25519VerifierFromBase64(publicKeyB64 string) (signer *Ed25519Signer, err error) {
	publicKey, err := base64.StdEncoding.DecodeString(publicKeyB64)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrInvalid, err, "could not decode public key from base64")
		return
	}

	return NewEd25519Verifier(publicKey)
}

// NewEd25519SignerFromSeed will create an Ed25519Signer based on a seed. It will automatically pad the seed to the correct length
// A seed for Ed25519 should be 32 characters long. Anything shorter will be padded with zeros and anything longer will be truncated
func NewEd25519SignerFromSeed(inputSeed string) (pair *Ed25519Signer, err error) {
	seed := make([]byte, ed25519.SeedSize)
	if inputSeed == "" {
		_, err = io.ReadFull(rand.Reader, seed)
		if err != nil {
			return
		}
	} else {
		for i := range seed {
			if i < len(inputSeed) {
				seed[i] = inputSeed[i]
			} else {
				seed[i] = '0'
			}
		}
	}

	privateKey := ed25519.NewKeyFromSeed(seed)
	pair, err = NewEd25519Signer(privateKey)
	pair.seed = seed
	return
}
