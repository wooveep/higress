# Higress — AI Native API Gateway

> **版本**: v2.2.1 | **语言**: Go 1.24 | **许可**: Apache 2.0  
> **上游基础**: Istio 1.27 + Envoy 1.36 (higress-group 定制分支)  
> **GitHub**: https://github.com/alibaba/higress

## 项目概述

Higress 是阿里巴巴开源的**云原生 AI API 网关**，基于 Istio 和 Envoy 构建。它同时服务于传统 API 网关场景（K8s Ingress / Gateway API / 微服务网关）和 **AI 网关**场景（LLM 代理、MCP Server 托管、Token 限流、AI 可观测性）。

### 核心使用场景

| 场景 | 说明 |
|---|---|
| **AI Gateway** | 统一协议对接所有 LLM 模型提供商，AI 可观测性、多模型负载均衡、Token 限流、缓存 |
| **MCP Server 托管** | 通过插件机制托管 MCP Server，让 AI Agent 调用各种工具和服务 |
| **K8s Ingress/Gateway** | 兼容 nginx ingress 注解，支持 Gateway API，路由变更毫秒生效 |
| **微服务网关** | 支持 Nacos / ZooKeeper / Consul / Eureka 等服务注册中心 |
| **安全网关** | WAF、key-auth / hmac-auth / jwt-auth / basic-auth / OIDC 等 |

---

## 目录结构

```
higress/
├── cmd/higress/          # 主程序入口 (main.go)
├── pkg/                  # 核心 Go 源码
│   ├── bootstrap/        #   启动引导逻辑
│   ├── cert/             #   证书管理 (Let's Encrypt 自动签发)
│   ├── config/           #   配置管理
│   ├── ingress/          #   Ingress 相关逻辑
│   │   ├── config/       #     Ingress 配置处理
│   │   ├── kube/         #     K8s Ingress 对接
│   │   ├── log/          #     日志
│   │   ├── mcp/          #     MCP (Mesh Configuration Protocol)
│   │   └── translation/  #     Ingress→Envoy 配置转换
│   ├── kube/             #   K8s 客户端封装
│   └── common/           #   通用工具函数
│
├── plugins/              # Wasm 插件生态
│   ├── wasm-go/          #   Go Wasm 插件 (主力)
│   │   ├── extensions/   #     56 个官方插件 (ai-proxy, waf, jwt-auth, mcp-server, ...)
│   │   ├── mcp-servers/  #     55+ MCP Server 插件 (github, weather, stock, ...)
│   │   └── pkg/          #     插件 SDK
│   ├── wasm-cpp/         #   C++ Wasm 插件
│   ├── wasm-rust/        #   Rust Wasm 插件
│   ├── wasm-assemblyscript/  # AssemblyScript Wasm 插件
│   └── golang-filter/    #   Golang 原生 filter
│
├── hgctl/                # Higress CLI 工具 (类似 istioctl)
│   ├── cmd/              #   CLI 命令定义
│   └── pkg/              #   CLI 核心逻辑
│
├── registry/             # 服务注册中心适配
│   ├── nacos/            #   Nacos v1 & v2
│   ├── consul/           #   Consul
│   ├── eureka/           #   Eureka
│   ├── zookeeper/        #   ZooKeeper
│   ├── direct/           #   直接指定地址
│   ├── memory/           #   内存缓存
│   └── reconcile/        #   服务发现 reconcile 逻辑
│
├── api/                  # API 定义 (protobuf, CRD)
│   ├── networking/       #   网络相关 API
│   ├── extensions/       #   扩展 API
│   └── kubernetes/       #   K8s CRD 定义
│
├── helm/                 # Helm Chart (K8s 部署)
│   └── higress/          #   Higress Helm Chart
│
├── docker/               # Docker 镜像构建
│   ├── Dockerfile.base   #   基础镜像
│   ├── Dockerfile.higress #  Higress 主镜像
│   └── build-local-images.sh  # 本地构建脚本
│
├── external/             # 外部依赖 (symlinks → istio/ 子模块)
│   ├── api → istio/api
│   ├── istio → istio/istio
│   ├── client-go → istio/client-go
│   ├── pkg → istio/pkg
│   ├── proxy → istio/proxy
│   ├── envoy → envoy/envoy
│   └── go-control-plane → envoy/go-control-plane
│
├── envoy/                # Envoy 定制 (git submodule)
│   ├── envoy/            #   higress-group/envoy (envoy-1.36 分支)
│   └── go-control-plane/ #   higress-group/go-control-plane
│
├── istio/                # Istio 定制 (git submodules)
│   ├── istio/            #   higress-group/istio (istio-1.27 分支)
│   ├── api/              #   higress-group/api
│   ├── client-go/        #   higress-group/client-go
│   ├── pkg/              #   higress-group/pkg
│   └── proxy/            #   higress-group/proxy
│
├── docs/                 # 文档
├── test/                 # 测试
├── tools/                # 构建工具脚本
├── samples/              # 示例配置
└── release-notes/        # 发布说明
```

