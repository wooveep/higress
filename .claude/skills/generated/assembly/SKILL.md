---
name: assembly
description: "Skill for the Assembly area of higress. 86 symbols across 9 files."
---

# Assembly

86 symbols | 9 files | Cohesion: 94%

## When to Use

- Working with code in `plugins/`
- Understanding how ParseConfigBy, ProcessRequestHeadersBy, ProcessRequestBodyBy work
- Modifying assembly-related functionality

## Key Files

| File | Symbols |
|------|---------|
| `plugins/wasm-assemblyscript/assembly/plugin_wrapper.ts` | Closure, setParseConfigFunc, setHttpHeadersFunc, setHttpBodyFunc, ParseConfigBy (+24) |
| `plugins/wasm-assemblyscript/assembly/rule_matcher.ts` | HostMatcher, RuleConfig, parseRuleConfig, parseRouteMatchConfig, parseRoutePrefixMatchConfig (+9) |
| `plugins/wasm-assemblyscript/assembly/cluster_wrapper.ts` | clusterName, hostName, Cluster, RouteCluster, K8sCluster (+6) |
| `plugins/wasm-assemblyscript/assembly/http_wrapper.ts` | httpCall, get, head, options, post (+5) |
| `plugins/wasm-assemblyscript/assembly/log_wrapper.ts` | Debug, log, Trace, Info, Warn (+3) |
| `plugins/wasm-assemblyscript/extensions/hello-world/assembly/index.ts` | HelloWorldConfig, parseConfig, onHttpResponseHeaders, TestContext, onHttpRequestHeaders |
| `plugins/wasm-assemblyscript/assembly/request_wrapper.ts` | getRequestHost, isBinaryRequestBody, getRequestScheme, getRequestPath, getRequestMethod |
| `plugins/wasm-assemblyscript/extensions/custom-response/assembly/index.ts` | onHttpResponseHeaders, CustomResponseConfig, parseConfig |
| `plugins/wasm-cpp/extensions/oauth/plugin_test.cc` | OAuthTest |

## Entry Points

Start here when exploring this area:

- **`ParseConfigBy`** (Function) — `plugins/wasm-assemblyscript/assembly/plugin_wrapper.ts:190`
- **`ProcessRequestHeadersBy`** (Function) — `plugins/wasm-assemblyscript/assembly/plugin_wrapper.ts:207`
- **`ProcessRequestBodyBy`** (Function) — `plugins/wasm-assemblyscript/assembly/plugin_wrapper.ts:224`
- **`ProcessResponseHeadersBy`** (Function) — `plugins/wasm-assemblyscript/assembly/plugin_wrapper.ts:241`
- **`ProcessResponseBodyBy`** (Function) — `plugins/wasm-assemblyscript/assembly/plugin_wrapper.ts:258`

## Key Symbols

| Symbol | Type | File | Line |
|--------|------|------|------|
| `HelloWorldConfig` | Class | `plugins/wasm-assemblyscript/extensions/hello-world/assembly/index.ts` | 4 |
| `Cluster` | Class | `plugins/wasm-assemblyscript/assembly/cluster_wrapper.ts` | 8 |
| `RouteCluster` | Class | `plugins/wasm-assemblyscript/assembly/cluster_wrapper.ts` | 13 |
| `K8sCluster` | Class | `plugins/wasm-assemblyscript/assembly/cluster_wrapper.ts` | 37 |
| `NacosCluster` | Class | `plugins/wasm-assemblyscript/assembly/cluster_wrapper.ts` | 72 |
| `StaticIpCluster` | Class | `plugins/wasm-assemblyscript/assembly/cluster_wrapper.ts` | 115 |
| `DnsCluster` | Class | `plugins/wasm-assemblyscript/assembly/cluster_wrapper.ts` | 139 |
| `ConsulCluster` | Class | `plugins/wasm-assemblyscript/assembly/cluster_wrapper.ts` | 160 |
| `FQDNCluster` | Class | `plugins/wasm-assemblyscript/assembly/cluster_wrapper.ts` | 191 |
| `ParseResult` | Class | `plugins/wasm-assemblyscript/assembly/rule_matcher.ts` | 52 |
| `RuleMatcher` | Class | `plugins/wasm-assemblyscript/assembly/rule_matcher.ts` | 61 |
| `Log` | Class | `plugins/wasm-assemblyscript/assembly/log_wrapper.ts` | 11 |
| `CustomResponseConfig` | Class | `plugins/wasm-assemblyscript/extensions/custom-response/assembly/index.ts` | 5 |
| `TestContext` | Class | `plugins/wasm-assemblyscript/extensions/hello-world/assembly/index.ts` | 23 |
| `ParseConfigBy` | Function | `plugins/wasm-assemblyscript/assembly/plugin_wrapper.ts` | 190 |
| `ProcessRequestHeadersBy` | Function | `plugins/wasm-assemblyscript/assembly/plugin_wrapper.ts` | 207 |
| `ProcessRequestBodyBy` | Function | `plugins/wasm-assemblyscript/assembly/plugin_wrapper.ts` | 224 |
| `ProcessResponseHeadersBy` | Function | `plugins/wasm-assemblyscript/assembly/plugin_wrapper.ts` | 241 |
| `ProcessResponseBodyBy` | Function | `plugins/wasm-assemblyscript/assembly/plugin_wrapper.ts` | 258 |
| `RegisterTickFunc` | Function | `plugins/wasm-assemblyscript/assembly/plugin_wrapper.ts` | 158 |

## How to Explore

1. `gitnexus_context({name: "ParseConfigBy"})` — see callers and callees
2. `gitnexus_query({query: "assembly"})` — find related execution flows
3. Read key files listed above for implementation details
