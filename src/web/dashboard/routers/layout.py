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
    
    # 1. System Config
    if Permissions.SYSTEM_CONFIG_VIEW in user_perms:
        menu.append({"id": "system_config", "label": "System Config", "icon": "settings"})
        
    # 2. User Management
    if Permissions.USER_MANAGER_VIEW in user_perms:
        menu.append({"id": "user_manager", "label": "User Manager", "icon": "people"})
        
    # 3. Schedule
    if Permissions.SCHEDULE_VIEW in user_perms:
        menu.append({"id": "schedule", "label": "Schedule", "icon": "schedule"})
        
    # 4. Scraper Modules
    if Permissions.SCRAPER_VIEW in user_perms:
        menu.append({"id": "scraper_config", "label": "Scraper Modules", "icon": "extension"})
        
    # 5. Push Modules
    if Permissions.PUSH_MODULE_VIEW in user_perms:
        menu.append({"id": "push_module_config", "label": "Push Modules", "icon": "send"})
        
    # 6. Frontend Config
    if Permissions.FRONTEND_CONFIG_VIEW in user_perms:
        menu.append({"id": "frontend_config", "label": "Frontend Config", "icon": "web"})
        
    # 7. User Push Config (Personal)
    if Permissions.USER_PUSH_SETTINGS in user_perms:
        menu.append({"id": "user_push_config", "label": "My Push Settings", "icon": "notifications_active"})
        
    return {"menu": menu}
