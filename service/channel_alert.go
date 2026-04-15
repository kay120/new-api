package service

import (
	"fmt"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
)

// SendChannelAlert 将渠道状态变更异步推送到全局告警 Webhook。
// event: "disabled" | "enabled"
func SendChannelAlert(event string, channelId int, channelName string, reason string) {
	if !common.ChannelAlertWebhookEnabled || common.ChannelAlertWebhookURL == "" {
		return
	}

	title := fmt.Sprintf("[渠道%s] #%d %s", channelEventLabel(event), channelId, channelName)
	content := reason
	if content == "" {
		content = channelEventLabel(event)
	}

	notify := dto.NewNotify(
		"channel_"+event,
		title,
		content,
		[]interface{}{channelId, channelName},
	)

	go func() {
		if err := SendWebhookNotify(
			common.ChannelAlertWebhookURL,
			common.ChannelAlertWebhookSecret,
			notify,
		); err != nil {
			common.SysLog(fmt.Sprintf("channel alert webhook failed: %s", err.Error()))
		}
	}()
}

func channelEventLabel(event string) string {
	switch event {
	case "disabled":
		return "已禁用"
	case "enabled":
		return "已恢复"
	default:
		return event
	}
}
