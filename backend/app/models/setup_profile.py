import uuid
from datetime import datetime

from sqlalchemy import String, Float, Text, DateTime, ForeignKey, func
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.orm import Mapped, mapped_column, relationship

from app.database import Base


class SetupProfile(Base):
    __tablename__ = "setup_profiles"

    id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    user_id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), ForeignKey("users.id"), nullable=False)
    name: Mapped[str] = mapped_column(String(200), nullable=False)
    description: Mapped[str | None] = mapped_column(Text)
    brace_height: Mapped[float | None] = mapped_column(Float)
    tiller: Mapped[float | None] = mapped_column(Float)
    draw_weight: Mapped[float | None] = mapped_column(Float)
    draw_length: Mapped[float | None] = mapped_column(Float)
    arrow_foc: Mapped[float | None] = mapped_column(Float)
    created_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), server_default=func.now())

    user: Mapped["User"] = relationship(back_populates="setup_profiles")
    equipment_links: Mapped[list["SetupEquipment"]] = relationship(
        back_populates="setup", lazy="selectin", cascade="all, delete-orphan"
    )


class SetupEquipment(Base):
    __tablename__ = "setup_equipment"

    id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    setup_id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True), ForeignKey("setup_profiles.id", ondelete="CASCADE"), nullable=False
    )
    equipment_id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True), ForeignKey("equipment.id", ondelete="CASCADE"), nullable=False
    )

    setup: Mapped["SetupProfile"] = relationship(back_populates="equipment_links")
    equipment: Mapped["Equipment"] = relationship(lazy="selectin")
