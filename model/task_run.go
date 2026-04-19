package model

import (
	"time"

	"github.com/QuantumNous/new-api/common"
)

// TaskRun 记录后台任务的一次执行（由 RecordTaskRun 包装写入）。
// 统一"数据运维"页面所有定时任务的状态来源。
type TaskRun struct {
	Id           int    `json:"id" gorm:"primaryKey"`
	TaskName     string `json:"task_name" gorm:"index:idx_task_started,priority:1;type:varchar(64);not null"`
	StartedAt    int64  `json:"started_at" gorm:"index:idx_task_started,priority:2;bigint"`
	FinishedAt   int64  `json:"finished_at" gorm:"default:0"`
	DurationMs   int64  `json:"duration_ms" gorm:"default:0"`
	Status       string `json:"status" gorm:"type:varchar(16);default:'running'"` // running / success / failed
	Message      string `json:"message" gorm:"type:text"`                         // 错误信息或结果 summary
	AffectedRows int64  `json:"affected_rows" gorm:"default:0"`
}

// LatestTaskRuns 返回每个 task_name 的最近一次运行（按 task_name 分组取 max(id)）。
func LatestTaskRuns() ([]TaskRun, error) {
	// 两次查询：先取每个 name 的 max(id)，再 IN 取全量字段。
	// 避免数据库方言差异（mysql / pg 的 DISTINCT ON / ROW_NUMBER 写法不同）。
	var maxIds []int
	if err := DB.Table("task_runs").
		Select("MAX(id) AS id").
		Group("task_name").
		Scan(&maxIds).Error; err != nil {
		return nil, err
	}
	if len(maxIds) == 0 {
		return nil, nil
	}
	var runs []TaskRun
	if err := DB.Where("id IN ?", maxIds).Order("started_at desc").Find(&runs).Error; err != nil {
		return nil, err
	}
	return runs, nil
}

// RecentTaskRuns 返回指定 task_name 的最近 N 次运行。
func RecentTaskRuns(taskName string, limit int) ([]TaskRun, error) {
	if limit <= 0 {
		limit = 20
	}
	var runs []TaskRun
	err := DB.Where("task_name = ?", taskName).
		Order("id desc").
		Limit(limit).
		Find(&runs).Error
	return runs, err
}

// RecordTaskRun 包装一次任务执行：自动写 running → success/failed 两阶段记录。
// fn 返回 (影响行数, 错误)；nil 表示成功。
func RecordTaskRun(taskName string, fn func() (int64, error)) {
	started := time.Now()
	run := &TaskRun{
		TaskName:  taskName,
		StartedAt: started.Unix(),
		Status:    "running",
	}
	if err := DB.Create(run).Error; err != nil {
		common.SysError("task_run create failed: " + err.Error())
		// 即使写库失败，任务本身仍照跑
	}

	rows, runErr := func() (int64, error) {
		// 防御：fn panic 不影响后续记录
		defer func() {
			if r := recover(); r != nil {
				common.SysError("task panic: " + taskName)
			}
		}()
		return fn()
	}()

	finished := time.Now()
	run.FinishedAt = finished.Unix()
	run.DurationMs = finished.Sub(started).Milliseconds()
	run.AffectedRows = rows
	if runErr != nil {
		run.Status = "failed"
		run.Message = runErr.Error()
	} else {
		run.Status = "success"
	}
	if run.Id > 0 {
		DB.Save(run)
	}
}
