package middleware

import (
	"os"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORS 返回跨源访问中间件。
//
// 行为：
//   - 未设置 CORS_ALLOW_ORIGINS 时，允许任意来源但禁用凭证（AllowCredentials=false），
//     避免 "AllowAllOrigins=true + AllowCredentials=true" 这种会导致任意网站向 API
//     发带凭证请求的致命组合。
//   - 设置了逗号分隔的 CORS_ALLOW_ORIGINS 时，仅放行该白名单，允许带凭证。
func CORS() gin.HandlerFunc {
	config := cors.DefaultConfig()
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"*"}

	raw := strings.TrimSpace(os.Getenv("CORS_ALLOW_ORIGINS"))
	if raw == "" {
		// 开发 / 单机默认：兼容任意来源，但禁用 credentials 以阻断 CSRF-via-XHR
		config.AllowAllOrigins = true
		config.AllowCredentials = false
		return cors.New(config)
	}

	origins := make([]string, 0)
	for _, o := range strings.Split(raw, ",") {
		if v := strings.TrimSpace(o); v != "" {
			origins = append(origins, v)
		}
	}
	config.AllowOrigins = origins
	config.AllowCredentials = true
	common.SysLog("CORS: allowlist = " + strings.Join(origins, ", "))
	return cors.New(config)
}

func PoweredBy() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-New-Api-Version", common.Version)
		c.Next()
	}
}
