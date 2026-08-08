"""User account — the tenant identity every other domain (wallet, contacts,
templates, campaigns...) hangs off via `owner_id`.

Business-verification and quality-rating fields mirror the consumer app's
`users.is_business` / `business_category` / `quality_flagged_at` pattern
(see CONTEXT.md) — reused here as the WhatsApp-quality-rating equivalent
for the TalkEx channel too, not just WhatsApp.
"""
import enum
from datetime import datetime

from sqlalchemy import Boolean, DateTime, Enum, String
from sqlalchemy.orm import Mapped, mapped_column

from app.database.base import Base, TimestampMixin, UUIDPrimaryKeyMixin


class UserRole(str, enum.Enum):
    OWNER = "owner"
    ADMIN = "admin"
    AGENT = "agent"
    DEVELOPER = "developer"


class User(UUIDPrimaryKeyMixin, TimestampMixin, Base):
    __tablename__ = "users"

    email: Mapped[str] = mapped_column(String(255), unique=True, index=True, nullable=False)
    hashed_password: Mapped[str] = mapped_column(String(255), nullable=False)
    full_name: Mapped[str] = mapped_column(String(255), nullable=False)
    role: Mapped[UserRole] = mapped_column(Enum(UserRole), default=UserRole.OWNER, nullable=False)

    is_active: Mapped[bool] = mapped_column(Boolean, default=True, nullable=False)

    # Business Verification (admin-granted, not self-service — see CONTEXT.md
    # "Business Verification / Embedded Signup").
    is_business_verified: Mapped[bool] = mapped_column(Boolean, default=False, nullable=False)
    business_category: Mapped[str | None] = mapped_column(String(100), nullable=True)

    # Auto-suspend after repeated blocks/reports in a rolling window —
    # mirrors WhatsApp quality rating dropping to Red. NULL = not flagged.
    quality_flagged_at: Mapped[datetime | None] = mapped_column(DateTime(timezone=True), nullable=True)
