---
name: test
description: "Skill for the Test area of higress. 139 symbols across 32 files."
---

# Test

139 symbols | 32 files | Cohesion: 94%

## When to Use

- Working with code in `plugins/`
- Understanding how TestProviderWasmSmoke, RunBaichuanWasmSmokeTests, RunYiWasmSmokeTests work
- Modifying test-related functionality

## Key Files

| File | Symbols |
|------|---------|
| `plugins/wasm-go/extensions/ai-proxy/main_test.go` | TestProviderWasmSmoke, TestVertex, TestQwen, TestGemini, TestBedrock (+16) |
| `plugins/wasm-go/extensions/ai-proxy/test/vertex.go` | RunVertexParseConfigTests, RunVertexExpressModeOnHttpRequestHeadersTests, RunVertexExpressModeOnHttpRequestBodyTests, RunVertexExpressModeOnHttpResponseBodyTests, RunVertexOpenAICompatibleModeOnHttpRequestHeadersTests (+10) |
| `plugins/wasm-go/extensions/ai-proxy/test/provider_wasm_smoke.go` | providerSmokeLegacyJSON, RunBaichuanWasmSmokeTests, RunYiWasmSmokeTests, RunOllamaWasmSmokeTests, RunBaiduWasmSmokeTests (+9) |
| `plugins/wasm-go/extensions/ai-proxy/test/bedrock.go` | RunBedrockParseConfigTests, RunBedrockOnHttpRequestHeadersTests, RunBedrockOnHttpResponseHeadersTests, RunBedrockToolCallTests, RunBedrockOnHttpResponseBodyTests (+6) |
| `plugins/wasm-go/extensions/ai-proxy/test/azure.go` | RunAzureParseConfigTests, RunAzureOnHttpRequestHeadersTests, RunAzureOnHttpResponseHeadersTests, RunAzureOnHttpResponseBodyTests, RunAzureBasePathHandlingTests (+3) |
| `plugins/wasm-go/extensions/ai-proxy/test/qwen.go` | hasUnsupportedAPINameError, RunQwenParseConfigTests, RunQwenOnHttpRequestHeadersTests, RunQwenOnHttpRequestBodyTests, RunQwenOnHttpResponseHeadersTests (+2) |
| `plugins/wasm-go/extensions/ai-proxy/test/gemini.go` | RunGeminiParseConfigTests, RunGeminiOnHttpRequestHeadersTests, RunGeminiOnHttpRequestBodyTests, RunGeminiOnHttpResponseHeadersTests, RunGeminiOnHttpResponseBodyTests (+2) |
| `plugins/wasm-go/extensions/ai-proxy/test/ai360.go` | RunAi360ParseConfigTests, RunAi360OnHttpRequestHeadersTests, RunAi360OnHttpRequestBodyTests, RunAi360OnHttpResponseHeadersTests, RunAi360OnHttpResponseBodyTests (+1) |
| `plugins/wasm-go/extensions/ai-proxy/test/openai.go` | RunOpenAIParseConfigTests, RunOpenAIOnHttpRequestHeadersTests, RunOpenAIOnHttpRequestBodyTests, RunOpenAIOnHttpResponseHeadersTests, RunOpenAIOnHttpResponseBodyTests (+1) |
| `plugins/wasm-go/extensions/jwt-auth/test/jwt_test.go` | genPrivateKey, genJWKs, genJWTs, TestMain |

## Entry Points

Start here when exploring this area:

- **`TestProviderWasmSmoke`** (Function) — `plugins/wasm-go/extensions/ai-proxy/main_test.go:408`
- **`RunBaichuanWasmSmokeTests`** (Function) — `plugins/wasm-go/extensions/ai-proxy/test/provider_wasm_smoke.go:19`
- **`RunYiWasmSmokeTests`** (Function) — `plugins/wasm-go/extensions/ai-proxy/test/provider_wasm_smoke.go:43`
- **`RunOllamaWasmSmokeTests`** (Function) — `plugins/wasm-go/extensions/ai-proxy/test/provider_wasm_smoke.go:59`
- **`RunBaiduWasmSmokeTests`** (Function) — `plugins/wasm-go/extensions/ai-proxy/test/provider_wasm_smoke.go:77`

