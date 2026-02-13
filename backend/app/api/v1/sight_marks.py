from uuid import UUID

from fastapi import APIRouter, Depends, Query, status
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from app.database import get_db
from app.dependencies import get_current_user
from app.models.user import User
from app.models.sight_mark import SightMark
from app.schemas.sight_mark import SightMarkCreate, SightMarkUpdate, SightMarkOut
from app.core.exceptions import NotFoundError

router = APIRouter(prefix="/sight-marks", tags=["sight-marks"])


@router.get("", response_model=list[SightMarkOut])
async def list_sight_marks(
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
    equipment_id: UUID | None = Query(None),
    setup_id: UUID | None = Query(None),
):
    query = select(SightMark).where(SightMark.user_id == user.id)
    if equipment_id:
        query = query.where(SightMark.equipment_id == equipment_id)
    if setup_id:
        query = query.where(SightMark.setup_id == setup_id)
    result = await db.execute(query.order_by(SightMark.distance, SightMark.date_recorded.desc()))
    return result.scalars().all()


@router.post("", response_model=SightMarkOut, status_code=status.HTTP_201_CREATED)
async def create_sight_mark(
    body: SightMarkCreate,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    sm = SightMark(
        user_id=user.id,
        equipment_id=body.equipment_id,
        setup_id=body.setup_id,
        distance=body.distance,
        setting=body.setting,
        notes=body.notes,
        date_recorded=body.date_recorded,
    )
    db.add(sm)
    await db.commit()
    await db.refresh(sm)
    return sm


@router.put("/{sight_mark_id}", response_model=SightMarkOut)
async def update_sight_mark(
    sight_mark_id: UUID,
    body: SightMarkUpdate,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    result = await db.execute(
        select(SightMark).where(SightMark.id == sight_mark_id, SightMark.user_id == user.id)
    )
    sm = result.scalar_one_or_none()
    if not sm:
        raise NotFoundError("Sight mark not found")
    if body.distance is not None:
        sm.distance = body.distance
    if body.setting is not None:
        sm.setting = body.setting
    if body.notes is not None:
        sm.notes = body.notes
    if body.date_recorded is not None:
        sm.date_recorded = body.date_recorded
    if body.equipment_id is not None:
        sm.equipment_id = body.equipment_id
    if body.setup_id is not None:
        sm.setup_id = body.setup_id
    await db.commit()
    await db.refresh(sm)
    return sm


@router.delete("/{sight_mark_id}", status_code=status.HTTP_204_NO_CONTENT)
async def delete_sight_mark(
    sight_mark_id: UUID,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    result = await db.execute(
        select(SightMark).where(SightMark.id == sight_mark_id, SightMark.user_id == user.id)
    )
    sm = result.scalar_one_or_none()
    if not sm:
        raise NotFoundError("Sight mark not found")
    await db.delete(sm)
    await db.commit()
