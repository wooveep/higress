---
name: tests
description: "Skill for the Tests area of higress. 75 symbols across 67 files."
---

# Tests

75 symbols | 67 files | Cohesion: 95%

## When to Use

- Working with code in `test/`
- Understanding how Register, TestRps10, TestRps10Burst3 work
- Modifying tests-related functionality

## Key Files

| File | Symbols |
|------|---------|
| `test/e2e/conformance/tests/httproute-limit.go` | init, TestRps10, TestRps10Burst3, TestRpm10Burst3, AssertRps (+4) |
| `test/e2e/conformance/tests/tests.go` | Register |
| `test/e2e/conformance/tests/rust-wasm-request-block.go` | init |
| `test/e2e/conformance/tests/rust-wasm-ai-data-masking.go` | init |
| `test/e2e/conformance/tests/ingress-loadbalance-mcp-sse.go` | init |
| `test/e2e/conformance/tests/httproute-whitelist-source-range.go` | init |
| `test/e2e/conformance/tests/httproute-timeout.go` | init |
| `test/e2e/conformance/tests/httproute-temporal-redirect.go` | init |
| `test/e2e/conformance/tests/httproute-static-registry.go` | init |
| `test/e2e/conformance/tests/httproute-simple-same-namespace.go` | init |

## Entry Points

Start here when exploring this area:

- **`Register`** (Function) — `test/e2e/conformance/tests/tests.go:18`
- **`TestRps10`** (Function) — `test/e2e/conformance/tests/httproute-limit.go:59`
- **`TestRps10Burst3`** (Function) — `test/e2e/conformance/tests/httproute-limit.go:97`
- **`TestRpm10Burst3`** (Function) — `test/e2e/conformance/tests/httproute-limit.go:135`
- **`AssertRps`** (Function) — `test/e2e/conformance/tests/httproute-limit.go:243`

## Key Symbols

| Symbol | Type | File | Line |
|--------|------|------|------|
| `Register` | Function | `test/e2e/conformance/tests/tests.go` | 18 |
| `TestRps10` | Function | `test/e2e/conformance/tests/httproute-limit.go` | 59 |
| `TestRps10Burst3` | Function | `test/e2e/conformance/tests/httproute-limit.go` | 97 |
| `TestRpm10Burst3` | Function | `test/e2e/conformance/tests/httproute-limit.go` | 135 |
| `AssertRps` | Function | `test/e2e/conformance/tests/httproute-limit.go` | 243 |
| `TestRps50` | Function | `test/e2e/conformance/tests/httproute-limit.go` | 78 |
| `TestRpm10` | Function | `test/e2e/conformance/tests/httproute-limit.go` | 116 |
| `DoRequest` | Function | `test/e2e/conformance/tests/httproute-limit.go` | 153 |
| `ParallelRunner` | Function | `test/e2e/conformance/tests/httproute-limit.go` | 197 |
| `init` | Function | `test/e2e/conformance/tests/rust-wasm-request-block.go` | 23 |
| `init` | Function | `test/e2e/conformance/tests/rust-wasm-ai-data-masking.go` | 23 |
| `init` | Function | `test/e2e/conformance/tests/ingress-loadbalance-mcp-sse.go` | 23 |
| `init` | Function | `test/e2e/conformance/tests/httproute-whitelist-source-range.go` | 23 |
| `init` | Function | `test/e2e/conformance/tests/httproute-timeout.go` | 23 |
| `init` | Function | `test/e2e/conformance/tests/httproute-temporal-redirect.go` | 24 |
| `init` | Function | `test/e2e/conformance/tests/httproute-static-registry.go` | 23 |
| `init` | Function | `test/e2e/conformance/tests/httproute-simple-same-namespace.go` | 23 |
| `init` | Function | `test/e2e/conformance/tests/httproute-same-host-and-path.go` | 23 |
| `init` | Function | `test/e2e/conformance/tests/httproute-rewrite-path.go` | 23 |
| `init` | Function | `test/e2e/conformance/tests/httproute-rewrite-host.go` | 23 |

## How to Explore

1. `gitnexus_context({name: "Register"})` — see callers and callees
2. `gitnexus_query({query: "tests"})` — find related execution flows
3. Read key files listed above for implementation details
