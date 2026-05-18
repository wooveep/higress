---
name: server
description: "Skill for the Server area of higress. 168 symbols across 18 files."
---

# Server

168 symbols | 18 files | Cohesion: 85%

## When to Use

- Working with code in `plugins/`
- Understanding how TestMcpProxyServerTransport, TestToolsListForwarding, TestToolsCallForwarding work
- Modifying server-related functionality

## Key Files

| File | Symbols |
|------|---------|
| `plugins/wasm-go/pkg/mcp/server/proxy_tool.go` | ensureHeader, parseSSEResponse, Initialize, ForwardToolsList, executeToolsList (+19) |
| `plugins/wasm-go/pkg/mcp/server/rest_server.go` | GetSecurityScheme, GetDefaultDownstreamSecurity, GetDefaultUpstreamSecurity, GetPassthroughAuthHeader, convertArgToString (+15) |
| `plugins/wasm-go/pkg/mcp/server/proxy_server.go` | NewMcpProxyServer, GetTimeout, GetTransport, GetToolConfig, GetSecurityScheme (+14) |
| `plugins/wasm-go/pkg/mcp/server/plugin.go` | validateURL, setupMcpProxyServer, onHttpStreamingResponseBody, RegisterTool, computeEffectiveAllowTools (+11) |
| `plugins/wasm-go/pkg/mcp/server/proxy_integration_test.go` | TestMcpProtocolInitialization, TestMcpInitializeRequest, CreateInitializeRequest, TestMcpSessionManagement, NewMcpSessionManager (+10) |
| `plugins/wasm-go/pkg/mcp/server/sse_proxy.go` | injectSSEResponseSuccess, injectSSEResponseError, ParseSSEMessage, ExtractEndpointURL, sendSSEInitialize (+9) |
| `plugins/wasm-go/pkg/mcp/server/proxy_auth_test.go` | TestApiKeyAuthentication, TestBearerAuthentication, TestBasicAuthentication, TestMultipleSecuritySchemes, TestToolsListAuthentication (+3) |
| `plugins/wasm-go/pkg/mcp/server/rest_server_test.go` | TestConvertArgToString, TestHasContentType, TestArgsToUrlParamAndFormBody, TestResponseTemplatePrependAppend, TestInputSchemaWithComplexTypes (+3) |
| `plugins/wasm-go/pkg/mcp/server/proxy_tools_test.go` | TestToolsListForwarding, TestToolsCallForwarding, TestToolsCallWithParameters, TestToolsCallWithCursor, TestBackendErrorHandling (+2) |
| `plugins/wasm-go/pkg/mcp/utils/mcp_rpc.go` | OnMCPResponseSuccess, OnMCPToolCallSuccess, OnMCPToolCallSuccessWithStructuredContent, SendMCPToolImageResult, SendMCPToolTextResultWithStructuredContent (+1) |

## Entry Points

Start here when exploring this area:

- **`TestMcpProxyServerTransport`** (Function) — `plugins/wasm-go/pkg/mcp/server/sse_proxy_test.go:226`
- **`TestToolsListForwarding`** (Function) — `plugins/wasm-go/pkg/mcp/server/proxy_tools_test.go:25`
- **`TestToolsCallForwarding`** (Function) — `plugins/wasm-go/pkg/mcp/server/proxy_tools_test.go:75`
- **`TestToolsCallWithParameters`** (Function) — `plugins/wasm-go/pkg/mcp/server/proxy_tools_test.go:120`
- **`TestToolsCallWithCursor`** (Function) — `plugins/wasm-go/pkg/mcp/server/proxy_tools_test.go:237`

## Key Symbols

| Symbol | Type | File | Line |
|--------|------|------|------|
| `TestMcpProxyServerTransport` | Function | `plugins/wasm-go/pkg/mcp/server/sse_proxy_test.go` | 226 |
| `TestToolsListForwarding` | Function | `plugins/wasm-go/pkg/mcp/server/proxy_tools_test.go` | 25 |
| `TestToolsCallForwarding` | Function | `plugins/wasm-go/pkg/mcp/server/proxy_tools_test.go` | 75 |
| `TestToolsCallWithParameters` | Function | `plugins/wasm-go/pkg/mcp/server/proxy_tools_test.go` | 120 |
| `TestToolsCallWithCursor` | Function | `plugins/wasm-go/pkg/mcp/server/proxy_tools_test.go` | 237 |
| `TestBackendErrorHandling` | Function | `plugins/wasm-go/pkg/mcp/server/proxy_tools_test.go` | 252 |
| `TestMcpProxyServerBasicInterface` | Function | `plugins/wasm-go/pkg/mcp/server/proxy_server_test.go` | 23 |
| `TestMcpProxyServerConfiguration` | Function | `plugins/wasm-go/pkg/mcp/server/proxy_server_test.go` | 42 |
| `TestMcpProxyServerAddTool` | Function | `plugins/wasm-go/pkg/mcp/server/proxy_server_test.go` | 70 |
| `TestMcpProxyServerSecuritySchemes` | Function | `plugins/wasm-go/pkg/mcp/server/proxy_server_test.go` | 95 |
| `NewMcpProxyServer` | Function | `plugins/wasm-go/pkg/mcp/server/proxy_server.go` | 81 |
| `TestMcpProtocolInitialization` | Function | `plugins/wasm-go/pkg/mcp/server/proxy_integration_test.go` | 33 |
| `TestMcpInitializeRequest` | Function | `plugins/wasm-go/pkg/mcp/server/proxy_integration_test.go` | 157 |
| `CreateInitializeRequest` | Function | `plugins/wasm-go/pkg/mcp/server/proxy_integration_test.go` | 291 |
| `TestApiKeyAuthentication` | Function | `plugins/wasm-go/pkg/mcp/server/proxy_auth_test.go` | 25 |
| `TestBearerAuthentication` | Function | `plugins/wasm-go/pkg/mcp/server/proxy_auth_test.go` | 95 |
| `TestBasicAuthentication` | Function | `plugins/wasm-go/pkg/mcp/server/proxy_auth_test.go` | 154 |
| `TestMultipleSecuritySchemes` | Function | `plugins/wasm-go/pkg/mcp/server/proxy_auth_test.go` | 220 |
| `TestToolsListAuthentication` | Function | `plugins/wasm-go/pkg/mcp/server/proxy_auth_test.go` | 255 |
| `TestDefaultSecurityFallback` | Function | `plugins/wasm-go/pkg/mcp/server/proxy_auth_test.go` | 300 |

## Execution Flows

| Flow | Type | Steps |
|------|------|-------|
| `Call → MakeHttpResponse` | cross_community | 7 |
| `Call → MakeHttpResponse` | cross_community | 7 |
| `CreateMcpProxyMethodHandlers → MakeHttpResponse` | cross_community | 5 |
| `Init → AddMCPTool` | cross_community | 4 |
| `Init → BaseMCPServer` | cross_community | 4 |
| `Init → AddMCPTool` | cross_community | 4 |
| `Init → BaseMCPServer` | cross_community | 4 |
| `Init → AddMCPTool` | cross_community | 4 |
| `Init → BaseMCPServer` | cross_community | 4 |
| `Call → GetConfig` | cross_community | 3 |

## Connected Areas

| Area | Connections |
|------|-------------|
| Tools | 8 calls |
| Cluster_476 | 2 calls |

## How to Explore

1. `gitnexus_context({name: "TestMcpProxyServerTransport"})` — see callers and callees
2. `gitnexus_query({query: "server"})` — find related execution flows
3. Read key files listed above for implementation details
