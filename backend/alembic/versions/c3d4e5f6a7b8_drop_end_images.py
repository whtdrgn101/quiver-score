"""drop end_images table and unused legacy_id column on attachments

End_images was the BYTEA-backed predecessor to GCS-stored attachments. The
project chose not to backfill the data, so the table holds nothing the API
exposes anymore — drop it along with the now-unused legacy_id column on
attachments (which existed to keep the backfill command idempotent).

Note: this migration is destructive. The downgrade recreates the schema but
cannot restore the bytea image data. That is acceptable; we made the call to
walk away from the data when we decided to skip backfill.

Revision ID: c3d4e5f6a7b8
Revises: a2b3c4d5e6f7
Create Date: 2026-05-08 16:30:00.000000
"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa
from sqlalchemy.dialects import postgresql

revision: str = 'c3d4e5f6a7b8'
down_revision: Union[str, None] = 'a2b3c4d5e6f7'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    op.drop_index('ix_attachments_legacy_id', table_name='attachments')
    op.drop_column('attachments', 'legacy_id')

    op.drop_index('ix_end_images_user_id', table_name='end_images')
    op.drop_index('ix_end_images_session_id', table_name='end_images')
    op.drop_index('ix_end_images_end_id', table_name='end_images')
    op.drop_table('end_images')


def downgrade() -> None:
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

    op.add_column(
        'attachments',
        sa.Column('legacy_id', sa.UUID(), nullable=True),
    )
    op.create_index(
        'ix_attachments_legacy_id',
        'attachments',
        ['legacy_id'],
        unique=True,
        postgresql_where=sa.text('legacy_id IS NOT NULL'),
    )
