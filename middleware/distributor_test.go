package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ===========================================================================
// extractModelNameFromGeminiPath — 纯字符串解析
// ===========================================================================

func TestExtractModelNameFromGeminiPath_HappyPath(t *testing.T) {
	cases := []struct {
		path     string
		expected string
	}{
		{"/v1/models/gemini-1.5-pro:generateContent", "gemini-1.5-pro"},
		{"/v1/models/gemini-1.5-pro:streamGenerateContent", "gemini-1.5-pro"},
		{"/v1beta/models/gemini-2.0-flash:generateContent", "gemini-2.0-flash"},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			assert.Equal(t, tc.expected, extractModelNameFromGeminiPath(tc.path))
		})
	}
}

func TestExtractModelNameFromGeminiPath_NoModelsSegment(t *testing.T) {
	assert.Empty(t, extractModelNameFromGeminiPath("/v1/chat/completions"))
}

func TestExtractModelNameFromGeminiPath_NoColonSuffix(t *testing.T) {
	// 没有 ":"，应返回 /models/ 后的所有内容
	assert.Equal(t, "gemini-pro", extractModelNameFromGeminiPath("/v1/models/gemini-pro"))
}

func TestExtractModelNameFromGeminiPath_EmptyAfterPrefix(t *testing.T) {
	assert.Empty(t, extractModelNameFromGeminiPath("/v1/models/"))
}

func TestExtractModelNameFromGeminiPath_EmptyInput(t *testing.T) {
	assert.Empty(t, extractModelNameFromGeminiPath(""))
}

// ===========================================================================
// SetupContextForSelectedChannel
// ===========================================================================

func newDistributorTestContext() *gin.Context {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c
}

func TestSetupContextForSelectedChannel_NilChannel(t *testing.T) {
	c := newDistributorTestContext()
	err := SetupContextForSelectedChannel(c, nil, "gpt-4")
	require.NotNil(t, err, "nil channel 应返回错误")
	assert.Equal(t, "gpt-4", c.GetString("original_model"))
}

func TestSetupContextForSelectedChannel_InjectsChannelKeys(t *testing.T) {
	c := newDistributorTestContext()
	ch := &model.Channel{
		Id:   42,
		Name: "openai-pri",
		Type: 1,
		Key:  "sk-x",
	}
	err := SetupContextForSelectedChannel(c, ch, "gpt-4")
	require.Nil(t, err)

	assert.Equal(t, 42, common.GetContextKeyInt(c, constant.ContextKeyChannelId))
	assert.Equal(t, "openai-pri", common.GetContextKeyString(c, constant.ContextKeyChannelName))
	assert.Equal(t, "sk-x", common.GetContextKeyString(c, constant.ContextKeyChannelKey))
	assert.Equal(t, "gpt-4", c.GetString("original_model"))
	assert.False(t, common.GetContextKeyBool(c, constant.ContextKeyChannelIsMultiKey),
		"单 key 渠道时 IsMultiKey 必须被显式设为 false")
}

func TestSetupContextForSelectedChannel_AzureInjectsApiVersion(t *testing.T) {
	c := newDistributorTestContext()
	ch := &model.Channel{
		Id:    43,
		Name:  "azure-ch",
		Type:  constant.ChannelTypeAzure,
		Key:   "sk-azure",
		Other: "2024-02-15",
	}
	err := SetupContextForSelectedChannel(c, ch, "gpt-4")
	require.Nil(t, err)
	assert.Equal(t, "2024-02-15", c.GetString("api_version"),
		"Azure 渠道必须把 channel.Other 注入为 api_version")
}

func TestSetupContextForSelectedChannel_GeminiInjectsApiVersion(t *testing.T) {
	c := newDistributorTestContext()
	ch := &model.Channel{
		Id:    44,
		Name:  "gemini-ch",
		Type:  constant.ChannelTypeGemini,
		Key:   "sk-gemini",
		Other: "v1beta",
	}
	err := SetupContextForSelectedChannel(c, ch, "gemini-1.5-pro")
	require.Nil(t, err)
	assert.Equal(t, "v1beta", c.GetString("api_version"))
}

func TestSetupContextForSelectedChannel_VertexAiInjectsRegion(t *testing.T) {
	c := newDistributorTestContext()
	ch := &model.Channel{
		Id:    45,
		Name:  "vertex-ch",
		Type:  constant.ChannelTypeVertexAi,
		Key:   "sk-vertex",
		Other: "us-central1",
	}
	err := SetupContextForSelectedChannel(c, ch, "gemini-pro")
	require.Nil(t, err)
	assert.Equal(t, "us-central1", c.GetString("region"),
		"Vertex AI 渠道 Other 应作为 region 注入")
}

func TestSetupContextForSelectedChannel_CozeInjectsBotId(t *testing.T) {
	c := newDistributorTestContext()
	ch := &model.Channel{
		Id:    46,
		Name:  "coze-ch",
		Type:  constant.ChannelTypeCoze,
		Key:   "sk-coze",
		Other: "bot_12345",
	}
	err := SetupContextForSelectedChannel(c, ch, "coze-bot")
	require.Nil(t, err)
	assert.Equal(t, "bot_12345", c.GetString("bot_id"))
}

func TestSetupContextForSelectedChannel_OtherTypeNoExtraInjection(t *testing.T) {
	c := newDistributorTestContext()
	ch := &model.Channel{
		Id:    47,
		Name:  "openai-ch",
		Type:  constant.ChannelTypeOpenAI,
		Key:   "sk-openai",
		Other: "should-be-ignored",
	}
	err := SetupContextForSelectedChannel(c, ch, "gpt-4")
	require.Nil(t, err)
	assert.Empty(t, c.GetString("api_version"))
	assert.Empty(t, c.GetString("region"))
	assert.Empty(t, c.GetString("bot_id"))
}
