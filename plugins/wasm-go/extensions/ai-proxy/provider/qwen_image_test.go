package provider

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/higress-group/wasm-go/pkg/iface"
	"github.com/stretchr/testify/require"
)

type fakeHttpContext struct {
	context map[string]any
	user    map[string]any
}

func newFakeHTTPContext() *fakeHttpContext {
	return &fakeHttpContext{
		context: make(map[string]any),
		user:    make(map[string]any),
	}
}

func (f *fakeHttpContext) Scheme() string                           { return "http" }
func (f *fakeHttpContext) Host() string                             { return "example.com" }
func (f *fakeHttpContext) Path() string                             { return "/v1/images/generations" }
func (f *fakeHttpContext) Method() string                           { return http.MethodPost }
func (f *fakeHttpContext) SetContext(key string, value interface{}) { f.context[key] = value }
func (f *fakeHttpContext) GetContext(key string) interface{}        { return f.context[key] }
func (f *fakeHttpContext) GetBoolContext(key string, defaultValue bool) bool {
	v, ok := f.context[key].(bool)
	if !ok {
		return defaultValue
	}
	return v
}
func (f *fakeHttpContext) GetStringContext(key, defaultValue string) string {
	v, ok := f.context[key].(string)
	if !ok {
		return defaultValue
	}
	return v
}
func (f *fakeHttpContext) GetByteSliceContext(key string, defaultValue []byte) []byte {
	v, ok := f.context[key].([]byte)
	if !ok {
		return defaultValue
	}
	return v
}
func (f *fakeHttpContext) GetUserAttribute(key string) interface{}          { return f.user[key] }
func (f *fakeHttpContext) SetUserAttribute(key string, value interface{})   { f.user[key] = value }
func (f *fakeHttpContext) SetUserAttributeMap(kvmap map[string]interface{}) { f.user = kvmap }
func (f *fakeHttpContext) GetUserAttributeMap() map[string]interface{}      { return f.user }
func (f *fakeHttpContext) WriteUserAttributeToLog() error                   { return nil }
func (f *fakeHttpContext) WriteUserAttributeToLogWithKey(key string) error  { return nil }
func (f *fakeHttpContext) WriteUserAttributeToTrace() error                 { return nil }
func (f *fakeHttpContext) DontReadRequestBody()                             {}
func (f *fakeHttpContext) DontReadResponseBody()                            {}
func (f *fakeHttpContext) BufferRequestBody()                               {}
func (f *fakeHttpContext) BufferResponseBody()                              {}
func (f *fakeHttpContext) NeedPauseStreamingResponse()                      {}
func (f *fakeHttpContext) PushBuffer(buffer []byte)                         {}
func (f *fakeHttpContext) PopBuffer() []byte                                { return nil }
func (f *fakeHttpContext) BufferQueueSize() int                             { return 0 }
func (f *fakeHttpContext) DisableReroute()                                  {}
func (f *fakeHttpContext) SetRequestBodyBufferLimit(byteSize uint32)        {}
func (f *fakeHttpContext) SetResponseBodyBufferLimit(byteSize uint32)       {}
func (f *fakeHttpContext) RouteCall(method, url string, headers [][2]string, body []byte, callback iface.RouteResponseCallback) error {
	return nil
}
func (f *fakeHttpContext) GetExecutionPhase() iface.HTTPExecutionPhase { return iface.DecodeData }
func (f *fakeHttpContext) HasRequestBody() bool                        { return true }
func (f *fakeHttpContext) HasResponseBody() bool                       { return true }
func (f *fakeHttpContext) IsWebsocket() bool                           { return false }
func (f *fakeHttpContext) IsBinaryRequestBody() bool                   { return false }
func (f *fakeHttpContext) IsBinaryResponseBody() bool                  { return false }

func TestQwenImageGenerationCapabilities(t *testing.T) {
	init := &qwenProviderInitializer{}
	capabilities := init.DefaultCapabilities(true)
	require.Equal(t, qwenMultimodalGenerationPath, capabilities[string(ApiNameImageGeneration)])
}

func TestQwenImageGenerationRequestTransform(t *testing.T) {
	init := &qwenProviderInitializer{}
	provider := &qwenProvider{
		config: ProviderConfig{
			typ:                  providerTypeQwen,
			protocol:             protocolOpenAI,
			apiTokens:            []string{"token"},
			qwenEnableCompatible: true,
			capabilities:         init.DefaultCapabilities(true),
		},
	}

	ctx := newFakeHTTPContext()
	headers := http.Header{}
	body := []byte(`{"model":"qwen-image-2.0-pro","prompt":"winter street","size":"1024x1024","n":2}`)

	transformed, err := provider.TransformRequestBodyHeaders(ctx, ApiNameImageGeneration, body, headers)
	require.NoError(t, err)

	var raw map[string]any
	require.NoError(t, json.Unmarshal(transformed, &raw))
	require.Equal(t, "qwen-image-2.0-pro", raw["model"])
	require.Equal(t, map[string]any{"n": float64(2), "size": "1024*1024"}, raw["parameters"])

	input := raw["input"].(map[string]any)
	messages := input["messages"].([]any)
	require.Len(t, messages, 1)

	message := messages[0].(map[string]any)
	require.Equal(t, roleUser, message["role"])
	content := message["content"].([]any)
	require.Len(t, content, 1)
	require.Equal(t, map[string]any{"text": "winter street"}, content[0])
}

func TestQwenImageGenerationResponseTransform(t *testing.T) {
	init := &qwenProviderInitializer{}
	provider := &qwenProvider{
		config: ProviderConfig{
			typ:                  providerTypeQwen,
			protocol:             protocolOpenAI,
			apiTokens:            []string{"token"},
			qwenEnableCompatible: true,
			capabilities:         init.DefaultCapabilities(true),
		},
	}

	body := []byte(`{
		"request_id":"req-1",
		"output":{
			"choices":[
				{
					"finish_reason":"stop",
					"message":{
						"role":"assistant",
						"content":[
							{"image":"https://example.com/a.png"},
							{"image":"https://example.com/b.png"}
						]
					}
				}
			]
		}
	}`)

	transformed, err := provider.TransformResponseBody(newFakeHTTPContext(), ApiNameImageGeneration, body)
	require.NoError(t, err)

	var response imageGenerationResponse
	require.NoError(t, json.Unmarshal(transformed, &response))
	require.Len(t, response.Data, 2)
	require.Equal(t, "https://example.com/a.png", response.Data[0].URL)
	require.Equal(t, "https://example.com/b.png", response.Data[1].URL)
	require.NotZero(t, response.Created)
}
