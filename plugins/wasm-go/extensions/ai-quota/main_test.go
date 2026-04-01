// Copyright (c) 2024 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/higress-group/proxy-wasm-go-sdk/proxywasm/types"
	"github.com/higress-group/wasm-go/pkg/test"
	"github.com/stretchr/testify/require"
)

// 测试配置：基础配置
var basicConfig = func() json.RawMessage {
	data, _ := json.Marshal(map[string]interface{}{
		"admin_consumer":   "admin",
		"redis_key_prefix": "chat_quota:",
		"admin_path":       "/quota",
		"redis": map[string]interface{}{
			"service_name": "redis.static",
			"service_port": 6379,
			"timeout":      1000,
			"database":     0,
		},
	})
	return data
}()

var amountConfig = func() json.RawMessage {
	data, _ := json.Marshal(map[string]interface{}{
		"admin_consumer":     "admin",
		"quota_unit":         "amount",
		"balance_key_prefix": "billing:balance:",
		"price_key_prefix":   "billing:model-price:",
		"usage_event_stream": "billing:usage:stream",
		"redis": map[string]interface{}{
			"service_name": "redis.static",
			"service_port": 6379,
			"timeout":      1000,
			"database":     0,
		},
	})
	return data
}()

// 测试配置：缺少admin_consumer
var missingAdminConsumerConfig = func() json.RawMessage {
	data, _ := json.Marshal(map[string]interface{}{
		"redis": map[string]interface{}{
			"service_name": "redis.static",
			"service_port": 6379,
		},
	})
	return data
}()

func TestParseConfig(t *testing.T) {
	test.RunGoTest(t, func(t *testing.T) {
		// 测试基础配置解析
		t.Run("basic config", func(t *testing.T) {
			host, status := test.NewTestHost(basicConfig)
			defer host.Reset()
			require.Equal(t, types.OnPluginStartStatusOK, status)
			config, err := host.GetMatchConfig()
			require.NoError(t, err)
			require.NotNil(t, config)

			quotaConfig := config.(*QuotaConfig)
			require.Equal(t, "admin", quotaConfig.AdminConsumer)
			require.Equal(t, "chat_quota:", quotaConfig.RedisKeyPrefix)
			require.Equal(t, "/quota", quotaConfig.AdminPath)
		})

		// 测试缺少admin_consumer的配置
		t.Run("missing admin_consumer", func(t *testing.T) {
			host, status := test.NewTestHost(missingAdminConsumerConfig)
			defer host.Reset()
			require.Equal(t, types.OnPluginStartStatusFailed, status)
		})

		t.Run("amount config", func(t *testing.T) {
			host, status := test.NewTestHost(amountConfig)
			defer host.Reset()
			require.Equal(t, types.OnPluginStartStatusOK, status)
			config, err := host.GetMatchConfig()
			require.NoError(t, err)
			quotaConfig := config.(*QuotaConfig)
			require.Equal(t, QuotaUnitAmount, quotaConfig.QuotaUnit)
			require.Equal(t, "billing:balance:", quotaConfig.BalanceKeyPrefix)
			require.Equal(t, "billing:model-price:", quotaConfig.PriceKeyPrefix)
		})
	})
}

