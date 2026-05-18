---
name: mcpserver
description: "Skill for the Mcpserver area of higress. 65 symbols across 10 files."
---

# Mcpserver

65 symbols | 10 files | Cohesion: 87%

## When to Use

- Working with code in `registry/`
- Understanding how NewWatcher, WithUpdateCacheWhenEmpty, NewWatcher work
- Modifying mcpserver-related functionality

## Key Files

| File | Symbols |
|------|---------|
| `registry/nacos/mcpserver/watcher.go` | NewWatcher, WithNacosNamespaceId, WithNacosNamespace, WithNacosGroups, WithNacosAddressServer (+32) |
| `registry/nacos/mcpserver/client.go` | NewMcpRegistryClient, ListenToMcpServer, onServerVersionChanged, triggerMcpServerChange, mapConfigMapToServerConfig (+10) |
| `registry/nacos/v2/watcher.go` | NewWatcher, WithUpdateCacheWhenEmpty, updateNacosClient |
| `registry/nacos/mcpserver/watcher_test.go` | newTestWatcher, testCallback, Test_Watcher |
| `registry/nacos/mcpserver/client_test.go` | TestNacosRegistryClient_ListenToMcpServer, TestNacosRegistryClient_ListMcpServer |
| `registry/nacos/address/address_discovery.go` | GetNacosAddress |
| `registry/reconcile/reconcile.go` | NewReconciler |
| `registry/memory/cache.go` | NewCache |
| `pkg/ingress/config/ingress_config.go` | AddOrUpdateMcpBridge |
| `pkg/ingress/kube/configmap/controller.go` | RegisterMcpServerProvider |

## Entry Points

Start here when exploring this area:

- **`NewWatcher`** (Function) — `registry/nacos/v2/watcher.go:84`
- **`WithUpdateCacheWhenEmpty`** (Function) — `registry/nacos/v2/watcher.go:278`
- **`NewWatcher`** (Function) — `registry/nacos/mcpserver/watcher.go:101`
- **`WithNacosNamespaceId`** (Function) — `registry/nacos/mcpserver/watcher.go:167`
- **`WithNacosNamespace`** (Function) — `registry/nacos/mcpserver/watcher.go:177`

## Key Symbols

| Symbol | Type | File | Line |
|--------|------|------|------|
| `NewWatcher` | Function | `registry/nacos/v2/watcher.go` | 84 |
| `WithUpdateCacheWhenEmpty` | Function | `registry/nacos/v2/watcher.go` | 278 |
| `NewWatcher` | Function | `registry/nacos/mcpserver/watcher.go` | 101 |
| `WithNacosNamespaceId` | Function | `registry/nacos/mcpserver/watcher.go` | 167 |
| `WithNacosNamespace` | Function | `registry/nacos/mcpserver/watcher.go` | 177 |
| `WithNacosGroups` | Function | `registry/nacos/mcpserver/watcher.go` | 183 |
| `WithNacosAddressServer` | Function | `registry/nacos/mcpserver/watcher.go` | 189 |
| `WithNacosAccessKey` | Function | `registry/nacos/mcpserver/watcher.go` | 195 |
| `WithNacosSecretKey` | Function | `registry/nacos/mcpserver/watcher.go` | 201 |
| `WithNacosRefreshInterval` | Function | `registry/nacos/mcpserver/watcher.go` | 207 |
| `WithType` | Function | `registry/nacos/mcpserver/watcher.go` | 216 |
| `WithName` | Function | `registry/nacos/mcpserver/watcher.go` | 222 |
| `WithDomain` | Function | `registry/nacos/mcpserver/watcher.go` | 228 |
| `WithPort` | Function | `registry/nacos/mcpserver/watcher.go` | 234 |
| `WithMcpExportDomains` | Function | `registry/nacos/mcpserver/watcher.go` | 240 |
| `WithMcpBaseUrl` | Function | `registry/nacos/mcpserver/watcher.go` | 246 |
| `WithEnableMcpServer` | Function | `registry/nacos/mcpserver/watcher.go` | 252 |
| `WithNamespace` | Function | `registry/nacos/mcpserver/watcher.go` | 258 |
| `WithClusterId` | Function | `registry/nacos/mcpserver/watcher.go` | 264 |
| `WithAuthOption` | Function | `registry/nacos/mcpserver/watcher.go` | 270 |

## Execution Flows

| Flow | Type | Steps |
|------|------|-------|
| `NewWatcher → NacosRegistryClient` | intra_community | 4 |
| `Run → ListMcpServerConfigs` | cross_community | 4 |
| `Run → BasicMcpServerInfo` | cross_community | 4 |
| `Run → IsMcpServerShouldBeDiscoveryForGateway` | cross_community | 4 |
| `Run → VersionsMcpServerInfo` | cross_community | 4 |
| `Run → VersionedMcpServerInfo` | cross_community | 4 |
| `NewWatcher → Watcher` | intra_community | 3 |

## Connected Areas

| Area | Connections |
|------|-------------|
| Config | 12 calls |
| Address | 1 calls |
| V2 | 1 calls |
| Direct | 1 calls |

## How to Explore

1. `gitnexus_context({name: "NewWatcher"})` — see callers and callees
2. `gitnexus_query({query: "mcpserver"})` — find related execution flows
3. Read key files listed above for implementation details
