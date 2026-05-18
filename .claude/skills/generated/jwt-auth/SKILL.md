---
name: jwt-auth
description: "Skill for the Jwt_auth area of higress. 100 symbols across 28 files."
---

# Jwt_auth

100 symbols | 28 files | Cohesion: 91%

## When to Use

- Working with code in `plugins/`
- Understanding how empty, JsonValueAs, JsonArrayIterate work
- Modifying jwt_auth-related functionality

## Key Files

| File | Symbols |
|------|---------|
| `plugins/wasm-cpp/common/route_rule_matcher.h` | setInvalidConfig, getRules, globalAuthDisable, checkAuthRule, setEmptyGlobalConfig (+7) |
| `plugins/wasm-cpp/extensions/jwt_auth/extractor.cc` | check, isClaimAllowed, claimsToHeaders, extract, JwtLocationBase (+4) |
| `plugins/wasm-cpp/extensions/jwt_auth/plugin.cc` | generateRcDetails, parsePluginConfig, consumerVerify, checkPlugin, onConfigure (+2) |
| `plugins/wasm-cpp/extensions/hmac_auth/plugin.cc` | deniedInvalidCaKey, deniedUnauthorizedConsumer, getStringToSignWithParam, parsePluginConfig, checkConsumer (+2) |
| `plugins/wasm-cpp/extensions/key_rate_limit/plugin.cc` | tooManyRequest, parsePluginConfig, checkPlugin, onConfigure, configure |
| `plugins/wasm-cpp/common/json_util.h` | JsonValueAs, value, JsonArrayIterate, JsonObjectIterate |
| `plugins/wasm-cpp/extensions/request_block/plugin.cc` | parsePluginConfig, onConfigure, configure, checkHeader |
| `plugins/wasm-cpp/extensions/oauth/plugin.cc` | parsePluginConfig, checkPlugin, onConfigure, configure |
| `plugins/wasm-cpp/extensions/model_router/plugin.cc` | parsePluginConfig, onConfigure, configure, onJsonBody |
| `plugins/wasm-cpp/extensions/model_mapper/plugin.cc` | parsePluginConfig, onConfigure, configure, onBody |

## Entry Points

Start here when exploring this area:

- **`empty`** (Function) — `plugins/wasm-rust/extensions/ai-data-masking/src/deny_word.rs:46`
- **`JsonValueAs`** (Function) — `plugins/wasm-cpp/common/json_util.h:38`
- **`JsonArrayIterate`** (Function) — `plugins/wasm-cpp/common/json_util.h:106`
- **`JsonObjectIterate`** (Function) — `plugins/wasm-cpp/common/json_util.h:113`
- **`JsonArrayIterate`** (Function) — `plugins/wasm-cpp/common/json_util.cc:144`

## Key Symbols

| Symbol | Type | File | Line |
|--------|------|------|------|
| `JwtLocation` | Class | `plugins/wasm-cpp/extensions/jwt_auth/extractor.h` | 46 |
| `Extractor` | Class | `plugins/wasm-cpp/extensions/jwt_auth/extractor.h` | 71 |
| `empty` | Function | `plugins/wasm-rust/extensions/ai-data-masking/src/deny_word.rs` | 46 |
| `JsonValueAs` | Function | `plugins/wasm-cpp/common/json_util.h` | 38 |
| `JsonArrayIterate` | Function | `plugins/wasm-cpp/common/json_util.h` | 106 |
| `JsonObjectIterate` | Function | `plugins/wasm-cpp/common/json_util.h` | 113 |
| `JsonArrayIterate` | Function | `plugins/wasm-cpp/common/json_util.cc` | 144 |
| `JsonObjectIterate` | Function | `plugins/wasm-cpp/common/json_util.cc` | 162 |
| `buildOriginalUri` | Function | `plugins/wasm-cpp/common/http_util.cc` | 252 |
| `hasRequestBody` | Function | `plugins/wasm-cpp/common/http_util.cc` | 313 |
| `getMD5` | Function | `plugins/wasm-cpp/common/crypto_util.cc` | 102 |
| `getMD5Base64` | Function | `plugins/wasm-cpp/common/crypto_util.cc` | 109 |
| `crypt_ssha` | Function | `plugins/wasm-cpp/common/crypto_util.cc` | 217 |
| `getToken` | Function | `plugins/wasm-cpp/extensions/key_rate_limit/bucket.h` | 37 |
| `initializeTokenBucket` | Function | `plugins/wasm-cpp/extensions/key_rate_limit/bucket.h` | 39 |
| `TEST_F` | Function | `plugins/wasm-cpp/extensions/basic_auth/plugin_test.cc` | 118 |
| `TEST_F` | Function | `plugins/wasm-cpp/extensions/jwt_auth/plugin_test.cc` | 127 |
| `setInvalidConfig` | Method | `plugins/wasm-cpp/common/route_rule_matcher.h` | 82 |
| `getRules` | Method | `plugins/wasm-cpp/common/route_rule_matcher.h` | 84 |
| `globalAuthDisable` | Method | `plugins/wasm-cpp/common/route_rule_matcher.h` | 96 |

## Connected Areas

| Area | Connections |
|------|-------------|
| Ai-cache | 1 calls |
| Key_auth | 1 calls |

## How to Explore

1. `gitnexus_context({name: "empty"})` — see callers and callees
2. `gitnexus_query({query: "jwt_auth"})` — find related execution flows
3. Read key files listed above for implementation details
