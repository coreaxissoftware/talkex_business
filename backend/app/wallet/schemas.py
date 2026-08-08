from datetime import datetime
from decimal import Decimal

from pydantic import BaseModel, ConfigDict

from app.wallet.models import TransactionType


class WalletRead(BaseModel):
    model_config = ConfigDict(from_attributes=True)

    id: str
    balance: Decimal
    currency: str


class WalletTransactionCreate(BaseModel):
    type: TransactionType
    amount: Decimal
    reference: str | None = None
    idempotency_key: str


class WalletTransactionRead(BaseModel):
    model_config = ConfigDict(from_attributes=True)

    id: str
    type: TransactionType
    amount: Decimal
    balance_after: Decimal
    reference: str | None
    created_at: datetime
