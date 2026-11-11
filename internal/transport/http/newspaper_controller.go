package http

import (
	"github.com/UtopiaXC/UtopiaDaily/internal/common"
	"github.com/UtopiaXC/UtopiaDaily/internal/service"
	"github.com/gin-gonic/gin"
)

type NewspaperController struct {
	newsSvc *service.NewspaperService
}

func NewNewspaperController(newsSvc *service.NewspaperService) *NewspaperController {
	return &NewspaperController{newsSvc: newsSvc}
}

func (ctl *NewspaperController) Latest(c *gin.Context) {
	source := c.Query("source")
	resp, err := ctl.newsSvc.GetLatest(source)
	if err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, resp)
}
