package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/higress-group/proxy-wasm-go-sdk/proxywasm"
	"github.com/higress-group/proxy-wasm-go-sdk/proxywasm/types"
	"github.com/higress-group/wasm-go/pkg/log"
	"github.com/higress-group/wasm-go/pkg/tokenusage"
	"github.com/higress-group/wasm-go/pkg/wrapper"
	"github.com/tidwall/gjson"
	"github.com/tidwall/resp"

	"github.com/alibaba/higress/plugins/wasm-go/extensions/ai-quota/util"
)

const (
	pluginName                   = "ai-quota"
	QuotaUnitToken               = "token"
	QuotaUnitAmount              = "amount"
	defaultQuotaRedisPrefix      = "chat_quota:"
	defaultBalanceKeyPrefix      = "billing:balance:"
	defaultPriceKeyPrefix        = "billing:model-price:"
	defaultUsageEventStream      = "billing:usage:stream"
	defaultUsageEventDedupPrefix = "billing:usage:event:"
	defaultUserPolicyKeyPrefix   = "billing:quota-policy:user:"
	defaultKeyPolicyKeyPrefix    = "billing:quota-policy:key:"
	defaultUserUsageKeyPrefix    = "billing:quota-usage:user:"
	defaultKeyUsageKeyPrefix     = "billing:quota-usage:key:"
	defaultQuotaNoopPrefix       = "billing:quota:noop:"
	ctxKeyAmountEventDispatched  = "amount_event_dispatched"
	ctxKeyRequestModelPrice      = "request_model_price"
	ctxKeyRequestModelName       = "request_model_name"
	ctxKeyRequestAPIKeyID        = "request_api_key_id"
	ctxKeyRequestID              = "request_id"
	ctxKeyTraceID                = "trace_id"
	ctxKeyRequestStartedAt       = "request_started_at"
	ctxKeyRequestPath            = "request_path"
	ctxKeyRequestKind            = "request_kind"
	headerAPIKeyID               = "x-higress-api-key-id"
)

type ChatMode string

const (
	ChatModeCompletion ChatMode = "completion"
	ChatModeAdmin      ChatMode = "admin"
	ChatModeNone       ChatMode = "none"
)

type AdminMode string

const (
	AdminModeRefresh AdminMode = "refresh"
	AdminModeQuery   AdminMode = "query"
	AdminModeDelta   AdminMode = "delta"
	AdminModeNone    AdminMode = "none"
)

func main() {}

func init() {
	wrapper.SetCtx(
		pluginName,
		wrapper.ParseConfig(parseConfig),
		wrapper.ProcessRequestHeaders(onHttpRequestHeaders),
		wrapper.ProcessRequestBody(onHttpRequestBody),
		wrapper.ProcessStreamingResponseBody(onHttpStreamingResponseBody),
		wrapper.ProcessStreamDone(onHttpStreamDone),
	)
}

type QuotaConfig struct {
	redisInfo             RedisInfo         `yaml:"redis"`
	QuotaUnit             string            `yaml:"quota_unit"`
	RedisKeyPrefix        string            `yaml:"redis_key_prefix"`
	BalanceKeyPrefix      string            `yaml:"balance_key_prefix"`
	PriceKeyPrefix        string            `yaml:"price_key_prefix"`
	UsageEventStream      string            `yaml:"usage_event_stream"`
	UsageEventDedupPrefix string            `yaml:"usage_event_dedup_prefix"`
	AdminConsumer         string            `yaml:"admin_consumer"`
	AdminPath             string            `yaml:"admin_path"`
	credential2Name       map[string]string `yaml:"-"`
	redisClient           wrapper.RedisClient
}

type Consumer struct {
	Name       string `yaml:"name"`
	Credential string `yaml:"credential"`
}

type RedisInfo struct {
	ServiceName string `required:"true" yaml:"service_name" json:"service_name"`
	ServicePort int    `required:"false" yaml:"service_port" json:"service_port"`
	Username    string `required:"false" yaml:"username" json:"username"`
	Password    string `required:"false" yaml:"password" json:"password"`
	Timeout     int    `required:"false" yaml:"timeout" json:"timeout"`
	Database    int    `required:"false" yaml:"database" json:"database"`
}

func parseConfig(json gjson.Result, config *QuotaConfig) error {
	log.Debugf("parse config()")
	// admin
	config.AdminPath = json.Get("admin_path").String()
	config.AdminConsumer = json.Get("admin_consumer").String()
	if config.AdminPath == "" {
		config.AdminPath = "/quota"
	}
	if config.AdminConsumer == "" {
		return errors.New("missing admin_consumer in config")
	}
	// Redis
	config.QuotaUnit = strings.ToLower(strings.TrimSpace(json.Get("quota_unit").String()))
	if config.QuotaUnit == "" {
		config.QuotaUnit = QuotaUnitToken
	}
	config.RedisKeyPrefix = json.Get("redis_key_prefix").String()
	if config.RedisKeyPrefix == "" {
		config.RedisKeyPrefix = defaultQuotaRedisPrefix
	}
	config.BalanceKeyPrefix = json.Get("balance_key_prefix").String()
	if config.BalanceKeyPrefix == "" && config.QuotaUnit == QuotaUnitAmount {
		config.BalanceKeyPrefix = defaultBalanceKeyPrefix
	}
	config.PriceKeyPrefix = json.Get("price_key_prefix").String()
	if config.PriceKeyPrefix == "" && config.QuotaUnit == QuotaUnitAmount {
		config.PriceKeyPrefix = defaultPriceKeyPrefix
	}
	config.UsageEventStream = json.Get("usage_event_stream").String()
	if config.UsageEventStream == "" && config.QuotaUnit == QuotaUnitAmount {
		config.UsageEventStream = defaultUsageEventStream
	}
	config.UsageEventDedupPrefix = json.Get("usage_event_dedup_prefix").String()
	if config.UsageEventDedupPrefix == "" && config.QuotaUnit == QuotaUnitAmount {
		config.UsageEventDedupPrefix = defaultUsageEventDedupPrefix
	}
	redisConfig := json.Get("redis")
	if !redisConfig.Exists() {
		return errors.New("missing redis in config")
	}
	serviceName := redisConfig.Get("service_name").String()
	if serviceName == "" {
		return errors.New("redis service name must not be empty")
	}
	servicePort := int(redisConfig.Get("service_port").Int())
	if servicePort == 0 {
		if strings.HasSuffix(serviceName, ".static") {
			// use default logic port which is 80 for static service
			servicePort = 80
		} else {
			servicePort = 6379
		}
	}
	username := redisConfig.Get("username").String()
	password := redisConfig.Get("password").String()
	timeout := int(redisConfig.Get("timeout").Int())
	if timeout == 0 {
		timeout = 1000
	}
	database := int(redisConfig.Get("database").Int())
	config.redisInfo.ServiceName = serviceName
	config.redisInfo.ServicePort = servicePort
	config.redisInfo.Username = username
	config.redisInfo.Password = password
	config.redisInfo.Timeout = timeout
	config.redisInfo.Database = database
	config.redisClient = wrapper.NewRedisClusterClient(wrapper.FQDNCluster{
		FQDN: serviceName,
		Port: int64(servicePort),
	})

	return config.redisClient.Init(username, password, int64(timeout), wrapper.WithDataBase(database))
}

