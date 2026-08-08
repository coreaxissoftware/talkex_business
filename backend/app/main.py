"""FastAPI app entrypoint. Run with: uvicorn app.main:app --reload"""
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from app.auth.router import router as auth_router
from app.contacts.router import router as contacts_router
from app.core.config import get_settings
from app.middleware.error_handlers import register_error_handlers
from app.templates.router import router as templates_router
from app.users.router import router as users_router
from app.wallet.router import router as wallet_router

settings = get_settings()

app = FastAPI(
    title="TalkEx Business API",
    description="Enterprise CPaaS dashboard API — one platform, multiple messaging channels.",
    version="0.1.0",
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=settings.cors_origin_list,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

register_error_handlers(app)

app.include_router(auth_router)
app.include_router(users_router)
app.include_router(wallet_router)
app.include_router(contacts_router)
app.include_router(templates_router)


@app.get("/health", tags=["health"])
async def health_check():
    return {"status": "ok", "environment": settings.environment}
