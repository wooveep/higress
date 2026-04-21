package main

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/higress-group/wasm-go/pkg/tokenusage"
	"github.com/higress-group/wasm-go/pkg/wrapper"
	"github.com/tidwall/gjson"
	"github.com/tidwall/resp"
)

const ctxKeyAmountUsageMetrics = "amount_usage_metrics"

type detailedUsageMetrics struct {
	Model                      string
	InputTokens                int64
	OutputTokens               int64
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
}

type rateLimitModelPrice struct {
	ModelID                                                string
	PriceVersion                                           int64
	InputMicroYuanPerToken                                 int64
	OutputMicroYuanPerToken                                int64
	InputRequestPriceMicroYuan                             int64
	CacheCreationInputTokenPriceMicroYuanPerToken          int64
	CacheCreationInputTokenPriceAbove1hrMicroYuanPerToken  int64
	CacheReadInputTokenPriceMicroYuanPerToken              int64
	InputTokenPriceAbove200kMicroYuanPerToken              int64
	OutputTokenPriceAbove200kMicroYuanPerToken             int64
	CacheCreationInputTokenPriceAbove200kMicroYuanPerToken int64
	CacheReadInputTokenPriceAbove200kMicroYuanPerToken     int64
	OutputImagePriceMicroYuan                              int64
	OutputImageTokenPriceMicroYuanPerToken                 int64
	InputImagePriceMicroYuan                               int64
	InputImageTokenPriceMicroYuanPerToken                  int64
}

func mergeDetailedUsageFromResponse(ctx wrapper.HttpContext, data []byte) detailedUsageMetrics {
	var metrics detailedUsageMetrics
	if existing, ok := ctx.GetContext(ctxKeyAmountUsageMetrics).(detailedUsageMetrics); ok {
		metrics = existing
	}
	metrics.merge(parseUsageMetricsChunk(data))
	ctx.SetContext(ctxKeyAmountUsageMetrics, metrics)
	return metrics
}

func getDetailedUsageFromContext(ctx wrapper.HttpContext) detailedUsageMetrics {
	if value, ok := ctx.GetContext(ctxKeyAmountUsageMetrics).(detailedUsageMetrics); ok {
		return value
	}
	return detailedUsageMetrics{}
}

func (m *detailedUsageMetrics) merge(other detailedUsageMetrics) {
	m.Model = firstNonEmptyString(m.Model, other.Model)
	m.InputTokens = maxInt64(m.InputTokens, other.InputTokens)
	m.OutputTokens = maxInt64(m.OutputTokens, other.OutputTokens)
	m.CacheCreationInputTokens = maxInt64(m.CacheCreationInputTokens, other.CacheCreationInputTokens)
	m.CacheCreation5mInputTokens = maxInt64(m.CacheCreation5mInputTokens, other.CacheCreation5mInputTokens)
	m.CacheCreation1hInputTokens = maxInt64(m.CacheCreation1hInputTokens, other.CacheCreation1hInputTokens)
	m.CacheReadInputTokens = maxInt64(m.CacheReadInputTokens, other.CacheReadInputTokens)
	m.InputImageTokens = maxInt64(m.InputImageTokens, other.InputImageTokens)
	m.OutputImageTokens = maxInt64(m.OutputImageTokens, other.OutputImageTokens)
	m.InputImageCount = maxInt64(m.InputImageCount, other.InputImageCount)
	m.OutputImageCount = maxInt64(m.OutputImageCount, other.OutputImageCount)
	m.RequestCount = maxInt64(m.RequestCount, other.RequestCount)
	if m.CacheTTL == "" {
		m.CacheTTL = other.CacheTTL
	}
	m.normalize()
}

func (m *detailedUsageMetrics) normalize() {
	m.CacheTTL = normalizeCacheTTL(m.CacheTTL)
	splitTotal := m.CacheCreation5mInputTokens + m.CacheCreation1hInputTokens
	if splitTotal > m.CacheCreationInputTokens {
		m.CacheCreationInputTokens = splitTotal
	}
	remaining := m.CacheCreationInputTokens - splitTotal
	if remaining > 0 {
		if m.CacheTTL == "1h" {
			m.CacheCreation1hInputTokens += remaining
		} else {
			m.CacheCreation5mInputTokens += remaining
		}
	}
	if m.RequestCount == 0 {
		m.RequestCount = 1
	}
}

