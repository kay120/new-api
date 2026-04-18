package channelctl

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ===========================================================================
// normalizeChannelTestEndpoint
// ===========================================================================

func TestNormalizeChannelTestEndpoint_ExplicitTypeWins(t *testing.T) {
	got := normalizeChannelTestEndpoint(nil, "gpt-4", "  openai_chat  ")
	assert.Equal(t, "openai_chat", got, "显式传入的 endpointType 应被 trim 后直接使用")
}

func TestNormalizeChannelTestEndpoint_CodexChannelDefaultsResponse(t *testing.T) {
	ch := &model.Channel{Type: constant.ChannelTypeCodex}
	got := normalizeChannelTestEndpoint(ch, "gpt-4", "")
	assert.Equal(t, string(constant.EndpointTypeOpenAIResponse), got,
		"Codex 渠道在无指定 endpoint 时默认 OpenAI Response")
}

func TestNormalizeChannelTestEndpoint_EmptyAllReturnsEmpty(t *testing.T) {
	got := normalizeChannelTestEndpoint(nil, "gpt-4", "")
	assert.Empty(t, got, "nil channel + 无 suffix + 无 endpointType 应返回空串")
}

// ===========================================================================
// isTransientUpstreamError
// ===========================================================================

func TestIsTransientUpstreamError(t *testing.T) {
	transient := []int{
		http.StatusTooManyRequests,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout,
	}
	for _, code := range transient {
		assert.True(t, isTransientUpstreamError(code), "%d 应被视为临时错误", code)
	}

	permanent := []int{
		http.StatusOK,
		http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusForbidden,
		http.StatusNotFound,
		http.StatusInternalServerError,
	}
	for _, code := range permanent {
		assert.False(t, isTransientUpstreamError(code), "%d 不应被视为临时错误", code)
	}
}

// ===========================================================================
// isReasoningTestModel & supportEnableThinkingField
// ===========================================================================

func TestIsReasoningTestModel(t *testing.T) {
	reasoning := []string{
		"o1-preview", "o3-mini", "o4-turbo",
		"deepseek-r1", "deepseek-reasoner",
		"qwq-32b", "custom-reasoner-x",
		"kimi-k2-thinking", "kimi-thinking",
		"glm-k2.5",
	}
	for _, m := range reasoning {
		assert.True(t, isReasoningTestModel(m), "%s 应识别为推理模型", m)
	}

	nonReasoning := []string{
		"gpt-4", "gpt-4o-mini", "claude-3-5-sonnet",
		"gemini-1.5-pro", "llama-3",
	}
	for _, m := range nonReasoning {
		assert.False(t, isReasoningTestModel(m), "%s 不应识别为推理模型", m)
	}
}

func TestSupportEnableThinkingField(t *testing.T) {
	supported := []string{
		"qwen3-max", "QWEN3-turbo",
		"deepseek-r1", "deepseek-v3.1", "deepseek-v3.2",
		"qwq-plus",
		"glm-4.5-flash",
	}
	for _, m := range supported {
		assert.True(t, supportEnableThinkingField(m), "%s 应支持 enable_thinking 字段", m)
	}

	unsupported := []string{
		"gpt-4",
		"claude-3-opus",
		"moonshot-v1-8k",
		"gemini-1.5-pro",
		"qwen2.5-max",
	}
	for _, m := range unsupported {
		assert.False(t, supportEnableThinkingField(m), "%s 不支持 enable_thinking 字段", m)
	}
}

// ===========================================================================
// coerceTestUsage
// ===========================================================================

func TestCoerceTestUsage_TypedUsagePointer(t *testing.T) {
	u := &dto.Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15}
	got, err := coerceTestUsage(u, false, 0)
	require.NoError(t, err)
	assert.Equal(t, u, got)
}

