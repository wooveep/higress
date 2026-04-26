package util

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeConsumerAuthHeaders(t *testing.T) {
	headers := http.Header{}
	headers.Set(HeaderAuthorization, "Bearer provider-token")
	headers.Set(HeaderXAPIKey, "consumer-key")
	headers.Set(HeaderAnthropicAPIKey, "anthropic-key")
	headers.Set(HeaderXGoogAPIKey, "google-key")
	headers.Set("traceparent", "00-abcdef-1234567890abcdef-01")

	SanitizeConsumerAuthHeaders(headers)

	assert.Empty(t, headers.Get(HeaderAuthorization))
	assert.Empty(t, headers.Get(HeaderXAPIKey))
	assert.Empty(t, headers.Get(HeaderAnthropicAPIKey))
	assert.Empty(t, headers.Get(HeaderXGoogAPIKey))
	assert.Equal(t, "00-abcdef-1234567890abcdef-01", headers.Get("traceparent"))
}

func TestOverwriteRequestAuthorizationHeaderDeletesEmptyCredential(t *testing.T) {
	headers := http.Header{}
	headers.Set(HeaderAuthorization, "Bearer provider-token")

	OverwriteRequestAuthorizationHeader(headers, "")

	assert.Empty(t, headers.Get(HeaderAuthorization))
}
