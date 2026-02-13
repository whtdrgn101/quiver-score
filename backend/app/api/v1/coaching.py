from uuid import UUID

from fastapi import APIRouter, Depends, status
from sqlalchemy import select, or_
from sqlalchemy.ext.asyncio import AsyncSession

from app.database import get_db
from app.dependencies import get_current_user
from app.models.user import User
from app.models.coaching import CoachAthleteLink, SessionAnnotation
from app.models.scoring import ScoringSession
from app.schemas.coaching import (
    CoachAthleteLinkOut,
    InviteRequest,
    RespondRequest,
    AnnotationCreate,
    AnnotationOut,
)
from app.core.exceptions import NotFoundError, ConflictError, ForbiddenError, ValidationError

router = APIRouter(prefix="/coaching", tags=["coaching"])


def _link_to_out(link: CoachAthleteLink) -> dict:
    return {
        "id": link.id,
        "coach_id": link.coach_id,
        "athlete_id": link.athlete_id,
        "coach_username": link.coach.username if link.coach else None,
        "athlete_username": link.athlete.username if link.athlete else None,
        "status": link.status,
        "created_at": link.created_at,
    }


@router.post("/invite", response_model=CoachAthleteLinkOut, status_code=status.HTTP_201_CREATED)
async def invite_athlete(
    body: InviteRequest,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    """Coach invites an athlete by username."""
    result = await db.execute(select(User).where(User.username == body.athlete_username))
    athlete = result.scalar_one_or_none()
    if not athlete:
        raise NotFoundError("User not found")
    if athlete.id == user.id:
        raise ValidationError("Cannot coach yourself")

    # Check existing link
    existing = await db.execute(
        select(CoachAthleteLink).where(
            CoachAthleteLink.coach_id == user.id,
            CoachAthleteLink.athlete_id == athlete.id,
        )
    )
    if existing.scalar_one_or_none():
        raise ConflictError("Link already exists")

    link = CoachAthleteLink(coach_id=user.id, athlete_id=athlete.id)
    db.add(link)
    await db.commit()
    await db.refresh(link)
    return CoachAthleteLinkOut.model_validate(_link_to_out(link))


@router.post("/respond", response_model=CoachAthleteLinkOut)
async def respond_to_invite(
    body: RespondRequest,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    """Athlete accepts or rejects a coaching invite."""
    result = await db.execute(
        select(CoachAthleteLink).where(
            CoachAthleteLink.id == body.link_id,
            CoachAthleteLink.athlete_id == user.id,
        )
    )
    link = result.scalar_one_or_none()
    if not link:
        raise NotFoundError("Invite not found")
    if link.status != "pending":
        raise ValidationError("Invite already responded to")

    link.status = "active" if body.accept else "revoked"
    await db.commit()
    await db.refresh(link)
    return CoachAthleteLinkOut.model_validate(_link_to_out(link))


@router.get("/athletes", response_model=list[CoachAthleteLinkOut])
async def list_athletes(
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    """List coach's athletes."""
    result = await db.execute(
        select(CoachAthleteLink)
        .where(CoachAthleteLink.coach_id == user.id)
        .order_by(CoachAthleteLink.created_at.desc())
    )
    return [CoachAthleteLinkOut.model_validate(_link_to_out(l)) for l in result.scalars().all()]


@router.get("/coaches", response_model=list[CoachAthleteLinkOut])
async def list_coaches(
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    """List athlete's coaches."""
    result = await db.execute(
        select(CoachAthleteLink)
        .where(CoachAthleteLink.athlete_id == user.id)
        .order_by(CoachAthleteLink.created_at.desc())
    )
    return [CoachAthleteLinkOut.model_validate(_link_to_out(l)) for l in result.scalars().all()]


@router.get("/athletes/{athlete_id}/sessions")
async def view_athlete_sessions(
    athlete_id: UUID,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    """Coach views athlete's completed sessions."""
    # Verify active coaching link
    result = await db.execute(
        select(CoachAthleteLink).where(
            CoachAthleteLink.coach_id == user.id,
            CoachAthleteLink.athlete_id == athlete_id,
            CoachAthleteLink.status == "active",
        )
    )
    if not result.scalar_one_or_none():
        raise ForbiddenError("No active coaching link")

    sessions = await db.execute(
        select(ScoringSession)
        .where(ScoringSession.user_id == athlete_id, ScoringSession.status == "completed")
        .order_by(ScoringSession.completed_at.desc())
    )
    return [
        {
            "id": s.id,
            "template_name": s.template.name if s.template else "Unknown",
            "total_score": s.total_score,
            "total_x_count": s.total_x_count,
            "total_arrows": s.total_arrows,
            "completed_at": s.completed_at,
        }
        for s in sessions.scalars().all()
    ]


@router.post("/sessions/{session_id}/annotations", response_model=AnnotationOut, status_code=status.HTTP_201_CREATED)
async def add_annotation(
    session_id: UUID,
    body: AnnotationCreate,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    """Add annotation to a session (coach or session owner)."""
    # Load session
    session_result = await db.execute(
        select(ScoringSession).where(ScoringSession.id == session_id)
    )
    session = session_result.scalar_one_or_none()
    if not session:
        raise NotFoundError("Session not found")

    # Must be session owner or active coach
    if session.user_id != user.id:
        link_result = await db.execute(
            select(CoachAthleteLink).where(
                CoachAthleteLink.coach_id == user.id,
                CoachAthleteLink.athlete_id == session.user_id,
                CoachAthleteLink.status == "active",
            )
        )
        if not link_result.scalar_one_or_none():
            raise ForbiddenError("Not authorized to annotate this session")

    annotation = SessionAnnotation(
        session_id=session_id,
        author_id=user.id,
        end_number=body.end_number,
        arrow_number=body.arrow_number,
        text=body.text,
    )
    db.add(annotation)
    await db.commit()
    await db.refresh(annotation)
    return AnnotationOut(
        id=annotation.id,
        session_id=annotation.session_id,
        author_id=annotation.author_id,
        author_username=user.username,
        end_number=annotation.end_number,
        arrow_number=annotation.arrow_number,
        text=annotation.text,
        created_at=annotation.created_at,
    )


@router.get("/sessions/{session_id}/annotations", response_model=list[AnnotationOut])
async def list_annotations(
    session_id: UUID,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    """List annotations for a session (session owner or coach)."""
    session_result = await db.execute(
        select(ScoringSession).where(ScoringSession.id == session_id)
    )
    session = session_result.scalar_one_or_none()
    if not session:
        raise NotFoundError("Session not found")

    # Must be session owner or active coach
    if session.user_id != user.id:
        link_result = await db.execute(
            select(CoachAthleteLink).where(
                CoachAthleteLink.coach_id == user.id,
                CoachAthleteLink.athlete_id == session.user_id,
                CoachAthleteLink.status == "active",
            )
        )
        if not link_result.scalar_one_or_none():
            raise ForbiddenError("Not authorized to view annotations")

    result = await db.execute(
        select(SessionAnnotation)
        .where(SessionAnnotation.session_id == session_id)
        .order_by(SessionAnnotation.created_at)
    )
    return [
        AnnotationOut(
            id=a.id,
            session_id=a.session_id,
            author_id=a.author_id,
            author_username=a.author.username if a.author else None,
            end_number=a.end_number,
            arrow_number=a.arrow_number,
            text=a.text,
            created_at=a.created_at,
        )
        for a in result.scalars().all()
    ]
