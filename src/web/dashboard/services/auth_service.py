import time
import hashlib
from sqlalchemy.orm import Session
from src.database.models import User, UserSession, UserRole, UserLog
from src.utils.security import crypto_manager
from src.utils.security.jwt_util import create_access_token
from src.web.dashboard.schemas import LoginRequest, LoginResponse, UserResponse
from src.utils.security.captcha import verify_captcha

class AuthService:
    def _record_attempt(self, db: Session, ip_address: str, username: str, success: bool, user_id: str = None):
        log = UserLog(
            user_id=user_id,
            action="LOGIN_SUCCESS" if success else "LOGIN_FAILED",
            status="SUCCESS" if success else "FAILURE",
            ip_address=ip_address,
            details={"username": username}
        )
        db.add(log)
        db.commit()

    def login(self, db: Session, request: LoginRequest, ip_address: str, user_agent: str) -> LoginResponse:
        # 1. Mandatory Captcha Check
        if not request.captcha_id or not request.captcha_code:
            raise Exception("CAPTCHA_REQUIRED") 
        
        if not verify_captcha(request.captcha_id, request.captcha_code):
            raise Exception("Invalid CAPTCHA")

        # 2. Find User
        user = db.query(User).filter(User.username == request.username).first()
        if not user:
            self._record_attempt(db, ip_address, request.username, False)
            return None
        
        # 3. Verify Password
        if not crypto_manager.verify_password(request.password, user.password_hash):
            self._record_attempt(db, ip_address, request.username, False, user.id)
            return None
            
        if not user.is_active:
            self._record_attempt(db, ip_address, request.username, False, user.id)
            raise Exception("User is disabled")

        # Login Success
        self._record_attempt(db, ip_address, request.username, True, user.id)

        # 4. Create JWT Token
        expire_minutes = 30 * 24 * 60 if request.remember_me else 24 * 60
        access_token = create_access_token(data={"sub": user.id}, expires_delta=expire_minutes * 60)
        
        # 5. Store Session Record
        expires_at = int(time.time() * 1000) + (expire_minutes * 60 * 1000)
        token_hash = hashlib.sha256(access_token.encode()).hexdigest()
        
        session = UserSession(
            user_id=user.id,
            token_hash=token_hash,
            expires_at=expires_at,
            ip_address=ip_address,
            user_agent=user_agent,
            is_valid=True
        )
        db.add(session)
        db.commit()
        
        # 6. Construct Response
        role = db.query(UserRole).filter(UserRole.id == user.role_id).first()
        
        return LoginResponse(
            token=access_token,
            expires_at=expires_at,
            user=UserResponse(
                id=user.id,
                username=user.username,
                email=user.email,
                role_name=role.name if role else "unknown",
                permissions=role.permissions if role else {}
            )
        )
