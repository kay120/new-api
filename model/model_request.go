package model

import (
	"errors"

	"github.com/QuantumNous/new-api/common"
)

const (
	ModelRequestStatusPending  = 0
	ModelRequestStatusApproved = 1
	ModelRequestStatusRejected = 2
	ModelRequestStatusDeployed = 3
)

type ModelRequest struct {
	Id          int    `json:"id" gorm:"primaryKey;autoIncrement"`
	UserId      int    `json:"user_id" gorm:"index;not null"`
	Username    string `json:"username" gorm:"type:varchar(64)"`
	ModelName   string `json:"model_name" gorm:"type:varchar(255);index;not null"`
	Reason      string `json:"reason" gorm:"type:text"`
	Status      int    `json:"status" gorm:"default:0;index"`
	AdminReply  string `json:"admin_reply" gorm:"type:text"`
	CreatedAt   int64  `json:"created_at" gorm:"bigint;index"`
	UpdatedAt   int64  `json:"updated_at" gorm:"bigint"`
}

func (ModelRequest) TableName() string {
	return "model_requests"
}

// CreateModelRequest 用户提交模型申请
func CreateModelRequest(userId int, username string, modelName string, reason string) (*ModelRequest, error) {
	if modelName == "" {
		return nil, errors.New("model name is required")
	}
	req := &ModelRequest{
		UserId:    userId,
		Username:  username,
		ModelName: modelName,
		Reason:    reason,
		Status:    ModelRequestStatusPending,
		CreatedAt: common.GetTimestamp(),
		UpdatedAt: common.GetTimestamp(),
	}
	if err := DB.Create(req).Error; err != nil {
		return nil, err
	}
	return req, nil
}

// GetModelRequests 获取模型申请列表（管理员）
func GetModelRequests(status int, page int, pageSize int) ([]ModelRequest, int64, error) {
	var requests []ModelRequest
	var total int64

	query := DB.Model(&ModelRequest{})
	if status >= 0 {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)

	err := query.Order("created_at desc").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&requests).Error
	return requests, total, err
}

// GetUserModelRequests 获取用户自己的申请
func GetUserModelRequests(userId int) ([]ModelRequest, error) {
	var requests []ModelRequest
	err := DB.Where("user_id = ?", userId).
		Order("created_at desc").
		Find(&requests).Error
	return requests, err
}

// UpdateModelRequestStatus 管理员审批模型申请
func UpdateModelRequestStatus(id int, status int, adminReply string) error {
	return DB.Model(&ModelRequest{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":      status,
		"admin_reply": adminReply,
		"updated_at":  common.GetTimestamp(),
	}).Error
}

// ModelRequestStat 模型申请统计（按模型名聚合）
type ModelRequestStat struct {
	ModelName    string `json:"model_name"`
	RequestCount int64  `json:"request_count"`
	UserCount    int64  `json:"user_count"`
}

// GetModelRequestStats 获取按模型聚合的申请统计
func GetModelRequestStats() ([]ModelRequestStat, error) {
	var stats []ModelRequestStat
	err := DB.Table("model_requests").
		Select("model_name, count(*) as request_count, count(distinct user_id) as user_count").
		Where("status = ?", ModelRequestStatusPending).
		Group("model_name").
		Order("request_count desc").
		Scan(&stats).Error
	return stats, err
}
