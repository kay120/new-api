package operation_setting

// 企业内部部署不使用充值/支付功能。以下变量仅用于：
// - USDExchangeRate：日志和 OpenAI 兼容 /dashboard/billing/* 接口的金额格式换算（tokens 模式下不生效）
// - Price：个别上游渠道查余额时的 CNY→USD 换算（企业自用通常不触发）
var (
	Price           = 7.3
	USDExchangeRate = 7.3
)
