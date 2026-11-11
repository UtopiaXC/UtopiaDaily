package http

import (
	"io"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strings"

	"github.com/UtopiaXC/UtopiaDaily/internal/common"
	"github.com/UtopiaXC/UtopiaDaily/internal/database"
	"github.com/UtopiaXC/UtopiaDaily/internal/logger"
	"github.com/UtopiaXC/UtopiaDaily/internal/security"
	"github.com/gin-gonic/gin"
)

type RouterDependencies struct {
	Mode       string
	JWTManager *security.JWTManager

	Common         *CommonController
	Auth           *AuthController
	Dashboard      *DashboardController
	Config         *ConfigController
	User           *UserController
	Scraper        *ScraperController
	Push           *PushController
	Schedule       *ScheduleController
	Newspaper      *NewspaperController
	Classification *ClassificationController
	Data           *DataController

	DashboardFS fs.FS
	NewspaperFS fs.FS
}

func NewRouter(deps RouterDependencies) *gin.Engine {
	if strings.EqualFold(deps.Mode, "debug") {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(logger.GinLoggerMiddleware())

	r.GET("/healthz", deps.Common.Health)

	api := r.Group("/api")
	api.Use(security.AuthMiddleware(deps.JWTManager))
	{
		commonGroup := api.Group("/common")
		{
			commonGroup.GET("/system-info", deps.Common.SystemInfo)
			commonGroup.GET("/site", deps.Common.SiteInfo)
			commonGroup.GET("/locales", deps.Common.Locales)
			commonGroup.GET("/i18n/locales", deps.Common.Locales)
			commonGroup.GET("/i18n/:locale", deps.Common.Bundle)
		}

		newspaperGroup := api.Group("/newspaper")
		{
			newspaperGroup.GET("/latest", deps.Newspaper.Latest)
		}

		authGroup := api.Group("/auth")
		{
			authGroup.GET("/setup/status", deps.Auth.SetupStatus)
			authGroup.POST("/setup", deps.Auth.SetupAdmin)
			authGroup.GET("/login-policy", deps.Auth.LoginPolicy)
			authGroup.GET("/captcha", deps.Auth.Captcha)
			authGroup.POST("/login", deps.Auth.Login)
			authGroup.POST("/refresh", deps.Auth.Refresh)
			authGroup.POST("/logout", deps.Auth.Logout)
			authGroup.POST("/register", deps.Auth.Register)
			authGroup.POST("/activate", deps.Auth.Activate)
			authGroup.POST("/forgot-password", deps.Auth.ForgotPassword)
			authGroup.POST("/reset-password", deps.Auth.ResetPassword)

			authed := authGroup.Group("")
			authed.Use(security.RequireAuth())
			{
				authed.GET("/me", deps.Auth.Me)
				authed.PATCH("/me", deps.Auth.UpdateProfile)
				authed.POST("/me/password", deps.Auth.ChangePassword)
				authed.POST("/me/logout-all", deps.Auth.LogoutAll)
			}
		}

		dash := api.Group("/dashboard")
		dash.Use(security.RequireAuth())
		{
			dash.GET("/layout/menu", deps.Dashboard.LayoutMenu)

			overviewGroup := dash.Group("")
			overviewGroup.Use(security.RequirePermission(database.PermSystemStatusView))
			{
				overviewGroup.GET("/overview", deps.Dashboard.Overview)
				overviewGroup.GET("/events", deps.Dashboard.Events)
			}

			sysCfgView := dash.Group("/system-config")
			sysCfgView.Use(security.RequirePermission(database.PermSystemConfigView))
			{
				sysCfgView.GET("", deps.Config.ListSystemConfigs)
				sysCfgView.GET("/classification-slot", deps.Config.ClassificationSlot)
			}

			sysCfgEdit := dash.Group("/system-config")
			sysCfgEdit.Use(security.RequirePermission(database.PermSystemConfigEdit))
			{
				sysCfgEdit.PUT("", deps.Config.UpdateSystemConfigsBatch)
				sysCfgEdit.PUT("/:key", deps.Config.UpdateSystemConfig)
				sysCfgEdit.POST("/smtp/test", deps.Config.TestSMTP)
			}

			feCfgView := dash.Group("/frontend-config")
			feCfgView.Use(security.RequirePermission(database.PermFrontendConfigView))
			{
				feCfgView.GET("", deps.Config.ListFrontendConfigs)
				feCfgView.GET("/modules", deps.Newspaper.ListModuleConfigs)
			}

			feCfgEdit := dash.Group("/frontend-config")
			feCfgEdit.Use(security.RequirePermission(database.PermFrontendConfigEdit))
			{
				feCfgEdit.PUT("", deps.Config.UpdateFrontendConfigsBatch)
				feCfgEdit.PUT("/modules", deps.Newspaper.SaveModuleConfigs)
				feCfgEdit.PUT("/:key", deps.Config.UpdateFrontendConfig)
			}

			umView := dash.Group("/user-manager")
			umView.Use(security.RequirePermission(database.PermUserManagerView))
			{
				umView.GET("/users", deps.User.ListUsers)
				umView.GET("/roles", deps.User.ListRoles)
				umView.GET("/permissions", deps.User.ListPermissions)
			}

			umEdit := dash.Group("/user-manager")
			umEdit.Use(security.RequirePermission(database.PermUserManagerEdit))
			{
				umEdit.POST("/users", deps.User.CreateUser)
				umEdit.PUT("/users/:id", deps.User.UpdateUser)
				umEdit.POST("/users/:id/password", deps.User.ResetUserPassword)
				umEdit.POST("/users/:id/ban", deps.User.BanUser)
				umEdit.POST("/users/:id/unfreeze", deps.User.UnfreezeUser)
				umEdit.DELETE("/users/:id", deps.User.DeleteUser)

				umEdit.POST("/roles", deps.User.CreateRole)
				umEdit.PUT("/roles/:id", deps.User.UpdateRole)
				umEdit.DELETE("/roles/:id", deps.User.DeleteRole)
			}

			scraperView := dash.Group("/scraper")
			scraperView.Use(security.RequirePermission(database.PermScraperView))
			{
				scraperView.GET("/modules", deps.Scraper.List)
				scraperView.GET("/modules/:id", deps.Scraper.Detail)
			}

			scraperEdit := dash.Group("/scraper")
			scraperEdit.Use(security.RequirePermission(database.PermScraperEdit))
			{
				scraperEdit.POST("/modules/reload", deps.Scraper.Reload)
				scraperEdit.POST("/modules/:id/enable", deps.Scraper.Enable)
				scraperEdit.POST("/modules/:id/disable", deps.Scraper.Disable)
				scraperEdit.POST("/modules/:id/prepare-env", deps.Scraper.PrepareEnv)
				scraperEdit.POST("/modules/:id/config", deps.Scraper.SaveConfig)
				scraperEdit.POST("/modules/:id/test", deps.Scraper.TestConfig)
				scraperEdit.POST("/modules/:id/tasks/:task/run", deps.Scraper.RunTask)
				scraperEdit.POST("/bilibili/qrcode/generate", deps.Scraper.GenerateBilibiliQR)
				scraperEdit.GET("/bilibili/qrcode/poll", deps.Scraper.PollBilibiliQR)
			}

			pushView := dash.Group("/push")
			pushView.Use(security.RequirePermission(database.PermPushModuleView))
			{
				pushView.GET("/modules", deps.Push.ListModules)
				pushView.GET("/modules/:id", deps.Push.Detail)
				pushView.GET("/logs", deps.Push.ListLogs)
			}

			pushEdit := dash.Group("/push")
			pushEdit.Use(security.RequirePermission(database.PermPushModuleEdit))
			{
				pushEdit.POST("/modules/reload", deps.Push.Reload)
				pushEdit.POST("/modules/:id/enable", deps.Push.Enable)
				pushEdit.POST("/modules/:id/disable", deps.Push.Disable)
				pushEdit.POST("/modules/:id/prepare-env", deps.Push.PrepareEnv)
				pushEdit.POST("/modules/:id/config", deps.Push.SaveConfig)
				pushEdit.POST("/modules/:id/test", deps.Push.TestConfig)
				pushEdit.POST("/modules/:id/tasks/:task/run", deps.Push.RunTask)
			}

			schedView := dash.Group("/schedule")
			schedView.Use(security.RequirePermission(database.PermScheduleView))
			{
				schedView.GET("", deps.Schedule.List)
			}

			schedEdit := dash.Group("/schedule")
			schedEdit.Use(security.RequirePermission(database.PermScheduleEdit))
			{
				schedEdit.PUT("", deps.Schedule.BatchUpdate)
				schedEdit.PUT("/:id", deps.Schedule.Update)
				schedEdit.POST("/:id/run", deps.Schedule.RunNow)
			}

			clsView := dash.Group("/classification")
			clsView.Use(security.RequirePermission(database.PermClassificationView))
			{
				clsView.GET("/overview", deps.Classification.Overview)
				clsView.GET("/profiles", deps.Classification.ListProfiles)
				clsView.GET("/modules/:id", deps.Classification.GetModuleConfig)
				clsView.GET("/queue/status", deps.Classification.QueueStatus)
				clsView.GET("/queue/tasks", deps.Classification.ListQueueTasks)
				clsView.GET("/queue/tasks/:id", deps.Classification.GetQueueTaskDetail)
				clsView.GET("/llm/ollama/models", deps.Classification.ListOllamaModels)
				clsView.GET("/jobs", deps.Classification.ListJobs)
				clsView.GET("/jobs/:id", deps.Classification.GetJob)
			}

			clsEdit := dash.Group("/classification")
			clsEdit.Use(security.RequirePermission(database.PermClassificationEdit))
			{
				clsEdit.POST("/onnx/prepare-env", deps.Classification.PrepareONNXEnv)
				clsEdit.POST("/onnx/models/download", deps.Classification.DownloadONNXModel)
				clsEdit.DELETE("/onnx/models/:name", deps.Classification.DeleteONNXModel)
				clsEdit.POST("/llm/ollama/pull", deps.Classification.PullOllamaModel)
				clsEdit.POST("/jobs/:id/cancel", deps.Classification.CancelJob)
				clsEdit.DELETE("/jobs/:id", deps.Classification.DismissJob)
				clsEdit.POST("/profiles", deps.Classification.CreateProfile)
				clsEdit.PUT("/profiles/:id", deps.Classification.UpdateProfile)
				clsEdit.DELETE("/profiles/:id", deps.Classification.DeleteProfile)
				clsEdit.POST("/profiles/:id/load-default", deps.Classification.LoadDefaultProfile)
				clsEdit.PUT("/modules/:id", deps.Classification.SaveModuleConfig)
				clsEdit.POST("/queue/pause", deps.Classification.PauseQueue)
				clsEdit.POST("/queue/resume", deps.Classification.ResumeQueue)
				clsEdit.POST("/queue/enqueue-untagged", deps.Classification.EnqueueUntagged)
				clsEdit.POST("/queue/cancel-pending", deps.Classification.CancelPending)
				clsEdit.POST("/queue/retry-failed", deps.Classification.RetryFailed)
				clsEdit.POST("/queue/clear-finished", deps.Classification.ClearFinished)
				clsEdit.POST("/queue/tasks/:id/cancel", deps.Classification.CancelTask)
				clsEdit.POST("/queue/tasks/:id/retry", deps.Classification.RetryTask)
			}

			dataView := dash.Group("/data")
			dataView.Use(security.RequirePermission(database.PermDataView))
			{
				dataView.GET("/articles", deps.Data.List)
				dataView.GET("/articles/:id", deps.Data.Detail)
				dataView.GET("/tags", deps.Data.ListTags)
			}

			dataEdit := dash.Group("/data")
			dataEdit.Use(security.RequirePermission(database.PermDataEdit))
			{
				dataEdit.POST("/articles/delete", deps.Data.DeleteBatch)
				dataEdit.DELETE("/articles/:id", deps.Data.DeleteOne)
				dataEdit.POST("/articles/:id/tags", deps.Data.AddTag)
				dataEdit.DELETE("/articles/:id/tags/:tag_id", deps.Data.RemoveTag)
				dataEdit.POST("/articles/reclassify", deps.Data.Reclassify)
			}

			userPush := dash.Group("/user-push")
			userPush.Use(security.RequirePermission(database.PermUserPushSettings))
			{
				userPush.GET("", deps.Push.ListUserSubscriptions)
				userPush.POST("", deps.Push.UpsertUserSubscription)
				userPush.DELETE("/:id", deps.Push.DeleteUserSubscription)
				userPush.POST("/test", deps.Push.TestUserPush)
			}
		}
	}

	r.GET("/dashboard", func(c *gin.Context) {
		serveStaticFS(c, deps.DashboardFS, "index.html", true)
	})
	r.GET("/dashboard/*filepath", func(c *gin.Context) {
		serveStaticFS(c, deps.DashboardFS, c.Param("filepath"), true)
	})
	r.GET("/", func(c *gin.Context) {
		serveStaticFS(c, deps.NewspaperFS, "index.html", true)
	})

	r.NoRoute(func(c *gin.Context) {
		reqPath := c.Request.URL.Path
		if strings.HasPrefix(reqPath, "/api/") || reqPath == "/api" {
			common.RespondError(c, http.StatusNotFound, common.ErrRecordNotFound)
			return
		}
		if reqPath == "/dashboard" || strings.HasPrefix(reqPath, "/dashboard/") {
			serveStaticFS(c, deps.DashboardFS, strings.TrimPrefix(reqPath, "/dashboard"), true)
			return
		}
		serveStaticFS(c, deps.NewspaperFS, reqPath, true)
	})

	return r
}

func serveStaticFS(c *gin.Context, staticFS fs.FS, relPath string, fallbackToIndex bool) {
	if staticFS == nil {
		c.Status(http.StatusNotFound)
		return
	}

	cleaned := strings.TrimPrefix(path.Clean("/"+relPath), "/")
	if cleaned == "" || cleaned == "." {
		cleaned = "index.html"
	}

	if file, err := staticFS.Open(cleaned); err == nil {
		defer func() { _ = file.Close() }()
		if stat, statErr := file.Stat(); statErr == nil && !stat.IsDir() {
			data, readErr := io.ReadAll(file)
			if readErr == nil {
				ctype := mime.TypeByExtension(path.Ext(cleaned))
				if ctype == "" {
					ctype = http.DetectContentType(data)
				}
				if strings.HasPrefix(cleaned, "assets/") {
					c.Header("Cache-Control", "public, max-age=31536000, immutable")
				}
				c.Data(http.StatusOK, ctype, data)
				return
			}
		}
	}

	if !fallbackToIndex {
		c.Status(http.StatusNotFound)
		return
	}

	indexData, err := fs.ReadFile(staticFS, "index.html")
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.Header("Cache-Control", "no-cache")
	c.Data(http.StatusOK, "text/html; charset=utf-8", indexData)
}
