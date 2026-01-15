import sys
import os
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from fastapi.staticfiles import StaticFiles
from fastapi.responses import RedirectResponse, FileResponse
from contextlib import asynccontextmanager
from src.utils.logger.logger import Log
from src.web.dashboard.routers import auth as dashboard_auth
from src.web.dashboard.routers import layout as dashboard_layout
from src.web.dashboard.routers import system_config as dashboard_sys_config
from src.web.newspaper.routers import news as newspaper_news
from src.web import common_routers
from src.web.middleware.security import SecurityMiddleware
from src.web.middleware.auth import AuthMiddleware

TAG = "WEB_ENTRY"

@asynccontextmanager
async def lifespan(app: FastAPI):
    # Startup logic
    Log.i(TAG, "Web Server starting...")
    yield
    # Shutdown logic
    Log.i(TAG, "Web Server shutting down...")

def create_app() -> FastAPI:
    app = FastAPI(
        title="Utopia Daily Web API",
        description="Unified Web API for Utopia Daily system.",
        version="1.0.0",
        docs_url="/api/docs",
        redoc_url="/api/redoc",
        openapi_url="/api/openapi.json",
        lifespan=lifespan
    )

    # 1. Security Middleware (First line of defense)
    app.add_middleware(SecurityMiddleware)

    # 2. CORS Configuration
    app.add_middleware(
        CORSMiddleware,
        allow_origins=["*"],
        allow_credentials=True,
        allow_methods=["*"],
        allow_headers=["*"],
    )
    
    # 3. Auth Middleware (Injects user into request.state)
    app.add_middleware(AuthMiddleware)
    
    # Register Common Routers (i18n, etc.)
    app.include_router(common_routers.router)

    # Register Dashboard Routers
    app.include_router(dashboard_auth.router)
    app.include_router(dashboard_layout.router)
    app.include_router(dashboard_sys_config.router)

    # Register Newspaper Routers
    app.include_router(newspaper_news.router)
    current_dir = os.path.dirname(os.path.abspath(__file__))
    dashboard_dist_dir = os.path.join(current_dir, "dashboard", "frontend", "dist")
    dashboard_src_dir = os.path.join(current_dir, "dashboard", "frontend")
    dashboard_dir_to_mount = None
    
    if os.path.exists(dashboard_dist_dir):
        dashboard_dir_to_mount = dashboard_dist_dir
        Log.i(TAG, f"Found Dashboard build artifacts. Mounting: {dashboard_dist_dir}")
    elif os.path.exists(dashboard_src_dir):
        Log.e(TAG, "NO WEB FRONTEND")
        sys.exit(1)

    if dashboard_dir_to_mount:
        @app.get("/dashboard", include_in_schema=False)
        async def dashboard_redirect():
            return RedirectResponse(url="/dashboard/")
        
        app.mount("/dashboard", StaticFiles(directory=dashboard_dir_to_mount, html=True), name="dashboard_frontend")

    themes_dir = os.path.join(current_dir, "newspaper", "frontend", "theme")
    if os.path.exists(themes_dir):
        app.mount("/themes", StaticFiles(directory=themes_dir, html=True), name="themes")
        Log.i(TAG, f"Themes directory mounted from {themes_dir}")
    else:
        Log.w(TAG, f"Themes directory not found at {themes_dir}")

    newspaper_frontend_dir = os.path.join(current_dir, "newspaper", "frontend")
    if os.path.exists(newspaper_frontend_dir):
        app.mount("/", StaticFiles(directory=newspaper_frontend_dir, html=True), name="newspaper_root")
        Log.i(TAG, f"Newspaper root frontend mounted from {newspaper_frontend_dir}")

    @app.get("/api/health")
    async def health_check():
        return {"status": "ok", "version": "1.0.0"}

    return app

app = create_app()
