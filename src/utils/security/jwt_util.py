import jwt
import time
from typing import Optional, Dict
from src.utils.env_manage.env_manager import EnvManager
from src.utils.event import EventManager
from src.utils.logger.logger import Log

TAG = "JWT_UTIL"

ALGORITHM = "HS256"
ACCESS_TOKEN_EXPIRE_MINUTES = 60 * 24 # 1 day

def get_secret_key():
    key = EnvManager.get_env("JWT_SECRET", "")
    if not key:
        EventManager.record(
            level=EventManager.LEVEL_FATAL,
            category=EventManager.CATEGORY_SYSTEM,
            event_type="missing_config",
            summary=f"No jwt secret found!",
            details={
                "msg": "System exit unexpectedly due to missing jwt secret in config/.env",
            },
            source_id="JWT_SECRET",
            is_resolved=False
        )
        Log.fatal(TAG,f"NO JWT SECRET CONFIGURATION! MUST SET A VARIABLE SECRET IN config/.env")
    return key

def create_access_token(data: dict, expires_delta: int = None) -> str:
    to_encode = data.copy()
    if expires_delta:
        expire = time.time() + expires_delta
    else:
        expire = time.time() + (ACCESS_TOKEN_EXPIRE_MINUTES * 60)
    
    to_encode.update({"exp": expire})
    encoded_jwt = jwt.encode(to_encode, get_secret_key(), algorithm=ALGORITHM)
    return encoded_jwt

def decode_access_token(token: str) -> Optional[Dict]:
    try:
        payload = jwt.decode(token, get_secret_key(), algorithms=[ALGORITHM])
        if payload.get("exp") < time.time():
            return None
        return payload
    except jwt.PyJWTError:
        return None
