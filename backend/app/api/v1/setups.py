from uuid import UUID

from fastapi import APIRouter, Depends, status
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.orm import selectinload

from app.database import get_db
from app.dependencies import get_current_user
from app.models.user import User
from app.models.equipment import Equipment
from app.models.setup_profile import SetupProfile, SetupEquipment
from app.schemas.setup_profile import (
    SetupProfileCreate, SetupProfileUpdate, SetupProfileOut, SetupProfileSummary,
)
from app.core.exceptions import NotFoundError, ConflictError

router = APIRouter(prefix="/setups", tags=["setups"])


async def _get_user_setup(db: AsyncSession, setup_id: UUID, user_id) -> SetupProfile:
    result = await db.execute(
        select(SetupProfile)
        .where(SetupProfile.id == setup_id, SetupProfile.user_id == user_id)
        .options(selectinload(SetupProfile.equipment_links).selectinload(SetupEquipment.equipment))
    )
    setup = result.scalar_one_or_none()
    if not setup:
        raise NotFoundError("Setup profile not found")
    return setup


@router.post("", response_model=SetupProfileOut, status_code=status.HTTP_201_CREATED)
async def create_setup(
    body: SetupProfileCreate,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    setup = SetupProfile(user_id=user.id, **body.model_dump())
    db.add(setup)
    await db.commit()
    setup = await _get_user_setup(db, setup.id, user.id)
    return SetupProfileOut.from_model(setup)


@router.get("", response_model=list[SetupProfileSummary])
async def list_setups(
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    result = await db.execute(
        select(SetupProfile)
        .where(SetupProfile.user_id == user.id)
        .order_by(SetupProfile.name)
        .options(selectinload(SetupProfile.equipment_links))
    )
    setups = result.scalars().all()
    out = []
    for s in setups:
        summary = SetupProfileSummary.model_validate(s)
        summary.equipment_count = len(s.equipment_links)
        out.append(summary)
    return out


@router.get("/{setup_id}", response_model=SetupProfileOut)
async def get_setup(
    setup_id: UUID,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    setup = await _get_user_setup(db, setup_id, user.id)
    return SetupProfileOut.from_model(setup)


@router.put("/{setup_id}", response_model=SetupProfileOut)
async def update_setup(
    setup_id: UUID,
    body: SetupProfileUpdate,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    setup = await _get_user_setup(db, setup_id, user.id)
    for field, value in body.model_dump(exclude_unset=True).items():
        setattr(setup, field, value)
    await db.commit()
    await db.refresh(setup)
    return SetupProfileOut.from_model(setup)


@router.delete("/{setup_id}", status_code=status.HTTP_204_NO_CONTENT)
async def delete_setup(
    setup_id: UUID,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    setup = await _get_user_setup(db, setup_id, user.id)
    await db.delete(setup)
    await db.commit()


@router.post("/{setup_id}/equipment/{equipment_id}", response_model=SetupProfileOut, status_code=status.HTTP_201_CREATED)
async def add_equipment_to_setup(
    setup_id: UUID,
    equipment_id: UUID,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    setup = await _get_user_setup(db, setup_id, user.id)

    # Verify equipment belongs to user
    result = await db.execute(
        select(Equipment).where(Equipment.id == equipment_id, Equipment.user_id == user.id)
    )
    if not result.scalar_one_or_none():
        raise NotFoundError("Equipment not found")

    # Check not already linked
    result = await db.execute(
        select(SetupEquipment).where(
            SetupEquipment.setup_id == setup_id,
            SetupEquipment.equipment_id == equipment_id,
        )
    )
    if result.scalar_one_or_none():
        raise ConflictError("Equipment already linked to this setup")

    link = SetupEquipment(setup_id=setup_id, equipment_id=equipment_id)
    db.add(link)
    await db.commit()
    # Expire cached state and re-fetch with eager loading
    db.expire(setup)
    setup = await _get_user_setup(db, setup_id, user.id)
    return SetupProfileOut.from_model(setup)


@router.delete("/{setup_id}/equipment/{equipment_id}", status_code=status.HTTP_204_NO_CONTENT)
async def remove_equipment_from_setup(
    setup_id: UUID,
    equipment_id: UUID,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    await _get_user_setup(db, setup_id, user.id)

    result = await db.execute(
        select(SetupEquipment).where(
            SetupEquipment.setup_id == setup_id,
            SetupEquipment.equipment_id == equipment_id,
        )
    )
    link = result.scalar_one_or_none()
    if not link:
        raise NotFoundError("Equipment not linked to this setup")
    await db.delete(link)
    await db.commit()
