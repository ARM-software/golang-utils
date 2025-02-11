package testhelpers

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math"
	"math/big"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/filesystem"
)

func GenerateTestCerts(t *testing.T) (certPath, keyPath string) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	serialNum, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	require.NoError(t, err)

	certTemplate := &x509.Certificate{
		SerialNumber:          serialNum,
		Subject:               pkix.Name{Organization: []string{faker.Name()}}, //nolint:misspell // library is american
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(1 * time.Minute),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	certB, err := x509.CreateCertificate(rand.Reader, certTemplate, certTemplate, &key.PublicKey, key)
	require.NoError(t, err)

	var certBuf, keyBuf bytes.Buffer
	err = pem.Encode(&certBuf, &pem.Block{Type: "CERTIFICATE", Bytes: certB})
	require.NoError(t, err)

	err = pem.Encode(&keyBuf, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	require.NoError(t, err)

	tmpDir := t.TempDir()

	certPath = filepath.Join(tmpDir, "test_cert.pem")
	err = filesystem.WriteFile(certPath, certBuf.Bytes(), 0644)
	require.NoError(t, err)

	keyPath = filepath.Join(tmpDir, "test_key.pem")
	err = filesystem.WriteFile(keyPath, keyBuf.Bytes(), 0644)
	require.NoError(t, err)

	return
}
