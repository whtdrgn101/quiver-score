from datetime import date, datetime, timezone
from uuid import UUID

from fastapi import APIRouter, Depends, Query, status
from fastapi.responses import Response
from sqlalchemy import or_, select
from sqlalchemy.ext.asyncio import AsyncSession

from app.database import get_db
from app.dependencies import get_current_user
from app.models.user import User
from app.models.round_template import RoundTemplateStage
from app.models.scoring import ScoringSession, End, Arrow, PersonalRecord
from app.models.setup_profile import SetupProfile
from app.schemas.scoring import (
    SessionCreate, SessionOut, SessionSummary, EndIn, EndOut, SessionComplete,
    StatsOut, RoundTypeAvg, RecentTrendItem, PersonalRecordOut, TrendDataItem,
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
    template_id: UUID | None = Query(None),
    date_from: date | None = Query(None),
    date_to: date | None = Query(None),
    search: str | None = Query(None),
):
    query = select(ScoringSession).where(ScoringSession.user_id == user.id)

    if template_id:
        query = query.where(ScoringSession.template_id == template_id)
    if date_from:
        query = query.where(ScoringSession.started_at >= datetime.combine(date_from, datetime.min.time(), tzinfo=timezone.utc))
    if date_to:
        query = query.where(ScoringSession.started_at <= datetime.combine(date_to, datetime.max.time(), tzinfo=timezone.utc))
    if search:
        pattern = f"%{search}%"
        query = query.where(
            or_(
                ScoringSession.notes.ilike(pattern),
                ScoringSession.location.ilike(pattern),
            )
        )

    result = await db.execute(query.order_by(ScoringSession.started_at.desc()))
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


@router.get("/export")
async def export_sessions(
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
    template_id: UUID | None = Query(None),
    date_from: date | None = Query(None),
    date_to: date | None = Query(None),
    search: str | None = Query(None),
):
    from app.services.export import generate_sessions_csv

    query = select(ScoringSession).where(ScoringSession.user_id == user.id)
    if template_id:
        query = query.where(ScoringSession.template_id == template_id)
    if date_from:
        query = query.where(ScoringSession.started_at >= datetime.combine(date_from, datetime.min.time(), tzinfo=timezone.utc))
    if date_to:
        query = query.where(ScoringSession.started_at <= datetime.combine(date_to, datetime.max.time(), tzinfo=timezone.utc))
    if search:
        pattern = f"%{search}%"
        query = query.where(or_(ScoringSession.notes.ilike(pattern), ScoringSession.location.ilike(pattern)))

    result = await db.execute(query.order_by(ScoringSession.started_at.desc()))
    sessions = result.scalars().all()
    csv_content = generate_sessions_csv(sessions)
    return Response(
        content=csv_content,
        media_type="text/csv",
        headers={"Content-Disposition": "attachment; filename=sessions.csv"},
    )


