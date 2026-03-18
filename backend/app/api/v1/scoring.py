"""
Scoring session export endpoints — PDF and CSV.

This is the only Python route still active. All other session operations
are handled by the Go API, which proxies PDF export requests here.
"""

from uuid import UUID

from fastapi import APIRouter, Depends, Query
from fastapi.responses import Response
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from app.database import get_db
from app.dependencies import get_current_user
from app.models.user import User
from app.models.scoring import ScoringSession
from app.core.exceptions import NotFoundError

router = APIRouter(prefix="/sessions", tags=["scoring"])


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
