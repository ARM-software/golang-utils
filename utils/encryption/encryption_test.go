package encryption

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/logs/logstest"
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
	decryptedMessage, err := DecryptWithKeyPair(pair, encrypted)
	require.NoError(t, err)
	assert.Equal(t, message, decryptedMessage)
}

func TestKeyPrint(t *testing.T) {
	// Test to make sure the private key does not get printing in the logs by mistake.
	pair, err := GenerateKeyPair()
	require.NoError(t, err)

	fmtString := fmt.Sprintf("test: %v", pair)
	assert.NotContains(t, fmtString, pair.GetPrivateKey())
	fmt.Println(fmtString)

	fmtString = fmt.Sprintf("test: %+v", pair)
	assert.NotContains(t, fmtString, pair.GetPrivateKey())
	fmt.Println(fmtString)

	fmtString = fmt.Sprintf("test: %q", pair)
	assert.NotContains(t, fmtString, pair.GetPrivateKey())
	fmt.Println(fmtString)

	fmtJSON, err := json.Marshal(pair)
	require.NoError(t, err)
	fmtString = string(fmtJSON)
	assert.NotContains(t, fmtString, pair.GetPrivateKey())
	fmt.Println(fmtString)
	logger := logstest.NewTestLogger(t)
	logger.Info("test", "key", pair)
}

func TestEncryptDecrypt_Failures(t *testing.T) {
	pair, err := GenerateKeyPair()
	require.NoError(t, err)

	_, err = EncryptWithPublicKey(faker.Name(), faker.Word())
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrInvalid)

	invalidPair := newBasicKeyPair(faker.Name(), faker.Name())
	_, err = DecryptWithKeyPair(invalidPair, faker.Word())
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrInvalid)

	invalidPair = newBasicKeyPair(pair.GetPublicKey(), faker.Name())
	_, err = DecryptWithKeyPair(invalidPair, faker.Word())
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrInvalid)

	invalidPair = newBasicKeyPair(faker.Name(), pair.GetPrivateKey())
	_, err = DecryptWithKeyPair(invalidPair, faker.Word())
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrInvalid)

	_, err = DecryptWithKeyPair(pair, faker.Word())
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrInvalid)
}
