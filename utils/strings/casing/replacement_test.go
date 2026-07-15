package casing

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewReplacer(t *testing.T) {
	r, err := NewReplacer(Rule{
		Token:       "Api",
		Replacement: "API",
		Exceptions:  []string{" apiClient ", "apiClient"},
	})
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "apiClient", r.Replace("apiClient"))
}

func TestNewReplacerRejectsInvalidRule(t *testing.T) {
	_, err := NewReplacer(Rule{Token: ""})
	require.Error(t, err)
}

func TestRuleIsCompatible(t *testing.T) {
	assert.True(t, Rule{Token: "Api", Replacement: "API"}.IsCompatible(&Rule{Token: " api ", Replacement: "API"}))
	assert.False(t, Rule{Token: "Api", Replacement: "API"}.IsCompatible(&Rule{Token: "Api", Replacement: "HTTP"}))
	assert.False(t, Rule{Token: "Api", Replacement: "API"}.IsCompatible(&Rule{Token: "Http", Replacement: "API"}))
	assert.False(t, Rule{Token: "Api", Replacement: "API"}.IsCompatible(nil))
}

func TestRuleMerge(t *testing.T) {
	merged := Rule{Token: "Api", Replacement: "API", Exceptions: []string{"client", " request "}}.Merge(
		&Rule{Token: "api", Replacement: "API", Exceptions: []string{"request", "identifier"}},
	)

	assert.Equal(t, "Api", merged.Token)
	assert.Equal(t, "API", merged.Replacement)
	assert.ElementsMatch(t, []string{"client", "request", "identifier"}, merged.Exceptions)

	assert.Equal(t, merged, merged.Merge(nil))
}

func TestMergeRules(t *testing.T) {
	merged := MergeRules(
		Rule{Token: "Api", Replacement: "API", Exceptions: []string{"client"}},
		Rule{Token: "api", Replacement: "API", Exceptions: []string{"identifier"}},
		Rule{Token: "Api", Replacement: "HTTP"},
	)

	require.Len(t, merged, 1)
	assert.Equal(t, Rule{Token: "Api", Replacement: "HTTP"}, merged[0])
}

func TestNewReplacerMergesCompatibleRules(t *testing.T) {
	r, err := NewReplacer(
		Rule{Token: "Api", Replacement: "API", Exceptions: []string{"client"}},
		Rule{Token: "api", Replacement: "API", Exceptions: []string{"identifier"}},
	)
	require.NoError(t, err)

	assert.Equal(t, "APIClient", r.Replace("ApiClient"))
	assert.Equal(t, "APIIdentifier", r.Replace("ApiIdentifier"))
	assert.Equal(t, "APIKey", r.Replace("ApiKey"))
}

func TestNewReplacerKeepsLastConflictingRule(t *testing.T) {
	r, err := NewReplacer(
		Rule{Token: "Api", Replacement: "API"},
		Rule{Token: "Api", Replacement: "HTTP"},
	)
	require.NoError(t, err)

	assert.Equal(t, "HTTPClient", r.Replace("ApiClient"))
}

func TestReplacerReplace(t *testing.T) {
	r, err := NewReplacer(
		Rule{Token: "Ai", Replacement: "AI"},
		Rule{Token: "Api", Replacement: "API"},
		Rule{Token: "Http", Replacement: "HTTP"},
		Rule{Token: "Https", Replacement: "HTTPS"},
		Rule{Token: "Id", Replacement: "ID", Exceptions: []string{"identifier", "idempotent"}},
		Rule{Token: "Url", Replacement: "URL"},
		Rule{Token: "Xss", Replacement: "XSS"},
	)
	require.NoError(t, err)

	assert.Equal(t, "APIClient", r.Replace("ApiClient"))
	assert.Equal(t, "apiClient", r.Replace("apiClient"))
	assert.Equal(t, "OpenAIAPIKey", r.Replace("OpenAiApiKey"))
	assert.Equal(t, "openAIAPIKey", r.Replace("openAiApiKey"))
	assert.Equal(t, "kemID", r.Replace("kemId"))
	assert.Equal(t, "AdrienIdentifier", r.Replace("AdrienIdentifier"))
	assert.Equal(t, "idempotentRetry", r.Replace("idempotentRetry"))
	assert.Equal(t, "OpenAIAPIKey", r.Replace("OpenAIAPIKey"))
	assert.Equal(t, "openAIAPIKey", r.Replace("openAIAPIKey"))
	assert.Equal(t, "sourceName", r.Replace("sourceName"))
	assert.Equal(t, "", r.Replace(""))

	assert.Equal(t, "AIAPI", r.Replace("AIAPI"))
	assert.Equal(t, "https", r.Replace("https"))
	assert.Equal(t, "HTTPS", r.Replace("Https"))
	assert.Equal(t, "HTTPS", r.Replace("HTTPS"))
	assert.Equal(t, "IHTTP", r.Replace("IHTTP"))
	assert.Equal(t, "IHTTPS", r.Replace("IHTTPS"))
	assert.Equal(t, "IHTTPS", r.Replace("IHttps"))
	assert.Equal(t, "aHTTPClient", r.Replace("aHTTPClient"))
	assert.Equal(t, "xss", r.Replace("xss"))
	assert.Equal(t, "XSS", r.Replace("Xss"))
	assert.Equal(t, "xURLValue", r.Replace("xURLValue"))
	assert.Equal(t, "URLs", r.Replace("Urls"))
	assert.Equal(t, "userURLs", r.Replace("userUrls"))
	assert.Equal(t, "UserURLs", r.Replace("UserUrls"))
	assert.Equal(t, "urls", r.Replace("urls"))
	assert.Equal(t, "HTTPAPIToken", r.Replace("HTTPApiToken"))
	assert.Equal(t, "HTTPAPIToken", r.Replace("HTTPAPIToken"))
}

