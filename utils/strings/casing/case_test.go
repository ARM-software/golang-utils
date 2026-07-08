package casing

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCaseHelpersWithOptionalReplacer(t *testing.T) {
	r, err := NewReplacer(
		Rule{Token: "Api", Replacement: "API"},
		Rule{Token: "Id", Replacement: "ID", Exceptions: []string{"identifier"}},
	)
	require.NoError(t, err)

	assert.Equal(t, "apiClient", ToCamelCase("api_client", r))
	assert.Equal(t, "APIClient", ToPascalCase("api_client", r))
	assert.Equal(t, "api_client", ToSnakeCase("api_client", r))
	assert.Equal(t, "api-client", ToKebabCase("api_client", r))
}

func TestCaseHelpersWithoutReplacer(t *testing.T) {
	assert.Equal(t, "sourceName", ToCamelCase("source_name"))
	assert.Equal(t, "SourceName", ToPascalCase("source_name"))
	assert.Equal(t, "source_name", ToSnakeCase("sourceName"))
	assert.Equal(t, "source-name", ToKebabCase("sourceName"))

	assert.Equal(t, "aiapi", ToSnakeCase("AIAPI"))
	assert.Equal(t, "aiapi", ToKebabCase("AIAPI"))
	assert.Equal(t, "httpapi_token", ToSnakeCase("HTTPAPIToken"))
	assert.Equal(t, "httpapi-token", ToKebabCase("HTTPAPIToken"))
	assert.Equal(t, "httpapiToken", ToCamelCase("HTTPAPIToken"))
	assert.Equal(t, "HttpapiToken", ToPascalCase("HTTPAPIToken"))
}

func TestCaseHelpersWithoutReplacer_StrcaseInspiredCases(t *testing.T) {
	// These cases are adapted from the broader casing corpus used by
	// github.com/ettle/strcase to make sure this package behaves consistently on
	// common acronym and mixed-token inputs, even though this package does not
	// apply Go initialism rules unless explicitly configured through a Replacer.
	assert.Equal(t, "Id", ToPascalCase("ID"))
	assert.Equal(t, "id", ToCamelCase("ID"))
	assert.Equal(t, "id", ToSnakeCase("ID"))

	assert.Equal(t, "userId", ToCamelCase("userID"))
	assert.Equal(t, "UserId", ToPascalCase("userID"))
	assert.Equal(t, "user_id", ToSnakeCase("userID"))

	assert.Equal(t, "jsonBlob", ToCamelCase("JSON_blob"))
	assert.Equal(t, "JsonBlob", ToPascalCase("JSON_blob"))
	assert.Equal(t, "json_blob", ToSnakeCase("JSON_blob"))

	assert.Equal(t, "httpStatusCode", ToCamelCase("HTTPStatusCode"))
	assert.Equal(t, "HttpStatusCode", ToPascalCase("HTTPStatusCode"))
	assert.Equal(t, "http_status_code", ToSnakeCase("HTTPStatusCode"))

	assert.Equal(t, "freeBsd", ToCamelCase("FreeBSD"))
	assert.Equal(t, "FreeBsd", ToPascalCase("FreeBSD"))
	assert.Equal(t, "free_bsd", ToSnakeCase("FreeBSD"))
}

func TestCaseHelpersWithOptionalReplacer_StrcaseInspiredInitialisms(t *testing.T) {
	r, err := NewReplacer(
		Rule{Token: "Api", Replacement: "API"},
		Rule{Token: "Http", Replacement: "HTTP"},
		Rule{Token: "Id", Replacement: "ID"},
		Rule{Token: "Json", Replacement: "JSON"},
	)
	require.NoError(t, err)

	assert.Equal(t, "userID", ToCamelCase("user_id", r))
	assert.Equal(t, "UserID", ToPascalCase("user_id", r))
	assert.Equal(t, "user_id", ToSnakeCase("UserID", r))

	assert.Equal(t, "httpStatusCode", ToCamelCase("http_status_code", r))
	assert.Equal(t, "HTTPStatusCode", ToPascalCase("http_status_code", r))
	assert.Equal(t, "http_status_code", ToSnakeCase("HTTPStatusCode", r))

	assert.Equal(t, "jsonBlob", ToCamelCase("json_blob", r))
	assert.Equal(t, "JSONBlob", ToPascalCase("json_blob", r))
	assert.Equal(t, "json_blob", ToSnakeCase("JSONBlob", r))

	acr, err := NewReplacer(
		Rule{Token: "Aes", Replacement: "AES"},
		Rule{Token: "Rsa", Replacement: "RSA"},
	)
	require.NoError(t, err)
	assert.Equal(t, "hybridAESRSAEncryptedPayload", ToCamelCase("hybridAESRSAEncryptedPayload", acr))
	assert.Equal(t, "hybridAESRSAEncryptedPayload", ToCamelCase("hybridAesrsaEncryptedPayload", acr))
	assert.Equal(t, "HybridAESRSAEncryptedPayload", ToPascalCase("HybridAESRSAEncryptedPayload", acr))
	assert.Equal(t, "HybridAESRSAEncryptedPayload", ToPascalCase("HybridAesrsaEncryptedPayload", acr))
	assert.Equal(t, "hybrid_aesrsa_encrypted_payload", ToSnakeCase("HybridAESRSAEncryptedPayload", acr))
	assert.Equal(t, "hybrid_aesrsa_encrypted_payload", ToSnakeCase("HybridAesrsaEncryptedPayload", acr))
	assert.Equal(t, "hybrid-aesrsa-encrypted-payload", ToKebabCase("HybridAESRSAEncryptedPayload", acr))
	assert.Equal(t, "hybrid-aesrsa-encrypted-payload", ToKebabCase("HybridAesrsaEncryptedPayload", acr))

	portReplacer, err := NewReplacer(Rule{Token: "Port", Replacement: "Port"})
	require.NoError(t, err)
	assert.Equal(t, "Port", ToPascalCase("port", portReplacer))
	assert.Equal(t, "port", ToCamelCase("port", portReplacer))
	assert.Equal(t, "Port", ToPascalCase("Port", portReplacer))
	assert.Equal(t, "port", ToCamelCase("Port", portReplacer))
}

func TestCaseHelpersUseOnlyFirstReplacer(t *testing.T) {
	first, err := NewReplacer(Rule{Token: "Api", Replacement: "API"})
	require.NoError(t, err)
	second, err := NewReplacer(Rule{Token: "Api", Replacement: "XXX"})
	require.NoError(t, err)

	assert.Equal(t, "APIClient", ToPascalCase("api_client", first, second))
}
