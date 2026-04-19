package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// NoCacheAPI 给所有 /api/* 动态接口响应加 Cache-Control: no-store。
//
// 原因：浏览器默认会把无显式 Cache-Control 的 200/404 GET 响应缓存一段时间，
// 如果后端在某段时间返回过错误（比如路由还没部署时的 404），该条目会被
// 缓存，之后就算后端修好了浏览器仍直接返旧 404 而不再发请求。
// 接口类响应天然不该缓存——statically cached by intermediate proxy 也会引发
// 登录信息错位等 security 风险。
//
// 该中间件放在 Gin 全局链路上，对所有请求生效但只对 /api/ 下发 no-store，
// 不影响 /web 静态资源或 /assets 的正常缓存。
func NoCacheAPI() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.Header("Cache-Control", "no-store")
			c.Header("Pragma", "no-cache")
		}
		c.Next()
	}
}