func TestOnHttpRequestHeaders(t *testing.T) {
	test.RunTest(t, func(t *testing.T) {
		// 测试聊天完成模式的请求头处理
		t.Run("chat completion mode", func(t *testing.T) {
			host, status := test.NewTestHost(basicConfig)
			defer host.Reset()
			require.Equal(t, types.OnPluginStartStatusOK, status)

			// 设置请求头，包含consumer信息
			action := host.CallOnHttpRequestHeaders([][2]string{
				{":authority", "example.com"},
				{":path", "/v1/chat/completions"},
				{":method", "POST"},
				{"x-mse-consumer", "consumer1"},
			})

			// 由于需要调用Redis检查配额，应该返回HeaderStopAllIterationAndWatermark
			require.Equal(t, types.HeaderStopAllIterationAndWatermark, action)

			// 模拟Redis调用响应（有足够配额）
			resp := test.CreateRedisResp(1000)
			host.CallOnRedisCall(0, resp)
			action = host.GetHttpStreamAction()
			require.Equal(t, types.ActionContinue, action)
			host.CompleteHttp()
		})

		// 测试管理员查询模式的请求头处理
		t.Run("admin query mode", func(t *testing.T) {
			host, status := test.NewTestHost(basicConfig)
			defer host.Reset()
			require.Equal(t, types.OnPluginStartStatusOK, status)

			// 设置请求头，包含admin consumer信息
			action := host.CallOnHttpRequestHeaders([][2]string{
				{":authority", "example.com"},
				{":path", "/v1/chat/completions/quota?consumer=consumer1"},
				{":method", "GET"},
				{"x-mse-consumer", "admin"},
			})

			// 管理员查询模式应该返回 ActionPause
			require.Equal(t, types.ActionPause, action)

			// 模拟Redis调用响应
			resp := test.CreateRedisResp(500)
			host.CallOnRedisCall(0, resp)

			response := host.GetLocalResponse()
			require.Equal(t, uint32(http.StatusOK), response.StatusCode)
			require.Equal(t, "{\"consumer\":\"consumer1\",\"quota\":500}", string(response.Data))
			host.CompleteHttp()
		})

		// 测试无consumer的情况
		t.Run("no consumer", func(t *testing.T) {
			host, status := test.NewTestHost(basicConfig)
			defer host.Reset()
			require.Equal(t, types.OnPluginStartStatusOK, status)

			// 设置请求头，不包含consumer信息
			action := host.CallOnHttpRequestHeaders([][2]string{
				{":authority", "example.com"},
				{":path", "/v1/chat/completions"},
				{":method", "POST"},
			})

			// 无consumer应该返回ActionContinue
			require.Equal(t, types.ActionContinue, action)
		})

		t.Run("amount mode should buffer request body", func(t *testing.T) {
			host, status := test.NewTestHost(amountConfig)
			defer host.Reset()
			require.Equal(t, types.OnPluginStartStatusOK, status)

			action := host.CallOnHttpRequestHeaders([][2]string{
				{":authority", "example.com"},
				{":path", "/v1/chat/completions"},
				{":method", "POST"},
				{"x-mse-consumer", "consumer1"},
			})

			require.Equal(t, types.HeaderStopIteration, action)
		})
	})
}

