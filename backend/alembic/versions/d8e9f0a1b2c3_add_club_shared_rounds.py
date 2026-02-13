"""add club_shared_rounds

Revision ID: d8e9f0a1b2c3
Revises: c7d8e9f0a1b2
Create Date: 2026-02-13 18:00:00.000000

"""
from alembic import op
import sqlalchemy as sa
from sqlalchemy.dialects import postgresql

# revision identifiers, used by Alembic.
revision = "d8e9f0a1b2c3"
down_revision = "c7d8e9f0a1b2"
branch_labels = None
depends_on = None


def upgrade() -> None:
    op.create_table(
        "club_shared_rounds",
        sa.Column("id", postgresql.UUID(as_uuid=True), primary_key=True),
        sa.Column("club_id", postgresql.UUID(as_uuid=True), sa.ForeignKey("clubs.id", ondelete="CASCADE"), nullable=False, index=True),
        sa.Column("template_id", postgresql.UUID(as_uuid=True), sa.ForeignKey("round_templates.id", ondelete="CASCADE"), nullable=False, index=True),
        sa.Column("shared_by", postgresql.UUID(as_uuid=True), sa.ForeignKey("users.id"), nullable=False),
        sa.Column("shared_at", sa.DateTime(timezone=True), server_default=sa.func.now()),
        sa.UniqueConstraint("club_id", "template_id", name="uq_club_template"),
    )


def downgrade() -> None:
    op.drop_table("club_shared_rounds")