func onHttpRequestHeaders(context wrapper.HttpContext, config QuotaConfig) types.Action {
	context.DisableReroute()
	log.Debugf("onHttpRequestHeaders()")
	// get tokens
	consumer, err := proxywasm.GetHttpRequestHeader("x-mse-consumer")
	if err != nil {
		return deniedNoKeyAuthData()
	}
	if consumer == "" {
		return deniedUnauthorizedConsumer()
	}

	rawPath := context.Path()
	path, _ := url.Parse(rawPath)
	chatMode, adminMode := getOperationMode(path.Path, config.AdminPath)
	context.SetContext("chatMode", chatMode)
	context.SetContext("adminMode", adminMode)
	context.SetContext("consumer", consumer)
	context.SetContext(ctxKeyRequestPath, path.Path)
	context.SetContext(ctxKeyRequestKind, requestKindFromPath(path.Path))
	context.SetContext(ctxKeyRequestStartedAt, time.Now().UTC().Format(time.RFC3339Nano))
	requestID := firstNonEmptyString(
		getRequestHeader("x-request-id"),
		getPropertyString("x_request_id"),
		fmt.Sprintf("%s-%d", consumer, time.Now().UnixNano()),
	)
	traceID := firstNonEmptyString(
		getRequestHeader("traceparent"),
		getRequestHeader("x-b3-traceid"),
		requestID,
	)
	context.SetContext(ctxKeyRequestID, requestID)
	context.SetContext(ctxKeyTraceID, traceID)
	if apiKeyID, _ := proxywasm.GetHttpRequestHeader(headerAPIKeyID); strings.TrimSpace(apiKeyID) != "" {
		context.SetContext(ctxKeyRequestAPIKeyID, strings.TrimSpace(apiKeyID))
	}
	log.Debugf("chatMode:%s, adminMode:%s, consumer:%s", chatMode, adminMode, consumer)
	if chatMode == ChatModeNone {
		return types.ActionContinue
	}
	if chatMode == ChatModeAdmin {
		// query quota
		if adminMode == AdminModeQuery {
			return queryQuota(context, config, consumer, path)
		}
		if adminMode == AdminModeRefresh || adminMode == AdminModeDelta {
			context.BufferRequestBody()
			return types.HeaderStopIteration
		}
		return types.ActionContinue
	}

	if config.QuotaUnit == QuotaUnitAmount {
		context.BufferRequestBody()
		return types.HeaderStopIteration
	}

	// there is no need to read request body when it is on chat completion mode
	context.DontReadRequestBody()
	// check quota here
	config.redisClient.Get(config.RedisKeyPrefix+consumer, func(response resp.Value) {
		isDenied := false
		if err := response.Error(); err != nil {
			isDenied = true
		}
		if response.IsNull() {
			isDenied = true
		}
		if response.Integer() <= 0 {
			isDenied = true
		}
		log.Debugf("get consumer:%s quota:%d isDenied:%t", consumer, response.Integer(), isDenied)
		if isDenied {
			util.SendResponse(http.StatusForbidden, "ai-quota.noquota", "text/plain", "Request denied by ai quota check, No quota left")
			return
		}
		proxywasm.ResumeHttpRequest()
	})
	return types.HeaderStopAllIterationAndWatermark
}

func onHttpRequestBody(ctx wrapper.HttpContext, config QuotaConfig, body []byte) types.Action {
	log.Debugf("onHttpRequestBody()")
	chatMode, ok := ctx.GetContext("chatMode").(ChatMode)
	if !ok {
		return types.ActionContinue
	}
	if chatMode == ChatModeNone {
		return types.ActionContinue
	}
	if chatMode == ChatModeCompletion {
		if config.QuotaUnit != QuotaUnitAmount {
			return types.ActionContinue
		}
		consumer, ok := ctx.GetContext("consumer").(string)
		if !ok || consumer == "" {
			return types.ActionContinue
		}
		captureDetailedRequestHints(ctx, body)
		return checkAmountQuota(ctx, config, consumer, body)
	}
	adminMode, ok := ctx.GetContext("adminMode").(AdminMode)
	if !ok {
		return types.ActionContinue
	}
	adminConsumer, ok := ctx.GetContext("consumer").(string)
	if !ok {
		return types.ActionContinue
	}

	if adminMode == AdminModeRefresh {
		return refreshQuota(ctx, config, adminConsumer, string(body))
	}
	if adminMode == AdminModeDelta {
		return deltaQuota(ctx, config, adminConsumer, string(body))
	}

	return types.ActionContinue
}

func onHttpStreamingResponseBody(ctx wrapper.HttpContext, config QuotaConfig, data []byte, endOfStream bool) []byte {
	chatMode, ok := ctx.GetContext("chatMode").(ChatMode)
	if !ok {
		return data
	}
	if chatMode == ChatModeNone || chatMode == ChatModeAdmin {
		return data
	}
	if usage := tokenusage.GetTokenUsage(ctx, data); usage.TotalToken > 0 {
		ctx.SetContext(tokenusage.CtxKeyInputToken, usage.InputToken)
		ctx.SetContext(tokenusage.CtxKeyOutputToken, usage.OutputToken)
	}
	detailedMetrics := mergeDetailedUsageFromResponse(ctx, data)
	if detailedMetrics.InputTokens > 0 {
		ctx.SetContext(tokenusage.CtxKeyInputToken, detailedMetrics.InputTokens)
	}
	if detailedMetrics.OutputTokens > 0 {
		ctx.SetContext(tokenusage.CtxKeyOutputToken, detailedMetrics.OutputTokens)
	}

	// chat completion mode
	if !endOfStream {
		return data
	}

	if ctx.GetContext("consumer") == nil {
		return data
	}

	consumer := ctx.GetContext("consumer").(string)
	statusCode := getResponseStatusCode()
	modelName := resolveEventModelName(ctx)
	if modelName == "" {
		modelName = "unknown"
	}

	inputToken := int64(0)
	outputToken := int64(0)
	if value := ctx.GetContext(tokenusage.CtxKeyInputToken); value != nil {
		inputToken = value.(int64)
	}
	if value := ctx.GetContext(tokenusage.CtxKeyOutputToken); value != nil {
		outputToken = value.(int64)
	}
	metrics := getDetailedUsageFromContext(ctx)
	if metrics.InputTokens > 0 {
		inputToken = metrics.InputTokens
	}
	if metrics.OutputTokens > 0 {
		outputToken = metrics.OutputTokens
	}
	totalToken := inputToken + outputToken
	if metrics.TotalTokens() > 0 {
		totalToken = metrics.TotalTokens()
	}
	if config.QuotaUnit == QuotaUnitAmount {
		if statusCode < 200 || statusCode >= 300 {
			emitAmountAuditEvent(ctx, config, consumer, modelName, "failed", "missing", statusCode,
				"upstream_failed", fmt.Sprintf("Upstream request finished with status %d.", statusCode))
			return data
		}
		if totalToken <= 0 && metrics.InputImageCount <= 0 && metrics.OutputImageCount <= 0 {
			emitAmountAuditEvent(ctx, config, consumer, modelName, "success", "missing", statusCode,
				"missing_usage", "Successful response did not expose billable usage.")
			return data
		}
		emitAmountUsageEvent(ctx, config, consumer, modelName, metrics, statusCode)
		return data
	}
	log.Debugf("update consumer:%s, totalToken:%d", consumer, totalToken)
	config.redisClient.DecrBy(config.RedisKeyPrefix+consumer, int(totalToken), nil)
	return data
}

