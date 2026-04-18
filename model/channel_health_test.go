package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func seedLogRow(t *testing.T, logType int, channelId int, createdAt int64, useTime int) {
	t.Helper()
	row := &Log{
		UserId:    1,
		Type:      logType,
		ChannelId: channelId,
		CreatedAt: createdAt,
		UseTime:   useTime,
		ModelName: "test-model",
	}
	require.NoError(t, LOG_DB.Create(row).Error)
}

func TestQueryChannelHealthSummary_AggregatesByChannel(t *testing.T) {
	require.NoError(t, DB.AutoMigrate(&Channel{}, &Log{}))
	DB.Exec("DELETE FROM channels")
	LOG_DB.Exec("DELETE FROM logs")

	// 准备两个渠道
	require.NoError(t, DB.Create(&Channel{Id: 901, Type: 1, Name: "OpenAI", Status: 1}).Error)
	require.NoError(t, DB.Create(&Channel{Id: 902, Type: 14, Name: "Claude", Status: 2}).Error)

	now := time.Now().Unix()
	// ch901: 3 成功 + 1 错误，use_time 1,2,3,4 秒
	seedLogRow(t, LogTypeConsume, 901, now-60, 1)
	seedLogRow(t, LogTypeConsume, 901, now-50, 2)
	seedLogRow(t, LogTypeConsume, 901, now-40, 3)
	seedLogRow(t, LogTypeError, 901, now-30, 4)
	// ch902: 1 成功
	seedLogRow(t, LogTypeConsume, 902, now-20, 2)

	rows, err := QueryChannelHealthSummary(1)
	require.NoError(t, err)
	require.Len(t, rows, 2)

	// 排序是按 total 降序，所以 ch901 在前
	assert.Equal(t, 901, rows[0].ChannelId)
	assert.Equal(t, "OpenAI", rows[0].ChannelName)
	assert.EqualValues(t, 4, rows[0].Total)
	assert.EqualValues(t, 3, rows[0].Successes)
	assert.EqualValues(t, 1, rows[0].Errors)
	assert.InDelta(t, 0.25, rows[0].ErrorRate, 0.001)
	assert.InDelta(t, 2.5, rows[0].AvgUseTime, 0.001)
	assert.EqualValues(t, 4, rows[0].MaxUseTime)
	assert.Equal(t, 1, rows[0].Status)
	assert.Equal(t, 1, rows[0].ChannelType)

	assert.Equal(t, 902, rows[1].ChannelId)
	assert.EqualValues(t, 1, rows[1].Total)
	assert.EqualValues(t, 1, rows[1].Successes)
	assert.Equal(t, "Claude", rows[1].ChannelName)
	assert.Equal(t, 2, rows[1].Status)
}

func TestQueryChannelHealthSummary_RespectsTimeWindow(t *testing.T) {
	require.NoError(t, DB.AutoMigrate(&Channel{}, &Log{}))
	DB.Exec("DELETE FROM channels")
	LOG_DB.Exec("DELETE FROM logs")

	require.NoError(t, DB.Create(&Channel{Id: 903, Type: 1, Name: "OldCh", Status: 1}).Error)

	now := time.Now().Unix()
	seedLogRow(t, LogTypeConsume, 903, now-10, 1)         // 窗口内
	seedLogRow(t, LogTypeConsume, 903, now-2*3600, 1)     // 窗口外（2h 前 vs 1h 窗）

	rows, err := QueryChannelHealthSummary(1)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.EqualValues(t, 1, rows[0].Total, "窗口外日志应被过滤")
}

func TestQueryChannelHealthSummary_ZeroHoursDefaultsTo24(t *testing.T) {
	require.NoError(t, DB.AutoMigrate(&Channel{}, &Log{}))
	DB.Exec("DELETE FROM channels")
	LOG_DB.Exec("DELETE FROM logs")

	require.NoError(t, DB.Create(&Channel{Id: 904, Type: 1, Name: "Ch", Status: 1}).Error)

	now := time.Now().Unix()
	seedLogRow(t, LogTypeConsume, 904, now-12*3600, 1) // 12h 前 → 24h 默认窗口内

	rows, err := QueryChannelHealthSummary(0)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.EqualValues(t, 1, rows[0].Total)
}