---

## 技术架构

```
┌──────────────────────────────────────────────────────┐
│                     Client / AI Agent                │
└──────────────────────┬───────────────────────────────┘
                       │ HTTP/HTTPS/gRPC
┌──────────────────────▼───────────────────────────────┐
│                  Envoy Proxy (数据面)                   │
│  ┌─────────────┐ ┌──────────────┐ ┌───────────────┐  │
│  │ Wasm Plugins │ │ Golang Filter│ │ Envoy Filters │  │
│  │ (Go/Rust/JS) │ │              │ │ (C++ 原生)    │  │
│  └─────────────┘ └──────────────┘ └───────────────┘  │
└──────────────────────┬───────────────────────────────┘
                       │ xDS / MCP
┌──────────────────────▼───────────────────────────────┐
│              Higress Controller (控制面)                │
│  ┌───────────┐ ┌──────────┐ ┌──────────────────────┐ │
│  │ Ingress   │ │ Gateway  │ │ Service Discovery    │ │
│  │ Controller│ │ API Ctrl │ │ (Nacos/Consul/ZK...) │ │
│  └───────────┘ └──────────┘ └──────────────────────┘ │
│  ┌───────────┐ ┌──────────┐ ┌──────────────────────┐ │
│  │ Cert Mgr  │ │ Config   │ │ Translation (Ingress │ │
│  │ (ACME)    │ │ Store    │ │  → EnvoyFilter)      │ │
│  └───────────┘ └──────────┘ └──────────────────────┘ │
└──────────────────────────────────────────────────────┘
```

---

## 核心依赖

| 依赖 | 版本/分支 | 说明 |
|---|---|---|
| Istio | 1.27 (higress-group fork) | 控制面基础 |
| Envoy | 1.36 (higress-group fork) | 数据面代理 |
| Go | 1.24.4 | 编译语言 |
| K8s client-go | v0.34.1 | K8s API 交互 |
| Gateway API | v1.4.0 | K8s Gateway API 支持 |
| Nacos SDK | v2.3.2 | Nacos 服务发现 |
| Consul API | v1.32.0 | Consul 服务发现 |
| higress-console | v2.1.9 | Web 管理控制台 (独立仓库) |

---

## 构建与运行

### Docker 快速启动

```bash
mkdir higress && cd higress
docker run -d --rm --name higress-ai -v ${PWD}:/data \
    -p 8001:8001 -p 8080:8080 -p 8443:8443 \
    higress-registry.cn-hangzhou.cr.aliyuncs.com/higress/all-in-one:latest
```

### 本地编译

```bash
# 初始化子模块
git submodule update --init --recursive

# 本地编译
make build

# 构建 Docker 镜像
docker/build-local-images.sh
```

### Helm 部署 (K8s)

```bash
helm install higress -n higress-system higress.io/higress --create-namespace
```

---

## Wasm 插件分类 (56 个官方插件)

