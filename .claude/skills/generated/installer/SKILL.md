---
name: installer
description: "Skill for the Installer area of higress. 77 symbols across 22 files."
---

# Installer

77 symbols | 22 files | Cohesion: 68%

## When to Use

- Working with code in `hgctl/`
- Understanding how TestParseK8sObjectsFromYAMLManifest, ParseK8sObjectsFromYAMLManifest, NewServerInfo work
- Modifying installer-related functionality

## Key Files

| File | Symbols |
|------|---------|
| `hgctl/pkg/installer/standalone_agent.go` | Version, NewAgent, run, Upgrade, promptRestart (+9) |
| `hgctl/pkg/installer/installer_k8s.go` | ApplyManifests, applyManifest, DeleteManifests, deleteManifest, isNamespacedObject (+8) |
| `hgctl/pkg/installer/component.go` | WithComponentNamespace, WithComponentChartPath, WithComponentCapabilities, WithDevel, renderComponentManifest (+4) |
| `hgctl/pkg/installer/standalone.go` | Install, Upgrade, UnInstall, NewStandaloneComponent, prepareProfile |
| `hgctl/pkg/helm/object/objects.go` | UnstructuredObject, Unstructured, UnstructuredItems, ParseK8sObjectsFromYAMLManifest |
| `hgctl/pkg/installer/installer_docker.go` | Install, Upgrade, UnInstall, NewDockerInstaller |
| `hgctl/pkg/installer/profile_store.go` | Save, applyConfigmap, Delete, deleteConfigmap |
| `hgctl/pkg/helm/profile.go` | IstioEnabled, GatewayAPIEnabled, GetIstioNamespace |
| `hgctl/pkg/util/http_fetcher.go` | Fetch, retryable, NewHTTPFetcher |
| `hgctl/pkg/installer/istio.go` | ComponentName, RenderManifest |

## Entry Points

Start here when exploring this area:

- **`TestParseK8sObjectsFromYAMLManifest`** (Function) — `hgctl/pkg/helm/object/objects_test.go:404`
- **`ParseK8sObjectsFromYAMLManifest`** (Function) — `hgctl/pkg/helm/object/objects.go:247`
- **`NewServerInfo`** (Function) — `hgctl/pkg/installer/server_info.go:60`
- **`NewK8sInstaller`** (Function) — `hgctl/pkg/installer/installer_k8s.go:256`
- **`WithComponentNamespace`** (Function) — `hgctl/pkg/installer/component.go:60`

## Key Symbols

| Symbol | Type | File | Line |
|--------|------|------|------|
| `TestParseK8sObjectsFromYAMLManifest` | Function | `hgctl/pkg/helm/object/objects_test.go` | 404 |
| `ParseK8sObjectsFromYAMLManifest` | Function | `hgctl/pkg/helm/object/objects.go` | 247 |
| `NewServerInfo` | Function | `hgctl/pkg/installer/server_info.go` | 60 |
| `NewK8sInstaller` | Function | `hgctl/pkg/installer/installer_k8s.go` | 256 |
| `WithComponentNamespace` | Function | `hgctl/pkg/installer/component.go` | 60 |
| `WithComponentChartPath` | Function | `hgctl/pkg/installer/component.go` | 66 |
| `WithComponentCapabilities` | Function | `hgctl/pkg/installer/component.go` | 90 |
| `WithDevel` | Function | `hgctl/pkg/installer/component.go` | 102 |
| `NewHelmAgent` | Function | `hgctl/pkg/installer/helm_agent.go` | 44 |
| `WriteFileString` | Function | `hgctl/pkg/util/util.go` | 84 |
| `NewInstaller` | Function | `hgctl/pkg/installer/installer.go` | 53 |
| `NewHTTPFetcher` | Function | `hgctl/pkg/util/http_fetcher.go` | 38 |
| `NewAgent` | Function | `hgctl/pkg/installer/standalone_agent.go` | 53 |
| `NewStandaloneComponent` | Function | `hgctl/pkg/installer/standalone.go` | 98 |
| `GetDefaultInstallPackagePath` | Function | `hgctl/pkg/installer/installer.go` | 95 |
| `NewDockerInstaller` | Function | `hgctl/pkg/installer/installer_docker.go` | 77 |
| `WithComponentChartName` | Function | `hgctl/pkg/installer/component.go` | 72 |
| `WithComponentRepoURL` | Function | `hgctl/pkg/installer/component.go` | 78 |
| `WithComponentVersion` | Function | `hgctl/pkg/installer/component.go` | 84 |
| `WithQuiet` | Function | `hgctl/pkg/installer/component.go` | 96 |

## Connected Areas

| Area | Connections |
|------|-------------|
| Pkg | 6 calls |
| Helm | 5 calls |
| Util | 4 calls |
| Agent | 2 calls |
| Object | 2 calls |

## How to Explore

1. `gitnexus_context({name: "TestParseK8sObjectsFromYAMLManifest"})` — see callers and callees
2. `gitnexus_query({query: "installer"})` — find related execution flows
3. Read key files listed above for implementation details
