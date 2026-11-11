package http

import (
	"net/http"
	"strconv"

	"github.com/UtopiaXC/UtopiaDaily/internal/common"
	"github.com/UtopiaXC/UtopiaDaily/internal/security"
	"github.com/UtopiaXC/UtopiaDaily/internal/service"
	"github.com/gin-gonic/gin"
)

type PushController struct {
	pushSvc    *service.PushService
	scraperSvc *service.ScraperService
}

func NewPushController(pushSvc *service.PushService, scraperSvc *service.ScraperService) *PushController {
	return &PushController{
		pushSvc:    pushSvc,
		scraperSvc: scraperSvc,
	}
}

func (ctl *PushController) ListModules(c *gin.Context) {
	list, err := ctl.pushSvc.ListModules()
	if err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, list)
}

func (ctl *PushController) Detail(c *gin.Context) {
	detail, err := ctl.pushSvc.GetModuleDetail(c.Param("id"))
	if err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, detail)
}

func (ctl *PushController) Enable(c *gin.Context) {
	if err := ctl.pushSvc.EnableModule(c.Request.Context(), c.Param("id")); err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, gin.H{"status": "ok"})
}

func (ctl *PushController) Disable(c *gin.Context) {
	if err := ctl.pushSvc.DisableModule(c.Param("id")); err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, gin.H{"status": "ok"})
}

func (ctl *PushController) Reload(c *gin.Context) {
	if err := ctl.pushSvc.ReloadAll(); err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, gin.H{"status": "ok"})
}

func (ctl *PushController) PrepareEnv(c *gin.Context) {
	out, err := ctl.pushSvc.PrepareEnv(c.Request.Context(), c.Param("id"))
	if err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, gin.H{"status": "ok", "output": out})
}

func (ctl *PushController) SaveConfig(c *gin.Context) {
	var payload map[string]any
	if err := c.ShouldBindJSON(&payload); err != nil {
		common.RespondError(c, http.StatusBadRequest, common.ErrValidationFailed)
		return
	}
	if err := ctl.pushSvc.SaveConfig(c.Param("id"), payload); err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, gin.H{"status": "ok"})
}

func (ctl *PushController) TestConfig(c *gin.Context) {
	var payload map[string]any
	_ = c.ShouldBindJSON(&payload)
	res, err := ctl.pushSvc.TestConfig(c.Request.Context(), c.Param("id"), payload)
	if err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, res)
}

func (ctl *PushController) RunTask(c *gin.Context) {
	if err := ctl.pushSvc.ExecuteModulePushTask(c.Request.Context(), c.Param("id"), c.Param("task")); err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, gin.H{"status": "ok"})
}

func (ctl *PushController) ListLogs(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "30"))
	logs, err := ctl.pushSvc.ListDeliveryLogs(limit)
	if err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, logs)
}

func (ctl *PushController) ListUserSubscriptions(c *gin.Context) {
	user, ok := security.CurrentUser(c)
	if !ok {
		common.RespondError(c, http.StatusUnauthorized, common.ErrAuthUnauthorized)
		return
	}
	subs, err := ctl.pushSvc.ListUserSubscriptions(user.ID)
	if err != nil {
		common.RespondFromError(c, err)
		return
	}
	pushMods, _ := ctl.pushSvc.ListModules()
	scraperMods, _ := ctl.scraperSvc.ListModules()
	common.RespondSuccess(c, gin.H{
		"subscriptions":   subs,
		"push_modules":    pushMods,
		"scraper_modules": scraperMods,
	})
}

func (ctl *PushController) UpsertUserSubscription(c *gin.Context) {
	user, ok := security.CurrentUser(c)
	if !ok {
		common.RespondError(c, http.StatusUnauthorized, common.ErrAuthUnauthorized)
		return
	}
	var req service.UserPushSubscriptionDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, http.StatusBadRequest, common.ErrValidationFailed)
		return
	}
	res, err := ctl.pushSvc.UpsertUserSubscription(user.ID, req)
	if err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, res)
}

func (ctl *PushController) DeleteUserSubscription(c *gin.Context) {
	user, ok := security.CurrentUser(c)
	if !ok {
		common.RespondError(c, http.StatusUnauthorized, common.ErrAuthUnauthorized)
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		common.RespondError(c, http.StatusBadRequest, common.ErrValidationFailed)
		return
	}
	if err := ctl.pushSvc.DeleteUserSubscription(user.ID, uint(id)); err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, gin.H{"status": "ok"})
}

func (ctl *PushController) TestUserPush(c *gin.Context) {
	user, ok := security.CurrentUser(c)
	if !ok {
		common.RespondError(c, http.StatusUnauthorized, common.ErrAuthUnauthorized)
		return
	}
	var req struct {
		PushModuleID string `json:"push_module_id"`
		Target       string `json:"target"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, http.StatusBadRequest, common.ErrValidationFailed)
		return
	}
	if _, err := ctl.pushSvc.TestUserSubscription(c.Request.Context(), user, req.PushModuleID, req.Target); err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, gin.H{"status": "ok"})
}