func onHttpStreamDone(ctx wrapper.HttpContext, config QuotaConfig) {
	if config.QuotaUnit != QuotaUnitAmount || isAmountEventDispatched(ctx) {
		return
	}

	chatMode, ok := ctx.GetContext("chatMode").(ChatMode)
	if !ok || chatMode != ChatModeCompletion {
		return
	}

	consumer, ok := ctx.GetContext("consumer").(string)
	if !ok || strings.TrimSpace(consumer) == "" {
		return
	}

	statusCode := getResponseStatusCode()
	if statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices {
		statusCode = 499
	}

	modelName := resolveEventModelName(ctx)
	if modelName == "" {
		modelName = "unknown"
	}

	metrics := getDetailedUsageFromContext(ctx)
	if metrics.TotalTokens() > 0 || metrics.InputImageCount > 0 || metrics.OutputImageCount > 0 {
		emitAmountInterruptedUsageEvent(ctx, config, consumer, modelName, metrics, statusCode)
		return
	}

	emitAmountAuditEvent(ctx, config, consumer, modelName, "failed", "missing", statusCode,
		"stream_interrupted", "Streaming response terminated before final billable usage was emitted.")
}

func deniedNoKeyAuthData() types.Action {
	util.SendResponse(http.StatusUnauthorized, "ai-quota.no_key", "text/plain", "Request denied by ai quota check. No Key Authentication information found.")
	return types.ActionContinue
}

func deniedUnauthorizedConsumer() types.Action {
	util.SendResponse(http.StatusForbidden, "ai-quota.unauthorized", "text/plain", "Request denied by ai quota check. Unauthorized consumer.")
	return types.ActionContinue
}

func getOperationMode(path string, adminPath string) (ChatMode, AdminMode) {
	fullAdminPath := "/v1/chat/completions" + adminPath
	if strings.HasSuffix(path, fullAdminPath+"/refresh") {
		return ChatModeAdmin, AdminModeRefresh
	}
	if strings.HasSuffix(path, fullAdminPath+"/delta") {
		return ChatModeAdmin, AdminModeDelta
	}
	if strings.HasSuffix(path, fullAdminPath) {
		return ChatModeAdmin, AdminModeQuery
	}
	if isSupportedAIPath(path) {
		return ChatModeCompletion, AdminModeNone
	}
	return ChatModeNone, AdminModeNone
}

func isSupportedAIPath(path string) bool {
	supported := []string{
		"/v1/chat/completions",
		"/v1/completions",
		"/v1/responses",
		"/v1/embeddings",
	}
	for _, item := range supported {
		if strings.HasSuffix(path, item) {
			return true
		}
	}
	return false
}

func requestKindFromPath(path string) string {
	switch {
	case strings.HasSuffix(path, "/v1/chat/completions"):
		return "chat.completions"
	case strings.HasSuffix(path, "/v1/completions"):
		return "completions"
	case strings.HasSuffix(path, "/v1/responses"):
		return "responses"
	case strings.HasSuffix(path, "/v1/embeddings"):
		return "embeddings"
	default:
		return "unknown"
	}
}

func refreshQuota(ctx wrapper.HttpContext, config QuotaConfig, adminConsumer string, body string) types.Action {
	// check consumer
	if adminConsumer != config.AdminConsumer {
		util.SendResponse(http.StatusForbidden, "ai-quota.unauthorized", "text/plain", "Request denied by ai quota check. Unauthorized admin consumer.")
		return types.ActionContinue
	}

	queryValues, _ := url.ParseQuery(body)
	values := make(map[string]string, len(queryValues))
	for k, v := range queryValues {
		values[k] = v[0]
	}
	queryConsumer := values["consumer"]
	quota, err := strconv.ParseInt(values["quota"], 10, 64)
	if queryConsumer == "" || err != nil {
		util.SendResponse(http.StatusForbidden, "ai-quota.unauthorized", "text/plain", "Request denied by ai quota check. consumer can't be empty and quota must be integer.")
		return types.ActionContinue
	}
	key := quotaStorageKey(config, queryConsumer)
	err2 := config.redisClient.Set(key, quota, func(response resp.Value) {
		log.Debugf("Redis set key = %s quota = %d", key, quota)
		if err := response.Error(); err != nil {
			util.SendResponse(http.StatusServiceUnavailable, "ai-quota.error", "text/plain", fmt.Sprintf("redis error:%v", err))
			return
		}
		util.SendResponse(http.StatusOK, "ai-quota.refreshquota", "text/plain", "refresh quota successful")
	})

	if err2 != nil {
		util.SendResponse(http.StatusServiceUnavailable, "ai-quota.error", "text/plain", fmt.Sprintf("redis error:%v", err))
		return types.ActionContinue
	}

	return types.ActionPause
}

func queryQuota(ctx wrapper.HttpContext, config QuotaConfig, adminConsumer string, url *url.URL) types.Action {
	// check consumer
	if adminConsumer != config.AdminConsumer {
		util.SendResponse(http.StatusForbidden, "ai-quota.unauthorized", "text/plain", "Request denied by ai quota check. Unauthorized admin consumer.")
		return types.ActionContinue
	}
	// check url
	queryValues := url.Query()
	values := make(map[string]string, len(queryValues))
	for k, v := range queryValues {
		values[k] = v[0]
	}
	if values["consumer"] == "" {
		util.SendResponse(http.StatusForbidden, "ai-quota.unauthorized", "text/plain", "Request denied by ai quota check. consumer can't be empty.")
		return types.ActionContinue
	}
	queryConsumer := values["consumer"]
	key := quotaStorageKey(config, queryConsumer)
	err := config.redisClient.Get(key, func(response resp.Value) {
		quota := int64(0)
		if err := response.Error(); err != nil {
			util.SendResponse(http.StatusServiceUnavailable, "ai-quota.error", "text/plain", fmt.Sprintf("redis error:%v", err))
			return
		} else if response.IsNull() {
			quota = 0
		} else {
			quota = parseRespInteger(response)
		}
		result := struct {
			Consumer string `json:"consumer"`
			Quota    int64  `json:"quota"`
		}{
			Consumer: queryConsumer,
			Quota:    quota,
		}
		body, _ := json.Marshal(result)
		util.SendResponse(http.StatusOK, "ai-quota.queryquota", "application/json", string(body))
	})
	if err != nil {
		util.SendResponse(http.StatusServiceUnavailable, "ai-quota.error", "text/plain", fmt.Sprintf("redis error:%v", err))
		return types.ActionContinue
	}
	return types.ActionPause
}

func deltaQuota(ctx wrapper.HttpContext, config QuotaConfig, adminConsumer string, body string) types.Action {
	// check consumer
	if adminConsumer != config.AdminConsumer {
		util.SendResponse(http.StatusForbidden, "ai-quota.unauthorized", "text/plain", "Request denied by ai quota check. Unauthorized admin consumer.")
		return types.ActionContinue
	}

	queryValues, _ := url.ParseQuery(body)
	values := make(map[string]string, len(queryValues))
	for k, v := range queryValues {
		values[k] = v[0]
	}
	queryConsumer := values["consumer"]
	value, err := strconv.ParseInt(values["value"], 10, 64)
	if queryConsumer == "" || err != nil {
		util.SendResponse(http.StatusForbidden, "ai-quota.unauthorized", "text/plain", "Request denied by ai quota check. consumer can't be empty and value must be integer.")
		return types.ActionContinue
	}
	key := quotaStorageKey(config, queryConsumer)

	if value >= 0 {
		err := config.redisClient.Command([]interface{}{"incrby", key, value}, func(response resp.Value) {
			log.Debugf("Redis Incr key = %s value = %d", key, value)
			if err := response.Error(); err != nil {
				util.SendResponse(http.StatusServiceUnavailable, "ai-quota.error", "text/plain", fmt.Sprintf("redis error:%v", err))
				return
			}
			util.SendResponse(http.StatusOK, "ai-quota.deltaquota", "text/plain", "delta quota successful")
		})
		if err != nil {
			util.SendResponse(http.StatusServiceUnavailable, "ai-quota.error", "text/plain", fmt.Sprintf("redis error:%v", err))
			return types.ActionContinue
		}
	} else {
		err := config.redisClient.Command([]interface{}{"decrby", key, 0 - value}, func(response resp.Value) {
			log.Debugf("Redis Decr key = %s value = %d", key, 0-value)
			if err := response.Error(); err != nil {
				util.SendResponse(http.StatusServiceUnavailable, "ai-quota.error", "text/plain", fmt.Sprintf("redis error:%v", err))
				return
			}
			util.SendResponse(http.StatusOK, "ai-quota.deltaquota", "text/plain", "delta quota successful")
		})
		if err != nil {
			util.SendResponse(http.StatusServiceUnavailable, "ai-quota.error", "text/plain", fmt.Sprintf("redis error:%v", err))
			return types.ActionContinue
		}
	}

	return types.ActionPause
}

