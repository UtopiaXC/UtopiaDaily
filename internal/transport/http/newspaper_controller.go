package http

import (
	"net/http"

	"github.com/UtopiaXC/UtopiaDaily/internal/common"
	"github.com/UtopiaXC/UtopiaDaily/internal/database"
	"github.com/UtopiaXC/UtopiaDaily/internal/model"
	"github.com/UtopiaXC/UtopiaDaily/internal/repository"
	"github.com/UtopiaXC/UtopiaDaily/internal/security"
	"github.com/UtopiaXC/UtopiaDaily/internal/service"
	"github.com/gin-gonic/gin"
)

type NewspaperController struct {
	newsSvc *service.NewspaperService
	cfgSvc  *service.ConfigService
}

func NewNewspaperController(newsSvc *service.NewspaperService, cfgSvc *service.ConfigService) *NewspaperController {
	return &NewspaperController{newsSvc: newsSvc, cfgSvc: cfgSvc}
}

func (ctl *NewspaperController) Latest(c *gin.Context) {
	if ctl.cfgSvc != nil && ctl.cfgSvc.GetNewspaperVisibility() != "public" {
		user, ok := security.CurrentUser(c)
		if !ok || user == nil || user.Status != model.UserStatusActive {
			common.RespondError(c, http.StatusUnauthorized, common.ErrAuthUnauthorized)
			return
		}
		if !security.HasPermission(c, database.PermNewspaperView) {
			common.RespondError(c, http.StatusForbidden, common.ErrNewspaperForbidden)
			return
		}
	}

	source := c.Query("source")
	resp, err := ctl.newsSvc.GetLatest(source)
	if err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, resp)
}

func (ctl *NewspaperController) ListModuleConfigs(c *gin.Context) {
	items, err := ctl.newsSvc.ListModuleConfigs()
	if err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, items)
}

func (ctl *NewspaperController) SaveModuleConfigs(c *gin.Context) {
	var req struct {
		Modules []repository.NewspaperModuleConfigInput `json:"modules"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, http.StatusBadRequest, common.ErrValidationFailed)
		return
	}
	if err := ctl.newsSvc.SaveModuleConfigs(req.Modules); err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, gin.H{"status": "ok"})
}
