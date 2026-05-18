---
name: helm
description: "Skill for the Helm area of higress. 76 symbols across 18 files."
---

# Helm

76 symbols | 18 files | Cohesion: 86%

## When to Use

- Working with code in `hgctl/`
- Understanding how TestOverlayTrees, TestOverlayYAML, OverlayTrees work
- Modifying helm-related functionality

## Key Files

| File | Symbols |
|------|---------|
| `hgctl/pkg/helm/render.go` | WithName, WithNamespace, WithFS, WithDir, WithVersion (+21) |
| `hgctl/pkg/helm/common.go` | GetProfileFromFlags, GetValuesOverylayFromFiles, ReadLayeredYAMLs, readLayeredYAMLs, GetValueForSetFlag (+12) |
| `hgctl/pkg/helm/profile.go` | SetFlags, ValuesYaml, Validate, ToString, isValidK8SResourceFormat |
| `hgctl/pkg/install.go` | applyFlagAliases, install, promptInstall, promptProfileName |
| `hgctl/pkg/util/util.go` | ParseValue, StripPrefix, IsFilePath, StringBoolMapToSlice |
| `hgctl/pkg/upgrade.go` | upgrade, promptUpgrade, promptProfileContexts |
| `hgctl/pkg/uninstall.go` | uninstall, promptUninstall |
| `hgctl/pkg/util/yaml_test.go` | TestOverlayTrees, TestOverlayYAML |
| `hgctl/pkg/util/yaml.go` | OverlayTrees, OverlayYAML |
| `hgctl/pkg/installer/profile_store.go` | List, listConfigmaps |

## Entry Points

Start here when exploring this area:

- **`TestOverlayTrees`** (Function) — `hgctl/pkg/util/yaml_test.go:66`
- **`TestOverlayYAML`** (Function) — `hgctl/pkg/util/yaml_test.go:128`
- **`OverlayTrees`** (Function) — `hgctl/pkg/util/yaml.go:121`
- **`OverlayYAML`** (Function) — `hgctl/pkg/util/yaml.go:161`
- **`ParseValue`** (Function) — `hgctl/pkg/util/util.go:69`

## Key Symbols

| Symbol | Type | File | Line |
|--------|------|------|------|
| `TestOverlayTrees` | Function | `hgctl/pkg/util/yaml_test.go` | 66 |
| `TestOverlayYAML` | Function | `hgctl/pkg/util/yaml_test.go` | 128 |
| `OverlayTrees` | Function | `hgctl/pkg/util/yaml.go` | 121 |
| `OverlayYAML` | Function | `hgctl/pkg/util/yaml.go` | 161 |
| `ParseValue` | Function | `hgctl/pkg/util/util.go` | 69 |
| `GetInstalledYamlPath` | Function | `hgctl/pkg/installer/installer.go` | 127 |
| `GetProfileFromFlags` | Function | `hgctl/pkg/helm/common.go` | 29 |
| `GetValuesOverylayFromFiles` | Function | `hgctl/pkg/helm/common.go` | 39 |
| `ReadLayeredYAMLs` | Function | `hgctl/pkg/helm/common.go` | 68 |
| `GetValueForSetFlag` | Function | `hgctl/pkg/helm/common.go` | 101 |
| `GenerateConfig` | Function | `hgctl/pkg/helm/common.go` | 123 |
| `UnmarshalProfile` | Function | `hgctl/pkg/helm/common.go` | 276 |
| `GenProfile` | Function | `hgctl/pkg/helm/common.go` | 286 |
| `GenProfileFromProfileContent` | Function | `hgctl/pkg/helm/common.go` | 329 |
| `BuiltinOrDir` | Function | `hgctl/pkg/manifests/manifest.go` | 34 |
| `NewIstioCRDComponent` | Function | `hgctl/pkg/installer/istio.go` | 39 |
| `NewHigressComponent` | Function | `hgctl/pkg/installer/higress.go` | 95 |
| `NewGatewayAPIComponent` | Function | `hgctl/pkg/installer/gateway_api.go` | 40 |
| `WithName` | Function | `hgctl/pkg/helm/render.go` | 147 |
| `WithNamespace` | Function | `hgctl/pkg/helm/render.go` | 153 |

## Execution Flows

| Flow | Type | Steps |
|------|------|-------|
| `NewIstioCRDComponent → RendererOptions` | intra_community | 3 |
| `NewIstioCRDComponent → VerifyRendererOptions` | intra_community | 3 |
| `NewIstioCRDComponent → LocalChartRenderer` | intra_community | 3 |
| `NewGatewayAPIComponent → RendererOptions` | intra_community | 3 |
| `NewGatewayAPIComponent → LocalFileRenderer` | intra_community | 3 |

## Connected Areas

| Area | Connections |
|------|-------------|
| Util | 6 calls |
| Installer | 4 calls |
| Pkg | 3 calls |
| Tpath | 1 calls |

## How to Explore

1. `gitnexus_context({name: "TestOverlayTrees"})` — see callers and callees
2. `gitnexus_query({query: "helm"})` — find related execution flows
3. Read key files listed above for implementation details
