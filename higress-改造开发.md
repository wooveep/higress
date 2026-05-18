# Higress 改造开发分析

分析范围：

- Higress 项目：`/home/cloudyi/CodeWorkspace/higress`
- 控制台项目：`/home/cloudyi/CodeWorkspace/higress-console`
- 控制台 SDK graphify 输出：`/home/cloudyi/CodeWorkspace/higress-console/backend/sdk/graphify-out`

## 1. 控制台如何配置当前 Higress

控制台不是直接改 Higress 进程内存，而是通过 `higress-console/backend/sdk` 写 Kubernetes 资源，Higress Controller/Gateway 再 watch 这些资源并生成网关配置。

整体链路：

```text
frontend/src/services/*.ts
  -> backend/console/src/main/java/.../controller/*
  -> backend/sdk/src/main/java/.../service/*
  -> KubernetesClientService
  -> Kubernetes API Server
  -> Higress Controller/Gateway 生效
```

关键入口：

- Spring Bean 初始化：`higress-console/backend/console/src/main/java/com/alibaba/higress/console/config/SdkConfig.java`
- SDK 聚合工厂：`higress-console/backend/sdk/src/main/java/com/alibaba/higress/sdk/service/HigressServiceProviderImpl.java`
- Kubernetes 读写封装：`higress-console/backend/sdk/src/main/java/com/alibaba/higress/sdk/service/kubernetes/KubernetesClientService.java`
- 模型和 K8s 资源转换：`higress-console/backend/sdk/src/main/java/com/alibaba/higress/sdk/service/kubernetes/KubernetesModelConverter.java`

`SdkConfig` 从系统配置读取 kubeconfig、namespace、controller service、ingressClass、access token 等参数，创建 `HigressServiceProvider`。Provider 内部创建：

- `KubernetesClientService`
- `KubernetesModelConverter`
- `RouteServiceImpl`
- `DomainServiceImpl`
- `WasmPluginServiceImpl`
- `WasmPluginInstanceServiceImpl`
- `AiRouteServiceImpl`
- `LlmProviderServiceImpl`
- `McpServiceContextImpl`

控制台配置资源的主要落点：

| 控制台能力 | HTTP API | SDK Service | K8s 资源 |
| --- | --- | --- | --- |
| 普通路由 | `/v1/routes` | `RouteServiceImpl` | `Ingress` |
| 域名 | `/v1/domains` | `DomainServiceImpl` | `ConfigMap` + 路由关联 |
| 服务来源 | `/v1/service-sources` | `ServiceSourceServiceImpl` | `ConfigMap` / Gateway 相关配置 |
| Wasm 插件定义 | `/v1/wasm-plugins` | `WasmPluginServiceImpl` | `extensions.higress.io/v1alpha1/WasmPlugin` |
| Wasm 插件实例配置 | `/v1/*/plugin-instances` | `WasmPluginInstanceServiceImpl` | `WasmPlugin.spec.defaultConfig` / `matchRules` |
| AI Provider | `/v1/ai/providers` | `LlmProviderServiceImpl` | 服务来源 + 插件实例 |
| AI Route | `/v1/ai/routes` | `AiRouteServiceImpl` | `ConfigMap` + `Ingress` + 插件实例 |
| MCP Server | `/v1/mcpServer` | `McpServiceContextImpl` | `ConfigMap` + `Ingress` + `WasmPlugin` + `McpBridge` |

`KubernetesClientService` 中可见它直接读写：

- `Ingress`
- `ConfigMap`
- `Secret`
- `extensions.higress.io/v1alpha1/WasmPlugin`
- `networking.istio.io/v1alpha3/EnvoyFilter`
- `McpBridge`
- controller debug 接口：`/debug/registryz`、`/debug/endpointShardz`

## 2. Higress 暴露出来的接口在哪里

这里有两类“接口”需要区分。

### 2.1 控制台 HTTP API

控制台后端对前端和外部调用者暴露 Spring REST API，集中在：

`/home/cloudyi/CodeWorkspace/higress-console/backend/console/src/main/java/com/alibaba/higress/console/controller`

核心接口：

