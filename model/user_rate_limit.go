package model

import (
	"context"
	"fmt"
	"time"

	"github.com/QuantumNous/new-api/common"
)

// CheckUserDailyTokenLimit 检查用户每日 Token 上限是否超出
func CheckUserDailyTokenLimit(userId, limit int) bool {
	if !common.RedisEnabled || limit <= 0 {
		return false
	}
	today := time.Now().Format("20060102")
	key := fmt.Sprintf("user_daily:%d:%s", userId, today)
	val, _ := common.RDB.Get(context.Background(), key).Int64()
	return val >= int64(limit)
}

// RecordUserDailyTokens 记录用户当日 Token 用量
func RecordUserDailyTokens(userId, tokens int) {
	if !common.RedisEnabled || tokens <= 0 {
		return
	}
	ctx := context.Background()
	today := time.Now().Format("20060102")
	key := fmt.Sprintf("user_daily:%d:%s", userId, today)
	common.RDB.IncrBy(ctx, key, int64(tokens))
	common.RDB.Expire(ctx, key, 25*time.Hour)
}

// GetUserDailyTokens 获取用户当日 Token 用量（用于前端进度展示）
func GetUserDailyTokens(userId int) int64 {
	if !common.RedisEnabled {
		return 0
	}
	today := time.Now().Format("20060102")
	key := fmt.Sprintf("user_daily:%d:%s", userId, today)
	val, _ := common.RDB.Get(context.Background(), key).Int64()
	return val
}
