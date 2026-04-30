"""add tournament_rounds and tournament_round_scores tables

Revision ID: b2c3d4e5f6a7
Revises: a1b2c3d4e5f7
Create Date: 2026-04-30 10:00:00.000000
"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa

revision: str = 'b2c3d4e5f6a7'
down_revision: Union[str, None] = 'a1b2c3d4e5f7'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    op.create_table(
        'tournament_rounds',
        sa.Column('id', sa.UUID(), nullable=False),
        sa.Column('tournament_id', sa.UUID(), nullable=False),
        sa.Column('round_number', sa.Integer(), nullable=False),
        sa.Column('name', sa.String(100), nullable=False),
        sa.Column('template_id', sa.UUID(), nullable=True),
        sa.Column('advancement', sa.Integer(), nullable=True),
        sa.Column('status', sa.String(20), server_default='pending', nullable=False),
        sa.Column('started_at', sa.DateTime(timezone=True), nullable=True),
        sa.Column('completed_at', sa.DateTime(timezone=True), nullable=True),
        sa.Column('created_at', sa.DateTime(timezone=True), server_default=sa.func.now()),
        sa.PrimaryKeyConstraint('id'),
        sa.ForeignKeyConstraint(['tournament_id'], ['tournaments.id'], ondelete='CASCADE'),
        sa.ForeignKeyConstraint(['template_id'], ['round_templates.id']),
        sa.CheckConstraint("status IN ('pending', 'in_progress', 'completed')", name='ck_tournament_round_status'),
        sa.UniqueConstraint('tournament_id', 'round_number', name='uq_tournament_round_number'),
    )
    op.create_index('ix_tournament_rounds_tournament_id', 'tournament_rounds', ['tournament_id'])
    op.create_index('ix_tournament_rounds_template_id', 'tournament_rounds', ['template_id'])

    op.create_table(
        'tournament_round_scores',
        sa.Column('id', sa.UUID(), nullable=False),
        sa.Column('round_id', sa.UUID(), nullable=False),
        sa.Column('participant_id', sa.UUID(), nullable=False),
        sa.Column('session_id', sa.UUID(), nullable=True),
        sa.Column('score', sa.Integer(), nullable=True),
        sa.Column('x_count', sa.Integer(), nullable=True),
        sa.Column('rank_in_round', sa.Integer(), nullable=True),
        sa.Column('advanced', sa.Boolean(), server_default='false', nullable=False),
        sa.Column('created_at', sa.DateTime(timezone=True), server_default=sa.func.now()),
        sa.PrimaryKeyConstraint('id'),
        sa.ForeignKeyConstraint(['round_id'], ['tournament_rounds.id'], ondelete='CASCADE'),
        sa.ForeignKeyConstraint(['participant_id'], ['tournament_participants.id'], ondelete='CASCADE'),
        sa.ForeignKeyConstraint(['session_id'], ['scoring_sessions.id'], ondelete='SET NULL'),
        sa.UniqueConstraint('round_id', 'participant_id', name='uq_round_participant'),
    )
    op.create_index('ix_tournament_round_scores_round_id', 'tournament_round_scores', ['round_id'])
    op.create_index('ix_tournament_round_scores_participant_id', 'tournament_round_scores', ['participant_id'])
    op.create_index('ix_tournament_round_scores_session_id', 'tournament_round_scores', ['session_id'])


def downgrade() -> None:
    op.drop_index('ix_tournament_round_scores_session_id')
    op.drop_index('ix_tournament_round_scores_participant_id')
    op.drop_index('ix_tournament_round_scores_round_id')
    op.drop_table('tournament_round_scores')
    op.drop_index('ix_tournament_rounds_template_id')
    op.drop_index('ix_tournament_rounds_tournament_id')
    op.drop_table('tournament_rounds')
