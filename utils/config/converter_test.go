package config

import (
	"context"
	"errors"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
)

type secretStringer string

func (s secretStringer) String() string {
	return string(s)
}

type secretTextValue string

func (s secretTextValue) MarshalText() ([]byte, error) {
	return []byte(string(s)), nil
}

type secretFailingTextValue struct {
	value string
}

func (secretFailingTextValue) MarshalText() ([]byte, error) {
	return nil, errors.New("boom")
}

type taggedSecretNestedConfig struct {
	Token   string `mapstructure:"token" secret:"true"`
	Visible string `mapstructure:"visible"`
}

func (c *taggedSecretNestedConfig) Validate() error {
	return nil
}

type taggedSecretConfig struct {
	Username string                   `mapstructure:"username"`
	Password string                   `mapstructure:"password" mask:"redact"`
	APIKey   string                   `mapstructure:"api_key" sensitive:"true"`
	Nested   taggedSecretNestedConfig `mapstructure:"nested"`
}

func (c *taggedSecretConfig) Validate() error {
	return nil
}

func TestSecretConverter(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		input := faker.Word() + faker.Word()
		converted, err := SecretConverter.ConvertValue(context.Background(), input)
		require.NoError(t, err)
		assert.Equal(t, string([]rune(input)[0])+"*****"+string([]rune(input)[len([]rune(input))-1]), converted)
	})

	t.Run("stringer", func(t *testing.T) {
		converted, err := SecretConverter.ConvertValue(context.Background(), secretStringer("secret"))
		require.NoError(t, err)
		assert.Equal(t, "s*****t", converted)
	})

	t.Run("text marshaler", func(t *testing.T) {
		converted, err := SecretConverter.ConvertValue(context.Background(), secretTextValue("secret"))
		require.NoError(t, err)
		assert.Equal(t, "s*****t", converted)
	})

	t.Run("nil", func(t *testing.T) {
		converted, err := SecretConverter.ConvertValue(context.Background(), nil)
		require.NoError(t, err)
		assert.Equal(t, "", converted)
	})

	t.Run("empty", func(t *testing.T) {
		converted, err := SecretConverter.ConvertValue(context.Background(), "")
		require.NoError(t, err)
		assert.Equal(t, "", converted)
	})

	t.Run("marshal error", func(t *testing.T) {
		_, err := SecretConverter.ConvertValue(context.Background(), secretFailingTextValue{value: "secret"})
		require.Error(t, err)
		errortest.AssertError(t, err, commonerrors.ErrMarshalling)
		errortest.AssertErrorDescription(t, err, "boom")
	})
}

func TestFieldTagSecretConverters(t *testing.T) {
	configValue := &taggedSecretConfig{
		Username: faker.Username(),
		Password: faker.Password(),
		APIKey:   faker.UUIDDigit(),
		Nested: taggedSecretNestedConfig{
			Token:   faker.UUIDHyphenated(),
			Visible: faker.Word(),
		},
	}

	t.Run("generic converter", func(t *testing.T) {
		defaults, err := DetermineConfigurationEnvironmentVariables("test", configValue, NewFieldTagSecretConverter(SensitiveFieldTagMask, SensitiveFieldTagSensitive))
		require.NoError(t, err)
		maskedPassword, err := SecretConverter.ConvertValue(context.Background(), configValue.Password)
		require.NoError(t, err)
		maskedAPIKey, err := SecretConverter.ConvertValue(context.Background(), configValue.APIKey)
		require.NoError(t, err)

		assert.Equal(t, configValue.Username, defaults["TEST_USERNAME"])
		assert.Equal(t, maskedPassword, defaults["TEST_PASSWORD"])
		assert.Equal(t, maskedAPIKey, defaults["TEST_API_KEY"])
		assert.Equal(t, configValue.Nested.Token, defaults["TEST_NESTED_TOKEN"])
		assert.Equal(t, configValue.Nested.Visible, defaults["TEST_NESTED_VISIBLE"])
	})

	t.Run("common converter", func(t *testing.T) {
		defaults, err := DetermineConfigurationEnvironmentVariables("test", configValue, CommonFieldTagSecretConverter)
		require.NoError(t, err)
		maskedPassword, err := SecretConverter.ConvertValue(context.Background(), configValue.Password)
		require.NoError(t, err)
		maskedAPIKey, err := SecretConverter.ConvertValue(context.Background(), configValue.APIKey)
		require.NoError(t, err)
		maskedToken, err := SecretConverter.ConvertValue(context.Background(), configValue.Nested.Token)
		require.NoError(t, err)

		assert.Equal(t, configValue.Username, defaults["TEST_USERNAME"])
		assert.Equal(t, maskedPassword, defaults["TEST_PASSWORD"])
		assert.Equal(t, maskedAPIKey, defaults["TEST_API_KEY"])
		assert.Equal(t, maskedToken, defaults["TEST_NESTED_TOKEN"])
		assert.Equal(t, configValue.Nested.Visible, defaults["TEST_NESTED_VISIBLE"])
	})
}

func TestFieldNameConverters(t *testing.T) {
	configValue := &taggedSecretConfig{
		Username: faker.Username(),
		Password: faker.Password(),
		APIKey:   faker.UUIDDigit(),
		Nested: taggedSecretNestedConfig{
			Token:   faker.UUIDHyphenated(),
			Visible: faker.Word(),
		},
	}

	t.Run("generic converter", func(t *testing.T) {
		defaults, err := DetermineConfigurationEnvironmentVariables("test", configValue, NewFieldNameConverter(func(fieldName string) bool {
			return fieldName == "Password" || fieldName == "APIKey"
		}, SecretConverter))
		require.NoError(t, err)
		maskedPassword, err := SecretConverter.ConvertValue(context.Background(), configValue.Password)
		require.NoError(t, err)
		maskedAPIKey, err := SecretConverter.ConvertValue(context.Background(), configValue.APIKey)
		require.NoError(t, err)

		assert.Equal(t, configValue.Username, defaults["TEST_USERNAME"])
		assert.Equal(t, maskedPassword, defaults["TEST_PASSWORD"])
		assert.Equal(t, maskedAPIKey, defaults["TEST_API_KEY"])
		assert.Equal(t, configValue.Nested.Token, defaults["TEST_NESTED_TOKEN"])
	})

	t.Run("secret converter", func(t *testing.T) {
		defaults, err := DetermineConfigurationEnvironmentVariables("test", configValue, NewFieldNameSecretConverter(func(fieldName string) bool {
			return fieldName == "Token"
		}))
		require.NoError(t, err)
		maskedToken, err := SecretConverter.ConvertValue(context.Background(), configValue.Nested.Token)
		require.NoError(t, err)

		assert.Equal(t, configValue.Password, defaults["TEST_PASSWORD"])
		assert.Equal(t, maskedToken, defaults["TEST_NESTED_TOKEN"])
		assert.Equal(t, configValue.Nested.Visible, defaults["TEST_NESTED_VISIBLE"])
	})
}
