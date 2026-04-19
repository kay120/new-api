package model

import (
	"errors"
	"time"

	"github.com/QuantumNous/new-api/common"
)

// GroupStat 按组统计
type GroupStat struct {
	Group            string `json:"group"`
	Quota            int    `json:"quota"`
	RequestCount     int    `json:"request_count"`
	PromptTokens     int    `json:"prompt_tokens"`
	CompletionTokens int    `json:"completion_tokens"`
	UserCount        int    `json:"user_count"`
}

// ModelStat 按模型统计
type ModelStat struct {
	ModelName        string `json:"model_name"`
	Quota            int    `json:"quota"`
	RequestCount     int    `json:"request_count"`
	PromptTokens     int    `json:"prompt_tokens"`
	CompletionTokens int    `json:"completion_tokens"`
	UserCount        int    `json:"user_count"`
}

// UserStat 按用户统计
type UserStat struct {
	Username         string `json:"username"`
	Group            string `json:"group"`
	Quota            int    `json:"quota"`
	RequestCount     int    `json:"request_count"`
	PromptTokens     int    `json:"prompt_tokens"`
	CompletionTokens int    `json:"completion_tokens"`
	ModelCount       int    `json:"model_count"`
}

// ModelUserBreakdown 单条 (model, user) 聚合，供 Dashboard 明细面板使用。
// 不依赖 /api/log 分页（会被 pageSize cap=100 截断）。
type ModelUserBreakdown struct {
	ModelName        string `json:"model_name"`
	Username         string `json:"username"`
	UserId           int    `json:"user_id"`
	RequestCount     int    `json:"request_count"`
	PromptTokens     int    `json:"prompt_tokens"`
	CompletionTokens int    `json:"completion_tokens"`
	FailCount        int    `json:"fail_count"`
}

// GetModelUserBreakdown 返回时间窗内每个 (model_name, user_id) 组合的完整聚合，
// 含成功 tokens / 失败次数。前端用来精确渲染"某模型的用户占比 / 某用户的模型占比"，
// 避免依赖 /api/log 列表 + pageSize<=100 的片面样本。
func GetModelUserBreakdown(startTimestamp int64, endTimestamp int64, groupFilter string) ([]ModelUserBreakdown, error) {
	if startTimestamp == 0 || endTimestamp == 0 {
		return nil, errors.New("start_timestamp and end_timestamp are required")
	}

	query := LOG_DB.Table("logs").
		Select(`logs.model_name, logs.username, logs.user_id,
			sum(case when logs.type = ? then 1 else 0 end) as request_count,
			sum(case when logs.type = ? then logs.prompt_tokens else 0 end) as prompt_tokens,
			sum(case when logs.type = ? then logs.completion_tokens else 0 end) as completion_tokens,
			sum(case when logs.type = ? then 1 else 0 end) as fail_count`,
			LogTypeConsume, LogTypeConsume, LogTypeConsume, LogTypeError).
		Where("logs.type IN ? and logs.created_at >= ? and logs.created_at <= ?",
			[]int{LogTypeConsume, LogTypeError}, startTimestamp, endTimestamp)

	if groupFilter != "" {
		query = query.Where("logs.group = ?", groupFilter)
	}

	query = query.Group("logs.model_name, logs.user_id, logs.username")

	var rows []ModelUserBreakdown
	if err := query.Scan(&rows).Error; err != nil {
		common.SysError("failed to query model-user breakdown: " + err.Error())
		return nil, errors.New("查询统计数据失败")
	}
	return rows, nil
}

// DailyStat 按日统计
type DailyStat struct {
	Date             string `json:"date"`
	Quota            int    `json:"quota"`
	RequestCount     int    `json:"request_count"`
	PromptTokens     int    `json:"prompt_tokens"`
	CompletionTokens int    `json:"completion_tokens"`
}

