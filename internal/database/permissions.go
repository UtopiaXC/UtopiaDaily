package database

import "strings"

const (
	PermSystemStatusView   = "system.status.view"
	PermSystemConfigView   = "system.config.view"
	PermSystemConfigEdit   = "system.config.edit"
	PermUserManagerView    = "user_manager.view"
	PermUserManagerEdit    = "user_manager.edit"
	PermScheduleView       = "schedule.view"
	PermScheduleEdit       = "schedule.edit"
	PermScraperView        = "scraper.view"
	PermScraperEdit        = "scraper.edit"
	PermPushModuleView     = "push_module.view"
	PermPushModuleEdit     = "push_module.edit"
	PermFrontendConfigView = "frontend.config.view"
	PermFrontendConfigEdit = "frontend.config.edit"
	PermUserPushSettings   = "user_push.settings"
)

func AllPermissions() []string {
	return []string{
		PermSystemStatusView,
		PermSystemConfigView,
		PermSystemConfigEdit,
		PermUserManagerView,
		PermUserManagerEdit,
		PermScheduleView,
		PermScheduleEdit,
		PermScraperView,
		PermScraperEdit,
		PermPushModuleView,
		PermPushModuleEdit,
		PermFrontendConfigView,
		PermFrontendConfigEdit,
		PermUserPushSettings,
	}
}

func DefaultUserPermissions() []string {
	return []string{
		PermUserPushSettings,
	}
}

func PermissionGroup(code string) string {
	if idx := strings.IndexByte(code, '.'); idx > 0 {
		return code[:idx]
	}
	return "system"
}
