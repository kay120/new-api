package model

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupMiniRedis 启动一个内存 Redis，替换 common.RDB 并在测试结束时恢复。
func setupMiniRedis(t *testing.T) *miniredis.Miniredis {
	t.Helper()
	mr := miniredis.RunT(t)
	origRDB := common.RDB
	origEnabled := common.RedisEnabled
	common.RDB = redis.NewClient(&redis.Options{Addr: mr.Addr()})
	common.RedisEnabled = true
	t.Cleanup(func() {
		common.RDB = origRDB
		common.RedisEnabled = origEnabled
	})
	return mr
}

// ===========================================================================
// CheckChannelRateLimit — guard 分支
// ===========================================================================

func TestCheckChannelRateLimit_RedisDisabled(t *testing.T) {
	orig := common.RedisEnabled
	common.RedisEnabled = false
	t.Cleanup(func() { common.RedisEnabled = orig })

	assert.False(t, CheckChannelRateLimit(1, 100, 1000, 10000),
		"Redis 未启用时总是放行")
}

func TestCheckChannelRateLimit_AllLimitsZero(t *testing.T) {
	setupMiniRedis(t)

	assert.False(t, CheckChannelRateLimit(1, 0, 0, 0),
		"三个 limit 全为 0 时不做任何检查")
}

// ===========================================================================
// RPM 限流
// ===========================================================================

func TestCheckChannelRateLimit_RPM_BelowLimit(t *testing.T) {
	setupMiniRedis(t)

	// RPM 限制 5，调用第一次应放行并计入 1
	assert.False(t, CheckChannelRateLimit(10, 5, 0, 0))

	val, _ := common.RDB.Get(context.Background(), "ch_rpm:10").Int64()
	assert.EqualValues(t, 1, val)
}

func TestCheckChannelRateLimit_RPM_AtLimit(t *testing.T) {
	mr := setupMiniRedis(t)

	// 预先设值到限额
	mr.Set("ch_rpm:11", "5")

	assert.True(t, CheckChannelRateLimit(11, 5, 0, 0),
		"当前值等于 RPM 限额应被限流")

	// 被限流时不应再 INCR
	val, _ := common.RDB.Get(context.Background(), "ch_rpm:11").Int64()
	assert.EqualValues(t, 5, val, "限流命中时不应继续递增")
}

func TestCheckChannelRateLimit_RPM_SetsTTL(t *testing.T) {
	setupMiniRedis(t)

	CheckChannelRateLimit(12, 100, 0, 0)

	ttl := common.RDB.TTL(context.Background(), "ch_rpm:12").Val()
	assert.Greater(t, ttl.Seconds(), 0.0, "RPM key 必须有 TTL 以避免永不过期")
	assert.LessOrEqual(t, ttl.Seconds(), 60.0)
}

// ===========================================================================
// TPM 限流
// ===========================================================================

func TestCheckChannelRateLimit_TPM_AtLimit(t *testing.T) {
	mr := setupMiniRedis(t)

	mr.Set("ch_tpm:20", "10000")
	assert.True(t, CheckChannelRateLimit(20, 0, 10000, 0))
}

func TestCheckChannelRateLimit_TPM_BelowLimit_NoIncrementFromCheck(t *testing.T) {
	setupMiniRedis(t)

	// TPM 由 RecordChannelTokenUsage 记，CheckChannelRateLimit 只读判断
	assert.False(t, CheckChannelRateLimit(21, 0, 10000, 0))

	exists, _ := common.RDB.Exists(context.Background(), "ch_tpm:21").Result()
	assert.EqualValues(t, 0, exists, "Check 操作不应写入 TPM key")
}

// ===========================================================================
// 每日 Token 限流
// ===========================================================================

func TestCheckChannelRateLimit_DailyToken_AtLimit(t *testing.T) {
	mr := setupMiniRedis(t)

	today := time.Now().Format("20060102")
	mr.Set(fmt.Sprintf("ch_daily:30:%s", today), "500000")

	assert.True(t, CheckChannelRateLimit(30, 0, 0, 500000))
}

func TestCheckChannelRateLimit_DailyToken_BelowLimit(t *testing.T) {
	setupMiniRedis(t)

	assert.False(t, CheckChannelRateLimit(31, 0, 0, 1000000))
}

// ===========================================================================
// RecordChannelTokenUsage
// ===========================================================================

func TestRecordChannelTokenUsage_UpdatesTPMAndDaily(t *testing.T) {
	setupMiniRedis(t)
	ctx := context.Background()

	RecordChannelTokenUsage(40, 2500)
	RecordChannelTokenUsage(40, 1500)

	tpm, _ := common.RDB.Get(ctx, "ch_tpm:40").Int64()
	assert.EqualValues(t, 4000, tpm)

	today := time.Now().Format("20060102")
	daily, _ := common.RDB.Get(ctx, fmt.Sprintf("ch_daily:40:%s", today)).Int64()
	assert.EqualValues(t, 4000, daily)

	tpmTTL := common.RDB.TTL(ctx, "ch_tpm:40").Val()
	assert.Greater(t, tpmTTL.Seconds(), 0.0, "TPM key 必须有 TTL")

	dailyTTL := common.RDB.TTL(ctx, fmt.Sprintf("ch_daily:40:%s", today)).Val()
	require.Greater(t, dailyTTL.Seconds(), 23*60*60.0, "每日 key TTL 必须覆盖到次日")
}

func TestRecordChannelTokenUsage_RedisDisabled(t *testing.T) {
	orig := common.RedisEnabled
	common.RedisEnabled = false
	t.Cleanup(func() { common.RedisEnabled = orig })

	// 不应 panic，即使 RDB 为 nil（依赖 guard 提前返回）
	RecordChannelTokenUsage(41, 100)
}

func TestRecordChannelTokenUsage_ZeroTokens(t *testing.T) {
	setupMiniRedis(t)

	RecordChannelTokenUsage(42, 0)

	exists, _ := common.RDB.Exists(context.Background(), "ch_tpm:42").Result()
	assert.EqualValues(t, 0, exists, "tokens=0 不应写入任何 key")
}
