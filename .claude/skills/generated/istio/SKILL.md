---
name: istio
description: "Skill for the Istio area of higress. 213 symbols across 28 files."
---

# Istio

213 symbols | 28 files | Cohesion: 74%

## When to Use

- Working with code in `pkg/`
- Understanding how HTTPRouteCollection, GRPCRouteCollection, TCPRouteCollection work
- Modifying istio-related functionality

## Key Files

| File | Symbols |
|------|---------|
| `pkg/ingress/kube/gateway/istio/conversion.go` | sortRoutesByCreationTime, parentTypes, augmentPortMatch, augmentTCPPortMatch, augmentTLSPortMatch (+81) |
| `pkg/ingress/kube/gateway/istio/inferencepool_collection.go` | translateShadowServiceToService, reconcileShadowService, canManageShadowServiceForInference, findGatewayParents, routeReferencesInferencePool (+13) |
| `pkg/ingress/kube/gateway/istio/inferencepool_status_test.go` | TestInferencePoolStatusReconciliation, assertConditionContains, WithRouteParentCondition, WithParentStatus, AsDefaultStatus (+11) |
| `pkg/ingress/kube/gateway/istio/route_collections.go` | HTTPRouteCollection, extractAncestorBackends, GRPCRouteCollection, TCPRouteCollection, TLSRouteCollection (+6) |
| `pkg/ingress/kube/gateway/istio/conversion_test.go` | setupClientCRDs, Statuses, Dump, TestConvertResources, dumpOnFailure (+5) |
| `pkg/ingress/kube/gateway/istio/backend_policies.go` | String, ResourceName, DestinationRuleCollection, BackendTLSPolicyCollection, getBackendTLSCredentialName (+5) |
| `pkg/ingress/kube/gateway/istio/gateway_collection.go` | fetch, FinalGatewayStatusCollection, BuildRouteParents, GatewayCollection, ListenerSetCollection (+3) |
| `pkg/ingress/kube/gateway/istio/controller.go` | Run, NewController, buildClient, pushXds, SecretAllowed (+1) |
| `pkg/ingress/kube/gateway/istio/references.go` | LocalPolicyTargetRef, XLocalPolicyTargetRef, LocalPolicyRef, internal, NewReferenceSet (+1) |
| `pkg/ingress/kube/gateway/istio/context.go` | ResolveGatewayInstances, GetService, GetEndpoints, checkServicePortExists, extractServiceName (+1) |

## Entry Points

Start here when exploring this area:

- **`HTTPRouteCollection`** (Function) — `pkg/ingress/kube/gateway/istio/route_collections.go:56`
- **`GRPCRouteCollection`** (Function) — `pkg/ingress/kube/gateway/istio/route_collections.go:258`
- **`TCPRouteCollection`** (Function) — `pkg/ingress/kube/gateway/istio/route_collections.go:402`
- **`TLSRouteCollection`** (Function) — `pkg/ingress/kube/gateway/istio/route_collections.go:493`
- **`ReferenceGrantsCollection`** (Function) — `pkg/ingress/kube/gateway/istio/references_collection.go:52`

## Key Symbols

| Symbol | Type | File | Line |
|--------|------|------|------|
| `HTTPRouteCollection` | Function | `pkg/ingress/kube/gateway/istio/route_collections.go` | 56 |
| `GRPCRouteCollection` | Function | `pkg/ingress/kube/gateway/istio/route_collections.go` | 258 |
| `TCPRouteCollection` | Function | `pkg/ingress/kube/gateway/istio/route_collections.go` | 402 |
| `TLSRouteCollection` | Function | `pkg/ingress/kube/gateway/istio/route_collections.go` | 493 |
| `ReferenceGrantsCollection` | Function | `pkg/ingress/kube/gateway/istio/references_collection.go` | 52 |
| `GetCommonRouteInfo` | Function | `pkg/ingress/kube/gateway/istio/conversion.go` | 2533 |
| `GetCommonRouteStateParents` | Function | `pkg/ingress/kube/gateway/istio/conversion.go` | 2549 |
| `GetBackendRef` | Function | `pkg/ingress/kube/gateway/istio/conversion.go` | 2634 |
| `TestCreateRouteStatus` | Function | `pkg/ingress/kube/gateway/istio/conditions_test.go` | 28 |
| `FilterInPlaceByIndex` | Function | `pkg/ingress/kube/gateway/istio/conditions.go` | 412 |
| `NewFakeClient` | Function | `pkg/kube/client.go` | 74 |
| `TestStartWithNoError` | Function | `pkg/bootstrap/server_test.go` | 29 |
| `TestController` | Function | `pkg/ingress/kube/secret/controller_test.go` | 45 |
| `TestStatusCollections` | Function | `pkg/ingress/kube/gateway/istio/status_test.go` | 33 |
| `TestClassController` | Function | `pkg/ingress/kube/gateway/istio/gatewayclass_test.go` | 31 |
| `NewClassController` | Function | `pkg/ingress/kube/gateway/istio/gatewayclass.go` | 42 |
| `TestListInvalidGroupVersionKind` | Function | `pkg/ingress/kube/gateway/istio/controller_test.go` | 96 |
| `NewSimpleClientset` | Function | `client/pkg/clientset/versioned/fake/clientset_generated.gen.go` | 35 |
| `DestinationRuleCollection` | Function | `pkg/ingress/kube/gateway/istio/backend_policies.go` | 124 |
| `BackendTLSPolicyCollection` | Function | `pkg/ingress/kube/gateway/istio/backend_policies.go` | 282 |

## Execution Flows

| Flow | Type | Steps |
|------|------|-------|
| `ListenerSetCollection → Condition` | cross_community | 4 |
| `HTTPRouteCollection → RouteContext` | intra_community | 4 |
| `HTTPRouteCollection → GetCommonRouteInfo` | intra_community | 4 |
| `HTTPRouteCollection → ParentReference` | intra_community | 4 |
| `HTTPRouteCollection → NormalizeReference` | intra_community | 4 |
| `GRPCRouteCollection → RouteContext` | intra_community | 4 |
| `GRPCRouteCollection → GetCommonRouteInfo` | intra_community | 4 |
| `GRPCRouteCollection → ParentReference` | intra_community | 4 |
| `GRPCRouteCollection → NormalizeReference` | intra_community | 4 |
| `GatewayCollection → Condition` | cross_community | 4 |

## Connected Areas

| Area | Connections |
|------|-------------|
| Externalversions | 2 calls |
| Provider | 1 calls |

## How to Explore

1. `gitnexus_context({name: "HTTPRouteCollection"})` — see callers and callees
2. `gitnexus_query({query: "istio"})` — find related execution flows
3. Read key files listed above for implementation details
