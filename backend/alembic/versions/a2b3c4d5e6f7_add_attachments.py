"""add attachments table for GCS-backed image storage

Revision ID: a2b3c4d5e6f7
Revises: b2c3d4e5f6a7
Create Date: 2026-05-08 14:30:00.000000
"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa

revision: str = 'a2b3c4d5e6f7'
down_revision: Union[str, None] = 'b2c3d4e5f6a7'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


# Polymorphic owner — application enforces FK validity per owner_type via the
# OwnerVerifier registry in the Go handler. Postgres can't model polymorphic
# FKs, so cascade on parent delete is handled in application code; user_id has
# a real FK so account deletion still cascades through the database.
def upgrade() -> None:
    op.create_table(
        'attachments',
        sa.Column('id', sa.UUID(), nullable=False),
        sa.Column('user_id', sa.UUID(), nullable=False),
        sa.Column('owner_type', sa.String(32), nullable=False),
        sa.Column('owner_id', sa.UUID(), nullable=False),
        sa.Column('storage_key', sa.Text(), nullable=False),
        sa.Column('thumb_key', sa.Text(), nullable=False),
        sa.Column('content_type', sa.String(50), nullable=False),
        sa.Column('full_size', sa.Integer(), nullable=False),
        sa.Column('thumb_size', sa.Integer(), nullable=False),
        sa.Column('width', sa.Integer(), nullable=False),
        sa.Column('height', sa.Integer(), nullable=False),
        # legacy_id holds end_images.id during the backfill window so the
        # backfill command is idempotent (re-running skips already-migrated
        # rows). Dropped in the cleanup migration after Phase 19.
        sa.Column('legacy_id', sa.UUID(), nullable=True),
        sa.Column('created_at', sa.DateTime(timezone=True), server_default=sa.func.now(), nullable=False),
        sa.PrimaryKeyConstraint('id'),
        sa.ForeignKeyConstraint(['user_id'], ['users.id'], ondelete='CASCADE'),
        sa.CheckConstraint(
            "owner_type IN ('session_end', 'equipment', 'setup')",
            name='ck_attachments_owner_type',
        ),
    )
    op.create_index('ix_attachments_owner', 'attachments', ['owner_type', 'owner_id'])
    op.create_index('ix_attachments_user_id', 'attachments', ['user_id'])
    op.create_index(
        'ix_attachments_legacy_id',
        'attachments',
        ['legacy_id'],
        unique=True,
        postgresql_where=sa.text('legacy_id IS NOT NULL'),
    )


def downgrade() -> None:
    op.drop_index('ix_attachments_legacy_id', table_name='attachments')
    op.drop_index('ix_attachments_user_id', table_name='attachments')
    op.drop_index('ix_attachments_owner', table_name='attachments')
    op.drop_table('attachments')
