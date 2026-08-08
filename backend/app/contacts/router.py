from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.ext.asyncio import AsyncSession

from app.auth.dependencies import get_current_user
from app.contacts.schemas import ContactCreate, ContactRead, ContactUpdate
from app.contacts.service import (
    create_contact,
    delete_contact,
    get_contact,
    list_contacts,
    update_contact,
)
from app.database.session import get_db
from app.users.models import User

router = APIRouter(prefix="/contacts", tags=["contacts"])


@router.get("", response_model=list[ContactRead])
async def read_contacts(
    current_user: User = Depends(get_current_user), db: AsyncSession = Depends(get_db)
):
    return await list_contacts(db, current_user.id)


@router.post("", response_model=ContactRead, status_code=status.HTTP_201_CREATED)
async def create(
    data: ContactCreate,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    return await create_contact(db, current_user.id, data)


async def _get_owned_contact_or_404(contact_id: str, current_user: User, db: AsyncSession):
    contact = await get_contact(db, current_user.id, contact_id)
    if contact is None:
        raise HTTPException(status.HTTP_404_NOT_FOUND, "Contact not found")
    return contact


@router.get("/{contact_id}", response_model=ContactRead)
async def read_contact(
    contact_id: str,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    return await _get_owned_contact_or_404(contact_id, current_user, db)


@router.patch("/{contact_id}", response_model=ContactRead)
async def update(
    contact_id: str,
    data: ContactUpdate,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    contact = await _get_owned_contact_or_404(contact_id, current_user, db)
    return await update_contact(db, contact, data)


@router.delete("/{contact_id}", status_code=status.HTTP_204_NO_CONTENT)
async def delete(
    contact_id: str,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    contact = await _get_owned_contact_or_404(contact_id, current_user, db)
    await delete_contact(db, contact)
