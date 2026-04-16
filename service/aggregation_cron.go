package service

import (
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
)

var aggregationOnce sync.Once

// StartAggregationCron 启动小时级聚合定时任务
// 每小时运行一次，聚合上一小时的 logs 数据到 model_usage_hourly
// 同时清理超过 90 天的旧数据
func StartAggregationCron() {
	if !common.IsMasterNode {
		return
	}
	aggregationOnce.Do(func() {
		go func() {
			// 启动时先聚合上一小时的数据
			runAggregation()

			for {
				now := time.Now()
				// 等到下一个整点 + 5 分钟（给 logs 写入留缓冲）
				nextHour := now.Truncate(time.Hour).Add(time.Hour).Add(5 * time.Minute)
				sleepDuration := nextHour.Sub(now)
				time.Sleep(sleepDuration)

				runAggregation()
			}
		}()
	})
}

func runAggregation() {
	now := time.Now()
	lastHour := now.Truncate(time.Hour).Add(-time.Hour)

	common.SysLog("starting hourly model usage aggregation for " + lastHour.Format("2006-01-02 15:04"))

	if err := model.AggregateHourlyUsage(lastHour); err != nil {
		common.SysError("hourly aggregation failed: " + err.Error())
	} else {
		common.SysLog("hourly aggregation completed")
	}

	// 检查用量里程碑
	CheckMilestones()

	// 每天 0 点执行清理 + 使用率下降检测
	if now.Hour() == 0 {
		CheckUsageDecline()
		common.SysLog("cleaning up old aggregation data (>90 days)")
		if err := model.CleanupOldHourlyUsage(90); err != nil {
			common.SysError("cleanup old hourly usage failed: " + err.Error())
		}
		if err := model.CleanupOldMissEvents(90); err != nil {
			common.SysError("cleanup old miss events failed: " + err.Error())
		}
		if err := model.CleanupOldHealthChecks(90); err != nil {
			common.SysError("cleanup old health checks failed: " + err.Error())
		}
	}
}
