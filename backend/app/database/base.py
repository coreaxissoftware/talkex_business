"""Shared declarative base + reusable mixins for every domain model.

Every module's `models.py` (users, wallet, contacts, templates, ...)
imports `Base` from here so they all end up on one MetaData object —
that's what alembic/env.py autogenerates migrations against.
"""
import uuid
from datetime import datetime

from sqlalchemy import DateTime, func
from sqlalchemy.orm import DeclarativeBase, Mapped, mapped_column


class Base(DeclarativeBase):
    pass


class UUIDPrimaryKeyMixin:
    """String(36) UUID pk — portable across SQLite (dev) and Postgres (prod),
    where a native UUID column type would otherwise diverge per-dialect."""

    id: Mapped[str] = mapped_column(
        primary_key=True, default=lambda: str(uuid.uuid4()), index=True
    )


class TimestampMixin:
    created_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), server_default=func.now())
    updated_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True), server_default=func.now(), onupdate=func.now()
    )
