---
title: AI Data Masking
keywords: [higress, ai data masking]
description: AI Data Masking Plugin Configuration Reference
---
## Function Description
  Sensitive-word interception for requests plus replacement and restoration for requests/responses
![image](https://img.alicdn.com/imgextra/i4/O1CN0156Wtko1T9JO0RiWow_!!6000000002339-0-tps-1314-638.jpg)

### Data Handling Scope
  - openai protocol: Request/response conversation content
  - jsonpath: Only process specified fields
  - raw: Entire request/response body

### Sensitive Word Interception
  - By default, only request content is intercepted for sensitive words and returns preset error messages
  - Supports system's built-in sensitive word library and custom sensitive words

### Sensitive Word Replacement
  - Replace sensitive words in request data with masked strings before passing to back-end services. Ensures that sensitive data does not leave the domain
  - Some masked data can be restored after being returned by the back-end service
  - Custom rules support standard regular expressions and grok rules, and replacement strings support variable substitution

## Execution Properties
Plugin Execution Phase: `Authentication Phase`  
Plugin Execution Priority: `100`

Recommended order: run after `model-router` so the request body is parsed and `x-higress-llm-model` is populated before request-side interception, while still remaining earlier than statistics and quota plugins.

## Configuration Fields
| Name                   | Data Type       | Default Value | Description                          |
| ---------------------- | ---------------- | -------------- | ------------------------------------ |
|  deny_openai           | bool             | true           |  Intercept openai protocol          |
|  deny_jsonpath         | string           |   []           |  Intercept specified jsonpath       |
|  deny_raw              | bool             | false          |  Intercept raw body                 |
|  system_deny           | bool             | false          |  Enable built-in interception rules  |
|  system_deny_words     | array of string  |   -            |  Console-projected system dictionary; falls back to the bundled dictionary when omitted |
|  deny_code             | int              | 200            |  HTTP status code when intercepted   |
|  deny_message          | string           | Sensitive words found in the question or answer have been blocked | AI returned message when intercepted |
|  deny_raw_message      | string           | {"errmsg":"Sensitive words found in the question or answer have been blocked"} | Content returned when not openai intercepted |
|  deny_content_type     | string           | application/json | Content type header returned when not openai intercepted |
|  deny_words            | array of string  | []             | Custom sensitive word list           |
|  deny_rules            | array            | []             | Structured detect rules with `contains / exact / regex` matching |
|  deny_rules.pattern    | string           |   -            | Rule pattern; grok patterns are supported for `regex` |
|  deny_rules.match_type | string           | contains       | Match type: `contains / exact / regex` |
|  deny_rules.description | string          |   -            | Rule description                     |
|  deny_rules.priority   | int              | 0              | Rule priority                        |
|  deny_rules.enabled    | bool             | true           | Whether the rule is enabled          |
|  replace_roles         | array            |   -            | Custom sensitive word regex replacement |
|  replace_roles.regex   | string           |   -            | Rule regex (built-in GROK rule)    |
|  replace_roles.type    | [replace, hash]  |   -            | Replacement type                     |
|  replace_roles.restore  | bool             | false          | Whether to restore                   |
|  replace_roles.value    | string          |   -            | Replacement value (supports regex variables) |
|  replace_rules         | array            | []             | Structured replace rules, backward compatible with `replace_roles` |
|  replace_rules.pattern | string           |   -            | Rule regex (supports built-in GROK patterns) |
|  replace_rules.replace_type | [replace, hash] | replace     | Replacement type                     |
|  replace_rules.replace_value | string     |   -            | Replacement value (supports regex variables) |
|  replace_rules.restore | bool             | false          | Whether to restore                   |
|  replace_rules.description | string       |   -            | Rule description                     |
|  replace_rules.priority | int             | 0              | Rule priority                        |
|  replace_rules.enabled | bool             | true           | Whether the rule is enabled          |
|  audit_sink            | object           |   -            | Audit reporting target configuration |
|  audit_sink.service_name | string         |   -            | Audit service name                   |
|  audit_sink.namespace  | string           |   -            | Audit service namespace              |
|  audit_sink.port       | int              |   -            | Audit service port                   |
|  audit_sink.path       | string           |   -            | Audit endpoint path                  |
|  audit_sink.timeout_ms | int              | 2000           | Audit request timeout                |

## Configuration Example
```yaml
system_deny: false
deny_openai: true
deny_jsonpath:
  - "$.messages[*].content"
deny_raw: true
deny_code: 200
deny_message: "Sensitive words found in the question or answer have been blocked"
deny_raw_message: "{\"errmsg\":\"Sensitive words found in the question or answer have been blocked\"}"
deny_content_type: "application/json"
deny_words:
  - "Custom sensitive word 1"
  - "Custom sensitive word 2"
deny_rules:
  - pattern: "spam"
    match_type: "contains"
    description: "Generic sensitive keyword block"
  - pattern: "exact phrase"
    match_type: "exact"
    description: "Exact phrase block"
  - pattern: "b[a@4]d[wW]o[rR]d"
    match_type: "regex"
    description: "Variant word block"
replace_roles:
  - regex: "%{MOBILE}"
    type: "replace"
    value: "****"
    # Mobile number  13800138000 -> ****
  - regex: "%{EMAILLOCALPART}@%{HOSTNAME:domain}"
    type: "replace"
    restore: true
    value: "****@$domain"
    # Email  admin@gmail.com -> ****@gmail.com
  - regex: "%{IP}"
    type: "replace"
    restore: true
    value: "***.***.***.***"
    # IP 192.168.0.1 -> ***.***.***.***
  - regex: "%{IDCARD}"
    type: "replace"
    value: "****"
    # ID card number 110000000000000000 -> ****
  - regex: "sk-[0-9a-zA-Z]*"
    restore: true
    type: "hash"
    # hash sk-12345 -> 9cb495455da32f41567dab1d07f1973d
    # The hashed value is provided to the large model, and the hash value will be restored to the original value from the data returned by the large model
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

## Sensitive Word Replacement Example
### User Request Content
  Please change `curl http://172.20.5.14/api/openai/v1/chat/completions -H "Authorization: sk-12345" -H "Auth: test@gmail.com"` to POST method

### Processed Request Large Model Content
  `curl http://***.***.***.***/api/openai/v1/chat/completions -H "Authorization: 48a7e98a91d93896d8dac522c5853948" -H "Auth: ****@gmail.com"` change to POST method

### Large Model Returned Content
  You want to convert a `curl` GET request to a POST request, and this request is sending data to a specific API. Below is the modified `curl` command to send as POST:
```sh
curl -X POST \
     -H "Authorization: 48a7e98a91d93896d8dac522c5853948" \
     -H "Auth: ****@gmail.com" \
     -H "Content-Type: application/json" \
     -d '{"key":"value"}' \
     http://***.***.***.***/api/openai/v1/chat/completions
```
Here are the following modifications made:
- `-X POST` sets the request method to POST.
- `-H "Content-Type: application/json"` sets the `Content-Type` in the request header to `application/json`, which is typically used to inform the server that the data you are sending is in JSON format.
- `-d '{"key":"value"}'` sets the data to be sent, where `'{"key":"value"}'` is a simple example of a JSON object. You need to replace it with the actual data you want to send.

Please note that you need to replace `"key":"value"` with the actual data content you want to send. If your API accepts a different data structure or requires specific fields, please adjust this part according to your actual situation.

### Processed Return to User Content
  You want to convert a `curl` GET request to a POST request, and this request is sending data to a specific API. Below is the modified `curl` command to send as POST:
```sh
curl -X POST \
     -H "Authorization: sk-12345" \
     -H "Auth: test@gmail.com" \
     -H "Content-Type: application/json" \
     -d '{"key":"value"}' \
     http://172.20.5.14/api/openai/v1/chat/completions
```
Here are the following modifications made:
- `-X POST` sets the request method to POST.
- `-H "Content-Type: application/json"` sets the `Content-Type` in the request header to `application/json`, which is typically used to inform the server that the data you are sending is in JSON format.
- `-d '{"key":"value"}'` sets the data to be sent, where `'{"key":"value"}'` is a simple example of a JSON object. You need to replace it with the actual data you want to send.

Please note that you need to replace `"key":"value"` with the actual data content you want to send. If your API accepts a different data structure or requires specific fields, please adjust this part according to your actual situation.

## Related Notes
 - In streaming mode, if the masked words are split across multiple chunks, restoration may not be possible
 - AI/OpenAI responses are not blocked by sensitive-word rules by default; only restore/masking behavior is preserved on the response side
 - Grok built-in rule list: https://help.aliyun.com/zh/sls/user-guide/grok-patterns
 - Built-in sensitive word library data source: https://github.com/houbb/sensitive-word-data/tree/main/src/main/resources
 - Short CJK entries in the system dictionary (up to 3 Han characters) use a more conservative exact whole-message match by default to reduce false positives for benign queries such as "天安门的景点"
 - Detect rules now support `contains / exact / regex` matching and are case-insensitive by default.
 - `deny_words` and `replace_roles` remain supported for backward compatibility, but `deny_rules` and `replace_rules` are recommended.
