package service

import (
	"fmt"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/model"
)

// CheckUsageDecline 检测模型使用率下降
// 规则：连续 7 天使用量低于前 30 天均值的 30%
// 在聚合 cron 中每天 0 点调用
func CheckUsageDecline() {
	now := time.Now()
	// 前 30 天均值
	thirtyDaysAgo := now.AddDate(0, 0, -30).Unix()
	sevenDaysAgo := now.AddDate(0, 0, -7).Unix()
	nowTs := now.Unix()

	avgStats, err := model.GetModelUsageSummary(thirtyDaysAgo, sevenDaysAgo)
	if err != nil {
		common.SysError("usage decline check: failed to get 30d avg: " + err.Error())
		return
	}

	recentStats, err := model.GetModelUsageSummary(sevenDaysAgo, nowTs)
	if err != nil {
		common.SysError("usage decline check: failed to get 7d stats: " + err.Error())
		return
	}

	recentMap := make(map[string]int64)
	for _, s := range recentStats {
		recentMap[s.ModelName] = s.RequestCount
	}

	for _, avg := range avgStats {
		if avg.RequestCount == 0 {
			continue
		}
		// 30 天期间的日均请求
		dailyAvg := float64(avg.RequestCount) / 23.0 // 30-7=23 天
		// 最近 7 天的日均
		recentCount := recentMap[avg.ModelName]
		recentDailyAvg := float64(recentCount) / 7.0

		threshold := dailyAvg * 0.3
		if recentDailyAvg < threshold && dailyAvg > 5 { // 日均 >5 才有意义
			sendDeclineAlert(avg.ModelName, dailyAvg, recentDailyAvg)
		}
	}
}

func sendDeclineAlert(modelName string, prevAvg float64, currentAvg float64) {
	title := fmt.Sprintf("[使用率下降] %s", modelName)
	content := fmt.Sprintf(
		"模型 %s 最近 7 天日均请求 %.0f 次，相比前 30 天日均 %.0f 次下降了 %.0f%%。\n建议检查是否需要下线或替换。",
		modelName, currentAvg, prevAvg, (1-currentAvg/prevAvg)*100,
	)

	common.SysLog(fmt.Sprintf("[USAGE_DECLINE] %s: %.0f -> %.0f", modelName, prevAvg, currentAvg))

	if !common.ChannelAlertWebhookEnabled || common.ChannelAlertWebhookURL == "" {
		return
	}

	notify := dto.NewNotify("usage_decline", title, content, nil)
	go func() {
		if err := SendWebhookNotify(
			common.ChannelAlertWebhookURL,
			common.ChannelAlertWebhookSecret,
			notify,
		); err != nil {
			common.SysLog("usage decline webhook failed: " + err.Error())
		}
	}()
}
