from src.database.connection import system_session_scope
from src.database.models import UserRole, User
from src.utils.logger.logger import Log
from src.utils.security import crypto_manager

VERSION_CODE = 1
DESCRIPTION = "Initialize default user roles and admin user"

TAG = "MIGRATION_002"

DEFAULT_ROLES = [
    {
        "name": "admin",
        "description": "System Administrator",
        "permissions": {
            "is_admin": True,
            "can_manage_users": True,
            "can_manage_modules": True,
            "can_manage_system": True
        }
    }
]

DEFAULT_ADMIN = {
    "username": "admin",
    "password": "admin",
    "email": "admin@admin.com",
    "role_name": "admin"
}

def upgrade():
    Log.i(TAG, "Starting upgrade...")
    with system_session_scope() as session:
        # 1. Create Roles
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
                Log.i(TAG, f"Role {role_data['name']} already exists. Skipping.")
        
        # Flush to ensure roles have IDs
        session.flush()

        # 2. Create Default Admin
        admin_role = session.query(UserRole).filter_by(name=DEFAULT_ADMIN["role_name"]).first()
        if admin_role:
            existing_user = session.query(User).filter_by(username=DEFAULT_ADMIN["username"]).first()
            if not existing_user:
                Log.i(TAG, f"Creating default admin user: {DEFAULT_ADMIN['username']}")
                new_user = User(
                    username=DEFAULT_ADMIN["username"],
                    password_hash=crypto_manager.get_password_hash(DEFAULT_ADMIN["password"]),
                    email=DEFAULT_ADMIN["email"],
                    role_id=admin_role.id,
                    is_active=True
                )
                session.add(new_user)
            else:
                Log.i(TAG, "Default admin user already exists. Skipping.")
        else:
            Log.e(TAG, "Failed to find admin role, cannot create admin user.")