func TestOnHttpRequestBody(t *testing.T) {
	test.RunTest(t, func(t *testing.T) {
		// 测试管理员刷新模式的请求体处理
		t.Run("admin refresh mode", func(t *testing.T) {
			host, status := test.NewTestHost(basicConfig)
			defer host.Reset()
			require.Equal(t, types.OnPluginStartStatusOK, status)

			// 先设置请求头
			host.CallOnHttpRequestHeaders([][2]string{
				{":authority", "example.com"},
				{":path", "/v1/chat/completions/quota/refresh"},
				{":method", "POST"},
				{"x-mse-consumer", "admin"},
			})

			// 设置请求体
			body := "consumer=consumer1&quota=1000"
			action := host.CallOnHttpRequestBody([]byte(body))

			// 管理员刷新模式应该返回ActionPause
			require.Equal(t, types.ActionPause, action)

			// 模拟Redis调用响应
			resp := test.CreateRedisRespArray([]interface{}{"OK"})
			host.CallOnRedisCall(0, resp)

			response := host.GetLocalResponse()
			require.Equal(t, uint32(http.StatusOK), response.StatusCode)
			require.Equal(t, "refresh quota successful", string(response.Data))
			host.CompleteHttp()
		})

		// 测试聊天完成模式的请求体处理
		t.Run("chat completion mode", func(t *testing.T) {
			host, status := test.NewTestHost(basicConfig)
			defer host.Reset()
			require.Equal(t, types.OnPluginStartStatusOK, status)

			// 先设置请求头
			host.CallOnHttpRequestHeaders([][2]string{
				{":authority", "example.com"},
				{":path", "/v1/chat/completions"},
				{":method", "POST"},
				{"x-mse-consumer", "consumer1"},
			})

			// 设置请求体
			body := `{"model": "gpt-3.5-turbo", "messages": [{"role": "user", "content": "Hello"}]}`
			action := host.CallOnHttpRequestBody([]byte(body))

			// 聊天完成模式应该返回ActionContinue
			require.Equal(t, types.ActionContinue, action)
		})

		t.Run("amount mode", func(t *testing.T) {
			host, status := test.NewTestHost(amountConfig)
			defer host.Reset()
			require.Equal(t, types.OnPluginStartStatusOK, status)

			action := host.CallOnHttpRequestHeaders([][2]string{
				{":authority", "example.com"},
				{":path", "/v1/chat/completions"},
				{":method", "POST"},
				{"x-mse-consumer", "consumer1"},
			})
			require.Equal(t, types.HeaderStopIteration, action)

			action = host.CallOnHttpRequestBody([]byte(`{"model":"qwen-plus","messages":[{"role":"user","content":"hello"}]}`))
			require.Equal(t, types.ActionPause, action)

			host.CallOnRedisCall(0, test.CreateRedisRespArray([]interface{}{
				"model_id", "qwen-plus",
				"price_version_id", "1",
				"input_price_per_1k_micro_yuan", "1000",
				"output_price_per_1k_micro_yuan", "2000",
			}))
			host.CallOnRedisCall(0, test.CreateRedisRespArray([]interface{}{1, "ok", 1000000}))

			action = host.GetHttpStreamAction()
			require.Equal(t, types.ActionContinue, action)
			host.CompleteHttp()
		})

		t.Run("amount mode denied by user daily window", func(t *testing.T) {
			host, status := test.NewTestHost(amountConfig)
			defer host.Reset()
			require.Equal(t, types.OnPluginStartStatusOK, status)

			action := host.CallOnHttpRequestHeaders([][2]string{
				{":authority", "example.com"},
				{":path", "/v1/chat/completions"},
				{":method", "POST"},
				{"x-mse-consumer", "consumer1"},
				{"x-higress-api-key-id", "KEY123"},
			})
			require.Equal(t, types.HeaderStopIteration, action)

			action = host.CallOnHttpRequestBody([]byte(`{"model":"qwen-plus","messages":[{"role":"user","content":"hello"}]}`))
			require.Equal(t, types.ActionPause, action)

			host.CallOnRedisCall(0, test.CreateRedisRespArray([]interface{}{
				"model_id", "qwen-plus",
				"price_version_id", "1",
				"input_price_per_1k_micro_yuan", "1000",
				"output_price_per_1k_micro_yuan", "2000",
			}))
			host.CallOnRedisCall(0, test.CreateRedisRespArray([]interface{}{0, "user_daily_limit_exceeded", 1000000}))

			response := host.GetLocalResponse()
			require.NotNil(t, response)
			require.Equal(t, uint32(http.StatusForbidden), response.StatusCode)
			require.Contains(t, string(response.Data), "User daily quota exceeded")
			host.CompleteHttp()
		})
	})
}

func TestOnHttpStreamingResponseBody(t *testing.T) {
	test.RunTest(t, func(t *testing.T) {
		// 测试聊天完成模式的流式响应体处理
		t.Run("chat completion mode", func(t *testing.T) {
			host, status := test.NewTestHost(basicConfig)
			defer host.Reset()
			require.Equal(t, types.OnPluginStartStatusOK, status)

			// 先设置请求头
			host.CallOnHttpRequestHeaders([][2]string{
				{":authority", "example.com"},
				{":path", "/v1/chat/completions"},
				{":method", "POST"},
				{"x-mse-consumer", "consumer1"},
			})

			// 测试流式响应体处理
			data := []byte(`{"choices": [{"delta": {"content": "Hello"}}]}`)
			action := host.CallOnHttpStreamingResponseBody(data, false)

			require.Equal(t, types.ActionContinue, action)
			result := host.GetResponseBody()
			// 非结束流应该返回原始数据
			require.Equal(t, data, result)

			// 测试结束流
			action = host.CallOnHttpStreamingResponseBody(data, true)

			require.Equal(t, types.ActionContinue, action)
			result = host.GetResponseBody()
			// 结束流应该返回原始数据
			require.Equal(t, data, result)

			// 模拟Redis调用响应（减少配额）
			resp := test.CreateRedisRespArray([]interface{}{30})
			host.CallOnRedisCall(0, resp)

			host.CompleteHttp()
		})

		// 测试非聊天完成模式的流式响应体处理
		t.Run("non-chat completion mode", func(t *testing.T) {
			host, status := test.NewTestHost(basicConfig)
			defer host.Reset()
			require.Equal(t, types.OnPluginStartStatusOK, status)

			// 先设置请求头
			host.CallOnHttpRequestHeaders([][2]string{
				{":authority", "example.com"},
				{":path", "/other/path"},
				{":method", "GET"},
				{"x-mse-consumer", "consumer1"},
			})

			// 测试流式响应体处理
			data := []byte("response data")
			action := host.CallOnHttpStreamingResponseBody(data, false)

			// 非聊天完成模式应该返回原始数据
			require.Equal(t, types.ActionContinue, action)
			result := host.GetResponseBody()
			require.Equal(t, data, result)
		})
	})
}

