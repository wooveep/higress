package main

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/higress-group/wasm-go/pkg/wrapper"
	"github.com/tidwall/gjson"
)

const (
	ctxKeyDetailedUsageMetrics = "detailed_usage_metrics"
	ctxKeyRequestCacheTTL      = "request_cache_ttl"
	ctxKeyRequestImageCount    = "request_image_count"
)

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

func (m *detailedUsageMetrics) Merge(other detailedUsageMetrics) {
	m.Model = firstNonEmptyString(m.Model, other.Model)
	m.InputTokens = maxInt64Value(m.InputTokens, other.InputTokens)
	m.OutputTokens = maxInt64Value(m.OutputTokens, other.OutputTokens)
	m.CacheCreationInputTokens = maxInt64Value(m.CacheCreationInputTokens, other.CacheCreationInputTokens)
	m.CacheCreation5mInputTokens = maxInt64Value(m.CacheCreation5mInputTokens, other.CacheCreation5mInputTokens)
	m.CacheCreation1hInputTokens = maxInt64Value(m.CacheCreation1hInputTokens, other.CacheCreation1hInputTokens)
	m.CacheReadInputTokens = maxInt64Value(m.CacheReadInputTokens, other.CacheReadInputTokens)
	m.InputImageTokens = maxInt64Value(m.InputImageTokens, other.InputImageTokens)
	m.OutputImageTokens = maxInt64Value(m.OutputImageTokens, other.OutputImageTokens)
	m.InputImageCount = maxInt64Value(m.InputImageCount, other.InputImageCount)
	m.OutputImageCount = maxInt64Value(m.OutputImageCount, other.OutputImageCount)
	m.RequestCount = maxInt64Value(m.RequestCount, other.RequestCount)
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
		if normalizeCacheTTL(m.CacheTTL) == "1h" {
			m.CacheCreation1hInputTokens += remaining
		} else {
			m.CacheCreation5mInputTokens += remaining
		}
	}
	if m.RequestCount == 0 {
		m.RequestCount = 1
	}
}

func (m detailedUsageMetrics) TotalTokens() int64 {
	cacheCreation := maxInt64Value(m.CacheCreationInputTokens, m.CacheCreation5mInputTokens+m.CacheCreation1hInputTokens)
	return m.InputTokens + m.OutputTokens + cacheCreation + m.CacheReadInputTokens + m.InputImageTokens + m.OutputImageTokens
}

func captureDetailedRequestHints(ctx wrapper.HttpContext, body []byte) {
	hints := parseRequestUsageHints(body)
	if hints.CacheTTL != "" {
		ctx.SetContext(ctxKeyRequestCacheTTL, hints.CacheTTL)
	}
	if hints.InputImageCount > 0 {
		ctx.SetContext(ctxKeyRequestImageCount, hints.InputImageCount)
	}
}

func mergeDetailedUsageFromResponse(ctx wrapper.HttpContext, data []byte) detailedUsageMetrics {
	var metrics detailedUsageMetrics
	if existing, ok := ctx.GetContext(ctxKeyDetailedUsageMetrics).(detailedUsageMetrics); ok {
		metrics = existing
	}
	requestTTL := ""
	if value, ok := ctx.GetContext(ctxKeyRequestCacheTTL).(string); ok {
		requestTTL = value
	}
	metrics.Merge(parseUsageMetricsChunk(data, requestTTL))
	if metrics.InputImageCount == 0 {
		if value, ok := ctx.GetContext(ctxKeyRequestImageCount).(int64); ok {
			metrics.InputImageCount = value
		}
	}
	ctx.SetContext(ctxKeyDetailedUsageMetrics, metrics)
	return metrics
}

