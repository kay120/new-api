package router

import (
	"github.com/QuantumNous/new-api/controller/auth"
	"github.com/QuantumNous/new-api/controller/billing"
	channelctl "github.com/QuantumNous/new-api/controller/channel"
	"github.com/QuantumNous/new-api/controller/modelctl"
	"github.com/QuantumNous/new-api/controller/observability"
	"github.com/QuantumNous/new-api/controller/relayctl"
	"github.com/QuantumNous/new-api/controller/system"
	"github.com/QuantumNous/new-api/controller/userctl"
	"github.com/QuantumNous/new-api/middleware"

	// Import oauth package to register providers via init()
	_ "github.com/QuantumNous/new-api/oauth"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

func SetApiRouter(router *gin.Engine) {
	apiRouter := router.Group("/api")
	apiRouter.Use(middleware.RouteTag("api"))
	apiRouter.Use(gzip.Gzip(gzip.DefaultCompression))
	apiRouter.Use(middleware.BodyStorageCleanup()) // 清理请求体存储
	apiRouter.Use(middleware.GlobalAPIRateLimit())
	{
		apiRouter.GET("/setup", auth.GetSetup)
		apiRouter.POST("/setup", auth.PostSetup)
		apiRouter.GET("/status", system.GetStatus)
		apiRouter.GET("/uptime/status", observability.GetUptimeKumaStatus)
		apiRouter.GET("/models", middleware.UserAuth(), modelctl.DashboardListModels)
		apiRouter.GET("/status/test", middleware.AdminAuth(), system.TestStatus)
		apiRouter.GET("/notice", system.GetNotice)
		apiRouter.GET("/user-agreement", system.GetUserAgreement)
		apiRouter.GET("/privacy-policy", system.GetPrivacyPolicy)
		apiRouter.GET("/about", system.GetAbout)
		apiRouter.GET("/home_page_content", system.GetHomePageContent)
		apiRouter.GET("/pricing", middleware.TryUserAuth(), billing.GetPricing)
		apiRouter.GET("/verification", middleware.EmailVerificationRateLimit(), middleware.TurnstileCheck(), system.SendEmailVerification)
		apiRouter.GET("/reset_password", middleware.CriticalRateLimit(), middleware.TurnstileCheck(), system.SendPasswordResetEmail)
		apiRouter.POST("/user/reset", middleware.CriticalRateLimit(), system.ResetPassword)
		// OAuth routes - specific routes must come before :provider wildcard
		apiRouter.GET("/oauth/state", middleware.CriticalRateLimit(), auth.GenerateOAuthCode)
		apiRouter.POST("/oauth/email/bind", middleware.CriticalRateLimit(), userctl.EmailBind)
		// Standard OAuth providers (GitHub, Discord, OIDC, LinuxDO) - unified route
		apiRouter.GET("/oauth/:provider", middleware.CriticalRateLimit(), auth.HandleOAuth)
		apiRouter.GET("/ratio_config", middleware.CriticalRateLimit(), billing.GetRatioConfig)

		// Universal secure verification routes
		apiRouter.POST("/verify", middleware.UserAuth(), middleware.CriticalRateLimit(), auth.UniversalVerify)

		userRoute := apiRouter.Group("/user")
		{
			userRoute.POST("/register", middleware.CriticalRateLimit(), middleware.TurnstileCheck(), userctl.Register)
			userRoute.POST("/login", middleware.CriticalRateLimit(), middleware.TurnstileCheck(), userctl.Login)
			userRoute.POST("/login/2fa", middleware.CriticalRateLimit(), auth.Verify2FALogin)
			userRoute.POST("/passkey/login/begin", middleware.CriticalRateLimit(), auth.PasskeyLoginBegin)
			userRoute.POST("/passkey/login/finish", middleware.CriticalRateLimit(), auth.PasskeyLoginFinish)
			//userRoute.POST("/tokenlog", middleware.CriticalRateLimit(), controller.TokenLog)
			userRoute.GET("/logout", userctl.Logout)
			userRoute.GET("/groups", userctl.GetUserGroups)

			selfRoute := userRoute.Group("/")
			selfRoute.Use(middleware.UserAuth())
			{
				selfRoute.GET("/self/groups", userctl.GetUserGroups)
				selfRoute.GET("/self", userctl.GetSelf)
				selfRoute.GET("/models", userctl.GetUserModels)
				selfRoute.PUT("/self", userctl.UpdateSelf)
				selfRoute.DELETE("/self", userctl.DeleteSelf)
				selfRoute.GET("/token", userctl.GenerateAccessToken)
				selfRoute.GET("/passkey", auth.PasskeyStatus)
				selfRoute.POST("/passkey/register/begin", auth.PasskeyRegisterBegin)
				selfRoute.POST("/passkey/register/finish", auth.PasskeyRegisterFinish)
				selfRoute.POST("/passkey/verify/begin", auth.PasskeyVerifyBegin)
				selfRoute.POST("/passkey/verify/finish", auth.PasskeyVerifyFinish)
				selfRoute.DELETE("/passkey", auth.PasskeyDelete)
				selfRoute.GET("/aff", userctl.GetAffCode)
				selfRoute.PUT("/setting", userctl.UpdateUserSetting)

				// 2FA routes
				selfRoute.GET("/2fa/status", auth.Get2FAStatus)
				selfRoute.POST("/2fa/setup", auth.Setup2FA)
				selfRoute.POST("/2fa/enable", auth.Enable2FA)
				selfRoute.POST("/2fa/disable", auth.Disable2FA)
				selfRoute.POST("/2fa/backup_codes", auth.RegenerateBackupCodes)

			}

			adminRoute := userRoute.Group("/")
			adminRoute.Use(middleware.AdminAuth())
			{
				adminRoute.GET("/", userctl.GetAllUsers)
				adminRoute.GET("/search", userctl.SearchUsers)
				adminRoute.DELETE("/:id/bindings/:binding_type", userctl.AdminClearUserBinding)
				adminRoute.GET("/:id", userctl.GetUser)
				adminRoute.POST("/", userctl.CreateUser)
				adminRoute.POST("/manage", userctl.ManageUser)
				adminRoute.PUT("/", userctl.UpdateUser)
				adminRoute.DELETE("/:id", userctl.DeleteUser)
				adminRoute.DELETE("/:id/reset_passkey", auth.AdminResetPasskey)

				// Admin 2FA routes
				adminRoute.GET("/2fa/stats", auth.Admin2FAStats)
				adminRoute.DELETE("/:id/2fa", auth.AdminDisable2FA)
			}
		}

		optionRoute := apiRouter.Group("/option")
		optionRoute.Use(middleware.RootAuth())
		{
			optionRoute.GET("/", system.GetOptions)
			optionRoute.PUT("/", system.UpdateOption)
			optionRoute.GET("/channel_affinity_cache", channelctl.GetChannelAffinityCacheStats)
			optionRoute.DELETE("/channel_affinity_cache", channelctl.ClearChannelAffinityCache)
			optionRoute.POST("/rest_model_ratio", billing.ResetModelRatio)
			optionRoute.POST("/migrate_console_setting", system.MigrateConsoleSetting) // 用于迁移检测的旧键，下个版本会删除
		}

		performanceRoute := apiRouter.Group("/performance")
		performanceRoute.Use(middleware.RootAuth())
		{
			performanceRoute.GET("/stats", observability.GetPerformanceStats)
			performanceRoute.DELETE("/disk_cache", observability.ClearDiskCache)
			performanceRoute.POST("/reset_stats", observability.ResetPerformanceStats)
			performanceRoute.POST("/gc", observability.ForceGC)
			performanceRoute.GET("/logs", observability.GetLogFiles)
			performanceRoute.DELETE("/logs", observability.CleanupLogFiles)
		}
		ratioSyncRoute := apiRouter.Group("/ratio_sync")
		ratioSyncRoute.Use(middleware.RootAuth())
		{
			ratioSyncRoute.GET("/channels", billing.GetSyncableChannels)
			ratioSyncRoute.POST("/fetch", billing.FetchUpstreamRatios)
		}
		channelRoute := apiRouter.Group("/channel")
		channelRoute.Use(middleware.AdminAuth())
		{
			channelRoute.GET("/", channelctl.GetAllChannels)
			channelRoute.GET("/search", channelctl.SearchChannels)
			channelRoute.GET("/models", modelctl.ChannelListModels)
			channelRoute.GET("/models_enabled", modelctl.EnabledListModels)
			channelRoute.GET("/:id", channelctl.GetChannel)
			channelRoute.POST("/:id/key", middleware.RootAuth(), middleware.CriticalRateLimit(), middleware.DisableCache(), middleware.SecureVerificationRequired(), channelctl.GetChannelKey)
			channelRoute.GET("/test", channelctl.TestAllChannels)
			channelRoute.GET("/test/:id", channelctl.TestChannel)
			channelRoute.GET("/update_balance", billing.UpdateAllChannelsBalance)
			channelRoute.GET("/update_balance/:id", billing.UpdateChannelBalance)
			channelRoute.POST("/", channelctl.AddChannel)
			channelRoute.PUT("/", channelctl.UpdateChannel)
			channelRoute.DELETE("/disabled", channelctl.DeleteDisabledChannel)
			channelRoute.POST("/tag/disabled", channelctl.DisableTagChannels)
			channelRoute.POST("/tag/enabled", channelctl.EnableTagChannels)
			channelRoute.PUT("/tag", channelctl.EditTagChannels)
			channelRoute.DELETE("/:id", channelctl.DeleteChannel)
			channelRoute.POST("/batch", channelctl.DeleteChannelBatch)
			channelRoute.POST("/fix", channelctl.FixChannelsAbilities)
			channelRoute.GET("/fetch_models/:id", channelctl.FetchUpstreamModels)
			channelRoute.POST("/fetch_models", middleware.RootAuth(), channelctl.FetchModels)
			channelRoute.POST("/codex/oauth/start", auth.StartCodexOAuth)
			channelRoute.POST("/codex/oauth/complete", auth.CompleteCodexOAuth)
			channelRoute.POST("/:id/codex/oauth/start", auth.StartCodexOAuthForChannel)
			channelRoute.POST("/:id/codex/oauth/complete", auth.CompleteCodexOAuthForChannel)
			channelRoute.POST("/:id/codex/refresh", channelctl.RefreshCodexChannelCredential)
			channelRoute.GET("/:id/codex/usage", userctl.GetCodexChannelUsage)
			channelRoute.POST("/ollama/pull", channelctl.OllamaPullModel)
			channelRoute.POST("/ollama/pull/stream", channelctl.OllamaPullModelStream)
			channelRoute.DELETE("/ollama/delete", channelctl.OllamaDeleteModel)
			channelRoute.GET("/ollama/version/:id", channelctl.OllamaVersion)
			channelRoute.POST("/batch/tag", channelctl.BatchSetChannelTag)
			channelRoute.GET("/tag/models", channelctl.GetTagModels)
			channelRoute.POST("/copy/:id", channelctl.CopyChannel)
			channelRoute.POST("/multi_key/manage", channelctl.ManageMultiKeys)
			channelRoute.POST("/upstream_updates/apply", channelctl.ApplyChannelUpstreamModelUpdates)
			channelRoute.POST("/upstream_updates/apply_all", channelctl.ApplyAllChannelUpstreamModelUpdates)
			channelRoute.POST("/upstream_updates/detect", channelctl.DetectChannelUpstreamModelUpdates)
			channelRoute.POST("/upstream_updates/detect_all", channelctl.DetectAllChannelUpstreamModelUpdates)
		}
		tokenRoute := apiRouter.Group("/token")
		tokenRoute.Use(middleware.UserAuth())
		{
			tokenRoute.GET("/", userctl.GetAllTokens)
			tokenRoute.GET("/search", middleware.SearchRateLimit(), userctl.SearchTokens)
			tokenRoute.GET("/:id", userctl.GetToken)
			tokenRoute.POST("/:id/key", middleware.CriticalRateLimit(), middleware.DisableCache(), userctl.GetTokenKey)
			tokenRoute.POST("/", userctl.AddToken)
			tokenRoute.PUT("/", userctl.UpdateToken)
			tokenRoute.DELETE("/:id", userctl.DeleteToken)
			tokenRoute.POST("/batch", userctl.DeleteTokenBatch)
			tokenRoute.POST("/batch/keys", middleware.CriticalRateLimit(), middleware.DisableCache(), userctl.GetTokenKeysBatch)
		}

		usageRoute := apiRouter.Group("/usage")
		usageRoute.Use(middleware.CORS(), middleware.CriticalRateLimit())
		{
			tokenUsageRoute := usageRoute.Group("/token")
			tokenUsageRoute.Use(middleware.TokenAuthReadOnly())
			{
				tokenUsageRoute.GET("/", userctl.GetTokenUsage)
			}
		}

		logRoute := apiRouter.Group("/log")
		logRoute.GET("/", middleware.AdminAuth(), observability.GetAllLogs)
		logRoute.DELETE("/", middleware.AdminAuth(), observability.DeleteHistoryLogs)
		logRoute.GET("/stat", middleware.AdminAuth(), observability.GetLogsStat)
		logRoute.GET("/self/stat", middleware.UserAuth(), observability.GetLogsSelfStat)
		logRoute.GET("/channel_affinity_usage_cache", middleware.AdminAuth(), channelctl.GetChannelAffinityUsageCacheStats)
		logRoute.GET("/search", middleware.AdminAuth(), observability.SearchAllLogs)
		logRoute.GET("/self", middleware.UserAuth(), observability.GetUserLogs)
		logRoute.GET("/self/search", middleware.UserAuth(), middleware.SearchRateLimit(), observability.SearchUserLogs)

		dataRoute := apiRouter.Group("/data")
		dataRoute.GET("/", middleware.AdminAuth(), billing.GetAllQuotaDates)
		dataRoute.GET("/users", middleware.AdminAuth(), billing.GetQuotaDatesByUser)
		dataRoute.GET("/self", middleware.UserAuth(), billing.GetUserQuotaDates)

		logRoute.Use(middleware.CORS(), middleware.CriticalRateLimit())
		{
			logRoute.GET("/token", middleware.TokenAuthReadOnly(), observability.GetLogByKey)
		}
		groupRoute := apiRouter.Group("/group")
		groupRoute.Use(middleware.AdminAuth())
		{
			groupRoute.GET("/", userctl.GetGroups)
		}

		prefillGroupRoute := apiRouter.Group("/prefill_group")
		prefillGroupRoute.Use(middleware.AdminAuth())
		{
			prefillGroupRoute.GET("/", billing.GetPrefillGroups)
			prefillGroupRoute.POST("/", billing.CreatePrefillGroup)
			prefillGroupRoute.PUT("/", billing.UpdatePrefillGroup)
			prefillGroupRoute.DELETE("/:id", billing.DeletePrefillGroup)
		}

		auditRoute := apiRouter.Group("/audit_log")
		auditRoute.Use(middleware.AdminAuth())
		{
			auditRoute.GET("/", observability.GetAuditLogs)
		}

		analyticsRoute := apiRouter.Group("/analytics")
		{
			// Admin-only: 全局分析
			analyticsRoute.GET("/model-usage", middleware.AdminAuth(), observability.GetModelUsage)
			analyticsRoute.GET("/model-ranking", middleware.AdminAuth(), observability.GetModelRanking)
			analyticsRoute.GET("/missed-models", middleware.AdminAuth(), observability.GetMissedModels)
			analyticsRoute.GET("/channel-health", middleware.AdminAuth(), observability.GetChannelHealthOverview)
			analyticsRoute.GET("/channel-health/:id", middleware.AdminAuth(), observability.GetChannelHealthDetail)
			analyticsRoute.GET("/feedback-stats", middleware.AdminAuth(), observability.GetFeedbackStats)
			analyticsRoute.GET("/feedback-list", middleware.AdminAuth(), observability.GetRecentFeedbackList)
		}

		feedbackRoute := apiRouter.Group("/feedback")
		{
			feedbackRoute.POST("/", middleware.UserAuth(), observability.PostFeedback)
		}

		modelRequestRoute := apiRouter.Group("/model-request")
		{
			modelRequestRoute.POST("/", middleware.UserAuth(), modelctl.CreateModelRequestHandler)
			modelRequestRoute.GET("/self", middleware.UserAuth(), modelctl.GetUserModelRequestsHandler)
			modelRequestRoute.GET("/", middleware.AdminAuth(), modelctl.GetAllModelRequestsHandler)
			modelRequestRoute.GET("/stats", middleware.AdminAuth(), modelctl.GetModelRequestStatsHandler)
			modelRequestRoute.PUT("/:id", middleware.AdminAuth(), modelctl.UpdateModelRequestHandler)
		}

		taskRoute := apiRouter.Group("/task")
		{
			taskRoute.GET("/self", middleware.UserAuth(), relayctl.GetUserTask)
			taskRoute.GET("/", middleware.AdminAuth(), relayctl.GetAllTask)
		}

		vendorRoute := apiRouter.Group("/vendors")
		vendorRoute.Use(middleware.AdminAuth())
		{
			vendorRoute.GET("/", modelctl.GetAllVendors)
			vendorRoute.GET("/search", modelctl.SearchVendors)
			vendorRoute.GET("/:id", modelctl.GetVendorMeta)
			vendorRoute.POST("/", modelctl.CreateVendorMeta)
			vendorRoute.PUT("/", modelctl.UpdateVendorMeta)
			vendorRoute.DELETE("/:id", modelctl.DeleteVendorMeta)
		}

		modelsRoute := apiRouter.Group("/models")
		modelsRoute.Use(middleware.AdminAuth())
		{
			modelsRoute.GET("/sync_upstream/preview", modelctl.SyncUpstreamPreview)
			modelsRoute.POST("/sync_upstream", modelctl.SyncUpstreamModels)
			modelsRoute.GET("/missing", modelctl.GetMissingModels)
			modelsRoute.GET("/", modelctl.GetAllModelsMeta)
			modelsRoute.GET("/search", modelctl.SearchModelsMeta)
			modelsRoute.GET("/:id", modelctl.GetModelMeta)
			modelsRoute.POST("/", modelctl.CreateModelMeta)
			modelsRoute.PUT("/", modelctl.UpdateModelMeta)
			modelsRoute.DELETE("/:id", modelctl.DeleteModelMeta)
		}

		// Deployments (model deployment management)
		deploymentsRoute := apiRouter.Group("/deployments")
		deploymentsRoute.Use(middleware.AdminAuth())
		{
			deploymentsRoute.GET("/settings", system.GetModelDeploymentSettings)
			deploymentsRoute.POST("/settings/test-connection", system.TestIoNetConnection)
			deploymentsRoute.GET("/", system.GetAllDeployments)
			deploymentsRoute.GET("/search", system.SearchDeployments)
			deploymentsRoute.POST("/test-connection", system.TestIoNetConnection)
			deploymentsRoute.GET("/hardware-types", system.GetHardwareTypes)
			deploymentsRoute.GET("/locations", system.GetLocations)
			deploymentsRoute.GET("/available-replicas", system.GetAvailableReplicas)
			deploymentsRoute.POST("/price-estimation", system.GetPriceEstimation)
			deploymentsRoute.GET("/check-name", system.CheckClusterNameAvailability)
			deploymentsRoute.POST("/", system.CreateDeployment)

			deploymentsRoute.GET("/:id", system.GetDeployment)
			deploymentsRoute.GET("/:id/logs", system.GetDeploymentLogs)
			deploymentsRoute.GET("/:id/containers", system.ListDeploymentContainers)
			deploymentsRoute.GET("/:id/containers/:container_id", system.GetContainerDetails)
			deploymentsRoute.PUT("/:id", system.UpdateDeployment)
			deploymentsRoute.PUT("/:id/name", system.UpdateDeploymentName)
			deploymentsRoute.POST("/:id/extend", system.ExtendDeployment)
			deploymentsRoute.DELETE("/:id", system.DeleteDeployment)
		}

		// Report routes
		reportRoute := apiRouter.Group("/report")
		reportRoute.Use(middleware.UserAuth())
		{
			reportRoute.GET("/overview", observability.GetReportOverview)
			reportRoute.GET("/by-group", observability.GetReportByGroup)
			reportRoute.GET("/by-model", observability.GetReportByModel)
			reportRoute.GET("/by-user", observability.GetReportByUser)
		}
		// Export route (no session auth needed, uses UserAuth)
		apiRouter.GET("/report/export", middleware.UserAuth(), observability.ExportReportCSV)
	}
}
