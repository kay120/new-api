package router

import (
	"github.com/QuantumNous/new-api/controller/billing"
	"github.com/QuantumNous/new-api/middleware"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

func SetDashboardRouter(router *gin.Engine) {
	apiRouter := router.Group("/")
	apiRouter.Use(middleware.RouteTag("old_api"))
	apiRouter.Use(gzip.Gzip(gzip.DefaultCompression))
	apiRouter.Use(middleware.GlobalAPIRateLimit())
	apiRouter.Use(middleware.CORS())
	apiRouter.Use(middleware.TokenAuth())
	{
		apiRouter.GET("/dashboard/billing/subscription", billing.GetSubscription)
		apiRouter.GET("/v1/dashboard/billing/subscription", billing.GetSubscription)
		apiRouter.GET("/dashboard/billing/usage", billing.GetUsage)
		apiRouter.GET("/v1/dashboard/billing/usage", billing.GetUsage)
	}
}
