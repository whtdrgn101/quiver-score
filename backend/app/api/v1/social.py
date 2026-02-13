from uuid import UUID

from fastapi import APIRouter, Depends, Query, status
from sqlalchemy import select, func as sa_func
from sqlalchemy.ext.asyncio import AsyncSession

from app.database import get_db
from app.dependencies import get_current_user
from app.models.user import User
from app.models.social import Follow, FeedItem
from app.schemas.social import FollowOut, FeedItemOut, FollowersCountOut
from app.core.exceptions import NotFoundError, ConflictError, ValidationError

router = APIRouter(prefix="/social", tags=["social"])


@router.post("/follow/{user_id}", response_model=FollowOut, status_code=status.HTTP_201_CREATED)
async def follow_user(
    user_id: UUID,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    if user_id == user.id:
        raise ValidationError("Cannot follow yourself")

    # Check target exists
    target = await db.execute(select(User).where(User.id == user_id))
    if not target.scalar_one_or_none():
        raise NotFoundError("User not found")

    # Check not already following
    existing = await db.execute(
        select(Follow).where(Follow.follower_id == user.id, Follow.following_id == user_id)
    )
    if existing.scalar_one_or_none():
        raise ConflictError("Already following this user")

    follow = Follow(follower_id=user.id, following_id=user_id)
    db.add(follow)
    await db.commit()
    await db.refresh(follow)
    return FollowOut(
        id=follow.id,
        follower_id=follow.follower_id,
        following_id=follow.following_id,
        follower_username=follow.follower.username if follow.follower else None,
        following_username=follow.following.username if follow.following else None,
        created_at=follow.created_at,
    )


@router.delete("/follow/{user_id}", status_code=status.HTTP_204_NO_CONTENT)
async def unfollow_user(
    user_id: UUID,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    result = await db.execute(
        select(Follow).where(Follow.follower_id == user.id, Follow.following_id == user_id)
    )
    follow = result.scalar_one_or_none()
    if not follow:
        raise NotFoundError("Not following this user")
    await db.delete(follow)
    await db.commit()


@router.get("/followers", response_model=list[FollowOut])
async def list_followers(
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    result = await db.execute(
        select(Follow).where(Follow.following_id == user.id).order_by(Follow.created_at.desc())
    )
    return [
        FollowOut(
            id=f.id,
            follower_id=f.follower_id,
            following_id=f.following_id,
            follower_username=f.follower.username if f.follower else None,
            following_username=f.following.username if f.following else None,
            created_at=f.created_at,
        )
        for f in result.scalars().all()
    ]


@router.get("/following", response_model=list[FollowOut])
async def list_following(
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    result = await db.execute(
        select(Follow).where(Follow.follower_id == user.id).order_by(Follow.created_at.desc())
    )
    return [
        FollowOut(
            id=f.id,
            follower_id=f.follower_id,
            following_id=f.following_id,
            follower_username=f.follower.username if f.follower else None,
            following_username=f.following.username if f.following else None,
            created_at=f.created_at,
        )
        for f in result.scalars().all()
    ]


@router.get("/feed", response_model=list[FeedItemOut])
async def get_feed(
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
    limit: int = Query(20, ge=1, le=100),
    offset: int = Query(0, ge=0),
):
    # Get IDs of users I follow
    following_result = await db.execute(
        select(Follow.following_id).where(Follow.follower_id == user.id)
    )
    following_ids = [row[0] for row in following_result.all()]

    if not following_ids:
        return []

    result = await db.execute(
        select(FeedItem)
        .where(FeedItem.user_id.in_(following_ids))
        .order_by(FeedItem.created_at.desc())
        .offset(offset)
        .limit(limit)
    )
    return [
        FeedItemOut(
            id=item.id,
            user_id=item.user_id,
            username=item.user.username if item.user else None,
            type=item.type,
            data=item.data,
            created_at=item.created_at,
        )
        for item in result.scalars().all()
    ]
