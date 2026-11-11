import hashlib
from src.database.connection import system_session_scope
from src.database.models import UserRole, User
from src.utils.logger.logger import Log
from src.utils.security import crypto_manager
from src.utils.constants.permissions import Permissions

VERSION_CODE = 1
DESCRIPTION = "Initialize default user roles and admin user"

TAG = "MIGRATION_002"

DEFAULT_ROLES = [
    {
        "name": "admin",
        "description": "Administrator",
        "permissions": Permissions.get_default_admin()
    },
    {
        "name": "user",
        "description": "User",
        "permissions": Permissions.get_default_user()
    }
]

DEFAULT_ADMIN = {
    "username": "admin",
    "nickname": "Administrator",
    "password": "admin",
    "email": "admin@admin.com",
    "role_name": "admin"
}

def upgrade():
    Log.i(TAG, "Starting upgrade...")
    with system_session_scope() as session:
        for role_data in DEFAULT_ROLES:
            existing = session.query(UserRole).filter_by(name=role_data["name"]).first()
            if not existing:
                Log.i(TAG, f"Adding role: {role_data['name']}")
                new_role = UserRole(
                    name=role_data["name"],
                    description=role_data["description"],
                    permissions=role_data["permissions"]
                )
                session.add(new_role)
            else:
                Log.i(TAG, f"Role {role_data['name']} already exists. Updating permissions.")
                existing.permissions = role_data["permissions"]
        session.flush()
        admin_role = session.query(UserRole).filter_by(name=DEFAULT_ADMIN["role_name"]).first()
        if admin_role:
            existing_user = session.query(User).filter_by(username=DEFAULT_ADMIN["username"]).first()
            if not existing_user:
                Log.i(TAG, f"Creating default admin user: {DEFAULT_ADMIN['username']}")
                raw_password = DEFAULT_ADMIN["password"]
                md5_password = hashlib.md5(raw_password.encode()).hexdigest()
                
                new_user = User(
                    username=DEFAULT_ADMIN["username"],
                    nickname=DEFAULT_ADMIN["nickname"],
                    password_hash=crypto_manager.get_password_hash(md5_password),
                    email=DEFAULT_ADMIN["email"],
                    role_id=admin_role.id,
                    is_active=True
                )
                session.add(new_user)
            else:
                Log.i(TAG, "Default admin user already exists. Skipping.")
        else:
            Log.e(TAG, "Failed to find admin role, cannot create admin user.")
