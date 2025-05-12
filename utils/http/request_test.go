package http

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/field"
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

func TestAuth(t *testing.T) {
	cfg, err := NewAuthConfiguration(nil)
	require.NoError(t, err)
	assert.False(t, cfg.Enforced)
	cfg.Enforced = true
	require.Error(t, cfg.Validate())
	cfg.AccessToken = faker.Password()
	cfg.Scheme = faker.Name()
	require.Error(t, cfg.Validate())
	cfg.Scheme = AuthorisationSchemeToken
	require.NoError(t, cfg.Validate())
	cfg2, err := NewAuthConfiguration(field.ToOptionalString(faker.Password()))
	require.Error(t, err)
	assert.Empty(t, cfg2)
	cfg2, err = NewAuthConfiguration(field.ToOptionalString(cfg.GetAuthorizationHeader()))
	require.NoError(t, err)
	require.NoError(t, cfg2.Validate())
	assert.True(t, cfg.Enforced)
	cfg.Scheme = faker.Name()
	cfg2, err = NewAuthConfiguration(field.ToOptionalString(cfg.GetAuthorizationHeader()))
	require.Error(t, err)
	require.Error(t, cfg2.Validate())
	assert.True(t, cfg.Enforced)
}
