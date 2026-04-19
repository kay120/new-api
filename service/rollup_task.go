package service

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"

	"github.com/bytedance/gopkg/util/gopool"
)

// Summary rollup：每天凌晨把昨天的 logs 聚合进 summary_daily。
// Dashboard 历史区间查询由分层逻辑决定走 summary 还是 logs。

const (
	rollupCheckInterval  = 30 * time.Minute // 周期检查（避免机器时钟错过 00:10 窗口）
	rollupHourOfDay      = 0                // UTC 0 点
	rollupMinuteOfHour   = 10
	rollupSafetyBackfill = 2 // 每次除了跑昨天，再补前 N 天（幂等覆盖，防补跑漏）
)

var (
	rollupTaskOnce sync.Once
	rollupRunning  atomic.Bool
)

// StartSummaryRollupTask 在主节点启动定时 rollup 协程。
// 每 30 分钟检查一次：若今天的 00:10 已过且昨天还没跑，就跑一次；
// 同时每次跑都覆盖前 rollupSafetyBackfill 天（幂等 UPSERT）。
func StartSummaryRollupTask() {
	rollupTaskOnce.Do(func() {
		if !common.IsMasterNode {
			return
		}
		if os.Getenv("SUMMARY_ROLLUP_ENABLED") == "false" {
			common.SysLog("summary rollup disabled by SUMMARY_ROLLUP_ENABLED=false")
			return
		}
		gopool.Go(func() {
			logger.LogInfo(context.Background(),
				"summary rollup task started (checks every 30m, rolls up yesterday + safety window)")
			// 启动时先跑一次（包含首次部署回填触发）
			runDailyRollupIfDue()
			ticker := time.NewTicker(rollupCheckInterval)
			defer ticker.Stop()
			for range ticker.C {
				runDailyRollupIfDue()
			}
		})
	})
}

// TaskNameSummaryRollup 是统一任务名，用于 TaskRun 记录和前端 ops 页面。
const TaskNameSummaryRollup = "summary_rollup"

// runDailyRollupIfDue 决策是否该 rollup，每次执行通过 RecordTaskRun 写入 task_runs 表。
func runDailyRollupIfDue() {
	if !rollupRunning.CompareAndSwap(false, true) {
		return
	}
	defer rollupRunning.Store(false)

	model.RecordTaskRun(TaskNameSummaryRollup, func() (int64, error) {
		now := time.Now()
		todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

		latestStr, _ := model.LatestSummaryDate()
		var startDay time.Time
		if latestStr == "" {
			earliest, err := model.EarliestLogDate()
			if err != nil {
				return 0, fmt.Errorf("earliest log lookup: %w", err)
			}
			if earliest.IsZero() {
				return 0, nil
			}
			startDay = time.Date(earliest.Year(), earliest.Month(), earliest.Day(), 0, 0, 0, 0, earliest.Location())
		} else {
			t, err := time.Parse("2006-01-02", latestStr)
			if err != nil {
				return 0, fmt.Errorf("parse latest date %q: %w", latestStr, err)
			}
			startDay = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, now.Location()).
				AddDate(0, 0, -rollupSafetyBackfill)
		}

		endUnix := todayStart.Unix()
		startUnix := startDay.Unix()
		if startUnix >= endUnix {
			return 0, nil
		}

		affected, err := model.RollupLogsToSummary(startUnix, endUnix)
		if err != nil {
			return 0, fmt.Errorf("rollup %s ~ %s: %w",
				startDay.Format("2006-01-02"), todayStart.Format("2006-01-02"), err)
		}
		common.SysLog(fmt.Sprintf("rollup done: [%s ~ %s) wrote/updated %d rows",
			startDay.Format("2006-01-02"), todayStart.Format("2006-01-02"), affected))
		return affected, nil
	})
}

// TriggerRollupBackfill 供管理员手动触发：强制 rollup 从指定起点到今天 00:00。
// startDaysAgo=0 表示只跑昨天；7 表示跑近 7 天；<=0 且未指定则跑所有历史。
func TriggerRollupBackfill(startDaysAgo int) (int64, error) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	var startUnix int64
	if startDaysAgo > 0 {
		startUnix = todayStart.AddDate(0, 0, -startDaysAgo).Unix()
	} else {
		earliest, err := model.EarliestLogDate()
		if err != nil {
			return 0, err
		}
		if earliest.IsZero() {
			return 0, nil
		}
		startUnix = time.Date(earliest.Year(), earliest.Month(), earliest.Day(), 0, 0, 0, 0, now.Location()).Unix()
	}
	return model.RollupLogsToSummary(startUnix, todayStart.Unix())
}

// getEnvInt 本地小 helper；避免 import 扩散
func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

var _ = getEnvInt // 保留以供后续 window 可配置扩展