func TestGetOperationMode(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		adminPath string
		chatMode  ChatMode
		adminMode AdminMode
	}{
		{
			name:      "chat completion mode",
			path:      "/v1/chat/completions",
			adminPath: "/quota",
			chatMode:  ChatModeCompletion,
			adminMode: AdminModeNone,
		},
		{
			name:      "admin query mode",
			path:      "/v1/chat/completions/quota",
			adminPath: "/quota",
			chatMode:  ChatModeAdmin,
			adminMode: AdminModeQuery,
		},
		{
			name:      "admin refresh mode",
			path:      "/v1/chat/completions/quota/refresh",
			adminPath: "/quota",
			chatMode:  ChatModeAdmin,
			adminMode: AdminModeRefresh,
		},
		{
			name:      "admin delta mode",
			path:      "/v1/chat/completions/quota/delta",
			adminPath: "/quota",
			chatMode:  ChatModeAdmin,
			adminMode: AdminModeDelta,
		},
		{
			name:      "none mode",
			path:      "/other/path",
			adminPath: "/quota",
			chatMode:  ChatModeNone,
			adminMode: AdminModeNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chatMode, adminMode := getOperationMode(tt.path, tt.adminPath)
			require.Equal(t, tt.chatMode, chatMode)
			require.Equal(t, tt.adminMode, adminMode)
		})
	}
}

func TestBuildAmountChargeArgs(t *testing.T) {
	startedAt := time.Date(2026, time.March, 25, 10, 0, 0, 0, time.UTC)
	finishedAt := time.Date(2026, time.March, 25, 10, 29, 59, 0, time.UTC)
	occurredAt := time.Date(2026, time.March, 25, 10, 30, 0, 0, time.UTC)
	args := buildAmountChargeArgs(
		308,
		"evt-1",
		"req-1",
		"trace-1",
		"consumer1",
		"route-a",
		"/v1/chat/completions",
		"chat.completions",
		"qwen-plus",
		"success",
		"parsed",
		200,
		"",
		"",
		120,
		130,
		250,
		30,
		20,
		10,
		5,
		12,
		18,
		2,
		1,
		1,
		"1h",
		`{"cached_tokens":5}`,
		`{"image_tokens":18}`,
		`{"cache_ttl":"1h"}`,
		9,
		"KEY123",
		startedAt,
		finishedAt,
		occurredAt,
		3600,
		7200,
		10800,
	)

	require.Len(t, args, 38)
	require.Equal(t, int64(308), args[0])
	require.Equal(t, "trace-1", args[3])
	require.Equal(t, "success", args[9])
	require.Equal(t, int64(30), args[17])
	require.Equal(t, "1h", args[26])
	require.Equal(t, `{"cache_ttl":"1h"}`, args[29])
	require.Equal(t, "KEY123", args[31])
	require.Equal(t, startedAt.Format(time.RFC3339Nano), args[32])
	require.Equal(t, finishedAt.Format(time.RFC3339Nano), args[33])
	require.Equal(t, occurredAt.Format(time.RFC3339Nano), args[34])
	require.Equal(t, int64(3600), args[35])
	require.Equal(t, int64(7200), args[36])
	require.Equal(t, int64(10800), args[37])
}

func TestPreferredBillingModelName(t *testing.T) {
	require.Equal(t, "qwen-plus", preferredBillingModelName("qwen-plus", "qwen-plus-2026-04-01"))
	require.Equal(t, "gemini-2.5-pro", preferredBillingModelName("", "gemini-2.5-pro"))
	require.Equal(t, "", preferredBillingModelName("", "unknown"))
}

func TestAmountWindowTTLSecondsUsesUTC(t *testing.T) {
	now := time.Date(2026, time.April, 1, 23, 30, 0, 0, time.UTC)
	dailyTTL, weeklyTTL, monthlyTTL := amountWindowTTLSeconds(now)

	require.Equal(t, int64(30*60), dailyTTL)
	require.Equal(t, int64(4*24*60*60+30*60), weeklyTTL)
	require.Equal(t, int64(29*24*60*60+30*60), monthlyTTL)
}
