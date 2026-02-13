"""add_classifications

Revision ID: d2e3f4a5b6c7
Revises: c1d2e3f4a5b6
Create Date: 2026-02-13 10:30:00.000000
"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa

revision: str = 'd2e3f4a5b6c7'
down_revision: Union[str, None] = 'c1d2e3f4a5b6'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    op.create_table(
        'classification_records',
        sa.Column('id', sa.UUID(), nullable=False),
        sa.Column('user_id', sa.UUID(), nullable=False),
        sa.Column('system', sa.String(50), nullable=False),
        sa.Column('classification', sa.String(50), nullable=False),
        sa.Column('round_type', sa.String(100), nullable=False),
        sa.Column('score', sa.Integer(), nullable=False),
        sa.Column('achieved_at', sa.DateTime(timezone=True), nullable=False),
        sa.Column('session_id', sa.UUID(), nullable=False),
        sa.Column('created_at', sa.DateTime(timezone=True), server_default=sa.func.now()),
        sa.ForeignKeyConstraint(['user_id'], ['users.id'], ondelete='CASCADE'),
        sa.ForeignKeyConstraint(['session_id'], ['scoring_sessions.id']),
        sa.PrimaryKeyConstraint('id'),
    )
    op.create_index('ix_classification_records_user_id', 'classification_records', ['user_id'])


def downgrade() -> None:
    op.drop_index('ix_classification_records_user_id', table_name='classification_records')
    op.drop_table('classification_records')
