---
name: util
description: "Skill for the Util area of higress. 122 symbols across 22 files."
---

# Util

122 symbols | 22 files | Cohesion: 81%

## When to Use

- Working with code in `hgctl/`
- Understanding how TestToYAML, TestYAMLDiff, TestMultipleYAMLDiff work
- Modifying util-related functionality

## Key Files

| File | Symbols |
|------|---------|
| `plugins/wasm-go/extensions/frontend-gray/util/utils.go` | GetRequestPath, IsRequestSkippedByHeaders, IsIndexRequest, IsGrayEnabled, CheckIsHtmlRequest (+20) |
| `hgctl/pkg/util/reflect.go` | kindOf, IsString, IsPtr, IsSlice, IsStruct (+14) |
| `hgctl/pkg/util/path.go` | PathFromString, IsValidPathElement, IsKVPathElement, IsVPathElement, IsNPathElement (+8) |
| `hgctl/pkg/util/path_test.go` | TestPathFromString, TestSplitEscaped, TestIsNPathElement, stringSlicesEqual, TestIsKVPathElement (+6) |
| `hgctl/pkg/util/yaml.go` | ToYAML, yamlDiff, yamlStringsToList, multiYamlDiffOutput, diffStringList (+2) |
| `hgctl/pkg/helm/tpath/tree_test.go` | TestWritePathContext, TestWriteNode, TestMergeNode, errToString, TestSecretVolumes (+1) |
| `plugins/wasm-go/extensions/frontend-gray/util/utils_test.go` | TestCheckIsHtmlRequest, TestGetCookieValue, TestIndexRewrite, TestIndexRewrite2, TestPrefixFileRewrite (+1) |
| `hgctl/pkg/helm/tpath/tree.go` | GetPathContext, getPathContext, stringsEqual, matchesRegex, mergeConditional |
| `hgctl/pkg/util/yaml_test.go` | TestToYAML, TestYAMLDiff, TestMultipleYAMLDiff, TestIsYAMLEmpty |
| `plugins/wasm-go/extensions/ai-token-ratelimit/util/utils.go` | ReconvertHeaders, GetRouteName, GetClusterName, GetConsumer |

## Entry Points

Start here when exploring this area:

- **`TestToYAML`** (Function) — `hgctl/pkg/util/yaml_test.go:22`
- **`TestYAMLDiff`** (Function) — `hgctl/pkg/util/yaml_test.go:173`
- **`TestMultipleYAMLDiff`** (Function) — `hgctl/pkg/util/yaml_test.go:218`
- **`ToYAML`** (Function) — `hgctl/pkg/util/yaml.go:67`
- **`YAMLDiff`** (Function) — `hgctl/pkg/util/yaml.go:281`

## Key Symbols

| Symbol | Type | File | Line |
|--------|------|------|------|
| `TestToYAML` | Function | `hgctl/pkg/util/yaml_test.go` | 22 |
| `TestYAMLDiff` | Function | `hgctl/pkg/util/yaml_test.go` | 173 |
| `TestMultipleYAMLDiff` | Function | `hgctl/pkg/util/yaml_test.go` | 218 |
| `ToYAML` | Function | `hgctl/pkg/util/yaml.go` | 67 |
| `YAMLDiff` | Function | `hgctl/pkg/util/yaml.go` | 281 |
| `TestPathFromString` | Function | `hgctl/pkg/util/path_test.go` | 116 |
| `PathFromString` | Function | `hgctl/pkg/util/path.go` | 47 |
| `TestGetConfigSubtree` | Function | `hgctl/pkg/helm/tpath/util_test.go` | 64 |
| `GetSpecSubtree` | Function | `hgctl/pkg/helm/tpath/util.go` | 37 |
| `GetConfigSubtree` | Function | `hgctl/pkg/helm/tpath/util.go` | 42 |
| `TestWritePathContext` | Function | `hgctl/pkg/helm/tpath/tree_test.go` | 23 |
| `TestWriteNode` | Function | `hgctl/pkg/helm/tpath/tree_test.go` | 366 |
| `TestMergeNode` | Function | `hgctl/pkg/helm/tpath/tree_test.go` | 594 |
| `TestSecretVolumes` | Function | `hgctl/pkg/helm/tpath/tree_test.go` | 683 |
| `TestWriteEscapedPathContext` | Function | `hgctl/pkg/helm/tpath/tree_test.go` | 786 |
| `GetPathContext` | Function | `hgctl/pkg/helm/tpath/tree.go` | 59 |
| `TestSplitEscaped` | Function | `hgctl/pkg/util/path_test.go` | 21 |
| `TestIsNPathElement` | Function | `hgctl/pkg/util/path_test.go` | 67 |
| `TestIsKVPathElement` | Function | `hgctl/pkg/util/path_test.go` | 170 |
| `TestIsVPathElement` | Function | `hgctl/pkg/util/path_test.go` | 217 |

## Execution Flows

| Flow | Type | Steps |
|------|------|-------|
| `OnHttpRequestHeaders → ContainsValue` | cross_community | 4 |
| `OnHttpRequestHeaders → IsIndexRequest` | intra_community | 3 |
| `OnHttpRequestHeaders → IsRequestSkippedByHeaders` | intra_community | 3 |

## Connected Areas

| Area | Connections |
|------|-------------|
| Tpath | 7 calls |
| Helm | 1 calls |
| Installer | 1 calls |
| Config | 1 calls |

## How to Explore

1. `gitnexus_context({name: "TestToYAML"})` — see callers and callees
2. `gitnexus_query({query: "util"})` — find related execution flows
3. Read key files listed above for implementation details
