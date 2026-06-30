"""merge attachments and challenges branches

Revision ID: 0c2446f1aa1a
Revises: c3d4e5f6a7b8, 2d3e4f5a6b7c
Create Date: 2026-06-30 08:44:16.068157
"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa


revision: str = '0c2446f1aa1a'
down_revision: Union[str, None] = ('c3d4e5f6a7b8', '2d3e4f5a6b7c')
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    pass


def downgrade() -> None:
    pass