type modelPrice struct {
	ModelID                                             string
	PriceVersion                                        int64
	InputPer1K                                          int64
	OutputPer1K                                         int64
	InputRequestPriceMicroYuan                          int64
	CacheCreationInputTokenPricePer1KMicroYuan          int64
	CacheCreationInputTokenPriceAbove1hrPer1KMicroYuan  int64
	CacheReadInputTokenPricePer1KMicroYuan              int64
	InputTokenPriceAbove200kPer1KMicroYuan              int64
	OutputTokenPriceAbove200kPer1KMicroYuan             int64
	CacheCreationInputTokenPriceAbove200kPer1KMicroYuan int64
	CacheReadInputTokenPriceAbove200kPer1KMicroYuan     int64
	OutputImagePriceMicroYuan                           int64
	OutputImageTokenPricePer1KMicroYuan                 int64
	InputImagePriceMicroYuan                            int64
	InputImageTokenPricePer1KMicroYuan                  int64
}

func quotaStorageKey(config QuotaConfig, consumer string) string {
	if config.QuotaUnit == QuotaUnitAmount && config.BalanceKeyPrefix != "" {
		return config.BalanceKeyPrefix + consumer
	}
	return config.RedisKeyPrefix + consumer
}

func checkAmountQuota(ctx wrapper.HttpContext, config QuotaConfig, consumer string, body []byte) types.Action {
	modelName := extractRequestModel(body)
	if modelName == "" {
		emitAmountAuditEvent(ctx, config, consumer, "unknown", "rejected", "missing", http.StatusServiceUnavailable,
			"model_missing", "Request denied by ai quota check. Request model is missing.")
		util.SendResponse(http.StatusServiceUnavailable, "ai-quota.model_price_missing", "text/plain", "Request denied by ai quota check. Model pricing is unavailable.")
		return types.ActionPause
	}
	ctx.SetContext(ctxKeyRequestModelName, modelName)
	ctx.SetContext(tokenusage.CtxKeyRequestModel, modelName)

	priceKey := config.PriceKeyPrefix + modelName
	err := config.redisClient.Command([]interface{}{"hgetall", priceKey}, func(response resp.Value) {
		price, ok := parseModelPriceResponse(modelName, response)
		if err := response.Error(); err != nil {
			_ = proxywasm.ResumeHttpRequest()
			return
		}
		if !ok {
			emitAmountAuditEvent(ctx, config, consumer, modelName, "rejected", "missing", http.StatusServiceUnavailable,
				"model_price_missing", "Request denied by ai quota check. Model pricing is unavailable.")
			util.SendResponse(http.StatusServiceUnavailable, "ai-quota.model_price_missing", "text/plain", "Request denied by ai quota check. Model pricing is unavailable.")
			return
		}
		ctx.SetContext(ctxKeyRequestModelPrice, price)
		if err := evaluateAmountAdmission(ctx, config, consumer, modelName); err != nil {
			_ = proxywasm.ResumeHttpRequest()
		}
	})
	if err != nil {
		return types.ActionContinue
	}
	return types.ActionPause
}

func emitAmountUsageEvent(ctx wrapper.HttpContext, config QuotaConfig, consumer string, modelName string, metrics detailedUsageMetrics, httpStatus int) {
	emitAmountUsageEventWithStatus(ctx, config, consumer, modelName, metrics, httpStatus, "success", "", "")
}

func emitAmountInterruptedUsageEvent(ctx wrapper.HttpContext, config QuotaConfig, consumer string, modelName string, metrics detailedUsageMetrics, httpStatus int) {
	emitAmountUsageEventWithStatus(
		ctx,
		config,
		consumer,
		modelName,
		metrics,
		httpStatus,
		"failed",
		"stream_interrupted",
		"Streaming response terminated before final billable usage was emitted; charged with partial billable usage captured before interruption.",
	)
}

func emitAmountUsageEventWithStatus(
	ctx wrapper.HttpContext,
	config QuotaConfig,
	consumer string,
	modelName string,
	metrics detailedUsageMetrics,
	httpStatus int,
	requestStatus string,
	errorCode string,
	errorMessage string,
) {
	if storedPrice, ok := ctx.GetContext(ctxKeyRequestModelPrice).(modelPrice); ok {
		billingModelName := normalizeModelName(storedPrice.ModelID)
		if billingModelName == "" {
			billingModelName = preferredBillingModelName(ctx.GetStringContext(ctxKeyRequestModelName, ""), modelName)
		}
		dispatchAmountCharge(ctx, config, consumer, billingModelName, storedPrice, metrics, httpStatus, requestStatus, errorCode, errorMessage)
		return
	}

	priceKey := config.PriceKeyPrefix + modelName
	_ = config.redisClient.Command([]interface{}{"hgetall", priceKey}, func(response resp.Value) {
		price, ok := parseModelPriceResponse(modelName, response)
		if err := response.Error(); err != nil || !ok {
			emitAmountAuditEvent(ctx, config, consumer, modelName, requestStatus, "missing", httpStatus,
				"model_price_missing", "Successful response could not find model pricing.")
			return
		}
		dispatchAmountCharge(ctx, config, consumer, modelName, price, metrics, httpStatus, requestStatus, errorCode, errorMessage)
	})
}

func dispatchAmountCharge(
	ctx wrapper.HttpContext,
	config QuotaConfig,
	consumer string,
	modelName string,
	price modelPrice,
	metrics detailedUsageMetrics,
	httpStatus int,
	requestStatus string,
	errorCode string,
	errorMessage string,
) {
	cost := calculateAmountCost(metrics, price)
	dispatchAmountEvent(ctx, config, amountEvent{
		Consumer:                   consumer,
		ModelName:                  modelName,
		RequestStatus:              requestStatus,
		UsageStatus:                "parsed",
		ErrorCode:                  errorCode,
		ErrorMessage:               errorMessage,
		HTTPStatus:                 httpStatus,
		InputToken:                 metrics.InputTokens,
		OutputToken:                metrics.OutputTokens,
		TotalToken:                 metrics.TotalTokens(),
		CacheCreationInputTokens:   metrics.CacheCreationInputTokens,
		CacheCreation5mInputTokens: metrics.CacheCreation5mInputTokens,
		CacheCreation1hInputTokens: metrics.CacheCreation1hInputTokens,
		CacheReadInputTokens:       metrics.CacheReadInputTokens,
		InputImageTokens:           metrics.InputImageTokens,
		OutputImageTokens:          metrics.OutputImageTokens,
		InputImageCount:            metrics.InputImageCount,
		OutputImageCount:           metrics.OutputImageCount,
		RequestCount:               maxInt64Value(metrics.RequestCount, 1),
		CacheTTL:                   metrics.CacheTTL,
		InputTokenDetailsJSON:      buildInputTokenDetailsJSON(metrics),
		OutputTokenDetailsJSON:     buildOutputTokenDetailsJSON(metrics),
		ProviderUsageJSON:          buildProviderUsageJSON(metrics),
		Cost:                       cost,
		PriceVersion:               price.PriceVersion,
	})
}