- `RoutesController.java`: `/v1/routes`
- `DomainsController.java`: `/v1/domains`
- `ServiceSourceController.java`: `/v1/service-sources`
- `WasmPluginsController.java`: `/v1/wasm-plugins`
- `WasmPluginInstancesController.java`: `/v1/global|domains|routes|services/.../plugin-instances`
- `ai/AiRoutesController.java`: `/v1/ai/routes`
- `ai/LlmProvidersController.java`: `/v1/ai/providers`
- `mcp/McpServerController.java`: `/v1/mcpServer`
- `ConsumersController.java`: `/v1/consumers`
- `TlsCertificatesController.java`: `/v1/tls-certificates`

前端调用位置：

- `frontend/src/services/route.ts`
- `frontend/src/services/plugin.ts`
- `frontend/src/services/ai-route.ts`
- `frontend/src/services/llm-provider.ts`
- `frontend/src/services/mcp.ts`

### 2.2 Higress 项目中的 MCP Server 工具接口

Higress 项目里 `plugins/golang-filter/mcp-server/servers/higress/higress-api` 暴露的是一个 MCP Server，不是普通 HTTP Controller。它通过 MCP tools 调用 Higress Console HTTP API。

入口：

- `plugins/golang-filter/mcp-server/servers/higress/higress-api/server.go`
- 配置说明：`plugins/golang-filter/mcp-server/servers/higress/higress-api/README.md`
- HTTP 客户端：`plugins/golang-filter/mcp-server/servers/higress/client.go`

`server.go` 注册 server 类型：

```text
common.GlobalRegistry.RegisterServer("higress-api", &HigressConfig{})
```

配置项中 `higressURL` 指向 Higress Console，例如：

```yaml
type: higress-api
config:
  higressURL: http://higress-console.higress-system.svc.cluster.local:8080
```

已注册的 MCP tools：

- 路由：`list-routes`、`get-route`、`add-route`、`update-route`、`delete-route`
- 服务来源：`list-service-sources`、`get-service-source`、`add-service-source`、`update-service-source`、`delete-service-source`
- AI 路由：`list-ai-routes`、`get-ai-route`、`add-ai-route`、`update-ai-route`、`delete-ai-route`
- AI Provider：`list-ai-providers`、`get-ai-provider`、`add-ai-provider`、`update-ai-provider`、`delete-ai-provider`
- MCP Server：`list-mcp-servers`、`get-mcp-server`、`add-or-update-mcp-server`、`delete-mcp-server`
- MCP consumers：`list-mcp-server-consumers`、`add-mcp-server-consumers`、`delete-mcp-server-consumers`
- Swagger 转 MCP：`swagger-to-mcp-config`
- 插件实例：`list-plugin-instances`、`get-plugin`、`delete-plugin`
- 特定插件快捷更新：`update-request-block-plugin`、`update-custom-response-plugin`

这些工具最终通过 `HigressClient` 转成 HTTP 调用，并透传 MCP 请求上下文中的 `Authorization` header。

## 3. `servers/gorm` 是否是插件对接外部数据库的依赖

结论：普通流量 Wasm 插件对接外部数据库，不依赖 `plugins/golang-filter/mcp-server/servers/gorm`。这个目录是 MCP Server 的 `database` 类型实现，用于把数据库作为 MCP 工具暴露给大模型或 MCP 客户端。

`servers/gorm` 的职责：

- 注册 server 类型：`common.GlobalRegistry.RegisterServer("database", &DBConfig{})`
- 支持数据库类型：MySQL、PostgreSQL、ClickHouse、SQLite
- 配置项：`dsn`、`dbType`、`description`
- 暴露 MCP tools：`query`、`execute`、`list tables`、`describe table`
- 使用 GORM 建立连接，并周期性重连

控制台 SDK 与它的关系：

- 控制台的 MCP Database 类型会在 `DatabaseSaveStrategy` 中把用户填写的 DB 配置转换为 DSN。
- 生成的配置写入 MCP Server ConfigMap。
- Higress MCP 运行时根据 server type `database` 加载 `servers/gorm`。

因此：

- 如果目标是“新增一个 MCP database server，让 MCP client 查询外部数据库”，应复用 `servers/gorm` 这条链路。
- 如果目标是“网关在线流量插件在请求/响应过程中做计费、扣费、记账”，不要直接依赖 `servers/gorm`。应在 Wasm 插件内使用 Higress Wasm SDK 支持的能力，例如 Redis、HTTP callout，或引入能在 proxy-wasm 环境中工作的外部调用方式。

