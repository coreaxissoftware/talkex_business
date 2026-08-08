from datetime import datetime

from pydantic import BaseModel, ConfigDict, field_validator


class ContactCreate(BaseModel):
    phone_number: str
    name: str | None = None
    email: str | None = None
    tags: list[str] = []
    custom_fields: dict = {}


class ContactUpdate(BaseModel):
    name: str | None = None
    email: str | None = None
    tags: list[str] | None = None
    custom_fields: dict | None = None


class ContactRead(BaseModel):
    model_config = ConfigDict(from_attributes=True)

    id: str
    phone_number: str
    name: str | None
    email: str | None
    tags: list[str]
    custom_fields: dict
    opted_in: bool
    opted_in_at: datetime | None
    last_inbound_at: datetime | None
    created_at: datetime

    @field_validator("tags", mode="before")
    @classmethod
    def _parse_tags(cls, v):
        import json

        return json.loads(v) if isinstance(v, str) else v

    @field_validator("custom_fields", mode="before")
    @classmethod
    def _parse_custom_fields(cls, v):
        import json

        return json.loads(v) if isinstance(v, str) else v
