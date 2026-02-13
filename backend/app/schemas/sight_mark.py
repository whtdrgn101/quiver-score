import uuid
from datetime import datetime
from pydantic import BaseModel, Field


class SightMarkCreate(BaseModel):
    equipment_id: uuid.UUID | None = None
    setup_id: uuid.UUID | None = None
    distance: str = Field(..., min_length=1, max_length=50)
    setting: str = Field(..., min_length=1, max_length=100)
    notes: str | None = Field(None, max_length=500)
    date_recorded: datetime


class SightMarkUpdate(BaseModel):
    equipment_id: uuid.UUID | None = None
    setup_id: uuid.UUID | None = None
    distance: str | None = Field(None, min_length=1, max_length=50)
    setting: str | None = Field(None, min_length=1, max_length=100)
    notes: str | None = Field(None, max_length=500)
    date_recorded: datetime | None = None


class SightMarkOut(BaseModel):
    id: uuid.UUID
    equipment_id: uuid.UUID | None
    setup_id: uuid.UUID | None
    distance: str
    setting: str
    notes: str | None
    date_recorded: datetime
    created_at: datetime

    model_config = {"from_attributes": True}
