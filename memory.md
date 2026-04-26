# Higress 项目 Memory

> 本文件只保留 Higress 子项目基本入口与实现 caveat。AIGateway 集成决策看根目录 `Memory.md`。

更新时间：2026-04-25

## 基本信息

- 名称：Higress，AI Native API Gateway。
- 当前仓库路径：`/home/cloudyi/code/aigateway-group/higress`。
- 主要语言：Go。
- 上游基础：Istio + Envoy 的 Higress 定制分支。
- 控制面：Higress Controller，基于 Istio pilot。
- 数据面：Envoy Proxy，带 Wasm 扩展。

## 关键路径

- 主程序入口：`cmd/higress/main.go`。
- 核心包：`pkg/`，覆盖 bootstrap、cert、config、ingress、kube、common。
- Go Wasm 插件：`plugins/wasm-go/extensions/`。
- MCP Server 插件：`plugins/wasm-go/mcp-servers/`。
- 插件 SDK：`plugins/wasm-go/pkg/`。
- 服务注册：`registry/`。
- CLI：`hgctl/`。
- API / CRD：`api/`。
- Helm Chart：`helm/higress/` 与根级父 Chart `../helm/higress`。
- Istio / Envoy 子模块：`istio/`、`envoy/` 指向本地 fork / symlink。

## 插件开发 caveat

- Go Wasm 插件使用 `plugins/wasm-go/pkg/` SDK。
- 支持 Go、Rust、C++、AssemblyScript 插件。
- Go Wasm 插件构建入口为 `plugins/wasm-go/Makefile`，容器构建见 `plugins/wasm-go/Dockerfile`。
- AI 相关重点插件：`ai-proxy`、`ai-quota`、`ai-token-ratelimit`、`ai-statistics`、`mcp-server`、`model-router`。

## 子模块与构建 caveat

- `go.mod` 中大量 `replace` 指向本地 Istio / Envoy fork。
- 子模块通常为 shallow clone，更新时使用 `git submodule update --init --recursive`。
- 构建脚本和 release 镜像由 AIGateway 根级脚本统一编排，优先从根目录入口执行。

## 端口约定

- `8001`：Higress UI 控制台历史端口。
- `8080`：网关 HTTP 入口历史端口。
- `8443`：网关 HTTPS 入口历史端口。
- AIGateway release / dev 的实际暴露方式以根级 Helm values 和 `Project.md` 为准。
