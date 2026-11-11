package http

import (
	"net/http"
	"strconv"

	"github.com/UtopiaXC/UtopiaDaily/internal/common"
	"github.com/UtopiaXC/UtopiaDaily/internal/service"
	"github.com/gin-gonic/gin"
)

type ScheduleController struct {
	schedSvc *service.ScheduleService
}

func NewScheduleController(schedSvc *service.ScheduleService) *ScheduleController {
	return &ScheduleController{schedSvc: schedSvc}
}

func (ctl *ScheduleController) List(c *gin.Context) {
	jobs, err := ctl.schedSvc.ListJobs()
	if err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, jobs)
}

func (ctl *ScheduleController) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		common.RespondError(c, http.StatusBadRequest, common.ErrValidationFailed)
		return
	}
	var req struct {
		CronExpr string `json:"cron_expr"`
		Enabled  bool   `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, http.StatusBadRequest, common.ErrValidationFailed)
		return
	}
	if err := ctl.schedSvc.UpdateJob(uint(id), req.CronExpr, req.Enabled); err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, gin.H{"status": "ok"})
}

func (ctl *ScheduleController) BatchUpdate(c *gin.Context) {
	var req struct {
		Jobs []service.ScheduleBatchUpdateItem `json:"jobs"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, http.StatusBadRequest, common.ErrValidationFailed)
		return
	}
	if err := ctl.schedSvc.BatchUpdateJobs(req.Jobs); err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, gin.H{"status": "ok"})
}

func (ctl *ScheduleController) RunNow(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		common.RespondError(c, http.StatusBadRequest, common.ErrValidationFailed)
		return
	}
	if err := ctl.schedSvc.TriggerJob(c.Request.Context(), uint(id)); err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, gin.H{"status": "ok"})
}
