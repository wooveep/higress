---
name: provider
description: "Skill for the Provider area of higress. 651 symbols across 66 files."
---

# Provider

651 symbols | 66 files | Cohesion: 77%

## When to Use

- Working with code in `plugins/`
- Understanding how GetRequestHeaders, ReplaceRequestHeaders, TestMergeConsecutiveMessages work
- Modifying provider-related functionality

## Key Files

| File | Symbols |
|------|---------|
| `plugins/wasm-go/extensions/ai-proxy/provider/bedrock.go` | OnRequestBody, TransformRequestHeaders, OnRequestHeaders, DefaultCapabilities, CreateProvider (+60) |
| `plugins/wasm-go/extensions/ai-proxy/provider/vertex.go` | isOpenAICompatibleMode, OnRequestBody, OnRequestHeaders, DefaultCapabilities, CreateProvider (+42) |
| `plugins/wasm-go/extensions/ai-proxy/provider/failover.go` | GetApiTokenInUse, handleAvailableApiToken, handleUnavailableApiToken, addApiToken, removeApiToken (+30) |
| `plugins/wasm-go/extensions/ai-proxy/provider/provider.go` | IsOriginal, isDeveloperRoleSupported, convertDeveloperRoleToSystem, isSupportedAPI, handleRequestBody (+26) |
| `plugins/wasm-go/extensions/ai-proxy/provider/gemini.go` | OnRequestBody, TransformRequestHeaders, OnRequestHeaders, DefaultCapabilities, CreateProvider (+26) |
| `plugins/wasm-go/extensions/ai-proxy/provider/qwen.go` | OnRequestBody, TransformRequestHeaders, OnRequestHeaders, DefaultCapabilities, CreateProvider (+17) |
| `plugins/wasm-go/extensions/ai-proxy/provider/bedrock_thinking_test.go` | TestBedrockRequestPreservesClaudeNativeThinkingAndToolResult, TestBedrockRequestPreservesClaudeNoArgToolUseInput, TestBedrockRequestToolResultWithTrailingTextDoesNotDuplicateToolResult, TestBedrockRequestPreservesClaudeNativeThinkingBudget, TestBedrockRequestMapsAdaptiveOutputEffortIntoThinking (+17) |
| `plugins/wasm-go/extensions/ai-proxy/provider/claude.go` | OnRequestBody, TransformRequestHeaders, OnRequestHeaders, NewStringContent, NewArrayContent (+15) |
| `plugins/wasm-go/extensions/ai-proxy/provider/model.go` | GetURL, GetImageURLs, HasMask, IsEmpty, IsStringContent (+15) |
| `plugins/wasm-go/extensions/ai-proxy/provider/hunyuan.go` | TransformRequestHeaders, OnRequestHeaders, DefaultCapabilities, CreateProvider, convertMessagesFromOpenAIToHunyuan (+12) |

## Entry Points

Start here when exploring this area:

- **`GetRequestHeaders`** (Function) — `plugins/wasm-go/extensions/ai-proxy/util/http.go:224`
- **`ReplaceRequestHeaders`** (Function) — `plugins/wasm-go/extensions/ai-proxy/util/http.go:234`
- **`TestMergeConsecutiveMessages`** (Function) — `plugins/wasm-go/extensions/ai-proxy/provider/request_helper_test.go:10`
- **`TestCleanupContextMessages`** (Function) — `plugins/wasm-go/extensions/ai-proxy/provider/request_helper_test.go:135`
- **`TestStripClaudeInternalMessageFields`** (Function) — `plugins/wasm-go/extensions/ai-proxy/provider/provider_test.go:720`

## Key Symbols

