package provider

import (
	"errors"
	"net/http"
	"strings"

	"github.com/alibaba/higress/plugins/wasm-go/extensions/ai-proxy/util"
	"github.com/higress-group/proxy-wasm-go-sdk/proxywasm/types"
	"github.com/higress-group/wasm-go/pkg/log"
	"github.com/higress-group/wasm-go/pkg/wrapper"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

const (
	volcengineDomain              = "ark.cn-beijing.volces.com"
	volcengineChatCompletionPath  = "/api/v3/chat/completions"
	volcengineEmbeddingsPath      = "/api/v3/embeddings"
	volcengineImageGenerationPath = "/api/v3/images/generations"
	volcengineResponsesPath       = "/api/v3/responses"
)

type volcengineProviderInitializer struct{}

func (m *volcengineProviderInitializer) ValidateConfig(config *ProviderConfig) error {
	if config.apiTokens == nil || len(config.apiTokens) == 0 {
		return errors.New("no apiToken found in provider config")
	}
	return nil
}

func (m *volcengineProviderInitializer) DefaultCapabilities() map[string]string {
	return map[string]string{
		string(ApiNameChatCompletion):  volcengineChatCompletionPath,
		string(ApiNameEmbeddings):      volcengineEmbeddingsPath,
		string(ApiNameImageGeneration): volcengineImageGenerationPath,
		string(ApiNameResponses):       volcengineResponsesPath,
	}
}

func (m *volcengineProviderInitializer) CreateProvider(config ProviderConfig) (Provider, error) {
	config.setDefaultCapabilities(m.DefaultCapabilities())
	return &volcengineProvider{
		config:       config,
		contextCache: createContextCache(&config),
	}, nil
}

type volcengineProvider struct {
	config       ProviderConfig
	contextCache *contextCache
}

func (m *volcengineProvider) GetProviderType() string {
	return providerTypeVolcengine
}

func (m *volcengineProvider) OnRequestHeaders(ctx wrapper.HttpContext, apiName ApiName) error {
	m.config.handleRequestHeaders(m, ctx, apiName)
	return nil
}

func (m *volcengineProvider) OnRequestBody(ctx wrapper.HttpContext, apiName ApiName, body []byte) (types.Action, error) {
	if !m.config.isSupportedAPI(apiName) {
		return types.ActionContinue, errUnsupportedApiName
	}
	return m.config.handleRequestBody(m, m.contextCache, ctx, apiName, body)
}

func (m *volcengineProvider) TransformRequestHeaders(ctx wrapper.HttpContext, apiName ApiName, headers http.Header) {
	util.OverwriteRequestPathHeaderByCapability(headers, string(apiName), m.config.capabilities)
	util.OverwriteRequestHostHeader(headers, volcengineDomain)
	util.OverwriteRequestAuthorizationHeader(headers, "Bearer "+m.config.GetApiTokenInUse(ctx))
	m.setProviderSpecificRequestHeaders(headers)
	headers.Del("Content-Length")
}

func (m *volcengineProvider) setProviderSpecificRequestHeaders(headers http.Header) {
	if m.config.volcengineClientRequestID != "" {
		headers.Set("X-Client-Request-Id", m.config.volcengineClientRequestID)
	}
	if m.config.volcengineEnableEncryption {
		headers.Set("x-is-encrypted", "true")
	}
	if m.config.volcengineEnableTrace {
		headers.Set("X-Fornax-Trace", "true")
	}
}

func (m *volcengineProvider) TransformRequestBody(ctx wrapper.HttpContext, apiName ApiName, body []byte) ([]byte, error) {
	var err error
	switch apiName {
	case ApiNameResponses:
		// 移除火山 responses 接口暂时不支持的参数
		// 参考: https://www.volcengine.com/docs/82379/1569618
		// TODO: 这里应该用 DTO 处理
		for _, param := range []string{"parallel_tool_calls", "tool_choice"} {
			body, err = sjson.DeleteBytes(body, param)
			if err != nil {
				log.Warnf("[volcengine] failed to delete %s in request body, err: %v", param, err)
			}
		}
	case ApiNameImageGeneration:
		// 火山生图接口默认会带上水印,但 OpenAI 接口不支持此参数
		// 参考: https://www.volcengine.com/docs/82379/1541523
		if res := gjson.GetBytes(body, "watermark"); !res.Exists() {
			body, err = sjson.SetBytes(body, "watermark", false)
			if err != nil {
				log.Warnf("[volcengine] failed to set watermark in request body, err: %v", err)
			}
		}
	}
	return m.config.defaultTransformRequestBody(ctx, apiName, body)
}

func (m *volcengineProvider) GetApiName(path string) ApiName {
	if strings.Contains(path, volcengineChatCompletionPath) {
		return ApiNameChatCompletion
	}
	if strings.Contains(path, volcengineEmbeddingsPath) {
		return ApiNameEmbeddings
	}
	if strings.Contains(path, volcengineImageGenerationPath) {
		return ApiNameImageGeneration
	}
	if strings.Contains(path, volcengineResponsesPath) {
		return ApiNameResponses
	}
	return ""
}
