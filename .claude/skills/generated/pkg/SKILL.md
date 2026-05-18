---
name: pkg
description: "Skill for the Pkg area of higress. 185 symbols across 57 files."
---

# Pkg

185 symbols | 57 files | Cohesion: 82%

## When to Use

- Working with code in `plugins/`
- Understanding how GetRootCommand, ClosePortForwarderOnInterrupt, AddKubeConfigFlags work
- Modifying pkg-related functionality

## Key Files

| File | Symbols |
|------|---------|
| `plugins/wasm-go/extensions/traffic-editor/pkg/command_test.go` | TestNewDeleteCommand_Success, TestNewDeleteCommand_MissingTarget, TestSetExecutor_Run_SingleStage, TestConcatExecutor_Run_SingleStage, TestConcatExecutor_Run_MultiStages (+17) |
| `plugins/wasm-go/extensions/traffic-editor/pkg/condition_test.go` | TestEqualsCondition_Match, TestEqualsCondition_NoMatch, TestPrefixCondition_Match, TestPrefixCondition_NoMatch, TestSuffixCondition_Match (+13) |
| `plugins/wasm-go/extensions/traffic-editor/pkg/context.go` | NewEditorContext, SetRequestPath, SetRequestQueries, SetRefValue, SetRefValues (+9) |
| `plugins/wasm-go/extensions/traffic-editor/pkg/context_test.go` | TestEditorContext_CommandSetAndExecutors, TestEditorContext_Stage, TestEditorContext_RequestPath, TestEditorContext_RequestHeaders, TestEditorContext_RequestQueries (+8) |
| `hgctl/pkg/dashboard.go` | newDashboardCmd, promDashCmd, consoleDashCmd, grafanaDashCmd, envoyDashCmd (+6) |
| `plugins/wasm-go/extensions/traffic-editor/pkg/command.go` | Command, newDeleteCommand, CreateExecutor, Run, FromJson (+6) |
| `plugins/wasm-go/extensions/traffic-editor/pkg/condition.go` | CreateCondition, newEqualsCondition, newPrefixCondition, newSuffixCondition, newContainsCondition (+5) |
| `hgctl/pkg/code_debug.go` | newCodeDebugCmd, getStartCodeDebugCmd, getStopCodeDebugCmd, getNonLoopbackIPv4, updateXdsIpAndRollout (+3) |
| `hgctl/pkg/completion.go` | newCompletionCmd, runCompletionBash, runCompletionZsh, runCompletionFish, runCompletionPowershell |
| `hgctl/pkg/version.go` | newVersionCommand, retrieveVersion, versions |

## Entry Points

Start here when exploring this area:

- **`GetRootCommand`** (Function) — `hgctl/pkg/root.go:26`
- **`ClosePortForwarderOnInterrupt`** (Function) — `hgctl/pkg/dashboard.go:392`
- **`AddKubeConfigFlags`** (Function) — `pkg/cmd/options/global.go:23`
- **`NewCommand`** (Function) — `hgctl/pkg/plugin/plugin.go:28`
- **`AddHigressNamespaceFlags`** (Function) — `hgctl/pkg/kubernetes/wasmplugin.go:44`

## Key Symbols

| Symbol | Type | File | Line |
|--------|------|------|------|
| `GetRootCommand` | Function | `hgctl/pkg/root.go` | 26 |
| `ClosePortForwarderOnInterrupt` | Function | `hgctl/pkg/dashboard.go` | 392 |
| `AddKubeConfigFlags` | Function | `pkg/cmd/options/global.go` | 23 |
| `NewCommand` | Function | `hgctl/pkg/plugin/plugin.go` | 28 |
| `AddHigressNamespaceFlags` | Function | `hgctl/pkg/kubernetes/wasmplugin.go` | 44 |
| `NewCLIClient` | Function | `hgctl/pkg/kubernetes/client.go` | 73 |
| `NewFileDirProfileStore` | Function | `hgctl/pkg/installer/profile_store.go` | 123 |
| `NewConfigmapProfileStore` | Function | `hgctl/pkg/installer/profile_store.go` | 241 |
| `GetHomeDir` | Function | `hgctl/pkg/installer/installer.go` | 70 |
| `GetHgctlPath` | Function | `hgctl/pkg/installer/installer.go` | 79 |
| `GetProfileInstalledPath` | Function | `hgctl/pkg/installer/installer.go` | 111 |
| `NewMCPCmd` | Function | `hgctl/pkg/agent/mcp.go` | 66 |
| `NewAgentCmd` | Function | `hgctl/pkg/agent/agent.go` | 32 |
| `NewCommand` | Function | `hgctl/pkg/plugin/uninstall/uninstall.go` | 30 |
| `NewCommand` | Function | `hgctl/pkg/plugin/test/test.go` | 20 |
| `AddOptionFileFlag` | Function | `hgctl/pkg/plugin/option/option.go` | 92 |
| `NewCommand` | Function | `hgctl/pkg/plugin/ls/ls.go` | 32 |
| `NewCommand` | Function | `hgctl/pkg/plugin/install/install.go` | 54 |
| `NewCommand` | Function | `hgctl/pkg/plugin/init/init.go` | 31 |
| `NewCommand` | Function | `hgctl/pkg/plugin/config/config.go` | 18 |

## Execution Flows

| Flow | Type | Steps |
|------|------|-------|
| `Run → GetRequestHeader` | cross_community | 5 |
| `Run → Ref` | intra_community | 5 |
| `Run → GetRequestQuery` | cross_community | 4 |
| `Run → GetResponseHeader` | cross_community | 4 |
| `OnHttpRequestHeaders → SetExecutor` | cross_community | 4 |
| `OnHttpRequestHeaders → ConcatExecutor` | cross_community | 4 |
| `OnHttpRequestHeaders → CopyExecutor` | cross_community | 4 |
| `OnHttpRequestHeaders → DeleteExecutor` | cross_community | 4 |
| `GetRootCommand → AddKubeConfigFlags` | intra_community | 4 |
| `GetRootCommand → AddOptionFileFlag` | intra_community | 4 |

## Connected Areas

| Area | Connections |
|------|-------------|
| Agent | 7 calls |
| Helm | 6 calls |
| Config | 6 calls |
| Docker | 5 calls |
| Test | 5 calls |
| Kubernetes | 3 calls |
| Types | 1 calls |
| Install | 1 calls |

## How to Explore

1. `gitnexus_context({name: "GetRootCommand"})` — see callers and callees
2. `gitnexus_query({query: "pkg"})` — find related execution flows
3. Read key files listed above for implementation details
