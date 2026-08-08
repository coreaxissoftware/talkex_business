from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.ext.asyncio import AsyncSession

from app.auth.dependencies import get_current_user
from app.database.session import get_db
from app.templates.schemas import TemplateCreate, TemplateRead, TemplateUpdate
from app.templates.service import (
    create_template,
    delete_template,
    get_template,
    list_templates,
    update_template,
)
from app.users.models import User

router = APIRouter(prefix="/templates", tags=["templates"])


@router.get("", response_model=list[TemplateRead])
async def read_templates(
    current_user: User = Depends(get_current_user), db: AsyncSession = Depends(get_db)
):
    return await list_templates(db, current_user.id)


@router.post("", response_model=TemplateRead, status_code=status.HTTP_201_CREATED)
async def create(
    data: TemplateCreate,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    return await create_template(db, current_user.id, data)


async def _get_owned_template_or_404(template_id: str, current_user: User, db: AsyncSession):
    template = await get_template(db, current_user.id, template_id)
    if template is None:
        raise HTTPException(status.HTTP_404_NOT_FOUND, "Template not found")
    return template


@router.get("/{template_id}", response_model=TemplateRead)
async def read_template(
    template_id: str,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    return await _get_owned_template_or_404(template_id, current_user, db)


@router.patch("/{template_id}", response_model=TemplateRead)
async def update(
    template_id: str,
    data: TemplateUpdate,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    template = await _get_owned_template_or_404(template_id, current_user, db)
    return await update_template(db, template, data)


@router.delete("/{template_id}", status_code=status.HTTP_204_NO_CONTENT)
async def delete(
    template_id: str,
    current_user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    template = await _get_owned_template_or_404(template_id, current_user, db)
    await delete_template(db, template)