type amountEvent struct {
	Consumer                   string
	ModelName                  string
	RequestStatus              string
	UsageStatus                string
	HTTPStatus                 int
	ErrorCode                  string
	ErrorMessage               string
	InputToken                 int64
	OutputToken                int64
	TotalToken                 int64
	CacheCreationInputTokens   int64
	CacheCreation5mInputTokens int64
	CacheCreation1hInputTokens int64
	CacheReadInputTokens       int64
	InputImageTokens           int64
	OutputImageTokens          int64
	InputImageCount            int64
	OutputImageCount           int64
	RequestCount               int64
	CacheTTL                   string
	InputTokenDetailsJSON      string
	OutputTokenDetailsJSON     string
	ProviderUsageJSON          string
	Cost                       int64
	PriceVersion               int64
}

func emitAmountAuditEvent(ctx wrapper.HttpContext, config QuotaConfig, consumer string, modelName string, requestStatus string, usageStatus string, httpStatus int, errorCode string, errorMessage string) {
	dispatchAmountEvent(ctx, config, amountEvent{
		Consumer:      consumer,
		ModelName:     modelName,
		RequestStatus: requestStatus,
		UsageStatus:   usageStatus,
		HTTPStatus:    httpStatus,
		ErrorCode:     errorCode,
		ErrorMessage:  errorMessage,
	})
}

func dispatchAmountEvent(ctx wrapper.HttpContext, config QuotaConfig, event amountEvent) {
	requestID := strings.TrimSpace(ctx.GetStringContext(ctxKeyRequestID, ""))
	eventID := requestID
	if eventID == "" {
		eventID = fmt.Sprintf("%s-%d", event.Consumer, time.Now().UnixNano())
	}
	dedupKey := config.UsageEventDedupPrefix + eventID
	keys := []interface{}{
		quotaStorageKey(config, event.Consumer),
		dedupKey,
		config.UsageEventStream,
		billingUsageKey(defaultUserUsageKeyPrefix, quotaWindowTotal, event.Consumer),
		billingUsageKey(defaultUserUsageKeyPrefix, quotaWindowDaily, event.Consumer),
		billingUsageKey(defaultUserUsageKeyPrefix, quotaWindowWeekly, event.Consumer),
		billingUsageKey(defaultUserUsageKeyPrefix, quotaWindowMonthly, event.Consumer),
		billingUsageKey(defaultKeyUsageKeyPrefix, quotaWindowTotal, firstNonEmptyString(ctx.GetStringContext(ctxKeyRequestAPIKeyID, ""), event.Consumer)),
		billingUsageKey(defaultKeyUsageKeyPrefix, quotaWindowDaily, firstNonEmptyString(ctx.GetStringContext(ctxKeyRequestAPIKeyID, ""), event.Consumer)),
		billingUsageKey(defaultKeyUsageKeyPrefix, quotaWindowWeekly, firstNonEmptyString(ctx.GetStringContext(ctxKeyRequestAPIKeyID, ""), event.Consumer)),
		billingUsageKey(defaultKeyUsageKeyPrefix, quotaWindowMonthly, firstNonEmptyString(ctx.GetStringContext(ctxKeyRequestAPIKeyID, ""), event.Consumer)),
	}
	now := time.Now().UTC()
	dailyTTL, weeklyTTL, monthlyTTL := amountWindowTTLSeconds(now)
	args := buildAmountChargeArgs(
		event.Cost,
		eventID,
		requestID,
		strings.TrimSpace(ctx.GetStringContext(ctxKeyTraceID, "")),
		event.Consumer,
		getPropertyString("route_name"),
		strings.TrimSpace(ctx.GetStringContext(ctxKeyRequestPath, "")),
		strings.TrimSpace(ctx.GetStringContext(ctxKeyRequestKind, "")),
		event.ModelName,
		event.RequestStatus,
		event.UsageStatus,
		event.HTTPStatus,
		event.ErrorCode,
		event.ErrorMessage,
		event.InputToken,
		event.OutputToken,
		event.TotalToken,
		event.CacheCreationInputTokens,
		event.CacheCreation5mInputTokens,
		event.CacheCreation1hInputTokens,
		event.CacheReadInputTokens,
		event.InputImageTokens,
		event.OutputImageTokens,
		event.InputImageCount,
		event.OutputImageCount,
		event.RequestCount,
		event.CacheTTL,
		event.InputTokenDetailsJSON,
		event.OutputTokenDetailsJSON,
		event.ProviderUsageJSON,
		event.PriceVersion,
		strings.TrimSpace(ctx.GetStringContext(ctxKeyRequestAPIKeyID, "")),
		parseRFC3339Time(ctx.GetStringContext(ctxKeyRequestStartedAt, "")),
		now,
		now,
		dailyTTL,
		weeklyTTL,
		monthlyTTL,
	)
	_ = config.redisClient.Eval(amountChargeScript, len(keys), keys, args, nil)
	markAmountEventDispatched(ctx)
}

func buildAmountChargeArgs(cost int64, eventID, requestID, traceID, consumer, routeName, requestPath, requestKind, modelName, requestStatus, usageStatus string, httpStatus int, errorCode, errorMessage string, inputToken, outputToken, totalToken, cacheCreationInputToken, cacheCreation5mInputToken, cacheCreation1hInputToken, cacheReadInputToken, inputImageToken, outputImageToken, inputImageCount, outputImageCount, requestCount int64, cacheTTL, inputTokenDetailsJSON, outputTokenDetailsJSON, providerUsageJSON string, priceVersion int64, apiKeyID string, startedAt time.Time, finishedAt time.Time, occurredAt time.Time, dailyTTL int64, weeklyTTL int64, monthlyTTL int64) []interface{} {
	return []interface{}{
		cost,
		eventID,
		requestID,
		traceID,
		consumer,
		routeName,
		requestPath,
		requestKind,
		modelName,
		requestStatus,
		usageStatus,
		httpStatus,
		errorCode,
		errorMessage,
		inputToken,
		outputToken,
		totalToken,
		cacheCreationInputToken,
		cacheCreation5mInputToken,
		cacheCreation1hInputToken,
		cacheReadInputToken,
		inputImageToken,
		outputImageToken,
		inputImageCount,
		outputImageCount,
		requestCount,
		cacheTTL,
		inputTokenDetailsJSON,
		outputTokenDetailsJSON,
		providerUsageJSON,
		priceVersion,
		apiKeyID,
		startedAt.Format(time.RFC3339Nano),
		finishedAt.Format(time.RFC3339Nano),
		occurredAt.Format(time.RFC3339Nano),
		dailyTTL,
		weeklyTTL,
		monthlyTTL,
	}
}

