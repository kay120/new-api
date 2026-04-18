package middleware

import (
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// SecurityHeaders 为所有响应注入通用安全响应头，提供 defense-in-depth。
//
// 默认行为：
//   - X-Content-Type-Options: nosniff
//   - X-Frame-Options: DENY（防 clickjacking）
//   - Referrer-Policy: strict-origin-when-cross-origin
//   - Cross-Origin-Opener-Policy: same-origin（降级 spectre 类副信道风险）
//
// 可选行为（由环境变量控制，避免破坏本地 HTTP 开发）：
//   - ENFORCE_HSTS=true → 加 Strict-Transport-Security（仅在站点全量 HTTPS 时打开）
//   - CSP_POLICY=<raw>  → 使用自定义 Content-Security-Policy；未设置则不下发，
//     避免误伤现有 /web/ 嵌入资源与 airouter 前端。
func SecurityHeaders() gin.HandlerFunc {
	enforceHSTS := strings.EqualFold(os.Getenv("ENFORCE_HSTS"), "true")
	csp := strings.TrimSpace(os.Getenv("CSP_POLICY"))
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Cross-Origin-Opener-Policy", "same-origin")
		if enforceHSTS {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		if csp != "" {
			c.Header("Content-Security-Policy", csp)
		}
		c.Next()
	}
}
