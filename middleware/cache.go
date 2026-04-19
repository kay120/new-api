package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// Cache 给静态资源加一周长缓存；动态接口（/api、/v1、/oauth、/dashboard/billing）
// 强制 no-store，避免 NoRoute 产生的 404 / 错误响应被浏览器缓存 7 天。
//
// 背景：web-router 把 Cache() 挂在整个 Engine 上，未命中路由时会走到
// NoRoute → RelayNotFound 返回 OpenAI 风格 404；若带上 max-age=604800
// 浏览器会缓存这个 404，之后即使后端新增了路由也继续返旧 404。
func Cache() func(c *gin.Context) {
	return func(c *gin.Context) {
		uri := c.Request.RequestURI
		if strings.HasPrefix(uri, "/api/") ||
			strings.HasPrefix(uri, "/v1/") ||
			strings.HasPrefix(uri, "/oauth") ||
			strings.HasPrefix(uri, "/dashboard/billing") {
			c.Header("Cache-Control", "no-store")
			c.Next()
			return
		}
		if uri == "/" {
			c.Header("Cache-Control", "no-cache")
		} else {
			c.Header("Cache-Control", "max-age=604800") // one week
		}
		c.Header("Cache-Version", "b688f2fb5be447c25e5aa3bd063087a83db32a288bf6a4f35f2d8db310e40b14")
		c.Next()
	}
}
