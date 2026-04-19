package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// NoCacheAPI 给所有 /api/* 动态接口响应默认加 Cache-Control: no-store。
//
// 背景：浏览器默认会把无显式 Cache-Control 的 200/404 GET 响应缓存一段时间；
// 如果后端曾在某段时间返回过错误（比如路由还没部署时的 404），该条目可能
// 被缓存 7 天，之后就算后端修好了浏览器仍直接返旧 404 而不再发请求。
// 接口类响应天然不该被浏览器 / 中间代理缓存——OWASP 推荐认证后的响应
// 使用 no-store，防止共享缓存（公司代理、多账号浏览器）跨用户泄露。
//
// **Override 语义**：
// 这个头只是"兜底默认"。单个 handler 如果明确想让浏览器缓存（例如只读的
// 模型列表 / 字典数据），在 c.JSON() 之前再写一次会覆盖默认：
//
//	c.Header("Cache-Control", "private, max-age=60")
//	c.JSON(200, data)
//
// 后写的覆盖前面——body flush 之前 Gin 的 header map 可以任改。
//
// 作用范围：只 /api/*；/web 静态资源 / /v1 relay 流式接口不受影响。
func NoCacheAPI() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.Header("Cache-Control", "no-store")
			c.Header("Pragma", "no-cache")
		}
		c.Next()
	}
}
