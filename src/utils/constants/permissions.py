class Permissions:
    """
    System Permission Nodes
    """
    
    # 1. System Config
    SYSTEM_CONFIG_VIEW = "system.config.view"
    SYSTEM_CONFIG_EDIT = "system.config.edit"
    
    # 2. User Management (Renamed & Merged)
    USER_MANAGER_VIEW = "user_manager.view"
    USER_MANAGER_EDIT = "user_manager.edit"  # Merged edit & manage
    
    # 3. Schedule
    SCHEDULE_VIEW = "schedule.view"
    SCHEDULE_EDIT = "schedule.edit"
    
    # 4. Scraper Modules
    SCRAPER_VIEW = "scraper.view"
    SCRAPER_EDIT = "scraper.edit" # Renamed from manage
    
    # 5. Push Modules (System Level)
    PUSH_MODULE_VIEW = "push_module.view"
    PUSH_MODULE_EDIT = "push_module.edit" # Renamed from manage
    
    # 6. Frontend Config
    FRONTEND_CONFIG_VIEW = "frontend.config.view"
    FRONTEND_CONFIG_EDIT = "frontend.config.edit"
    
    # 7. User Push Config (Personal)
    USER_PUSH_SETTINGS = "user_push.settings" # Renamed

    @classmethod
    def get_all(cls):
        return [v for k, v in cls.__dict__.items() if not k.startswith("__") and isinstance(v, str)]

    @classmethod
    def get_default_admin(cls):
        return cls.get_all()

    @classmethod
    def get_default_user(cls):
        return [
            cls.USER_PUSH_SETTINGS
        ]
