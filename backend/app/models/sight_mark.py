import uuid
from datetime import datetime

from sqlalchemy import String, Text, DateTime, ForeignKey, func
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.orm import Mapped, mapped_column

from app.database import Base


class SightMark(Base):
    __tablename__ = "sight_marks"

    id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    user_id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), ForeignKey("users.id", ondelete="CASCADE"), nullable=False, index=True)
    equipment_id: Mapped[uuid.UUID | None] = mapped_column(UUID(as_uuid=True), ForeignKey("equipment.id", ondelete="SET NULL"), index=True)
    setup_id: Mapped[uuid.UUID | None] = mapped_column(UUID(as_uuid=True), ForeignKey("setup_profiles.id", ondelete="SET NULL"), index=True)
    distance: Mapped[str] = mapped_column(String(50), nullable=False)  # "18m", "20yd"
    setting: Mapped[str] = mapped_column(String(100), nullable=False)  # "3.5 turns" or "47mm"
    notes: Mapped[str | None] = mapped_column(Text)
    date_recorded: Mapped[datetime] = mapped_column(DateTime(timezone=True), nullable=False)
    created_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), server_default=func.now())
