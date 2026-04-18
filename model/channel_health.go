package model

import (
	"time"
)

// ChannelHealthRow 单个渠道在观察窗内的聚合指标。
type ChannelHealthRow struct {
	ChannelId   int     `json:"channel_id"`
	ChannelName string  `json:"channel_name"`
	ChannelType int     `json:"channel_type"`
	Status      int     `json:"status"`
	Total       int64   `json:"total"`
	Successes   int64   `json:"successes"`
	Errors      int64   `json:"errors"`
	ErrorRate   float64 `json:"error_rate"`
	AvgUseTime  float64 `json:"avg_use_time_sec"` // 秒
	MaxUseTime  int64   `json:"max_use_time_sec"`
}

// QueryChannelHealthSummary 聚合最近 `hours` 小时各渠道成功/失败/延迟。
// 统计口径：type=2（consume）视为成功，type=5（error）视为失败。
// 返回按 total 降序排列的结果。
func QueryChannelHealthSummary(hours int) ([]ChannelHealthRow, error) {
	if hours <= 0 {
		hours = 24
	}
	since := time.Now().Add(-time.Duration(hours) * time.Hour).Unix()

	type agg struct {
		ChannelId  int
		Total      int64
		Successes  int64
		Errors     int64
		AvgUseTime float64
		MaxUseTime int64
	}
	var aggs []agg
	// NOTE: 跨 SQLite / MySQL / PostgreSQL 通用聚合，避免 PERCENTILE_CONT 等专属函数。
	// use_time 口径与 Log 表一致（秒）。
	err := LOG_DB.Table("logs").
		Select(`channel_id AS channel_id,
			COUNT(*) AS total,
			SUM(CASE WHEN type = ? THEN 1 ELSE 0 END) AS successes,
			SUM(CASE WHEN type = ? THEN 1 ELSE 0 END) AS errors,
			AVG(use_time) AS avg_use_time,
			MAX(use_time) AS max_use_time`,
			LogTypeConsume, LogTypeError).
		Where("created_at >= ?", since).
		Group("channel_id").
		Scan(&aggs).Error
	if err != nil {
		return nil, err
	}

	// 批量补齐渠道名 + 当前状态（从 DB 侧查一次，避免 N+1）
	channelIds := make([]int, 0, len(aggs))
	for _, a := range aggs {
		channelIds = append(channelIds, a.ChannelId)
	}
	chMap := make(map[int]*Channel, len(channelIds))
	if len(channelIds) > 0 {
		var chs []Channel
		if err := DB.Where("id IN ?", channelIds).Find(&chs).Error; err == nil {
			for i := range chs {
				chMap[chs[i].Id] = &chs[i]
			}
		}
	}

	rows := make([]ChannelHealthRow, 0, len(aggs))
	for _, a := range aggs {
		row := ChannelHealthRow{
			ChannelId:  a.ChannelId,
			Total:      a.Total,
			Successes:  a.Successes,
			Errors:     a.Errors,
			AvgUseTime: a.AvgUseTime,
			MaxUseTime: a.MaxUseTime,
		}
		if a.Total > 0 {
			row.ErrorRate = float64(a.Errors) / float64(a.Total)
		}
		if ch, ok := chMap[a.ChannelId]; ok {
			row.ChannelName = ch.Name
			row.ChannelType = ch.Type
			row.Status = ch.Status
		}
		rows = append(rows, row)
	}

	// 按 total 降序
	for i := 0; i < len(rows); i++ {
		for j := i + 1; j < len(rows); j++ {
			if rows[j].Total > rows[i].Total {
				rows[i], rows[j] = rows[j], rows[i]
			}
		}
	}
	return rows, nil
}
