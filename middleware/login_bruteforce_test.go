package middleware

import (
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
)

// resetLoginFailMem 清空内存 fallback 状态，保证每个测试独立。
func resetLoginFailMem(t *testing.T) {
	t.Helper()
	loginFailMemMu.Lock()
	loginFailMem = make(map[string]*loginFailEntry)
	loginFailMemMu.Unlock()
}

func TestLoginBruteForce_MemoryFallback_BelowThreshold(t *testing.T) {
	common.RedisEnabled = false
	resetLoginFailMem(t)

	for i := 0; i < loginFailThreshold-1; i++ {
		RecordLoginFailure("alice")
	}
	assert.False(t, LoginBruteForceLocked("alice"), "%d 次失败未达阈值，不应锁定", loginFailThreshold-1)
}

func TestLoginBruteForce_MemoryFallback_LocksAtThreshold(t *testing.T) {
	common.RedisEnabled = false
	resetLoginFailMem(t)

	for i := 0; i < loginFailThreshold; i++ {
		RecordLoginFailure("bob")
	}
	assert.True(t, LoginBruteForceLocked("bob"), "达到阈值应立即锁定")

	// 锁定期内即使无更多失败尝试也保持锁定
	loginFailMemMu.Lock()
	entry := loginFailMem[loginFailMemKey("bob")]
	loginFailMemMu.Unlock()
	assert.True(t, entry.lockedUntil.After(time.Now()))
}

func TestLoginBruteForce_MemoryFallback_ClearResets(t *testing.T) {
	common.RedisEnabled = false
	resetLoginFailMem(t)

	for i := 0; i < loginFailThreshold; i++ {
		RecordLoginFailure("carol")
	}
	assert.True(t, LoginBruteForceLocked("carol"))

	ClearLoginFailure("carol")
	assert.False(t, LoginBruteForceLocked("carol"), "Clear 后立即解锁")
}

func TestLoginBruteForce_MemoryFallback_WindowResets(t *testing.T) {
	common.RedisEnabled = false
	resetLoginFailMem(t)

	// 先记录几次失败
	for i := 0; i < 3; i++ {
		RecordLoginFailure("dave")
	}
	// 强行把 windowStart 设到过期之前
	loginFailMemMu.Lock()
	loginFailMem[loginFailMemKey("dave")].windowStart = time.Now().Add(-time.Duration(loginFailWindowSec+60) * time.Second)
	loginFailMemMu.Unlock()

	// 下一次失败应开启新窗口，计数重置为 1
	RecordLoginFailure("dave")
	loginFailMemMu.Lock()
	entry := loginFailMem[loginFailMemKey("dave")]
	loginFailMemMu.Unlock()
	assert.Equal(t, 1, entry.count, "超出 window 后计数应被重置")
	assert.False(t, LoginBruteForceLocked("dave"))
}

func TestLoginBruteForce_EmptyUsernameIgnored(t *testing.T) {
	common.RedisEnabled = false
	resetLoginFailMem(t)

	RecordLoginFailure("")
	assert.False(t, LoginBruteForceLocked(""))

	loginFailMemMu.Lock()
	mapSize := len(loginFailMem)
	loginFailMemMu.Unlock()
	assert.Equal(t, 0, mapSize, "空 username 不应在 map 中占位")
}

func TestLoginBruteForce_CaseInsensitiveNormalization(t *testing.T) {
	common.RedisEnabled = false
	resetLoginFailMem(t)

	for i := 0; i < loginFailThreshold; i++ {
		RecordLoginFailure("  Eve  ")
	}
	assert.True(t, LoginBruteForceLocked("eve"), "大小写 + 空白应规范化后命中同一条目")
	assert.True(t, LoginBruteForceLocked("EVE"))
}
