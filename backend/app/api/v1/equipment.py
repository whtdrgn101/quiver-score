from uuid import UUID

from fastapi import APIRouter, Depends, status
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from app.database import get_db
from app.dependencies import get_current_user
from app.models.user import User
from app.models.equipment import Equipment
from app.schemas.equipment import EquipmentCreate, EquipmentUpdate, EquipmentOut
from app.core.exceptions import NotFoundError

router = APIRouter(prefix="/equipment", tags=["equipment"])


@router.post("", response_model=EquipmentOut, status_code=status.HTTP_201_CREATED)
async def create_equipment(
    body: EquipmentCreate,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    equipment = Equipment(user_id=user.id, **body.model_dump())
    db.add(equipment)
    await db.commit()
    await db.refresh(equipment)
    return equipment


@router.get("", response_model=list[EquipmentOut])
async def list_equipment(
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    result = await db.execute(
        select(Equipment)
        .where(Equipment.user_id == user.id)
        .order_by(Equipment.category, Equipment.name)
    )
    return result.scalars().all()


@router.get("/{equipment_id}", response_model=EquipmentOut)
async def get_equipment(
    equipment_id: UUID,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    result = await db.execute(
        select(Equipment).where(Equipment.id == equipment_id, Equipment.user_id == user.id)
    )
    equipment = result.scalar_one_or_none()
    if not equipment:
        raise NotFoundError("Equipment not found")
    return equipment


@router.put("/{equipment_id}", response_model=EquipmentOut)
async def update_equipment(
    equipment_id: UUID,
    body: EquipmentUpdate,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    result = await db.execute(
        select(Equipment).where(Equipment.id == equipment_id, Equipment.user_id == user.id)
    )
    equipment = result.scalar_one_or_none()
    if not equipment:
        raise NotFoundError("Equipment not found")
    for field, value in body.model_dump(exclude_unset=True).items():
        setattr(equipment, field, value)
    await db.commit()
    await db.refresh(equipment)
    return equipment


@router.delete("/{equipment_id}", status_code=status.HTTP_204_NO_CONTENT)
async def delete_equipment(
    equipment_id: UUID,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    result = await db.execute(
        select(Equipment).where(Equipment.id == equipment_id, Equipment.user_id == user.id)
    )
    equipment = result.scalar_one_or_none()
    if not equipment:
        raise NotFoundError("Equipment not found")
    await db.delete(equipment)
    await db.commit()
