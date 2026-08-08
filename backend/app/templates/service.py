import json

from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from app.templates.models import MessageTemplate
from app.templates.schemas import TemplateCreate, TemplateUpdate


async def list_templates(db: AsyncSession, owner_id: str) -> list[MessageTemplate]:
    result = await db.execute(
        select(MessageTemplate)
        .where(MessageTemplate.owner_id == owner_id)
        .order_by(MessageTemplate.created_at.desc())
    )
    return list(result.scalars().all())


async def get_template(db: AsyncSession, owner_id: str, template_id: str) -> MessageTemplate | None:
    result = await db.execute(
        select(MessageTemplate).where(
            MessageTemplate.id == template_id, MessageTemplate.owner_id == owner_id
        )
    )
    return result.scalar_one_or_none()


async def create_template(db: AsyncSession, owner_id: str, data: TemplateCreate) -> MessageTemplate:
    template = MessageTemplate(
        owner_id=owner_id,
        name=data.name,
        category=data.category,
        channel=data.channel,
        body=data.body,
        variables=json.dumps(data.variables),
    )
    db.add(template)
    await db.commit()
    await db.refresh(template)
    return template


async def update_template(
    db: AsyncSession, template: MessageTemplate, data: TemplateUpdate
) -> MessageTemplate:
    if data.name is not None:
        template.name = data.name
    if data.body is not None:
        template.body = data.body
    if data.variables is not None:
        template.variables = json.dumps(data.variables)
    if data.status is not None:
        template.status = data.status
    await db.commit()
    await db.refresh(template)
    return template


async def delete_template(db: AsyncSession, template: MessageTemplate) -> None:
    await db.delete(template)
    await db.commit()
