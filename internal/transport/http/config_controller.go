package http

import (
	"net/http"

	"github.com/UtopiaXC/UtopiaDaily/internal/common"
	"github.com/UtopiaXC/UtopiaDaily/internal/service"
	"github.com/gin-gonic/gin"
)

type ConfigController struct {
	cfgSvc *service.ConfigService
}

func NewConfigController(cfgSvc *service.ConfigService) *ConfigController {
	return &ConfigController{cfgSvc: cfgSvc}
}

func (ctl *ConfigController) ListSystemConfigs(c *gin.Context) {
	allowed := map[string]bool{
		"system":         true,
		"security":       true,
		"smtp":           true,
		"classification": true,
	}
	items, err := ctl.cfgSvc.ListSystemConfigs(allowed)
	if err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, items)
}

func (ctl *ConfigController) UpdateSystemConfigsBatch(c *gin.Context) {
	var req struct {
		Configs map[string]any `json:"configs"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Configs == nil {
		common.RespondError(c, http.StatusBadRequest, common.ErrValidationFailed)
		return
	}
	if err := ctl.cfgSvc.UpdateBatchConfigs(req.Configs, false); err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, gin.H{"status": "ok"})
}

func (ctl *ConfigController) UpdateSystemConfig(c *gin.Context) {
	key := c.Param("key")
	var req struct {
		Value string `json:"value"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, http.StatusBadRequest, common.ErrValidationFailed)
		return
	}
	if err := ctl.cfgSvc.UpdateConfig(key, req.Value, false); err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, gin.H{"status": "ok"})
}

func (ctl *ConfigController) TestSMTP(c *gin.Context) {
	var req service.SMTPTestInput
	_ = c.ShouldBindJSON(&req)
	if err := ctl.cfgSvc.TestSMTP(req); err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, gin.H{"status": "ok"})
}

func (ctl *ConfigController) ClassificationSlot(c *gin.Context) {
	common.RespondSuccess(c, ctl.cfgSvc.GetClassificationSlot())
}

func (ctl *ConfigController) ListFrontendConfigs(c *gin.Context) {
	items, err := ctl.cfgSvc.ListSystemConfigs(map[string]bool{"frontend": true})
	if err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, items)
}

func (ctl *ConfigController) UpdateFrontendConfigsBatch(c *gin.Context) {
	var req struct {
		Configs map[string]any `json:"configs"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Configs == nil {
		common.RespondError(c, http.StatusBadRequest, common.ErrValidationFailed)
		return
	}
	if err := ctl.cfgSvc.UpdateBatchConfigs(req.Configs, true); err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, gin.H{"status": "ok"})
}

func (ctl *ConfigController) UpdateFrontendConfig(c *gin.Context) {
	key := c.Param("key")
	var req struct {
		Value string `json:"value"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, http.StatusBadRequest, common.ErrValidationFailed)
		return
	}
	if err := ctl.cfgSvc.UpdateConfig(key, req.Value, true); err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, gin.H{"status": "ok"})
}
