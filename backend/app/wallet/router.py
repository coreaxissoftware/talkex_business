from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.ext.asyncio import AsyncSession

from app.auth.dependencies import get_current_user
from app.database.session import get_db
from app.users.models import User
from app.wallet.schemas import WalletRead, WalletTransactionCreate, WalletTransactionRead
from app.wallet.service import (
    InsufficientBalanceError,
    apply_transaction,
    get_or_create_wallet,
    list_transactions,
)

router = APIRouter(prefix="/wallet", tags=["wallet"])


@router.get("", response_model=WalletRead)
async def read_wallet(
    current_user: User = Depends(get_current_user), db: AsyncSession = Depends(get_db)
):
    return await get_or_create_wallet(db, current_user.id)


@router.get("/transactions", response_model=list[WalletTransactionRead])
async def read_transactions(
    current_user: User = Depends(get_current_user), db: AsyncSession = Depends(get_db)
):
    wallet = await get_or_create_wallet(db, current_user.id)
    return await list_transactions(db, wallet.id)


@router.post("/transactions", response_model=WalletTransactionRead, status_code=status.HTTP_201_CREATED)
async def create_transaction(
    data: WalletTransactionCreate,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    wallet = await get_or_create_wallet(db, current_user.id)
    try:
        return await apply_transaction(
            db, wallet, data.type, data.amount, data.idempotency_key, data.reference
        )
    except InsufficientBalanceError as exc:
        raise HTTPException(status.HTTP_400_BAD_REQUEST, str(exc)) from exc