## 4. 现有 `ai-quota` 和 `ai-statistics` 的能力边界

### 4.1 `ai-quota`

位置：

`plugins/wasm-go/extensions/ai-quota`

核心文件：

- `main.go`
- `plugin.yaml`
- `README.md`
- `main_test.go`

职责：

- 依赖认证插件写入的 `x-mse-consumer` 识别 consumer。
- 请求进入时从 Redis 读取 `redis_key_prefix + consumer`。
- 配额小于等于 0 时拒绝请求。
- 流式响应结束后通过 `tokenusage.GetTokenUsage` 获取 token 用量。
- 根据 input + output token 对 Redis 配额执行 `DecrBy`。
- 提供管理接口：
  - 查询：`/v1/chat/completions{admin_path}?consumer=xxx`
  - 刷新：`/v1/chat/completions{admin_path}/refresh`
  - 增减：`/v1/chat/completions{admin_path}/delta`

配置重点：

- `redis.service_name`
- `redis.service_port`
- `redis_key_prefix`
- `admin_consumer`
- `admin_path`
- `enable_path_suffixes`

### 4.2 `ai-statistics`

位置：

`plugins/wasm-go/extensions/ai-statistics`

职责：

- 采集 AI 请求/响应日志属性。
- 从流式或非流式响应提取 token usage。
- 写入用户属性和 span attribute。
- 通过 proxy-wasm counter metric 记录：
  - input token
  - output token
  - total token
  - first token duration
  - service duration
- 支持 OpenAI、Claude、Gemini 等格式的部分兼容。

它更偏观测和指标，不负责配额扣减或账单落库。

## 5. 新增计费能力的建议方案

建议新增独立插件，例如 `ai-billing`，不要直接把计费逻辑塞进 `ai-statistics`。原因：

- `ai-statistics` 是观测插件，职责是日志、指标、span，改成计费会让失败语义复杂化。
- 计费通常需要幂等、重试、补偿、账单状态，和纯指标不同。
- `ai-quota` 已经承担配额拦截和 Redis 扣减，可以作为“请求前准入 + 快速余额”的基础。

推荐职责拆分：

```text
认证插件 key-auth/jwt-auth
  -> 写入 x-mse-consumer

ai-quota
  -> 请求前校验余额/额度
  -> 响应结束后快速扣减 Redis quota

ai-statistics
  -> 记录日志、指标、span

新增 ai-billing
  -> 响应结束后生成计费事件
  -> 调用外部计费服务或写入可靠队列
  -> 支持幂等 key 和失败降级策略
```

### 5.1 ai-billing 插件建议行为

插件阶段建议放在默认阶段，优先级需要结合现有插件顺序设计：

- key-auth 示例优先级：`300`
- ai-statistics 示例优先级：`250`
- ai-quota README 写的是 `750`，但示例 `plugin.yaml` 中 priority 为 `280`，实际部署时要确认生效值。

建议：

- 认证插件必须先执行，确保 `x-mse-consumer` 已存在。
- `ai-quota` 应在真正转发前完成准入。
- `ai-billing` 可在响应结束时执行，不应阻塞主链路太久。
- `ai-statistics` 可继续独立记录指标。

计费事件建议字段：

```json
{
  "request_id": "...",
  "consumer": "...",
  "route": "...",
  "cluster": "...",
  "model": "...",
  "input_tokens": 0,
  "output_tokens": 0,
  "total_tokens": 0,
  "unit_price": "...",
  "amount": "...",
  "currency": "CNY",
  "timestamp": 0,
  "response_type": "stream|normal",
  "status": "success|failed"
}
```

幂等 key 可用：

```text
billing:{request_id}:{consumer}:{route}:{model}
```

如果 proxy-wasm 上下文能稳定拿到 request id，优先使用已有 request id；否则组合时间戳、consumer、route、模型和随机值。

### 5.2 数据库对接方式

不建议 Wasm 插件直接连 MySQL/PostgreSQL。更稳的方式：

