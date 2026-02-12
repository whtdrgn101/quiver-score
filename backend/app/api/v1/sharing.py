import secrets
import uuid

from fastapi import APIRouter, Depends
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.orm import selectinload

from app.database import get_db
from app.dependencies import get_current_user
from app.models.user import User
from app.models.scoring import ScoringSession
from app.schemas.scoring import ShareLinkOut, SharedSessionOut, EndOut
from app.core.exceptions import NotFoundError, AuthError
from app.config import settings

router = APIRouter(prefix="/share", tags=["sharing"])


@router.post("/sessions/{session_id}", response_model=ShareLinkOut)
async def create_share_link(
    session_id: uuid.UUID,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    result = await db.execute(
        select(ScoringSession).where(
            ScoringSession.id == session_id,
            ScoringSession.user_id == user.id,
        )
    )
    session = result.scalar_one_or_none()
    if not session:
        raise NotFoundError("Session not found")

    if not session.share_token:
        session.share_token = secrets.token_urlsafe(16)
        await db.commit()
        await db.refresh(session)

    return ShareLinkOut(
        share_token=session.share_token,
        url=f"{settings.FRONTEND_URL}/shared/{session.share_token}",
    )


@router.delete("/sessions/{session_id}")
async def revoke_share_link(
    session_id: uuid.UUID,
    user: User = Depends(get_current_user),
    db: AsyncSession = Depends(get_db),
):
    result = await db.execute(
        select(ScoringSession).where(
            ScoringSession.id == session_id,
            ScoringSession.user_id == user.id,
        )
    )
    session = result.scalar_one_or_none()
    if not session:
        raise NotFoundError("Session not found")

    session.share_token = None
    await db.commit()
    return {"detail": "Share link revoked"}


@router.get("/s/{token}", response_model=SharedSessionOut)
async def get_shared_session(
    token: str,
    db: AsyncSession = Depends(get_db),
):
    result = await db.execute(
        select(ScoringSession)
        .where(ScoringSession.share_token == token)
        .options(selectinload(ScoringSession.user))
    )
    session = result.scalar_one_or_none()
    if not session:
        raise NotFoundError("Shared session not found")

    return SharedSessionOut(
        archer_name=session.user.display_name or session.user.username,
        archer_avatar=session.user.avatar,
        template_name=session.template.name if session.template else "Unknown",
        template=session.template,
        total_score=session.total_score,
        total_x_count=session.total_x_count,
        total_arrows=session.total_arrows,
        notes=session.notes,
        location=session.location,
        weather=session.weather,
        started_at=session.started_at,
        completed_at=session.completed_at,
        ends=[EndOut.model_validate(e) for e in session.ends],
    )
