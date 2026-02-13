import uuid
from datetime import datetime
from pydantic import BaseModel, Field, field_validator


# Round templates
class StageCreate(BaseModel):
    name: str = Field(..., min_length=1, max_length=100)
    distance: str | None = Field(None, max_length=50)
    num_ends: int = Field(..., gt=0)
    arrows_per_end: int = Field(..., gt=0)
    allowed_values: list[str] = Field(..., min_length=1)
    value_score_map: dict[str, int]
    max_score_per_arrow: int = Field(..., gt=0)

    @field_validator("value_score_map")
    @classmethod
    def validate_map_matches_values(cls, v, info):
        allowed = info.data.get("allowed_values", [])
        for val in allowed:
            if val not in v:
                raise ValueError(f"value_score_map missing key '{val}'")
        return v


class RoundTemplateCreate(BaseModel):
    name: str = Field(..., min_length=1, max_length=100)
    organization: str = Field(..., min_length=1, max_length=50)
    description: str | None = Field(None, max_length=500)
    stages: list[StageCreate] = Field(..., min_length=1)


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
    created_by: uuid.UUID | None = None
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
    stage_id: uuid.UUID | None = None
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
    share_token: str | None = None
    is_personal_best: bool = False
    started_at: datetime
    completed_at: datetime | None
    ends: list[EndOut]

    model_config = {"from_attributes": True}


class ShareLinkOut(BaseModel):
    share_token: str
    url: str


class SharedSessionOut(BaseModel):
    archer_name: str
    archer_avatar: str | None = None
    template_name: str
    template: RoundTemplateOut | None = None
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
    location: str | None = Field(None, max_length=200)
    weather: str | None = Field(None, max_length=100)


class RecentTrendItem(BaseModel):
    score: int
    max_score: int
    template_name: str
    date: datetime


class RoundTypeAvg(BaseModel):
    template_name: str
    avg_score: float
    count: int


class PersonalRecordOut(BaseModel):
    template_name: str
    score: int
    max_score: int
    achieved_at: datetime
    session_id: uuid.UUID


class TrendDataItem(BaseModel):
    session_id: uuid.UUID
    template_name: str
    total_score: int
    max_score: int
    percentage: float
    completed_at: datetime


class StatsOut(BaseModel):
    total_sessions: int
    completed_sessions: int
    total_arrows: int
    total_x_count: int
    personal_best_score: int | None = None
    personal_best_template: str | None = None
    avg_by_round_type: list[RoundTypeAvg]
    recent_trend: list[RecentTrendItem]
    personal_records: list[PersonalRecordOut] = []
