from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from fastapi.staticfiles import StaticFiles
from fastapi.responses import RedirectResponse, FileResponse
from contextlib import asynccontextmanager
import os
from src.utils.logger.logger import Log
from src.web.dashboard.routers import auth as dashboard_auth
from src.web.newspaper.routers import news as newspaper_news
from src.web import common_routers

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

    # CORS Configuration
    app.add_middleware(
        CORSMiddleware,
        allow_origins=["*"],
        allow_credentials=True,
        allow_methods=["*"],
        allow_headers=["*"],
    )
    
    # Register Common Routers (i18n, etc.)
    app.include_router(common_routers.router)

    # Register Dashboard Routers
    app.include_router(dashboard_auth.router)

    # Register Newspaper Routers
    app.include_router(newspaper_news.router)

    current_dir = os.path.dirname(os.path.abspath(__file__))

    # 1. Mount Dashboard Frontend
    dashboard_frontend_dir = os.path.join(current_dir, "dashboard", "frontend")
    if os.path.exists(dashboard_frontend_dir):
        @app.get("/dashboard", include_in_schema=False)
        async def dashboard_redirect():
            return RedirectResponse(url="/dashboard/")
        app.mount("/dashboard", StaticFiles(directory=dashboard_frontend_dir, html=True), name="dashboard_frontend")
        Log.i(TAG, f"Dashboard frontend mounted from {dashboard_frontend_dir}")

    # 2. Mount Themes Directory
    # This allows accessing themes via /themes/default/index.html, /themes/dark/index.html, etc.
    themes_dir = os.path.join(current_dir, "newspaper", "frontend", "theme")
    if os.path.exists(themes_dir):
        app.mount("/themes", StaticFiles(directory=themes_dir, html=True), name="themes")
        Log.i(TAG, f"Themes directory mounted from {themes_dir}")
    else:
        Log.w(TAG, f"Themes directory not found at {themes_dir}")

    # 3. Mount Root Frontend (The Loader)
    # This serves the index.html that redirects to the active theme
    newspaper_frontend_dir = os.path.join(current_dir, "newspaper", "frontend")
    if os.path.exists(newspaper_frontend_dir):
        app.mount("/", StaticFiles(directory=newspaper_frontend_dir, html=True), name="newspaper_root")
        Log.i(TAG, f"Newspaper root frontend mounted from {newspaper_frontend_dir}")

    @app.get("/api/health")
    async def health_check():
        return {"status": "ok", "version": "1.0.0"}

    return app

app = create_app()
