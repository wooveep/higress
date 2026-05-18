---
name: v1
description: "Skill for the V1 area of higress. 80 symbols across 17 files."
---

# V1

80 symbols | 17 files | Cohesion: 92%

## When to Use

- Working with code in `client/`
- Understanding how NewController, NewMcpBridgeLister, NewMcpBridgeInformer work
- Modifying v1-related functionality

## Key Files

| File | Symbols |
|------|---------|
| `client/pkg/applyconfiguration/networking/v1/mcpbridge.go` | WithGenerateName, WithUID, WithResourceVersion, WithGeneration, WithCreationTimestamp (+12) |
| `client/pkg/applyconfiguration/networking/v1/http2rpc.go` | WithGenerateName, WithUID, WithResourceVersion, WithGeneration, WithCreationTimestamp (+12) |
| `client/pkg/clientset/versioned/typed/networking/v1/networking_client.gen.go` | Http2Rpcs, McpBridges, RESTClient, NewForConfig, NewForConfigAndClient (+2) |
| `api/networking/v1/http_2_rpc.pb.go` | GetDestination, GetDubbo, GetGrpc, Descriptor, file_networking_v1_http_2_rpc_proto_rawDescGZIP (+2) |
| `client/pkg/informers/externalversions/networking/v1/mcpbridge.gen.go` | NewMcpBridgeInformer, NewFilteredMcpBridgeInformer, defaultInformer, Informer, Lister |
| `client/pkg/informers/externalversions/networking/v1/http2rpc.gen.go` | NewHttp2RpcInformer, NewFilteredHttp2RpcInformer, defaultInformer, Informer, Lister |
| `api/networking/v1/mcp_bridge.pb.go` | Descriptor, file_networking_v1_mcp_bridge_proto_rawDescGZIP, init, file_networking_v1_mcp_bridge_proto_init |
| `api/networking/v1/mcp_bridge_deepcopy.gen.go` | DeepCopyInto, DeepCopy, DeepCopyInterface |
| `api/networking/v1/http_2_rpc_deepcopy.gen.go` | DeepCopyInto, DeepCopy, DeepCopyInterface |
| `client/pkg/apis/networking/v1/zz_generated.deepcopy.gen.go` | DeepCopyInto, DeepCopy, DeepCopyObject |

## Entry Points

Start here when exploring this area:

- **`NewController`** (Function) — `pkg/ingress/kube/mcpbridge/controller.go:34`
- **`NewMcpBridgeLister`** (Function) — `client/pkg/listers/networking/v1/mcpbridge.gen.go:42`
- **`NewMcpBridgeInformer`** (Function) — `client/pkg/informers/externalversions/networking/v1/mcpbridge.gen.go:48`
- **`NewFilteredMcpBridgeInformer`** (Function) — `client/pkg/informers/externalversions/networking/v1/mcpbridge.gen.go:55`
- **`NewController`** (Function) — `pkg/ingress/kube/http2rpc/controller.go:34`

## Key Symbols

| Symbol | Type | File | Line |
|--------|------|------|------|
| `NewController` | Function | `pkg/ingress/kube/mcpbridge/controller.go` | 34 |
| `NewMcpBridgeLister` | Function | `client/pkg/listers/networking/v1/mcpbridge.gen.go` | 42 |
| `NewMcpBridgeInformer` | Function | `client/pkg/informers/externalversions/networking/v1/mcpbridge.gen.go` | 48 |
| `NewFilteredMcpBridgeInformer` | Function | `client/pkg/informers/externalversions/networking/v1/mcpbridge.gen.go` | 55 |
| `NewController` | Function | `pkg/ingress/kube/http2rpc/controller.go` | 34 |
| `NewHttp2RpcLister` | Function | `client/pkg/listers/networking/v1/http2rpc.gen.go` | 42 |
| `NewHttp2RpcInformer` | Function | `client/pkg/informers/externalversions/networking/v1/http2rpc.gen.go` | 48 |
| `NewFilteredHttp2RpcInformer` | Function | `client/pkg/informers/externalversions/networking/v1/http2rpc.gen.go` | 55 |
| `McpBridge` | Function | `client/pkg/applyconfiguration/networking/v1/mcpbridge.go` | 37 |
| `Http2Rpc` | Function | `client/pkg/applyconfiguration/networking/v1/http2rpc.go` | 37 |
| `NewForConfig` | Function | `client/pkg/clientset/versioned/typed/networking/v1/networking_client.gen.go` | 48 |
| `NewForConfigAndClient` | Function | `client/pkg/clientset/versioned/typed/networking/v1/networking_client.gen.go` | 62 |
| `NewForConfigOrDie` | Function | `client/pkg/clientset/versioned/typed/networking/v1/networking_client.gen.go` | 76 |
| `Resource` | Function | `client/pkg/apis/networking/v1/register.gen.go` | 37 |
| `WithGenerateName` | Method | `client/pkg/applyconfiguration/networking/v1/mcpbridge.go` | 74 |
| `WithUID` | Method | `client/pkg/applyconfiguration/networking/v1/mcpbridge.go` | 92 |
| `WithResourceVersion` | Method | `client/pkg/applyconfiguration/networking/v1/mcpbridge.go` | 101 |
| `WithGeneration` | Method | `client/pkg/applyconfiguration/networking/v1/mcpbridge.go` | 110 |
| `WithCreationTimestamp` | Method | `client/pkg/applyconfiguration/networking/v1/mcpbridge.go` | 119 |
| `WithDeletionTimestamp` | Method | `client/pkg/applyconfiguration/networking/v1/mcpbridge.go` | 128 |

## Execution Flows

| Flow | Type | Steps |
|------|------|-------|
| `NewIngressConfig → NewFilteredMcpBridgeInformer` | cross_community | 4 |
| `NewIngressConfig → AddEventHandler` | cross_community | 4 |
| `NewIngressConfig → McpBridgeLister` | cross_community | 4 |
| `McpBridge → ObjectMetaApplyConfiguration` | cross_community | 4 |
| `Http2Rpc → ObjectMetaApplyConfiguration` | cross_community | 4 |

## Connected Areas

| Area | Connections |
|------|-------------|
| V1alpha1 | 2 calls |

## How to Explore

1. `gitnexus_context({name: "NewController"})` — see callers and callees
2. `gitnexus_query({query: "v1"})` — find related execution flows
3. Read key files listed above for implementation details
