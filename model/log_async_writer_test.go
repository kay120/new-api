package model

import (
	"os"
	"sync"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupLogTable(t *testing.T) {
	t.Helper()
	require.NoError(t, DB.AutoMigrate(&Log{}))
	DB.Exec("DELETE FROM logs")
}

func makeLog(i int) *Log {
	return &Log{
		UserId:    1,
		Type:      LogTypeConsume,
		ModelName: "test-model",
		Content:   "req-" + itoa(i),
		CreatedAt: time.Now().Unix(),
	}
}

// ===========================================================================
// 同步路径 — LOG_CONSUME_ASYNC 未设
// ===========================================================================

func TestEnqueueConsumeLog_WritesSyncWhenDisabled(t *testing.T) {
	setupLogTable(t)
	// 确保 async 未启用
	_ = os.Unsetenv("LOG_CONSUME_ASYNC")
	StopConsumeLogWriter()

	returned := enqueueConsumeLog(makeLog(1))
	assert.False(t, returned, "未启用时应立即同步写并返回 false")

	var count int64
	require.NoError(t, DB.Model(&Log{}).Count(&count).Error)
	assert.EqualValues(t, 1, count, "同步路径应立即落库")
}

// ===========================================================================
// 异步路径 — LOG_CONSUME_ASYNC=true
// ===========================================================================

func TestEnqueueConsumeLog_AsyncQueuesAndFlushes(t *testing.T) {
	setupLogTable(t)
	t.Setenv("LOG_CONSUME_ASYNC", "true")
	StopConsumeLogWriter() // 清掉前一个测试可能留下的状态
	StartConsumeLogWriter()

	const n = 50
	for i := 0; i < n; i++ {
		ok := enqueueConsumeLog(makeLog(i))
		assert.True(t, ok, "buffer 有空间时应异步入队")
	}
	// 关闭 channel 并等待 worker 消费完
	StopConsumeLogWriter()

	var count int64
	require.NoError(t, DB.Model(&Log{}).Count(&count).Error)
	assert.EqualValues(t, n, count, "所有入队日志都应落库")
}

func TestStartConsumeLogWriter_NoopIfEnvOff(t *testing.T) {
	t.Setenv("LOG_CONSUME_ASYNC", "")
	StopConsumeLogWriter()
	StartConsumeLogWriter()

	// 未启用时 channel 应为 nil
	assert.Nil(t, consumeLogChan)
}

// ===========================================================================
// buffer 满 fallback
// ===========================================================================

func TestEnqueueConsumeLog_FallbackToSyncOnFullBuffer(t *testing.T) {
	setupLogTable(t)
	t.Setenv("LOG_CONSUME_ASYNC", "true")
	StopConsumeLogWriter()

	// 造一个小 buffer 的实例以便触发满
	consumeLogOnce = sync.Once{}
	consumeLogOnce.Do(func() {
		consumeLogChan = make(chan *Log, 2)
		// 故意不启动 consumer，让 buffer 填满
	})

	ok1 := enqueueConsumeLog(makeLog(1))
	ok2 := enqueueConsumeLog(makeLog(2))
	ok3 := enqueueConsumeLog(makeLog(3)) // 第三条满 → fallback 同步写

	assert.True(t, ok1)
	assert.True(t, ok2)
	assert.False(t, ok3, "buffer 满后应 fallback 同步写并返回 false")

	var count int64
	require.NoError(t, DB.Model(&Log{}).Count(&count).Error)
	assert.EqualValues(t, 1, count, "仅 fallback 的那条落库；前两条还在 channel 里")

	// 清理（不调 Stop 因为没 consumer，会死锁）
	consumeLogChan = nil
	consumeLogOnce = sync.Once{}
}

var _ = common.LogConsumeEnabled
