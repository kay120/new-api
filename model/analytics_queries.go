package model

import (
	"time"
)

// GetActiveUserCount 返回有过 consume 日志的独立用户数
func GetActiveUserCount() int {
	var count int64
	LOG_DB.Table("logs").
		Where("type = ?", LogTypeConsume).
		Distinct("user_id").
		Count(&count)
	return int(count)
}

// GetTodayRequestCount 返回今日 consume 请求总数
func GetTodayRequestCount() int64 {
	todayStart := time.Now().Truncate(24 * time.Hour).Unix()
	var count int64
	LOG_DB.Table("logs").
		Where("type = ? AND created_at >= ?", LogTypeConsume, todayStart).
		Count(&count)
	return count
}
