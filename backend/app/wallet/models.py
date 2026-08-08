"""Wallet balance + ledger. One Wallet per user; every balance change is a
WalletTransaction row so `balance` is always reconstructable/auditable.

`idempotency_key` on WalletTransaction is called out as an MVP-blocking gap
in CONTEXT.md ("Idempotency keys on send/wallet-debit") — a retried debit
(e.g. a campaign-send retry after a timeout) must not double-charge.
"""
import enum
import uuid
from decimal import Decimal

from sqlalchemy import Enum, ForeignKey, Numeric, String, UniqueConstraint
from sqlalchemy.orm import Mapped, mapped_column, relationship

from app.database.base import Base, TimestampMixin, UUIDPrimaryKeyMixin


class TransactionType(str, enum.Enum):
    CREDIT = "credit"
    DEBIT = "debit"


class Wallet(UUIDPrimaryKeyMixin, TimestampMixin, Base):
    __tablename__ = "wallets"

    user_id: Mapped[str] = mapped_column(ForeignKey("users.id"), unique=True, nullable=False)
    balance: Mapped[Decimal] = mapped_column(Numeric(14, 4), default=Decimal("0"), nullable=False)
    currency: Mapped[str] = mapped_column(String(3), default="INR", nullable=False)

    transactions: Mapped[list["WalletTransaction"]] = relationship(back_populates="wallet")


class WalletTransaction(UUIDPrimaryKeyMixin, TimestampMixin, Base):
    __tablename__ = "wallet_transactions"
    __table_args__ = (
        # Same idempotency_key can't be applied twice against the same
        # wallet — that's the double-charge guard.
        UniqueConstraint("wallet_id", "idempotency_key", name="uq_wallet_idempotency_key"),
    )

    wallet_id: Mapped[str] = mapped_column(ForeignKey("wallets.id"), nullable=False)
    type: Mapped[TransactionType] = mapped_column(Enum(TransactionType), nullable=False)
    amount: Mapped[Decimal] = mapped_column(Numeric(14, 4), nullable=False)
    balance_after: Mapped[Decimal] = mapped_column(Numeric(14, 4), nullable=False)
    reference: Mapped[str | None] = mapped_column(String(255), nullable=True)
    idempotency_key: Mapped[str] = mapped_column(
        String(64), default=lambda: str(uuid.uuid4()), nullable=False
    )

    wallet: Mapped["Wallet"] = relationship(back_populates="transactions")
