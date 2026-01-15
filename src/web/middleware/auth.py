from starlette.middleware.base import BaseHTTPMiddleware
from starlette.requests import Request
from starlette.responses import JSONResponse
from src.utils.security.jwt_util import decode_access_token
from src.database.connection import system_db_manager
from src.database.models.user import User
from src.utils.logger.logger import Log

TAG = "AUTH_MIDDLEWARE"

class AuthMiddleware(BaseHTTPMiddleware):
    async def dispatch(self, request: Request, call_next):
        path = request.url.path
        if path.startswith("/static") or path in ["/api/auth/login", "/docs", "/openapi.json"]:
            return await call_next(request)

        auth_header = request.headers.get("Authorization")
        user = None
        
        if auth_header and auth_header.startswith("Bearer "):
            token = auth_header.split(" ")[1]
            payload = decode_access_token(token)
            
            if payload:
                user_id = payload.get("sub")
                db = system_db_manager.get_session()
                try:
                    user = db.query(User).filter(User.id == user_id, User.is_active == True).first()
                    if user:
                        request.state.user = user
                        perms = user.role.permissions
                        if isinstance(perms, dict):
                            request.state.permissions = [k for k, v in perms.items() if v]
                        else:
                            request.state.permissions = perms or []
                except Exception as e:
                    Log.e(TAG, "DB Error during auth", error=e)
                finally:
                    db.close()

        if path.startswith("/api/dashboard") and not user:
             public_endpoints = [
                 "/api/dashboard/auth/login",
                 "/api/dashboard/auth/captcha"
             ]
             if path not in public_endpoints:
                 return JSONResponse({"error": "Unauthorized"}, status_code=401)

        response = await call_next(request)
        return response
