package channelctl

import (
	"net/http"
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

// GetChannelHealthSummary 返回最近 N 小时各渠道成功/失败/延迟聚合。
//
// Query:
//   - hours: 观察窗口（小时），默认 24，上限 168（7 天）避免全表扫描
//
// Response:
//
//	{
//	  success: true,
//	  data: {
//	    hours: 24,
//	    rows: [{ channel_id, channel_name, total, successes, errors,
//	             error_rate, avg_use_time_sec, max_use_time_sec, status,
//	             channel_type }]
//	  }
//	}
func GetChannelHealthSummary(c *gin.Context) {
	hours, _ := strconv.Atoi(c.Query("hours"))
	if hours <= 0 {
		hours = 24
	}
	if hours > 168 {
		hours = 168
	}

	rows, err := model.QueryChannelHealthSummary(hours)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"hours": hours,
			"rows":  rows,
		},
	})
}