1. 插件通过 HTTP callout 调用外部 billing-service。
2. billing-service 负责数据库事务、幂等、重试、价格表、账单状态。
3. 插件仅做最小事件上报，失败时可选择：
   - fail-open：不影响用户请求，但记录错误指标。
   - fail-close：计费失败则拒绝请求，适合强一致场景。

如果必须在网关侧做快速扣费，优先使用 Redis：

- `ai-quota` 已经使用 Redis。
- 可扩展为“扣余额 + 写流水队列”的模式。
- 真正数据库落库交给异步服务消费 Redis Stream / MQ。

`servers/gorm` 更适合 MCP 查询数据库，不适合高并发在线请求中的计费写路径。

### 5.3 基于现有插件改造的选择

方案 A：改造 `ai-quota`

- 适合：只需要按 token 扣减 Redis 额度，不需要正式账单流水。
- 修改点：
  - 扩展配置，增加价格规则或倍率。
  - 响应结束时按 token 和价格扣减。
  - 管理接口支持余额/套餐刷新。
- 风险：
  - 现有 quota 是整数 token 额度模型，金额计费会改变语义。
  - 管理接口和 Redis key 结构可能需要迁移。

方案 B：改造 `ai-statistics`

- 适合：只增加计费相关日志字段或指标。
- 不适合：真实扣费、账单落库、失败重试。
- 风险：
  - 统计插件当前设计是观测，计费失败不容易定义是否影响主请求。

方案 C：新增 `ai-billing`

- 适合：正式计费、账单、外部 DB、异步补偿。
- 推荐作为主方案。
- 可以复用：
  - `tokenusage.GetTokenUsage`
  - `x-mse-consumer`
  - route/cluster/model 获取逻辑，可参考 `ai-statistics`
  - Redis 接入方式，可参考 `ai-quota`
  - HTTP 响应/拒绝工具，可参考 `ai-quota/util`

## 6. 新增插件的落地路径

建议目录：

```text
plugins/wasm-go/extensions/ai-billing/
  go.mod
  go.sum
  main.go
  main_test.go
  README.md
  README_EN.md
  VERSION
  plugin.yaml
```

最小实现步骤：

1. 复制 `ai-quota` 或 `ai-statistics` 的工程结构。
2. `wrapper.SetCtx("ai-billing", ...)` 注册插件。
3. `parseConfig` 支持：
   - `billing_service_url`
   - `timeout`
   - `fail_policy`
   - `enable_path_suffixes`
   - `price_rules`
   - `redis` 可选，用于幂等或队列。
4. 请求头阶段读取：
   - `x-mse-consumer`
   - route name
   - cluster name
   - request path
5. 请求体阶段提取 model。
6. 响应体/流式响应结束阶段调用 `tokenusage.GetTokenUsage`。
7. 生成 billing event。
8. 通过 HTTP callout 或 Redis Stream 上报。
9. 增加单测覆盖：
   - 配置解析
   - path suffix 判断
   - token usage 事件构造
   - fail-open/fail-close
   - 幂等 key 生成

控制台侧如果要展示/配置该插件：

1. 确保插件镜像可发布到 OCI registry。
2. 在控制台 SDK 内置插件资源中加入 spec/readme/icon，或通过 `/v1/wasm-plugins` 创建自定义插件。
3. 通过 `/v1/global/plugin-instances/{name}`、`/v1/routes/{routeName}/plugin-instances/{name}` 等接口下发实例配置。
4. 如果需要专门 UI，可在 `frontend/src/services/plugin.ts` 复用现有插件实例接口。

## 7. 控制台和插件联动建议

短期可行方案：

- 使用现有控制台插件实例接口配置 `ai-quota`、`ai-statistics`、`ai-billing`。
- 先不改控制台后端模型，只把 `ai-billing` 作为普通 Wasm 插件配置。
- 自定义插件定义通过 `/v1/wasm-plugins` 或控制台内置插件资源加入。

中期增强：

- 在控制台新增“计费配置”页面。
- 后端新增 billing config DTO，但最终仍写入 WasmPlugin instance config。
- 增加 billing-service 地址、失败策略、价格规则、启用路径、消费者套餐等表单。

长期建议：

