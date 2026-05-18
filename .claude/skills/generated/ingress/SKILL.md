---
name: ingress
description: "Skill for the Ingress area of higress. 87 symbols across 19 files."
---

# Ingress

87 symbols | 19 files | Cohesion: 91%

## When to Use

- Working with code in `pkg/`
- Understanding how TestParseTLSSecret, ParseTLSSecret, TestCreateRuleKey work
- Modifying ingress-related functionality

## Key Files

| File | Symbols |
|------|---------|
| `pkg/ingress/kube/ingress/controller.go` | extractTLSSecretName, ConvertGateway, ConvertHTTPRoute, ApplyDefaultBackend, ApplyCanaryIngress (+16) |
| `pkg/ingress/kube/ingressv1/controller.go` | extractTLSSecretName, ConvertGateway, ConvertHTTPRoute, generateHttpMatches, ApplyDefaultBackend (+10) |
| `pkg/ingress/kube/kingress/controller.go` | extractTLSSecretName, ConvertGateway, ConvertHTTPRoute, IngressRouteBuilderServicesCheck, createRuleKey (+3) |
| `pkg/ingress/kube/common/tool.go` | ValidateBackendResource, SortHTTPRoutes, GenerateUniqueRouteName, GenerateUniqueRouteNameWithSuffix, SplitServiceFQDN (+3) |
| `pkg/ingress/kube/common/tool_test.go` | TestGenerateUniqueRouteName, TestSortRoutes, TestSortHTTPRoutesWithMoreRules, TestValidateBackendResource, TestSplitServiceFQDN (+2) |
| `pkg/ingress/kube/common/model.go` | New, NewAndAdd, Add, Update, Delete (+1) |
| `pkg/ingress/kube/kingress/controller_test.go` | TestCreateRuleKey, buildHigressAnnotationKey, TestKingressPathHeadersKey |
| `pkg/ingress/kube/annotations/annotations.go` | IsCanary, CanaryKind, NeedTrafficPolicy |
| `pkg/ingress/kube/ingress/status.go` | run, runUpdateStatus, updateStatus |
| `pkg/ingress/kube/ingress/controller_test.go` | TestCreateRuleKey, buildHigressAnnotationKey |

## Entry Points

Start here when exploring this area:

- **`TestParseTLSSecret`** (Function) — `pkg/cert/config_test.go:123`
- **`ParseTLSSecret`** (Function) — `pkg/cert/config.go:88`
- **`TestCreateRuleKey`** (Function) — `pkg/ingress/kube/kingress/controller_test.go:589`
- **`TestKingressPathHeadersKey`** (Function) — `pkg/ingress/kube/kingress/controller_test.go:624`
- **`TestCreateRuleKey`** (Function) — `pkg/ingress/kube/ingress/controller_test.go:1328`

## Key Symbols

| Symbol | Type | File | Line |
|--------|------|------|------|
| `TestParseTLSSecret` | Function | `pkg/cert/config_test.go` | 123 |
| `ParseTLSSecret` | Function | `pkg/cert/config.go` | 88 |
| `TestCreateRuleKey` | Function | `pkg/ingress/kube/kingress/controller_test.go` | 589 |
| `TestKingressPathHeadersKey` | Function | `pkg/ingress/kube/kingress/controller_test.go` | 624 |
| `TestCreateRuleKey` | Function | `pkg/ingress/kube/ingress/controller_test.go` | 1328 |
| `TestGenerateUniqueRouteName` | Function | `pkg/ingress/kube/common/tool_test.go` | 154 |
| `TestSortRoutes` | Function | `pkg/ingress/kube/common/tool_test.go` | 310 |
| `TestSortHTTPRoutesWithMoreRules` | Function | `pkg/ingress/kube/common/tool_test.go` | 420 |
| `TestValidateBackendResource` | Function | `pkg/ingress/kube/common/tool_test.go` | 559 |
| `TestSplitServiceFQDN` | Function | `pkg/ingress/kube/common/tool_test.go` | 764 |
| `TestConvertBackendService` | Function | `pkg/ingress/kube/common/tool_test.go` | 812 |
| `ValidateBackendResource` | Function | `pkg/ingress/kube/common/tool.go` | 37 |
| `SortHTTPRoutes` | Function | `pkg/ingress/kube/common/tool.go` | 171 |
| `GenerateUniqueRouteName` | Function | `pkg/ingress/kube/common/tool.go` | 315 |
| `GenerateUniqueRouteNameWithSuffix` | Function | `pkg/ingress/kube/common/tool.go` | 322 |
| `SplitServiceFQDN` | Function | `pkg/ingress/kube/common/tool.go` | 326 |
| `ConvertBackendService` | Function | `pkg/ingress/kube/common/tool.go` | 334 |
| `TestIngressDomainBuilder` | Function | `pkg/ingress/kube/common/model_test.go` | 56 |
| `IncrementInvalidIngress` | Function | `pkg/ingress/kube/common/metrics.go` | 68 |
| `CreateMcpServiceKey` | Function | `pkg/ingress/kube/common/controller.go` | 57 |

## Execution Flows

| Flow | Type | Steps |
|------|------|-------|
| `ApplyDefaultBackend → Meta` | cross_community | 5 |
| `ApplyDefaultBackend → Meta` | cross_community | 5 |
| `ApplyDefaultBackend → CreateConvertedName` | cross_community | 4 |
| `ApplyDefaultBackend → CreateConvertedName` | cross_community | 4 |
| `InitConfig → ParseTLSSecret` | cross_community | 4 |
| `ApplyCanaryIngress → IsCanary` | intra_community | 3 |
| `ApplyCanaryIngress → IngressRouteBuilder` | intra_community | 3 |
| `ApplyCanaryIngress → IsCanary` | intra_community | 3 |
| `ApplyCanaryIngress → IngressRouteBuilder` | intra_community | 3 |
| `Run → ServiceKey` | cross_community | 3 |

## Connected Areas

| Area | Connections |
|------|-------------|
| Config | 13 calls |
| Kingress | 2 calls |
| Resources | 1 calls |
| Annotations | 1 calls |
| Ingressv1 | 1 calls |
| Cluster_375 | 1 calls |

## How to Explore

1. `gitnexus_context({name: "TestParseTLSSecret"})` — see callers and callees
2. `gitnexus_query({query: "ingress"})` — find related execution flows
3. Read key files listed above for implementation details
