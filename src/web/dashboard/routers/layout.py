from fastapi import APIRouter, Depends, Request
from src.web.dependencies import get_current_user
from src.utils.constants.permissions import Permissions

router = APIRouter(prefix="/api/dashboard/layout", tags=["Dashboard Layout"])

@router.get("/menu")
async def get_menu(request: Request, user = Depends(get_current_user)):
    """
    Returns the list of menu items (tags) visible to the current user.
    """
    user_perms = getattr(request.state, "permissions", [])
    
    menu = []

    if Permissions.SYSTEM_CONFIG_VIEW in user_perms:
        menu.append({"id": "system_config", "label": "System Config", "icon": "settings"})

    if Permissions.USER_MANAGER_VIEW in user_perms:
        menu.append({"id": "user_manager", "label": "User Manager", "icon": "people"})

    # Reordered: Scraper & Push Modules before Schedule
    if Permissions.SCRAPER_VIEW in user_perms:
        menu.append({"id": "scraper_modules", "label": "Scraper Modules", "icon": "extension"})

    if Permissions.PUSH_MODULE_VIEW in user_perms:
        menu.append({"id": "push_module_config", "label": "Push Modules", "icon": "send"})

    if Permissions.SCHEDULE_VIEW in user_perms:
        menu.append({"id": "schedule", "label": "Schedule", "icon": "schedule"})

    if Permissions.FRONTEND_CONFIG_VIEW in user_perms:
        menu.append({"id": "frontend_config", "label": "Frontend Config", "icon": "web"})

    if Permissions.USER_PUSH_SETTINGS in user_perms:
        menu.append({"id": "user_push_config", "label": "My Push Settings", "icon": "notifications_active"})
        
    return {"menu": menu}
