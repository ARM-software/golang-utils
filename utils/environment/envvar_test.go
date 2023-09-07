package environment

import (
	"strings"
	"testing"

	"github.com/bxcodec/faker/v3"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
)

func TestEnvVar_Validate(t *testing.T) {
	tests := []struct {
		key         string
		value       string
		name        string
		valueRules  []validation.Rule
		expectError bool
	}{
		{
			name:        "variable with empty key",
			expectError: true,
		},
		{
			key:         faker.Sentence(),
			name:        "variable with whitespaces",
			expectError: true,
		},
		{
			key:         faker.Name(),
			name:        "variable with whitespaces",
			expectError: true,
		},

		{
			key:         faker.Word() + "=" + faker.Word(),
			name:        "variable with `=``",
			expectError: true,
		},
		{
			key:         faker.Word() + "$",
			name:        "variable with special character",
			expectError: true,
		},
		{
			key:         "0" + faker.Word(),
			name:        "variable starting with a digit",
			expectError: true,
		},
		{
			key:         faker.Word(),
			name:        "valid variable",
			expectError: false,
		},
		{
			key:         faker.Word() + "_0",
			name:        "valid variable with digit & underscore",
			expectError: false,
		},
		{
			key:         "_" + faker.Word(),
			name:        "variable starting with an underscore",
			expectError: false,
		},
		{
			key:         faker.Word(),
			name:        "variable value compliant with one rule",
			value:       faker.Sentence(),
			valueRules:  []validation.Rule{validation.Required},
			expectError: false,
		},
		{
			key:         faker.Word(),
			name:        "variable value compliant with several rules",
			value:       faker.Word(),
			valueRules:  []validation.Rule{validation.Required, is.Alphanumeric},
			expectError: false,
		},
		{
			key:         faker.Word(),
			name:        "non compliant variable value",
			valueRules:  []validation.Rule{validation.Required},
			expectError: true,
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			var env IEnvironmentVariable
			if len(test.valueRules) == 0 {
				env = NewEnvironmentVariable(test.key, test.value)
			} else {
				env = NewEnvironmentVariableWithValidation(test.key, test.value, test.valueRules...)
			}
			if test.expectError {
				require.Error(t, env.Validate())
			} else {
				require.NoError(t, env.Validate())
			}
			assert.Equal(t, test.key, env.GetKey())
			assert.Equal(t, test.value, env.GetValue())
		})
	}

}

func TestValidateEnvironmentVariables(t *testing.T) {
	require.NoError(t, ValidateEnvironmentVariables())
	require.NoError(t, ValidateEnvironmentVariables(NewEnvironmentVariable(faker.Word(), "")))
	require.NoError(t, ValidateEnvironmentVariables(NewEnvironmentVariable(faker.Word(), ""), NewEnvironmentVariable(faker.Word(), "")))
	require.Error(t, ValidateEnvironmentVariables(NewEnvironmentVariable(faker.Name(), ""), NewEnvironmentVariable(faker.Word(), "")))
	require.Error(t, ValidateEnvironmentVariables(NewEnvironmentVariable(faker.Word(), ""), NewEnvironmentVariable(faker.Name(), "")))
}

func TestParseEnvironmentVariable(t *testing.T) {
	env, err := ParseEnvironmentVariable(faker.Word())
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrInvalid)
	assert.Nil(t, env)
	_, err = ParseEnvironmentVariable(faker.Word() + "$=" + faker.Word())
	require.Error(t, err)
	errortest.AssertError(t, err, commonerrors.ErrInvalid)
	_, err = ParseEnvironmentVariable(faker.Word() + "=" + faker.Word())
	require.NoError(t, err)
	key := strings.ReplaceAll(strings.ReplaceAll(faker.Sentence(), " ", "_"), ".", "")
	value := faker.Sentence()
	envTest := NewEnvironmentVariable(key, value)
	envTest2 := envTest
	assert.True(t, envTest.Equal(envTest2))
	assert.True(t, envTest2.Equal(envTest))
	env, err = ParseEnvironmentVariable(envTest.String())
	require.NoError(t, err)
	assert.Equal(t, key, env.GetKey())
	assert.Equal(t, value, env.GetValue())
	assert.True(t, env.Equal(envTest))
	txt, err := envTest.MarshalText()
	require.NoError(t, err)
	require.NoError(t, env.UnmarshalText(txt))
	assert.True(t, envTest.Equal(env))
}
