"""add_sight_marks

Revision ID: e3f4a5b6c7d8
Revises: d2e3f4a5b6c7
Create Date: 2026-02-13 11:00:00.000000
"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa

revision: str = 'e3f4a5b6c7d8'
down_revision: Union[str, None] = 'd2e3f4a5b6c7'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    op.create_table(
        'sight_marks',
        sa.Column('id', sa.UUID(), nullable=False),
        sa.Column('user_id', sa.UUID(), nullable=False),
        sa.Column('equipment_id', sa.UUID(), nullable=True),
        sa.Column('distance', sa.String(50), nullable=False),
        sa.Column('setting', sa.String(100), nullable=False),
        sa.Column('notes', sa.Text(), nullable=True),
        sa.Column('date_recorded', sa.DateTime(timezone=True), nullable=False),
        sa.Column('created_at', sa.DateTime(timezone=True), server_default=sa.func.now()),
        sa.ForeignKeyConstraint(['user_id'], ['users.id'], ondelete='CASCADE'),
        sa.ForeignKeyConstraint(['equipment_id'], ['equipment.id'], ondelete='SET NULL'),
        sa.PrimaryKeyConstraint('id'),
    )
    op.create_index('ix_sight_marks_user_id', 'sight_marks', ['user_id'])
    op.create_index('ix_sight_marks_equipment_id', 'sight_marks', ['equipment_id'])


def downgrade() -> None:
    op.drop_index('ix_sight_marks_equipment_id', table_name='sight_marks')
    op.drop_index('ix_sight_marks_user_id', table_name='sight_marks')
    op.drop_table('sight_marks')
