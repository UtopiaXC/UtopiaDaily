from pydantic import BaseModel
from typing import Optional, List, Dict, Any

class LoginRequest(BaseModel):
    username: str
    password: str

class UserResponse(BaseModel):
    id: str
    username: str
    email: Optional[str]
    role_name: str
    permissions: Dict[str, Any]

class LoginResponse(BaseModel):
    token: str
    expires_at: int
    user: UserResponse
