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

// HourlyStat 单日分时统计（今日从 logs 实时算，历史单日从 summary 聚合），
// Dashboard 用来画"今天 0-23 小时"柱图。
type HourlyStat struct {
	Hour             int   `json:"hour"` // 0-23
	RequestCount     int   `json:"request_count"`
	PromptTokens     int   `json:"prompt_tokens"`
	CompletionTokens int   `json:"completion_tokens"`
	Quota            int64 `json:"quota"`
}

// GetHourlyStatsToday 返回今日 0-23 小时（本地时区）的请求 / token 分布。
// 只看 logs（今日 < 1 天数据量可控），走 idx_created_at_type。
func GetHourlyStatsToday(groupFilter string) ([]HourlyStat, error) {
	now := time.Now()
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	dayEnd := dayStart.Add(24 * time.Hour)

	var hourExpr string
	switch {
	case common.UsingSQLite:
		hourExpr = "CAST(strftime('%H', datetime(created_at, 'unixepoch')) AS INTEGER)"
	case common.UsingPostgreSQL:
		hourExpr = "EXTRACT(HOUR FROM to_timestamp(created_at))"
	default:
		hourExpr = "HOUR(FROM_UNIXTIME(created_at))"
	}

	q := LOG_DB.Table("logs").
		Select(hourExpr+` AS hour,
			COUNT(*) AS request_count,
			SUM(prompt_tokens) AS prompt_tokens,
			SUM(completion_tokens) AS completion_tokens,
			SUM(quota) AS quota`).
		Where("type = ? AND created_at >= ? AND created_at < ?",
			LogTypeConsume, dayStart.Unix(), dayEnd.Unix())
	if groupFilter != "" {
		q = q.Where(commonGroupCol+" = ?", groupFilter)
	}
	q = q.Group(hourExpr).Order("hour asc")

	var rows []HourlyStat
	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}

	// 填 0：所有 24 小时都给一条，空时段为 0，前端画图时均匀显示
	filled := make([]HourlyStat, 24)
	for h := 0; h < 24; h++ {
		filled[h] = HourlyStat{Hour: h}
	}
	for _, r := range rows {
		if r.Hour >= 0 && r.Hour < 24 {
			filled[r.Hour] = r
		}
	}
	return filled, nil
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

// todayStartUnix 返回今天本地 00:00 的 Unix 秒。分层查询的分界点。
func todayStartUnix() int64 {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
}

// splitTimeRange 按今天 00:00 把 [start,end] 切成历史段（走 summary_daily）+
// 今日段（走 logs）。返回值中 histEnd==0 表示无历史段，todayStart==0 表示无今日段。
func splitTimeRange(start, end int64) (histStart, histEnd, todayStart, todayEnd int64) {
	today := todayStartUnix()
	switch {
	case end <= today:
		return start, end, 0, 0
	case start >= today:
		return 0, 0, start, end
	default:
		return start, today, today, end
	}
}

// dateRangeStrs 把 [histStart, histEnd) 转成 YYYY-MM-DD 列表（含 start，不含 end）。
// 用于 summary_daily 的 date IN (...) 过滤。
func dateRangeStrs(startUnix, endUnix int64) []string {
	if endUnix <= startUnix {
		return nil
	}
	start := time.Unix(startUnix, 0)
	end := time.Unix(endUnix, 0)
	// 对齐到本地天的 00:00
	d := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	dates := make([]string, 0, 32)
	for d.Before(end) {
		dates = append(dates, d.Format("2006-01-02"))
		d = d.AddDate(0, 0, 1)
	}
	return dates
}

