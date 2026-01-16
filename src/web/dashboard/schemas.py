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

class ScraperModuleResponse(BaseModel):
    id: str
    module_id: str
    name: str
    description: Optional[str]
    version: str
    author: Optional[str]
    is_enable: bool
    meta: Optional[Dict[str, Any]]
    source: Optional[str] = "unknown"
    updated_at: int
    created_at: int

class ScraperModuleConfigItem(BaseModel):
    value: Any
    description: str
    hint: str
    regular: str

class ScraperModuleTaskItem(BaseModel):
    cron: str
    description: str

class ScraperModuleDetailResponse(ScraperModuleResponse):
    config: Dict[str, ScraperModuleConfigItem]
    tasks: Dict[str, ScraperModuleTaskItem]

class TestModuleResponse(BaseModel):
    success: bool
    message: str
