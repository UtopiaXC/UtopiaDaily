from fastapi import APIRouter, Depends, HTTPException, Request
from fastapi.responses import Response
from sqlalchemy.orm import Session
from src.database.connection import system_db_manager
from src.web.dashboard.schemas import LoginRequest, LoginResponse
from src.web.dashboard.services.auth_service import AuthService
from src.utils.security.captcha import generate_captcha
from src.utils.logger.logger import Log
from src.utils.i18n import i18n

router = APIRouter(prefix="/api/dashboard/auth", tags=["Dashboard Authentication"])
auth_service = AuthService()

def get_db():
    db = system_db_manager.get_session()
    try:
        yield db
    finally:
        db.close()

def get_locale(req: Request) -> str:
    accept_language = req.headers.get("accept-language")
    if accept_language:
        lang = accept_language.split(",")[0].strip().replace("-", "_")
        if lang in i18n.translations:
            return lang
    return i18n.default_locale

@router.get("/captcha")
async def get_captcha():
    """
    Returns a new CAPTCHA (SVG image and ID).
    """
    captcha_id, svg = generate_captcha()
    return {"captcha_id": captcha_id, "svg": svg}

@router.post("/login", response_model=LoginResponse)
async def login(request: LoginRequest, req: Request, db: Session = Depends(get_db)):
    locale = get_locale(req)
    try:
        ip = req.client.host
        ua = req.headers.get("user-agent", "unknown")
        response = auth_service.login(db, request, ip, ua)
        if not response:
            # Use i18n for error message
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
             # Assuming we add this key later, or fallback
             detail = "User is disabled" 
             raise HTTPException(status_code=401, detail=detail)
            
        Log.e("DASHBOARD_AUTH", "Login error", error=e)
        detail = i18n.t("login.failed", locale=locale)
        raise HTTPException(status_code=500, detail=detail)
