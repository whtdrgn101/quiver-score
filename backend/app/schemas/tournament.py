import uuid
from datetime import datetime
from pydantic import BaseModel, Field


class TournamentCreate(BaseModel):
    name: str = Field(..., min_length=1, max_length=200)
    description: str | None = Field(None, max_length=2000)
    template_id: uuid.UUID
    max_participants: int | None = Field(None, ge=2, le=10000)
    registration_deadline: datetime | None = None
    start_date: datetime | None = None
    end_date: datetime | None = None


class TournamentUpdate(BaseModel):
    name: str | None = Field(None, min_length=1, max_length=200)
    description: str | None = None
    max_participants: int | None = Field(None, ge=2, le=10000)
    registration_deadline: datetime | None = None
    start_date: datetime | None = None
    end_date: datetime | None = None


class ParticipantOut(BaseModel):
    id: uuid.UUID
    user_id: uuid.UUID
    username: str | None = None
    status: str
    final_score: int | None
    final_x_count: int | None
    rank: int | None
    registered_at: datetime

    model_config = {"from_attributes": True}


class TournamentOut(BaseModel):
    id: uuid.UUID
    name: str
    description: str | None
    organizer_id: uuid.UUID
    organizer_name: str | None = None
    template_id: uuid.UUID
    template_name: str | None = None
    status: str
    max_participants: int | None
    registration_deadline: datetime | None
    start_date: datetime | None
    end_date: datetime | None
    participant_count: int = 0
    club_id: uuid.UUID | None = None
    club_name: str | None = None
    created_at: datetime

    model_config = {"from_attributes": True}


class TournamentDetailOut(TournamentOut):
    participants: list[ParticipantOut] = []


class LeaderboardEntry(BaseModel):
    rank: int
    user_id: uuid.UUID
    username: str | None = None
    final_score: int | None
    final_x_count: int | None
    status: str
