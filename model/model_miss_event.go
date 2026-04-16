package model

import (
	"github.com/QuantumNous/new-api/common"
	"github.com/bytedance/gopkg/util/gopool"
)

type ModelMissEvent struct {
	Id        int    `json:"id" gorm:"primaryKey;autoIncrement"`
	ModelName string `json:"model_name" gorm:"type:varchar(255);index;not null"`
	UserId    int    `json:"user_id" gorm:"not null"`
	Group     string `json:"group" gorm:"type:varchar(64);index"`
	UserAgent string `json:"user_agent" gorm:"type:varchar(512)"`
	CreatedAt int64  `json:"created_at" gorm:"bigint;index"`
}

func (ModelMissEvent) TableName() string {
	return "model_miss_events"
}

// RecordModelMiss 异步记录未命中模型事件（fire-and-forget，不阻塞请求）
func RecordModelMiss(modelName string, userId int, group string, userAgent string) {
	gopool.Go(func() {
		event := &ModelMissEvent{
			ModelName: modelName,
			UserId:    userId,
			Group:     group,
			UserAgent: userAgent,
			CreatedAt: common.GetTimestamp(),
		}
		if err := LOG_DB.Create(event).Error; err != nil {
			common.SysError("failed to record model miss event: " + err.Error())
		}
	})
}

// MissedModelStat 未命中模型统计
type MissedModelStat struct {
	ModelName   string `json:"model_name"`
	MissCount   int64  `json:"miss_count"`
	UserCount   int64  `json:"user_count"`
	LastMissAt  int64  `json:"last_miss_at"`
}

// GetMissedModelStats 获取未命中模型统计，按未命中次数降序
func GetMissedModelStats(startTimestamp int64, endTimestamp int64) ([]MissedModelStat, error) {
	var stats []MissedModelStat
	query := LOG_DB.Table("model_miss_events").
		Select("model_name, count(*) as miss_count, count(distinct user_id) as user_count, max(created_at) as last_miss_at").
		Where("created_at >= ? AND created_at <= ?", startTimestamp, endTimestamp).
		Group("model_name").
		Order("miss_count desc")

	if err := query.Scan(&stats).Error; err != nil {
		return nil, err
	}
	return stats, nil
}

// CleanupOldMissEvents 清理超过指定天数的旧事件
func CleanupOldMissEvents(retentionDays int) error {
	threshold := common.GetTimestamp() - int64(retentionDays*86400)
	return LOG_DB.Where("created_at < ?", threshold).Delete(&ModelMissEvent{}).Error
}
