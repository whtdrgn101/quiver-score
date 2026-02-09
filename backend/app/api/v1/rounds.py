from fastapi import APIRouter, Depends
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from app.database import get_db
from app.models.round_template import RoundTemplate
from app.schemas.scoring import RoundTemplateOut

router = APIRouter(prefix="/rounds", tags=["rounds"])


@router.get("", response_model=list[RoundTemplateOut])
async def list_rounds(db: AsyncSession = Depends(get_db)):
    result = await db.execute(select(RoundTemplate).order_by(RoundTemplate.name))
    return result.scalars().all()


@router.get("/{round_id}", response_model=RoundTemplateOut)
async def get_round(round_id: str, db: AsyncSession = Depends(get_db)):
    from uuid import UUID
    from app.core.exceptions import NotFoundError
    result = await db.execute(select(RoundTemplate).where(RoundTemplate.id == UUID(round_id)))
    template = result.scalar_one_or_none()
    if not template:
        raise NotFoundError("Round template not found")
    return template
