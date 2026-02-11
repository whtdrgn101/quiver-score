import uuid
from datetime import datetime

from sqlalchemy import Boolean, String, Text, DateTime, func
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.orm import Mapped, mapped_column, relationship

from app.database import Base


class User(Base):
    __tablename__ = "users"

    id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    email: Mapped[str] = mapped_column(String(255), unique=True, nullable=False, index=True)
    username: Mapped[str] = mapped_column(String(50), unique=True, nullable=False, index=True)
    hashed_password: Mapped[str] = mapped_column(String(255), nullable=False)
    display_name: Mapped[str | None] = mapped_column(String(100))
    bow_type: Mapped[str | None] = mapped_column(String(50))
    classification: Mapped[str | None] = mapped_column(String(50))
    bio: Mapped[str | None] = mapped_column(Text)
    avatar: Mapped[str | None] = mapped_column(Text)
    email_verified: Mapped[bool] = mapped_column(Boolean, server_default="false", nullable=False)
    email_verification_token: Mapped[str | None] = mapped_column(String(255), nullable=True)
    profile_public: Mapped[bool] = mapped_column(Boolean, server_default="false", nullable=False)
    created_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), server_default=func.now())
    updated_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now())

    scoring_sessions: Mapped[list["ScoringSession"]] = relationship(back_populates="user", lazy="selectin")
    equipment: Mapped[list["Equipment"]] = relationship(back_populates="user", lazy="noload")
    setup_profiles: Mapped[list["SetupProfile"]] = relationship(back_populates="user", lazy="noload")
