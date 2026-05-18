---
name: tools
description: "Skill for the Tools area of higress. 149 symbols across 35 files."
---

# Tools

149 symbols | 35 files | Cohesion: 97%

## When to Use

- Working with code in `plugins/`
- Understanding how FormatJSONResponse, CreateToolResult, CreateErrorResult work
- Modifying tools-related functionality

## Key Files

| File | Symbols |
|------|---------|
| `plugins/golang-filter/mcp-server/servers/higress/higress-api/tools/mcp_server.go` | RegisterMcpServerTools, handleListMcpServers, handleGetMcpServer, handleAddOrUpdateMcpServer, handleDeleteMcpServer (+10) |
| `plugins/golang-filter/mcp-server/servers/higress/nginx-migration/tools/lua_converter.go` | ConvertLuaToWasm, generateGoCode, generateConfigFields, generateRequestHeadersLogic, generateRequestBodyLogic (+9) |
| `plugins/golang-filter/mcp-server/servers/higress/higress-api/tools/service.go` | RegisterServiceTools, handleListServiceSources, handleGetServiceSource, handleAddServiceSource, handleUpdateServiceSource (+5) |
| `plugins/golang-filter/mcp-server/servers/higress/higress-api/tools/route.go` | RegisterRouteTools, handleListRoutes, handleGetRoute, handleAddRoute, handleUpdateRoute (+5) |
| `plugins/golang-filter/mcp-server/servers/higress/higress-api/tools/ai_route.go` | RegisterAiRouteTools, handleListAiRoutes, handleGetAiRoute, handleAddAiRoute, handleUpdateAiRoute (+5) |
| `plugins/golang-filter/mcp-server/servers/higress/higress-api/tools/ai_provider.go` | RegisterAiProviderTools, handleListAiProviders, handleGetAiProvider, handleAddAiProvider, handleUpdateAiProvider (+5) |
| `plugins/golang-filter/mcp-server/servers/higress/higress-ops/tools/envoy.go` | RegisterEnvoyTools, handleEnvoyConfigDump, handleEnvoyClusters, handleEnvoyListeners, handleEnvoyStats (+4) |
| `plugins/golang-filter/mcp-server/servers/higress/nginx-migration/tools/tool_chain.go` | AnalyzeLuaPluginForAI, ValidateWasmCode, removeComments, containsImport, checkCallbackReturnErrors (+4) |
| `plugins/golang-filter/mcp-server/servers/higress/higress-ops/tools/istiod.go` | RegisterIstiodTools, handleIstiodSyncz, handleIstiodEndpointz, handleIstiodConfigz, handleIstiodClusters (+2) |
| `plugins/golang-filter/mcp-server/servers/higress/higress-ops/tools/utils.go` | FormatJSONResponse, CreateToolResult, CreateErrorResult, GetStringParam, CreateSimpleSchema (+1) |

## Entry Points

Start here when exploring this area:

- **`FormatJSONResponse`** (Function) — `plugins/golang-filter/mcp-server/servers/higress/higress-ops/tools/utils.go:11`
- **`CreateToolResult`** (Function) — `plugins/golang-filter/mcp-server/servers/higress/higress-ops/tools/utils.go:27`
- **`CreateErrorResult`** (Function) — `plugins/golang-filter/mcp-server/servers/higress/higress-ops/tools/utils.go:51`
- **`GetStringParam`** (Function) — `plugins/golang-filter/mcp-server/servers/higress/higress-ops/tools/utils.go:56`
- **`CreateSimpleSchema`** (Function) — `plugins/golang-filter/mcp-server/servers/higress/higress-ops/tools/utils.go:82`

## Key Symbols

| Symbol | Type | File | Line |
|--------|------|------|------|
| `FormatJSONResponse` | Function | `plugins/golang-filter/mcp-server/servers/higress/higress-ops/tools/utils.go` | 11 |
| `CreateToolResult` | Function | `plugins/golang-filter/mcp-server/servers/higress/higress-ops/tools/utils.go` | 27 |
| `CreateErrorResult` | Function | `plugins/golang-filter/mcp-server/servers/higress/higress-ops/tools/utils.go` | 51 |
| `GetStringParam` | Function | `plugins/golang-filter/mcp-server/servers/higress/higress-ops/tools/utils.go` | 56 |
| `CreateSimpleSchema` | Function | `plugins/golang-filter/mcp-server/servers/higress/higress-ops/tools/utils.go` | 82 |
| `CreateParameterSchema` | Function | `plugins/golang-filter/mcp-server/servers/higress/higress-ops/tools/utils.go` | 92 |
| `RegisterIstiodTools` | Function | `plugins/golang-filter/mcp-server/servers/higress/higress-ops/tools/istiod.go` | 10 |
| `RegisterEnvoyTools` | Function | `plugins/golang-filter/mcp-server/servers/higress/higress-ops/tools/envoy.go` | 10 |
| `OnMCPToolCallError` | Function | `plugins/wasm-go/pkg/mcp/utils/mcp_rpc.go` | 54 |
| `SendMCPToolTextResult` | Function | `plugins/wasm-go/pkg/mcp/utils/mcp_rpc.go` | 70 |
| `ToInputSchema` | Function | `plugins/wasm-go/pkg/mcp/server/plugin.go` | 712 |
| `TestWebSearchInputSchema` | Function | `plugins/wasm-go/mcp-servers/quark-search/tools/web_search_test.go` | 24 |
| `RegisterMcpServerTools` | Function | `plugins/golang-filter/mcp-server/servers/higress/higress-api/tools/mcp_server.go` | 68 |
| `RegisterServiceTools` | Function | `plugins/golang-filter/mcp-server/servers/higress/higress-api/tools/service.go` | 37 |
| `RegisterRouteTools` | Function | `plugins/golang-filter/mcp-server/servers/higress/higress-api/tools/route.go` | 57 |
| `RegisterAiRouteTools` | Function | `plugins/golang-filter/mcp-server/servers/higress/higress-api/tools/ai_route.go` | 67 |
| `RegisterAiProviderTools` | Function | `plugins/golang-filter/mcp-server/servers/higress/higress-api/tools/ai_provider.go` | 36 |
| `ConvertLuaToWasm` | Function | `plugins/golang-filter/mcp-server/servers/higress/nginx-migration/tools/lua_converter.go` | 180 |
| `NewMCPServer` | Function | `plugins/wasm-go/pkg/mcp/mcp.go` | 15 |
| `LoadTools` | Function | `plugins/wasm-go/mcp-servers/quark-search/tools/load_tools.go` | 21 |

## Execution Flows

| Flow | Type | Steps |
|------|------|-------|
| `Call → MakeHttpResponse` | cross_community | 7 |
| `NewServer → FormatJSONResponse` | cross_community | 5 |
| `NewServer → CreateErrorResult` | cross_community | 4 |
| `Init → AddMCPTool` | cross_community | 4 |
| `Init → BaseMCPServer` | cross_community | 4 |
| `RegisterEnvoyTools → FormatJSONResponse` | intra_community | 4 |
| `Init → AddMCPTool` | cross_community | 4 |
| `Init → BaseMCPServer` | cross_community | 4 |
| `Init → AddMCPTool` | cross_community | 4 |
| `Init → BaseMCPServer` | cross_community | 4 |

## Connected Areas

| Area | Connections |
|------|-------------|
| Server | 4 calls |

## How to Explore

1. `gitnexus_context({name: "FormatJSONResponse"})` — see callers and callees
2. `gitnexus_query({query: "tools"})` — find related execution flows
3. Read key files listed above for implementation details
