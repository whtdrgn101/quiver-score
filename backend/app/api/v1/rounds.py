from uuid import UUID

from fastapi import APIRouter, Depends, status
from sqlalchemy import or_, select
from sqlalchemy.ext.asyncio import AsyncSession

from app.database import get_db
from app.dependencies import get_current_user, get_current_user_optional
from app.models.user import User
from app.models.round_template import RoundTemplate, RoundTemplateStage
from app.models.club import ClubMember, ClubSharedRound
from app.models.scoring import ScoringSession
from app.schemas.scoring import RoundTemplateOut, RoundTemplateCreate
from app.schemas.club import ShareRoundToClub, ClubSharedRoundOut
from app.core.exceptions import AuthError, ConflictError, ForbiddenError, NotFoundError, ValidationError

router = APIRouter(prefix="/rounds", tags=["rounds"])


@router.get("", response_model=list[RoundTemplateOut])
async def list_rounds(
    db: AsyncSession = Depends(get_db),
    user: User | None = Depends(get_current_user_optional),
):
    query = select(RoundTemplate)
    if user:
        # Get IDs of clubs the user belongs to
        club_ids_result = await db.execute(
            select(ClubMember.club_id).where(ClubMember.user_id == user.id)
        )
        club_ids = list(club_ids_result.scalars().all())

        # Get template IDs shared with those clubs
        shared_template_ids: list[UUID] = []
        if club_ids:
            shared_result = await db.execute(
                select(ClubSharedRound.template_id).where(ClubSharedRound.club_id.in_(club_ids))
            )
            shared_template_ids = list(shared_result.scalars().all())

        conditions = [
            RoundTemplate.is_official == True,
            RoundTemplate.created_by == user.id,
        ]
        if shared_template_ids:
            conditions.append(RoundTemplate.id.in_(shared_template_ids))
        query = query.where(or_(*conditions))
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


@router.put("/{round_id}", response_model=RoundTemplateOut)
async def update_round(
    round_id: UUID,
    body: RoundTemplateCreate,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    result = await db.execute(select(RoundTemplate).where(RoundTemplate.id == round_id))
    template = result.scalar_one_or_none()
    if not template:
        raise NotFoundError("Round template not found")
    if template.is_official:
        raise ForbiddenError("Cannot edit official round templates")
    if template.created_by != user.id:
        raise ForbiddenError("You can only edit your own custom rounds")

    # Block editing if there's an in-progress session using this template
    in_progress = await db.execute(
        select(ScoringSession.id).where(
            ScoringSession.template_id == round_id,
            ScoringSession.status == "in_progress",
        ).limit(1)
    )
    if in_progress.scalar_one_or_none():
        raise ValidationError("Cannot edit a round template while a scoring session is in progress")

    # Update template metadata
    template.name = body.name
    template.organization = body.organization
    template.description = body.description

    # Delete old stages (End.stage_id will be SET NULL for historical ends)
    for stage in template.stages:
        await db.delete(stage)
    await db.flush()

    # Create new stages
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


@router.post("/{round_id}/share", status_code=status.HTTP_201_CREATED)
async def share_round_with_club(
    round_id: UUID,
    body: ShareRoundToClub,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    result = await db.execute(select(RoundTemplate).where(RoundTemplate.id == round_id))
    template = result.scalar_one_or_none()
    if not template:
        raise NotFoundError("Round template not found")
    if template.is_official:
        raise ForbiddenError("Cannot share official round templates")
    if template.created_by != user.id:
        raise ForbiddenError("You can only share your own custom rounds")

    # Check user is a member of the club
    member_result = await db.execute(
        select(ClubMember).where(ClubMember.club_id == body.club_id, ClubMember.user_id == user.id)
    )
    if not member_result.scalar_one_or_none():
        raise AuthError("You are not a member of this club")

    # Check duplicate
    existing = await db.execute(
        select(ClubSharedRound).where(
            ClubSharedRound.club_id == body.club_id,
            ClubSharedRound.template_id == round_id,
        )
    )
    if existing.scalar_one_or_none():
        raise ConflictError("Round is already shared with this club")

    share = ClubSharedRound(
        club_id=body.club_id,
        template_id=round_id,
        shared_by=user.id,
    )
    db.add(share)
    await db.commit()
    return {"detail": "Round shared with club"}


@router.delete("/{round_id}/share/{club_id}", status_code=status.HTTP_204_NO_CONTENT)
async def unshare_round_from_club(
    round_id: UUID,
    club_id: UUID,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    result = await db.execute(select(RoundTemplate).where(RoundTemplate.id == round_id))
    template = result.scalar_one_or_none()
    if not template:
        raise NotFoundError("Round template not found")
    if template.created_by != user.id:
        raise ForbiddenError("You can only unshare your own custom rounds")

    share_result = await db.execute(
        select(ClubSharedRound).where(
            ClubSharedRound.club_id == club_id,
            ClubSharedRound.template_id == round_id,
        )
    )
    share = share_result.scalar_one_or_none()
    if not share:
        raise NotFoundError("Share not found")
    await db.delete(share)
    await db.commit()
