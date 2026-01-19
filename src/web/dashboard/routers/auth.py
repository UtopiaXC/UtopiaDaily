from fastapi import APIRouter, Depends, HTTPException, Request
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select
from src.database.models import User, UserRole
from src.web.dashboard.schemas import LoginRequest, LoginResponse, UserResponse
from src.web.dashboard.services.auth_service import AuthService
from src.utils.security.captcha import generate_captcha
from src.utils.logger.logger import Log
from src.utils.i18n import i18n
from src.web.dependencies import get_async_db, get_current_user

router = APIRouter(prefix="/api/dashboard/auth", tags=["Dashboard Authentication"])
auth_service = AuthService()

def get_locale(req: Request) -> str:
    accept_language = req.headers.get("accept-language")
    if accept_language:
        lang = accept_language.split(",")[0].strip().replace("-", "_")
        if lang in i18n.translations:
            return lang
    return i18n.default_locale

@router.get("/captcha")
def get_captcha():
    captcha_id, svg = generate_captcha()
    return {"captcha_id": captcha_id, "svg": svg}

@router.post("/login", response_model=LoginResponse)
async def login(request: LoginRequest, req: Request, db: AsyncSession = Depends(get_async_db)):
    locale = get_locale(req)
    try:
        ip = req.client.host
        ua = req.headers.get("user-agent", "unknown")
        response = await auth_service.login(db, request, ip, ua)
        if not response:
            detail = i18n.t("login.invalid_credentials", locale=locale)
            raise HTTPException(status_code=401, detail=detail)
        Log.i("DASHBOARD_AUTH", f"User {request.username} logged in from {ip}")
        return response
    except HTTPException:
        raise
    except Exception as e:
        error_msg = str(e)
        if error_msg == "CAPTCHA_REQUIRED":
            detail = i18n.t("captcha.required", locale=locale)
            raise HTTPException(status_code=403, detail=detail)
        if error_msg == "Invalid CAPTCHA":
            detail = i18n.t("captcha.invalid", locale=locale)
            raise HTTPException(status_code=400, detail=detail)
        if error_msg == "User is disabled":
             detail = "User is disabled" 
             raise HTTPException(status_code=401, detail=detail)
            
        Log.e("DASHBOARD_AUTH", "Login error", error=e)
        detail = i18n.t("login.failed", locale=locale)
        raise HTTPException(status_code=500, detail=detail)

@router.get("/me", response_model=UserResponse)
async def get_current_user_info(user: User = Depends(get_current_user), db: AsyncSession = Depends(get_async_db)):
    result = await db.execute(select(User).where(User.id == user.id))
    current_user = result.scalars().first()
    
    if not current_user:
        raise HTTPException(status_code=404, detail="User not found")

    role_result = await db.execute(select(UserRole).where(UserRole.id == current_user.role_id))
    role = role_result.scalars().first()

    return UserResponse(
        id=current_user.id,
        username=current_user.username,
        nickname=current_user.nickname,
        email=current_user.email,
        role_name=role.name if role else "unknown",
        permissions=role.permissions if role else {}
    )