// ReportOverview 报表概览
type ReportOverview struct {
	TotalQuota            int         `json:"total_quota"`
	TotalRequests         int         `json:"total_requests"`
	TotalUsers            int         `json:"total_users"`
	TotalModels           int         `json:"total_models"`
	TotalPromptTokens     int         `json:"total_prompt_tokens"`
	TotalCompletionTokens int         `json:"total_completion_tokens"`
	DailyStats            []DailyStat `json:"daily_stats"`
	TopGroups             []GroupStat `json:"top_groups"`
	TopModels             []ModelStat `json:"top_models"`
	TopUsers              []UserStat  `json:"top_users"`
}

func GetGroupsByReport(startTimestamp int64, endTimestamp int64, groupFilter string) ([]GroupStat, error) {
	if startTimestamp == 0 || endTimestamp == 0 {
		return nil, errors.New("start_timestamp and end_timestamp are required")
	}

	query := LOG_DB.Table("logs").
		Select("logs.group, sum(logs.quota) as quota, count(*) as request_count, sum(logs.prompt_tokens) as prompt_tokens, sum(logs.completion_tokens) as completion_tokens, count(distinct logs.user_id) as user_count").
		Where("logs.type = ? and logs.created_at >= ? and logs.created_at <= ?", LogTypeConsume, startTimestamp, endTimestamp)

	if groupFilter != "" {
		query = query.Where("logs.group = ?", groupFilter)
	}

	query = query.Group("logs.group").Order("quota desc")

	var stats []GroupStat
	if err := query.Scan(&stats).Error; err != nil {
		common.SysError("failed to query group stats: " + err.Error())
		return nil, errors.New("查询统计数据失败")
	}

	return stats, nil
}

func GetModelsByReport(startTimestamp int64, endTimestamp int64, groupFilter string) ([]ModelStat, error) {
	if startTimestamp == 0 || endTimestamp == 0 {
		return nil, errors.New("start_timestamp and end_timestamp are required")
	}

	query := LOG_DB.Table("logs").
		Select("logs.model_name, sum(logs.quota) as quota, count(*) as request_count, sum(logs.prompt_tokens) as prompt_tokens, sum(logs.completion_tokens) as completion_tokens, count(distinct logs.user_id) as user_count").
		Where("logs.type = ? and logs.created_at >= ? and logs.created_at <= ?", LogTypeConsume, startTimestamp, endTimestamp)

	if groupFilter != "" {
		query = query.Where("logs.group = ?", groupFilter)
	}

	query = query.Group("logs.model_name").Order("quota desc")

	var stats []ModelStat
	if err := query.Scan(&stats).Error; err != nil {
		common.SysError("failed to query model stats: " + err.Error())
		return nil, errors.New("查询统计数据失败")
	}

	return stats, nil
}

func GetUsersByReport(startTimestamp int64, endTimestamp int64, groupFilter string, limit int) ([]UserStat, error) {
	if startTimestamp == 0 || endTimestamp == 0 {
		return nil, errors.New("start_timestamp and end_timestamp are required")
	}

	if limit <= 0 {
		limit = 50
	}

	query := LOG_DB.Table("logs").
		Select("logs.username, logs.group, sum(logs.quota) as quota, count(*) as request_count, sum(logs.prompt_tokens) as prompt_tokens, sum(logs.completion_tokens) as completion_tokens, count(distinct logs.model_name) as model_count").
		Where("logs.type = ? and logs.created_at >= ? and logs.created_at <= ?", LogTypeConsume, startTimestamp, endTimestamp)

	if groupFilter != "" {
		query = query.Where("logs.group = ?", groupFilter)
	}

	query = query.Group("logs.username, logs.group").Order("quota desc").Limit(limit)

	var stats []UserStat
	if err := query.Scan(&stats).Error; err != nil {
		common.SysError("failed to query user stats: " + err.Error())
		return nil, errors.New("查询统计数据失败")
	}

	return stats, nil
}