func (m detailedUsageMetrics) totalTokens() int64 {
	cacheCreation := maxInt64(m.CacheCreationInputTokens, m.CacheCreation5mInputTokens+m.CacheCreation1hInputTokens)
	return m.InputTokens + m.OutputTokens + cacheCreation + m.CacheReadInputTokens + m.InputImageTokens + m.OutputImageTokens
}

func parseUsageMetricsChunk(data []byte) detailedUsageMetrics {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return detailedUsageMetrics{}
	}
	if gjson.ValidBytes(trimmed) {
		return parseUsageMetricsJSON(trimmed)
	}
	if !bytes.Contains(trimmed, []byte("data:")) {
		return detailedUsageMetrics{}
	}
	metrics := detailedUsageMetrics{}
	for _, payload := range extractSSEPayloads(trimmed) {
		if gjson.ValidBytes(payload) {
			metrics.merge(parseUsageMetricsJSON(payload))
		}
	}
	return metrics
}

func parseUsageMetricsJSON(data []byte) detailedUsageMetrics {
	root := gjson.ParseBytes(data)
	metrics := detailedUsageMetrics{
		Model:            normalizeModelName(firstNonEmptyString(root.Get("model").String(), root.Get("message.model").String())),
		OutputImageCount: countJSONOutputImages(data),
	}
	switch strings.ToLower(strings.TrimSpace(root.Get("type").String())) {
	case "message_start":
		metrics.merge(extractUsageNode(root.Get("message.usage")))
	case "message_delta":
		metrics.merge(extractUsageNode(root.Get("usage")))
	}
	metrics.merge(extractUsageNode(root.Get("usage")))
	metrics.merge(extractUsageNode(root.Get("response.usage")))
	metrics.merge(extractUsageNode(root.Get("message.usage")))
	metrics.merge(extractGeminiUsageNode(root.Get("usageMetadata")))
	return metrics
}

func extractUsageNode(node gjson.Result) detailedUsageMetrics {
	if !node.Exists() || node.Type == gjson.Null {
		return detailedUsageMetrics{}
	}
	metrics := detailedUsageMetrics{
		InputTokens:              node.Get("input_tokens").Int(),
		OutputTokens:             node.Get("output_tokens").Int(),
		CacheCreationInputTokens: node.Get("cache_creation_input_tokens").Int(),
		CacheReadInputTokens:     node.Get("cache_read_input_tokens").Int(),
		OutputImageCount:         node.Get("generated_images").Int(),
		InputImageCount:          node.Get("input_images").Int(),
		RequestCount:             1,
		CacheTTL: firstNonEmptyString(
			normalizeCacheTTL(node.Get("cache_creation.ttl").String()),
			normalizeCacheTTL(node.Get("cache_ttl").String()),
			normalizeCacheTTL(node.Get("ttl").String()),
		),
	}
	metrics.CacheCreation5mInputTokens = node.Get("cache_creation.ephemeral_5m_input_tokens").Int()
	metrics.CacheCreation1hInputTokens = node.Get("cache_creation.ephemeral_1h_input_tokens").Int()
	if promptTokens := node.Get("prompt_tokens"); promptTokens.Exists() {
		cachedTokens := node.Get("input_tokens_details.cached_tokens").Int()
		metrics.CacheReadInputTokens = maxInt64(metrics.CacheReadInputTokens, cachedTokens)
		textInput := promptTokens.Int() - cachedTokens
		if textInput < 0 {
			textInput = 0
		}
		metrics.InputTokens = maxInt64(metrics.InputTokens, textInput)
	}
	if completionTokens := node.Get("completion_tokens"); completionTokens.Exists() {
		metrics.OutputTokens = maxInt64(metrics.OutputTokens, completionTokens.Int())
	}
	metrics.normalize()
	return metrics
}