## Key Symbols

| Symbol | Type | File | Line |
|--------|------|------|------|
| `TestProviderWasmSmoke` | Function | `plugins/wasm-go/extensions/ai-proxy/main_test.go` | 408 |
| `RunBaichuanWasmSmokeTests` | Function | `plugins/wasm-go/extensions/ai-proxy/test/provider_wasm_smoke.go` | 19 |
| `RunYiWasmSmokeTests` | Function | `plugins/wasm-go/extensions/ai-proxy/test/provider_wasm_smoke.go` | 43 |
| `RunOllamaWasmSmokeTests` | Function | `plugins/wasm-go/extensions/ai-proxy/test/provider_wasm_smoke.go` | 59 |
| `RunBaiduWasmSmokeTests` | Function | `plugins/wasm-go/extensions/ai-proxy/test/provider_wasm_smoke.go` | 77 |
| `RunHunyuanWasmSmokeTests` | Function | `plugins/wasm-go/extensions/ai-proxy/test/provider_wasm_smoke.go` | 93 |
| `RunStepfunWasmSmokeTests` | Function | `plugins/wasm-go/extensions/ai-proxy/test/provider_wasm_smoke.go` | 104 |
| `RunCloudflareWasmSmokeTests` | Function | `plugins/wasm-go/extensions/ai-proxy/test/provider_wasm_smoke.go` | 120 |
| `RunDeeplWasmSmokeTests` | Function | `plugins/wasm-go/extensions/ai-proxy/test/provider_wasm_smoke.go` | 138 |
| `RunCohereWasmSmokeTests` | Function | `plugins/wasm-go/extensions/ai-proxy/test/provider_wasm_smoke.go` | 151 |
| `RunCozeWasmSmokeTests` | Function | `plugins/wasm-go/extensions/ai-proxy/test/provider_wasm_smoke.go` | 167 |
| `RunDifyWasmSmokeTests` | Function | `plugins/wasm-go/extensions/ai-proxy/test/provider_wasm_smoke.go` | 183 |
| `RunTritonWasmSmokeTests` | Function | `plugins/wasm-go/extensions/ai-proxy/test/provider_wasm_smoke.go` | 201 |
| `RunVllmWasmSmokeTests` | Function | `plugins/wasm-go/extensions/ai-proxy/test/provider_wasm_smoke.go` | 220 |
| `TestVertex` | Function | `plugins/wasm-go/extensions/ai-proxy/main_test.go` | 316 |
| `RunVertexParseConfigTests` | Function | `plugins/wasm-go/extensions/ai-proxy/test/vertex.go` | 194 |
| `RunVertexExpressModeOnHttpRequestHeadersTests` | Function | `plugins/wasm-go/extensions/ai-proxy/test/vertex.go` | 278 |
| `RunVertexExpressModeOnHttpRequestBodyTests` | Function | `plugins/wasm-go/extensions/ai-proxy/test/vertex.go` | 342 |
| `RunVertexExpressModeOnHttpResponseBodyTests` | Function | `plugins/wasm-go/extensions/ai-proxy/test/vertex.go` | 923 |
| `RunVertexOpenAICompatibleModeOnHttpRequestHeadersTests` | Function | `plugins/wasm-go/extensions/ai-proxy/test/vertex.go` | 1001 |

## Connected Areas

| Area | Connections |
|------|-------------|
| Provider | 2 calls |

## How to Explore

1. `gitnexus_context({name: "TestProviderWasmSmoke"})` — see callers and callees
2. `gitnexus_query({query: "test"})` — find related execution flows
3. Read key files listed above for implementation details
