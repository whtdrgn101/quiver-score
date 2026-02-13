"""make end stage_id nullable with SET NULL

Revision ID: e9f0a1b2c3d4
Revises: d8e9f0a1b2c3
Create Date: 2026-02-13 19:00:00.000000

"""
from alembic import op
import sqlalchemy as sa
from sqlalchemy.dialects import postgresql

revision = "e9f0a1b2c3d4"
down_revision = "d8e9f0a1b2c3"
branch_labels = None
depends_on = None


def upgrade() -> None:
    op.alter_column("ends", "stage_id", existing_type=postgresql.UUID(), nullable=True)
    op.drop_constraint("ends_stage_id_fkey", "ends", type_="foreignkey")
    op.create_foreign_key(
        "ends_stage_id_fkey",
        "ends",
        "round_template_stages",
        ["stage_id"],
        ["id"],
        ondelete="SET NULL",
    )


def downgrade() -> None:
    op.drop_constraint("ends_stage_id_fkey", "ends", type_="foreignkey")
    op.create_foreign_key(
        "ends_stage_id_fkey",
        "ends",
        "round_template_stages",
        ["stage_id"],
        ["id"],
    )
    op.alter_column("ends", "stage_id", existing_type=postgresql.UUID(), nullable=False)
