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

func TestReplacerReplace(t *testing.T) {
	r, err := NewReplacer(
		Rule{Token: "Ai", Replacement: "AI"},
		Rule{Token: "Api", Replacement: "API"},
		Rule{Token: "Id", Replacement: "ID", Exceptions: []string{"identifier", "idempotent"}},
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
	assert.Equal(t, "HTTPAPIToken", r.Replace("HttpApiToken"))
	assert.Equal(t, "JSONBlob", r.Replace("JsonBlob"))
}

func TestInitialismReplacer(t *testing.T) {
	require.NotNil(t, InitialismReplacer)
	assert.Equal(t, "UserID", InitialismReplacer.Replace("UserId"))
	assert.Equal(t, "userID", InitialismReplacer.Replace("userId"))
	assert.Equal(t, "HTTPAPIToken", InitialismReplacer.Replace("HttpApiToken"))
}
