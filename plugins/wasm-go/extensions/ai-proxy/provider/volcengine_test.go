package provider

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVolcengineSetProviderSpecificRequestHeaders(t *testing.T) {
	p := &volcengineProvider{
		config: ProviderConfig{
			volcengineClientRequestID:  "request-id-1",
			volcengineEnableEncryption: true,
			volcengineEnableTrace:      true,
		},
	}
	headers := http.Header{}

	p.setProviderSpecificRequestHeaders(headers)

	require.Equal(t, "request-id-1", headers.Get("X-Client-Request-Id"))
	require.Equal(t, "true", headers.Get("x-is-encrypted"))
	require.Equal(t, "true", headers.Get("X-Fornax-Trace"))
}

func TestVolcengineGetApiName(t *testing.T) {
	p := &volcengineProvider{}

	require.Equal(t, ApiNameChatCompletion, p.GetApiName("/api/v3/chat/completions"))
	require.Equal(t, ApiNameEmbeddings, p.GetApiName("/api/v3/embeddings"))
	require.Equal(t, ApiNameImageGeneration, p.GetApiName("/api/v3/images/generations"))
	require.Equal(t, ApiNameResponses, p.GetApiName("/api/v3/responses"))
	require.Equal(t, ApiName(""), p.GetApiName("/v1/models"))
}
