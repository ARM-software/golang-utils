package casing

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrepareCaseInput(t *testing.T) {
	assert.Equal(t, "", prepareCaseInput(""))
	assert.Equal(t, "source_name", prepareCaseInput("source_name"))
	assert.Equal(t, "HTTPAPI_Token", prepareCaseInput("HTTPAPIToken"))
	assert.Equal(t, "I_HTTP", prepareCaseInput("IHTTP"))
}

func TestSplitCamelWords(t *testing.T) {
	assert.Nil(t, splitCamelWords(""))
	assert.ElementsMatch(t, []string{"source", "Name"}, splitCamelWords("sourceName"))
	assert.ElementsMatch(t, []string{"HTTPAPI", "Token"}, splitCamelWords("HTTPAPIToken"))
	assert.ElementsMatch(t, []string{"HTTPS"}, splitCamelWords("HTTPS"))
	assert.ElementsMatch(t, []string{"Https"}, splitCamelWords("Https"))
	assert.ElementsMatch(t, []string{"I", "HTTP"}, splitCamelWords("IHTTP"))
	assert.ElementsMatch(t, []string{"I", "HTTPS"}, splitCamelWords("IHTTPS"))
	assert.ElementsMatch(t, []string{"I", "Https"}, splitCamelWords("IHttps"))
	assert.ElementsMatch(t, []string{"i", "Http"}, splitCamelWords("iHttp"))
	assert.ElementsMatch(t, []string{"uRLs"}, splitCamelWords("uRLs"))
	assert.ElementsMatch(t, []string{"URLs"}, splitCamelWords("URLs"))
	assert.ElementsMatch(t, []string{"user", "URLs"}, splitCamelWords("userURLs"))
	assert.ElementsMatch(t, []string{"User", "Urls"}, splitCamelWords("UserUrls"))
	assert.ElementsMatch(t, []string{"a", "HTTP", "Client"}, splitCamelWords("aHTTPClient"))
	assert.ElementsMatch(t, []string{"x", "URL", "Value"}, splitCamelWords("xURLValue"))
	assert.ElementsMatch(t, []string{"No", "Keyring"}, splitCamelWords("NoKeyring"))
}

func TestSplitLeadingLetterCompoundAvoidsPascalCaseWords(t *testing.T) {
	r, err := NewReplacer(InitialismRules...)
	require.NoError(t, err)

	parts, ok := splitLeadingLetterCompound("aHTTPClient", r)
	require.True(t, ok)
	assert.Equal(t, []string{"a", "HTTP", "Client"}, parts)

	parts, ok = splitLeadingLetterCompound("xURLValue", r)
	require.True(t, ok)
	assert.Equal(t, []string{"x", "URL", "Value"}, parts)

	parts, ok = splitLeadingLetterCompound("NoKeyring", r)
	assert.False(t, ok)
	assert.Nil(t, parts)

	parts, ok = splitLeadingLetterCompound("nOKeyring", r)
	assert.False(t, ok)
	assert.Nil(t, parts)
}

func TestFormSnakeCasedWords(t *testing.T) {
	assert.Equal(t, "", formSnakeCasedWords(nil))
	assert.Equal(t, "http_api_token", formSnakeCasedWords([]string{"HTTP", "API", "Token"}))
	assert.Equal(t, "x_url_value", formSnakeCasedWords([]string{"x", "URL", "Value"}))
}

func TestFormKebabCasedWords(t *testing.T) {
	assert.Equal(t, "", formKebabCasedWords(nil))
	assert.Equal(t, "http-api-token", formKebabCasedWords([]string{"HTTP", "API", "Token"}))
	assert.Equal(t, "x-url-value", formKebabCasedWords([]string{"x", "URL", "Value"}))
}