func extractGeminiUsageNode(node gjson.Result) detailedUsageMetrics {
	if !node.Exists() || node.Type == gjson.Null {
		return detailedUsageMetrics{}
	}
	metrics := detailedUsageMetrics{
		CacheReadInputTokens: node.Get("cachedContentTokenCount").Int(),
		RequestCount:         1,
	}
	inputText, inputImage := sumModalityTokens(node.Get("promptTokensDetails"))
	outputText, outputImage := sumModalityTokens(node.Get("candidatesTokensDetails"))
	if inputText+inputImage > 0 {
		metrics.InputTokens = inputText
		metrics.InputImageTokens = inputImage
	} else if promptCount := node.Get("promptTokenCount").Int(); promptCount > 0 {
		adjusted := promptCount - metrics.CacheReadInputTokens
		if adjusted < 0 {
			adjusted = 0
		}
		metrics.InputTokens = adjusted
	}
	if outputText+outputImage > 0 {
		metrics.OutputTokens = outputText
		metrics.OutputImageTokens = outputImage
	} else if candidateCount := node.Get("candidatesTokenCount").Int(); candidateCount > 0 {
		metrics.OutputTokens = candidateCount
	}
	metrics.normalize()
	return metrics
}

func sumModalityTokens(node gjson.Result) (int64, int64) {
	if !node.Exists() || !node.IsArray() {
		return 0, 0
	}
	var textTokens int64
	var imageTokens int64
	for _, item := range node.Array() {
		count := item.Get("tokenCount").Int()
		if strings.EqualFold(item.Get("modality").String(), "IMAGE") {
			imageTokens += count
			continue
		}
		textTokens += count
	}
	return textTokens, imageTokens
}

func extractSSEPayloads(data []byte) [][]byte {
	lines := bytes.Split(data, []byte("\n"))
	payloads := make([][]byte, 0, len(lines))
	for _, line := range lines {
		trimmed := bytes.TrimSpace(line)
		if !bytes.HasPrefix(trimmed, []byte("data:")) {
			continue
		}
		payload := bytes.TrimSpace(bytes.TrimPrefix(trimmed, []byte("data:")))
		if len(payload) == 0 || bytes.Equal(payload, []byte("[DONE]")) {
			continue
		}
		payloads = append(payloads, payload)
	}
	return payloads
}

func normalizeCacheTTL(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1h", "60m", "3600", "3600s":
		return "1h"
	case "5m", "300", "300s", "":
		if strings.TrimSpace(value) == "" {
			return ""
		}
		return "5m"
	default:
		return "5m"
	}
}

func countJSONOutputImages(data []byte) int64 {
	var root any
	if err := json.Unmarshal(data, &root); err != nil {
		return 0
	}
	return countImageNodes(root, true)
}

func countImageNodes(value any, response bool) int64 {
	switch typed := value.(type) {
	case []any:
		var total int64
		for _, item := range typed {
			total += countImageNodes(item, response)
		}
		return total
	case map[string]any:
		if isImageNode(typed, response) {
			return 1
		}
		var total int64
		for _, item := range typed {
			total += countImageNodes(item, response)
		}
		return total
	default:
		return 0
	}
}

func isImageNode(node map[string]any, response bool) bool {
	typeName := strings.ToLower(strings.TrimSpace(stringValue(node["type"])))
	if response {
		if typeName == "output_image" || typeName == "image" || typeName == "generated_image" {
			return true
		}
		if _, ok := node["b64_json"]; ok {
			return true
		}
		if _, ok := node["inlineData"]; ok {
			return true
		}
		if _, ok := node["fileData"]; ok {
			return true
		}
	}
	return false
}

func stringValue(value any) string {
	if text, ok := value.(string); ok {
		return text
	}
	return ""
}

func maxInt64(left int64, right int64) int64 {
	if left >= right {
		return left
	}
	return right
}

func resolveAmountModelName(ctx wrapper.HttpContext, metrics detailedUsageMetrics) string {
	if normalized := normalizeModelName(metrics.Model); normalized != "" {
		return normalized
	}
	if model, ok := ctx.GetUserAttribute(tokenusage.CtxKeyModel).(string); ok {
		if normalized := normalizeModelName(model); normalized != "" {
			return normalized
		}
	}
	return ""
}

