class Permissions:
    
    SYSTEM_STATUS_VIEW = "system.status.view"

    SYSTEM_CONFIG_VIEW = "system.config.view"
    SYSTEM_CONFIG_EDIT = "system.config.edit"

    USER_MANAGER_VIEW = "user_manager.view"
    USER_MANAGER_EDIT = "user_manager.edit"  # Merged edit & manage

    SCHEDULE_VIEW = "schedule.view"
    SCHEDULE_EDIT = "schedule.edit"

    SCRAPER_VIEW = "scraper.view"
    SCRAPER_EDIT = "scraper.edit" # Renamed from manage

    PUSH_MODULE_VIEW = "push_module.view"
    PUSH_MODULE_EDIT = "push_module.edit" # Renamed from manage

    FRONTEND_CONFIG_VIEW = "frontend.config.view"
    FRONTEND_CONFIG_EDIT = "frontend.config.edit"

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
