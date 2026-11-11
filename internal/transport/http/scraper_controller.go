package http

import (
	"net/http"

	"github.com/UtopiaXC/UtopiaDaily/internal/common"
	"github.com/UtopiaXC/UtopiaDaily/internal/service"
	"github.com/gin-gonic/gin"
)

type ScraperController struct {
	scraperSvc *service.ScraperService
}

func NewScraperController(scraperSvc *service.ScraperService) *ScraperController {
	return &ScraperController{scraperSvc: scraperSvc}
}

func (ctl *ScraperController) List(c *gin.Context) {
	list, err := ctl.scraperSvc.ListModules()
	if err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, list)
}

func (ctl *ScraperController) Detail(c *gin.Context) {
	detail, err := ctl.scraperSvc.GetModuleDetail(c.Param("id"))
	if err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, detail)
}

func (ctl *ScraperController) Reload(c *gin.Context) {
	if err := ctl.scraperSvc.ReloadAll(); err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, gin.H{"status": "ok"})
}

func (ctl *ScraperController) Enable(c *gin.Context) {
	if err := ctl.scraperSvc.EnableModule(c.Request.Context(), c.Param("id")); err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, gin.H{"status": "ok"})
}

func (ctl *ScraperController) Disable(c *gin.Context) {
	if err := ctl.scraperSvc.DisableModule(c.Param("id")); err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, gin.H{"status": "ok"})
}

func (ctl *ScraperController) PrepareEnv(c *gin.Context) {
	out, err := ctl.scraperSvc.PrepareEnv(c.Request.Context(), c.Param("id"))
	if err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, gin.H{"status": "ok", "output": out})
}

func (ctl *ScraperController) SaveConfig(c *gin.Context) {
	var payload map[string]any
	if err := c.ShouldBindJSON(&payload); err != nil {
		common.RespondError(c, http.StatusBadRequest, common.ErrValidationFailed)
		return
	}
	if err := ctl.scraperSvc.SaveConfig(c.Param("id"), payload); err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, gin.H{"status": "ok"})
}

func (ctl *ScraperController) TestConfig(c *gin.Context) {
	var payload map[string]any
	_ = c.ShouldBindJSON(&payload)
	res, err := ctl.scraperSvc.TestConfig(c.Request.Context(), c.Param("id"), payload)
	if err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, res)
}

func (ctl *ScraperController) RunTask(c *gin.Context) {
	inserted, err := ctl.scraperSvc.RunTask(c.Request.Context(), c.Param("id"), c.Param("task"))
	if err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, gin.H{"status": "ok", "inserted": inserted})
}
