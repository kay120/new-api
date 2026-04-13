package model

import (
	"context"
	"fmt"
	"strconv"
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

	// RPM 检查
	if rpmLimit > 0 {
		key := fmt.Sprintf("ch_rpm:%d", channelId)
		val, err := common.RDB.Incr(ctx, key).Result()
		if err == nil && val == 1 {
			common.RDB.Expire(ctx, key, time.Minute)
		}
		if val > int64(rpmLimit) {
			return true
		}
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
	tokensStr := strconv.FormatInt(int64(tokens), 10)
	_ = tokensStr

	// TPM
	tpmKey := fmt.Sprintf("ch_tpm:%d", channelId)
	val, err := common.RDB.IncrBy(ctx, tpmKey, int64(tokens)).Result()
	if err == nil && val == int64(tokens) {
		common.RDB.Expire(ctx, tpmKey, time.Minute)
	}

	// 每日
	today := time.Now().Format("20060102")
	dailyKey := fmt.Sprintf("ch_daily:%d:%s", channelId, today)
	val, err = common.RDB.IncrBy(ctx, dailyKey, int64(tokens)).Result()
	if err == nil && val == int64(tokens) {
		common.RDB.Expire(ctx, dailyKey, 25*time.Hour)
	}
}
