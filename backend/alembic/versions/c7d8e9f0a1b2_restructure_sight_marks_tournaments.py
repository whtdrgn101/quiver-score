"""restructure sight marks and tournaments

Revision ID: c7d8e9f0a1b2
Revises: b6c7d8e9f0a1
Create Date: 2026-02-13 12:00:00.000000

"""
from alembic import op
import sqlalchemy as sa
from sqlalchemy.dialects import postgresql

# revision identifiers, used by Alembic.
revision = "c7d8e9f0a1b2"
down_revision = "b6c7d8e9f0a1"
branch_labels = None
depends_on = None


def upgrade() -> None:
    op.add_column(
        "sight_marks",
        sa.Column("setup_id", sa.Uuid(), nullable=True),
    )
    op.create_index(op.f("ix_sight_marks_setup_id"), "sight_marks", ["setup_id"])
    op.create_foreign_key(
        "fk_sight_marks_setup_id",
        "sight_marks",
        "setup_profiles",
        ["setup_id"],
        ["id"],
        ondelete="SET NULL",
    )

    op.add_column(
        "tournaments",
        sa.Column("club_id", sa.Uuid(), nullable=True),
    )
    op.create_index(op.f("ix_tournaments_club_id"), "tournaments", ["club_id"])
    op.create_foreign_key(
        "fk_tournaments_club_id",
        "tournaments",
        "clubs",
        ["club_id"],
        ["id"],
        ondelete="CASCADE",
    )


def downgrade() -> None:
    op.drop_constraint("fk_tournaments_club_id", "tournaments", type_="foreignkey")
    op.drop_index(op.f("ix_tournaments_club_id"), table_name="tournaments")
    op.drop_column("tournaments", "club_id")

    op.drop_constraint("fk_sight_marks_setup_id", "sight_marks", type_="foreignkey")
    op.drop_index(op.f("ix_sight_marks_setup_id"), table_name="sight_marks")
    op.drop_column("sight_marks", "setup_id")
