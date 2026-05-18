---
name: config
description: "Skill for the Config area of higress. 321 symbols across 76 files."
---

# Config

321 symbols | 76 files | Cohesion: 83%

## When to Use

- Working with code in `plugins/`
- Understanding how TestEvaluateRiskWithConsumerRiskAction, TestRiskLevelFunctions, TestCustomLabelDetailExceedsThreshold work
- Modifying config-related functionality

## Key Files

| File | Symbols |
|------|---------|
| `plugins/wasm-go/extensions/ai-security-guard/config/evaluate_risk_test.go` | baseConfig, TestTC_EVAL_001, TestTC_EVAL_002, TestTC_EVAL_003, TestTC_EVAL_004 (+40) |
| `pkg/ingress/config/ingress_config.go` | NewIngressConfig, convertEnvoyFilter, constructHttp2RpcEnvoyFilter, constructHttp2RpcMethods, buildPatchStruct (+31) |
| `plugins/wasm-go/extensions/ai-security-guard/config/config.go` | LevelToInt, EvaluateRisk, dimensionActionKey, getGlobalDimensionAction, enforceMaskBoundary (+17) |
| `pkg/ingress/config/kingress_config.go` | NewKIngressConfig, convertVirtualService, normalizeWeightedKCluster, applyAppRoot, List (+11) |
| `plugins/wasm-go/extensions/ai-security-guard/config/action_resolver_test.go` | TestTC_RESOLVE_001, TestTC_RESOLVE_002, TestTC_RESOLVE_003, TestTC_RESOLVE_004, TestTC_RESOLVE_005 (+10) |
| `pkg/ingress/kube/common/tool.go` | ConvertToDNSLabelValid, CreateConvertedName, constructRouteName, V1Available, NetworkingIngressAvailable (+5) |
| `plugins/wasm-go/extensions/de-graphql/config/degraphql_config.go` | SetEndpoint, SetDomain, SetTimeout, GetDomain, GetEndpoint (+5) |
| `plugins/wasm-go/extensions/ai-security-guard/config/evaluate_risk_property_test.go` | TestProperty1_AboveThresholdMaskProducesRiskMask, TestProperty2_BelowThresholdMaskProducesRiskPass, TestProperty3_PerDetailThresholdIndependence, describeLevels, TestProperty4a_SuggestionBlockRespectsThreshold (+4) |
| `pkg/ingress/translation/translation.go` | NewIngressTranslation, AddLocalCluster, List, Run, SetWatchErrorHandler (+4) |
| `hgctl/cmd/hgctl/config/gateway_config_test.go` | newFakePortForwarder, Start, Stop, Address, TestExtractAllConfigDump (+3) |

## Entry Points

Start here when exploring this area:

- **`TestEvaluateRiskWithConsumerRiskAction`** (Function) — `plugins/wasm-go/extensions/ai-security-guard/main_test.go:1049`
- **`TestRiskLevelFunctions`** (Function) — `plugins/wasm-go/extensions/ai-security-guard/main_test.go:1159`
- **`TestCustomLabelDetailExceedsThreshold`** (Function) — `plugins/wasm-go/extensions/ai-security-guard/main_test.go:1315`
- **`TestTC_EVAL_001`** (Function) — `plugins/wasm-go/extensions/ai-security-guard/config/evaluate_risk_test.go:41`
- **`TestTC_EVAL_002`** (Function) — `plugins/wasm-go/extensions/ai-security-guard/config/evaluate_risk_test.go:63`

## Key Symbols

