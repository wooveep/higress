---
name: v2
description: "Skill for the V2 area of higress. 72 symbols across 11 files."
---

# V2

72 symbols | 11 files | Cohesion: 94%

## When to Use

- Working with code in `registry/`
- Understanding how WithZkServicesPath, WithType, WithName work
- Modifying v2-related functionality

## Key Files

| File | Symbols |
|------|---------|
| `registry/nacos/v2/watcher.go` | WithVport, WithNacosAddressServer, WithNacosAccessKey, WithNacosSecretKey, WithNacosNamespaceId (+21) |
| `registry/nacos/watcher.go` | NewWatcher, WithVport, WithNacosNamespaceId, WithNacosNamespace, WithNacosGroups (+7) |
| `registry/consul/watcher.go` | WithType, WithName, WithDomain, WithPort, WithDatacenter (+4) |
| `registry/direct/watcher.go` | WithType, WithName, WithDomain, WithPort, WithProtocol (+2) |
| `registry/eureka/watcher.go` | NewWatcher, WithVport, WithType, WithName, WithDomain (+1) |
| `registry/zookeeper/types.go` | WithType, WithName, WithDomain, WithPort |
| `registry/reconcile/reconcile.go` | generateWatcherFromRegistryConfig, getAuthOption |
| `registry/eureka/client/http_client.go` | NewEurekaHttpClient, NewDefaultConfig |
| `registry/nacos/address/address_discovery.go` | Trigger, Stop |
| `registry/zookeeper/watcher.go` | WithZkServicesPath |

## Entry Points

Start here when exploring this area:

- **`WithZkServicesPath`** (Function) — `registry/zookeeper/watcher.go:118`
- **`WithType`** (Function) — `registry/zookeeper/types.go:77`
- **`WithName`** (Function) — `registry/zookeeper/types.go:83`
- **`WithDomain`** (Function) — `registry/zookeeper/types.go:89`
- **`WithPort`** (Function) — `registry/zookeeper/types.go:95`

## Key Symbols

| Symbol | Type | File | Line |
|--------|------|------|------|
| `WithZkServicesPath` | Function | `registry/zookeeper/watcher.go` | 118 |
| `WithType` | Function | `registry/zookeeper/types.go` | 77 |
| `WithName` | Function | `registry/zookeeper/types.go` | 83 |
| `WithDomain` | Function | `registry/zookeeper/types.go` | 89 |
| `WithPort` | Function | `registry/zookeeper/types.go` | 95 |
| `NewWatcher` | Function | `registry/nacos/watcher.go` | 70 |
| `WithVport` | Function | `registry/nacos/watcher.go` | 122 |
| `WithNacosNamespaceId` | Function | `registry/nacos/watcher.go` | 128 |
| `WithNacosNamespace` | Function | `registry/nacos/watcher.go` | 138 |
| `WithNacosGroups` | Function | `registry/nacos/watcher.go` | 144 |
| `WithNacosRefreshInterval` | Function | `registry/nacos/watcher.go` | 150 |
| `WithType` | Function | `registry/nacos/watcher.go` | 159 |
| `WithName` | Function | `registry/nacos/watcher.go` | 165 |
| `WithDomain` | Function | `registry/nacos/watcher.go` | 171 |
| `WithPort` | Function | `registry/nacos/watcher.go` | 177 |
| `WithUpdateCacheWhenEmpty` | Function | `registry/nacos/watcher.go` | 183 |
| `WithAuthOption` | Function | `registry/nacos/watcher.go` | 189 |
| `NewWatcher` | Function | `registry/eureka/watcher.go` | 60 |
| `WithVport` | Function | `registry/eureka/watcher.go` | 83 |
| `WithType` | Function | `registry/eureka/watcher.go` | 98 |

## Execution Flows

| Flow | Type | Steps |
|------|------|-------|
| `Run → ParseProtocol` | cross_community | 6 |
| `Run → IsValidIP` | cross_community | 6 |
| `Run → ServiceWrapper` | intra_community | 5 |
| `GenerateWatcherFromRegistryConfig → AuthOption` | intra_community | 3 |
| `GenerateWatcherFromRegistryConfig → Watcher` | intra_community | 3 |
| `GenerateWatcherFromRegistryConfig → WithUpdateCacheWhenEmpty` | intra_community | 3 |
| `Run → Trigger` | intra_community | 3 |
| `Run → ShouldSubscribe` | intra_community | 3 |

## Connected Areas

| Area | Connections |
|------|-------------|
| Mcpserver | 1 calls |
| Zookeeper | 1 calls |
| Direct | 1 calls |
| Registry | 1 calls |

## How to Explore

1. `gitnexus_context({name: "WithZkServicesPath"})` — see callers and callees
2. `gitnexus_query({query: "v2"})` — find related execution flows
3. Read key files listed above for implementation details
