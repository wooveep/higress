# Higress 项目 Memory

> 上次更新: 2026-03-23

## 项目基本信息

- **名称**: Higress (AI Native API Gateway)
- **版本**: v2.2.1
- **组织**: alibaba / higress-group
- **语言**: Go 1.24.4
- **路径**: `/home/cloudyi/code/aigateway-group/higress`
- **上游**: Istio 1.27 + Envoy 1.36 (higress-group 定制分支)
- **Console 依赖版本**: higress-console v2.1.9

---

## 关键入口 & 常用路径

| 用途 | 路径 |
|---|---|
| 主程序入口 | `cmd/higress/main.go` |
| 核心包 | `pkg/` (bootstrap, cert, config, ingress, kube, common) |
| Wasm 插件 (Go) | `plugins/wasm-go/extensions/` (56 个插件) |
| MCP Server 插件 | `plugins/wasm-go/mcp-servers/` (55+ 个) |
| 插件 SDK | `plugins/wasm-go/pkg/` |
| 服务注册 | `registry/` (nacos, consul, eureka, zookeeper) |
| CLI 工具 | `hgctl/` |
| API 定义 | `api/` (protobuf, CRD) |
| Helm Chart | `helm/higress/` |
| Docker 构建 | `docker/` |
| Istio 子模块 | `istio/` → `external/` (symlinks) |
| Envoy 子模块 | `envoy/` → `external/` (symlinks) |
| 构建系统 | `Makefile` → `Makefile.core.mk` |

---

## Git 子模块 (7 个)

| 子模块 | 远程仓库 | 分支 |
|---|---|---|
| `istio/api` | higress-group/api | istio-1.27 |
| `istio/istio` | higress-group/istio | istio-1.27 |
| `istio/client-go` | higress-group/client-go | istio-1.27 |
| `istio/pkg` | higress-group/pkg | istio-1.19 |
| `istio/proxy` | higress-group/proxy | envoy-1.36 |
| `envoy/envoy` | higress-group/envoy | envoy-1.36 |
| `envoy/go-control-plane` | higress-group/go-control-plane | envoy-1.36 |

---

## 重要提醒 & 经验

### 编译相关
- 使用 `go.mod` 中大量 `replace` 指令将 istio/envoy 依赖指向 `external/` 下的本地 fork
- 子模块为 shallow clone，更新时使用 `git submodule update --init --recursive`
- 构建脚本: `docker/build-local-images.sh`（之前遇到过构建错误，参考对话 7973c91d）

### 插件开发
- Go Wasm 插件使用 `plugins/wasm-go/pkg/` 中的 SDK
- 支持 Go/Rust/C++/AssemblyScript 四种语言
- 插件编译使用 `plugins/wasm-go/Makefile`
- 插件 Dockerfile: `plugins/wasm-go/Dockerfile`

### 架构要点
- 控制面 = Higress Controller (基于 Istio pilot)
- 数据面 = Envoy Proxy (带 Wasm 扩展)
- 配置变更毫秒生效，无 reload 抖动
- 支持 Ingress API 和 Gateway API 双模
- 内置 ACME 自动证书管理 (Let's Encrypt)

### 相关项目路径
- higress-console 后端: `/home/cloudyi/code/aigateway-group/higress-console/backend`（参考对话 caf9d6fb）

---

## 端口约定

| 端口 | 用途 |
|---|---|
| 8001 | Higress UI 控制台 |
| 8080 | 网关 HTTP 入口 |
| 8443 | 网关 HTTPS 入口 |

---

## 待办 / 关注点

- [ ] 了解 `pkg/ingress/translation/` 中 Ingress→EnvoyFilter 的转换逻辑
- [ ] 深入分析 `ai-proxy` 插件如何对接各 LLM provider
- [ ] 研究 `mcp-server` 插件的 MCP 协议实现
- [ ] 跟踪 higress-console v2.1.9 → 最新版本的变化
