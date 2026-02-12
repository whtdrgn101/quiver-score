"""add personal_records table

Revision ID: a1b2c3d4e5f6
Revises: 459225dbf800
Create Date: 2026-02-11 20:00:00.000000
"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa
from sqlalchemy.dialects import postgresql


revision: str = 'a1b2c3d4e5f6'
down_revision: Union[str, None] = '459225dbf800'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    op.create_table(
        'personal_records',
        sa.Column('id', postgresql.UUID(as_uuid=True), primary_key=True),
        sa.Column('user_id', postgresql.UUID(as_uuid=True), sa.ForeignKey('users.id'), nullable=False),
        sa.Column('template_id', postgresql.UUID(as_uuid=True), sa.ForeignKey('round_templates.id'), nullable=False),
        sa.Column('session_id', postgresql.UUID(as_uuid=True), sa.ForeignKey('scoring_sessions.id'), nullable=False),
        sa.Column('score', sa.Integer(), nullable=False),
        sa.Column('achieved_at', sa.DateTime(timezone=True), nullable=False),
        sa.UniqueConstraint('user_id', 'template_id', name='uq_user_template_pr'),
    )


def downgrade() -> None:
    op.drop_table('personal_records')
