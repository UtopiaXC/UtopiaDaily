package http

import (
	"net/http"

	"github.com/UtopiaXC/UtopiaDaily/internal/common"
	"github.com/UtopiaXC/UtopiaDaily/internal/pkg/i18n"
	"github.com/UtopiaXC/UtopiaDaily/internal/service"
	"github.com/gin-gonic/gin"
)

type CommonController struct {
	cfgSvc  *service.ConfigService
	i18nMgr *i18n.Assembler
}

func NewCommonController(cfgSvc *service.ConfigService, i18nMgr *i18n.Assembler) *CommonController {
	return &CommonController{
		cfgSvc:  cfgSvc,
		i18nMgr: i18nMgr,
	}
}

func (ctl *CommonController) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (ctl *CommonController) SystemInfo(c *gin.Context) {
	common.RespondSuccess(c, ctl.cfgSvc.GetPublicSystemInfo())
}

func (ctl *CommonController) SiteInfo(c *gin.Context) {
	common.RespondSuccess(c, ctl.cfgSvc.GetSiteInfo())
}

func (ctl *CommonController) Locales(c *gin.Context) {
	common.RespondSuccess(c, gin.H{
		"languages":      ctl.i18nMgr.Languages(),
		"default_locale": ctl.cfgSvc.GetDefaultLocale(),
	})
}

func (ctl *CommonController) Bundle(c *gin.Context) {
	locale := c.Param("locale")
	common.RespondSuccess(c, ctl.i18nMgr.Bundle(locale))
}