func normalizeModelName(modelName string) string {
	normalized := strings.TrimSpace(modelName)
	if normalized == "" || strings.EqualFold(normalized, tokenusage.ModelUnknown) {
		return ""
	}
	return normalized
}

func parseModelPriceResponse(modelName string, response resp.Value) (rateLimitModelPrice, bool) {
	if response.IsNull() || response.Error() != nil {
		return rateLimitModelPrice{}, false
	}
	fields := response.Array()
	if len(fields) < 2 {
		return rateLimitModelPrice{}, false
	}
	kv := make(map[string]string, len(fields)/2)
	for i := 0; i+1 < len(fields); i += 2 {
		kv[fields[i].String()] = fields[i+1].String()
	}
	inputPerToken, err := parseTokenPriceMicroYuanPerToken(kv, "input_price_micro_yuan_per_token", "input_price_per_1k_micro_yuan")
	if err != nil {
		return rateLimitModelPrice{}, false
	}
	outputPerToken, err := parseTokenPriceMicroYuanPerToken(kv, "output_price_micro_yuan_per_token", "output_price_per_1k_micro_yuan")
	if err != nil {
		return rateLimitModelPrice{}, false
	}
	priceVersion, _ := strconv.ParseInt(strings.TrimSpace(kv["price_version_id"]), 10, 64)
	modelID := normalizeModelName(kv["model_id"])
	if modelID == "" {
		modelID = modelName
	}
	return rateLimitModelPrice{
		ModelID:                    modelID,
		PriceVersion:               priceVersion,
		InputMicroYuanPerToken:     inputPerToken,
		OutputMicroYuanPerToken:    outputPerToken,
		InputRequestPriceMicroYuan: parseInt64String(kv["input_request_price_micro_yuan"]),
		CacheCreationInputTokenPriceMicroYuanPerToken:          parseTokenPriceWithFallback(kv, "cache_creation_input_token_price_micro_yuan_per_token", "cache_creation_input_token_price_per_1k_micro_yuan"),
		CacheCreationInputTokenPriceAbove1hrMicroYuanPerToken:  parseTokenPriceWithFallback(kv, "cache_creation_input_token_price_above_1hr_micro_yuan_per_token", "cache_creation_input_token_price_above_1hr_per_1k_micro_yuan"),
		CacheReadInputTokenPriceMicroYuanPerToken:              parseTokenPriceWithFallback(kv, "cache_read_input_token_price_micro_yuan_per_token", "cache_read_input_token_price_per_1k_micro_yuan"),
		InputTokenPriceAbove200kMicroYuanPerToken:              parseTokenPriceWithFallback(kv, "input_token_price_above_200k_micro_yuan_per_token", "input_token_price_above_200k_per_1k_micro_yuan"),
		OutputTokenPriceAbove200kMicroYuanPerToken:             parseTokenPriceWithFallback(kv, "output_token_price_above_200k_micro_yuan_per_token", "output_token_price_above_200k_per_1k_micro_yuan"),
		CacheCreationInputTokenPriceAbove200kMicroYuanPerToken: parseTokenPriceWithFallback(kv, "cache_creation_input_token_price_above_200k_micro_yuan_per_token", "cache_creation_input_token_price_above_200k_per_1k_micro_yuan"),
		CacheReadInputTokenPriceAbove200kMicroYuanPerToken:     parseTokenPriceWithFallback(kv, "cache_read_input_token_price_above_200k_micro_yuan_per_token", "cache_read_input_token_price_above_200k_per_1k_micro_yuan"),
		OutputImagePriceMicroYuan:                              parseInt64String(kv["output_image_price_micro_yuan"]),
		OutputImageTokenPriceMicroYuanPerToken:                 parseTokenPriceWithFallback(kv, "output_image_token_price_micro_yuan_per_token", "output_image_token_price_per_1k_micro_yuan"),
		InputImagePriceMicroYuan:                               parseInt64String(kv["input_image_price_micro_yuan"]),
		InputImageTokenPriceMicroYuanPerToken:                  parseTokenPriceWithFallback(kv, "input_image_token_price_micro_yuan_per_token", "input_image_token_price_per_1k_micro_yuan"),
	}, true
}

