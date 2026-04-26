package provider

import (
	"net/http"
	"testing"

	"github.com/alibaba/higress/plugins/wasm-go/extensions/ai-proxy/util"
	"github.com/higress-group/proxy-wasm-go-sdk/proxywasm"
	"github.com/higress-group/proxy-wasm-go-sdk/proxywasm/proxytest"
	"github.com/higress-group/proxy-wasm-go-sdk/proxywasm/types"
	"github.com/stretchr/testify/assert"
)

func setProviderToken(ctx *fakeHttpContext, config *ProviderConfig, token string) {
	if config.failover == nil {
		config.failover = &failover{}
	}
	config.failover.ctxApiTokenInUse = "test-api-token"
	ctx.SetContext(config.failover.ctxApiTokenInUse, token)
}

func withRequestPath(t *testing.T, path string, fn func()) {
	t.Helper()

	host, reset := proxytest.NewHostEmulator(proxytest.NewEmulatorOption().WithHttpContext(func(uint32) types.HttpContext {
		return &types.DefaultHttpContext{}
	}))
	t.Cleanup(reset)

	contextID := host.InitializeHttpContext()
	host.SetHttpRequestHeaders(contextID, [][2]string{{util.HeaderPath, path}})
	assert.NoError(t, proxywasm.SetEffectiveContext(contextID))

	fn()
}

func TestQwenProviderTransformRequestHeadersRewritesConsumerAuth(t *testing.T) {
	ctx := newFakeHTTPContext()
	config := ProviderConfig{protocol: protocolOriginal}
	setProviderToken(ctx, &config, "provider-token")

	provider := &qwenProvider{config: config}
	headers := http.Header{}
	headers.Set(util.HeaderAuthorization, "Bearer consumer-token")
	headers.Set(util.HeaderXAPIKey, "consumer-key")
	headers.Set(util.HeaderAnthropicAPIKey, "consumer-anthropic-key")

	provider.TransformRequestHeaders(ctx, ApiNameAnthropicMessages, headers)

	assert.Equal(t, "Bearer provider-token", headers.Get(util.HeaderAuthorization))
	assert.Equal(t, "provider-token", headers.Get(util.HeaderXAPIKey))
	assert.Empty(t, headers.Get(util.HeaderAnthropicAPIKey))
	assert.Equal(t, "dashscope.aliyuncs.com", headers.Get(util.HeaderAuthority))
}

func TestGeminiProviderTransformRequestHeadersKeepsOnlyGoogleAPIKey(t *testing.T) {
	ctx := newFakeHTTPContext()
	config := ProviderConfig{}
	setProviderToken(ctx, &config, "provider-token")

	provider := &geminiProvider{config: config}
	headers := http.Header{}
	headers.Set(util.HeaderAuthorization, "Bearer consumer-token")
	headers.Set(util.HeaderXAPIKey, "consumer-key")
	headers.Set(util.HeaderAnthropicAPIKey, "consumer-anthropic-key")

	provider.TransformRequestHeaders(ctx, ApiNameChatCompletion, headers)

	assert.Empty(t, headers.Get(util.HeaderAuthorization))
	assert.Empty(t, headers.Get(util.HeaderXAPIKey))
	assert.Empty(t, headers.Get(util.HeaderAnthropicAPIKey))
	assert.Equal(t, "provider-token", headers.Get(geminiApiKeyHeader))
}

func TestClaudeProviderTransformRequestHeadersStandardModeRewritesXAPIKey(t *testing.T) {
	ctx := newFakeHTTPContext()
	config := ProviderConfig{}
	setProviderToken(ctx, &config, "provider-token")

	provider := &claudeProvider{config: config}
	headers := http.Header{}
	headers.Set(util.HeaderAuthorization, "Bearer consumer-token")
	headers.Set(util.HeaderXAPIKey, "consumer-key")
	headers.Set(":path", "/v1/messages")

	withRequestPath(t, "/v1/messages", func() {
		provider.TransformRequestHeaders(ctx, ApiNameAnthropicMessages, headers)
	})

	assert.Empty(t, headers.Get(util.HeaderAuthorization))
	assert.Equal(t, "provider-token", headers.Get(util.HeaderXAPIKey))
	assert.Equal(t, claudeDefaultVersion, headers.Get("anthropic-version"))
}

func TestClaudeProviderTransformRequestHeadersClaudeCodeModeKeepsOnlyBearer(t *testing.T) {
	ctx := newFakeHTTPContext()
	config := ProviderConfig{claudeCodeMode: true}
	setProviderToken(ctx, &config, "provider-token")

	provider := &claudeProvider{config: config}
	headers := http.Header{}
	headers.Set(util.HeaderAuthorization, "Bearer consumer-token")
	headers.Set(util.HeaderXAPIKey, "consumer-key")
	headers.Set(":path", "/v1/messages")

	withRequestPath(t, "/v1/messages", func() {
		provider.TransformRequestHeaders(ctx, ApiNameAnthropicMessages, headers)
	})

	assert.Equal(t, "Bearer provider-token", headers.Get(util.HeaderAuthorization))
	assert.Empty(t, headers.Get(util.HeaderXAPIKey))
	assert.Equal(t, "cli", headers.Get("x-app"))
	assert.Equal(t, "/v1/messages?beta=true", headers.Get(":path"))
}

func TestVertexProviderTransformRequestHeadersSanitizesConsumerAuth(t *testing.T) {
	provider := &vertexProvider{
		config: ProviderConfig{
			apiTokens: []string{"provider-token"},
		},
	}
	headers := http.Header{}
	headers.Set(util.HeaderAuthorization, "Bearer consumer-token")
	headers.Set(util.HeaderXAPIKey, "consumer-key")
	headers.Set(util.HeaderXGoogAPIKey, "consumer-google-key")

	provider.TransformRequestHeaders(newFakeHTTPContext(), ApiNameChatCompletion, headers)

	assert.Empty(t, headers.Get(util.HeaderAuthorization))
	assert.Empty(t, headers.Get(util.HeaderXAPIKey))
	assert.Empty(t, headers.Get(util.HeaderXGoogAPIKey))
	assert.Equal(t, vertexDomain, headers.Get(util.HeaderAuthority))
}
