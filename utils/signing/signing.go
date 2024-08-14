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
		err = fmt.Errorf("%w: missing private key", commonerrors.ErrInvalid)
		return
	}
	if len(k.private) != ed25519.PrivateKeySize {
		err = fmt.Errorf("%w: invalid private key length %v", commonerrors.ErrInvalid, len(k.private))
		return
	}

	signature, err = k.private.Sign(nil, message, &ed25519.Options{})
	if err != nil {
		err = fmt.Errorf("%w: error occured whilst signing: %v", commonerrors.ErrUnexpected, err.Error())
		return
	}

	return
}

func (k *Ed25519Signer) Verify(message, signature []byte) (ok bool, err error) {
	if len(k.Public) == 0 {
		err = fmt.Errorf("%w: missing public key", commonerrors.ErrInvalid)
		return
	}
	if len(k.Public) != ed25519.PublicKeySize {
		err = fmt.Errorf("%w: invalid public key length %v", commonerrors.ErrInvalid, len(k.Public))
		return
	}

	ok = ed25519.Verify(k.Public, message, signature)
	return
}

// NewEd25519Signer will create a Ed25519Signer that can both sign new messages as well as verify them
func NewEd25519Signer(privateKey ed25519.PrivateKey) (signer *Ed25519Signer, err error) {
	if len(privateKey) != ed25519.PrivateKeySize {
		err = fmt.Errorf("%w: private key must have length %v, it has length %v", commonerrors.ErrInvalid, ed25519.PrivateKeySize, len(privateKey))
		return
	}

	publicKey, ok := privateKey.Public().(ed25519.PublicKey)
	if !ok {
		err = fmt.Errorf("%w: could not extract public key from private key", commonerrors.ErrUnexpected)
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
	if len(publicKey) != ed25519.PublicKeySize {
		err = fmt.Errorf("%w: public key must have length %v, it has length %v", commonerrors.ErrInvalid, ed25519.PrivateKeySize, len(publicKey))
		return
	}

	signer = &Ed25519Signer{
		Public: publicKey,
	}

	return
}

// NewEd25519SignerFromSeed will create an Ed25519Signer based on a seed. It will automatically pad the seed to the correct length
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
