"""add_tournaments

Revision ID: f4a5b6c7d8e9
Revises: e3f4a5b6c7d8
Create Date: 2026-02-13 14:00:00.000000
"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa

revision: str = 'f4a5b6c7d8e9'
down_revision: Union[str, None] = 'e3f4a5b6c7d8'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    op.create_table(
        'tournaments',
        sa.Column('id', sa.UUID(), nullable=False),
        sa.Column('name', sa.String(200), nullable=False),
        sa.Column('description', sa.Text(), nullable=True),
        sa.Column('organizer_id', sa.UUID(), nullable=False),
        sa.Column('template_id', sa.UUID(), nullable=False),
        sa.Column('status', sa.String(20), server_default='draft'),
        sa.Column('max_participants', sa.Integer(), nullable=True),
        sa.Column('registration_deadline', sa.DateTime(timezone=True), nullable=True),
        sa.Column('start_date', sa.DateTime(timezone=True), nullable=True),
        sa.Column('end_date', sa.DateTime(timezone=True), nullable=True),
        sa.Column('created_at', sa.DateTime(timezone=True), server_default=sa.func.now()),
        sa.ForeignKeyConstraint(['organizer_id'], ['users.id']),
        sa.ForeignKeyConstraint(['template_id'], ['round_templates.id']),
        sa.PrimaryKeyConstraint('id'),
        sa.CheckConstraint("status IN ('draft', 'registration', 'in_progress', 'completed')", name='ck_tournament_status'),
    )
    op.create_index('ix_tournaments_organizer_id', 'tournaments', ['organizer_id'])
    op.create_index('ix_tournaments_template_id', 'tournaments', ['template_id'])

    op.create_table(
        'tournament_participants',
        sa.Column('id', sa.UUID(), nullable=False),
        sa.Column('tournament_id', sa.UUID(), nullable=False),
        sa.Column('user_id', sa.UUID(), nullable=False),
        sa.Column('session_id', sa.UUID(), nullable=True),
        sa.Column('status', sa.String(20), server_default='registered'),
        sa.Column('final_score', sa.Integer(), nullable=True),
        sa.Column('final_x_count', sa.Integer(), nullable=True),
        sa.Column('rank', sa.Integer(), nullable=True),
        sa.Column('registered_at', sa.DateTime(timezone=True), server_default=sa.func.now()),
        sa.ForeignKeyConstraint(['tournament_id'], ['tournaments.id'], ondelete='CASCADE'),
        sa.ForeignKeyConstraint(['user_id'], ['users.id']),
        sa.ForeignKeyConstraint(['session_id'], ['scoring_sessions.id'], ondelete='SET NULL'),
        sa.PrimaryKeyConstraint('id'),
        sa.UniqueConstraint('tournament_id', 'user_id', name='uq_tournament_user'),
        sa.CheckConstraint("status IN ('registered', 'active', 'completed', 'withdrawn')", name='ck_tournament_participant_status'),
    )
    op.create_index('ix_tournament_participants_tournament_id', 'tournament_participants', ['tournament_id'])
    op.create_index('ix_tournament_participants_user_id', 'tournament_participants', ['user_id'])
    op.create_index('ix_tournament_participants_session_id', 'tournament_participants', ['session_id'])


def downgrade() -> None:
    op.drop_index('ix_tournament_participants_session_id', table_name='tournament_participants')
    op.drop_index('ix_tournament_participants_user_id', table_name='tournament_participants')
    op.drop_index('ix_tournament_participants_tournament_id', table_name='tournament_participants')
    op.drop_table('tournament_participants')
    op.drop_index('ix_tournaments_template_id', table_name='tournaments')
    op.drop_index('ix_tournaments_organizer_id', table_name='tournaments')
    op.drop_table('tournaments')
