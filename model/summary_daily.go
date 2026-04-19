package model

import (
	"errors"
	"fmt"
	"time"

	"github.com/QuantumNous/new-api/common"
)

// SummaryDaily 按日预聚合，key = (date, user_id, group, model_name, channel_id)。
// 由 rollup 任务每日凌晨扫昨天的 logs 写入；Dashboard 查历史区间时走这里，
// 避免扫 logs 明细。索引：
//   - PRIMARY (date, user_id, model_name, channel_id, group) — 保证 UPSERT 唯一
//   - idx_date                    — 时间窗过滤
//   - idx_sd_user_date            — 按用户
//   - idx_sd_model_date           — 按模型
type SummaryDaily struct {
	Date       string `json:"date" gorm:"type:varchar(10);primaryKey;index:idx_date;index:idx_sd_user_date,priority:2;index:idx_sd_model_date,priority:2"`
	UserId     int    `json:"user_id" gorm:"primaryKey;autoIncrement:false;index:idx_sd_user_date,priority:1"`
	Username   string `json:"username" gorm:"type:varchar(64);default:''"`
	Group      string `json:"group" gorm:"column:group;type:varchar(64);primaryKey;default:''"`
	ModelName  string `json:"model_name" gorm:"type:varchar(128);primaryKey;default:'';index:idx_sd_model_date,priority:1"`
	ChannelId  int    `json:"channel_id" gorm:"primaryKey;autoIncrement:false;default:0"`

	RequestCount     int   `json:"request_count" gorm:"default:0"` // type=2
	FailCount        int   `json:"fail_count" gorm:"default:0"`    // type=5
	PromptTokens     int64 `json:"prompt_tokens" gorm:"default:0"`
	CompletionTokens int64 `json:"completion_tokens" gorm:"default:0"`
	Quota            int64 `json:"quota" gorm:"default:0"`
	UseTimeSum       int64 `json:"use_time_sum" gorm:"default:0"` // 秒，用于 avg latency
}

// TableName 固定表名，避免 gorm 复数化。
func (SummaryDaily) TableName() string {
	return "summary_daily"
}

// --------------------------------------------------------------------------
// Rollup：把 logs 按 (date, user, group, model, channel) 聚合写入 summary_daily
// --------------------------------------------------------------------------

