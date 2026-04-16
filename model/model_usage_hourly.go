package model

import (
	"math"
	"sort"
	"time"
)

type ModelUsageHourly struct {
	Id           int     `json:"id" gorm:"primaryKey;autoIncrement"`
	ModelName    string  `json:"model_name" gorm:"type:varchar(255);index:idx_model_hour,priority:1;not null"`
	HourBucket   int64   `json:"hour_bucket" gorm:"bigint;index:idx_model_hour,priority:2;not null"`
	RequestCount int64   `json:"request_count" gorm:"default:0"`
	TokenCount   int64   `json:"token_count" gorm:"default:0"`
	UniqueUsers  int     `json:"unique_users" gorm:"default:0"`
	ErrorCount   int64   `json:"error_count" gorm:"default:0"`
	LatencyP50   float64 `json:"latency_p50" gorm:"default:0"`
	LatencyP95   float64 `json:"latency_p95" gorm:"default:0"`
	LatencyP99   float64 `json:"latency_p99" gorm:"default:0"`
	TtftP50      float64 `json:"ttft_p50" gorm:"default:0"`
	TtftP95      float64 `json:"ttft_p95" gorm:"default:0"`
	TtftP99      float64 `json:"ttft_p99" gorm:"default:0"`
}

func (ModelUsageHourly) TableName() string {
	return "model_usage_hourly"
}

// hourlyRawRow 聚合查询的中间结果
type hourlyRawRow struct {
	ModelName    string
	RequestCount int64
	TokenCount   int64
	UniqueUsers  int
	ErrorCount   int64
}

// percentile 在已排序的切片中计算百分位值（应用层计算，兼容所有数据库）
func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := p / 100.0 * float64(len(sorted)-1)
	lower := int(math.Floor(idx))
	upper := int(math.Ceil(idx))
	if lower == upper || upper >= len(sorted) {
		return sorted[lower]
	}
	frac := idx - float64(lower)
	return sorted[lower]*(1-frac) + sorted[upper]*frac
}

// AggregateHourlyUsage 对指定小时时间桶聚合 logs 表数据，写入 model_usage_hourly
func AggregateHourlyUsage(hourBucket time.Time) error {
	bucketTs := hourBucket.Unix()
	nextBucketTs := hourBucket.Add(time.Hour).Unix()

	// 1. 基础统计（请求数、Token、用户数、错误数）
	var rows []hourlyRawRow
	err := LOG_DB.Table("logs").
		Select("model_name, count(*) as request_count, "+
			"sum(prompt_tokens + completion_tokens) as token_count, "+
			"count(distinct user_id) as unique_users, "+
			"sum(CASE WHEN type = ? THEN 1 ELSE 0 END) as error_count", LogTypeError).
		Where("type IN (?, ?) AND created_at >= ? AND created_at < ?",
			LogTypeConsume, LogTypeError, bucketTs, nextBucketTs).
		Group("model_name").
		Scan(&rows).Error
	if err != nil {
		return err
	}

	if len(rows) == 0 {
		return nil // 无数据不写入
	}

	// 2. 延迟数据（按模型分组取原始值，在 Go 层计算百分位）
	type latencyRow struct {
		ModelName      string
		UseTime        int
		FirstTokenTime int
	}
	var latencies []latencyRow
	err = LOG_DB.Table("logs").
		Select("model_name, use_time, first_token_time").
		Where("type = ? AND created_at >= ? AND created_at < ? AND use_time > 0",
			LogTypeConsume, bucketTs, nextBucketTs).
		Scan(&latencies).Error
	if err != nil {
		return err
	}

	// 按模型名分组延迟
	latencyByModel := make(map[string][]float64)
	ttftByModel := make(map[string][]float64)
	for _, l := range latencies {
		// use_time 单位是秒，转为毫秒
		latencyByModel[l.ModelName] = append(latencyByModel[l.ModelName], float64(l.UseTime)*1000)
		if l.FirstTokenTime > 0 {
			ttftByModel[l.ModelName] = append(ttftByModel[l.ModelName], float64(l.FirstTokenTime))
		}
	}

	// 3. 构建并写入聚合记录
	for _, row := range rows {
		if row.ModelName == "" {
			continue
		}

		// 排序后计算百分位
		lats := latencyByModel[row.ModelName]
		sort.Float64s(lats)
		ttfts := ttftByModel[row.ModelName]
		sort.Float64s(ttfts)

		record := &ModelUsageHourly{
			ModelName:    row.ModelName,
			HourBucket:   bucketTs,
			RequestCount: row.RequestCount,
			TokenCount:   row.TokenCount,
			UniqueUsers:  row.UniqueUsers,
			ErrorCount:   row.ErrorCount,
			LatencyP50:   percentile(lats, 50),
			LatencyP95:   percentile(lats, 95),
			LatencyP99:   percentile(lats, 99),
			TtftP50:      percentile(ttfts, 50),
			TtftP95:      percentile(ttfts, 95),
			TtftP99:      percentile(ttfts, 99),
		}

		// upsert：如果已存在则更新
		existing := &ModelUsageHourly{}
		result := LOG_DB.Where("model_name = ? AND hour_bucket = ?", row.ModelName, bucketTs).First(existing)
		if result.Error == nil {
			LOG_DB.Model(existing).Updates(record)
		} else {
			LOG_DB.Create(record)
		}
	}

	return nil
}

// ModelUsageSummary 模型使用汇总（聚合表查询结果）
type ModelUsageSummary struct {
	ModelName    string  `json:"model_name"`
	RequestCount int64   `json:"request_count"`
	TokenCount   int64   `json:"token_count"`
	UniqueUsers  int64   `json:"unique_users"`
	ErrorCount   int64   `json:"error_count"`
	AvgLatencyMs float64 `json:"avg_latency_ms"`
	AvgTtftMs    float64 `json:"avg_ttft_ms"`
}

// GetModelUsageSummary 从聚合表查询模型使用汇总
func GetModelUsageSummary(startTs int64, endTs int64) ([]ModelUsageSummary, error) {
	var results []ModelUsageSummary
	err := LOG_DB.Table("model_usage_hourly").
		Select("model_name, "+
			"sum(request_count) as request_count, "+
			"sum(token_count) as token_count, "+
			"sum(unique_users) as unique_users, "+
			"sum(error_count) as error_count, "+
			"avg(latency_p50) as avg_latency_ms, "+
			"avg(ttft_p50) as avg_ttft_ms").
		Where("hour_bucket >= ? AND hour_bucket < ?", startTs, endTs).
		Group("model_name").
		Order("request_count desc").
		Scan(&results).Error
	return results, err
}

// CleanupOldHourlyUsage 清理超过指定天数的小时级数据
func CleanupOldHourlyUsage(retentionDays int) error {
	threshold := time.Now().Unix() - int64(retentionDays*86400)
	return LOG_DB.Where("hour_bucket < ?", threshold).Delete(&ModelUsageHourly{}).Error
}
