from pydantic import BaseModel
from typing import Optional, List, Dict, Any, Union, Generic, TypeVar

T = TypeVar('T')

class LoginRequest(BaseModel):
    username: str
    password: str
    remember_me: bool = False
    captcha_code: Optional[str] = None
    captcha_id: Optional[str] = None

class UserResponse(BaseModel):
    id: str
    username: str
    nickname: Optional[str] = None
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
    type: str = "text"
    options: Optional[Union[List, Dict]] = None
    hint: Optional[str] = None
    regular: Optional[str] = None
    source: str = "default"
    is_override: bool = False

class ScraperModuleTaskItem(BaseModel):
    id: str
    key: str
    name: str
    description: str

class ScraperModuleDetailResponse(ScraperModuleResponse):
    config: Dict[str, ScraperModuleConfigItem]

class ScraperModuleTaskResponse(BaseModel):
    module_id: str
    module_name: str
    tasks: List[ScraperModuleTaskItem]

class TestModuleResponse(BaseModel):
    success: bool
    message: str

class SystemEventResponse(BaseModel):
    id: str
    level: str
    category: str
    event_type: str
    summary: str
    details: Optional[Dict[str, Any]]
    source_id: Optional[str]
    is_resolved: bool
    created_at: int
    updated_at: Optional[int]

class PaginatedResponse(BaseModel, Generic[T]):
    items: List[T]
    total: int
    page: int
    page_size: int