func TestIsIdentifierWithoutSeparators(t *testing.T) {
	assert.False(t, isIdentifierWithoutSeparators(""))
	assert.True(t, isIdentifierWithoutSeparators("SourceName1"))
	assert.False(t, isIdentifierWithoutSeparators("source_name"))
	assert.False(t, isIdentifierWithoutSeparators("source-name"))
	assert.True(t, isIdentifierWithoutSeparators("Éclair"))
}

func TestHasUppercase(t *testing.T) {
	assert.False(t, hasUppercase(""))
	assert.False(t, hasUppercase("lowercase"))
	assert.True(t, hasUppercase("Lowercase"))
	assert.True(t, hasUppercase("HTTP"))
}

func TestLowerFirstWord(t *testing.T) {
	assert.Equal(t, "", lowerFirstWord(""))
	assert.Equal(t, "hTTP", lowerFirstWord("HTTP"))
	assert.Equal(t, "aPIClient", lowerFirstWord("APIClient"))
	assert.Equal(t, "xURLValue", lowerFirstWord("XURLValue"))
	assert.Equal(t, "uRLs", lowerFirstWord("URLs"))
}

func TestUpperFirstWord(t *testing.T) {
	assert.Equal(t, "", upperFirstWord(""))
	assert.Equal(t, "APIClient", upperFirstWord("aPIClient"))
	assert.Equal(t, "XURLValue", upperFirstWord("xURLValue"))
}

func TestStartsWithLowercase(t *testing.T) {
	assert.False(t, startsWithLowercase(""))
	assert.True(t, startsWithLowercase("abc"))
	assert.False(t, startsWithLowercase("Abc"))
	assert.False(t, startsWithLowercase("1abc"))
	assert.True(t, startsWithLowercase("éclair"))
}

func TestReplaceIdentifierWords(t *testing.T) {
	r, err := NewReplacer(
		Rule{Token: "Http", Replacement: "HTTP"},
		Rule{Token: "Https", Replacement: "HTTPS"},
		Rule{Token: "Url", Replacement: "URL"},
	)
	require.NoError(t, err)

	assert.Equal(t, "HTTPClient", replaceIdentifierWords("HttpClient", r, false))
	assert.Equal(t, "httpClient", replaceIdentifierWords("HttpClient", r, true))
	assert.Equal(t, "URLs", replaceIdentifierWords("Urls", r, false))
	assert.Equal(t, "urls", replaceIdentifierWords("Urls", r, true))
	assert.Equal(t, "xURLValue", replaceIdentifierWords("xUrlValue", r, true))
	assert.Equal(t, "XURLValue", replaceIdentifierWords("xUrlValue", r, false))
	assert.Equal(t, "aHTTPClient", replaceIdentifierWords("aHTTPClient", r, true))
	assert.Equal(t, "AHTTPClient", replaceIdentifierWords("aHTTPClient", r, false))
}

func TestIsUpperInitialismOrPlural(t *testing.T) {
	assert.False(t, isUpperInitialismOrPlural(""))
	assert.True(t, isUpperInitialismOrPlural("HTTP"))
	assert.True(t, isUpperInitialismOrPlural("HTTP2"))
	assert.True(t, isUpperInitialismOrPlural("URLs"))
	assert.False(t, isUpperInitialismOrPlural("Http"))
	assert.False(t, isUpperInitialismOrPlural("UserURLs"))
}

func TestSplitLeadingLetterCompound(t *testing.T) {
	r, err := NewReplacer(
		Rule{Token: "Http", Replacement: "HTTP"},
		Rule{Token: "Url", Replacement: "URL"},
	)
	require.NoError(t, err)

	parts, ok := splitLeadingLetterCompound("xURLValue", r)
	require.True(t, ok)
	assert.ElementsMatch(t, []string{"x", "URL", "Value"}, parts)

	parts, ok = splitLeadingLetterCompound("aHTTPClient", r)
	require.True(t, ok)
	assert.ElementsMatch(t, []string{"a", "HTTP", "Client"}, parts)

	parts, ok = splitLeadingLetterCompound("UserUrls", r)
	assert.False(t, ok)
	assert.Nil(t, parts)
}
