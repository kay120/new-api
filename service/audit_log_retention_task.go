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

const (
	defaultAuditLogRetentionDays = 90
	auditLogPurgeTickInterval    = 6 * time.Hour
)

var (
	auditLogPurgeOnce    sync.Once
	auditLogPurgeRunning atomic.Bool
)

// getAuditLogRetentionDays 从环境变量读取保留天数，缺省/非法时返回默认 90。
// 设置为 0 或负数表示"永不清理"。
func getAuditLogRetentionDays() int {
	raw := os.Getenv("AUDIT_LOG_RETENTION_DAYS")
	if raw == "" {
		return defaultAuditLogRetentionDays
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		common.SysError(fmt.Sprintf("AUDIT_LOG_RETENTION_DAYS invalid (%q), using default %d", raw, defaultAuditLogRetentionDays))
		return defaultAuditLogRetentionDays
	}
	return n
}

// StartAuditLogRetentionTask 在主节点上启动审计日志保留期清理协程。
// 每 6 小时跑一次 purge；保留期通过 AUDIT_LOG_RETENTION_DAYS 环境变量控制。
func StartAuditLogRetentionTask() {
	auditLogPurgeOnce.Do(func() {
		if !common.IsMasterNode {
			return
		}
		days := getAuditLogRetentionDays()
		if days <= 0 {
			common.SysLog("audit log retention task disabled (AUDIT_LOG_RETENTION_DAYS <= 0)")
			return
		}

		gopool.Go(func() {
			logger.LogInfo(context.Background(), fmt.Sprintf(
				"audit log retention task started: keep %d days, tick=%s",
				days, auditLogPurgeTickInterval,
			))
			ticker := time.NewTicker(auditLogPurgeTickInterval)
			defer ticker.Stop()

			runAuditLogPurgeOnce()
			for range ticker.C {
				runAuditLogPurgeOnce()
			}
		})
	})
}

// TaskNameAuditLogRetention 统一任务名。
const TaskNameAuditLogRetention = "audit_log_retention"

// TriggerAuditLogPurge 与定时任务共用保留策略的手动触发入口。
func TriggerAuditLogPurge() (int64, error) {
	days := getAuditLogRetentionDays()
	if days <= 0 {
		return 0, nil
	}
	cutoff := time.Now().Add(-time.Duration(days) * 24 * time.Hour).Unix()
	return model.PurgeAuditLogsOlderThan(cutoff)
}

func runAuditLogPurgeOnce() {
	if !auditLogPurgeRunning.CompareAndSwap(false, true) {
		return
	}
	defer auditLogPurgeRunning.Store(false)

	model.RecordTaskRun(TaskNameAuditLogRetention, func() (int64, error) {
		days := getAuditLogRetentionDays()
		if days <= 0 {
			return 0, nil
		}
		cutoff := time.Now().Add(-time.Duration(days) * 24 * time.Hour).Unix()
		deleted, err := model.PurgeAuditLogsOlderThan(cutoff)
		if err != nil {
			return 0, err
		}
		if deleted > 0 {
			common.SysLog(fmt.Sprintf("audit log purge: deleted %d rows older than %d days", deleted, days))
		}
		return deleted, nil
	})
}
