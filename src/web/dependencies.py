from fastapi import Request, HTTPException, Depends
from src.utils.constants.permissions import Permissions
from src.database.connection import system_db_manager

def get_db():
    session = system_db_manager.get_session()
    try:
        yield session
    finally:
        session.close()

def get_current_user(request: Request):
    user = getattr(request.state, "user", None)
    if not user:
        raise HTTPException(status_code=401, detail="Not authenticated")
    return user

class RequirePermission:
    def __init__(self, permission: str):
        self.permission = permission

    def __call__(self, request: Request, user = Depends(get_current_user)):
        user_perms = getattr(request.state, "permissions", [])
        
        # Check if user has the required permission
        if self.permission not in user_perms:
            raise HTTPException(status_code=403, detail=f"Missing permission: {self.permission}")
        return True
