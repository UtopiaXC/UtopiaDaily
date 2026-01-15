from starlette.middleware.base import BaseHTTPMiddleware
from starlette.requests import Request
from starlette.responses import JSONResponse
from src.utils.logger.logger import Log

TAG = "SECURITY_MIDDLEWARE"

class SecurityMiddleware(BaseHTTPMiddleware):
    async def dispatch(self, request: Request, call_next):
        # Basic Header Check
        host = request.headers.get("host", "")
        
        # 2. SQL Injection & XSS Basic Filter
        query_string = request.url.query.lower()
        suspicious_patterns = ["union select", "drop table", "<script>"]
        for pattern in suspicious_patterns:
            if pattern in query_string:
                Log.w(TAG, f"Blocked suspicious request from {request.client.host}: {query_string}")
                return JSONResponse({"error": "Illegal request detected"}, status_code=403)

        # Security Headers
        response = await call_next(request)
        response.headers["X-Content-Type-Options"] = "nosniff"
        response.headers["X-Frame-Options"] = "DENY"
        response.headers["X-XSS-Protection"] = "1; mode=block"
        
        return response
