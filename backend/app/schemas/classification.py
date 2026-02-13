import uuid
from datetime import datetime
from pydantic import BaseModel


class ClassificationRecordOut(BaseModel):
    id: uuid.UUID
    system: str
    classification: str
    round_type: str
    score: int
    achieved_at: datetime
    session_id: uuid.UUID

    model_config = {"from_attributes": True}


class CurrentClassificationOut(BaseModel):
    system: str
    classification: str
    round_type: str
    score: int
    achieved_at: datetime
