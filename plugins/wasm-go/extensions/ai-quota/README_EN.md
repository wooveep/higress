---
title: AI Quota Management
keywords: [ AI Gateway, AI Quota ]
description: AI quota management plugin configuration reference
---
## Function Description
The `ai-quota` plugin supports two operating modes:

- `token`: legacy mode that directly checks and deducts token quotas stored in Redis.
- `amount`: billing mode. The plugin checks wallet balance and model pricing before forwarding the request, then calculates the final charge from real `input/output` token usage and writes a usage event to Redis Stream for durable billing.

The plugin works together with authentication plugins such as `key-auth` and `jwt-auth` to identify the consumer. In `amount` mode it no longer relies on `ai-statistics` as the primary billing source.

## Runtime Properties
Plugin execution phase: `default phase`
Plugin execution priority: `750`

## Configuration Description
| Name | Data Type | Required Conditions | Default Value | Description |
| --- | --- | --- | --- | --- |
| `quota_unit` | string | Optional | `token` | Quota unit, either `token` or `amount` |
| `redis_key_prefix` | string | Optional | `chat_quota:` | Quota Redis key prefix in `token` mode |
| `balance_key_prefix` | string | Optional | `billing:balance:` | Wallet balance key prefix in `amount` mode |
| `price_key_prefix` | string | Optional | `billing:model-price:` | Model pricing hash prefix in `amount` mode |
| `usage_event_stream` | string | Optional | `billing:usage:stream` | Usage event stream in `amount` mode |
| `usage_event_dedup_prefix` | string | Optional | `billing:usage:event:` | Event deduplication key prefix in `amount` mode |
| `admin_consumer` | string | Required | - | Consumer name for quota administration |
| `admin_path` | string | Optional | `/quota` | Prefix for quota management paths |
| `redis` | object | Yes | - | Redis related configuration |
Explanation of each configuration field in `redis`
| Configuration Item | Type   | Required | Default Value                                           | Explanation                                                                                             |
|--------------------|--------|----------|---------------------------------------------------------|---------------------------------------------------------------------------------------------------------|
| service_name       | string | Required | -                                                       | Redis service name, full FQDN name with service type, e.g., my-redis.dns, redis.my-ns.svc.cluster.local |
| service_port       | int    | No       | Default value for static service is 80; others are 6379 | Service port for the redis service                                                                      |
| username           | string | No       | -                                                       | Redis username                                                                                          |
| password           | string | No       | -                                                       | Redis password                                                                                          |
| timeout            | int    | No       | 1000                                                    | Redis connection timeout in milliseconds                                                                |
| database           | int    | No       | 0                                                       | The database ID used, for example, configured as 1, corresponds to `SELECT 1`.                          |

## Configuration Example
### Amount mode
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

In `amount` mode:

- Request admission checks the consumer balance and model pricing.
- Missing pricing or non-positive balance will reject the request and emit an audit event.
- Successful responses are charged using the real token usage reported by the upstream model.
- The plugin writes a usage event into Redis Stream, and the backend billing consumer persists it into MySQL.

### Legacy token mode
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

### Refresh Quota
If the suffix of the current request URL matches the admin_path, for example, if the plugin is effective on the route example.com/v1/chat/completions, then the quota can be updated via:
curl https://example.com/v1/chat/completions/quota/refresh -H "Authorization: Bearer credential3" -d "consumer=consumer1&quota=10000"
In `token` mode, the value of `chat_quota:consumer1` in Redis will be refreshed to `10000`.

In `amount` mode, the same admin path writes the wallet balance key, which is usually projected from the billing control plane.

### Query Quota
To query the quota of a specific user, you can use: 
curl https://example.com/v1/chat/completions/quota?consumer=consumer1 -H "Authorization: Bearer credential3"
The response will return: {"quota": 10000, "consumer": "consumer1"}

### Increase or Decrease Quota
To increase or decrease the quota of a specific user, you can use:
curl https://example.com/v1/chat/completions/quota/delta -d "consumer=consumer1&value=100" -H "Authorization: Bearer credential3"
This will increase the value of the key `chat_quota:consumer1` in Redis by 100, and negative values can also be supported, thus subtracting the corresponding value.
