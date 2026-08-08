"""Contacts — shared across every channel (per CONTEXT.md: Contacts/
Templates/Campaigns/Analytics/Webhooks/API Keys are shared; only the
Channel Connector differs).

`opted_in` / `opted_in_at` mirror the consumer app's `business_optins`
table + opt-in endpoints — the consent gate required before a business can
message a contact outside an open conversation window.
"""
import json
from datetime import datetime

from sqlalchemy import Boolean, DateTime, ForeignKey, String, Text
from sqlalchemy.orm import Mapped, mapped_column

from app.database.base import Base, TimestampMixin, UUIDPrimaryKeyMixin


class Contact(UUIDPrimaryKeyMixin, TimestampMixin, Base):
    __tablename__ = "contacts"

    owner_id: Mapped[str] = mapped_column(ForeignKey("users.id"), index=True, nullable=False)
    phone_number: Mapped[str] = mapped_column(String(20), index=True, nullable=False)
    name: Mapped[str | None] = mapped_column(String(255), nullable=True)
    email: Mapped[str | None] = mapped_column(String(255), nullable=True)

    # JSON-encoded lists/dicts kept as Text for SQLite/Postgres portability
    # (avoids committing to JSONB now, before real query patterns exist).
    tags: Mapped[str] = mapped_column(Text, default="[]", nullable=False)
    custom_fields: Mapped[str] = mapped_column(Text, default="{}", nullable=False)

    # Consent gate — mirrors the consumer app's business_optins table
    # (see CONTEXT.md). Required before business-initiated sends outside
    # an open 24h conversation window.
    opted_in: Mapped[bool] = mapped_column(Boolean, default=False, nullable=False)
    opted_in_at: Mapped[datetime | None] = mapped_column(DateTime(timezone=True), nullable=True)

    # Last inbound message timestamp — drives has_open_conversation_window()
    # equivalent logic (24h customer-service-window, see CONTEXT.md).
    last_inbound_at: Mapped[datetime | None] = mapped_column(DateTime(timezone=True), nullable=True)

    def tags_list(self) -> list[str]:
        return json.loads(self.tags)

    def custom_fields_dict(self) -> dict:
        return json.loads(self.custom_fields)
