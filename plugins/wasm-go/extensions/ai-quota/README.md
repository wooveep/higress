---
title: AI 配额管理
keywords: [ AI网关, AI配额 ]
description: AI 配额管理插件配置参考
---

## 功能说明

`ai-quota` 插件支持两种工作模式：

- `token`：兼容旧模式，直接对 Redis 中的 token 配额做准入和扣减。
- `amount`：金额模式。请求开始时校验用户余额和模型价格，请求结束后按真实 `input/output token` 计算金额扣减，并把消费事件写入 Redis Stream，供后端持久化到账本。

插件需要配合 `key-auth`、`jwt-auth` 等认证插件获取 consumer 名称；在金额模式下不再依赖 `ai-statistics` 作为主账务来源，真实消费会直接按请求写入事件流。

`amount` 模式下的日/周/月自然窗口统一按 `UTC` 计算：

- 日窗口在上海时区的次日 `00:00` 重置
- 周窗口以周一 `00:00` 为起点
- 月窗口以下月 1 日 `00:00` 为起点

## 运行属性

插件执行阶段：`默认阶段`
插件执行优先级：`750`

## 配置说明

| 名称 | 数据类型 | 填写要求 | 默认值 | 描述 |
| --- | --- | --- | --- | --- |
| `quota_unit` | string | 选填 | `token` | 配额单位，支持 `token` 或 `amount` |
| `redis_key_prefix` | string | 选填 | `chat_quota:` | `token` 模式下的 quota Redis key 前缀 |
| `balance_key_prefix` | string | 选填 | `billing:balance:` | `amount` 模式下的用户余额 Redis key 前缀 |
| `price_key_prefix` | string | 选填 | `billing:model-price:` | `amount` 模式下的模型价格 Redis Hash 前缀 |
| `usage_event_stream` | string | 选填 | `billing:usage:stream` | `amount` 模式下的消费事件 Stream |
| `usage_event_dedup_prefix` | string | 选填 | `billing:usage:event:` | `amount` 模式下的事件去重 key 前缀 |
| `admin_consumer` | string | 必填 | - | 管理 quota 接口身份的 consumer 名称 |
| `admin_path` | string | 选填 | `/quota` | 管理 quota 请求 path 前缀 |
| `redis` | object | 必填 | - | Redis 相关配置 |

`redis`中每一项的配置字段说明

| 配置项       | 类型   | 必填 | 默认值                                                     | 说明                                                                                         |
| ------------ | ------ | ---- | ---------------------------------------------------------- | ---------------------------                                                                  |
| service_name | string | 必填 | -                                                          | redis 服务名称，带服务类型的完整 FQDN 名称，例如 my-redis.dns、redis.my-ns.svc.cluster.local |
| service_port | int    | 否   | 服务类型为固定地址（static service）默认值为80，其他为6379 | 输入redis服务的服务端口                                                                      |
| username     | string | 否   | -                                                          | redis用户名                                                                                  |
| password     | string | 否   | -                                                          | redis密码                                                                                    |
| timeout      | int    | 否   | 1000                                                       | redis连接超时时间，单位毫秒                                                                  |
| database     | int    | 否   | 0                                                          | 使用的数据库id，例如配置为1，对应`SELECT 1`                                                  |


## 配置示例

### 金额模式
```yaml
quota_unit: amount
balance_key_prefix: "billing:balance:"
price_key_prefix: "billing:model-price:"
usage_event_stream: "billing:usage:stream"
admin_consumer: consumer3
admin_path: /quota
redis:
  service_name: redis-service.default.svc.cluster.local
  service_port: 6379
  timeout: 2000
```

金额模式下：

- 请求开始时读取 `x-mse-consumer`、请求模型和用户余额。
- 若模型没有有效价格或余额小于等于 0，会直接拒绝，并写入审计事件。
- 请求成功结束后会按真实 token 用量计算 `micro_yuan` 扣减金额，并写入 Redis Stream。
- 后端 billing consumer 再从 Stream 中落库到 MySQL 账本。
- 日/周/月金额窗口的 TTL 统一按 `UTC` 自然日历推导，而不是按部署机器本地时区。

### 金额模式计费口径

`amount` 模式会优先读取 `price_key_prefix + <provider>/<model>` 对应的 Redis Hash。该价格快照由 Portal 预先物化，已经包含 `ModelPriceData` 的回退结果，因此运行态不会再做浮点回退。

当前会写入账务事件并参与扣费/聚合的 usage 维度包括：

- `input_tokens`
- `output_tokens`
- `cache_creation_input_tokens`
- `cache_creation_5m_input_tokens`
- `cache_creation_1h_input_tokens`
- `cache_read_input_tokens`
- `input_image_tokens`
- `output_image_tokens`
- `input_image_count`
- `output_image_count`
- `request_count`

金额计算会同时覆盖：

- 文本输入/输出 token 单价
- 缓存创建、缓存读取、`above_200k` 分级 token 单价
- 输入/输出图像 token 单价
- `input_cost_per_request`
- `input_cost_per_image` / `output_cost_per_image`

当上游只返回聚合 `cache_creation_input_tokens` 时，插件会优先使用请求侧 `cache_control.ttl` 或等价 `ttl` 把剩余 token 分配到 `5m/1h` 桶；若未提供 TTL，则默认归入 `5m`。

### 兼容 token 模式
```yaml
quota_unit: token
redis_key_prefix: "chat_quota:"
admin_consumer: consumer3
admin_path: /quota
redis:
  service_name: redis-service.default.svc.cluster.local
  service_port: 6379
  timeout: 2000
```


###  刷新 quota

如果当前请求 url 的后缀符合 admin_path，例如插件在 example.com/v1/chat/completions 这个路由上生效，那么更新 quota 可以通过
curl https://example.com/v1/chat/completions/quota/refresh -H "Authorization: Bearer credential3" -d "consumer=consumer1&quota=10000" 

在 `token` 模式下，Redis 中 key 为 `chat_quota:consumer1` 的值会被刷新为 `10000`。

在 `amount` 模式下，管理接口仍然直接写 `balance_key_prefix + consumer`，通常由控制面或后端投影器统一维护。

### 查询 quota

查询特定用户的 quota 可以通过 curl https://example.com/v1/chat/completions/quota?consumer=consumer1 -H "Authorization: Bearer credential3"
将返回： {"quota": 10000, "consumer": "consumer1"}

### 增减 quota 

增减特定用户的 quota 可以通过 curl https://example.com/v1/chat/completions/quota/delta -d "consumer=consumer1&value=100" -H "Authorization: Bearer credential3"
这样 Redis 中 Key 为 chat_quota:consumer1 的值就会增加100，可以支持负数，则减去对应值。
