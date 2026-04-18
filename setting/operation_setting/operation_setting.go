package operation_setting

import "strings"

var DemoSiteEnabled = false

// SelfUseModeEnabled 企业内部模式开关（默认开启）。
// 为 true 时：未配置倍率的模型仍可使用、无须要求每个模型都配置价格。
// 计费扣费/额度检查已从 NewBillingSession 彻底移除，不再依赖此开关。
var SelfUseModeEnabled = true

var AutomaticDisableKeywords = []string{
	"Your credit balance is too low",
	"This organization has been disabled.",
	"You exceeded your current quota",
	"Permission denied",
	"The security token included in the request is invalid",
	"Operation not allowed",
	"Your account is not authorized",
}

func AutomaticDisableKeywordsToString() string {
	return strings.Join(AutomaticDisableKeywords, "\n")
}

func AutomaticDisableKeywordsFromString(s string) {
	AutomaticDisableKeywords = []string{}
	ak := strings.Split(s, "\n")
	for _, k := range ak {
		k = strings.TrimSpace(k)
		k = strings.ToLower(k)
		if k != "" {
			AutomaticDisableKeywords = append(AutomaticDisableKeywords, k)
		}
	}
}
