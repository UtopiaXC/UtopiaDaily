import re
from starlette.middleware.base import BaseHTTPMiddleware
from starlette.requests import Request
from starlette.responses import JSONResponse
from src.utils.constants.permissions import Permissions
from src.utils.logger.logger import Log

TAG = "PERMISSION_MIDDLEWARE"

class PermissionRule:
    def __init__(self, path_regex: str, methods: list, permission: str):
        self.regex = re.compile(path_regex)
        self.methods = methods
        self.permission = permission

RULES = [
    PermissionRule(r"^/api/dashboard/auth/.*", ["GET", "POST"], "ALLOW"),
    PermissionRule(r"^/api/dashboard/layout/menu", ["GET"], "ALLOW"),

    PermissionRule(r"^/api/dashboard/system-config.*", ["GET"], Permissions.SYSTEM_CONFIG_VIEW),
    PermissionRule(r"^/api/dashboard/system-config.*", ["PUT"], Permissions.SYSTEM_CONFIG_EDIT),

    PermissionRule(r"^/api/dashboard/user-manager/permissions", ["GET"], Permissions.USER_MANAGER_VIEW),
    PermissionRule(r"^/api/dashboard/user-manager.*", ["GET"], Permissions.USER_MANAGER_VIEW),
    PermissionRule(r"^/api/dashboard/user-manager.*", ["POST", "PUT", "DELETE"], Permissions.USER_MANAGER_EDIT),

    PermissionRule(r"^/api/dashboard/scraper/modules.*", ["GET"], Permissions.SCRAPER_VIEW),
    PermissionRule(r"^/api/dashboard/scraper/modules.*", ["POST"], Permissions.SCRAPER_EDIT),

    PermissionRule(r"^/api/dashboard/schedule.*", ["GET"], Permissions.SCHEDULE_VIEW),
    PermissionRule(r"^/api/dashboard/schedule.*", ["POST", "PUT", "DELETE"], Permissions.SCHEDULE_EDIT),

    PermissionRule(r"^/api/dashboard/push/modules.*", ["GET"], Permissions.PUSH_MODULE_VIEW),
    PermissionRule(r"^/api/dashboard/push/modules.*", ["POST", "PUT", "DELETE"], Permissions.PUSH_MODULE_EDIT),

    PermissionRule(r"^/api/dashboard/frontend-config.*", ["GET"], Permissions.FRONTEND_CONFIG_VIEW),
    PermissionRule(r"^/api/dashboard/frontend-config.*", ["POST", "PUT"], Permissions.FRONTEND_CONFIG_EDIT),

    PermissionRule(r"^/api/dashboard/user-push.*", ["GET", "POST", "PUT"], Permissions.USER_PUSH_SETTINGS),
]

class PermissionMiddleware(BaseHTTPMiddleware):
    async def dispatch(self, request: Request, call_next):
        path = request.url.path
        method = request.method
        if not path.startswith("/api/dashboard"):
            return await call_next(request)
        matched_rule = None
        for rule in RULES:
            if rule.regex.match(path) and (method in rule.methods or "*" in rule.methods):
                matched_rule = rule
                break
        
        if matched_rule:
            if matched_rule.permission == "ALLOW":
                return await call_next(request)
            user = getattr(request.state, "user", None)
            if not user:
                return JSONResponse({"error": "Unauthorized"}, status_code=401)
            user_perms = getattr(request.state, "permissions", [])
            if matched_rule.permission not in user_perms:
                Log.w(TAG, f"Access denied for user {user.username} to {method} {path}. Missing: {matched_rule.permission}")
                return JSONResponse({"error": "Forbidden", "detail": f"Missing permission: {matched_rule.permission}"}, status_code=403)
        
        else:
            Log.w(TAG, f"No permission rule matched for {method} {path}. Defaulting to Deny.")
            return JSONResponse({"error": "Forbidden", "detail": "Access denied (No rule matched)"}, status_code=403)

        return await call_next(request)