func TestReplacerWriteString(t *testing.T) {
	r, err := NewReplacer(Rule{Token: "Api", Replacement: "API"})
	require.NoError(t, err)

	var buf bytes.Buffer
	n, err := r.WriteString(context.Background(), &buf, "ApiClient")
	require.NoError(t, err)
	assert.Equal(t, len("APIClient"), n)
	assert.Equal(t, "APIClient", buf.String())
}

func TestInitialismRules(t *testing.T) {
	r, err := NewReplacer(InitialismRules...)
	require.NoError(t, err)

	assert.Equal(t, "UserID", r.Replace("UserId"))
	assert.Equal(t, "userID", r.Replace("userId"))
	assert.Equal(t, "IHTTP", r.Replace("IHTTP"))
	assert.Equal(t, "https", r.Replace("https"))
	assert.Equal(t, "HTTPS", r.Replace("Https"))
	assert.Equal(t, "HTTPS", r.Replace("HTTPS"))
	assert.Equal(t, "IHTTPS", r.Replace("IHTTPS"))
	assert.Equal(t, "IHTTPS", r.Replace("IHttps"))
	assert.Equal(t, "aHTTPClient", r.Replace("aHTTPClient"))
	assert.Equal(t, "xss", r.Replace("xss"))
	assert.Equal(t, "XSS", r.Replace("Xss"))
	assert.Equal(t, "xURLValue", r.Replace("xURLValue"))
	assert.Equal(t, "URLs", r.Replace("Urls"))
	assert.Equal(t, "UserURLs", r.Replace("UserUrls"))
	assert.Equal(t, "userURLs", r.Replace("userUrls"))
	assert.Equal(t, "HTTPAPIToken", r.Replace("HttpApiToken"))
	assert.Equal(t, "JSONBlob", r.Replace("JsonBlob"))
}

func TestInitialismReplacer(t *testing.T) {
	require.NotNil(t, InitialismReplacer)
	assert.Equal(t, "UserID", InitialismReplacer.Replace("UserId"))
	assert.Equal(t, "userID", InitialismReplacer.Replace("userId"))
	assert.Equal(t, "IHTTP", InitialismReplacer.Replace("IHTTP"))
	assert.Equal(t, "https", InitialismReplacer.Replace("https"))
	assert.Equal(t, "HTTPS", InitialismReplacer.Replace("Https"))
	assert.Equal(t, "HTTPS", InitialismReplacer.Replace("HTTPS"))
	assert.Equal(t, "IHTTPS", InitialismReplacer.Replace("IHTTPS"))
	assert.Equal(t, "IHTTPS", InitialismReplacer.Replace("IHttps"))
	assert.Equal(t, "aHTTPClient", InitialismReplacer.Replace("aHTTPClient"))
	assert.Equal(t, "xss", InitialismReplacer.Replace("xss"))
	assert.Equal(t, "XSS", InitialismReplacer.Replace("Xss"))
	assert.Equal(t, "xURLValue", InitialismReplacer.Replace("xURLValue"))
	assert.Equal(t, "URLs", InitialismReplacer.Replace("Urls"))
	assert.Equal(t, "UserURLs", InitialismReplacer.Replace("UserUrls"))
	assert.Equal(t, "userURLs", InitialismReplacer.Replace("userUrls"))
	assert.Equal(t, "HTTPAPIToken", InitialismReplacer.Replace("HttpApiToken"))
}
