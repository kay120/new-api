package controller

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"

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
