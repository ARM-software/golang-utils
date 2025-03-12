package aesrsa

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/encryption/aesrsa/testhelpers"
	"github.com/ARM-software/golang-utils/utils/filesystem"
)

func TestEncryption(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		testCertPath, testKeyPath := testhelpers.GenerateTestCerts(t)
		require.FileExists(t, testCertPath)
		require.FileExists(t, testKeyPath)

		content := []byte(faker.Sentence())

		encrypted, err := EncryptHybridAESRSAEncryptedPayloadFromCertificate(testCertPath, content)
		require.NoError(t, err)
		require.NotEmpty(t, encrypted)

		decoded, err := DecryptHybridAESRSAEncryptedPayloadFromPrivateKey(testKeyPath, encrypted)
		assert.NoError(t, err)
		assert.Equal(t, content, decoded)
	})

	t.Run("decrypt: invalid key (missing)", func(t *testing.T) {
		testCertPath, _ := testhelpers.GenerateTestCerts(t)
		require.FileExists(t, testCertPath)

		testKeyPath := filepath.Join(strings.Split(faker.Sentence(), " ")...)

		content := []byte(faker.Sentence())

		encrypted, err := EncryptHybridAESRSAEncryptedPayloadFromCertificate(testCertPath, content)
		require.NoError(t, err)
		require.NotEmpty(t, encrypted)

		decoded, err := DecryptHybridAESRSAEncryptedPayloadFromPrivateKey(testKeyPath, encrypted)
		errortest.AssertErrorDescription(t, err, "could not find certificate")
		assert.Empty(t, decoded)
	})

	t.Run("decrypt: invalid key (wrong key)", func(t *testing.T) {
		testCertPath, _ := testhelpers.GenerateTestCerts(t)
		require.FileExists(t, testCertPath)

		_, testKeyPath := testhelpers.GenerateTestCerts(t)
		require.FileExists(t, testKeyPath)

		content := []byte(faker.Sentence())

		encrypted, err := EncryptHybridAESRSAEncryptedPayloadFromCertificate(testCertPath, content)
		require.NoError(t, err)
		require.NotEmpty(t, encrypted)

		decoded, err := DecryptHybridAESRSAEncryptedPayloadFromPrivateKey(testKeyPath, encrypted)
		errortest.AssertErrorDescription(t, err, "decryption error")
		assert.Empty(t, decoded)
	})

	t.Run("decrypt: invalid key (invalid file)", func(t *testing.T) {
		testCertPath, _ := testhelpers.GenerateTestCerts(t)
		require.FileExists(t, testCertPath)

		testKeyPath := filepath.Join(t.TempDir(), faker.Word())
		err := filesystem.WriteFile(testKeyPath, []byte(faker.Sentence()), 0644)
		require.NoError(t, err)
		require.FileExists(t, testKeyPath)

		content := []byte(faker.Sentence())

		encrypted, err := EncryptHybridAESRSAEncryptedPayloadFromCertificate(testCertPath, content)
		require.NoError(t, err)
		require.NotEmpty(t, encrypted)

		decoded, err := DecryptHybridAESRSAEncryptedPayloadFromPrivateKey(testKeyPath, encrypted)
		errortest.AssertErrorDescription(t, err, "failed to decode PEM block from certificate")
		assert.Empty(t, decoded)
	})

	t.Run("encrypt: invalid key (missing)", func(t *testing.T) {
		testCertPath := filepath.Join(strings.Split(faker.Sentence(), " ")...)
		content := []byte(faker.Sentence())
		require.NoFileExists(t, testCertPath)

		encrypted, err := EncryptHybridAESRSAEncryptedPayloadFromCertificate(testCertPath, content)
		errortest.AssertErrorDescription(t, err, "could not find certificate")
		assert.Nil(t, encrypted)
	})

	t.Run("encrypt: invalid key (invalid file)", func(t *testing.T) {
		testCertPath := filepath.Join(t.TempDir(), faker.Word())
		err := filesystem.WriteFile(testCertPath, []byte(faker.Sentence()), 0644)
		require.NoError(t, err)
		require.FileExists(t, testCertPath)

		content := []byte(faker.Sentence())

		encrypted, err := EncryptHybridAESRSAEncryptedPayloadFromCertificate(testCertPath, content)
		errortest.AssertErrorDescription(t, err, "failed to decode PEM block from certificate")
		assert.Nil(t, encrypted)
	})
}
