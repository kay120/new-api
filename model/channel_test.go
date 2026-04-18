package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// truncateChannels 清空 channel + ability 表以保证每个测试独立。
func truncateChannels(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		DB.Exec("DELETE FROM channels")
		DB.Exec("DELETE FROM abilities")
	})
}

func buildTestChannel(id int, models string, groups string) *Channel {
	return &Channel{
		Id:       id,
		Type:     1,
		Name:     "test_ch",
		Key:      "sk-test",
		Status:   1,
		Models:   models,
		Group:    groups,
		Priority: func(v int64) *int64 { return &v }(100),
		Weight:   func(v uint) *uint { return &v }(10),
	}
}

// ===========================================================================
// Insert / GetChannelById / Delete
// ===========================================================================

func TestChannel_Insert_CreatesAbilities(t *testing.T) {
	truncateChannels(t)
	ch := buildTestChannel(801, "gpt-4,gpt-3.5", "default,vip")

	require.NoError(t, ch.Insert())

	got, err := GetChannelById(801, true)
	require.NoError(t, err)
	assert.Equal(t, "test_ch", got.Name)
	assert.Equal(t, "gpt-4,gpt-3.5", got.Models)

	// models × groups 笛卡尔积生成 ability 记录：2 × 2 = 4
	var count int64
	require.NoError(t, DB.Model(&Ability{}).Where("channel_id = ?", 801).Count(&count).Error)
	assert.EqualValues(t, 4, count, "每个 model/group 组合应产生一条 ability")
}

func TestChannel_Delete_RemovesAbilities(t *testing.T) {
	truncateChannels(t)
	ch := buildTestChannel(802, "claude-3", "default")
	require.NoError(t, ch.Insert())

	require.NoError(t, ch.Delete())

	_, err := GetChannelById(802, true)
	assert.Error(t, err, "删除后 GetChannelById 应返回错误")

	var count int64
	require.NoError(t, DB.Model(&Ability{}).Where("channel_id = ?", 802).Count(&count).Error)
	assert.EqualValues(t, 0, count, "渠道删除时必须连同 ability 清理")
}

func TestGetChannelById_OmitsKeyWhenRequested(t *testing.T) {
	truncateChannels(t)
	ch := buildTestChannel(803, "gpt-4", "default")
	require.NoError(t, ch.Insert())

	got, err := GetChannelById(803, false)
	require.NoError(t, err)
	assert.Empty(t, got.Key, "selectAll=false 时 key 字段应被 Omit")
}

// ===========================================================================
// Update
// ===========================================================================

func TestChannel_Update_RegeneratesAbilities(t *testing.T) {
	truncateChannels(t)
	ch := buildTestChannel(804, "gpt-4", "default")
	require.NoError(t, ch.Insert())

	// 原本 1 个 ability；改模型列表为 2 个
	ch.Models = "gpt-4,claude-3"
	require.NoError(t, ch.Update())

	var count int64
	require.NoError(t, DB.Model(&Ability{}).Where("channel_id = ?", 804).Count(&count).Error)
	assert.EqualValues(t, 2, count, "Update 后 ability 应覆盖新的 models")
}

// ===========================================================================
// BatchInsertChannels / BatchDeleteChannels
// ===========================================================================

func TestBatchInsertChannels_InsertsMultiple(t *testing.T) {
	truncateChannels(t)
	chs := []Channel{
		*buildTestChannel(810, "gpt-4", "default"),
		*buildTestChannel(811, "gpt-4", "default"),
		*buildTestChannel(812, "claude-3", "vip"),
	}
	require.NoError(t, BatchInsertChannels(chs))

	var total int64
	require.NoError(t, DB.Model(&Channel{}).Where("id IN ?", []int{810, 811, 812}).Count(&total).Error)
	assert.EqualValues(t, 3, total)
}

func TestBatchDeleteChannels_RemovesByIds(t *testing.T) {
	truncateChannels(t)
	chs := []Channel{
		*buildTestChannel(820, "gpt-4", "default"),
		*buildTestChannel(821, "gpt-4", "default"),
		*buildTestChannel(822, "gpt-4", "default"),
	}
	require.NoError(t, BatchInsertChannels(chs))

	require.NoError(t, BatchDeleteChannels([]int{820, 821}))

	var remaining int64
	require.NoError(t, DB.Model(&Channel{}).Where("id IN ?", []int{820, 821, 822}).Count(&remaining).Error)
	assert.EqualValues(t, 1, remaining, "只应保留 822")
}

// ===========================================================================
// Count helpers
// ===========================================================================

func TestCountAllChannels(t *testing.T) {
	truncateChannels(t)
	require.NoError(t, buildTestChannel(830, "gpt-4", "default").Insert())
	require.NoError(t, buildTestChannel(831, "gpt-4", "default").Insert())

	n, err := CountAllChannels()
	require.NoError(t, err)
	assert.EqualValues(t, 2, n)
}

func TestCountChannelsByType(t *testing.T) {
	truncateChannels(t)
	ch1 := buildTestChannel(840, "gpt-4", "default")
	ch1.Type = 1
	ch2 := buildTestChannel(841, "gpt-4", "default")
	ch2.Type = 14
	require.NoError(t, ch1.Insert())
	require.NoError(t, ch2.Insert())

	n, err := CountChannelsByType(1)
	require.NoError(t, err)
	assert.EqualValues(t, 1, n)

	n, err = CountChannelsByType(14)
	require.NoError(t, err)
	assert.EqualValues(t, 1, n)
}

// ===========================================================================
// GetChannelsByIds
// ===========================================================================

func TestGetChannelsByIds_ReturnsOrderedSubset(t *testing.T) {
	truncateChannels(t)
	require.NoError(t, buildTestChannel(850, "gpt-4", "default").Insert())
	require.NoError(t, buildTestChannel(851, "gpt-4", "default").Insert())
	require.NoError(t, buildTestChannel(852, "gpt-4", "default").Insert())

	chs, err := GetChannelsByIds([]int{850, 852})
	require.NoError(t, err)
	require.Len(t, chs, 2)

	ids := []int{chs[0].Id, chs[1].Id}
	assert.ElementsMatch(t, []int{850, 852}, ids)
}
