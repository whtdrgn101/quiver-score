import uuid
from datetime import datetime
from pydantic import BaseModel, Field


class CoachAthleteLinkOut(BaseModel):
    id: uuid.UUID
    coach_id: uuid.UUID
    athlete_id: uuid.UUID
    coach_username: str | None = None
    athlete_username: str | None = None
    status: str
    created_at: datetime

    model_config = {"from_attributes": True}


class InviteRequest(BaseModel):
    athlete_username: str = Field(..., min_length=1)


class RespondRequest(BaseModel):
    link_id: uuid.UUID
    accept: bool


class AnnotationCreate(BaseModel):
    end_number: int | None = None
    arrow_number: int | None = None
    text: str = Field(..., min_length=1, max_length=2000)


class AnnotationOut(BaseModel):
    id: uuid.UUID
    session_id: uuid.UUID
    author_id: uuid.UUID
    author_username: str | None = None
    end_number: int | None
    arrow_number: int | None
    text: str
    created_at: datetime

    model_config = {"from_attributes": True}