func TestCoerceTestUsage_TypedUsageValue(t *testing.T) {
	got, err := coerceTestUsage(dto.Usage{PromptTokens: 1}, false, 0)
	require.NoError(t, err)
	assert.Equal(t, 1, got.PromptTokens)
}

func TestCoerceTestUsage_NilWithoutStreamIsError(t *testing.T) {
	_, err := coerceTestUsage(nil, false, 0)
	require.Error(t, err, "非流式 + nil usage 必须报错")
}

func TestCoerceTestUsage_NilWithStreamFallsBackToEstimate(t *testing.T) {
	got, err := coerceTestUsage(nil, true, 42)
	require.NoError(t, err)
	assert.Equal(t, 42, got.PromptTokens)
	assert.Equal(t, 42, got.TotalTokens)
}

func TestCoerceTestUsage_UnknownTypeStreamFallback(t *testing.T) {
	// 流式下未知类型（例如 map）也应 fallback 为估算值
	got, err := coerceTestUsage(map[string]any{"foo": "bar"}, true, 7)
	require.NoError(t, err)
	assert.Equal(t, 7, got.PromptTokens)
}

func TestCoerceTestUsage_UnknownTypeNonStreamIsError(t *testing.T) {
	_, err := coerceTestUsage(map[string]any{"foo": "bar"}, false, 10)
	require.Error(t, err)
}

// ===========================================================================
// detectErrorMessageFromJSONBytes
// ===========================================================================

func TestDetectErrorMessageFromJSONBytes_StandardOpenAI(t *testing.T) {
	js := []byte(`{"error":{"message":"invalid api key","type":"authentication_error"}}`)
	assert.Equal(t, "invalid api key", detectErrorMessageFromJSONBytes(js))
}

func TestDetectErrorMessageFromJSONBytes_NestedError(t *testing.T) {
	// 某些代理返回 error.error.message 双层嵌套
	js := []byte(`{"error":{"error":{"message":"upstream blew up"}}}`)
	assert.Equal(t, "upstream blew up", detectErrorMessageFromJSONBytes(js))
}

func TestDetectErrorMessageFromJSONBytes_ErrorAsString(t *testing.T) {
	js := []byte(`{"error":"rate limited"}`)
	assert.Equal(t, "rate limited", detectErrorMessageFromJSONBytes(js))
}

func TestDetectErrorMessageFromJSONBytes_NotJSONObject(t *testing.T) {
	assert.Empty(t, detectErrorMessageFromJSONBytes([]byte("plain text error")))
	assert.Empty(t, detectErrorMessageFromJSONBytes([]byte("")))
}

func TestDetectErrorMessageFromJSONBytes_NoErrorField(t *testing.T) {
	js := []byte(`{"choices":[{"text":"hello"}]}`)
	assert.Empty(t, detectErrorMessageFromJSONBytes(js))
}

func TestDetectErrorMessageFromJSONBytes_EmptyErrorObject(t *testing.T) {
	// error 存在但 message 为空，应降级为通用提示
	js := []byte(`{"error":{}}`)
	got := detectErrorMessageFromJSONBytes(js)
	assert.NotEmpty(t, got, "存在 error 字段即便空也应返回非空占位")
}

// ===========================================================================
// readTestResponseBody
// ===========================================================================

func TestReadTestResponseBody_NonStreamReadsAll(t *testing.T) {
	body := io.NopCloser(bytes.NewBufferString("full response body"))
	got, err := readTestResponseBody(body, false)
	require.NoError(t, err)
	assert.Equal(t, "full response body", string(got))
}

func TestReadTestResponseBody_StreamCapsAt8KB(t *testing.T) {
	// 构造 20KB 流数据，流式读取必须被限制到 8KB
	big := bytes.Repeat([]byte("a"), 20*1024)
	body := io.NopCloser(bytes.NewBuffer(big))
	got, err := readTestResponseBody(body, true)
	require.NoError(t, err)
	assert.Equal(t, 8*1024, len(got), "流式响应读取上限应为 8KB")
}
