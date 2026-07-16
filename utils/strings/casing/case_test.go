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
		Rule{Token: "Https", Replacement: "HTTPS"},
		Rule{Token: "Id", Replacement: "ID"},
		Rule{Token: "Json", Replacement: "JSON"},
		Rule{Token: "Url", Replacement: "URL"},
		Rule{Token: "Xss", Replacement: "XSS"},
	)
	require.NoError(t, err)

	assert.Equal(t, "userID", ToCamelCase("user_id", r))
	assert.Equal(t, "UserID", ToPascalCase("user_id", r))
	assert.Equal(t, "user_id", ToSnakeCase("UserID", r))

	assert.Equal(t, "httpStatusCode", ToCamelCase("http_status_code", r))
	assert.Equal(t, "HTTPStatusCode", ToPascalCase("http_status_code", r))
	assert.Equal(t, "http_status_code", ToSnakeCase("HTTPStatusCode", r))
	assert.Equal(t, "https", ToCamelCase("https", r))
	assert.Equal(t, "HTTPS", ToPascalCase("https", r))
	assert.Equal(t, "HTTPS", ToPascalCase("Https", r))
	assert.Equal(t, "HTTPS", ToPascalCase("HTTPS", r))
	assert.Equal(t, "iHTTPS", ToCamelCase("IHTTPS", r))
	assert.Equal(t, "IHTTPS", ToPascalCase("IHTTPS", r))
	assert.Equal(t, "IHTTPS", ToPascalCase("IHttps", r))
	assert.Equal(t, "ihttps", ToSnakeCase("IHTTPS", r))
	assert.Equal(t, "ihttps", ToKebabCase("IHTTPS", r))
	assert.Equal(t, "aHTTPClient", ToCamelCase("aHTTPClient", r))
	assert.Equal(t, "AHTTPClient", ToPascalCase("aHTTPClient", r))
	assert.Equal(t, "iHTTP", ToCamelCase("IHTTP", r))
	assert.Equal(t, "iHTTP", ToCamelCase("ihttp", r))
	assert.Equal(t, "iHTTP", ToCamelCase("iHttp", r))
	assert.Equal(t, "iHTTP2", ToCamelCase("IHTTP2", r))
	assert.Equal(t, "iHTTP2", ToCamelCase("ihttp2", r))
	assert.Equal(t, "iHTTP2", ToCamelCase("iHttp2", r))
	assert.Equal(t, "IHTTP", ToPascalCase("IHTTP", r))
	assert.Equal(t, "IHTTP", ToPascalCase("iHTTP", r))
	assert.Equal(t, "IHTTP", ToPascalCase("ihttp", r))
	assert.Equal(t, "IHTTP", ToPascalCase("iHttp", r))
	assert.Equal(t, "IHTTP2", ToPascalCase("IHTTP2", r))
	assert.Equal(t, "IHTTP2", ToPascalCase("ihttp2", r))
	assert.Equal(t, "IHTTP2", ToPascalCase("iHttp2", r))
	assert.Equal(t, "ihttp", ToSnakeCase("IHTTP", r))
	assert.Equal(t, "ihttp", ToSnakeCase("iHTTP", r))
	assert.Equal(t, "ihttp", ToSnakeCase("iHttp", r))
	assert.Equal(t, "ihttp2", ToSnakeCase("IHTTP2", r))
	assert.Equal(t, "ihttp2", ToSnakeCase("iHTTP2", r))
	assert.Equal(t, "ihttp2", ToSnakeCase("iHttp2", r))
	assert.Equal(t, "ihttp", ToKebabCase("IHTTP", r))
	assert.Equal(t, "ihttp", ToKebabCase("iHTTP", r))
	assert.Equal(t, "ihttp", ToKebabCase("iHttp", r))
	assert.Equal(t, "ihttp2", ToKebabCase("IHTTP2", r))
	assert.Equal(t, "ihttp2", ToKebabCase("iHTTP2", r))
	assert.Equal(t, "ihttp2", ToKebabCase("iHttp2", r))
	assert.Equal(t, "URLs", ToPascalCase("urls", r))
	assert.Equal(t, "URLs", ToPascalCase("uRLs", r))
	assert.Equal(t, "urls", ToCamelCase("urls", r))
	assert.Equal(t, "URLs", ToPascalCase("Urls", r))
	assert.Equal(t, "xss", ToCamelCase("xss", r))
	assert.Equal(t, "XSS", ToPascalCase("xss", r))
	assert.Equal(t, "xss", ToSnakeCase("XSS", r))
	assert.Equal(t, "xss", ToKebabCase("XSS", r))
	assert.Equal(t, "xURLValue", ToCamelCase("xURLValue", r))
	assert.Equal(t, "XURLValue", ToPascalCase("xURLValue", r))
	assert.Equal(t, "x_url_value", ToSnakeCase("xURLValue", r))
	assert.Equal(t, "x-url-value", ToKebabCase("xURLValue", r))
	assert.Equal(t, "noKeyring", ToCamelCase("NoKeyring", r))
	assert.Equal(t, "NoKeyring", ToPascalCase("NoKeyring", r))
	assert.Equal(t, "no_keyring", ToSnakeCase("NoKeyring", r))
	assert.Equal(t, "no-keyring", ToKebabCase("NoKeyring", r))
	assert.Equal(t, "userURLs", ToCamelCase("userUrls", r))
	assert.Equal(t, "userURLs", ToCamelCase("userURLs", r))
	assert.Equal(t, "userURLs", ToCamelCase("UserUrls", r))
	assert.Equal(t, "userURLs", ToCamelCase("UserURLs", r))
	assert.Equal(t, "userURLs", ToCamelCase("user_urls", r))
	assert.Equal(t, "UserURLs", ToPascalCase("userUrls", r))
	assert.Equal(t, "UserURLs", ToPascalCase("userURLs", r))
	assert.Equal(t, "UserURLs", ToPascalCase("UserUrls", r))
	assert.Equal(t, "UserURLs", ToPascalCase("UserURLs", r))
	assert.Equal(t, "UserURLs", ToPascalCase("user_urls", r))
	assert.Equal(t, "user_urls", ToSnakeCase("UserURLs", r))
	assert.Equal(t, "user-urls", ToKebabCase("UserURLs", r))
	assert.Equal(t, "urls", ToSnakeCase("URLs", r))

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

