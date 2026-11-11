package http

import (
	"strconv"

	"github.com/UtopiaXC/UtopiaDaily/internal/common"
	"github.com/UtopiaXC/UtopiaDaily/internal/config"
	"github.com/UtopiaXC/UtopiaDaily/internal/database"
	"github.com/UtopiaXC/UtopiaDaily/internal/repository"
	"github.com/UtopiaXC/UtopiaDaily/internal/security"
	"github.com/UtopiaXC/UtopiaDaily/internal/service"
	"github.com/gin-gonic/gin"
)

type MenuItem struct {
	ID    string `json:"id"`
	Icon  string `json:"icon"`
	Label string `json:"label"`
}

type DashboardController struct {
	modRepo     *repository.ModuleRepository
	scraperRepo *repository.ScraperRepository
	eventRepo   *repository.EventRepository
	cfgSvc      *service.ConfigService
}

func NewDashboardController(
	modRepo *repository.ModuleRepository,
	scraperRepo *repository.ScraperRepository,
	eventRepo *repository.EventRepository,
	cfgSvc *service.ConfigService,
) *DashboardController {
	return &DashboardController{
		modRepo:     modRepo,
		scraperRepo: scraperRepo,
		eventRepo:   eventRepo,
		cfgSvc:      cfgSvc,
	}
}

func (ctl *DashboardController) LayoutMenu(c *gin.Context) {
	candidates := []struct {
		id   string
		icon string
		perm string
	}{
		{id: "overview", icon: "dashboard", perm: database.PermSystemStatusView},
		{id: "system_config", icon: "settings", perm: database.PermSystemConfigView},
		{id: "user_manager", icon: "people", perm: database.PermUserManagerView},
		{id: "scraper_modules", icon: "cloud_download", perm: database.PermScraperView},
		{id: "classification", icon: "auto_awesome", perm: database.PermClassificationView},
		{id: "data_manager", icon: "inventory_2", perm: database.PermDataView},
		{id: "push_modules", icon: "send", perm: database.PermPushModuleView},
		{id: "schedule", icon: "schedule", perm: database.PermScheduleView},
		{id: "frontend_config", icon: "newspaper", perm: database.PermFrontendConfigView},
		{id: "user_push", icon: "mark_email_unread", perm: database.PermUserPushSettings},
	}

	menu := make([]MenuItem, 0, len(candidates))
	for _, item := range candidates {
		if security.HasPermission(c, item.perm) {
			menu = append(menu, MenuItem{
				ID:    item.id,
				Icon:  item.icon,
				Label: "menu." + item.id,
			})
		}
	}
	common.RespondSuccess(c, gin.H{"menu": menu})
}

func (ctl *DashboardController) Overview(c *gin.Context) {
	activeMods, _ := ctl.modRepo.CountEnabled()
	articles, _ := ctl.scraperRepo.CountResults()
	clsSlot := ctl.cfgSvc.GetClassificationSlot()

	common.RespondSuccess(c, gin.H{
		"api_status":          "online",
		"db_status":           "connected",
		"version":             config.Version,
		"active_modules":      activeMods,
		"scraped_articles":    articles,
		"classification_slot": clsSlot,
	})
}

func (ctl *DashboardController) Events(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "15"))

	items, total, err := ctl.eventRepo.Query(repository.EventQueryParams{
		Page:     page,
		PageSize: pageSize,
		Level:    c.Query("level"),
		Category: c.Query("category"),
		Search:   c.Query("search"),
		Date:     c.Query("date"),
		SortBy:   c.DefaultQuery("sort_by", "created_at"),
		Order:    c.DefaultQuery("order", "desc"),
	})
	if err != nil {
		common.RespondError(c, 500, common.ErrDatabaseError)
		return
	}
	cats, _ := ctl.eventRepo.DistinctCategories()
	common.RespondSuccess(c, gin.H{
		"items":      items,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
		"categories": cats,
	})
}
