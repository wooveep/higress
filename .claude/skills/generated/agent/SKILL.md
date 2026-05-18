---
name: agent
description: "Skill for the Agent area of higress. 85 symbols across 13 files."
---

# Agent

85 symbols | 13 files | Cohesion: 77%

## When to Use

- Working with code in `hgctl/`
- Understanding how NewHimarketClient, GetHigressGatewayServiceIP, GetSpecificAgentDir work
- Modifying agent-related functionality

## Key Files

| File | Symbols |
|------|---------|
| `hgctl/pkg/agent/utils.go` | getAgentConfig, importAgentFromCore, queryAgentSysPrompt, queryAgentTools, queryAgentModel (+17) |
| `hgctl/pkg/agent/deploy.go` | validate, RunCmd, checkAgentRunEnvironment, checkRequiredEnvironment, getAgentType (+12) |
| `hgctl/pkg/agent/base.go` | generateAgentPromptByCore, init, check, checkAgentInstall, promptAgentInstall (+10) |
| `hgctl/pkg/agent/core.go` | run, Start, AddMCPServer, ImproveNewAgent, runInTargetDir (+4) |
| `hgctl/pkg/agent/mcp.go` | newHanlder, addHTTPMCP, addOpenAPIMCP, parseOpenapiSpec, handleAddMCP (+2) |
| `hgctl/pkg/agent/agent.go` | newAgentAddCmd, invokeAgentCore, handleAddAgent, validateArg |
| `hgctl/pkg/agent/new.go` | afterCreatedAgent, runAgenticCoreImprovement, promptAfterCreatedAgent |
| `hgctl/pkg/util/util.go` | GetSpecificAgentDir, GetHomeHgctlDir |
| `hgctl/pkg/util/env.go` | GetPythonVersion, CompareVersions |
| `hgctl/pkg/agent/services/client.go` | NewHimarketClient |

## Entry Points

Start here when exploring this area:

- **`NewHimarketClient`** (Function) — `hgctl/pkg/agent/services/client.go:72`
- **`GetHigressGatewayServiceIP`** (Function) — `hgctl/pkg/agent/utils.go:140`
- **`GetSpecificAgentDir`** (Function) — `hgctl/pkg/util/util.go:105`
- **`InitConfig`** (Function) — `hgctl/pkg/agent/config.go:101`
- **`GetHomeHgctlDir`** (Function) — `hgctl/pkg/util/util.go:99`

## Key Symbols

| Symbol | Type | File | Line |
|--------|------|------|------|
| `NewHimarketClient` | Function | `hgctl/pkg/agent/services/client.go` | 72 |
| `GetHigressGatewayServiceIP` | Function | `hgctl/pkg/agent/utils.go` | 140 |
| `GetSpecificAgentDir` | Function | `hgctl/pkg/util/util.go` | 105 |
| `InitConfig` | Function | `hgctl/pkg/agent/config.go` | 101 |
| `GetHomeHgctlDir` | Function | `hgctl/pkg/util/util.go` | 99 |
| `GetPythonVersion` | Function | `hgctl/pkg/util/env.go` | 24 |
| `CompareVersions` | Function | `hgctl/pkg/util/env.go` | 49 |
| `ExtractEmbedFiles` | Function | `hgctl/pkg/manifests/manifest.go` | 42 |
| `NewAgenticCore` | Function | `hgctl/pkg/agent/core.go` | 35 |
| `Start` | Method | `hgctl/pkg/agent/core.go` | 230 |
| `AddMCPServer` | Method | `hgctl/pkg/agent/core.go` | 235 |
| `RunCmd` | Method | `hgctl/pkg/agent/deploy.go` | 80 |
| `Deploy` | Method | `hgctl/pkg/agent/deploy.go` | 289 |
| `HandleAgentRun` | Method | `hgctl/pkg/agent/deploy.go` | 321 |
| `CheckServerlessAccessKey` | Method | `hgctl/pkg/agent/deploy.go` | 347 |
| `ImproveNewAgent` | Method | `hgctl/pkg/agent/core.go` | 64 |
| `GetCoreDirName` | Method | `hgctl/pkg/agent/core.go` | 53 |
| `Setup` | Method | `hgctl/pkg/agent/core.go` | 107 |
| `RunPythonCmd` | Method | `hgctl/pkg/agent/deploy.go` | 103 |
| `HandleLocal` | Method | `hgctl/pkg/agent/deploy.go` | 366 |

## Execution Flows

| Flow | Type | Steps |
|------|------|-------|
| `GetRootCommand → MCPAddArg` | cross_community | 4 |

## Connected Areas

| Area | Connections |
|------|-------------|
| Services | 7 calls |
| Pkg | 4 calls |
| Helm | 1 calls |

## How to Explore

1. `gitnexus_context({name: "NewHimarketClient"})` — see callers and callees
2. `gitnexus_query({query: "agent"})` — find related execution flows
3. Read key files listed above for implementation details
