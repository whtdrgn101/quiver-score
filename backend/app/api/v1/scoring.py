from datetime import datetime, timezone
from uuid import UUID

from fastapi import APIRouter, Depends, status
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from app.database import get_db
from app.dependencies import get_current_user
from app.models.user import User
from app.models.round_template import RoundTemplateStage
from app.models.scoring import ScoringSession, End, Arrow
from app.models.setup_profile import SetupProfile
from app.schemas.scoring import (
    SessionCreate, SessionOut, SessionSummary, EndIn, EndOut, SessionComplete,
    StatsOut, RoundTypeAvg, RecentTrendItem,
)
from app.core.exceptions import NotFoundError, ValidationError

router = APIRouter(prefix="/sessions", tags=["scoring"])


def _session_out(session: ScoringSession) -> SessionOut:
    out = SessionOut.model_validate(session)
    if session.setup_profile:
        out.setup_profile_name = session.setup_profile.name
    return out


@router.post("", response_model=SessionOut, status_code=status.HTTP_201_CREATED)
async def create_session(
    body: SessionCreate,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    if body.setup_profile_id:
        result = await db.execute(
            select(SetupProfile).where(
                SetupProfile.id == body.setup_profile_id,
                SetupProfile.user_id == user.id,
            )
        )
        if not result.scalar_one_or_none():
            raise NotFoundError("Setup profile not found")

    session = ScoringSession(
        user_id=user.id,
        template_id=body.template_id,
        setup_profile_id=body.setup_profile_id,
        notes=body.notes,
        location=body.location,
        weather=body.weather,
    )
    db.add(session)
    await db.commit()
    await db.refresh(session)
    return _session_out(session)


@router.get("", response_model=list[SessionSummary])
async def list_sessions(
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    result = await db.execute(
        select(ScoringSession)
        .where(ScoringSession.user_id == user.id)
        .order_by(ScoringSession.started_at.desc())
    )
    sessions = result.scalars().all()
    out = []
    for s in sessions:
        summary = SessionSummary.model_validate(s)
        if s.template:
            summary.template_name = s.template.name
        if s.setup_profile:
            summary.setup_profile_name = s.setup_profile.name
        out.append(summary)
    return out


@router.get("/stats", response_model=StatsOut)
async def session_stats(
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    result = await db.execute(
        select(ScoringSession)
        .where(ScoringSession.user_id == user.id)
        .order_by(ScoringSession.completed_at.desc())
    )
    sessions = result.scalars().all()

    completed = [s for s in sessions if s.status == "completed"]
    total_arrows = sum(s.total_arrows for s in sessions)
    total_x_count = sum(s.total_x_count for s in sessions)

    # Personal best
    best = max(completed, key=lambda s: s.total_score, default=None)

    # Avg by round type
    from collections import defaultdict
    by_type: dict[str, list[int]] = defaultdict(list)
    for s in completed:
        name = s.template.name if s.template else "Unknown"
        by_type[name].append(s.total_score)
    avg_by_round = [
        RoundTypeAvg(template_name=name, avg_score=round(sum(scores) / len(scores), 1), count=len(scores))
        for name, scores in by_type.items()
    ]

    # Recent trend (last 10 completed)
    recent = []
    for s in completed[:10]:
        template = s.template
        if template and template.stages:
            max_score = sum(
                st.num_ends * st.arrows_per_end * st.max_score_per_arrow
                for st in template.stages
            )
        else:
            max_score = 0
        recent.append(RecentTrendItem(
            score=s.total_score,
            max_score=max_score,
            template_name=template.name if template else "Unknown",
            date=s.completed_at or s.started_at,
        ))

    return StatsOut(
        total_sessions=len(sessions),
        completed_sessions=len(completed),
        total_arrows=total_arrows,
        total_x_count=total_x_count,
        personal_best_score=best.total_score if best else None,
        personal_best_template=best.template.name if best and best.template else None,
        avg_by_round_type=avg_by_round,
        recent_trend=recent,
    )


@router.get("/{session_id}", response_model=SessionOut)
async def get_session(
    session_id: UUID,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    result = await db.execute(
        select(ScoringSession).where(ScoringSession.id == session_id, ScoringSession.user_id == user.id)
    )
    session = result.scalar_one_or_none()
    if not session:
        raise NotFoundError("Session not found")
    return _session_out(session)


@router.post("/{session_id}/ends", response_model=EndOut, status_code=status.HTTP_201_CREATED)
async def submit_end(
    session_id: UUID,
    body: EndIn,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    # Load session
    result = await db.execute(
        select(ScoringSession).where(ScoringSession.id == session_id, ScoringSession.user_id == user.id)
    )
    session = result.scalar_one_or_none()
    if not session:
        raise NotFoundError("Session not found")
    if session.status != "in_progress":
        raise ValidationError("Session is not in progress")

    # Load stage for validation
    result = await db.execute(select(RoundTemplateStage).where(RoundTemplateStage.id == body.stage_id))
    stage = result.scalar_one_or_none()
    if not stage:
        raise NotFoundError("Stage not found")

    # Validate arrow count
    if len(body.arrows) != stage.arrows_per_end:
        raise ValidationError(f"Expected {stage.arrows_per_end} arrows, got {len(body.arrows)}")

    # Validate arrow values
    allowed = set(stage.allowed_values)
    for a in body.arrows:
        if a.score_value not in allowed:
            raise ValidationError(f"Invalid arrow value '{a.score_value}'. Allowed: {stage.allowed_values}")

    # Determine end number
    end_number = len(session.ends) + 1

    # Create end
    end = End(session_id=session.id, stage_id=body.stage_id, end_number=end_number)
    db.add(end)
    await db.flush()

    # Create arrows
    end_total = 0
    x_count = 0
    value_map = stage.value_score_map
    for i, a in enumerate(body.arrows, 1):
        numeric = value_map[a.score_value]
        arrow = Arrow(
            end_id=end.id,
            arrow_number=i,
            score_value=a.score_value,
            score_numeric=numeric,
            x_pos=a.x_pos,
            y_pos=a.y_pos,
        )
        db.add(arrow)
        end_total += numeric
        if a.score_value == "X":
            x_count += 1

    end.end_total = end_total

    # Update session totals
    session.total_score += end_total
    session.total_x_count += x_count
    session.total_arrows += len(body.arrows)

    await db.commit()
    await db.refresh(end)
    return end


@router.post("/{session_id}/complete", response_model=SessionOut)
async def complete_session(
    session_id: UUID,
    body: SessionComplete | None = None,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    result = await db.execute(
        select(ScoringSession).where(ScoringSession.id == session_id, ScoringSession.user_id == user.id)
    )
    session = result.scalar_one_or_none()
    if not session:
        raise NotFoundError("Session not found")

    session.status = "completed"
    session.completed_at = datetime.now(timezone.utc)
    if body and body.notes:
        session.notes = body.notes

    await db.commit()
    await db.refresh(session)
    return _session_out(session)
