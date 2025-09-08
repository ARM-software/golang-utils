package signing

import (
	"crypto/ed25519"
	"encoding/base64"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
)

func TestSigning(t *testing.T) {
	t.Run("Successful sign", func(t *testing.T) {
		message := []byte(faker.Word())

		signer, err := NewEd25519SignerFromSeed(faker.Word())
		require.NoError(t, err)

		signature, err := signer.Sign(message)
		require.NoError(t, err)

		ok, err := signer.Verify(message, signature)
		require.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("Invalid message", func(t *testing.T) {
		message := []byte(faker.Word())

		signer, err := NewEd25519SignerFromSeed(faker.Word())
		require.NoError(t, err)

		signature, err := signer.Sign(message)
		require.NoError(t, err)

		ok, err := signer.Verify([]byte(faker.Word()+faker.Word()), signature)
		require.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("Invalid signature", func(t *testing.T) {
		message := []byte(faker.Word())

		signer, err := NewEd25519SignerFromSeed(faker.Word())
		require.NoError(t, err)

		wrongSignature, err := signer.Sign([]byte(faker.Word() + faker.Word()))
		require.NoError(t, err)

		ok, err := signer.Verify(message, wrongSignature)
		require.NoError(t, err)
		assert.False(t, ok)
	})
}

func TestNewSigner(t *testing.T) {
	t.Run("Test signing key from seed", func(t *testing.T) {
		k, err := NewEd25519SignerFromSeed("1234")
		require.NoError(t, err)
		assert.Equal(t, k.seed, []byte("12340000000000000000000000000000"))
		assert.EqualValues(t, k.private, []byte{0x31, 0x32, 0x33, 0x34, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0xf1, 0x8, 0x8e, 0x21, 0xe9, 0x5c, 0xc9, 0xf6, 0x34, 0xf5, 0x73, 0xbe, 0xc4, 0x6, 0xd7, 0xa6, 0xd, 0xc1, 0x7b, 0x9d, 0xb3, 0x9a, 0xa0, 0x92, 0xdb, 0x70, 0x8e, 0x6, 0x5, 0x1e, 0x16, 0xfb})
		assert.EqualValues(t, k.Public, []byte{0xf1, 0x8, 0x8e, 0x21, 0xe9, 0x5c, 0xc9, 0xf6, 0x34, 0xf5, 0x73, 0xbe, 0xc4, 0x6, 0xd7, 0xa6, 0xd, 0xc1, 0x7b, 0x9d, 0xb3, 0x9a, 0xa0, 0x92, 0xdb, 0x70, 0x8e, 0x6, 0x5, 0x1e, 0x16, 0xfb})
		assert.Equal(t, k.Public, k.private.Public())
	})

	t.Run("Test NewEd25519Signer", func(t *testing.T) {
		testValidKey, err := NewEd25519SignerFromSeed("12234778")
		require.NoError(t, err)
		k, err := NewEd25519Signer(testValidKey.private)
		assert.NoError(t, err)
		assert.Equal(t, testValidKey, k)
		assert.Equal(t, "MTIyMzQ3NzgwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAODBfpjfvG3GlkUsrmSBcFnmUSSh63RXhczGbdd9oUfQ==", k.GetPrivateKey())
	})

	t.Run("Test NewEd25519Signer (not valid)", func(t *testing.T) {
		k, err := NewEd25519Signer([]byte("not a private key"))
		errortest.AssertError(t, err, commonerrors.ErrInvalid)
		assert.Nil(t, k)
	})

	t.Run("Test NewEd25519Signer (private empty)", func(t *testing.T) {
		k, err := NewEd25519Signer([]byte{})
		errortest.AssertError(t, err, commonerrors.ErrInvalid)
		assert.Nil(t, k)
	})

	t.Run("test base64 private key", func(t *testing.T) {
		testKey, err := NewEd25519SignerFromSeed("1234")
		require.NoError(t, err)

		_, err = NewEd25519SignerFromBase64(base64.StdEncoding.EncodeToString(testKey.private))
		require.NoError(t, err)
	})

	t.Run("test invalid base64 public key", func(t *testing.T) {
		testKey, err := NewEd25519SignerFromBase64("i am not base64 8907987_^?%+")
		errortest.AssertError(t, err, commonerrors.ErrInvalid)
		assert.Nil(t, testKey)
	})
}

func TestSign(t *testing.T) {
	t.Run("test valid private key", func(t *testing.T) {
		testKey, err := NewEd25519SignerFromSeed("1234")
		require.NoError(t, err)
		_, err = testKey.Sign([]byte(faker.Sentence()))
		assert.NoError(t, err)
	})

	t.Run("test invalid private key", func(t *testing.T) {
		testKey, err := NewEd25519SignerFromSeed("1234")
		require.NoError(t, err)
		testKey.private = ed25519.PrivateKey([]byte("thiswontbevalid"))
		_, err = testKey.Sign([]byte(faker.Sentence()))
		errortest.AssertError(t, err, commonerrors.ErrInvalid)
	})

	t.Run("test no private key", func(t *testing.T) {
		testKey, err := NewEd25519SignerFromSeed("1234")
		require.NoError(t, err)
		testKey.private = ed25519.PrivateKey{}
		_, err = testKey.Sign([]byte(faker.Sentence()))
		errortest.AssertError(t, err, commonerrors.ErrUndefined)
	})

	t.Run("test base64", func(t *testing.T) {
		testKey, err := NewEd25519SignerFromSeed("1234")
		require.NoError(t, err)

		testMessage := []byte(faker.Sentence())

		sigNormal, err := testKey.Sign(testMessage)
		require.NoError(t, err)

		sigBase64, err := testKey.GenerateSignature(testMessage)
		require.NoError(t, err)

		assert.Equal(t, sigBase64, base64.StdEncoding.EncodeToString(sigNormal))
	})
}

func TestVerify(t *testing.T) {
	testKey, err := NewEd25519SignerFromSeed("1234")
	require.NoError(t, err)

	publicKey := testKey.Public
	message := []byte(faker.Sentence())
	signature, err := testKey.Sign(message)
	require.NoError(t, err)
	signatureBase64 := base64.StdEncoding.EncodeToString(signature)

	t.Run("test valid private key", func(t *testing.T) {
		testKey, err := NewEd25519Verifier(publicKey)
		require.NoError(t, err)
		ok, err := testKey.Verify(message, signature)
		assert.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("test missing public key", func(t *testing.T) {
		testKey, err := NewEd25519Verifier(publicKey)
		require.NoError(t, err)
		testKey.Public = ed25519.PublicKey{}
		ok, err := testKey.Verify(message, signature)
		errortest.AssertError(t, err, commonerrors.ErrUndefined)
		assert.False(t, ok)
	})

	t.Run("test invalid length public key", func(t *testing.T) {
		testKey, err := NewEd25519Verifier(publicKey)
		require.NoError(t, err)
		testKey.Public = testKey.Public[2:18]
		ok, err := testKey.Verify(message, signature)
		errortest.AssertError(t, err, commonerrors.ErrInvalid)
		assert.False(t, ok)
	})

	t.Run("test valid private key (using base64)", func(t *testing.T) {
		testKey, err := NewEd25519Verifier(publicKey)
		require.NoError(t, err)
		ok, err := testKey.VerifySignature(message, signatureBase64)
		assert.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("test missing public key (using base64)", func(t *testing.T) {
		testKey, err := NewEd25519Verifier(publicKey)
		require.NoError(t, err)
		testKey.Public = ed25519.PublicKey{}
		ok, err := testKey.VerifySignature(message, signatureBase64)
		errortest.AssertError(t, err, commonerrors.ErrUndefined)
		assert.False(t, ok)
	})

	t.Run("test invalid length public key (using base64)", func(t *testing.T) {
		testKey, err := NewEd25519Verifier(publicKey)
		require.NoError(t, err)
		testKey.Public = testKey.Public[2:18]
		ok, err := testKey.VerifySignature(message, signatureBase64)
		errortest.AssertError(t, err, commonerrors.ErrInvalid)
		assert.False(t, ok)
	})
}

func TestNewVerifier(t *testing.T) {
	t.Run("test valid public key", func(t *testing.T) {
		testKey, err := NewEd25519SignerFromSeed("1234")
		require.NoError(t, err)

		publicKey := testKey.Public
		_, err = NewEd25519Verifier(publicKey)
		require.NoError(t, err)
	})

	t.Run("test missing public key", func(t *testing.T) {
		testKey, err := NewEd25519Verifier(ed25519.PublicKey{})
		errortest.AssertError(t, err, commonerrors.ErrInvalid)
		assert.Nil(t, testKey)
	})

	t.Run("test invalid length public key", func(t *testing.T) {
		testKey, err := NewEd25519SignerFromSeed("1234")
		require.NoError(t, err)

		publicKey := testKey.Public

		testKey, err = NewEd25519Verifier(publicKey[2:18])
		errortest.AssertError(t, err, commonerrors.ErrInvalid)
		assert.Nil(t, testKey)
	})

	t.Run("test base64 public key", func(t *testing.T) {
		testKey, err := NewEd25519SignerFromSeed("1234")
		require.NoError(t, err)

		_, err = NewEd25519VerifierFromBase64(base64.StdEncoding.EncodeToString(testKey.Public))
		require.NoError(t, err)
	})

	t.Run("test invalid base64 public key", func(t *testing.T) {
		testKey, err := NewEd25519VerifierFromBase64("i am not base64 8907987_^?%+")
		errortest.AssertError(t, err, commonerrors.ErrInvalid)
		assert.Nil(t, testKey)
	})
}

func TestEncryptionRequiredMethods(t *testing.T) {
	testKey, err := NewEd25519SignerFromSeed("1234791289218")
	require.NoError(t, err)

	t.Run("String", func(t *testing.T) {
		s := testKey.String()
		assert.Equal(t, "{Public: fK8uaDt/2+1RoVwnj4JotkAu3SGm7cf5RhpZNWbLPSA=}", s)
	})

	t.Run("GoString", func(t *testing.T) {
		s := testKey.GoString()
		assert.Equal(t, "KeyPair(\"{Public: fK8uaDt/2+1RoVwnj4JotkAu3SGm7cf5RhpZNWbLPSA=}\")", s)
	})

	t.Run("MarshalJSON", func(t *testing.T) {
		b, err := testKey.MarshalJSON()
		require.NoError(t, err)
		assert.Equal(t, "{\"public\":\"fK8uaDt/2+1RoVwnj4JotkAu3SGm7cf5RhpZNWbLPSA=\"}", string(b))
	})
}
