import uuid
from datetime import datetime

from typing import TYPE_CHECKING

from sqlalchemy import CheckConstraint, DateTime, ForeignKey, Integer, String, Text, UniqueConstraint, func
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.orm import Mapped, mapped_column, relationship

if TYPE_CHECKING:
    from app.models.club import Club

from app.database import Base


class Tournament(Base):
    __tablename__ = "tournaments"
    __table_args__ = (
        CheckConstraint(
            "status IN ('draft', 'registration', 'in_progress', 'completed')",
            name="ck_tournament_status",
        ),
    )

    id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    name: Mapped[str] = mapped_column(String(200), nullable=False)
    description: Mapped[str | None] = mapped_column(Text)
    organizer_id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), ForeignKey("users.id"), nullable=False, index=True)
    club_id: Mapped[uuid.UUID | None] = mapped_column(UUID(as_uuid=True), ForeignKey("clubs.id", ondelete="CASCADE"), index=True)
    template_id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), ForeignKey("round_templates.id"), nullable=False, index=True)
    status: Mapped[str] = mapped_column(String(20), default="draft")
    max_participants: Mapped[int | None] = mapped_column(Integer)
    registration_deadline: Mapped[datetime | None] = mapped_column(DateTime(timezone=True))
    start_date: Mapped[datetime | None] = mapped_column(DateTime(timezone=True))
    end_date: Mapped[datetime | None] = mapped_column(DateTime(timezone=True))
    created_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), server_default=func.now())

    organizer: Mapped["User"] = relationship(lazy="selectin")
    club: Mapped["Club"] = relationship(lazy="selectin")
    template: Mapped["RoundTemplate"] = relationship(lazy="selectin")
    participants: Mapped[list["TournamentParticipant"]] = relationship(
        back_populates="tournament", lazy="selectin", cascade="all, delete-orphan"
    )


class TournamentParticipant(Base):
    __tablename__ = "tournament_participants"
    __table_args__ = (
        UniqueConstraint("tournament_id", "user_id", name="uq_tournament_user"),
        CheckConstraint(
            "status IN ('registered', 'active', 'completed', 'withdrawn')",
            name="ck_tournament_participant_status",
        ),
    )

    id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    tournament_id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True), ForeignKey("tournaments.id", ondelete="CASCADE"), nullable=False, index=True
    )
    user_id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), ForeignKey("users.id"), nullable=False, index=True)
    session_id: Mapped[uuid.UUID | None] = mapped_column(
        UUID(as_uuid=True), ForeignKey("scoring_sessions.id", ondelete="SET NULL"), index=True
    )
    status: Mapped[str] = mapped_column(String(20), default="registered")
    final_score: Mapped[int | None] = mapped_column(Integer)
    final_x_count: Mapped[int | None] = mapped_column(Integer)
    rank: Mapped[int | None] = mapped_column(Integer)
    registered_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), server_default=func.now())

    tournament: Mapped["Tournament"] = relationship(back_populates="participants")
    user: Mapped["User"] = relationship(lazy="selectin")
