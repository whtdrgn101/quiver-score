import uuid
from datetime import datetime

from pydantic import BaseModel

from app.schemas.equipment import EquipmentOut


class SetupProfileCreate(BaseModel):
    name: str
    description: str | None = None
    brace_height: float | None = None
    tiller: float | None = None
    draw_weight: float | None = None
    draw_length: float | None = None
    arrow_foc: float | None = None


class SetupProfileUpdate(BaseModel):
    name: str | None = None
    description: str | None = None
    brace_height: float | None = None
    tiller: float | None = None
    draw_weight: float | None = None
    draw_length: float | None = None
    arrow_foc: float | None = None


class SetupProfileOut(BaseModel):
    id: uuid.UUID
    name: str
    description: str | None
    brace_height: float | None
    tiller: float | None
    draw_weight: float | None
    draw_length: float | None
    arrow_foc: float | None
    equipment: list[EquipmentOut]
    created_at: datetime

    model_config = {"from_attributes": True}

    @classmethod
    def from_model(cls, profile):
        data = {
            "id": profile.id,
            "name": profile.name,
            "description": profile.description,
            "brace_height": profile.brace_height,
            "tiller": profile.tiller,
            "draw_weight": profile.draw_weight,
            "draw_length": profile.draw_length,
            "arrow_foc": profile.arrow_foc,
            "created_at": profile.created_at,
            "equipment": [link.equipment for link in profile.equipment_links],
        }
        return cls.model_validate(data)


class SetupProfileSummary(BaseModel):
    id: uuid.UUID
    name: str
    description: str | None
    equipment_count: int = 0
    created_at: datetime

    model_config = {"from_attributes": True}