| Symbol | Type | File | Line |
|--------|------|------|------|
| `GetRequestHeaders` | Function | `plugins/wasm-go/extensions/ai-proxy/util/http.go` | 224 |
| `ReplaceRequestHeaders` | Function | `plugins/wasm-go/extensions/ai-proxy/util/http.go` | 234 |
| `TestMergeConsecutiveMessages` | Function | `plugins/wasm-go/extensions/ai-proxy/provider/request_helper_test.go` | 10 |
| `TestCleanupContextMessages` | Function | `plugins/wasm-go/extensions/ai-proxy/provider/request_helper_test.go` | 135 |
| `TestStripClaudeInternalMessageFields` | Function | `plugins/wasm-go/extensions/ai-proxy/provider/provider_test.go` | 720 |
| `TestUtil` | Function | `plugins/wasm-go/extensions/ai-proxy/main_test.go` | 298 |
| `OverwriteRequestPath` | Function | `plugins/wasm-go/extensions/ai-proxy/util/http.go` | 59 |
| `OverwriteRequestHostHeader` | Function | `plugins/wasm-go/extensions/ai-proxy/util/http.go` | 67 |
| `OverwriteRequestPathHeader` | Function | `plugins/wasm-go/extensions/ai-proxy/util/http.go` | 71 |
| `OverwriteRequestPathHeaderByCapability` | Function | `plugins/wasm-go/extensions/ai-proxy/util/http.go` | 75 |
| `MapRequestPathByCapability` | Function | `plugins/wasm-go/extensions/ai-proxy/util/http.go` | 85 |
| `GetOriginalRequestPath` | Function | `plugins/wasm-go/extensions/ai-proxy/util/http.go` | 155 |
| `OverwriteRequestAuthorizationHeader` | Function | `plugins/wasm-go/extensions/ai-proxy/util/http.go` | 200 |
| `TestOverwriteRequestPathHeader` | Function | `plugins/wasm-go/extensions/ai-proxy/util/header_slice_test.go` | 32 |
| `RunMapRequestPathByCapabilityTests` | Function | `plugins/wasm-go/extensions/ai-proxy/test/util.go` | 16 |
| `TestLongcatProvider_GetProviderType` | Function | `plugins/wasm-go/extensions/ai-proxy/provider/longcat_test.go` | 72 |
| `TestClaudeProvider_GetProviderType` | Function | `plugins/wasm-go/extensions/ai-proxy/provider/claude_test.go` | 75 |
| `TestBedrockRequestPreservesClaudeNativeThinkingAndToolResult` | Function | `plugins/wasm-go/extensions/ai-proxy/provider/bedrock_thinking_test.go` | 207 |
| `TestBedrockRequestPreservesClaudeNoArgToolUseInput` | Function | `plugins/wasm-go/extensions/ai-proxy/provider/bedrock_thinking_test.go` | 253 |
| `TestBedrockRequestToolResultWithTrailingTextDoesNotDuplicateToolResult` | Function | `plugins/wasm-go/extensions/ai-proxy/provider/bedrock_thinking_test.go` | 282 |

## Execution Flows

| Flow | Type | Steps |
|------|------|-------|
| `OnHttpRequestHeader → GetConsumerFromContext` | cross_community | 5 |
| `OnHttpRequestHeader → GetTokenWithConsumerAffinity` | cross_community | 5 |
| `OnHttpRequestHeader → GetRandomToken` | cross_community | 5 |
| `OnHttpRequestHeader → IsStatefulAPI` | cross_community | 5 |
| `OnHttpResponseHeaders → HealthCheckEndpoint` | cross_community | 5 |
| `OnHttpResponseHeaders → GetApiTokenRequestCount` | cross_community | 5 |
| `OnStreamingResponseBody → IsContentEmpty` | cross_community | 5 |
| `OnRequestBody → HmacSha256` | cross_community | 5 |
| `OnStreamingResponseBody → ChatCompletionResponse` | cross_community | 5 |
| `OnHttpRequestHeader → GetApiTokens` | cross_community | 4 |

## Connected Areas

| Area | Connections |
|------|-------------|
| Ai-proxy | 6 calls |
| Basic_auth | 6 calls |
| Util | 2 calls |
| Istio | 1 calls |

## How to Explore

1. `gitnexus_context({name: "GetRequestHeaders"})` — see callers and callees
2. `gitnexus_query({query: "provider"})` — find related execution flows
3. Read key files listed above for implementation details