@router.get("/{session_id}/export")
async def export_session(
    session_id: UUID,
    format: str = Query("csv"),
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    from app.services.export import generate_session_csv, generate_session_pdf

    result = await db.execute(
        select(ScoringSession).where(ScoringSession.id == session_id, ScoringSession.user_id == user.id)
    )
    session = result.scalar_one_or_none()
    if not session:
        raise NotFoundError("Session not found")

    if format == "pdf":
        pdf_bytes = generate_session_pdf(session)
        return Response(
            content=pdf_bytes,
            media_type="application/pdf",
            headers={"Content-Disposition": f"attachment; filename=session-{session_id}.pdf"},
        )
    else:
        csv_content = generate_session_csv(session)
        return Response(
            content=csv_content,
            media_type="text/csv",
            headers={"Content-Disposition": f"attachment; filename=session-{session_id}.csv"},
        )


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

    # Personal records
    pr_result = await db.execute(
        select(PersonalRecord).where(PersonalRecord.user_id == user.id)
    )
    prs = pr_result.scalars().all()
    pr_list = []
    for pr in prs:
        template = pr.template
        if template and template.stages:
            pr_max = sum(
                st.num_ends * st.arrows_per_end * st.max_score_per_arrow
                for st in template.stages
            )
        else:
            pr_max = 0
        pr_list.append(PersonalRecordOut(
            template_name=template.name if template else "Unknown",
            score=pr.score,
            max_score=pr_max,
            achieved_at=pr.achieved_at,
            session_id=pr.session_id,
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
        personal_records=pr_list,
    )


@router.get("/personal-records", response_model=list[PersonalRecordOut])
async def list_personal_records(
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    result = await db.execute(
        select(PersonalRecord).where(PersonalRecord.user_id == user.id)
    )
    prs = result.scalars().all()
    out = []
    for pr in prs:
        template = pr.template
        if template and template.stages:
            pr_max = sum(
                st.num_ends * st.arrows_per_end * st.max_score_per_arrow
                for st in template.stages
            )
        else:
            pr_max = 0
        out.append(PersonalRecordOut(
            template_name=template.name if template else "Unknown",
            score=pr.score,
            max_score=pr_max,
            achieved_at=pr.achieved_at,
            session_id=pr.session_id,
        ))
    return out


@router.get("/trends", response_model=list[TrendDataItem])
async def session_trends(
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    result = await db.execute(
        select(ScoringSession)
        .where(ScoringSession.user_id == user.id, ScoringSession.status == "completed")
        .order_by(ScoringSession.completed_at.desc())
    )
    sessions = result.scalars().all()
    out = []
    for s in sessions:
        template = s.template
        if template and template.stages:
            max_score = sum(
                st.num_ends * st.arrows_per_end * st.max_score_per_arrow
                for st in template.stages
            )
        else:
            max_score = 0
        percentage = round((s.total_score / max_score) * 100, 1) if max_score > 0 else 0
        out.append(TrendDataItem(
            session_id=s.id,
            template_name=template.name if template else "Unknown",
            total_score=s.total_score,
            max_score=max_score,
            percentage=percentage,
            completed_at=s.completed_at or s.started_at,
        ))
    return out


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
    out = _session_out(session)
    # Check if this session holds a personal record
    pr_result = await db.execute(
        select(PersonalRecord).where(
            PersonalRecord.user_id == user.id,
            PersonalRecord.session_id == session.id,
        )
    )
    if pr_result.scalar_one_or_none():
        out.is_personal_best = True
    return out


@router.delete("/{session_id}/ends/last", response_model=SessionOut)
async def undo_last_end(
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
    if session.status != "in_progress":
        raise ValidationError("Session is not in progress")
    if not session.ends:
        raise ValidationError("No ends to undo")

    last_end = session.ends[-1]

    # Subtract totals
    x_count = sum(1 for a in last_end.arrows if a.score_value == "X")
    session.total_score -= last_end.end_total
    session.total_x_count -= x_count
    session.total_arrows -= len(last_end.arrows)

    # Delete arrows then end
    for arrow in last_end.arrows:
        await db.delete(arrow)
    await db.delete(last_end)

    await db.commit()
    await db.refresh(session)
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
    if body and body.location:
        session.location = body.location
    if body and body.weather:
        session.weather = body.weather

    # Check personal record
    is_personal_best = False
    result = await db.execute(
        select(PersonalRecord).where(
            PersonalRecord.user_id == user.id,
            PersonalRecord.template_id == session.template_id,
        )
    )
    existing_pr = result.scalar_one_or_none()
    if existing_pr is None or session.total_score > existing_pr.score:
        if existing_pr:
            existing_pr.session_id = session.id
            existing_pr.score = session.total_score
            existing_pr.achieved_at = session.completed_at
        else:
            pr = PersonalRecord(
                user_id=user.id,
                template_id=session.template_id,
                session_id=session.id,
                score=session.total_score,
                achieved_at=session.completed_at,
            )
            db.add(pr)
        is_personal_best = True

    # Create notification for personal record
    if is_personal_best:
        from app.services.notifications import create_notification
        template_name = session.template.name if session.template else "Unknown"
        await create_notification(
            db, user.id,
            type="personal_record",
            title="New Personal Record!",
            message=f"You scored {session.total_score} on {template_name} — a new personal best!",
            link=f"/sessions/{session.id}",
        )

    # Check classification
    from app.services.classification import calculate_classification
    from app.models.classification import ClassificationRecord
    template_name_for_class = session.template.name if session.template else None
    if template_name_for_class:
        result_class = calculate_classification(session.total_score, template_name_for_class)
        if result_class:
            system, classification = result_class
            cr = ClassificationRecord(
                user_id=user.id,
                system=system,
                classification=classification,
                round_type=template_name_for_class,
                score=session.total_score,
                achieved_at=session.completed_at,
                session_id=session.id,
            )
            db.add(cr)

    # Create feed item for completed session
    from app.services.feed import create_feed_item
    template_name_feed = session.template.name if session.template else "Unknown"
    feed_type = "personal_record" if is_personal_best else "session_completed"
    await create_feed_item(db, user.id, type=feed_type, data={
        "template_name": template_name_feed,
        "total_score": session.total_score,
        "session_id": str(session.id),
    })

    await db.commit()
    await db.refresh(session)
    out = _session_out(session)
    out.is_personal_best = is_personal_best
    return out


@router.post("/{session_id}/abandon", status_code=status.HTTP_200_OK)
async def abandon_session(
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
    if session.status != "in_progress":
        raise ValidationError("Only in-progress sessions can be abandoned")
    session.status = "abandoned"
    await db.commit()
    return {"detail": "Session abandoned"}


@router.delete("/{session_id}", status_code=status.HTTP_204_NO_CONTENT)
async def delete_session(
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
    if session.status != "abandoned":
        raise ValidationError("Only abandoned sessions can be deleted")
    for end in session.ends:
        for arrow in end.arrows:
            await db.delete(arrow)
        await db.delete(end)
    await db.delete(session)
    await db.commit()
