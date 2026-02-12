"""add_club_team_tables

Revision ID: 481541dcc0cf
Revises: b3c4d5e6f7a8
Create Date: 2026-02-12 12:20:10.692960
"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa

revision: str = '481541dcc0cf'
down_revision: Union[str, None] = 'b3c4d5e6f7a8'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    op.create_table('club_teams',
    sa.Column('id', sa.UUID(), nullable=False),
    sa.Column('club_id', sa.UUID(), nullable=False),
    sa.Column('name', sa.String(length=100), nullable=False),
    sa.Column('description', sa.Text(), nullable=True),
    sa.Column('leader_id', sa.UUID(), nullable=False),
    sa.Column('created_at', sa.DateTime(timezone=True), server_default=sa.text('now()'), nullable=False),
    sa.ForeignKeyConstraint(['club_id'], ['clubs.id'], ),
    sa.ForeignKeyConstraint(['leader_id'], ['users.id'], ),
    sa.PrimaryKeyConstraint('id')
    )
    op.create_table('club_team_members',
    sa.Column('id', sa.UUID(), nullable=False),
    sa.Column('team_id', sa.UUID(), nullable=False),
    sa.Column('user_id', sa.UUID(), nullable=False),
    sa.Column('joined_at', sa.DateTime(timezone=True), server_default=sa.text('now()'), nullable=False),
    sa.ForeignKeyConstraint(['team_id'], ['club_teams.id'], ),
    sa.ForeignKeyConstraint(['user_id'], ['users.id'], ),
    sa.PrimaryKeyConstraint('id'),
    sa.UniqueConstraint('team_id', 'user_id', name='uq_team_user')
    )


def downgrade() -> None:
    op.drop_table('club_team_members')
    op.drop_table('club_teams')