func calculateAmountCost(metrics detailedUsageMetrics, price rateLimitModelPrice) int64 {
	inputContextTokens := metrics.InputTokens + maxInt64(metrics.CacheCreationInputTokens,
		metrics.CacheCreation5mInputTokens+metrics.CacheCreation1hInputTokens) + metrics.CacheReadInputTokens + metrics.InputImageTokens
	useAbove200k := inputContextTokens > 200_000
	inputPerToken := choosePrice(useAbove200k, price.InputTokenPriceAbove200kMicroYuanPerToken, price.InputMicroYuanPerToken)
	outputPerToken := choosePrice(useAbove200k, price.OutputTokenPriceAbove200kMicroYuanPerToken, price.OutputMicroYuanPerToken)
	cacheCreationPerToken := choosePrice(useAbove200k, price.CacheCreationInputTokenPriceAbove200kMicroYuanPerToken,
		price.CacheCreationInputTokenPriceMicroYuanPerToken)
	cacheReadPerToken := choosePrice(useAbove200k, price.CacheReadInputTokenPriceAbove200kMicroYuanPerToken,
		price.CacheReadInputTokenPriceMicroYuanPerToken)
	cacheCreation1hPerToken := choosePrice(false, 0, price.CacheCreationInputTokenPriceAbove1hrMicroYuanPerToken)
	if cacheCreation1hPerToken == 0 {
		cacheCreation1hPerToken = cacheCreationPerToken
	}
	return roundTokenCost(metrics.InputTokens, inputPerToken) +
		roundTokenCost(metrics.OutputTokens, outputPerToken) +
		roundTokenCost(metrics.CacheCreation5mInputTokens, cacheCreationPerToken) +
		roundTokenCost(metrics.CacheCreation1hInputTokens, cacheCreation1hPerToken) +
		roundTokenCost(metrics.CacheReadInputTokens, cacheReadPerToken) +
		roundTokenCost(metrics.InputImageTokens, price.InputImageTokenPriceMicroYuanPerToken) +
		roundTokenCost(metrics.OutputImageTokens, price.OutputImageTokenPriceMicroYuanPerToken) +
		price.InputRequestPriceMicroYuan*maxInt64(metrics.RequestCount, 1) +
		price.InputImagePriceMicroYuan*metrics.InputImageCount +
		price.OutputImagePriceMicroYuan*metrics.OutputImageCount
}

func roundTokenCost(tokens int64, microPerToken int64) int64 {
	if tokens <= 0 || microPerToken <= 0 {
		return 0
	}
	return tokens * microPerToken
}

func parseInt64String(value string) int64 {
	parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil {
		return 0
	}
	return parsed
}

func parseNonNegativeInt64String(value string) (int64, error) {
	raw := strings.TrimSpace(value)
	if raw == "" {
		return 0, strconv.ErrSyntax
	}
	parsed, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, err
	}
	if parsed < 0 {
		return 0, strconv.ErrSyntax
	}
	return parsed, nil
}

func parseTokenPriceMicroYuanPerToken(kv map[string]string, modernKey string, legacyKey string) (int64, error) {
	if value, err := parseNonNegativeInt64String(kv[modernKey]); err == nil {
		return value, nil
	}
	legacy, err := parseNonNegativeInt64String(kv[legacyKey])
	if err != nil {
		return 0, err
	}
	return (legacy + 500) / 1000, nil
}

func parseTokenPriceWithFallback(kv map[string]string, modernKey string, legacyKey string) int64 {
	value, _ := parseTokenPriceMicroYuanPerToken(kv, modernKey, legacyKey)
	return value
}

func choosePrice(usePrimary bool, primary int64, fallback int64) int64 {
	if usePrimary && primary > 0 {
		return primary
	}
	return fallback
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
