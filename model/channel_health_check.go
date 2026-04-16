package model

import (
	"github.com/QuantumNous/new-api/common"
)

type ChannelHealthCheck struct {
	Id           int    `json:"id" gorm:"primaryKey;autoIncrement"`
	ChannelId    int    `json:"channel_id" gorm:"index;not null"`
	Success      bool   `json:"success"`
	ResponseTime int64  `json:"response_time" gorm:"default:0"`
	FirstTokenMs int64  `json:"first_token_ms" gorm:"default:0"`
	ErrorMessage string `json:"error_message" gorm:"type:text"`
	CreatedAt    int64  `json:"created_at" gorm:"bigint;index"`
}

func (ChannelHealthCheck) TableName() string {
	return "channel_health_checks"
}

// RecordHealthCheck 记录一次健康检查结果
func RecordHealthCheck(channelId int, success bool, responseTimeMs int64, firstTokenMs int64, errorMsg string) {
	check := &ChannelHealthCheck{
		ChannelId:    channelId,
		Success:      success,
		ResponseTime: responseTimeMs,
		FirstTokenMs: firstTokenMs,
		ErrorMessage: errorMsg,
		CreatedAt:    common.GetTimestamp(),
	}
	if err := LOG_DB.Create(check).Error; err != nil {
		common.SysError("failed to record health check: " + err.Error())
	}
}

// ChannelHealthStat 渠道健康统计
type ChannelHealthStat struct {
	ChannelId    int     `json:"channel_id"`
	TotalChecks  int64   `json:"total_checks"`
	SuccessCount int64   `json:"success_count"`
	Availability float64 `json:"availability"`
	AvgResponse  float64 `json:"avg_response"`
	AvgFirstToken float64 `json:"avg_first_token"`
}

// GetChannelHealthStats 获取所有渠道的健康统计
func GetChannelHealthStats(startTimestamp int64) ([]ChannelHealthStat, error) {
	var stats []ChannelHealthStat

	successExpr := "1"
	if common.UsingPostgreSQL {
		successExpr = "CASE WHEN success = true THEN 1 ELSE 0 END"
	} else {
		successExpr = "CASE WHEN success = 1 THEN 1 ELSE 0 END"
	}

	query := LOG_DB.Table("channel_health_checks").
		Select("channel_id, count(*) as total_checks, sum("+successExpr+") as success_count, "+
			"CAST(sum("+successExpr+") AS FLOAT) / CAST(count(*) AS FLOAT) * 100 as availability, "+
			"avg(CASE WHEN response_time > 0 THEN response_time ELSE NULL END) as avg_response, "+
			"avg(CASE WHEN first_token_ms > 0 THEN first_token_ms ELSE NULL END) as avg_first_token").
		Where("created_at >= ?", startTimestamp).
		Group("channel_id")

	if err := query.Scan(&stats).Error; err != nil {
		return nil, err
	}
	return stats, nil
}

// GetRecentHealthChecks 获取某渠道最近 N 次健康检查
func GetRecentHealthChecks(channelId int, limit int) ([]ChannelHealthCheck, error) {
	var checks []ChannelHealthCheck
	err := LOG_DB.Where("channel_id = ?", channelId).
		Order("created_at desc").
		Limit(limit).
		Find(&checks).Error
	return checks, err
}

// CleanupOldHealthChecks 清理超过指定天数的旧记录
func CleanupOldHealthChecks(retentionDays int) error {
	threshold := common.GetTimestamp() - int64(retentionDays*86400)
	return LOG_DB.Where("created_at < ?", threshold).Delete(&ChannelHealthCheck{}).Error
}
