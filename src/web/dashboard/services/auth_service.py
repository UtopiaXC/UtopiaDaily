import time
import hashlib
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import or_, select
from starlette.concurrency import run_in_threadpool
from src.database.models import User, UserSession, UserRole
from src.utils.security import crypto_manager
from src.utils.security.jwt_util import create_access_token
from src.web.dashboard.schemas import LoginRequest, LoginResponse, UserResponse
from src.utils.security.captcha import verify_captcha
from src.utils.event import EventManager

class AuthService:
    async def _record_attempt(self, ip_address: str, username: str, success: bool, user_id: str = None):
        level = EventManager.LEVEL_NORMAL if success else EventManager.LEVEL_WARNING
        event_type = "login_success" if success else "login_failed"
        summary = f"User {username} login {'success' if success else 'failed'}"
        
        await EventManager.record_async(
            level=level,
            category=EventManager.CATEGORY_USER,
            event_type=event_type,
            summary=summary,
            details={"username": username, "ip": ip_address},
            source_id=user_id,
            is_resolved=True
        )

    async def login(self, db: AsyncSession, request: LoginRequest, ip_address: str, user_agent: str) -> LoginResponse:
        if not request.captcha_id or not request.captcha_code:
            raise Exception("CAPTCHA_REQUIRED")
        if not verify_captcha(request.captcha_id, request.captcha_code):
            raise Exception("Invalid CAPTCHA")

        stmt = select(User).where(
            or_(
                User.username == request.username,
                User.email == request.username
            )
        )
        result = await db.execute(stmt)
        user = result.scalars().first()

        if not user:
            await self._record_attempt(ip_address, request.username, False)
            return None

        user_id = user.id
        username = user.username
        nickname = user.nickname
        email = user.email
        role_id = user.role_id
        is_active = user.is_active
        password_hash = user.password_hash
        is_password_valid = await run_in_threadpool(crypto_manager.verify_password, request.password, password_hash)

        if not is_password_valid:
            await self._record_attempt(ip_address, request.username, False, user_id)
            return None
            
        if not is_active:
            await self._record_attempt(ip_address, request.username, False, user_id)
            raise Exception("User is disabled")

        await self._record_attempt(ip_address, request.username, True, user_id)

        expire_minutes = 30 * 24 * 60 if request.remember_me else 24 * 60
        access_token = create_access_token(data={"sub": user_id}, expires_delta=expire_minutes * 60)

        expires_at = int(time.time() * 1000) + (expire_minutes * 60 * 1000)
        token_hash = hashlib.sha256(access_token.encode()).hexdigest()
        
        session = UserSession(
            user_id=user_id,
            token_hash=token_hash,
            expires_at=expires_at,
            ip_address=ip_address,
            user_agent=user_agent,
            is_valid=True
        )
        db.add(session)
        await db.commit()

        role_stmt = select(UserRole).where(UserRole.id == role_id)
        role_result = await db.execute(role_stmt)
        role = role_result.scalars().first()
        
        return LoginResponse(
            token=access_token,
            expires_at=expires_at,
            user=UserResponse(
                id=user_id,
                username=username,
                nickname=nickname,
                email=email,
                role_name=role.name if role else "unknown",
                permissions=role.permissions if role else {}
            )
        )
