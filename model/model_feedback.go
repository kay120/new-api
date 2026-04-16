package model

import (
	"github.com/QuantumNous/new-api/common"
	"github.com/bytedance/gopkg/util/gopool"
)

type ModelFeedback struct {
	Id        int    `json:"id" gorm:"primaryKey;autoIncrement"`
	UserId    int    `json:"user_id" gorm:"index;not null"`
	ModelName string `json:"model_name" gorm:"type:varchar(255);index;not null"`
	RequestId string `json:"request_id" gorm:"type:varchar(64);index"`
	Rating    int    `json:"rating" gorm:"not null"` // 1 = 赞, -1 = 踩
	Comment   string `json:"comment" gorm:"type:text"`
	CreatedAt int64  `json:"created_at" gorm:"bigint;index"`
}

func (ModelFeedback) TableName() string {
	return "model_feedbacks"
}

// RecordFeedback 异步记录用户反馈
func RecordFeedback(userId int, modelName string, requestId string, rating int, comment string) {
	gopool.Go(func() {
		fb := &ModelFeedback{
			UserId:    userId,
			ModelName: modelName,
			RequestId: requestId,
			Rating:    rating,
			Comment:   comment,
			CreatedAt: common.GetTimestamp(),
		}
		if err := LOG_DB.Create(fb).Error; err != nil {
			common.SysError("failed to record feedback: " + err.Error())
		}
	})
}

// ModelFeedbackStat 模型反馈统计
type ModelFeedbackStat struct {
	ModelName  string  `json:"model_name"`
	TotalCount int64   `json:"total_count"`
	LikeCount  int64   `json:"like_count"`
	DislikeCount int64 `json:"dislike_count"`
	LikeRate   float64 `json:"like_rate"`
}

// GetModelFeedbackStats 获取模型反馈统计
func GetModelFeedbackStats(startTs int64, endTs int64) ([]ModelFeedbackStat, error) {
	var stats []ModelFeedbackStat

	likeExpr := "CASE WHEN rating = 1 THEN 1 ELSE 0 END"
	dislikeExpr := "CASE WHEN rating = -1 THEN 1 ELSE 0 END"

	err := LOG_DB.Table("model_feedbacks").
		Select("model_name, count(*) as total_count, "+
			"sum("+likeExpr+") as like_count, "+
			"sum("+dislikeExpr+") as dislike_count, "+
			"CAST(sum("+likeExpr+") AS FLOAT) / CAST(count(*) AS FLOAT) * 100 as like_rate").
		Where("created_at >= ? AND created_at <= ?", startTs, endTs).
		Group("model_name").
		Order("total_count desc").
		Scan(&stats).Error
	return stats, err
}

// GetRecentFeedbacks 获取某模型最近的反馈
func GetRecentFeedbacks(modelName string, limit int) ([]ModelFeedback, error) {
	var feedbacks []ModelFeedback
	query := LOG_DB.Order("created_at desc").Limit(limit)
	if modelName != "" {
		query = query.Where("model_name = ?", modelName)
	}
	err := query.Find(&feedbacks).Error
	return feedbacks, err
}
