from starlette.middleware.base import BaseHTTPMiddleware
from starlette.requests import Request
from starlette.responses import JSONResponse
from sqlalchemy import select
from sqlalchemy.orm import selectinload
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
                
                # 使用原生异步会话
                async with system_db_manager.get_async_session() as session:
                    try:
                        # 使用 selectinload 预加载 role 关系，避免 N+1 问题和异步 lazy load 错误
                        stmt = select(User).options(selectinload(User.role)).where(User.id == user_id, User.is_active == True)
                        result = await session.execute(stmt)
                        user = result.scalars().first()
                        
                        if user:
                            # 此时 user.role 已经加载，可以直接访问
                            raw_perms = user.role.permissions
                            if isinstance(raw_perms, dict):
                                permissions = [k for k, v in raw_perms.items() if v]
                            else:
                                permissions = raw_perms or []
                                
                            request.state.user = user
                            request.state.permissions = permissions
                    except Exception as e:
                        Log.e(TAG, "Async DB Error during auth", error=e)

        if path.startswith("/api/dashboard") and not user:
             public_endpoints = [
                 "/api/dashboard/auth/login",
                 "/api/dashboard/auth/captcha"
             ]
             if path not in public_endpoints:
                 return JSONResponse({"error": "Unauthorized"}, status_code=401)

        response = await call_next(request)
        return response
