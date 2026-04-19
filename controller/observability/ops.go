package observability

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"

	"github.com/gin-gonic/gin"
)

// TaskMeta 静态任务元数据：前端用来画"数据运维"页面。
type TaskMeta struct {
	Name          string `json:"name"`                     // 唯一 key（与 RecordTaskRun 一致）
	Label         string `json:"label"`                    // 展示名
	Description   string `json:"description"`              // 做什么
	Schedule      string `json:"schedule"`                 // 人类可读的调度描述
	ConfigEnv     string `json:"config_env"`               // 影响它的 env 名
	Manual        bool   `json:"manual"`                   // 是否支持手动触发
	ManualPath    string `json:"manual_path,omitempty"`    // POST 路径（admin-only）
	ManualExample string `json:"manual_example,omitempty"` // 示例参数说明
}

// 已接入 RecordTaskRun 的任务 + 一些仅展示元数据的任务。
// 新增任务时在此处注册一行（Manual=true 的还要挂到 router 里）。
var registeredTasks = []TaskMeta{
	{
		Name:          "summary_rollup",
		Label:         "每日聚合 Rollup",
		Description:   "把 logs 按 (日期 × 用户 × 分组 × 模型 × 渠道) 预聚合到 summary_daily，Dashboard 历史区间走这张表。",
		Schedule:      "每 30 分钟检查，每天 00:10 之后回填昨天 + 向前 2 天覆盖",
		ConfigEnv:     "SUMMARY_ROLLUP_ENABLED (设 false 可禁用)",
		Manual:        true,
		ManualPath:    "/api/ops/tasks/summary_rollup/trigger",
		ManualExample: "可选 ?days=7 只回填最近 7 天；不带参数则回填所有历史",
	},
	{
		Name:          "audit_log_retention",
		Label:         "审计日志保留",
		Description:   "删除 audit_logs 表中超过保留期的记录。",
		Schedule:      "每 6 小时一次",
		ConfigEnv:     "AUDIT_LOG_RETENTION_DAYS (默认 90，<=0 禁用)",
		Manual:        true,
		ManualPath:    "/api/ops/tasks/audit_log_retention/trigger",
		ManualExample: "无参",
	},
	{
		Name:        "channel_auto_test",
		Label:       "渠道自动健康检查",
		Description: "按配置频率自动测试所有渠道，失败自动熔断。",
		Schedule:    "按 CHANNEL_TEST_FREQUENCY 配置（分钟）",
		ConfigEnv:   "CHANNEL_TEST_FREQUENCY",
		Manual:      false,
	},
	{
		Name:        "codex_credential_refresh",
		Label:       "Codex 凭证刷新",
		Description: "检查 Codex OAuth token，即将过期（1 天内）时自动刷新。",
		Schedule:    "每 10 分钟",
		Manual:      false,
	},
}

// GetOpsTasks 返回所有任务 + 最近一次运行状态（已接入记录的才有）。
func GetOpsTasks(c *gin.Context) {
	runs, _ := model.LatestTaskRuns()
	runByName := make(map[string]model.TaskRun, len(runs))
	for _, r := range runs {
		runByName[r.TaskName] = r
	}

	type row struct {
		TaskMeta
		LatestRun *model.TaskRun `json:"latest_run,omitempty"`
	}
	out := make([]row, 0, len(registeredTasks))
	for _, t := range registeredTasks {
		r := row{TaskMeta: t}
		if v, ok := runByName[t.Name]; ok {
			v := v
			r.LatestRun = &v
		}
		out = append(out, r)
	}
	common.ApiSuccess(c, out)
}

// GetOpsTaskRuns 某任务的最近 N 次运行历史。
func GetOpsTaskRuns(c *gin.Context) {
	name := c.Param("name")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	runs, err := model.RecentTaskRuns(name, limit)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, runs)
}

// TriggerOpsTask 手动触发某个任务。白名单内任务都会写 task_runs 记录。
func TriggerOpsTask(c *gin.Context) {
	name := c.Param("name")
	switch name {
	case "summary_rollup":
		days, _ := strconv.Atoi(c.DefaultQuery("days", "0"))
		var rows int64
		model.RecordTaskRun("summary_rollup", func() (int64, error) {
			n, err := service.TriggerRollupBackfill(days)
			rows = n
			return n, err
		})
		common.ApiSuccess(c, gin.H{"rows_written": rows})

	case "audit_log_retention":
		var deleted int64
		model.RecordTaskRun("audit_log_retention", func() (int64, error) {
			n, err := service.TriggerAuditLogPurge()
			deleted = n
			return n, err
		})
		common.ApiSuccess(c, gin.H{"deleted": deleted})

	default:
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": fmt.Sprintf("task %q does not support manual trigger", name),
		})
	}
}
