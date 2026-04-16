package controller

import (
	"net/http"
	"strconv"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

func parseAnalyticsPeriod(c *gin.Context) (int64, int64) {
	period := c.DefaultQuery("period", "7d")
	now := time.Now().Unix()

	var duration int64
	switch period {
	case "1d":
		duration = 86400
	case "7d":
		duration = 7 * 86400
	case "30d":
		duration = 30 * 86400
	case "90d":
		duration = 90 * 86400
	default:
		duration = 7 * 86400
	}

	// 支持自定义时间戳
	if start := c.Query("start_timestamp"); start != "" {
		if s, err := strconv.ParseInt(start, 10, 64); err == nil {
			end := now
			if e := c.Query("end_timestamp"); e != "" {
				if ev, err := strconv.ParseInt(e, 10, 64); err == nil {
					end = ev
				}
			}
			return s, end
		}
	}

	return now - duration, now
}

// GetModelUsage 模型使用量统计
func GetModelUsage(c *gin.Context) {
	start, end := parseAnalyticsPeriod(c)
	groupFilter := c.Query("group")

	stats, err := model.GetModelsByReport(start, end, groupFilter)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, stats)
}

// GetModelRanking 模型排行榜（按请求量/Token量排序）
func GetModelRanking(c *gin.Context) {
	start, end := parseAnalyticsPeriod(c)
	groupFilter := c.Query("group")

	stats, err := model.GetModelsByReport(start, end, groupFilter)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, stats)
}

// GetMissedModels 未命中模型统计
func GetMissedModels(c *gin.Context) {
	start, end := parseAnalyticsPeriod(c)

	stats, err := model.GetMissedModelStats(start, end)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if stats == nil {
		stats = []model.MissedModelStat{}
	}
	common.ApiSuccess(c, stats)
}

// GetChannelHealthOverview 所有渠道健康概览（服务状态页）
func GetChannelHealthOverview(c *gin.Context) {
	start, _ := parseAnalyticsPeriod(c)

	stats, err := model.GetChannelHealthStats(start)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if stats == nil {
		stats = []model.ChannelHealthStat{}
	}
	common.ApiSuccess(c, stats)
}

// GetChannelHealthDetail 单渠道健康检查历史
func GetChannelHealthDetail(c *gin.Context) {
	idStr := c.Param("id")
	channelId, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid channel id"})
		return
	}

	limitStr := c.DefaultQuery("limit", "60")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 || limit > 200 {
		limit = 60
	}

	checks, err := model.GetRecentHealthChecks(channelId, limit)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if checks == nil {
		checks = []model.ChannelHealthCheck{}
	}
	common.ApiSuccess(c, checks)
}