func GetDailyStatsByReport(startTimestamp int64, endTimestamp int64, groupFilter string) ([]DailyStat, error) {
	if startTimestamp == 0 || endTimestamp == 0 {
		return nil, errors.New("start_timestamp and end_timestamp are required")
	}

	type rawDailyStat struct {
		DateUnix         int64 `gorm:"column:date"`
		Quota            int   `gorm:"column:quota"`
		RequestCount     int   `gorm:"column:request_count"`
		PromptTokens     int   `gorm:"column:prompt_tokens"`
		CompletionTokens int   `gorm:"column:completion_tokens"`
	}

	// 使用 MySQL 的 DATE(FROM_UNIXTIME()) 按本地时区分组，避免 UTC 偏移问题
	query := LOG_DB.Table("logs").
		Select("UNIX_TIMESTAMP(DATE(FROM_UNIXTIME(logs.created_at))) as `date`, sum(logs.quota) as quota, count(*) as request_count, sum(logs.prompt_tokens) as prompt_tokens, sum(logs.completion_tokens) as completion_tokens").
		Where("logs.type = ? and logs.created_at >= ? and logs.created_at <= ?", LogTypeConsume, startTimestamp, endTimestamp)

	if groupFilter != "" {
		query = query.Where("logs.group = ?", groupFilter)
	}

	query = query.Group("DATE(FROM_UNIXTIME(logs.created_at))").Order("`date` asc")

	var rawStats []rawDailyStat
	if err := query.Scan(&rawStats).Error; err != nil {
		common.SysError("failed to query daily stats: " + err.Error())
		return nil, errors.New("查询统计数据失败")
	}

	stats := make([]DailyStat, 0, len(rawStats))
	for _, r := range rawStats {
		stats = append(stats, DailyStat{
			Date:             time.Unix(r.DateUnix, 0).Format("2006-01-02"),
			Quota:            r.Quota,
			RequestCount:     r.RequestCount,
			PromptTokens:     r.PromptTokens,
			CompletionTokens: r.CompletionTokens,
		})
	}

	return stats, nil
}

func GetReportOverview(startTimestamp int64, endTimestamp int64, groupFilter string) (*ReportOverview, error) {
	if startTimestamp == 0 || endTimestamp == 0 {
		now := time.Now()
		startTimestamp = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Unix()
		endTimestamp = now.Unix()
	}

	var overview ReportOverview

	// 总计
	type totalResult struct {
		TotalQuota            int
		TotalRequests         int
		TotalUsers            int
		TotalModels           int
		TotalPromptTokens     int
		TotalCompletionTokens int
	}

	query := LOG_DB.Table("logs").
		Select("sum(logs.quota) as total_quota, count(*) as total_requests, count(distinct logs.user_id) as total_users, count(distinct logs.model_name) as total_models, sum(logs.prompt_tokens) as total_prompt_tokens, sum(logs.completion_tokens) as total_completion_tokens").
		Where("logs.type = ? and logs.created_at >= ? and logs.created_at <= ?", LogTypeConsume, startTimestamp, endTimestamp)

	if groupFilter != "" {
		query = query.Where("logs.group = ?", groupFilter)
	}

	var total totalResult
	if err := query.Scan(&total).Error; err != nil {
		common.SysError("failed to query overview total: " + err.Error())
		return nil, errors.New("查询统计数据失败")
	}

	overview.TotalQuota = total.TotalQuota
	overview.TotalRequests = total.TotalRequests
	overview.TotalUsers = total.TotalUsers
	overview.TotalModels = total.TotalModels
	overview.TotalPromptTokens = total.TotalPromptTokens
	overview.TotalCompletionTokens = total.TotalCompletionTokens

	// 每日趋势
	dailyStats, _ := GetDailyStatsByReport(startTimestamp, endTimestamp, groupFilter)
	overview.DailyStats = dailyStats

	// Top groups
	topGroups, _ := GetGroupsByReport(startTimestamp, endTimestamp, "")
	if len(topGroups) > 10 {
		topGroups = topGroups[:10]
	}
	overview.TopGroups = topGroups

	// Top models
	topModels, _ := GetModelsByReport(startTimestamp, endTimestamp, groupFilter)
	if len(topModels) > 10 {
		topModels = topModels[:10]
	}
	overview.TopModels = topModels

	// Top users
	topUsers, _ := GetUsersByReport(startTimestamp, endTimestamp, groupFilter, 10)
	overview.TopUsers = topUsers

	return &overview, nil
}
