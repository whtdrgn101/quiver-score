import uuid
from datetime import datetime
from enum import Enum

from pydantic import BaseModel


class EquipmentCategory(str, Enum):
    riser = "riser"
    limbs = "limbs"
    arrows = "arrows"
    sight = "sight"
    stabilizer = "stabilizer"
    rest = "rest"
    release = "release"
    scope = "scope"
    string = "string"
    other = "other"


class EquipmentCreate(BaseModel):
    category: EquipmentCategory
    name: str
    brand: str | None = None
    model: str | None = None
    specs: dict | None = None
    notes: str | None = None


class EquipmentUpdate(BaseModel):
    category: EquipmentCategory | None = None
    name: str | None = None
    brand: str | None = None
    model: str | None = None
    specs: dict | None = None
    notes: str | None = None


class EquipmentOut(BaseModel):
    id: uuid.UUID
    category: str
    name: str
    brand: str | None
    model: str | None
    specs: dict | None
    notes: str | None
    created_at: datetime

    model_config = {"from_attributes": True}


class EquipmentUsageOut(BaseModel):
    item_id: uuid.UUID
    item_name: str
    category: str
    sessions_count: int
    total_arrows: int
    last_used: datetime | None
