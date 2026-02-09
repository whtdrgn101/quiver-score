import uuid
from datetime import datetime
from pydantic import BaseModel, Field


# Round templates
class StageOut(BaseModel):
    id: uuid.UUID
    stage_order: int
    name: str
    distance: str | None
    num_ends: int
    arrows_per_end: int
    allowed_values: list[str]
    value_score_map: dict[str, int]
    max_score_per_arrow: int

    model_config = {"from_attributes": True}


class RoundTemplateOut(BaseModel):
    id: uuid.UUID
    name: str
    organization: str
    description: str | None
    is_official: bool
    stages: list[StageOut]

    model_config = {"from_attributes": True}


# Arrows
class ArrowIn(BaseModel):
    score_value: str
    x_pos: float | None = None
    y_pos: float | None = None


class ArrowOut(BaseModel):
    id: uuid.UUID
    arrow_number: int
    score_value: str
    score_numeric: int
    x_pos: float | None
    y_pos: float | None

    model_config = {"from_attributes": True}


# Ends
class EndIn(BaseModel):
    stage_id: uuid.UUID
    arrows: list[ArrowIn]


class EndOut(BaseModel):
    id: uuid.UUID
    end_number: int
    end_total: int
    stage_id: uuid.UUID
    arrows: list[ArrowOut]
    created_at: datetime

    model_config = {"from_attributes": True}


# Sessions
class SessionCreate(BaseModel):
    template_id: uuid.UUID
    setup_profile_id: uuid.UUID | None = None
    notes: str | None = Field(None, max_length=1000)
    location: str | None = Field(None, max_length=200)
    weather: str | None = Field(None, max_length=100)


class SessionOut(BaseModel):
    id: uuid.UUID
    template_id: uuid.UUID
    setup_profile_id: uuid.UUID | None = None
    setup_profile_name: str | None = None
    template: RoundTemplateOut | None = None
    status: str
    total_score: int
    total_x_count: int
    total_arrows: int
    notes: str | None
    location: str | None
    weather: str | None
    started_at: datetime
    completed_at: datetime | None
    ends: list[EndOut]

    model_config = {"from_attributes": True}


class SessionSummary(BaseModel):
    id: uuid.UUID
    template_id: uuid.UUID
    setup_profile_id: uuid.UUID | None = None
    setup_profile_name: str | None = None
    status: str
    total_score: int
    total_x_count: int
    total_arrows: int
    started_at: datetime
    completed_at: datetime | None
    template_name: str | None = None

    model_config = {"from_attributes": True}


class SessionComplete(BaseModel):
    notes: str | None = Field(None, max_length=1000)


class RecentTrendItem(BaseModel):
    score: int
    max_score: int
    template_name: str
    date: datetime


class RoundTypeAvg(BaseModel):
    template_name: str
    avg_score: float
    count: int


class StatsOut(BaseModel):
    total_sessions: int
    completed_sessions: int
    total_arrows: int
    total_x_count: int
    personal_best_score: int | None = None
    personal_best_template: str | None = None
    avg_by_round_type: list[RoundTypeAvg]
    recent_trend: list[RecentTrendItem]
