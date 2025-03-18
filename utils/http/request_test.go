package http

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/require"
)

func TestRequestConfiguration_Validate(t *testing.T) {
	cfg := DefaultHTTPRequestConfiguration(faker.Name())
	require.Error(t, cfg.Validate())
	cfg.Host = faker.URL()
	require.NoError(t, cfg.Validate())
	cfg.Port = "123"
	require.NoError(t, cfg.Validate())
	cfg = DefaultHTTPRequestWithAuthorisationConfigurationEnforced(faker.Name())
	cfg.Host = faker.URL()
	require.Error(t, cfg.Validate())
	cfg.Authorisation.AccessToken = faker.Password()
	cfg.Authorisation.Scheme = faker.Name()
	require.Error(t, cfg.Validate())
	cfg.Authorisation.Scheme = AuthorisationSchemeToken
	require.NoError(t, cfg.Validate())
}
