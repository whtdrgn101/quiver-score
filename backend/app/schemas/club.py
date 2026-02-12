import uuid
from datetime import datetime

from pydantic import BaseModel, Field


# ── Club ──────────────────────────────────────────────────────────────

class ClubCreate(BaseModel):
    name: str = Field(min_length=1, max_length=100)
    description: str | None = Field(None, max_length=500)


class ClubUpdate(BaseModel):
    name: str | None = Field(None, min_length=1, max_length=100)
    description: str | None = Field(None, max_length=500)


class ClubMemberOut(BaseModel):
    user_id: uuid.UUID
    username: str
    display_name: str | None
    avatar: str | None
    role: str
    joined_at: datetime

    model_config = {"from_attributes": True}


class ClubOut(BaseModel):
    id: uuid.UUID
    name: str
    description: str | None
    avatar: str | None
    owner_id: uuid.UUID
    member_count: int
    my_role: str | None = None
    created_at: datetime

    model_config = {"from_attributes": True}


class ClubDetailOut(ClubOut):
    members: list[ClubMemberOut]


# ── Invites ───────────────────────────────────────────────────────────

class InviteCreate(BaseModel):
    max_uses: int | None = Field(None, ge=1)
    expires_in_hours: int | None = Field(None, ge=1, le=720)


class InviteOut(BaseModel):
    id: uuid.UUID
    code: str
    url: str
    max_uses: int | None
    use_count: int
    expires_at: datetime | None
    active: bool
    created_at: datetime

    model_config = {"from_attributes": True}


class JoinResult(BaseModel):
    club_id: uuid.UUID
    club_name: str
    role: str


# ── Leaderboard ───────────────────────────────────────────────────────

class LeaderboardEntry(BaseModel):
    user_id: uuid.UUID
    username: str
    display_name: str | None
    avatar: str | None
    best_score: int
    best_x_count: int
    session_id: uuid.UUID
    achieved_at: datetime


class LeaderboardOut(BaseModel):
    template_id: uuid.UUID
    template_name: str
    entries: list[LeaderboardEntry]


# ── Activity ──────────────────────────────────────────────────────────

class ActivityItem(BaseModel):
    type: str  # "session_completed" or "personal_record"
    user_id: uuid.UUID
    username: str
    display_name: str | None
    avatar: str | None
    template_name: str
    score: int
    x_count: int
    session_id: uuid.UUID
    occurred_at: datetime


# ── Events ────────────────────────────────────────────────────────────

class EventCreate(BaseModel):
    name: str = Field(min_length=1, max_length=200)
    description: str | None = Field(None, max_length=1000)
    template_id: uuid.UUID
    event_date: datetime
    location: str | None = Field(None, max_length=200)


class EventUpdate(BaseModel):
    name: str | None = Field(None, min_length=1, max_length=200)
    description: str | None = Field(None, max_length=1000)
    event_date: datetime | None = None
    location: str | None = Field(None, max_length=200)


class EventParticipantOut(BaseModel):
    user_id: uuid.UUID
    username: str
    display_name: str | None
    avatar: str | None
    status: str
    score: int | None = None
    x_count: int | None = None
    session_id: uuid.UUID | None = None


class EventOut(BaseModel):
    id: uuid.UUID
    club_id: uuid.UUID
    name: str
    description: str | None
    template_id: uuid.UUID
    template_name: str | None = None
    event_date: datetime
    location: str | None
    created_by: uuid.UUID
    participants: list[EventParticipantOut]
    created_at: datetime

    model_config = {"from_attributes": True}


class EventRSVP(BaseModel):
    status: str = Field(pattern=r"^(going|maybe|declined)$")
