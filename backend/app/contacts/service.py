import json

from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from app.contacts.models import Contact
from app.contacts.schemas import ContactCreate, ContactUpdate


async def list_contacts(db: AsyncSession, owner_id: str) -> list[Contact]:
    result = await db.execute(
        select(Contact).where(Contact.owner_id == owner_id).order_by(Contact.created_at.desc())
    )
    return list(result.scalars().all())


async def get_contact(db: AsyncSession, owner_id: str, contact_id: str) -> Contact | None:
    result = await db.execute(
        select(Contact).where(Contact.id == contact_id, Contact.owner_id == owner_id)
    )
    return result.scalar_one_or_none()


async def create_contact(db: AsyncSession, owner_id: str, data: ContactCreate) -> Contact:
    contact = Contact(
        owner_id=owner_id,
        phone_number=data.phone_number,
        name=data.name,
        email=data.email,
        tags=json.dumps(data.tags),
        custom_fields=json.dumps(data.custom_fields),
    )
    db.add(contact)
    await db.commit()
    await db.refresh(contact)
    return contact


async def update_contact(db: AsyncSession, contact: Contact, data: ContactUpdate) -> Contact:
    if data.name is not None:
        contact.name = data.name
    if data.email is not None:
        contact.email = data.email
    if data.tags is not None:
        contact.tags = json.dumps(data.tags)
    if data.custom_fields is not None:
        contact.custom_fields = json.dumps(data.custom_fields)
    await db.commit()
    await db.refresh(contact)
    return contact


async def delete_contact(db: AsyncSession, contact: Contact) -> None:
    await db.delete(contact)
    await db.commit()
