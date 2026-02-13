import uuid
from datetime import datetime

from sqlalchemy import String, Integer, DateTime, ForeignKey, func, Text
from sqlalchemy import JSON
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.orm import Mapped, mapped_column, relationship

from app.database import Base


class RoundTemplate(Base):
    __tablename__ = "round_templates"

    id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    name: Mapped[str] = mapped_column(String(100), nullable=False)
    organization: Mapped[str] = mapped_column(String(50), nullable=False)  # WA, NFAA, Lancaster, ASA, IBO
    description: Mapped[str | None] = mapped_column(Text)
    is_official: Mapped[bool] = mapped_column(default=True)
    created_by: Mapped[uuid.UUID | None] = mapped_column(UUID(as_uuid=True), ForeignKey("users.id"))
    created_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), server_default=func.now())

    stages: Mapped[list["RoundTemplateStage"]] = relationship(
        back_populates="template", lazy="selectin", order_by="RoundTemplateStage.stage_order"
    )


class RoundTemplateStage(Base):
    __tablename__ = "round_template_stages"

    id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    template_id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), ForeignKey("round_templates.id", ondelete="CASCADE"), nullable=False, index=True)
    stage_order: Mapped[int] = mapped_column(Integer, nullable=False)
    name: Mapped[str] = mapped_column(String(100), nullable=False)
    distance: Mapped[str | None] = mapped_column(String(50))  # "18m", "20yd"
    num_ends: Mapped[int] = mapped_column(Integer, nullable=False)
    arrows_per_end: Mapped[int] = mapped_column(Integer, nullable=False)
    allowed_values: Mapped[list] = mapped_column(JSON, nullable=False)  # ["X","10","9",...,"M"]
    value_score_map: Mapped[dict] = mapped_column(JSON, nullable=False)  # {"X":10,"10":10,...,"M":0}
    max_score_per_arrow: Mapped[int] = mapped_column(Integer, nullable=False)

    template: Mapped["RoundTemplate"] = relationship(back_populates="stages")
