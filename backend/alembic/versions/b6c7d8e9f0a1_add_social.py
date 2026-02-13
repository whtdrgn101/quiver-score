"""add_social

Revision ID: b6c7d8e9f0a1
Revises: a5b6c7d8e9f0
Create Date: 2026-02-13 16:00:00.000000
"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa

revision: str = 'b6c7d8e9f0a1'
down_revision: Union[str, None] = 'a5b6c7d8e9f0'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    op.create_table(
        'follows',
        sa.Column('id', sa.UUID(), nullable=False),
        sa.Column('follower_id', sa.UUID(), nullable=False),
        sa.Column('following_id', sa.UUID(), nullable=False),
        sa.Column('created_at', sa.DateTime(timezone=True), server_default=sa.func.now()),
        sa.ForeignKeyConstraint(['follower_id'], ['users.id'], ondelete='CASCADE'),
        sa.ForeignKeyConstraint(['following_id'], ['users.id'], ondelete='CASCADE'),
        sa.PrimaryKeyConstraint('id'),
        sa.UniqueConstraint('follower_id', 'following_id', name='uq_follower_following'),
    )
    op.create_index('ix_follows_follower_id', 'follows', ['follower_id'])
    op.create_index('ix_follows_following_id', 'follows', ['following_id'])

    op.create_table(
        'feed_items',
        sa.Column('id', sa.UUID(), nullable=False),
        sa.Column('user_id', sa.UUID(), nullable=False),
        sa.Column('type', sa.String(50), nullable=False),
        sa.Column('data', sa.JSON(), server_default='{}'),
        sa.Column('created_at', sa.DateTime(timezone=True), server_default=sa.func.now()),
        sa.ForeignKeyConstraint(['user_id'], ['users.id'], ondelete='CASCADE'),
        sa.PrimaryKeyConstraint('id'),
    )
    op.create_index('ix_feed_items_user_id', 'feed_items', ['user_id'])
    op.create_index('ix_feed_items_created_at', 'feed_items', ['created_at'])


def downgrade() -> None:
    op.drop_index('ix_feed_items_created_at', table_name='feed_items')
    op.drop_index('ix_feed_items_user_id', table_name='feed_items')
    op.drop_table('feed_items')
    op.drop_index('ix_follows_following_id', table_name='follows')
    op.drop_index('ix_follows_follower_id', table_name='follows')
    op.drop_table('follows')