- 网关只做准入、采集和事件投递。
- 计费服务做价格、账单、账户、幂等、补偿。
- Redis 作为高性能配额缓存或事件队列。
- 数据库只由计费服务访问，不由 Wasm 插件直接访问。

## 8. 改造风险点

- `x-mse-consumer` 是现有配额和统计的关键身份来源，必须保证认证插件先执行。
- token usage 对不同模型供应商的响应格式敏感，应复用 `tokenusage.GetTokenUsage` 并补充测试样例。
- 流式响应只有 end-of-stream 时才能拿到完整用量，计费逻辑要避免重复扣费。
- `ai-quota` 当前 Redis 扣减是响应结束后做的，如果响应中断或解析不到 usage，可能不扣减。
- 在线请求路径不适合直接同步写数据库，容易引入延迟和故障扩散。
- 若 fail-open，账单可能漏记，需要异步补偿；若 fail-close，用户体验会受计费服务可用性影响。
- 控制台 `AiRoutesController.put` 中 name 判断逻辑看起来可疑：`StringUtils.isNotEmpty(route.getName())` 时直接 `route.setName(name)`，否则比较空值和 name。此处不影响本文档结论，但后续改 AI Route 时建议单独核查。

## 9. 独立部署与 Nacos 服务发现

独立部署时，Higress 仍然通过控制台配置服务来源；控制台把 Nacos 注册中心配置写入 Kubernetes 资源，Higress Controller/Gateway 侧 watch 这些资源后启动对应 registry watcher。

控制台侧入口：

- 前端：`higress-console/frontend/src/services/service-source.ts`
- 后端 API：`higress-console/backend/console/src/main/java/com/alibaba/higress/console/controller/ServiceSourceController.java`
- SDK 模型：`higress-console/backend/sdk/src/main/java/com/alibaba/higress/sdk/model/ServiceSource.java`
- SDK 写入：`ServiceSourceServiceImpl`
- 转换逻辑：`KubernetesModelConverter.fillServiceSourceInfo`、`addV1McpBridgeRegistry`、`fillV1RegistryConfig`

Higress 侧入口：

- `registry/reconcile/reconcile.go`
- `registry/nacos/v2/watcher.go`
- `registry/nacos/mcpserver/watcher.go`

配合方式：

```text
console 配置 Nacos service-source
  -> /v1/service-sources
  -> ServiceSourceServiceImpl 写 McpBridge/相关配置
  -> Higress Reconciler 读取 RegistryConfig
  -> generateWatcherFromRegistryConfig 创建 Nacos watcher
  -> Nacos watcher 发现服务实例
  -> 路由 upstream/service 名称解析到 Nacos 服务实例
```

`ServiceSource` 支持 `nacos`、`nacos2`、`nacos3`，Nacos 类型要求配置 `domain`、`port`、`properties.nacosGroups` 等字段。`nacos3` 还可启用 MCP Server 导出能力：

- `enableMCPServer`
- `mcpServerBaseUrl`
- `mcpServerExportDomains`

如果 console 也要作为 `billing-service` 对外提供服务，有两种部署方式：

- Kubernetes 内部服务：插件 HTTP callout 访问 `higress-console` 的 ClusterIP Service。
- 独立部署 + Nacos：把 console 或独立 billing-service 注册到 Nacos，再在 Higress console 中添加 Nacos service-source，给 billing-service 配一个内部路由或可解析 upstream。

推荐先把 billing-service 做成 console 后端模块，但按独立服务边界设计 API。这样早期少部署一个服务，后续也可以拆成独立进程并注册到 Nacos。

## 10. 三个 AI 插件的解耦协作设计

目标是每个插件都可以独立启用，符合自身功能，不形成强依赖：

| 插件 | 独立职责 | 可独立启用 | 外部契约 |
| --- | --- | --- | --- |
| `ai-statistics` | 观测：日志、指标、span、token 统计 | 是 | `tokenusage.GetTokenUsage`、用户属性、metric |
| `ai-quota` | 准入和额度限制：请求前检查余额，响应后扣减额度 | 是 | `x-mse-consumer`、Redis 余额、可选价格缓存 |
| `ai-billing` | 账单：生成计费事件并投递 billing-service | 是 | `x-mse-consumer`、HTTP callout、幂等 key |

