"""add_tournament_matchups

Revision ID: 1c0254d7648f
Revises: b2c3d4e5f6a7
Create Date: 2026-06-29 14:23:54.735956
"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa


revision: str = '1c0254d7648f'
down_revision: Union[str, None] = 'b2c3d4e5f6a7'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    op.add_column(
        'tournament_rounds',
        sa.Column('round_type', sa.String(length=20), server_default='qualification', nullable=False)
    )
    op.create_check_constraint(
        'ck_tournament_round_type',
        'tournament_rounds',
        "round_type IN ('qualification', 'elimination')"
    )

    op.create_table(
        'tournament_matchups',
        sa.Column('id', sa.UUID(), nullable=False),
        sa.Column('round_id', sa.UUID(), nullable=False),
        sa.Column('match_number', sa.Integer(), nullable=False),
        sa.Column('participant_a_id', sa.UUID(), nullable=True),
        sa.Column('participant_b_id', sa.UUID(), nullable=True),
        sa.Column('score_a', sa.Integer(), nullable=True),
        sa.Column('score_b', sa.Integer(), nullable=True),
        sa.Column('winner_id', sa.UUID(), nullable=True),
        sa.Column('created_at', sa.DateTime(timezone=True), server_default=sa.func.now(), nullable=False),
        sa.PrimaryKeyConstraint('id'),
        sa.ForeignKeyConstraint(['round_id'], ['tournament_rounds.id'], ondelete='CASCADE'),
        sa.ForeignKeyConstraint(['participant_a_id'], ['tournament_participants.id'], ondelete='SET NULL'),
        sa.ForeignKeyConstraint(['participant_b_id'], ['tournament_participants.id'], ondelete='SET NULL'),
        sa.ForeignKeyConstraint(['winner_id'], ['tournament_participants.id'], ondelete='SET NULL'),
        sa.UniqueConstraint('round_id', 'match_number', name='uq_round_match_number')
    )
    op.create_index('ix_tournament_matchups_round_id', 'tournament_matchups', ['round_id'])


def downgrade() -> None:
    op.drop_index('ix_tournament_matchups_round_id', table_name='tournament_matchups')
    op.drop_table('tournament_matchups')
    
    op.drop_constraint('ck_tournament_round_type', 'tournament_rounds', type_='check')
    op.drop_column('tournament_rounds', 'round_type')

