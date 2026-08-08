from datetime import datetime

from pydantic import BaseModel, ConfigDict, field_validator

from app.templates.models import TemplateCategory, TemplateStatus


class TemplateCreate(BaseModel):
    name: str
    category: TemplateCategory
    channel: str
    body: str
    variables: list[str] = []


class TemplateUpdate(BaseModel):
    name: str | None = None
    body: str | None = None
    variables: list[str] | None = None
    status: TemplateStatus | None = None


class TemplateRead(BaseModel):
    model_config = ConfigDict(from_attributes=True)

    id: str
    name: str
    category: TemplateCategory
    channel: str
    body: str
    variables: list[str]
    status: TemplateStatus
    created_at: datetime

    @field_validator("variables", mode="before")
    @classmethod
    def _parse_variables(cls, v):
        import json

        return json.loads(v) if isinstance(v, str) else v
