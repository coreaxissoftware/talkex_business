from decimal import Decimal

from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from app.wallet.models import TransactionType, Wallet, WalletTransaction


class InsufficientBalanceError(Exception):
    pass


async def get_or_create_wallet(db: AsyncSession, user_id: str) -> Wallet:
    result = await db.execute(select(Wallet).where(Wallet.user_id == user_id))
    wallet = result.scalar_one_or_none()
    if wallet is None:
        wallet = Wallet(user_id=user_id)
        db.add(wallet)
        await db.commit()
        await db.refresh(wallet)
    return wallet


async def list_transactions(db: AsyncSession, wallet_id: str) -> list[WalletTransaction]:
    result = await db.execute(
        select(WalletTransaction)
        .where(WalletTransaction.wallet_id == wallet_id)
        .order_by(WalletTransaction.created_at.desc())
    )
    return list(result.scalars().all())


async def apply_transaction(
    db: AsyncSession,
    wallet: Wallet,
    type_: TransactionType,
    amount: Decimal,
    idempotency_key: str,
    reference: str | None = None,
) -> WalletTransaction:
    """Credit/debit `wallet` by `amount`, guarded by `idempotency_key` so a
    retried call (e.g. a campaign-send retry) returns the original result
    instead of applying twice."""
    existing = await db.execute(
        select(WalletTransaction).where(
            WalletTransaction.wallet_id == wallet.id,
            WalletTransaction.idempotency_key == idempotency_key,
        )
    )
    existing_txn = existing.scalar_one_or_none()
    if existing_txn is not None:
        return existing_txn

    delta = amount if type_ == TransactionType.CREDIT else -amount
    new_balance = wallet.balance + delta
    if new_balance < 0:
        raise InsufficientBalanceError("Wallet balance cannot go negative")

    wallet.balance = new_balance
    txn = WalletTransaction(
        wallet_id=wallet.id,
        type=type_,
        amount=amount,
        balance_after=new_balance,
        reference=reference,
        idempotency_key=idempotency_key,
    )
    db.add(txn)
    await db.commit()
    await db.refresh(txn)
    return txn
