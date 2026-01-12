from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from fastapi.staticfiles import StaticFiles
import os
from src.utils.logger.logger import Log
from src.management.routers import auth

TAG = "MANAGEMENT_API"

def create_app() -> FastAPI:
    app = FastAPI(
        title="Utopia Daily Management API",
        description="Backend API for managing Utopia Daily system.",
        version="1.0.0",
        docs_url="/api/docs",
        redoc_url="/api/redoc",
        openapi_url="/api/openapi.json"
    )

    # CORS Configuration
    app.add_middleware(
        CORSMiddleware,
        allow_origins=["*"],
        allow_credentials=True,
        allow_methods=["*"],
        allow_headers=["*"],
    )
    
    # Register Routers
    app.include_router(auth.router)

    # Mount Frontend (Static Files)
    # Path: src/management/frontend
    current_dir = os.path.dirname(os.path.abspath(__file__))
    frontend_dir = os.path.join(current_dir, "frontend")
    
    if os.path.exists(frontend_dir):
        app.mount("/", StaticFiles(directory=frontend_dir, html=True), name="frontend")
        Log.i(TAG, f"Frontend mounted from {frontend_dir}")
    else:
        Log.w(TAG, f"Frontend directory not found at {frontend_dir}")

    @app.on_event("startup")
    async def startup_event():
        Log.i(TAG, "Management API Server starting...")

    @app.on_event("shutdown")
    async def shutdown_event():
        Log.i(TAG, "Management API Server shutting down...")

    @app.get("/api/health")
    async def health_check():
        return {"status": "ok", "version": "1.0.0"}

    return app

app = create_app()
