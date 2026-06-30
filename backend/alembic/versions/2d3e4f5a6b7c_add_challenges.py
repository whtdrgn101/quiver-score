"""add_challenges

Revision ID: 2d3e4f5a6b7c
Revises: 1c0254d7648f
Create Date: 2026-06-30 08:35:00.000000
"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa

revision: str = '2d3e4f5a6b7c'
down_revision: Union[str, None] = '1c0254d7648f'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    op.create_table(
        'challenges',
        sa.Column('id', sa.UUID(), nullable=False),
        sa.Column('challenger_id', sa.UUID(), nullable=False),
        sa.Column('challengee_id', sa.UUID(), nullable=False),
        sa.Column('template_id', sa.UUID(), nullable=False),
        sa.Column('challenger_session_id', sa.UUID(), nullable=True),
        sa.Column('challengee_session_id', sa.UUID(), nullable=True),
        sa.Column('status', sa.String(length=50), server_default='pending', nullable=False),
        sa.Column('created_at', sa.DateTime(timezone=True), server_default=sa.func.now(), nullable=False),
        sa.Column('updated_at', sa.DateTime(timezone=True), server_default=sa.func.now(), nullable=False),
        sa.Column('expires_at', sa.DateTime(timezone=True), nullable=True),
        sa.ForeignKeyConstraint(['challenger_id'], ['users.id'], ondelete='CASCADE'),
        sa.ForeignKeyConstraint(['challengee_id'], ['users.id'], ondelete='CASCADE'),
        sa.ForeignKeyConstraint(['template_id'], ['round_templates.id'], ondelete='CASCADE'),
        sa.ForeignKeyConstraint(['challenger_session_id'], ['scoring_sessions.id'], ondelete='SET NULL'),
        sa.ForeignKeyConstraint(['challengee_session_id'], ['scoring_sessions.id'], ondelete='SET NULL'),
        sa.PrimaryKeyConstraint('id')
    )
    op.create_index('ix_challenges_challenger_id', 'challenges', ['challenger_id'])
    op.create_index('ix_challenges_challengee_id', 'challenges', ['challengee_id'])
    op.create_index('ix_challenges_status', 'challenges', ['status'])


def downgrade() -> None:
    op.drop_index('ix_challenges_status', table_name='challenges')
    op.drop_index('ix_challenges_challengee_id', table_name='challenges')
    op.drop_index('ix_challenges_challenger_id', table_name='challenges')
    op.drop_table('challenges')
