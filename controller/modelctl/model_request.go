package modelctl

import (
	"net/http"
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

type CreateModelRequestBody struct {
	ModelName string `json:"model_name" binding:"required"`
	Reason    string `json:"reason"`
}

// CreateModelRequestHandler 用户提交模型申请
func CreateModelRequestHandler(c *gin.Context) {
	var req CreateModelRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "model_name is required"})
		return
	}

	userId := c.GetInt("id")
	username := c.GetString("username")
	result, err := model.CreateModelRequest(userId, username, req.ModelName, req.Reason)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, result)
}

// GetUserModelRequestsHandler 用户查看自己的申请
func GetUserModelRequestsHandler(c *gin.Context) {
	userId := c.GetInt("id")
	requests, err := model.GetUserModelRequests(userId)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if requests == nil {
		requests = []model.ModelRequest{}
	}
	common.ApiSuccess(c, requests)
}

// GetAllModelRequestsHandler 管理员查看所有申请
func GetAllModelRequestsHandler(c *gin.Context) {
	statusStr := c.DefaultQuery("status", "-1")
	status, _ := strconv.Atoi(statusStr)
	page, _ := strconv.Atoi(c.DefaultQuery("p", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	requests, total, err := model.GetModelRequests(status, page, pageSize)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if requests == nil {
		requests = []model.ModelRequest{}
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    requests,
		"total":   total,
	})
}

type UpdateModelRequestBody struct {
	Status     int    `json:"status" binding:"required"`
	AdminReply string `json:"admin_reply"`
}

// UpdateModelRequestHandler 管理员审批申请
func UpdateModelRequestHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid id"})
		return
	}

	var req UpdateModelRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid request"})
		return
	}

	if err := model.UpdateModelRequestStatus(id, req.Status, req.AdminReply); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, nil)
}

// GetModelRequestStatsHandler 模型申请统计
func GetModelRequestStatsHandler(c *gin.Context) {
	stats, err := model.GetModelRequestStats()
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if stats == nil {
		stats = []model.ModelRequestStat{}
	}
	common.ApiSuccess(c, stats)
}
