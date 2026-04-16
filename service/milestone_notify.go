package service

import (
	"fmt"
	"sync"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/model"
)

var (
	milestoneOnce     sync.Once
	milestonesReached = make(map[string]bool)
)

// CheckMilestones 检查是否达到关键里程碑并通知
// 在聚合 cron 中调用，每小时检查一次
func CheckMilestones() {
	milestoneOnce.Do(func() {
		// 启动时从数据库加载已触发的里程碑（简单实现：用内存缓存，重启后重新检查）
	})

	checkUserMilestone()
	checkRequestMilestone()
}

func checkUserMilestone() {
	thresholds := []int{1, 5, 10, 20, 50, 100}
	count := model.GetActiveUserCount()

	for _, t := range thresholds {
		key := fmt.Sprintf("users_%d", t)
		if count >= t && !milestonesReached[key] {
			milestonesReached[key] = true
			sendMilestone(
				fmt.Sprintf("用户里程碑: %d 个活跃用户", t),
				fmt.Sprintf("恭喜！网关已有 %d 个活跃用户（当前 %d 个）。", t, count),
			)
		}
	}
}

func checkRequestMilestone() {
	thresholds := []int64{100, 1000, 5000, 10000, 50000, 100000}
	// 查询今日请求量
	count := model.GetTodayRequestCount()

	for _, t := range thresholds {
		key := fmt.Sprintf("daily_requests_%d", t)
		if count >= t && !milestonesReached[key] {
			milestonesReached[key] = true
			sendMilestone(
				fmt.Sprintf("请求里程碑: 日请求量突破 %d", t),
				fmt.Sprintf("今日请求量已达 %d 次。", count),
			)
		}
	}
}

func sendMilestone(title string, content string) {
	if !common.ChannelAlertWebhookEnabled || common.ChannelAlertWebhookURL == "" {
		common.SysLog(fmt.Sprintf("[MILESTONE] %s - %s", title, content))
		return
	}

	notify := dto.NewNotify("milestone", title, content, nil)
	go func() {
		if err := SendWebhookNotify(
			common.ChannelAlertWebhookURL,
			common.ChannelAlertWebhookSecret,
			notify,
		); err != nil {
			common.SysLog(fmt.Sprintf("milestone webhook failed: %s", err.Error()))
		}
	}()
}
