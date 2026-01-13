from fastapi import APIRouter, Depends, HTTPException, Request
from sqlalchemy.orm import Session
from src.database.connection import system_db_manager
from src.web.dashboard.schemas import LoginRequest, LoginResponse
from src.web.dashboard.services.auth_service import AuthService
from src.utils.logger.logger import Log

router = APIRouter(prefix="/api/dashboard/auth", tags=["Dashboard Authentication"])
auth_service = AuthService()

# Dependency to get DB session
def get_db():
    db = system_db_manager.get_session()
    try:
        yield db
    finally:
        db.close()

@router.post("/login", response_model=LoginResponse)
async def login(request: LoginRequest, req: Request, db: Session = Depends(get_db)):
    try:
        ip = req.client.host
        ua = req.headers.get("user-agent", "unknown")
        
        response = auth_service.login(db, request, ip, ua)
        if not response:
            raise HTTPException(status_code=401, detail="Invalid username or password")
            
        Log.i("DASHBOARD_AUTH", f"User {request.username} logged in from {ip}")
        return response
    except Exception as e:
        Log.e("DASHBOARD_AUTH", "Login error", error=e)
        raise HTTPException(status_code=500, detail=str(e))
