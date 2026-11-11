package http

import (
	"net/http"
	"strconv"

	"github.com/UtopiaXC/UtopiaDaily/internal/common"
	"github.com/UtopiaXC/UtopiaDaily/internal/repository"
	"github.com/UtopiaXC/UtopiaDaily/internal/service"
	"github.com/gin-gonic/gin"
)

type DataController struct {
	dataSvc *service.DataService
	clsSvc  *service.ClassificationService
}

func NewDataController(dataSvc *service.DataService, clsSvc *service.ClassificationService) *DataController {
	return &DataController{dataSvc: dataSvc, clsSvc: clsSvc}
}

func (ctl *DataController) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	tagID64, _ := strconv.ParseUint(c.Query("tag_id"), 10, 64)
	untagged := c.Query("untagged") == "true" || c.Query("untagged") == "1"

	items, total, mods, err := ctl.dataSvc.ListArticles(repository.DataQueryParams{
		Page:         page,
		PageSize:     pageSize,
		Search:       c.Query("search"),
		ModuleID:     c.Query("module_id"),
		TagID:        uint(tagID64),
		UntaggedOnly: untagged,
		DateFrom:     c.Query("date_from"),
		DateTo:       c.Query("date_to"),
	})
	if err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, gin.H{
		"items":     items,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"modules":   mods,
	})
}

func (ctl *DataController) ListTags(c *gin.Context) {
	tags, err := ctl.clsSvc.ListAllTags()
	if err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, tags)
}

func (ctl *DataController) Detail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		common.RespondError(c, http.StatusBadRequest, common.ErrValidationFailed)
		return
	}
	dto, err := ctl.dataSvc.GetArticleDetail(uint(id))
	if err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, dto)
}

func (ctl *DataController) DeleteOne(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		common.RespondError(c, http.StatusBadRequest, common.ErrValidationFailed)
		return
	}
	cnt, err := ctl.dataSvc.DeleteArticles([]uint{uint(id)})
	if err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, gin.H{"deleted": cnt})
}

func (ctl *DataController) DeleteBatch(c *gin.Context) {
	var req struct {
		IDs []uint `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || len(req.IDs) == 0 {
		common.RespondError(c, http.StatusBadRequest, common.ErrValidationFailed)
		return
	}
	cnt, err := ctl.dataSvc.DeleteArticles(req.IDs)
	if err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, gin.H{"deleted": cnt})
}

func (ctl *DataController) AddTag(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		common.RespondError(c, http.StatusBadRequest, common.ErrValidationFailed)
		return
	}
	var req struct {
		TagID uint `json:"tag_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.TagID == 0 {
		common.RespondError(c, http.StatusBadRequest, common.ErrValidationFailed)
		return
	}
	if err := ctl.dataSvc.AddTag(uint(id), req.TagID); err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, gin.H{"status": "ok"})
}

func (ctl *DataController) RemoveTag(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		common.RespondError(c, http.StatusBadRequest, common.ErrValidationFailed)
		return
	}
	tagID, err := strconv.ParseUint(c.Param("tag_id"), 10, 64)
	if err != nil {
		common.RespondError(c, http.StatusBadRequest, common.ErrValidationFailed)
		return
	}
	if err := ctl.dataSvc.RemoveTag(uint(id), uint(tagID)); err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, gin.H{"status": "ok"})
}

func (ctl *DataController) Reclassify(c *gin.Context) {
	var req struct {
		IDs []uint `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || len(req.IDs) == 0 {
		common.RespondError(c, http.StatusBadRequest, common.ErrValidationFailed)
		return
	}
	added, err := ctl.dataSvc.EnqueueReclassify(req.IDs)
	if err != nil {
		common.RespondFromError(c, err)
		return
	}
	common.RespondSuccess(c, gin.H{"added": added})
}
