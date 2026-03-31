---
title: AI 数据脱敏
keywords: [higress,ai data masking]
description: AI 数据脱敏插件配置参考
---

## 功能说明

  对请求中的敏感词拦截，以及对请求/返回中的敏感内容进行替换、还原

![image](https://img.alicdn.com/imgextra/i4/O1CN0156Wtko1T9JO0RiWow_!!6000000002339-0-tps-1314-638.jpg)

### 处理数据范围
  - openai协议：请求/返回对话内容
  - jsonpath：只处理指定字段
  - raw：整个请求/返回body

### 敏感词拦截
  - 默认只对请求内容进行敏感词拦截，返回预设错误信息
  - 支持系统内置敏感词库和自定义敏感词

### 敏感词替换
  - 将请求数据中出现的敏感词替换为脱敏字符串，传递给后端服务。可保证敏感数据不出域
  - 部分脱敏数据在后端服务返回后可进行还原
  - 自定义规则支持标准正则和grok规则，替换字符串支持变量替换

## 运行属性

插件执行阶段：`认证阶段`
插件执行优先级：`100`

推荐执行顺序：应在 `model-router` 之后执行，以便先从请求体提取 `model` 并补齐 `x-higress-llm-model`，同时仍早于统计与配额插件完成请求侧拦截。

## 配置字段

| 名称 | 数据类型 | 默认值 | 描述 |
| -------- | --------  | -------- | -------- |
|  deny_openai            | bool            | true  |  对openai协议进行拦截 |
|  deny_jsonpath          | string          |   []  |  对指定jsonpath拦截 |
|  deny_raw               | bool            | false |  对原始body拦截 |
|  system_deny            | bool            | false |  开启内置拦截规则  |
|  system_deny_words      | array of string |   -   |  Console 投影的系统词库；为空时回退到内置词库  |
|  deny_code              | int             | 200   |  拦截时http状态码   |
|  deny_message           | string          | 提问或回答中包含敏感词，已被屏蔽 |  拦截时ai返回消息   |
|  deny_raw_message       | string          | {"errmsg":"提问或回答中包含敏感词，已被屏蔽"} |  非openai拦截时返回内容   |
|  deny_content_type      | string          | application/json  |  非openai拦截时返回content_type头 |
|  deny_words             | array of string | []    |  自定义敏感词列表  |
|  deny_rules             | array           | []    |  自定义拦截规则，支持 `contains / exact / regex` 三种匹配方式  |
|  deny_rules.pattern     | string          |   -   |  规则内容；当 `match_type=regex` 时支持 grok 内置规则  |
|  deny_rules.match_type  | string          | contains |  匹配方式：`contains / exact / regex`  |
|  deny_rules.description | string          |   -   |  规则备注  |
|  deny_rules.priority    | int             | 0     |  规则优先级，值越大越先展示  |
|  deny_rules.enabled     | bool            | true  |  是否启用该规则  |
|  replace_roles          | array           |   -   |  自定义敏感词正则替换  |
|  replace_roles.regex    | string          |   -   |  规则正则(内置GROK规则) |
|  replace_roles.type     | [replace, hash] |   -   |  替换类型  |
|  replace_roles.restore  | bool            | false |  是否恢复  |
|  replace_roles.value    | string          |   -   |  替换值（支持正则变量）  |
|  replace_rules          | array           | []    |  标准化替换规则配置，和 `replace_roles` 向后兼容  |
|  replace_rules.pattern  | string          |   -   |  规则正则(支持内置GROK规则) |
|  replace_rules.replace_type | [replace, hash] | replace |  替换类型  |
|  replace_rules.replace_value | string      |   -   |  替换值（支持正则变量）  |
|  replace_rules.restore  | bool            | false |  是否恢复  |
|  replace_rules.description | string       |   -   |  规则备注  |
|  replace_rules.priority | int             | 0     |  规则优先级  |
|  replace_rules.enabled  | bool            | true  |  是否启用该规则  |
|  audit_sink             | object          |   -   |  审计上报目标配置  |
|  audit_sink.service_name | string         |   -   |  审计服务名  |
|  audit_sink.namespace   | string          |   -   |  审计服务命名空间  |
|  audit_sink.port        | int             |   -   |  审计服务端口  |
|  audit_sink.path        | string          |   -   |  审计接口路径  |
|  audit_sink.timeout_ms  | int             | 2000  |  审计请求超时时间  |

## 配置示例

```yaml
system_deny: false
deny_openai: true
deny_jsonpath:
  - "$.messages[*].content"
deny_raw: true
deny_code: 200
deny_message: "提问或回答中包含敏感词，已被屏蔽"
deny_raw_message: "{\"errmsg\":\"提问或回答中包含敏感词，已被屏蔽\"}"
deny_content_type: "application/json"
deny_words: 
  - "自定义敏感词1"
  - "自定义敏感词2"
deny_rules:
  - pattern: "spam"
    match_type: "contains"
    description: "通用敏感词拦截"
  - pattern: "exact phrase"
    match_type: "exact"
    description: "精确短语拦截"
  - pattern: "b[a@4]d[wW]o[rR]d"
    match_type: "regex"
    description: "变体敏感词拦截"
replace_roles:
  - regex: "%{MOBILE}"
    type: "replace"
    value: "****"
    # 手机号  13800138000 -> ****
  - regex: "%{EMAILLOCALPART}@%{HOSTNAME:domain}"
    type: "replace"
    restore: true
    value: "****@$domain"
    # 电子邮箱  admin@gmail.com -> ****@gmail.com
  - regex: "%{IP}"
    type: "replace"
    restore: true
    value: "***.***.***.***"
    # ip 192.168.0.1 -> ***.***.***.***
  - regex: "%{IDCARD}"
    type: "replace"
    value: "****"
    # 身份证号 110000000000000000 -> ****
  - regex: "sk-[0-9a-zA-Z]*"
    restore: true
    type: "hash"
    # hash sk-12345 -> 9cb495455da32f41567dab1d07f1973d
    # hash后的值提供给大模型，从大模型返回的数据中会将hash值还原为原始值
replace_rules:
  - pattern: "%{MOBILE}"
    replace_type: "replace"
    replace_value: "****"
  - pattern: "sk-[0-9a-zA-Z]*"
    replace_type: "hash"
    restore: true
audit_sink:
  service_name: "aigateway-console"
  namespace: "default"
  port: 8080
  path: "/v1/internal/ai/sensitive-block-events"
  timeout_ms: 2000
```

## 敏感词替换样例

### 用户请求内容

  请将 `curl http://172.20.5.14/api/openai/v1/chat/completions -H "Authorization: sk-12345" -H "Auth: test@gmail.com"` 改成post方式

### 处理后请求大模型内容

  `curl http://***.***.***.***/api/openai/v1/chat/completions -H "Authorization: 48a7e98a91d93896d8dac522c5853948" -H "Auth: ****@gmail.com"` 改成post方式

### 大模型返回内容

  您想要将一个 `curl` 的 GET 请求转换为 POST 请求，并且这个请求是向一个特定的 API 发送数据。下面是修改后的 `curl` 命令，以 POST 方式发送：

```sh
curl -X POST \
     -H "Authorization: 48a7e98a91d93896d8dac522c5853948" \
     -H "Auth: ****@gmail.com" \
     -H "Content-Type: application/json" \
     -d '{"key":"value"}' \
     http://***.***.***.***/api/openai/v1/chat/completions
```

这里做了如下几个修改:

- `-X POST` 设置请求方式为 POST。
- `-H "Content-Type: application/json"` 设置请求头中的 `Content-Type` 为 `application/json`，这通常用来告诉服务器您发送的数据格式是 JSON。
- `-d '{"key":"value"}'` 这里设置了要发送的数据，`'{"key":"value"}'` 是一个简单的 JSON 对象示例。您需要将其替换为您实际想要发送的数据。

请注意，您需要将 `"key":"value"` 替换为您实际要发送的数据内容。如果您的 API 接受不同的数据结构或者需要特定的字段，请根据实际情况调整这部分内容。

### 处理后返回用户内容

  您想要将一个 `curl` 的 GET 请求转换为 POST 请求，并且这个请求是向一个特定的 API 发送数据。下面是修改后的 `curl` 命令，以 POST 方式发送：

```sh
curl -X POST \
     -H "Authorization: sk-12345" \
     -H "Auth: test@gmail.com" \
     -H "Content-Type: application/json" \
     -d '{"key":"value"}' \
     http://172.20.5.14/api/openai/v1/chat/completions
```

这里做了如下几个修改:

- `-X POST` 设置请求方式为 POST。
- `-H "Content-Type: application/json"` 设置请求头中的 `Content-Type` 为 `application/json`，这通常用来告诉服务器您发送的数据格式是 JSON。
- `-d '{"key":"value"}'` 这里设置了要发送的数据，`'{"key":"value"}'` 是一个简单的 JSON 对象示例。您需要将其替换为您实际想要发送的数据。

请注意，您需要将 `"key":"value"` 替换为您实际要发送的数据内容。如果您的 API 接受不同的数据结构或者需要特定的字段，请根据实际情况调整这部分内容。


## 相关说明

 - 流模式中如果脱敏后的词被多个chunk拆分，可能无法进行还原
 - AI/OpenAI 响应默认不做敏感词拦截，仅保留脱敏内容恢复能力
 - grok 内置规则列表 https://help.aliyun.com/zh/sls/user-guide/grok-patterns
 - 内置敏感词库数据来源 https://github.com/houbb/sensitive-word-data/tree/main/src/main/resources
 - 系统词库中的短中文词条（3 个汉字及以内）默认按更保守的精确整句方式命中，避免误伤“天安门的景点”这类普通查询
 - 当前插件对拦截规则支持 `contains / exact / regex` 三种匹配方式，且默认大小写不敏感
 - `deny_words` 和 `replace_roles` 仍可继续使用，但建议优先使用 `deny_rules` 与 `replace_rules`
