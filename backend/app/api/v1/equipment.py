from uuid import UUID

from fastapi import APIRouter, Depends, status
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from app.database import get_db
from app.dependencies import get_current_user
from app.models.user import User
from app.models.equipment import Equipment
from app.models.scoring import ScoringSession
from app.models.setup_profile import SetupProfile, SetupEquipment
from app.schemas.equipment import EquipmentCreate, EquipmentUpdate, EquipmentOut, EquipmentUsageOut
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


@router.get("/stats", response_model=list[EquipmentUsageOut])
async def equipment_stats(
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    # Get all user equipment
    eq_result = await db.execute(
        select(Equipment).where(Equipment.user_id == user.id)
    )
    all_equipment = eq_result.scalars().all()

    # Get sessions with setup profiles linked to equipment
    result = await db.execute(
        select(ScoringSession, SetupEquipment.equipment_id)
        .join(SetupProfile, ScoringSession.setup_profile_id == SetupProfile.id)
        .join(SetupEquipment, SetupEquipment.setup_id == SetupProfile.id)
        .where(ScoringSession.user_id == user.id)
    )
    rows = result.all()

    # Aggregate per equipment item
    from collections import defaultdict
    usage: dict[str, dict] = defaultdict(lambda: {"sessions": set(), "arrows": 0, "last_used": None})
    for session, eq_id in rows:
        key = str(eq_id)
        usage[key]["sessions"].add(session.id)
        usage[key]["arrows"] += session.total_arrows
        ts = session.completed_at or session.started_at
        if usage[key]["last_used"] is None or ts > usage[key]["last_used"]:
            usage[key]["last_used"] = ts

    out = []
    for eq in all_equipment:
        key = str(eq.id)
        u = usage.get(key)
        out.append(EquipmentUsageOut(
            item_id=eq.id,
            item_name=eq.name,
            category=eq.category,
            sessions_count=len(u["sessions"]) if u else 0,
            total_arrows=u["arrows"] if u else 0,
            last_used=u["last_used"] if u else None,
        ))
    return out


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