func GetGroupsByReport(startTimestamp int64, endTimestamp int64, groupFilter string) ([]GroupStat, error) {
	if startTimestamp == 0 || endTimestamp == 0 {
		return nil, errors.New("start_timestamp and end_timestamp are required")
	}

	histStart, histEnd, todayStart, todayEnd := splitTimeRange(startTimestamp, endTimestamp)
	merged := make(map[string]*GroupStat)

	if histEnd > histStart {
		dates := dateRangeStrs(histStart, histEnd)
		if len(dates) > 0 {
			q := DB.Table("summary_daily").
				Select(commonGroupCol+` AS `+commonGroupCol+`,
					SUM(quota) AS quota, SUM(request_count) AS request_count,
					SUM(prompt_tokens) AS prompt_tokens, SUM(completion_tokens) AS completion_tokens,
					COUNT(DISTINCT user_id) AS user_count`).
				Where("date IN ?", dates)
			if groupFilter != "" {
				q = q.Where(commonGroupCol+" = ?", groupFilter)
			}
			var rows []GroupStat
			if err := q.Group(commonGroupCol).Scan(&rows).Error; err != nil {
				common.SysError("group stats (summary) query failed: " + err.Error())
				return nil, errors.New("查询统计数据失败")
			}
			for _, r := range rows {
				r2 := r
				merged[r.Group] = &r2
			}
		}
	}

	if todayEnd > todayStart {
		q := LOG_DB.Table("logs").
			Select(commonGroupCol+` AS `+commonGroupCol+`,
				SUM(quota) AS quota, COUNT(*) AS request_count,
				SUM(prompt_tokens) AS prompt_tokens, SUM(completion_tokens) AS completion_tokens,
				COUNT(DISTINCT user_id) AS user_count`).
			Where("type = ? AND created_at >= ? AND created_at <= ?", LogTypeConsume, todayStart, todayEnd)
		if groupFilter != "" {
			q = q.Where(commonGroupCol+" = ?", groupFilter)
		}
		var rows []GroupStat
		if err := q.Group(commonGroupCol).Scan(&rows).Error; err != nil {
			common.SysError("group stats (logs) query failed: " + err.Error())
			return nil, errors.New("查询统计数据失败")
		}
		for _, r := range rows {
			if ex, ok := merged[r.Group]; ok {
				ex.Quota += r.Quota
				ex.RequestCount += r.RequestCount
				ex.PromptTokens += r.PromptTokens
				ex.CompletionTokens += r.CompletionTokens
				ex.UserCount += r.UserCount
			} else {
				r2 := r
				merged[r.Group] = &r2
			}
		}
	}

	stats := make([]GroupStat, 0, len(merged))
	for _, v := range merged {
		stats = append(stats, *v)
	}
	for i := 0; i < len(stats); i++ {
		for j := i + 1; j < len(stats); j++ {
			if stats[j].Quota > stats[i].Quota {
				stats[i], stats[j] = stats[j], stats[i]
			}
		}
	}
	return stats, nil
}