### AI 类
`ai-proxy` · `ai-agent` · `ai-cache` · `ai-history` · `ai-intent` · `ai-json-resp` · `ai-load-balancer` · `ai-prompt-decorator` · `ai-prompt-template` · `ai-quota` · `ai-rag` · `ai-search` · `ai-security-guard` · `ai-statistics` · `ai-token-ratelimit` · `ai-transformer` · `ai-image-reader`

### MCP 类
`mcp-server` · `mcp-router`

### 认证 & 安全
`basic-auth` · `jwt-auth` · `key-auth` · `hmac-auth-apisix` · `oidc` · `simple-jwt-auth` · `ext-auth` · `waf` · `bot-detect` · `ip-restriction` · `opa` · `replay-protection`

### 流量管理
`cors` · `request-block` · `request-validation` · `custom-response` · `traffic-editor` · `traffic-tag` · `transformer` · `cache-control` · `response-cache` · `cluster-key-rate-limit` · `model-router` · `model-mapper`

### 其他
`geo-ip` · `frontend-gray` · `de-graphql` · `http-call` · `log-request-response` · `sni-misdirect` · `api-workflow` · `jsonrpc-converter` · `chatgpt-proxy` · `gw-error-format`

---

## 关联仓库

| 仓库 | 说明 |
|---|---|
| [higress-console](https://github.com/higress-group/higress-console) | Web 管理控制台 |
| [higress-standalone](https://github.com/higress-group/higress-standalone) | 独立部署版本 |
| [wasm-go](https://github.com/higress-group/wasm-go) | Wasm Go Plugin SDK |
| [plugin-server](https://github.com/higress-group/plugin-server) | 插件服务器 |
| [openapi-to-mcpserver](https://github.com/higress-group/openapi-to-mcpserver) | OpenAPI → MCP Server 转换工具 |

---

## Envoy & Istio 定制化分析

Higress 并非直接使用上游 Istio/Envoy，而是维护了一套深度定制的 fork 体系。以下是具体的定制内容。

### 一、Istio 控制面定制 (istio-1.27 分支)

Higress **不使用标准 Istio Pilot**，而是实现了自己的轻量控制面 `Higress Controller` (`pkg/bootstrap/server.go`)，但内部复用了大量 Istio Pilot 库。

#### 1.1 自定义 xDS 资源生成器 (MCP 协议)

标准 Istio Pilot 通过 xDS 推送 Envoy 配置。Higress 替换了其中的资源生成器，使用 **Istio MCP (Mesh Configuration Protocol)** 封装输出：

| 资源类型 | 生成器 | 源文件 |
|---|---|---|
| `WasmPlugin` | `WasmPluginGenerator` | `pkg/ingress/mcp/generator.go` |
| `DestinationRule` | `DestinationRuleGenerator` | 同上 |
| `EnvoyFilter` | `EnvoyFilterGenerator` | 同上 |
| `Gateway` | `GatewayGenerator` | 同上 |
| `VirtualService` | `VirtualServiceGenerator` | 同上 |
| `ServiceEntry` | `ServiceEntryGenerator` | 同上 |
| 其他 | `FallbackGenerator` (返回空) | 同上 |

这些生成器将 Istio 配置对象序列化为 `mcp.Resource` 格式，然后通过 `anypb` 封装成标准 xDS `discovery.Resource` 下发。这使得 Higress Console 可以作为 MCP 客户端接收配置。

**关键区别**：
- 标准 Pilot 的 `ProxyNeedsPush` 有复杂的过滤逻辑；Higress 简化为 **始终推送** (`return req, true`)
- 支持通过 `config.Extra` 字段在 MCP Resource 中携带额外 JSON 数据（利用 protobuf unknown fields, field number 100）

#### 1.2 Ingress → Istio 配置转换引擎

这是 Higress 最核心的定制。位于 `pkg/ingress/config/ingress_config.go`（2131 行），实现了将 K8s Ingress 资源自动转换为 Istio 原生配置：

```
K8s Ingress → Higress IngressConfig → Istio Gateway
                                    → Istio VirtualService
                                    → Istio DestinationRule
                                    → Istio ServiceEntry
                                    → Istio EnvoyFilter
                                    → Istio WasmPlugin
```

**转换流程**：
1. **多集群 Ingress 收集**：支持从多个 K8s 集群的 IngressController 收集 Ingress 资源
2. **注解解析** (`annotations.AnnotationHandler`)：解析 nginx 风格注解 + Higress 扩展注解
3. **Gateway 转换**：每个 host 生成一个 Istio Gateway，支持 HTTPS 自动配置
4. **VirtualService 转换**：支持 Canary 灰度发布、加权路由、AppRoot 重定向
5. **EnvoyFilter 生成**：自动生成 BasicAuth、Http2Rpc、MCP SSE StatefulSession 等 EnvoyFilter
6. **ServiceEntry 生成**：从服务注册中心（Nacos/Consul/Eureka/ZK）自动生成 ServiceEntry
7. **WasmPlugin 管理**：通过 CRD `WasmPlugin` 管理 Wasm 插件配置
8. **模板处理**：支持在配置中引用 K8s Secret 值（`TemplateProcessor`）

#### 1.3 自定义 CRD (Custom Resource Definitions)

Higress 在标准 Istio CRD 基础上新增了以下 CRD：

| CRD | API 定义 | 用途 |
|---|---|---|
| `McpBridge` | `api/networking/v1/mcp_bridge.pb.go` | 定义服务注册中心连接（Nacos/Consul/Eureka/ZK 地址和认证） |
| `Http2Rpc` | `api/networking/v1/http_2_rpc.pb.go` | HTTP 到 Dubbo RPC 的协议转换映射 |
| `WasmPlugin` | `api/extensions/v1alpha1/wasmplugin.pb.go` | Wasm 插件配置（Higress 扩展版本） |

#### 1.4 多注册中心服务发现

Higress 在 Istio 的服务发现之上增加了对 **非 K8s 注册中心** 的支持，通过 `McpBridge` CRD 配置：

```
McpBridge CRD → RegistryReconciler → ServiceEntry (Istio 标准格式)
                                   → VirtualService / WasmPlugin
```

支持的注册中心：`Nacos v1/v2`、`Consul`、`Eureka`、`ZooKeeper`、`Direct (静态地址)`

#### 1.5 其他 Istio 定制

- **自动 HTTPS**：内置 ACME 客户端（`pkg/cert/`），自动从 Let's Encrypt 签发和续期证书
- **Gateway API 支持**：`pkg/ingress/kube/gateway/` 实现 K8s Gateway API 控制器
- **nginx 兼容注解**：`pkg/ingress/kube/annotations/` 兼容 nginx-ingress 注解
- **ConfigMap 驱动配置**：`pkg/ingress/kube/configmap/` 从 ConfigMap 生成 EnvoyFilter

---

### 二、Envoy 数据面定制 (envoy-1.36 分支)

#### 2.1 自定义 Golang Filter (原生 C 共享库)

Higress 使用 Envoy 的 **Golang HTTP Filter** 机制（`envoy/contrib/golang`），将 Go 代码编译为 `.so` 文件嵌入 Envoy，处理 MCP 协议：

| Filter 名称 | 源文件 | 功能 |
|---|---|---|
| `mcp-server` | `plugins/golang-filter/mcp-server/filter.go` | 处理 MCP Server 的 `message` 端点 (POST 请求)，将请求体传递给 SSEServer 并返回 local reply |
| `mcp-session` | `plugins/golang-filter/mcp-session/filter.go` | MCP 会话管理，支持 SSE/Streamable HTTP/REST 三种上游模式、Redis 会话持久化、路径重写、用户级配置、限流 |

**编译方式**：
```bash
# 编译为共享库
TARGET_ARCH=amd64 ./tools/hack/build-golang-filters.sh
# 产物: external/package/golang-filter_amd64.so (约 73MB)
```

**mcp-session Filter 核心能力**：
- **三种 MCP 上游模式**：`RestUpstream`（REST API）、`SSEUpstream`（SSE 事件流）、`StreamableUpstream`（流式 HTTP）
- **SSE 端点 URL 重写**：解析上游 SSE 事件中的 `endpoint` 事件数据，按路径规则重写
- **Redis Pub/Sub**：通过 Redis 发布/订阅实现 SSE 会话持久化
- **用户级 MCP Server 配置**：支持 per-user 的 MCP Server 配置（通过 `x-higress-mcpserver-config` header）
- **JSON-RPC 限流**：对 `tools/call` 请求进行限流控制

#### 2.2 自定义 Envoy/Proxy 构建

Higress 使用 `higress-group/proxy` (envoy-1.36 分支) 作为 Envoy 代码基础：

- **预编译二进制**：生产使用从 `higress-group/proxy` releases 下载预编译 Envoy 二进制
  ```
  ENVOY_PACKAGE_URL_PATTERN=https://github.com/higress-group/proxy/releases/download/v2.2.1/envoy-symbol-ARCH.tar.gz
  ```
- **自定义补丁**：构建时应用 `tools/hack/build-envoy.patch` 补丁
- **构建工具镜像**：使用专门的 build-tools-proxy 镜像编译

#### 2.3 自定义 go-control-plane

Higress 维护了 `higress-group/go-control-plane` (envoy-1.36 分支)，扩展了 Envoy 的 xDS 控制面 API 定义，用于支持自定义的 filter 配置等。

---

### 三、Fork 管理与构建关系

```
┌─────────────────────────────────────────────────────────────┐
│             Git Submodules (shallow clone)                    │
│                                                               │
│  istio/istio ─────→ external/istio ──→ go.mod replace       │
│  istio/api ───────→ external/api ────→ go.mod replace       │
│  istio/client-go ─→ external/client-go → go.mod replace     │
│  istio/pkg ───────→ external/pkg ────→ go.mod replace       │
│  istio/proxy ─────→ external/proxy ──→ Envoy 二进制构建      │
│  envoy/envoy ─────→ external/envoy ──→ Envoy 源码构建        │
│  envoy/go-control-plane → external/go-control-plane          │
│                              → go.mod replace                │
└─────────────────────────────────────────────────────────────┘

prebuild.sh: 将 istio/ 和 envoy/ 下的子模块复制到 external/ 目录
             （不是 symlink，而是 cp -RP + 修改 .git 指向）
```

**构建产物组成**：

| 镜像 | 构建方式 | 说明 |
|---|---|---|
| `higress/higress` | `make build` | Higress Controller (Go 编译) |
| `higress/pilot` | `build-istio-pilot.sh` | Fork 版 Istio Pilot (在 `external/istio` 中编译) |
| `higress/gateway` | `build-gateway` + `build-golang-filter` | Fork 版 Envoy + Golang Filter .so |

### 四、定制总结

| 维度 | 标准 Istio/Envoy | Higress 定制 |
|---|---|---|
| **控制面** | Istiod (Pilot + Citadel + Galley) | Higress Controller (轻量化，仅保留路由/配置下发) |
| **配置下发** | xDS (Envoy 原生格式) | MCP 协议封装 xDS (支持 Console 作为 MCP 客户端) |
| **配置来源** | K8s CRD + Istio API | K8s Ingress + Gateway API + 自定义 CRD (McpBridge/Http2Rpc/WasmPlugin) |
| **服务发现** | K8s 仅 | K8s + Nacos + Consul + Eureka + ZooKeeper + Direct |
| **数据面扩展** | C++ Filter | Wasm (Go/Rust/JS) + Golang Filter (.so) |
| **MCP 支持** | 无 | 原生 MCP Server/Session 处理 (Golang Filter) |
| **证书管理** | Citadel mTLS | ACME 自动签发 (Let's Encrypt) |
| **协议转换** | 无 | HTTP → Dubbo RPC (Http2Rpc) |
