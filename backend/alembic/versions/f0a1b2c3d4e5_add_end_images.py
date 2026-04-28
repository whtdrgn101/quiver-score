"""add end_images table

Revision ID: f0a1b2c3d4e5
Revises: e9f0a1b2c3d4
Create Date: 2026-04-28 20:00:00.000000
"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa
from sqlalchemy.dialects import postgresql

revision: str = 'f0a1b2c3d4e5'
down_revision: Union[str, None] = 'e9f0a1b2c3d4'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    op.create_table(
        'end_images',
        sa.Column('id', sa.UUID(), nullable=False),
        sa.Column('end_id', sa.UUID(), nullable=False),
        sa.Column('session_id', sa.UUID(), nullable=False),
        sa.Column('user_id', sa.UUID(), nullable=False),
        sa.Column('image_data', postgresql.BYTEA(), nullable=False),
        sa.Column('content_type', sa.String(50), nullable=False),
        sa.Column('file_size', sa.Integer(), nullable=False),
        sa.Column('created_at', sa.DateTime(timezone=True), server_default=sa.func.now()),
        sa.ForeignKeyConstraint(['end_id'], ['ends.id'], ondelete='CASCADE'),
        sa.ForeignKeyConstraint(['session_id'], ['scoring_sessions.id'], ondelete='CASCADE'),
        sa.ForeignKeyConstraint(['user_id'], ['users.id'], ondelete='CASCADE'),
        sa.PrimaryKeyConstraint('id'),
    )
    op.create_index('ix_end_images_end_id', 'end_images', ['end_id'])
    op.create_index('ix_end_images_session_id', 'end_images', ['session_id'])
    op.create_index('ix_end_images_user_id', 'end_images', ['user_id'])


def downgrade() -> None:
    op.drop_index('ix_end_images_user_id', table_name='end_images')
    op.drop_index('ix_end_images_session_id', table_name='end_images')
    op.drop_index('ix_end_images_end_id', table_name='end_images')
    op.drop_table('end_images')