// RollupLogsToSummary 聚合 [startUnix, endUnix) 区间内的 logs 并 upsert 到 summary_daily。
// 常见用法：
//   - 每日 00:10 调度：rollup 昨天（startUnix=昨天00:00, endUnix=今天00:00）
//   - 手动回填：rollup 从历史最早到今天 00:00 的所有数据
//
// 返回写入/更新的行数（rollup 产生的 summary_daily 条数）。
func RollupLogsToSummary(startUnix, endUnix int64) (int64, error) {
	if endUnix <= startUnix {
		return 0, errors.New("invalid time range")
	}

	// 跨库时（LOG_DB != DB）无法直接 INSERT INTO ... SELECT FROM logs。
	// 先按 date 分批聚合到内存，再 upsert 到 DB。
	// 粒度：以 UTC 天对齐即可；summary_daily.date 存 'YYYY-MM-DD'。
	type aggRow struct {
		Date             string
		UserId           int
		Username         string
		Group            string
		ModelName        string
		ChannelId        int
		RequestCount     int
		FailCount        int
		PromptTokens     int64
		CompletionTokens int64
		Quota            int64
		UseTimeSum       int64
	}

	var dateExpr string
	switch {
	case common.UsingSQLite:
		dateExpr = "strftime('%Y-%m-%d', datetime(created_at, 'unixepoch'))"
	case common.UsingPostgreSQL:
		dateExpr = "to_char(to_timestamp(created_at), 'YYYY-MM-DD')"
	default:
		dateExpr = "FROM_UNIXTIME(created_at, '%Y-%m-%d')"
	}

	sel := fmt.Sprintf(`
		%s AS date,
		user_id, COALESCE(NULLIF(username, ''), '') AS username,
		COALESCE(NULLIF(%s, ''), '') AS `+"`group`"+`,
		COALESCE(NULLIF(model_name, ''), '') AS model_name,
		channel_id,
		SUM(CASE WHEN type = ? THEN 1 ELSE 0 END) AS request_count,
		SUM(CASE WHEN type = ? THEN 1 ELSE 0 END) AS fail_count,
		SUM(CASE WHEN type = ? THEN prompt_tokens ELSE 0 END) AS prompt_tokens,
		SUM(CASE WHEN type = ? THEN completion_tokens ELSE 0 END) AS completion_tokens,
		SUM(CASE WHEN type = ? THEN quota ELSE 0 END) AS quota,
		SUM(CASE WHEN type = ? THEN use_time ELSE 0 END) AS use_time_sum
	`, dateExpr, commonGroupCol)

	var rows []aggRow
	err := LOG_DB.Table("logs").
		Select(sel,
			LogTypeConsume, LogTypeError,
			LogTypeConsume, LogTypeConsume, LogTypeConsume, LogTypeConsume).
		Where("type IN ? AND created_at >= ? AND created_at < ?",
			[]int{LogTypeConsume, LogTypeError}, startUnix, endUnix).
		Group("date, user_id, username, " + commonGroupCol + ", model_name, channel_id").
		Scan(&rows).Error
	if err != nil {
		return 0, fmt.Errorf("aggregate logs: %w", err)
	}
	if len(rows) == 0 {
		return 0, nil
	}

	// UPSERT（PG / MySQL / SQLite 通用 ON CONFLICT，使用 gorm v2 Clauses）
	// 为避免 gorm OnConflict 在不同方言下的差异，用逐条 FirstOrCreate + Select/Updates 较慢；
	// 这里直接写原生 INSERT ... ON CONFLICT / ON DUPLICATE KEY UPDATE。
	tx := DB.Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}
	inserted := int64(0)
	for _, r := range rows {
		entry := SummaryDaily{
			Date:             r.Date,
			UserId:           r.UserId,
			Username:         r.Username,
			Group:            r.Group,
			ModelName:        r.ModelName,
			ChannelId:        r.ChannelId,
			RequestCount:     r.RequestCount,
			FailCount:        r.FailCount,
			PromptTokens:     r.PromptTokens,
			CompletionTokens: r.CompletionTokens,
			Quota:            r.Quota,
			UseTimeSum:       r.UseTimeSum,
		}
		// Upsert：存在则覆盖最新聚合（支持补跑幂等）
		res := tx.Where(&SummaryDaily{
			Date: entry.Date, UserId: entry.UserId,
			Group: entry.Group, ModelName: entry.ModelName, ChannelId: entry.ChannelId,
		}).Assign(map[string]any{
			"username":          entry.Username,
			"request_count":     entry.RequestCount,
			"fail_count":        entry.FailCount,
			"prompt_tokens":     entry.PromptTokens,
			"completion_tokens": entry.CompletionTokens,
			"quota":             entry.Quota,
			"use_time_sum":      entry.UseTimeSum,
		}).FirstOrCreate(&SummaryDaily{})
		if res.Error != nil {
			tx.Rollback()
			return inserted, res.Error
		}
		inserted += res.RowsAffected
	}
	if err := tx.Commit().Error; err != nil {
		return inserted, err
	}
	return inserted, nil
}

// LatestSummaryDate 返回 summary_daily 里最新一天的日期字符串（YYYY-MM-DD）。
// 空时返回 ""，由调度器据此决定从哪里开始回填。
func LatestSummaryDate() (string, error) {
	var row SummaryDaily
	err := DB.Order("date desc").Limit(1).Take(&row).Error
	if err != nil {
		return "", nil
	}
	return row.Date, nil
}

// EarliestLogDate 返回 logs 表里最早那条 type=2/5 日志的日期（UTC 天），
// 用于首次回填起点。空表返回零值。
func EarliestLogDate() (time.Time, error) {
	var ts int64
	err := LOG_DB.Table("logs").
		Select("MIN(created_at)").
		Where("type IN ?", []int{LogTypeConsume, LogTypeError}).
		Scan(&ts).Error
	if err != nil || ts == 0 {
		return time.Time{}, err
	}
	return time.Unix(ts, 0), nil
}
