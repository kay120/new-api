package observability

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"

	"github.com/gin-gonic/gin"
)

func isReportAdmin(c *gin.Context) bool {
	role := c.GetInt("role")
	return role >= common.RoleAdminUser
}

func GetReportOverview(c *gin.Context) {
	groupFilter := c.Query("group")
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)

	// 非管理员只能查看自己部门（组）的数据
	if !isReportAdmin(c) && groupFilter == "" {
		groupFilter = c.GetString("group")
	}

	overview, err := model.GetReportOverview(startTimestamp, endTimestamp, groupFilter)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	common.ApiSuccess(c, overview)
}

func GetReportByGroup(c *gin.Context) {
	groupFilter := c.Query("group")
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)

	stats, err := model.GetGroupsByReport(startTimestamp, endTimestamp, groupFilter)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	common.ApiSuccess(c, stats)
}

func GetReportByModel(c *gin.Context) {
	groupFilter := c.Query("group")
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)

	stats, err := model.GetModelsByReport(startTimestamp, endTimestamp, groupFilter)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	common.ApiSuccess(c, stats)
}

func GetReportByUser(c *gin.Context) {
	groupFilter := c.Query("group")
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	stats, err := model.GetUsersByReport(startTimestamp, endTimestamp, groupFilter, limit)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	common.ApiSuccess(c, stats)
}

// GetReportHourlyToday 今天 0-23 小时的请求 / tokens 分布（本地时区，固定 24 条）。
func GetReportHourlyToday(c *gin.Context) {
	groupFilter := c.Query("group")
	if !isReportAdmin(c) && groupFilter == "" {
		groupFilter = c.GetString("group")
	}
	rows, err := model.GetHourlyStatsToday(groupFilter)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, rows)
}

// TriggerReportRollup 供管理员手动回填 summary_daily。
// Query: days=N 表示只回填最近 N 天；不传则回填全部历史。
func TriggerReportRollup(c *gin.Context) {
	if !isReportAdmin(c) {
		common.ApiError(c, fmt.Errorf("需要管理员权限"))
		return
	}
	days, _ := strconv.Atoi(c.DefaultQuery("days", "0"))
	n, err := service.TriggerRollupBackfill(days)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{"rows_written": n})
}

// GetReportModelUserBreakdown 返回时间窗内 (model, user) 完整聚合，
// 前端 Dashboard 下钻面板用来精确渲染"模型×用户"占比，替代依赖
// /api/log 列表（受 pageSize<=100 限制）拼出来的不精确数据。
func GetReportModelUserBreakdown(c *gin.Context) {
	groupFilter := c.Query("group")
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)

	if !isReportAdmin(c) && groupFilter == "" {
		groupFilter = c.GetString("group")
	}

	rows, err := model.GetModelUserBreakdown(startTimestamp, endTimestamp, groupFilter)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, rows)
}

// GetReportByToken 时间窗内按 token (API key) 聚合：请求量 / tokens / 失败数 / 最后使用时间。
// "活跃 Key Top 10" 数据源。
func GetReportByToken(c *gin.Context) {
	groupFilter := c.Query("group")
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)

	if !isReportAdmin(c) && groupFilter == "" {
		groupFilter = c.GetString("group")
	}

	rows, err := model.GetTokenBreakdown(startTimestamp, endTimestamp, groupFilter)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, rows)
}

// GetReportTokenModelBreakdown (token, model) 二维聚合，给"key → 模型"下钻用。
func GetReportTokenModelBreakdown(c *gin.Context) {
	groupFilter := c.Query("group")
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)

	if !isReportAdmin(c) && groupFilter == "" {
		groupFilter = c.GetString("group")
	}

	rows, err := model.GetTokenModelBreakdown(startTimestamp, endTimestamp, groupFilter)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, rows)
}

// GetReportModelLatency 返回时间窗内按模型聚合的请求量 / 错误率 / 延迟分位。
// 走 SQL + 内存分位计算，不受 /api/log 分页 pageSize<=100 的限制。
func GetReportModelLatency(c *gin.Context) {
	groupFilter := c.Query("group")
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)

	if !isReportAdmin(c) && groupFilter == "" {
		groupFilter = c.GetString("group")
	}

	rows, err := model.GetModelLatencyStats(startTimestamp, endTimestamp, groupFilter)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, rows)
}

func ExportReportCSV(c *gin.Context) {
	groupFilter := c.Query("group")
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	exportType := c.DefaultQuery("type", "user") // user, model, group

	if startTimestamp == 0 || endTimestamp == 0 {
		// 默认导出最近 7 天
		endTimestamp = time.Now().Unix()
		startTimestamp = time.Now().AddDate(0, 0, -7).Unix()
	}

	var headers []string
	var rows [][]string

	switch exportType {
	case "user":
		stats, err := model.GetUsersByReport(startTimestamp, endTimestamp, groupFilter, 1000)
		if err != nil {
			common.ApiError(c, err)
			return
		}
		headers = []string{"用户名", "组", "消费额度", "请求次数", "输入Tokens", "输出Tokens"}
		for _, s := range stats {
			rows = append(rows, []string{
				s.Username, s.Group, strconv.Itoa(s.Quota),
				strconv.Itoa(s.RequestCount), strconv.Itoa(s.PromptTokens),
				strconv.Itoa(s.CompletionTokens),
			})
		}
	case "model":
		stats, err := model.GetModelsByReport(startTimestamp, endTimestamp, groupFilter)
		if err != nil {
			common.ApiError(c, err)
			return
		}
		headers = []string{"模型", "消费额度", "请求次数", "输入Tokens", "输出Tokens", "使用人数"}
		for _, s := range stats {
			rows = append(rows, []string{
				s.ModelName, strconv.Itoa(s.Quota), strconv.Itoa(s.RequestCount),
				strconv.Itoa(s.PromptTokens), strconv.Itoa(s.CompletionTokens),
				strconv.Itoa(s.UserCount),
			})
		}
	case "group":
		stats, err := model.GetGroupsByReport(startTimestamp, endTimestamp, groupFilter)
		if err != nil {
			common.ApiError(c, err)
			return
		}
		headers = []string{"组", "消费额度", "请求次数", "输入Tokens", "输出Tokens", "使用人数"}
		for _, s := range stats {
			rows = append(rows, []string{
				s.Group, strconv.Itoa(s.Quota), strconv.Itoa(s.RequestCount),
				strconv.Itoa(s.PromptTokens), strconv.Itoa(s.CompletionTokens),
				strconv.Itoa(s.UserCount),
			})
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "invalid export type",
		})
		return
	}

	// 写入 CSV
	w := csv.NewWriter(c.Writer)
	w.Write(headers)
	for _, row := range rows {
		w.Write(row)
	}
	w.Flush()

	filename := fmt.Sprintf("report_%s_%d_%d.csv", exportType, startTimestamp, endTimestamp)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "text/csv; charset=utf-8")
}