func calculateAmountCost(metrics detailedUsageMetrics, price modelPrice) int64 {
	inputContextTokens := metrics.InputTokens + maxInt64Value(metrics.CacheCreationInputTokens,
		metrics.CacheCreation5mInputTokens+metrics.CacheCreation1hInputTokens) + metrics.CacheReadInputTokens + metrics.InputImageTokens
	useAbove200k := inputContextTokens > 200_000
	inputPer1K := choosePrice(useAbove200k, price.InputTokenPriceAbove200kPer1KMicroYuan, price.InputPer1K)
	outputPer1K := choosePrice(useAbove200k, price.OutputTokenPriceAbove200kPer1KMicroYuan, price.OutputPer1K)
	cacheCreationPer1K := choosePrice(useAbove200k, price.CacheCreationInputTokenPriceAbove200kPer1KMicroYuan,
		price.CacheCreationInputTokenPricePer1KMicroYuan)
	cacheReadPer1K := choosePrice(useAbove200k, price.CacheReadInputTokenPriceAbove200kPer1KMicroYuan,
		price.CacheReadInputTokenPricePer1KMicroYuan)
	cacheCreation1hPer1K := choosePrice(false, 0, price.CacheCreationInputTokenPriceAbove1hrPer1KMicroYuan)
	if cacheCreation1hPer1K == 0 {
		cacheCreation1hPer1K = cacheCreationPer1K
	}
	return roundTokenCost(metrics.InputTokens, inputPer1K) +
		roundTokenCost(metrics.OutputTokens, outputPer1K) +
		roundTokenCost(metrics.CacheCreation5mInputTokens, cacheCreationPer1K) +
		roundTokenCost(metrics.CacheCreation1hInputTokens, cacheCreation1hPer1K) +
		roundTokenCost(metrics.CacheReadInputTokens, cacheReadPer1K) +
		roundTokenCost(metrics.InputImageTokens, price.InputImageTokenPricePer1KMicroYuan) +
		roundTokenCost(metrics.OutputImageTokens, price.OutputImageTokenPricePer1KMicroYuan) +
		price.InputRequestPriceMicroYuan*maxInt64Value(metrics.RequestCount, 1) +
		price.InputImagePriceMicroYuan*metrics.InputImageCount +
		price.OutputImagePriceMicroYuan*metrics.OutputImageCount
}

func roundTokenCost(tokens int64, microPer1K int64) int64 {
	if tokens <= 0 || microPer1K <= 0 {
		return 0
	}
	numerator := tokens * microPer1K
	return (numerator + 500) / 1000
}

func parseModelPriceResponse(modelName string, response resp.Value) (modelPrice, bool) {
	if response.IsNull() || response.Error() != nil {
		return modelPrice{}, false
	}
	fields := response.Array()
	if len(fields) < 2 {
		return modelPrice{}, false
	}
	kv := make(map[string]string, len(fields)/2)
	for i := 0; i+1 < len(fields); i += 2 {
		kv[fields[i].String()] = fields[i+1].String()
	}
	inputPer1K, err := strconv.ParseInt(strings.TrimSpace(kv["input_price_per_1k_micro_yuan"]), 10, 64)
	if err != nil {
		return modelPrice{}, false
	}
	outputPer1K, err := strconv.ParseInt(strings.TrimSpace(kv["output_price_per_1k_micro_yuan"]), 10, 64)
	if err != nil {
		return modelPrice{}, false
	}
	priceVersion, _ := strconv.ParseInt(strings.TrimSpace(kv["price_version_id"]), 10, 64)
	modelID := normalizeModelName(kv["model_id"])
	if modelID == "" {
		modelID = modelName
	}
	return modelPrice{
		ModelID:                    modelID,
		PriceVersion:               priceVersion,
		InputPer1K:                 inputPer1K,
		OutputPer1K:                outputPer1K,
		InputRequestPriceMicroYuan: parseInt64String(kv["input_request_price_micro_yuan"]),
		CacheCreationInputTokenPricePer1KMicroYuan:          parseInt64String(kv["cache_creation_input_token_price_per_1k_micro_yuan"]),
		CacheCreationInputTokenPriceAbove1hrPer1KMicroYuan:  parseInt64String(kv["cache_creation_input_token_price_above_1hr_per_1k_micro_yuan"]),
		CacheReadInputTokenPricePer1KMicroYuan:              parseInt64String(kv["cache_read_input_token_price_per_1k_micro_yuan"]),
		InputTokenPriceAbove200kPer1KMicroYuan:              parseInt64String(kv["input_token_price_above_200k_per_1k_micro_yuan"]),
		OutputTokenPriceAbove200kPer1KMicroYuan:             parseInt64String(kv["output_token_price_above_200k_per_1k_micro_yuan"]),
		CacheCreationInputTokenPriceAbove200kPer1KMicroYuan: parseInt64String(kv["cache_creation_input_token_price_above_200k_per_1k_micro_yuan"]),
		CacheReadInputTokenPriceAbove200kPer1KMicroYuan:     parseInt64String(kv["cache_read_input_token_price_above_200k_per_1k_micro_yuan"]),
		OutputImagePriceMicroYuan:                           parseInt64String(kv["output_image_price_micro_yuan"]),
		OutputImageTokenPricePer1KMicroYuan:                 parseInt64String(kv["output_image_token_price_per_1k_micro_yuan"]),
		InputImagePriceMicroYuan:                            parseInt64String(kv["input_image_price_micro_yuan"]),
		InputImageTokenPricePer1KMicroYuan:                  parseInt64String(kv["input_image_token_price_per_1k_micro_yuan"]),
	}, true
}

func parseRespInteger(value resp.Value) int64 {
	raw := strings.TrimSpace(value.String())
	if raw == "" {
		return 0
	}
	parsed, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0
	}
	return parsed
}

func parseInt64String(value string) int64 {
	parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil {
		return 0
	}
	return parsed
}

func choosePrice(usePrimary bool, primary int64, fallback int64) int64 {
	if usePrimary && primary > 0 {
		return primary
	}
	return fallback
}

func buildInputTokenDetailsJSON(metrics detailedUsageMetrics) string {
	payload := map[string]int64{}
	if metrics.CacheReadInputTokens > 0 {
		payload["cached_tokens"] = metrics.CacheReadInputTokens
	}
	if metrics.CacheCreationInputTokens > 0 {
		payload["cache_creation_input_tokens"] = metrics.CacheCreationInputTokens
	}
	if metrics.CacheCreation5mInputTokens > 0 {
		payload["cache_creation_5m_input_tokens"] = metrics.CacheCreation5mInputTokens
	}
	if metrics.CacheCreation1hInputTokens > 0 {
		payload["cache_creation_1h_input_tokens"] = metrics.CacheCreation1hInputTokens
	}
	if metrics.InputImageTokens > 0 {
		payload["image_tokens"] = metrics.InputImageTokens
	}
	if len(payload) == 0 {
		return ""
	}
	body, _ := json.Marshal(payload)
	return string(body)
}

func buildOutputTokenDetailsJSON(metrics detailedUsageMetrics) string {
	payload := map[string]int64{}
	if metrics.OutputImageTokens > 0 {
		payload["image_tokens"] = metrics.OutputImageTokens
	}
	if len(payload) == 0 {
		return ""
	}
	body, _ := json.Marshal(payload)
	return string(body)
}

