package service

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// seedChannelFull 写一个状态与模型可控的 Channel（findChannelFromAllowedList 用到）。
func seedChannelFull(t *testing.T, id int, status int, models string) {
	t.Helper()
	ch := &model.Channel{
		Id:     id,
		Name:   "test_ch",
		Key:    "sk-test",
		Status: status,
		Models: models,
	}
	require.NoError(t, model.DB.Create(ch).Error)
}

// ===========================================================================
// findChannelFromAllowedList — 白名单解析与匹配
// ===========================================================================

func TestFindChannelFromAllowedList_EmptyString(t *testing.T) {
	ch, err := findChannelFromAllowedList("", "gpt-4")
	require.NoError(t, err)
	assert.Nil(t, ch)
}

func TestFindChannelFromAllowedList_AllEntriesMalformed(t *testing.T) {
	// 无冒号、非数字 ID、空 entry 全部应被跳过
	ch, err := findChannelFromAllowedList("abc,1,:gpt-4,not_a_number:gpt-4", "gpt-4")
	require.NoError(t, err)
	assert.Nil(t, ch)
}

func TestFindChannelFromAllowedList_ExactModelMatch(t *testing.T) {
	truncate(t)
	seedChannelFull(t, 701, 1, "gpt-4,gpt-3.5")

	ch, err := findChannelFromAllowedList("701:gpt-4", "gpt-4")
	require.NoError(t, err)
	require.NotNil(t, ch)
	assert.Equal(t, 701, ch.Id)
}

func TestFindChannelFromAllowedList_WildcardMatch(t *testing.T) {
	truncate(t)
	seedChannelFull(t, 702, 1, "claude-3-opus")

	ch, err := findChannelFromAllowedList("702:*", "any-model-works")
	require.NoError(t, err)
	require.NotNil(t, ch)
	assert.Equal(t, 702, ch.Id)
}

func TestFindChannelFromAllowedList_NoMatchingModel(t *testing.T) {
	truncate(t)
	seedChannelFull(t, 703, 1, "gpt-4")

	// 白名单里没 claude，也没通配符
	ch, err := findChannelFromAllowedList("703:gpt-4", "claude-3")
	require.NoError(t, err)
	assert.Nil(t, ch, "模型未列入白名单应返回 nil")
}

func TestFindChannelFromAllowedList_SkipsDisabledChannel(t *testing.T) {
	truncate(t)
	// Status=2（手动禁用）或 3（自动熔断）都应跳过
	seedChannelFull(t, 704, 2, "gpt-4")

	ch, err := findChannelFromAllowedList("704:gpt-4", "gpt-4")
	require.NoError(t, err)
	assert.Nil(t, ch, "被禁用的渠道不应被选中")
}

func TestFindChannelFromAllowedList_FallsBackToNextCandidate(t *testing.T) {
	truncate(t)
	seedChannelFull(t, 705, 2, "gpt-4") // 禁用
	seedChannelFull(t, 706, 1, "gpt-4") // 启用

	ch, err := findChannelFromAllowedList("705:gpt-4,706:gpt-4", "gpt-4")
	require.NoError(t, err)
	require.NotNil(t, ch)
	assert.Equal(t, 706, ch.Id, "首个禁用时应降级到下一个候选")
}

func TestFindChannelFromAllowedList_MixedWildcardAndExact(t *testing.T) {
	truncate(t)
	seedChannelFull(t, 707, 1, "gpt-4") // 精确匹配
	seedChannelFull(t, 708, 1, "claude-3")

	// 同时有 "707:gpt-4"（精确）和 "708:*"（通配符），请求 gpt-4 时两个都是候选
	ch, err := findChannelFromAllowedList("707:gpt-4,708:*", "gpt-4")
	require.NoError(t, err)
	require.NotNil(t, ch)
	assert.Contains(t, []int{707, 708}, ch.Id)
}

func TestFindChannelFromAllowedList_WhitespaceTolerant(t *testing.T) {
	truncate(t)
	seedChannelFull(t, 709, 1, "gpt-4")

	// 空格应被容忍
	ch, err := findChannelFromAllowedList("  709 : gpt-4 ", "gpt-4")
	require.NoError(t, err)
	require.NotNil(t, ch)
	assert.Equal(t, 709, ch.Id)
}

func TestFindChannelFromAllowedList_ChannelNotFoundInDB(t *testing.T) {
	truncate(t)
	// 没 seed 任何 channel
	ch, err := findChannelFromAllowedList("999:gpt-4", "gpt-4")
	require.NoError(t, err)
	assert.Nil(t, ch, "DB 中无此渠道应优雅降级为 nil")
}

// ensure model package is kept even if all tests above skip import
var _ = common.UsingSQLite
