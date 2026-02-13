from sqlalchemy.ext.asyncio import AsyncSession
from app.models.social import FeedItem


async def create_feed_item(db: AsyncSession, user_id, type: str, data: dict):
    item = FeedItem(user_id=user_id, type=type, data=data)
    db.add(item)
