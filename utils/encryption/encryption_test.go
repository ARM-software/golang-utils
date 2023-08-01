package encryption

import (
	"encoding/base64"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
)

func TestGenerate(t *testing.T) {
	pair, err := GenerateKeyPair()
	require.NoError(t, err)
	assert.NotEmpty(t, pair.GetPublicKey())
	assert.NotEmpty(t, pair.GetPrivateKey())

	b, err := base64.StdEncoding.DecodeString(pair.GetPublicKey())
	require.NoError(t, err)
	assert.Equal(t, 32, len(b))
	b, err = base64.StdEncoding.DecodeString(pair.GetPrivateKey())
	require.NoError(t, err)
	assert.Equal(t, 32, len(b))
}

func TestEncryptDecrypt(t *testing.T) {
	message := faker.Paragraph()
	pair, err := GenerateKeyPair()
	require.NoError(t, err)

	encrypted, err := EncryptWithPublicKey(pair.GetPublicKey(), message)
	require.NoError(t, err)
	decryptedMessage, err := DecryptWithKeyPair(pair.GetPublicKey(), pair.GetPrivateKey(), encrypted)
	require.NoError(t, err)
	assert.Equal(t, message, decryptedMessage)
}

func TestEncryptDecrypt_Failures(t *testing.T) {
	pair, err := GenerateKeyPair()
	require.NoError(t, err)

	_, err = EncryptWithPublicKey(faker.Name(), faker.Word())
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrInvalid)

	_, err = DecryptWithKeyPair(faker.Name(), faker.Name(), faker.Word())
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrInvalid)

	_, err = DecryptWithKeyPair(pair.GetPublicKey(), faker.Name(), faker.Word())
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrInvalid)

	_, err = DecryptWithKeyPair(pair.GetPublicKey(), pair.GetPrivateKey(), faker.Word())
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrInvalid)

}
