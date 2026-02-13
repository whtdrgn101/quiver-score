import uuid
from datetime import datetime
from pydantic import BaseModel


class FollowOut(BaseModel):
    id: uuid.UUID
    follower_id: uuid.UUID
    following_id: uuid.UUID
    follower_username: str | None = None
    following_username: str | None = None
    created_at: datetime

    model_config = {"from_attributes": True}


class FeedItemOut(BaseModel):
    id: uuid.UUID
    user_id: uuid.UUID
    username: str | None = None
    type: str
    data: dict
    created_at: datetime

    model_config = {"from_attributes": True}


class FollowersCountOut(BaseModel):
    followers: int
    following: int
