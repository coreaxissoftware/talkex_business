"""Central settings, loaded from environment / .env via pydantic-settings.

Every other module reads config through `get_settings()` (cached) rather
than `os.environ` directly, so tests can override settings by dependency-
overriding this one function.
"""
from functools import lru_cache

from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(env_file=".env", env_file_encoding="utf-8", extra="ignore")

    environment: str = "development"

    # SQLite for local dev, Postgres for production — one DATABASE_URL env
    # var switches both, per CONTEXT.md. Must be an async-driver URL
    # (sqlite+aiosqlite:// / postgresql+asyncpg://) since the app engine is
    # async; alembic/env.py converts it to a sync driver for migrations.
    database_url: str = "sqlite+aiosqlite:///./talkex_business.db"

    jwt_secret_key: str = "changeme-generate-a-real-secret"
    jwt_algorithm: str = "HS256"
    access_token_expire_minutes: int = 15
    refresh_token_expire_days: int = 30

    cors_origins: str = "http://localhost:5173"

    @property
    def cors_origin_list(self) -> list[str]:
        return [origin.strip() for origin in self.cors_origins.split(",") if origin.strip()]


@lru_cache
def get_settings() -> Settings:
    return Settings()
