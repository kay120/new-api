package model

import (
	"context"
	"fmt"
	"time"

	"github.com/QuantumNous/new-api/common"
)

// CheckChannelRateLimit 检查渠道级限流，返回 true 表示被限流
func CheckChannelRateLimit(channelId, rpmLimit, tpmLimit, dailyLimit int) bool {
	if !common.RedisEnabled {
		return false
	}
	if rpmLimit <= 0 && tpmLimit <= 0 && dailyLimit <= 0 {
		return false
	}
	ctx := context.Background()

	// RPM 检查：先 GET 判断是否超限，未超限才 INCR
	if rpmLimit > 0 {
		key := fmt.Sprintf("ch_rpm:%d", channelId)
		val, _ := common.RDB.Get(ctx, key).Int64()
		if val >= int64(rpmLimit) {
			return true
		}
		// 未超限，递增并确保有 TTL
		common.RDB.Incr(ctx, key)
		common.RDB.Expire(ctx, key, time.Minute) // 每次刷新 TTL，避免 key 永不过期
	}

	// TPM 检查
	if tpmLimit > 0 {
		key := fmt.Sprintf("ch_tpm:%d", channelId)
		val, _ := common.RDB.Get(ctx, key).Int64()
		if val >= int64(tpmLimit) {
			return true
		}
	}

	// 每日 Token 上限
	if dailyLimit > 0 {
		today := time.Now().Format("20060102")
		key := fmt.Sprintf("ch_daily:%d:%s", channelId, today)
		val, _ := common.RDB.Get(ctx, key).Int64()
		if val >= int64(dailyLimit) {
			return true
		}
	}

	return false
}

// RecordChannelTokenUsage 请求完成后记录渠道 token 使用量
func RecordChannelTokenUsage(channelId, tokens int) {
	if !common.RedisEnabled || tokens <= 0 {
		return
	}
	ctx := context.Background()

	// TPM — IncrBy 并确保有 TTL
	tpmKey := fmt.Sprintf("ch_tpm:%d", channelId)
	common.RDB.IncrBy(ctx, tpmKey, int64(tokens))
	common.RDB.Expire(ctx, tpmKey, time.Minute)

	// 每日
	today := time.Now().Format("20060102")
	dailyKey := fmt.Sprintf("ch_daily:%d:%s", channelId, today)
	common.RDB.IncrBy(ctx, dailyKey, int64(tokens))
	common.RDB.Expire(ctx, dailyKey, 25*time.Hour)
}
