import uuid
from datetime import datetime

from sqlalchemy import String, Integer, Float, DateTime, ForeignKey, UniqueConstraint, func, Text
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.orm import Mapped, mapped_column, relationship

from app.database import Base


class ScoringSession(Base):
    __tablename__ = "scoring_sessions"

    id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    user_id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), ForeignKey("users.id"), nullable=False)
    template_id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), ForeignKey("round_templates.id"), nullable=False)
    status: Mapped[str] = mapped_column(String(20), default="in_progress")  # in_progress, completed, abandoned
    total_score: Mapped[int] = mapped_column(Integer, default=0)
    total_x_count: Mapped[int] = mapped_column(Integer, default=0)
    total_arrows: Mapped[int] = mapped_column(Integer, default=0)
    setup_profile_id: Mapped[uuid.UUID | None] = mapped_column(UUID(as_uuid=True), ForeignKey("setup_profiles.id"), nullable=True)
    notes: Mapped[str | None] = mapped_column(Text)
    location: Mapped[str | None] = mapped_column(String(200))
    weather: Mapped[str | None] = mapped_column(String(100))
    started_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), server_default=func.now())
    share_token: Mapped[str | None] = mapped_column(String(32), nullable=True, unique=True, index=True)
    completed_at: Mapped[datetime | None] = mapped_column(DateTime(timezone=True))

    user: Mapped["User"] = relationship(back_populates="scoring_sessions")
    template: Mapped["RoundTemplate"] = relationship(lazy="selectin")
    setup_profile: Mapped["SetupProfile | None"] = relationship(lazy="selectin")
    ends: Mapped[list["End"]] = relationship(back_populates="session", lazy="selectin", order_by="End.end_number")


class End(Base):
    __tablename__ = "ends"

    id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    session_id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), ForeignKey("scoring_sessions.id"), nullable=False)
    stage_id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), ForeignKey("round_template_stages.id"), nullable=False)
    end_number: Mapped[int] = mapped_column(Integer, nullable=False)
    end_total: Mapped[int] = mapped_column(Integer, default=0)
    created_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), server_default=func.now())

    session: Mapped["ScoringSession"] = relationship(back_populates="ends")
    stage: Mapped["RoundTemplateStage"] = relationship(lazy="selectin")
    arrows: Mapped[list["Arrow"]] = relationship(back_populates="end", lazy="selectin", order_by="Arrow.arrow_number")


class Arrow(Base):
    __tablename__ = "arrows"

    id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    end_id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), ForeignKey("ends.id"), nullable=False)
    arrow_number: Mapped[int] = mapped_column(Integer, nullable=False)
    score_value: Mapped[str] = mapped_column(String(5), nullable=False)  # "X", "10", "M", etc.
    score_numeric: Mapped[int] = mapped_column(Integer, nullable=False)  # numeric equivalent
    x_pos: Mapped[float | None] = mapped_column(Float)  # optional target position
    y_pos: Mapped[float | None] = mapped_column(Float)

    end: Mapped["End"] = relationship(back_populates="arrows")


class PersonalRecord(Base):
    __tablename__ = "personal_records"
    __table_args__ = (UniqueConstraint("user_id", "template_id", name="uq_user_template_pr"),)

    id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    user_id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), ForeignKey("users.id"), nullable=False)
    template_id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), ForeignKey("round_templates.id"), nullable=False)
    session_id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), ForeignKey("scoring_sessions.id"), nullable=False)
    score: Mapped[int] = mapped_column(Integer, nullable=False)
    achieved_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), nullable=False)

    template: Mapped["RoundTemplate"] = relationship(lazy="selectin")
    session: Mapped["ScoringSession"] = relationship(lazy="selectin")
