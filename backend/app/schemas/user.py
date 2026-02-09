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
    created_at: datetime

    model_config = {"from_attributes": True}


class UserUpdate(BaseModel):
    display_name: str | None = Field(None, max_length=100)
    bow_type: str | None = Field(None, max_length=50)
    classification: str | None = Field(None, max_length=50)
    bio: str | None = Field(None, max_length=500)


class AvatarUrlUpload(BaseModel):
    url: HttpUrl
