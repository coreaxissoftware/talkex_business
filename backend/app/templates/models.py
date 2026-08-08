"""Message templates — shared engine, per-channel mapping (per CONTEXT.md).

`category` is flagged as MVP-blocking, non-negotiable: it drives Meta's
2025 per-template pricing model and needs to be wired into Billing, not
bolted on later.
"""
import enum
import json

from sqlalchemy import Enum, ForeignKey, String, Text
from sqlalchemy.orm import Mapped, mapped_column

from app.database.base import Base, TimestampMixin, UUIDPrimaryKeyMixin


class TemplateCategory(str, enum.Enum):
    MARKETING = "marketing"
    UTILITY = "utility"
    AUTHENTICATION = "authentication"


class TemplateStatus(str, enum.Enum):
    DRAFT = "draft"
    PENDING_REVIEW = "pending_review"
    APPROVED = "approved"
    REJECTED = "rejected"


class MessageTemplate(UUIDPrimaryKeyMixin, TimestampMixin, Base):
    __tablename__ = "message_templates"

    owner_id: Mapped[str] = mapped_column(ForeignKey("users.id"), index=True, nullable=False)
    name: Mapped[str] = mapped_column(String(255), nullable=False)
    category: Mapped[TemplateCategory] = mapped_column(Enum(TemplateCategory), nullable=False)
    # "talkex" | "whatsapp" | ... — which channel connector this mapping
    # targets; the same logical template can have one row per channel.
    channel: Mapped[str] = mapped_column(String(50), nullable=False)
    body: Mapped[str] = mapped_column(Text, nullable=False)
    # JSON list of variable names used in `body`, e.g. ["customer_name", "order_id"]
    variables: Mapped[str] = mapped_column(Text, default="[]", nullable=False)
    status: Mapped[TemplateStatus] = mapped_column(
        Enum(TemplateStatus), default=TemplateStatus.DRAFT, nullable=False
    )

    def variables_list(self) -> list[str]:
        return json.loads(self.variables)