func TestCaseHelpersCompoundTransformations(t *testing.T) {
	r, err := NewReplacer(
		Rule{Token: "Http", Replacement: "HTTP"},
		Rule{Token: "Https", Replacement: "HTTPS"},
		Rule{Token: "Url", Replacement: "URL"},
	)
	require.NoError(t, err)

	assert.Equal(t, "itthps", ToSnakeCase(ToCamelCase("itthps")))
	assert.Equal(t, "itthps", ToSnakeCase(ToPascalCase("itthps")))
	assert.Equal(t, "i_tthps", ToSnakeCase(ToCamelCase("i_tthps")))
	assert.Equal(t, "i_tthps", ToSnakeCase(ToPascalCase("i_tthps")))
	assert.Equal(t, "ihttps", ToSnakeCase(ToCamelCase("i_https", r), r))
	assert.Equal(t, "ihttps", ToSnakeCase(ToPascalCase("i_https", r), r))
	assert.Equal(t, "user_urls", ToSnakeCase(ToCamelCase("user_urls", r), r))
	assert.Equal(t, "user_urls", ToSnakeCase(ToPascalCase("user_urls", r), r))
	assert.Equal(t, "x-url-value", ToKebabCase(ToCamelCase("x_url_value", r), r))
	assert.Equal(t, "x_url_value", ToSnakeCase(ToPascalCase("x_url_value", r), r))
}

func TestCaseHelpersOverlappingCompoundRulesBacktrack(t *testing.T) {
	r, err := NewReplacer(
		Rule{Token: "A", Replacement: "A"},
		Rule{Token: "Ab", Replacement: "AB"},
		Rule{Token: "Bc", Replacement: "BC"},
	)
	require.NoError(t, err)

	assert.Equal(t, "ABC", r.Replace("Abc"))
	assert.Equal(t, "aBC", ToCamelCase("abc", r))
	assert.Equal(t, "ABC", ToPascalCase("abc", r))
}