func parseRequestUsageHints(body []byte) detailedUsageMetrics {
	if !gjson.ValidBytes(body) {
		return detailedUsageMetrics{}
	}
	var root any
	if err := json.Unmarshal(body, &root); err != nil {
		return detailedUsageMetrics{}
	}
	parsed := gjson.ParseBytes(body)
	return detailedUsageMetrics{
		CacheTTL: firstNonEmptyString(
			normalizeCacheTTL(parsed.Get("cache_control.ttl").String()),
			normalizeCacheTTL(parsed.Get("cacheControl.ttl").String()),
			normalizeCacheTTL(parsed.Get("cache_ttl").String()),
			normalizeCacheTTL(parsed.Get("ttl").String()),
		),
		InputImageCount: countImageNodes(root, false),
		RequestCount:    1,
	}
}

func parseUsageMetricsChunk(data []byte, requestTTL string) detailedUsageMetrics {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return detailedUsageMetrics{}
	}
	if gjson.ValidBytes(trimmed) {
		return parseUsageMetricsJSON(trimmed, requestTTL)
	}
	if !bytes.Contains(trimmed, []byte("data:")) {
		return detailedUsageMetrics{}
	}
	metrics := detailedUsageMetrics{}
	for _, payload := range extractSSEPayloads(trimmed) {
		if gjson.ValidBytes(payload) {
			metrics.Merge(parseUsageMetricsJSON(payload, requestTTL))
		}
	}
	return metrics
}

func parseUsageMetricsJSON(data []byte, requestTTL string) detailedUsageMetrics {
	root := gjson.ParseBytes(data)
	metrics := detailedUsageMetrics{
		Model:            firstNonEmptyString(root.Get("model").String(), root.Get("message.model").String()),
		OutputImageCount: countJSONOutputImages(data),
	}
	switch strings.ToLower(strings.TrimSpace(root.Get("type").String())) {
	case "message_start":
		metrics.Merge(extractUsageNode(root.Get("message.usage"), requestTTL))
	case "message_delta":
		metrics.Merge(extractUsageNode(root.Get("usage"), requestTTL))
	}
	metrics.Merge(extractUsageNode(root.Get("usage"), requestTTL))
	metrics.Merge(extractUsageNode(root.Get("response.usage"), requestTTL))
	metrics.Merge(extractUsageNode(root.Get("message.usage"), requestTTL))
	metrics.Merge(extractGeminiUsageNode(root.Get("usageMetadata"), requestTTL))
	return metrics
}

func extractUsageNode(node gjson.Result, requestTTL string) detailedUsageMetrics {
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
			normalizeCacheTTL(requestTTL),
		),
	}
	metrics.CacheCreation5mInputTokens = node.Get("cache_creation.ephemeral_5m_input_tokens").Int()
	metrics.CacheCreation1hInputTokens = node.Get("cache_creation.ephemeral_1h_input_tokens").Int()

	if promptTokens := node.Get("prompt_tokens"); promptTokens.Exists() {
		cachedTokens := node.Get("input_tokens_details.cached_tokens").Int()
		metrics.CacheReadInputTokens = maxInt64Value(metrics.CacheReadInputTokens, cachedTokens)
		textInput := promptTokens.Int() - cachedTokens
		if textInput < 0 {
			textInput = 0
		}
		metrics.InputTokens = maxInt64Value(metrics.InputTokens, textInput)
	}
	if completionTokens := node.Get("completion_tokens"); completionTokens.Exists() {
		metrics.OutputTokens = maxInt64Value(metrics.OutputTokens, completionTokens.Int())
	}
	metrics.normalize()
	return metrics
}

func extractGeminiUsageNode(node gjson.Result, requestTTL string) detailedUsageMetrics {
	if !node.Exists() || node.Type == gjson.Null {
		return detailedUsageMetrics{}
	}
	metrics := detailedUsageMetrics{
		CacheReadInputTokens: node.Get("cachedContentTokenCount").Int(),
		RequestCount:         1,
		CacheTTL:             normalizeCacheTTL(requestTTL),
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
	if !response {
		if typeName == "image_url" || typeName == "input_image" || typeName == "image" {
			return true
		}
		if _, ok := node["image_url"]; ok {
			return true
		}
		if _, ok := node["inline_data"]; ok {
			return true
		}
		if _, ok := node["image"]; ok && typeName == "" {
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

func maxInt64Value(left int64, right int64) int64 {
	if left >= right {
		return left
	}
	return right
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
