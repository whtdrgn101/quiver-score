import uuid
from datetime import datetime
from pydantic import BaseModel, Field, HttpUrl


class UserOut(BaseModel):
    id: uuid.UUID
    email: str
    username: str
    display_name: str | None
    bow_type: str | None
    classification: str | None
    bio: str | None
    avatar: str | None
    email_verified: bool
    profile_public: bool
    created_at: datetime

    model_config = {"from_attributes": True}


class UserUpdate(BaseModel):
    display_name: str | None = Field(None, max_length=100)
    bow_type: str | None = Field(None, max_length=50)
    classification: str | None = Field(None, max_length=50)
    bio: str | None = Field(None, max_length=500)
    profile_public: bool | None = None


class PublicProfileOut(BaseModel):
    username: str
    display_name: str | None
    bow_type: str | None
    bio: str | None
    avatar: str | None
    created_at: datetime
    total_sessions: int
    completed_sessions: int
    total_arrows: int
    total_x_count: int
    personal_best_score: int | None = None
    personal_best_template: str | None = None
    recent_sessions: list["PublicSessionSummary"]
    clubs: list["ProfileClubOut"] = []

    model_config = {"from_attributes": True}


class PublicSessionSummary(BaseModel):
    template_name: str | None
    total_score: int
    total_x_count: int
    total_arrows: int
    completed_at: datetime | None
    share_token: str | None


class ProfileClubTeamOut(BaseModel):
    team_id: uuid.UUID
    team_name: str


class ProfileClubOut(BaseModel):
    club_id: uuid.UUID
    club_name: str
    role: str
    teams: list[ProfileClubTeamOut]


class AvatarUrlUpload(BaseModel):
    url: HttpUrl
