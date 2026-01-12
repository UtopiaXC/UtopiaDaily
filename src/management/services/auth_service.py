import uuid
import time
from sqlalchemy.orm import Session
from src.database.models import User, UserSession, UserRole
from src.utils.security import crypto_manager
from src.management.schemas import LoginRequest, LoginResponse, UserResponse

class AuthService:
    def login(self, db: Session, request: LoginRequest, ip_address: str, user_agent: str) -> LoginResponse:
        # 1. Find User
        user = db.query(User).filter(User.username == request.username).first()
        if not user:
            return None
        
        # 2. Verify Password
        if not crypto_manager.verify_password(request.password, user.password_hash):
            return None
            
        if not user.is_active:
            raise Exception("User is disabled")

        # 3. Create Session Token
        # Simple token generation (in production, use JWT or similar)
        raw_token = str(uuid.uuid4())
        token_hash = crypto_manager.get_password_hash(raw_token) # Hash token before storing
        
        expires_at = int(time.time() * 1000) + (24 * 60 * 60 * 1000) # 24 hours
        
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
        
        # 4. Construct Response
        # Note: We return the raw_token to the user ONLY ONCE here.
        # The DB only stores the hash.
        
        role = db.query(UserRole).filter(UserRole.id == user.role_id).first()
        
        return LoginResponse(
            token=f"{session.id}:{raw_token}", # Format: session_id:token
            expires_at=expires_at,
            user=UserResponse(
                id=user.id,
                username=user.username,
                email=user.email,
                role_name=role.name if role else "unknown",
                permissions=role.permissions if role else {}
            )
        )
