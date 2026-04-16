package controller

import (
	"net/http"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

type FeedbackRequest struct {
	ModelName string `json:"model_name" binding:"required"`
	RequestId string `json:"request_id"`
	Rating    int    `json:"rating" binding:"required"` // 1 or -1
	Comment   string `json:"comment"`
}

// PostFeedback 用户提交模型反馈
func PostFeedback(c *gin.Context) {
	var req FeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid request"})
		return
	}
	if req.Rating != 1 && req.Rating != -1 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "rating must be 1 or -1"})
		return
	}

	userId := c.GetInt("id")
	model.RecordFeedback(userId, req.ModelName, req.RequestId, req.Rating, req.Comment)
	common.ApiSuccess(c, nil)
}

// GetFeedbackStats 管理员查看模型反馈统计
func GetFeedbackStats(c *gin.Context) {
	start, end := parseAnalyticsPeriod(c)
	stats, err := model.GetModelFeedbackStats(start, end)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if stats == nil {
		stats = []model.ModelFeedbackStat{}
	}
	common.ApiSuccess(c, stats)
}

// GetRecentFeedbackList 管理员查看最近反馈
func GetRecentFeedbackList(c *gin.Context) {
	modelName := c.Query("model")
	feedbacks, err := model.GetRecentFeedbacks(modelName, 50)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if feedbacks == nil {
		feedbacks = []model.ModelFeedback{}
	}
	common.ApiSuccess(c, feedbacks)
}
