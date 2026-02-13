"""add_coaching

Revision ID: a5b6c7d8e9f0
Revises: f4a5b6c7d8e9
Create Date: 2026-02-13 15:00:00.000000
"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa

revision: str = 'a5b6c7d8e9f0'
down_revision: Union[str, None] = 'f4a5b6c7d8e9'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    op.create_table(
        'coach_athlete_links',
        sa.Column('id', sa.UUID(), nullable=False),
        sa.Column('coach_id', sa.UUID(), nullable=False),
        sa.Column('athlete_id', sa.UUID(), nullable=False),
        sa.Column('status', sa.String(20), server_default='pending'),
        sa.Column('created_at', sa.DateTime(timezone=True), server_default=sa.func.now()),
        sa.ForeignKeyConstraint(['coach_id'], ['users.id']),
        sa.ForeignKeyConstraint(['athlete_id'], ['users.id']),
        sa.PrimaryKeyConstraint('id'),
        sa.UniqueConstraint('coach_id', 'athlete_id', name='uq_coach_athlete'),
        sa.CheckConstraint("status IN ('pending', 'active', 'revoked')", name='ck_coach_athlete_status'),
    )
    op.create_index('ix_coach_athlete_links_coach_id', 'coach_athlete_links', ['coach_id'])
    op.create_index('ix_coach_athlete_links_athlete_id', 'coach_athlete_links', ['athlete_id'])

    op.create_table(
        'session_annotations',
        sa.Column('id', sa.UUID(), nullable=False),
        sa.Column('session_id', sa.UUID(), nullable=False),
        sa.Column('author_id', sa.UUID(), nullable=False),
        sa.Column('end_number', sa.Integer(), nullable=True),
        sa.Column('arrow_number', sa.Integer(), nullable=True),
        sa.Column('text', sa.Text(), nullable=False),
        sa.Column('created_at', sa.DateTime(timezone=True), server_default=sa.func.now()),
        sa.ForeignKeyConstraint(['session_id'], ['scoring_sessions.id'], ondelete='CASCADE'),
        sa.ForeignKeyConstraint(['author_id'], ['users.id']),
        sa.PrimaryKeyConstraint('id'),
    )
    op.create_index('ix_session_annotations_session_id', 'session_annotations', ['session_id'])
    op.create_index('ix_session_annotations_author_id', 'session_annotations', ['author_id'])


def downgrade() -> None:
    op.drop_index('ix_session_annotations_author_id', table_name='session_annotations')
    op.drop_index('ix_session_annotations_session_id', table_name='session_annotations')
    op.drop_table('session_annotations')
    op.drop_index('ix_coach_athlete_links_athlete_id', table_name='coach_athlete_links')
    op.drop_index('ix_coach_athlete_links_coach_id', table_name='coach_athlete_links')
    op.drop_table('coach_athlete_links')