func GetModelsByReport(startTimestamp int64, endTimestamp int64, groupFilter string) ([]ModelStat, error) {
	if startTimestamp == 0 || endTimestamp == 0 {
		return nil, errors.New("start_timestamp and end_timestamp are required")
	}

	histStart, histEnd, todayStart, todayEnd := splitTimeRange(startTimestamp, endTimestamp)

	// 用 map 聚合两段（按 model_name），应用层合并。user_count 跨段分别算，
	// 跨天的同一 user 会被双计一次——相对总量可忽略，换取实现简洁与性能。
	merged := make(map[string]*ModelStat)

	if histEnd > histStart {
		dates := dateRangeStrs(histStart, histEnd)
		if len(dates) > 0 {
			q := DB.Table("summary_daily").
				Select(`model_name,
					SUM(quota) AS quota,
					SUM(request_count) AS request_count,
					SUM(prompt_tokens) AS prompt_tokens,
					SUM(completion_tokens) AS completion_tokens,
					COUNT(DISTINCT user_id) AS user_count`).
				Where("date IN ?", dates)
			if groupFilter != "" {
				q = q.Where(commonGroupCol+" = ?", groupFilter)
			}
			var rows []ModelStat
			if err := q.Group("model_name").Scan(&rows).Error; err != nil {
				common.SysError("model stats (summary) query failed: " + err.Error())
				return nil, errors.New("查询统计数据失败")
			}
			for _, r := range rows {
				r2 := r
				merged[r.ModelName] = &r2
			}
		}
	}

	if todayEnd > todayStart {
		q := LOG_DB.Table("logs").
			Select(`model_name,
				SUM(quota) AS quota,
				COUNT(*) AS request_count,
				SUM(prompt_tokens) AS prompt_tokens,
				SUM(completion_tokens) AS completion_tokens,
				COUNT(DISTINCT user_id) AS user_count`).
			Where("type = ? AND created_at >= ? AND created_at <= ?", LogTypeConsume, todayStart, todayEnd)
		if groupFilter != "" {
			q = q.Where(commonGroupCol+" = ?", groupFilter)
		}
		var rows []ModelStat
		if err := q.Group("model_name").Scan(&rows).Error; err != nil {
			common.SysError("model stats (logs) query failed: " + err.Error())
			return nil, errors.New("查询统计数据失败")
		}
		for _, r := range rows {
			if ex, ok := merged[r.ModelName]; ok {
				ex.Quota += r.Quota
				ex.RequestCount += r.RequestCount
				ex.PromptTokens += r.PromptTokens
				ex.CompletionTokens += r.CompletionTokens
				ex.UserCount += r.UserCount
			} else {
				r2 := r
				merged[r.ModelName] = &r2
			}
		}
	}

	stats := make([]ModelStat, 0, len(merged))
	for _, v := range merged {
		stats = append(stats, *v)
	}
	// 按 quota desc 排序
	for i := 0; i < len(stats); i++ {
		for j := i + 1; j < len(stats); j++ {
			if stats[j].Quota > stats[i].Quota {
				stats[i], stats[j] = stats[j], stats[i]
			}
		}
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

	histStart, histEnd, todayStart, todayEnd := splitTimeRange(startTimestamp, endTimestamp)

	type key struct {
		Username string
		Group    string
	}
	merged := make(map[key]*UserStat)

	if histEnd > histStart {
		dates := dateRangeStrs(histStart, histEnd)
		if len(dates) > 0 {
			q := DB.Table("summary_daily").
				Select(`username, `+commonGroupCol+` AS `+commonGroupCol+`,
					SUM(quota) AS quota,
					SUM(request_count) AS request_count,
					SUM(prompt_tokens) AS prompt_tokens,
					SUM(completion_tokens) AS completion_tokens,
					COUNT(DISTINCT model_name) AS model_count`).
				Where("date IN ?", dates)
			if groupFilter != "" {
				q = q.Where(commonGroupCol+" = ?", groupFilter)
			}
			var rows []UserStat
			if err := q.Group("username, " + commonGroupCol).Scan(&rows).Error; err != nil {
				common.SysError("user stats (summary) query failed: " + err.Error())
				return nil, errors.New("查询统计数据失败")
			}
			for _, r := range rows {
				r2 := r
				merged[key{r.Username, r.Group}] = &r2
			}
		}
	}

	if todayEnd > todayStart {
		q := LOG_DB.Table("logs").
			Select(`username, `+commonGroupCol+` AS `+commonGroupCol+`,
				SUM(quota) AS quota, COUNT(*) AS request_count,
				SUM(prompt_tokens) AS prompt_tokens, SUM(completion_tokens) AS completion_tokens,
				COUNT(DISTINCT model_name) AS model_count`).
			Where("type = ? AND created_at >= ? AND created_at <= ?", LogTypeConsume, todayStart, todayEnd)
		if groupFilter != "" {
			q = q.Where(commonGroupCol+" = ?", groupFilter)
		}
		var rows []UserStat
		if err := q.Group("username, " + commonGroupCol).Scan(&rows).Error; err != nil {
			common.SysError("user stats (logs) query failed: " + err.Error())
			return nil, errors.New("查询统计数据失败")
		}
		for _, r := range rows {
			k := key{r.Username, r.Group}
			if ex, ok := merged[k]; ok {
				ex.Quota += r.Quota
				ex.RequestCount += r.RequestCount
				ex.PromptTokens += r.PromptTokens
				ex.CompletionTokens += r.CompletionTokens
				ex.ModelCount += r.ModelCount
			} else {
				r2 := r
				merged[k] = &r2
			}
		}
	}

	stats := make([]UserStat, 0, len(merged))
	for _, v := range merged {
		stats = append(stats, *v)
	}
	for i := 0; i < len(stats); i++ {
		for j := i + 1; j < len(stats); j++ {
			if stats[j].Quota > stats[i].Quota {
				stats[i], stats[j] = stats[j], stats[i]
			}
		}
	}
	if len(stats) > limit {
		stats = stats[:limit]
	}
	return stats, nil
}

func GetDailyStatsByReport(startTimestamp int64, endTimestamp int64, groupFilter string) ([]DailyStat, error) {
	if startTimestamp == 0 || endTimestamp == 0 {
		return nil, errors.New("start_timestamp and end_timestamp are required")
	}

	histStart, histEnd, todayStart, todayEnd := splitTimeRange(startTimestamp, endTimestamp)
	merged := make(map[string]*DailyStat)

	// 历史段：按 date 字段 SUM
	if histEnd > histStart {
		dates := dateRangeStrs(histStart, histEnd)
		if len(dates) > 0 {
			q := DB.Table("summary_daily").
				Select(`date,
					SUM(quota) AS quota,
					SUM(request_count) AS request_count,
					SUM(prompt_tokens) AS prompt_tokens,
					SUM(completion_tokens) AS completion_tokens`).
				Where("date IN ?", dates)
			if groupFilter != "" {
				q = q.Where(commonGroupCol+" = ?", groupFilter)
			}
			var rows []DailyStat
			if err := q.Group("date").Scan(&rows).Error; err != nil {
				common.SysError("daily stats (summary) query failed: " + err.Error())
				return nil, errors.New("查询统计数据失败")
			}
			for _, r := range rows {
				r2 := r
				merged[r.Date] = &r2
			}
		}
	}

	// 今日段：从 logs 实时聚合（只有一天，量很小）
	if todayEnd > todayStart {
		var dateExpr string
		switch {
		case common.UsingSQLite:
			dateExpr = "strftime('%Y-%m-%d', datetime(created_at, 'unixepoch'))"
		case common.UsingPostgreSQL:
			dateExpr = "to_char(to_timestamp(created_at), 'YYYY-MM-DD')"
		default:
			dateExpr = "FROM_UNIXTIME(created_at, '%Y-%m-%d')"
		}
		q := LOG_DB.Table("logs").
			Select(dateExpr+` AS date,
				SUM(quota) AS quota, COUNT(*) AS request_count,
				SUM(prompt_tokens) AS prompt_tokens, SUM(completion_tokens) AS completion_tokens`).
			Where("type = ? AND created_at >= ? AND created_at <= ?", LogTypeConsume, todayStart, todayEnd)
		if groupFilter != "" {
			q = q.Where(commonGroupCol+" = ?", groupFilter)
		}
		var rows []DailyStat
		if err := q.Group(dateExpr).Scan(&rows).Error; err != nil {
			common.SysError("daily stats (logs) query failed: " + err.Error())
			return nil, errors.New("查询统计数据失败")
		}
		for _, r := range rows {
			if ex, ok := merged[r.Date]; ok {
				ex.Quota += r.Quota
				ex.RequestCount += r.RequestCount
				ex.PromptTokens += r.PromptTokens
				ex.CompletionTokens += r.CompletionTokens
			} else {
				r2 := r
				merged[r.Date] = &r2
			}
		}
	}

	stats := make([]DailyStat, 0, len(merged))
	for _, v := range merged {
		stats = append(stats, *v)
	}
	// 按 date asc 排序
	for i := 0; i < len(stats); i++ {
		for j := i + 1; j < len(stats); j++ {
			if stats[j].Date < stats[i].Date {
				stats[i], stats[j] = stats[j], stats[i]
			}
		}
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