| Symbol | Type | File | Line |
|--------|------|------|------|
| `TestEvaluateRiskWithConsumerRiskAction` | Function | `plugins/wasm-go/extensions/ai-security-guard/main_test.go` | 1049 |
| `TestRiskLevelFunctions` | Function | `plugins/wasm-go/extensions/ai-security-guard/main_test.go` | 1159 |
| `TestCustomLabelDetailExceedsThreshold` | Function | `plugins/wasm-go/extensions/ai-security-guard/main_test.go` | 1315 |
| `TestTC_EVAL_001` | Function | `plugins/wasm-go/extensions/ai-security-guard/config/evaluate_risk_test.go` | 41 |
| `TestTC_EVAL_002` | Function | `plugins/wasm-go/extensions/ai-security-guard/config/evaluate_risk_test.go` | 63 |
| `TestTC_EVAL_003` | Function | `plugins/wasm-go/extensions/ai-security-guard/config/evaluate_risk_test.go` | 83 |
| `TestTC_EVAL_004` | Function | `plugins/wasm-go/extensions/ai-security-guard/config/evaluate_risk_test.go` | 104 |
| `TestTC_EVAL_005` | Function | `plugins/wasm-go/extensions/ai-security-guard/config/evaluate_risk_test.go` | 132 |
| `TestTC_EVAL_006` | Function | `plugins/wasm-go/extensions/ai-security-guard/config/evaluate_risk_test.go` | 159 |
| `TestTC_EVAL_007` | Function | `plugins/wasm-go/extensions/ai-security-guard/config/evaluate_risk_test.go` | 183 |
| `TestTC_EVAL_008` | Function | `plugins/wasm-go/extensions/ai-security-guard/config/evaluate_risk_test.go` | 196 |
| `TestTC_EVAL_009` | Function | `plugins/wasm-go/extensions/ai-security-guard/config/evaluate_risk_test.go` | 210 |
| `TestTC_EVAL_010` | Function | `plugins/wasm-go/extensions/ai-security-guard/config/evaluate_risk_test.go` | 234 |
| `TestTC_EVAL_011` | Function | `plugins/wasm-go/extensions/ai-security-guard/config/evaluate_risk_test.go` | 272 |
| `TestTC_EVAL_012` | Function | `plugins/wasm-go/extensions/ai-security-guard/config/evaluate_risk_test.go` | 286 |
| `TestTC_EVAL_013` | Function | `plugins/wasm-go/extensions/ai-security-guard/config/evaluate_risk_test.go` | 301 |
| `TestTC_EVAL_014` | Function | `plugins/wasm-go/extensions/ai-security-guard/config/evaluate_risk_test.go` | 326 |
| `TestTC_EVAL_015` | Function | `plugins/wasm-go/extensions/ai-security-guard/config/evaluate_risk_test.go` | 345 |
| `TestTC_EVAL_016` | Function | `plugins/wasm-go/extensions/ai-security-guard/config/evaluate_risk_test.go` | 365 |
| `TestTC_EVAL_017` | Function | `plugins/wasm-go/extensions/ai-security-guard/config/evaluate_risk_test.go` | 386 |

## Execution Flows

| Flow | Type | Steps |
|------|------|-------|
| `HandleQwenImageGenerationRequestBody → Match` | cross_community | 7 |
| `HandleOpenAIImageGenerationRequestBody → Match` | cross_community | 7 |
| `HandleTextGenerationRequestBody → Match` | cross_community | 6 |
| `HandleQwenImageGenerationRequestBody → GetGlobalDimensionAction` | cross_community | 6 |
| `HandleQwenImageGenerationRequestBody → DimensionActionKey` | cross_community | 6 |
| `HandleQwenImageGenerationRequestBody → EnforceMaskBoundary` | cross_community | 6 |
| `HandleOpenAIImageGenerationRequestBody → GetGlobalDimensionAction` | cross_community | 6 |
| `HandleOpenAIImageGenerationRequestBody → DimensionActionKey` | cross_community | 6 |
| `HandleOpenAIImageGenerationRequestBody → EnforceMaskBoundary` | cross_community | 6 |
| `HandleTextGenerationRequestBody → GetGlobalDimensionAction` | cross_community | 5 |

## Connected Areas

| Area | Connections |
|------|-------------|
| Ingress | 8 calls |
| Kingress | 6 calls |
| Ai-security-guard | 6 calls |
| Istio | 3 calls |
| Provider | 3 calls |
| V1 | 2 calls |
| V1alpha1 | 2 calls |
| Configmap | 2 calls |

## How to Explore

1. `gitnexus_context({name: "TestEvaluateRiskWithConsumerRiskAction"})` — see callers and callees
2. `gitnexus_query({query: "config"})` — find related execution flows
3. Read key files listed above for implementation details