不建议让 `ai-billing` 从 `ai-statistics` 读取 token。原因：

- 插件之间不应依赖同一请求上下文中的私有状态。
- `ai-statistics` 关闭时，计费仍应工作。
- `ai-statistics` 的职责是观测，不是计费事实来源。

推荐做法：

- `ai-statistics` 自己调用 `tokenusage.GetTokenUsage`，用于日志和指标。
- `ai-quota` 自己调用 `tokenusage.GetTokenUsage`，用于响应结束后扣减额度。
- `ai-billing` 自己调用 `tokenusage.GetTokenUsage`，用于生成账单事件。
- 三者共享同一个 tokenusage 解析库和字段约定，而不是共享插件运行时状态。

Redis 可以作为协作介质，但只承载外部状态，不承载插件间调用关系：

- `ai-quota` 使用 Redis 读写余额/额度。
- `ai-billing` 可用 Redis 做幂等、事件缓冲或失败重试队列。
- `ai-statistics` 不需要 Redis，除非后续增加自定义聚合。

## 11. 金额额度限制与模型价格映射

`ai-quota` 从 token 额度改造成金额额度时，需要把余额和价格分开看：

- 余额是强一致业务数据，最终以 console billing-service 的数据库为准。
- Redis 是网关侧快速判断和扣减缓存。
- 模型价格可以放 Redis，但只应作为缓存，不应作为唯一来源。

推荐数据归属：

| 数据 | 主存储 | Redis 作用 |
| --- | --- | --- |
| consumer 账户余额 | console billing-service DB | 热余额缓存 |
| 模型价格表 | console billing-service DB | 热价格缓存 |
| 计费流水 | console billing-service DB | 幂等 key / 异步事件队列 |
| 请求幂等状态 | console billing-service DB | 短 TTL 去重 |

建议 Redis key：

```text
billing:balance:{consumer}                       # 金额余额，单位建议用最小货币单位或 decimal 字符串
billing:price:{provider}:{model}:input           # 输入 token 单价
billing:price:{provider}:{model}:output          # 输出 token 单价
billing:price_version                            # 价格版本
billing:idempotency:{request_id}                 # 账单事件幂等
billing:event:stream                             # 可选，Redis Stream 事件队列
```

价格映射是否需要放 Redis：

- 需要缓存，但不要只放 Redis。
- console billing-service 管理价格表，并负责刷新 Redis。
- 插件读取 Redis 价格可以降低请求路径延迟。
- Redis 未命中时插件不应直接查数据库，建议调用 billing-service 的轻量接口或按配置选择 fail-open/fail-close。

金额额度扣减有两个阶段：

1. 请求前准入：只能判断余额是否大于 0，或根据 `max_tokens` 做预授权冻结。
2. 响应结束结算：拿到真实 token 后按价格计算金额，扣减实际费用。

如果业务要求严格不透支，应增加“预冻结”：

```text
请求前：
  estimate = input_tokens_estimate + max_tokens * output_price
  Redis/DB 冻结 estimate

响应结束：
  actual = input_tokens * input_price + output_tokens * output_price
  释放 estimate - actual，或补扣 actual - estimate
```

如果业务允许极小概率透支，早期可采用轻量模式：

```text
请求前：
  balance > 0 才放行

响应结束：
  根据真实 token 扣减金额
```

轻量模式改动最小、用户体验最好，但存在并发透支风险。预冻结模式更准确，但需要 billing-service 支持冻结/解冻事务。

## 12. `ai-billing` 与 console billing-service

推荐在 console 后端发布 billing-service 接口，插件通过 HTTP callout 调用。console billing-service 负责数据库事务、价格表、账户、幂等、账单流水和补偿。

建议接口：

```text
GET  /v1/billing/prices
GET  /v1/billing/prices/{provider}/{model}
GET  /v1/billing/accounts/{consumer}/balance
POST /v1/billing/reservations
POST /v1/billing/settlements
POST /v1/billing/events
GET  /v1/billing/events/{requestId}
```

最小可用版本可以只做：

```text
POST /v1/billing/events
```

请求体：

