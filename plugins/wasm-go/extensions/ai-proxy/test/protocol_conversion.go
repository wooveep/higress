package test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/higress-group/proxy-wasm-go-sdk/proxywasm/types"
	"github.com/higress-group/wasm-go/pkg/test"
	"github.com/stretchr/testify/require"
)

func providerProtocolConfig(protocol string) json.RawMessage {
	data, _ := json.Marshal(map[string]interface{}{
		"provider": map[string]interface{}{
			"type":      "openai",
			"protocol":  protocol,
			"apiTokens": []string{"sk-openai-test123456789"},
			"modelMapping": map[string]string{
				"*": "gpt-4o-mini",
			},
		},
	})
	return data
}

func RunProtocolConversionTests(t *testing.T) {
	test.RunTest(t, func(t *testing.T) {
		t.Run("explicit_openai_protocol_keeps_claude_messages_fallback", func(t *testing.T) {
			host, status := test.NewTestHost(providerProtocolConfig("openai"))
			defer host.Reset()
			require.Equal(t, types.OnPluginStartStatusOK, status)

			action := host.CallOnHttpRequestHeaders([][2]string{
				{":authority", "example.com"},
				{":path", "/v1/messages"},
				{":method", "POST"},
				{"Content-Type", "application/json"},
			})
			require.Equal(t, types.HeaderStopIteration, action)

			requestHeaders := host.GetRequestHeaders()
			path, found := test.GetHeaderValue(requestHeaders, ":path")
			require.True(t, found)
			require.Equal(t, "/v1/chat/completions", path)

			body := `{
				"model": "claude-3-7-sonnet",
				"max_tokens": 128,
				"messages": [{"role": "user", "content": "Hello from Claude"}]
			}`
			action = host.CallOnHttpRequestBody([]byte(body))
			require.Equal(t, types.ActionContinue, action)

			transformedBody := string(host.GetRequestBody())
			require.Contains(t, transformedBody, "\"messages\"")
			require.Contains(t, transformedBody, "\"role\":\"user\"")
			require.Contains(t, transformedBody, "\"max_tokens\":128")

		})

		t.Run("explicit_anthropic_protocol_is_accepted_and_reuses_conversion_chain", func(t *testing.T) {
			host, status := test.NewTestHost(providerProtocolConfig("anthropic"))
			defer host.Reset()
			require.Equal(t, types.OnPluginStartStatusOK, status)

			action := host.CallOnHttpRequestHeaders([][2]string{
				{":authority", "example.com"},
				{":path", "/v1/messages"},
				{":method", "POST"},
				{"Content-Type", "application/json"},
			})
			require.Equal(t, types.HeaderStopIteration, action)

			requestHeaders := host.GetRequestHeaders()
			path, found := test.GetHeaderValue(requestHeaders, ":path")
			require.True(t, found)
			require.Equal(t, "/v1/chat/completions", path)
		})

		t.Run("original_protocol_disables_claude_messages_fallback", func(t *testing.T) {
			host, status := test.NewTestHost(providerProtocolConfig("original"))
			defer host.Reset()
			require.Equal(t, types.OnPluginStartStatusOK, status)

			action := host.CallOnHttpRequestHeaders([][2]string{
				{":authority", "example.com"},
				{":path", "/v1/messages"},
				{":method", "POST"},
				{"Content-Type", "application/json"},
			})
			require.Equal(t, types.HeaderStopIteration, action)

			requestHeaders := host.GetRequestHeaders()
			path, found := test.GetHeaderValue(requestHeaders, ":path")
			require.True(t, found)
			require.Equal(t, "/v1/messages", path)

			body := `{
				"model": "claude-3-7-sonnet",
				"max_tokens": 128,
				"messages": [{"role": "user", "content": "Hello from Claude"}]
			}`
			action = host.CallOnHttpRequestBody([]byte(body))
			require.Equal(t, types.ActionContinue, action)

			transformedBody := strings.TrimSpace(string(host.GetRequestBody()))
			require.Contains(t, transformedBody, "\"max_tokens\"")
		})
	})
}