func buildProviderUsageJSON(metrics detailedUsageMetrics) string {
	body, _ := json.Marshal(map[string]any{
		"input_tokens":                   metrics.InputTokens,
		"output_tokens":                  metrics.OutputTokens,
		"cache_creation_input_tokens":    metrics.CacheCreationInputTokens,
		"cache_creation_5m_input_tokens": metrics.CacheCreation5mInputTokens,
		"cache_creation_1h_input_tokens": metrics.CacheCreation1hInputTokens,
		"cache_read_input_tokens":        metrics.CacheReadInputTokens,
		"input_image_tokens":             metrics.InputImageTokens,
		"output_image_tokens":            metrics.OutputImageTokens,
		"input_image_count":              metrics.InputImageCount,
		"output_image_count":             metrics.OutputImageCount,
		"request_count":                  metrics.RequestCount,
		"cache_ttl":                      metrics.CacheTTL,
	})
	return string(body)
}

func extractRequestModel(body []byte) string {
	modelName := strings.TrimSpace(gjson.GetBytes(body, "model").String())
	return normalizeModelName(modelName)
}

func normalizeModelName(modelName string) string {
	normalized := strings.TrimSpace(modelName)
	if normalized == "" || strings.EqualFold(normalized, tokenusage.ModelUnknown) {
		return ""
	}
	return normalized
}

func resolveEventModelName(ctx wrapper.HttpContext) string {
	responseModel := ""
	if model, ok := ctx.GetUserAttribute(tokenusage.CtxKeyModel).(string); ok {
		responseModel = model
	}
	return preferredBillingModelName(ctx.GetStringContext(ctxKeyRequestModelName, ""), responseModel)
}

func preferredBillingModelName(requestModel string, responseModel string) string {
	return firstNonEmptyString(normalizeModelName(requestModel), normalizeModelName(responseModel))
}

func getResponseStatusCode() int {
	statusText, err := proxywasm.GetHttpResponseHeader(":status")
	if err != nil {
		return http.StatusOK
	}
	statusCode, parseErr := strconv.Atoi(strings.TrimSpace(statusText))
	if parseErr != nil || statusCode <= 0 {
		return http.StatusOK
	}
	return statusCode
}

func isAmountEventDispatched(ctx wrapper.HttpContext) bool {
	value := ctx.GetContext(ctxKeyAmountEventDispatched)
	dispatched, ok := value.(bool)
	return ok && dispatched
}

func markAmountEventDispatched(ctx wrapper.HttpContext) {
	ctx.SetContext(ctxKeyAmountEventDispatched, true)
}

func parseRFC3339Time(value string) time.Time {
	if strings.TrimSpace(value) == "" {
		return time.Now().UTC()
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Now().UTC()
	}
	return parsed
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func getRequestHeader(name string) string {
	value, err := proxywasm.GetHttpRequestHeader(name)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(value)
}

func getPropertyString(key string) string {
	raw, err := proxywasm.GetProperty([]string{key})
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(raw))
}

const (
	quotaWindowTotal   = "total"
	quotaWindow5h      = "5h"
	quotaWindowDaily   = "daily"
	quotaWindowWeekly  = "weekly"
	quotaWindowMonthly = "monthly"
)

func evaluateAmountAdmission(ctx wrapper.HttpContext, config QuotaConfig, consumer string, modelName string) error {
	apiKeyID := strings.TrimSpace(ctx.GetStringContext(ctxKeyRequestAPIKeyID, ""))
	keys := buildAmountAdmissionKeys(config, consumer, apiKeyID)
	return config.redisClient.Eval(amountAdmissionScript, len(keys), keys, []interface{}{apiKeyID}, func(response resp.Value) {
		if err := response.Error(); err != nil {
			_ = proxywasm.ResumeHttpRequest()
			return
		}
		allowed, reason := parseAmountAdmissionResponse(response)
		if allowed {
			_ = proxywasm.ResumeHttpRequest()
			return
		}
		emitAmountAuditEvent(ctx, config, consumer, modelName, "rejected", "missing", http.StatusForbidden,
			reason, amountAdmissionMessage(reason))
		util.SendResponse(http.StatusForbidden, "ai-quota.noquota", "text/plain", amountAdmissionMessage(reason))
	})
}

func buildAmountAdmissionKeys(config QuotaConfig, consumer string, apiKeyID string) []interface{} {
	apiKeyTarget := apiKeyID
	if apiKeyTarget == "" {
		apiKeyTarget = defaultQuotaNoopPrefix + consumer
	}
	return []interface{}{
		quotaStorageKey(config, consumer),
		defaultUserPolicyKeyPrefix + consumer,
		defaultKeyPolicyKeyPrefix + apiKeyTarget,
		billingUsageKey(defaultUserUsageKeyPrefix, quotaWindowTotal, consumer),
		billingUsageKey(defaultUserUsageKeyPrefix, quotaWindow5h, consumer),
		billingUsageKey(defaultUserUsageKeyPrefix, quotaWindowDaily, consumer),
		billingUsageKey(defaultUserUsageKeyPrefix, quotaWindowWeekly, consumer),
		billingUsageKey(defaultUserUsageKeyPrefix, quotaWindowMonthly, consumer),
		billingUsageKey(defaultKeyUsageKeyPrefix, quotaWindowTotal, apiKeyTarget),
		billingUsageKey(defaultKeyUsageKeyPrefix, quotaWindow5h, apiKeyTarget),
		billingUsageKey(defaultKeyUsageKeyPrefix, quotaWindowDaily, apiKeyTarget),
		billingUsageKey(defaultKeyUsageKeyPrefix, quotaWindowWeekly, apiKeyTarget),
		billingUsageKey(defaultKeyUsageKeyPrefix, quotaWindowMonthly, apiKeyTarget),
	}
}

func billingUsageKey(prefix string, window string, target string) string {
	return prefix + window + ":" + target
}

func parseAmountAdmissionResponse(response resp.Value) (bool, string) {
	fields := response.Array()
	if len(fields) < 2 {
		return true, "ok"
	}
	allowed := parseRespInteger(fields[0]) > 0
	reason := strings.TrimSpace(fields[1].String())
	if reason == "" {
		reason = "ok"
	}
	return allowed, reason
}

func amountAdmissionMessage(reason string) string {
	switch reason {
	case "no_quota":
		return "Request denied by ai quota check, No quota left"
	case "key_total_limit_exceeded":
		return "Request denied by ai quota check, API Key total quota exceeded"
	case "user_total_limit_exceeded":
		return "Request denied by ai quota check, User total quota exceeded"
	case "key_5h_limit_exceeded":
		return "Request denied by ai quota check, API Key 5h quota exceeded"
	case "user_5h_limit_exceeded":
		return "Request denied by ai quota check, User 5h quota exceeded"
	case "key_daily_limit_exceeded":
		return "Request denied by ai quota check, API Key daily quota exceeded"
	case "user_daily_limit_exceeded":
		return "Request denied by ai quota check, User daily quota exceeded"
	case "key_weekly_limit_exceeded":
		return "Request denied by ai quota check, API Key weekly quota exceeded"
	case "user_weekly_limit_exceeded":
		return "Request denied by ai quota check, User weekly quota exceeded"
	case "key_monthly_limit_exceeded":
		return "Request denied by ai quota check, API Key monthly quota exceeded"
	case "user_monthly_limit_exceeded":
		return "Request denied by ai quota check, User monthly quota exceeded"
	default:
		return "Request denied by ai quota check, Quota window exceeded"
	}
}

func amountWindowTTLSeconds(now time.Time) (int64, int64, int64) {
	location := time.UTC
	localNow := now.In(location)
	nextDay := time.Date(localNow.Year(), localNow.Month(), localNow.Day()+1, 0, 0, 0, 0, location)
	weekStart := localNow.AddDate(0, 0, -((int(localNow.Weekday()) + 6) % 7))
	nextWeek := time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, location).AddDate(0, 0, 7)
	nextMonth := time.Date(localNow.Year(), localNow.Month(), 1, 0, 0, 0, 0, location).AddDate(0, 1, 0)
	return int64(nextDay.Sub(localNow).Seconds()),
		int64(nextWeek.Sub(localNow).Seconds()),
		int64(nextMonth.Sub(localNow).Seconds())
}

