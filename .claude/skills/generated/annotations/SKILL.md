---
name: annotations
description: "Skill for the Annotations area of higress. 121 symbols across 38 files."
---

# Annotations

121 symbols | 38 files | Cohesion: 87%

## When to Use

- Working with code in `pkg/`
- Understanding how TestIPAccessControlParse, TestSplitStringWithSpaceTrim, TestCorsParse work
- Modifying annotations-related functionality

## Key Files

| File | Symbols |
|------|---------|
| `pkg/ingress/kube/annotations/parser.go` | ParseBool, ParseBoolForHigress, ParseBoolASAP, ParseIntForHigress, ParseUint32ForHigress (+15) |
| `pkg/ingress/kube/annotations/match.go` | Parse, matchByMethod, matchByHeaderOrQueryParma, doMatch, isMethod (+5) |
| `pkg/ingress/kube/annotations/upstreamtls.go` | ApplyTrafficPolicy, processMTLS, processSimple, isH2, isHTTPS (+3) |
| `pkg/ingress/kube/annotations/cors.go` | Parse, ApplyRoute, needCorsConfig, splitStringWithSpaceTrim |
| `pkg/ingress/kube/annotations/downstreamtls.go` | Parse, needDownstreamTLS, ApplyGateway, convertTLSVersion |
| `pkg/ingress/kube/annotations/redirect.go` | Parse, ApplyRoute, needRedirectConfig, isValidURL |
| `pkg/ingress/kube/annotations/loadbalance.go` | Parse, isCookieAffinity, isOtherAffinity, needLoadBalanceConfig |
| `pkg/ingress/kube/annotations/header_control.go` | Parse, needHeaderControlConfig, trimQuotes, convertAddOrUpdate |
| `pkg/ingress/kube/annotations/ip_access_control.go` | Parse, needIPAccessControlConfig, ApplyRoute |
| `pkg/ingress/kube/annotations/downstreamtls_test.go` | TestParse, TestNeedDownstreamTLS, TestConvertTLSVersion |

## Entry Points

Start here when exploring this area:

- **`TestIPAccessControlParse`** (Function) — `pkg/ingress/kube/annotations/ip_access_control_test.go:23`
- **`TestSplitStringWithSpaceTrim`** (Function) — `pkg/ingress/kube/annotations/cors_test.go:25`
- **`TestCorsParse`** (Function) — `pkg/ingress/kube/annotations/cors_test.go:54`
- **`TestParseMirror`** (Function) — `pkg/ingress/kube/annotations/mirror_test.go:26`
- **`TestMCPServer_Parse`** (Function) — `pkg/ingress/kube/annotations/mcpserver_test.go:24`

## Key Symbols

| Symbol | Type | File | Line |
|--------|------|------|------|
| `TestIPAccessControlParse` | Function | `pkg/ingress/kube/annotations/ip_access_control_test.go` | 23 |
| `TestSplitStringWithSpaceTrim` | Function | `pkg/ingress/kube/annotations/cors_test.go` | 25 |
| `TestCorsParse` | Function | `pkg/ingress/kube/annotations/cors_test.go` | 54 |
| `TestParseMirror` | Function | `pkg/ingress/kube/annotations/mirror_test.go` | 26 |
| `TestMCPServer_Parse` | Function | `pkg/ingress/kube/annotations/mcpserver_test.go` | 24 |
| `TestParse` | Function | `pkg/ingress/kube/annotations/downstreamtls_test.go` | 26 |
| `TestFallbackParse` | Function | `pkg/ingress/kube/annotations/default_backend_test.go` | 55 |
| `TestApplyTrafficPolicy` | Function | `pkg/ingress/kube/annotations/upstreamtls_test.go` | 90 |
| `TestSplitNamespacedName` | Function | `pkg/ingress/kube/util/util_test.go` | 41 |
| `SplitNamespacedName` | Function | `pkg/ingress/kube/util/util.go` | 64 |
| `TestNeedDownstreamTLS` | Function | `pkg/ingress/kube/annotations/downstreamtls_test.go` | 284 |
| `ParseServiceInfo` | Function | `pkg/ingress/kube/util/util.go` | 133 |
| `TestExtraSecret` | Function | `pkg/ingress/kube/annotations/util_test.go` | 24 |
| `TestConvertTLSVersion` | Function | `pkg/ingress/kube/annotations/downstreamtls_test.go` | 96 |
| `TestRewriteParse` | Function | `pkg/ingress/kube/annotations/rewrite_test.go` | 60 |
| `TestRetryParse` | Function | `pkg/ingress/kube/annotations/retry_test.go` | 25 |
| `TestLoadBalanceParse` | Function | `pkg/ingress/kube/annotations/loadbalance_test.go` | 24 |
| `TestCanaryParse` | Function | `pkg/ingress/kube/annotations/canary_test.go` | 25 |
| `IsMissingAnnotations` | Function | `pkg/ingress/kube/annotations/parser.go` | 46 |
| `TestSplitBySeparator` | Function | `pkg/ingress/kube/annotations/util_test.go` | 55 |

## Execution Flows

| Flow | Type | Steps |
|------|------|-------|
| `ApplyDefaultBackend → Meta` | cross_community | 5 |
| `ApplyDefaultBackend → Meta` | cross_community | 5 |

## How to Explore

1. `gitnexus_context({name: "TestIPAccessControlParse"})` — see callers and callees
2. `gitnexus_query({query: "annotations"})` — find related execution flows
3. Read key files listed above for implementation details
