package http

import (
	"net/http"
	"strconv"

	"github.com/UtopiaXC/UtopiaDaily/internal/common"
	"github.com/UtopiaXC/UtopiaDaily/internal/repository"
	"github.com/UtopiaXC/UtopiaDaily/internal/service"
	"github.com/gin-gonic/gin"
)

type ClassificationController struct {
	clsSvc *service.ClassificationService
}

func NewClassificationController(clsSvc *service.ClassificationService) *ClassificationController {
	return &ClassificationController{clsSvc: clsSvc}
}

func (ctl *ClassificationController) Overview(c *gin.Context) {
	view, err := ctl.clsSvc.GetOverview()
	if err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, view)
}

func (ctl *ClassificationController) PrepareONNXEnv(c *gin.Context) {
	job := ctl.clsSvc.StartPrepareONNXEnvJob()
	common.RespondSuccess(c, gin.H{
		"status": "accepted",
		"job":    job,
	})
}

func (ctl *ClassificationController) DownloadONNXModel(c *gin.Context) {
	var req struct {
		RepoID string `json:"repo_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.RepoID == "" {
		common.RespondError(c, http.StatusBadRequest, common.ErrValidationFailed)
		return
	}
	job, err := ctl.clsSvc.StartDownloadONNXModelJob(req.RepoID)
	if err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, gin.H{
		"status": "accepted",
		"job":    job,
	})
}

func (ctl *ClassificationController) DeleteONNXModel(c *gin.Context) {
	if err := ctl.clsSvc.DeleteONNXModel(c.Param("name")); err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, gin.H{"status": "ok"})
}

func (ctl *ClassificationController) ListOllamaModels(c *gin.Context) {
	models, err := ctl.clsSvc.ListOllamaModels(c.Request.Context(), c.Query("url"))
	if err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, models)
}

func (ctl *ClassificationController) PullOllamaModel(c *gin.Context) {
	var req struct {
		URL   string `json:"url"`
		Model string `json:"model"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Model == "" {
		common.RespondError(c, http.StatusBadRequest, common.ErrValidationFailed)
		return
	}
	job, err := ctl.clsSvc.StartPullOllamaModelJob(req.URL, req.Model)
	if err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, gin.H{
		"status": "accepted",
		"job":    job,
	})
}

func (ctl *ClassificationController) ListJobs(c *gin.Context) {
	common.RespondSuccess(c, ctl.clsSvc.ListJobs())
}

func (ctl *ClassificationController) GetJob(c *gin.Context) {
	job, err := ctl.clsSvc.GetJob(c.Param("id"))
	if err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, job)
}

func (ctl *ClassificationController) CancelJob(c *gin.Context) {
	job, err := ctl.clsSvc.CancelJob(c.Param("id"))
	if err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, job)
}

func (ctl *ClassificationController) DismissJob(c *gin.Context) {
	if err := ctl.clsSvc.DismissJob(c.Param("id")); err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, gin.H{"status": "ok"})
}

func (ctl *ClassificationController) ListProfiles(c *gin.Context) {
	profiles, err := ctl.clsSvc.ListProfiles()
	if err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, profiles)
}

func (ctl *ClassificationController) CreateProfile(c *gin.Context) {
	var req repository.ProfileSaveInput
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, http.StatusBadRequest, common.ErrValidationFailed)
		return
	}
	p, err := ctl.clsSvc.CreateProfile(req)
	if err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, p)
}

func (ctl *ClassificationController) UpdateProfile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		common.RespondError(c, http.StatusBadRequest, common.ErrValidationFailed)
		return
	}
	var req repository.ProfileSaveInput
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, http.StatusBadRequest, common.ErrValidationFailed)
		return
	}
	p, err := ctl.clsSvc.UpdateProfile(uint(id), req)
	if err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, p)
}

func (ctl *ClassificationController) DeleteProfile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		common.RespondError(c, http.StatusBadRequest, common.ErrValidationFailed)
		return
	}
	if err := ctl.clsSvc.DeleteProfile(uint(id)); err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, gin.H{"status": "ok"})
}

func (ctl *ClassificationController) LoadDefaultProfile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		common.RespondError(c, http.StatusBadRequest, common.ErrValidationFailed)
		return
	}
	var req struct {
		Locale string `json:"locale"`
	}
	_ = c.ShouldBindJSON(&req)
	p, err := ctl.clsSvc.LoadDefaultProfileTemplate(uint(id), req.Locale)
	if err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, p)
}

func (ctl *ClassificationController) GetModuleConfig(c *gin.Context) {
	cfg, err := ctl.clsSvc.GetModuleConfig(c.Param("id"))
	if err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, cfg)
}

func (ctl *ClassificationController) SaveModuleConfig(c *gin.Context) {
	var req service.ModuleClassificationConfigView
	if err := c.ShouldBindJSON(&req); err != nil {
		common.RespondError(c, http.StatusBadRequest, common.ErrValidationFailed)
		return
	}
	if err := ctl.clsSvc.SaveModuleConfig(c.Param("id"), req); err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, gin.H{"status": "ok"})
}

func (ctl *ClassificationController) QueueStatus(c *gin.Context) {
	status, err := ctl.clsSvc.GetQueueStatus()
	if err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, status)
}

func (ctl *ClassificationController) ListQueueTasks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.DefaultQuery("status", "all")
	items, total, err := ctl.clsSvc.ListQueueTasks(status, page, pageSize)
	if err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, gin.H{
		"items":     items,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (ctl *ClassificationController) GetQueueTaskDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		common.RespondError(c, http.StatusBadRequest, common.ErrValidationFailed)
		return
	}
	detail, err := ctl.clsSvc.GetQueueTaskDetail(uint(id))
	if err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, detail)
}

func (ctl *ClassificationController) PauseQueue(c *gin.Context) {
	ctl.clsSvc.PauseWorker()
	common.RespondSuccess(c, gin.H{"status": "paused"})
}

func (ctl *ClassificationController) ResumeQueue(c *gin.Context) {
	ctl.clsSvc.ResumeWorker()
	common.RespondSuccess(c, gin.H{"status": "running"})
}

func (ctl *ClassificationController) EnqueueUntagged(c *gin.Context) {
	var req struct {
		Limit int `json:"limit"`
	}
	_ = c.ShouldBindJSON(&req)
	added, err := ctl.clsSvc.EnqueueUntagged(req.Limit)
	if err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, gin.H{"added": added})
}

func (ctl *ClassificationController) CancelPending(c *gin.Context) {
	cnt, err := ctl.clsSvc.CancelPendingTasks()
	if err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, gin.H{"cancelled": cnt})
}

func (ctl *ClassificationController) RetryFailed(c *gin.Context) {
	cnt, err := ctl.clsSvc.RetryFailedTasks()
	if err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, gin.H{"retried": cnt})
}

func (ctl *ClassificationController) ClearFinished(c *gin.Context) {
	cnt, err := ctl.clsSvc.ClearFinishedTasks()
	if err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, gin.H{"cleared": cnt})
}

func (ctl *ClassificationController) CancelTask(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		common.RespondError(c, http.StatusBadRequest, common.ErrValidationFailed)
		return
	}
	if err := ctl.clsSvc.CancelTask(uint(id)); err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, gin.H{"status": "ok"})
}

func (ctl *ClassificationController) RetryTask(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		common.RespondError(c, http.StatusBadRequest, common.ErrValidationFailed)
		return
	}
	if err := ctl.clsSvc.RetryTask(uint(id)); err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, gin.H{"status": "ok"})
}