const amountChargeScript = `
local balanceKey = KEYS[1]
local dedupKey = KEYS[2]
local streamKey = KEYS[3]
local userTotalKey = KEYS[4]
local userDailyKey = KEYS[5]
local userWeeklyKey = KEYS[6]
local userMonthlyKey = KEYS[7]
local keyTotalKey = KEYS[8]
local keyDailyKey = KEYS[9]
local keyWeeklyKey = KEYS[10]
local keyMonthlyKey = KEYS[11]
local cost = tonumber(ARGV[1])
if redis.call('SETNX', dedupKey, ARGV[2]) == 0 then
  return {0, redis.call('GET', balanceKey) or '0'}
end
redis.call('EXPIRE', dedupKey, 86400)
local usageStatus = ARGV[11]
local apiKeyId = ARGV[32]
local nextBalance = redis.call('GET', balanceKey) or '0'
if cost and cost > 0 and usageStatus == 'parsed' then
  nextBalance = redis.call('DECRBY', balanceKey, cost)
  redis.call('INCRBY', userTotalKey, cost)
  redis.call('INCRBY', userDailyKey, cost)
  redis.call('INCRBY', userWeeklyKey, cost)
  redis.call('INCRBY', userMonthlyKey, cost)
  local dailyTTL = tonumber(ARGV[36]) or 0
  local weeklyTTL = tonumber(ARGV[37]) or 0
  local monthlyTTL = tonumber(ARGV[38]) or 0
  if dailyTTL > 0 then redis.call('EXPIRE', userDailyKey, dailyTTL) end
  if weeklyTTL > 0 then redis.call('EXPIRE', userWeeklyKey, weeklyTTL) end
  if monthlyTTL > 0 then redis.call('EXPIRE', userMonthlyKey, monthlyTTL) end
  if apiKeyId ~= '' then
    redis.call('INCRBY', keyTotalKey, cost)
    redis.call('INCRBY', keyDailyKey, cost)
    redis.call('INCRBY', keyWeeklyKey, cost)
    redis.call('INCRBY', keyMonthlyKey, cost)
    if dailyTTL > 0 then redis.call('EXPIRE', keyDailyKey, dailyTTL) end
    if weeklyTTL > 0 then redis.call('EXPIRE', keyWeeklyKey, weeklyTTL) end
    if monthlyTTL > 0 then redis.call('EXPIRE', keyMonthlyKey, monthlyTTL) end
  end
end
redis.call('XADD', streamKey, '*',
  'event_id', ARGV[2],
  'request_id', ARGV[3],
  'trace_id', ARGV[4],
  'consumer_name', ARGV[5],
  'route_name', ARGV[6],
  'request_path', ARGV[7],
  'request_kind', ARGV[8],
  'model_id', ARGV[9],
  'request_status', ARGV[10],
  'usage_status', ARGV[11],
  'http_status', ARGV[12],
  'error_code', ARGV[13],
  'error_message', ARGV[14],
  'input_tokens', ARGV[15],
  'output_tokens', ARGV[16],
  'total_tokens', ARGV[17],
  'cache_creation_input_tokens', ARGV[18],
  'cache_creation_5m_input_tokens', ARGV[19],
  'cache_creation_1h_input_tokens', ARGV[20],
  'cache_read_input_tokens', ARGV[21],
  'input_image_tokens', ARGV[22],
  'output_image_tokens', ARGV[23],
  'input_image_count', ARGV[24],
  'output_image_count', ARGV[25],
  'request_count', ARGV[26],
  'cache_ttl', ARGV[27],
  'input_token_details_json', ARGV[28],
  'output_token_details_json', ARGV[29],
  'provider_usage_json', ARGV[30],
  'cost_micro_yuan', ARGV[1],
  'price_version_id', ARGV[31],
  'api_key_id', ARGV[32],
  'started_at', ARGV[33],
  'finished_at', ARGV[34],
  'occurred_at', ARGV[35]
)
return {1, nextBalance}
`

const amountAdmissionScript = `
local function intOrZero(value)
  local parsed = tonumber(value)
  if parsed == nil then
    return 0
  end
  return parsed
end

local function exceeds(limitValue, currentValue)
  return limitValue > 0 and currentValue >= limitValue
end

local balance = intOrZero(redis.call('GET', KEYS[1]))
if balance <= 0 then
  return {0, 'no_quota', balance}
end

local userLimits = redis.call('HMGET', KEYS[2],
  'limit_total_micro_yuan',
  'limit_5h_micro_yuan',
  'limit_daily_micro_yuan',
  'limit_weekly_micro_yuan',
  'limit_monthly_micro_yuan')
local keyLimits = redis.call('HMGET', KEYS[3],
  'limit_total_micro_yuan',
  'limit_5h_micro_yuan',
  'limit_daily_micro_yuan',
  'limit_weekly_micro_yuan',
  'limit_monthly_micro_yuan')

local userTotal = intOrZero(redis.call('GET', KEYS[4]))
local user5h = intOrZero(redis.call('GET', KEYS[5]))
local userDaily = intOrZero(redis.call('GET', KEYS[6]))
local userWeekly = intOrZero(redis.call('GET', KEYS[7]))
local userMonthly = intOrZero(redis.call('GET', KEYS[8]))
local keyTotal = intOrZero(redis.call('GET', KEYS[9]))
local key5h = intOrZero(redis.call('GET', KEYS[10]))
local keyDaily = intOrZero(redis.call('GET', KEYS[11]))
local keyWeekly = intOrZero(redis.call('GET', KEYS[12]))
local keyMonthly = intOrZero(redis.call('GET', KEYS[13]))

local hasKeyPolicy = ARGV[1] ~= ''
if hasKeyPolicy and exceeds(intOrZero(keyLimits[1]), keyTotal) then
  return {0, 'key_total_limit_exceeded', balance}
end
if exceeds(intOrZero(userLimits[1]), userTotal) then
  return {0, 'user_total_limit_exceeded', balance}
end
if hasKeyPolicy and exceeds(intOrZero(keyLimits[2]), key5h) then
  return {0, 'key_5h_limit_exceeded', balance}
end
if exceeds(intOrZero(userLimits[2]), user5h) then
  return {0, 'user_5h_limit_exceeded', balance}
end
if hasKeyPolicy and exceeds(intOrZero(keyLimits[3]), keyDaily) then
  return {0, 'key_daily_limit_exceeded', balance}
end
if exceeds(intOrZero(userLimits[3]), userDaily) then
  return {0, 'user_daily_limit_exceeded', balance}
end
if hasKeyPolicy and exceeds(intOrZero(keyLimits[4]), keyWeekly) then
  return {0, 'key_weekly_limit_exceeded', balance}
end
if exceeds(intOrZero(userLimits[4]), userWeekly) then
  return {0, 'user_weekly_limit_exceeded', balance}
end
if hasKeyPolicy and exceeds(intOrZero(keyLimits[5]), keyMonthly) then
  return {0, 'key_monthly_limit_exceeded', balance}
end
if exceeds(intOrZero(userLimits[5]), userMonthly) then
  return {0, 'user_monthly_limit_exceeded', balance}
end
return {1, 'ok', balance}
`
