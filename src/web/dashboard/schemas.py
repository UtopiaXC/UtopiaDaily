from pydantic import BaseModel
from typing import Optional, List, Dict, Any, Union

class LoginRequest(BaseModel):
    username: str
    password: str
    remember_me: bool = False
    captcha_code: Optional[str] = None
    captcha_id: Optional[str] = None

class UserResponse(BaseModel):
    id: str
    username: str
    email: Optional[str]
    role_name: str
    permissions: Union[List[str], Dict[str, Any]]

class LoginResponse(BaseModel):
    token: str
    expires_at: int
    user: UserResponse
