class Permissions:
    """
    System Permission Nodes
    """
    
    # 1. System Config
    SYSTEM_CONFIG_VIEW = "system.config.view"
    SYSTEM_CONFIG_EDIT = "system.config.edit"
    
    # 2. User Management
    USER_VIEW = "user.view"
    USER_EDIT = "user.edit"      # Edit basic info
    USER_MANAGE = "user.manage"  # Create/Delete/Assign Roles
    
    # 3. Schedule
    SCHEDULE_VIEW = "schedule.view"
    SCHEDULE_EDIT = "schedule.edit"
    
    # 4. Scraper Modules
    SCRAPER_VIEW = "scraper.view"
    SCRAPER_MANAGE = "scraper.manage" # Enable/Disable/Config
    
    # 5. Push Modules (System Level)
    PUSH_MODULE_VIEW = "push_module.view"
    PUSH_MODULE_MANAGE = "push_module.manage"
    
    # 6. Frontend Config
    FRONTEND_CONFIG_VIEW = "frontend.config.view"
    FRONTEND_CONFIG_EDIT = "frontend.config.edit"
    
    # 7. User Push Config (Personal)
    # Usually all logged-in users have this, but we can restrict it
    USER_PUSH_CONFIG = "user.push_config"

    # Special
    ADMIN_ACCESS = "admin.access" # Superuser access

    @classmethod
    def get_all(cls):
        return [v for k, v in cls.__dict__.items() if not k.startswith("__") and isinstance(v, str)]

    @classmethod
    def get_default_admin(cls):
        return cls.get_all()

    @classmethod
    def get_default_user(cls):
        return [
            cls.USER_PUSH_CONFIG
        ]
