---
name: configmap
description: "Skill for the Configmap area of higress. 89 symbols across 10 files."
---

# Configmap

89 symbols | 10 files | Cohesion: 83%

## When to Use

- Working with code in `pkg/`
- Understanding how Test_validGlobal, Test_compareGlobal, Test_deepCopyGlobal work
- Modifying configmap-related functionality

## Key Files

| File | Symbols |
|------|---------|
| `pkg/ingress/kube/configmap/global.go` | validGlobal, compareGlobal, deepCopyGlobal, NewDefaultGlobalOption, NewDefaultDownstream (+20) |
| `pkg/ingress/kube/configmap/tracing.go` | validServiceAndPort, validTracing, compareTracing, deepCopyTracing, NewDefaultTracing (+10) |
| `pkg/ingress/kube/configmap/mcp_server.go` | NewDefaultMcpServer, NewMcpServerController, validMcpServer, compareMcpServer, deepCopyMcpServer (+7) |
| `pkg/ingress/kube/configmap/gzip.go` | validGzip, compareGzip, deepCopyGzip, NewDefaultGzip, NewGzipController (+6) |
| `pkg/ingress/kube/configmap/mcp_server_test.go` | TestMcpServerController_AddOrUpdateHigressConfig, TestMcpServerController_ValidHigressConfig, TestMcpServerController_ConstructEnvoyFilters, TestMcpServerController_constructMcpSessionStruct, TestMcpServerController_constructMcpServerStruct (+3) |
| `pkg/ingress/kube/configmap/controller.go` | NewConfigmapMgr, AddItemControllers, initEventHandlers, SetHigressConfig, GetHigressConfig (+2) |
| `pkg/ingress/kube/configmap/global_test.go` | Test_validGlobal, Test_compareGlobal, Test_deepCopyGlobal, Test_AddOrUpdateHigressConfig |
| `pkg/ingress/kube/configmap/gzip_test.go` | Test_validGzip, Test_compareGzip, Test_deepCopyGzip, TestGzipController_AddOrUpdateHigressConfig |
| `pkg/ingress/kube/configmap/config.go` | NewDefaultHigressConfig, GetHigressConfigString |
| `pkg/ingress/kube/util/util.go` | BuildPatchStruct |

## Entry Points

Start here when exploring this area:

- **`Test_validGlobal`** (Function) — `pkg/ingress/kube/configmap/global_test.go:23`
- **`Test_compareGlobal`** (Function) — `pkg/ingress/kube/configmap/global_test.go:58`
- **`Test_deepCopyGlobal`** (Function) — `pkg/ingress/kube/configmap/global_test.go:117`
- **`Test_AddOrUpdateHigressConfig`** (Function) — `pkg/ingress/kube/configmap/global_test.go:170`
- **`NewDefaultGlobalOption`** (Function) — `pkg/ingress/kube/configmap/global.go:175`

## Key Symbols

| Symbol | Type | File | Line |
|--------|------|------|------|
| `Test_validGlobal` | Function | `pkg/ingress/kube/configmap/global_test.go` | 23 |
| `Test_compareGlobal` | Function | `pkg/ingress/kube/configmap/global_test.go` | 58 |
| `Test_deepCopyGlobal` | Function | `pkg/ingress/kube/configmap/global_test.go` | 117 |
| `Test_AddOrUpdateHigressConfig` | Function | `pkg/ingress/kube/configmap/global_test.go` | 170 |
| `NewDefaultGlobalOption` | Function | `pkg/ingress/kube/configmap/global.go` | 175 |
| `NewDefaultDownstream` | Function | `pkg/ingress/kube/configmap/global.go` | 185 |
| `NewDefaultUpStream` | Function | `pkg/ingress/kube/configmap/global.go` | 196 |
| `NewDefaultHttp2` | Function | `pkg/ingress/kube/configmap/global.go` | 204 |
| `NewGlobalOptionController` | Function | `pkg/ingress/kube/configmap/global.go` | 221 |
| `Test_validGzip` | Function | `pkg/ingress/kube/configmap/gzip_test.go` | 26 |
| `Test_compareGzip` | Function | `pkg/ingress/kube/configmap/gzip_test.go` | 174 |
| `Test_deepCopyGzip` | Function | `pkg/ingress/kube/configmap/gzip_test.go` | 270 |
| `TestGzipController_AddOrUpdateHigressConfig` | Function | `pkg/ingress/kube/configmap/gzip_test.go` | 340 |
| `NewDefaultGzip` | Function | `pkg/ingress/kube/configmap/gzip.go` | 144 |
| `NewGzipController` | Function | `pkg/ingress/kube/configmap/gzip.go` | 166 |
| `TestMcpServerController_AddOrUpdateHigressConfig` | Function | `pkg/ingress/kube/configmap/mcp_server_test.go` | 410 |
| `TestMcpServerController_ValidHigressConfig` | Function | `pkg/ingress/kube/configmap/mcp_server_test.go` | 537 |
| `TestMcpServerController_ConstructEnvoyFilters` | Function | `pkg/ingress/kube/configmap/mcp_server_test.go` | 593 |
| `TestMcpServerController_constructMcpSessionStruct` | Function | `pkg/ingress/kube/configmap/mcp_server_test.go` | 640 |
| `TestMcpServerController_constructMcpServerStruct` | Function | `pkg/ingress/kube/configmap/mcp_server_test.go` | 853 |

## Execution Flows

| Flow | Type | Steps |
|------|------|-------|
| `AddOrUpdateHigressConfig → Http2` | cross_community | 6 |
| `AddOrUpdateHigressConfig → Downstream` | cross_community | 5 |
| `AddOrUpdateHigressConfig → Upstream` | cross_community | 5 |
| `AddOrUpdateHigressConfig → Global` | cross_community | 4 |
| `AddOrUpdateHigressConfig → Tracing` | cross_community | 4 |
| `AddOrUpdateHigressConfig → Gzip` | cross_community | 4 |
| `ConstructEnvoyFilters → BuildPatchStruct` | cross_community | 3 |

## How to Explore

1. `gitnexus_context({name: "Test_validGlobal"})` — see callers and callees
2. `gitnexus_query({query: "configmap"})` — find related execution flows
3. Read key files listed above for implementation details