```json
{
  "request_id": "...",
  "idempotency_key": "...",
  "consumer": "...",
  "provider": "...",
  "model": "...",
  "route": "...",
  "cluster": "...",
  "input_tokens": 0,
  "output_tokens": 0,
  "total_tokens": 0,
  "stream": true,
  "status_code": 200,
  "started_at": 0,
  "ended_at": 0
}
```

billing-service 内部职责：

- 根据 provider/model 查价格。
- 计算费用。
- 用 `idempotency_key` 保证重复上报只结算一次。
- 写入账单流水。
- 更新余额或后付费账单。
- 刷新 Redis 余额和价格缓存。
- 提供管理 API 给 console UI。

插件调用策略：

- 默认 fail-open：billing-service 超时不阻断用户响应，只写错误日志和指标，可选写 Redis Stream 待补偿。
- 强管控场景可配置 fail-close，但不建议作为默认值。
- HTTP callout 必须有短超时，例如 100-500ms。
- 响应结束后异步上报，避免影响流式响应尾包返回。

## 13. 流式响应计费

流式响应要避免两个问题：

- 每个 chunk 都扣费导致重复计费。
- 等 billing-service 返回后再放行最后一个 chunk，影响用户体验。

推荐处理方式：

```text
onHttpStreamingResponseBody(data, endOfStream=false)
  -> 调用 tokenusage.GetTokenUsage(ctx, data)
  -> 如果解析到 usage，暂存 input/output/model 到 ctx
  -> 原样返回 data

onHttpStreamingResponseBody(data, endOfStream=true)
  -> 确认 ctx 中有最终 usage
  -> 生成 idempotency_key
  -> 投递 billing event
  -> 原样返回 data
```

准确性策略：

- 优先使用模型响应中的官方 usage 字段。
- 对 OpenAI/兼容 API，要求上游启用 stream usage，例如 `stream_options.include_usage=true`。
- 对无法提供 usage 的模型，采用 fallback：
  - 不计费，只记录 `usage_missing`。
  - 或调用外部 tokenizer 估算，但这会增加延迟和误差。
  - 或按请求固定价格计费。

推荐默认：

- 有官方 usage：按 token 精确计费。
- 无官方 usage：不在插件内估算金额，向 billing-service 上报 `usage_missing` 事件，由后端补偿或标记异常。

`ai-quota` 金额扣减也应使用同样策略：

- 有 usage 才执行真实扣减。
- 无 usage 不强行扣减，避免误扣。
- 可配置 `missing_usage_policy`：
  - `skip`
  - `fixed_fee`
  - `deny_next_request`

## 14. 最小改造方案

在不大范围修改现有逻辑的前提下，推荐分三步走。

第一步：新增 `ai-billing`

- 复制 `ai-statistics` 的 token 提取和流式处理思路。
- 不依赖 `ai-statistics`。
- 响应结束后调用 console billing-service。
- 默认 fail-open。

第二步：小改 `ai-quota`

- 保留现有 Redis quota 逻辑。
- 新增 `quota_mode`：
  - `token`: 兼容现有行为。
  - `amount`: 金额额度。
- `amount` 模式下读取 Redis 余额和模型价格缓存。
- 响应结束后按真实 usage 扣减金额。

第三步：console 增加 billing-service

- 先做后端 API 和 DB 表。
- 再做 Redis 价格/余额同步。
- 最后做 UI。

这样三个插件可以独立启用：

- 只开 `ai-statistics`：只有观测。
- 只开 `ai-quota`：只有额度限制。
- 只开 `ai-billing`：只有账单事件。
- 同时开启：通过 `x-mse-consumer`、Redis、billing-service 外部契约自然协作。

## 15. 推荐实施顺序

1. 明确计费模型：按 token、按请求、按模型、按 consumer、预付费还是后付费。
2. 保留 `ai-quota` 做额度准入。
3. 新增 `ai-billing` 插件，只负责生成计费事件并投递。
4. 建一个外部 billing-service 处理数据库落库。
5. 控制台先用现有 Wasm 插件实例接口下发配置。
6. 稳定后再做专门 UI 和后端 DTO。

最终建议：不要把 `servers/gorm` 当成在线 Wasm 插件对外部数据库的基础依赖；它是 MCP database server。计费能力应新增 `ai-billing` 插件，并把数据库事务放到外部 billing-service。
