from uuid import UUID

from fastapi import APIRouter, Depends, status
from sqlalchemy import or_, select
from sqlalchemy.ext.asyncio import AsyncSession

from app.database import get_db
from app.dependencies import get_current_user, get_current_user_optional
from app.models.user import User
from app.models.round_template import RoundTemplate, RoundTemplateStage
from app.schemas.scoring import RoundTemplateOut, RoundTemplateCreate
from app.core.exceptions import NotFoundError, ForbiddenError

router = APIRouter(prefix="/rounds", tags=["rounds"])


@router.get("", response_model=list[RoundTemplateOut])
async def list_rounds(
    db: AsyncSession = Depends(get_db),
    user: User | None = Depends(get_current_user_optional),
):
    query = select(RoundTemplate)
    if user:
        query = query.where(
            or_(RoundTemplate.is_official == True, RoundTemplate.created_by == user.id)
        )
    else:
        query = query.where(RoundTemplate.is_official == True)
    result = await db.execute(query.order_by(RoundTemplate.name))
    return result.scalars().all()


@router.post("", response_model=RoundTemplateOut, status_code=status.HTTP_201_CREATED)
async def create_round(
    body: RoundTemplateCreate,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    template = RoundTemplate(
        name=body.name,
        organization=body.organization,
        description=body.description,
        is_official=False,
        created_by=user.id,
    )
    db.add(template)
    await db.flush()

    for i, stage_data in enumerate(body.stages, 1):
        stage = RoundTemplateStage(
            template_id=template.id,
            stage_order=i,
            name=stage_data.name,
            distance=stage_data.distance,
            num_ends=stage_data.num_ends,
            arrows_per_end=stage_data.arrows_per_end,
            allowed_values=stage_data.allowed_values,
            value_score_map=stage_data.value_score_map,
            max_score_per_arrow=stage_data.max_score_per_arrow,
        )
        db.add(stage)

    await db.commit()
    await db.refresh(template)
    return template


@router.delete("/{round_id}", status_code=status.HTTP_204_NO_CONTENT)
async def delete_round(
    round_id: UUID,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    result = await db.execute(select(RoundTemplate).where(RoundTemplate.id == round_id))
    template = result.scalar_one_or_none()
    if not template:
        raise NotFoundError("Round template not found")
    if template.is_official:
        raise ForbiddenError("Cannot delete official round templates")
    if template.created_by != user.id:
        raise ForbiddenError("You can only delete your own custom rounds")
    for stage in template.stages:
        await db.delete(stage)
    await db.delete(template)
    await db.commit()


@router.get("/{round_id}", response_model=RoundTemplateOut)
async def get_round(round_id: UUID, db: AsyncSession = Depends(get_db)):
    result = await db.execute(select(RoundTemplate).where(RoundTemplate.id == round_id))
    template = result.scalar_one_or_none()
    if not template:
        raise NotFoundError("Round template not found")
    return template
