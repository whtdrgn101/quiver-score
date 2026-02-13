from fastapi import APIRouter, Depends
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from app.database import get_db
from app.dependencies import get_current_user
from app.models.user import User
from app.models.classification import ClassificationRecord
from app.schemas.classification import ClassificationRecordOut, CurrentClassificationOut

router = APIRouter(prefix="/users/me/classifications", tags=["classifications"])


@router.get("", response_model=list[ClassificationRecordOut])
async def list_classifications(
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    result = await db.execute(
        select(ClassificationRecord)
        .where(ClassificationRecord.user_id == user.id)
        .order_by(ClassificationRecord.achieved_at.desc())
    )
    return result.scalars().all()


@router.get("/current", response_model=list[CurrentClassificationOut])
async def current_classifications(
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    result = await db.execute(
        select(ClassificationRecord)
        .where(ClassificationRecord.user_id == user.id)
        .order_by(ClassificationRecord.achieved_at.desc())
    )
    records = result.scalars().all()

    # Get highest classification per system
    best_by_system: dict[str, ClassificationRecord] = {}
    for r in records:
        key = f"{r.system}:{r.round_type}"
        if key not in best_by_system:
            best_by_system[key] = r

    return [
        CurrentClassificationOut(
            system=r.system,
            classification=r.classification,
            round_type=r.round_type,
            score=r.score,
            achieved_at=r.achieved_at,
        )
        for r in best_by_system.values()
    ]
