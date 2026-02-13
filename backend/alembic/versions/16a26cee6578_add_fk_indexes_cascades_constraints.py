"""add_fk_indexes_cascades_constraints

Revision ID: 16a26cee6578
Revises: 481541dcc0cf
Create Date: 2026-02-12 20:15:18.919604
"""
from typing import Sequence, Union

from alembic import op

revision: str = '16a26cee6578'
down_revision: Union[str, None] = '481541dcc0cf'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    # Indexes on foreign keys
    op.create_index(op.f('ix_arrows_end_id'), 'arrows', ['end_id'], unique=False)
    op.create_index(op.f('ix_club_event_participants_event_id'), 'club_event_participants', ['event_id'], unique=False)
    op.create_index(op.f('ix_club_event_participants_user_id'), 'club_event_participants', ['user_id'], unique=False)
    op.create_index(op.f('ix_club_events_club_id'), 'club_events', ['club_id'], unique=False)
    op.create_index(op.f('ix_club_events_created_by'), 'club_events', ['created_by'], unique=False)
    op.create_index(op.f('ix_club_events_template_id'), 'club_events', ['template_id'], unique=False)
    op.create_index(op.f('ix_club_invites_club_id'), 'club_invites', ['club_id'], unique=False)
    op.create_index(op.f('ix_club_invites_created_by'), 'club_invites', ['created_by'], unique=False)
    op.create_index(op.f('ix_club_members_club_id'), 'club_members', ['club_id'], unique=False)
    op.create_index(op.f('ix_club_members_user_id'), 'club_members', ['user_id'], unique=False)
    op.create_index(op.f('ix_club_team_members_team_id'), 'club_team_members', ['team_id'], unique=False)
    op.create_index(op.f('ix_club_team_members_user_id'), 'club_team_members', ['user_id'], unique=False)
    op.create_index(op.f('ix_club_teams_club_id'), 'club_teams', ['club_id'], unique=False)
    op.create_index(op.f('ix_club_teams_leader_id'), 'club_teams', ['leader_id'], unique=False)
    op.create_index(op.f('ix_ends_session_id'), 'ends', ['session_id'], unique=False)
    op.create_index(op.f('ix_ends_stage_id'), 'ends', ['stage_id'], unique=False)
    op.create_index(op.f('ix_equipment_user_id'), 'equipment', ['user_id'], unique=False)
    op.create_index(op.f('ix_personal_records_session_id'), 'personal_records', ['session_id'], unique=False)
    op.create_index(op.f('ix_personal_records_template_id'), 'personal_records', ['template_id'], unique=False)
    op.create_index(op.f('ix_personal_records_user_id'), 'personal_records', ['user_id'], unique=False)
    op.create_index(op.f('ix_round_template_stages_template_id'), 'round_template_stages', ['template_id'], unique=False)
    op.create_index(op.f('ix_scoring_sessions_setup_profile_id'), 'scoring_sessions', ['setup_profile_id'], unique=False)
    op.create_index(op.f('ix_scoring_sessions_template_id'), 'scoring_sessions', ['template_id'], unique=False)
    op.create_index(op.f('ix_scoring_sessions_user_id'), 'scoring_sessions', ['user_id'], unique=False)
    op.create_index(op.f('ix_setup_equipment_equipment_id'), 'setup_equipment', ['equipment_id'], unique=False)
    op.create_index(op.f('ix_setup_equipment_setup_id'), 'setup_equipment', ['setup_id'], unique=False)
    op.create_index(op.f('ix_setup_profiles_user_id'), 'setup_profiles', ['user_id'], unique=False)

    # Cascade deletes on parent-child FKs
    op.drop_constraint(op.f('arrows_end_id_fkey'), 'arrows', type_='foreignkey')
    op.create_foreign_key(op.f('arrows_end_id_fkey'), 'arrows', 'ends', ['end_id'], ['id'], ondelete='CASCADE')

    op.drop_constraint(op.f('ends_session_id_fkey'), 'ends', type_='foreignkey')
    op.create_foreign_key(op.f('ends_session_id_fkey'), 'ends', 'scoring_sessions', ['session_id'], ['id'], ondelete='CASCADE')

    op.drop_constraint(op.f('round_template_stages_template_id_fkey'), 'round_template_stages', type_='foreignkey')
    op.create_foreign_key(op.f('round_template_stages_template_id_fkey'), 'round_template_stages', 'round_templates', ['template_id'], ['id'], ondelete='CASCADE')

    op.drop_constraint(op.f('club_members_club_id_fkey'), 'club_members', type_='foreignkey')
    op.create_foreign_key(op.f('club_members_club_id_fkey'), 'club_members', 'clubs', ['club_id'], ['id'], ondelete='CASCADE')
    op.drop_constraint(op.f('club_members_user_id_fkey'), 'club_members', type_='foreignkey')
    op.create_foreign_key(op.f('club_members_user_id_fkey'), 'club_members', 'users', ['user_id'], ['id'], ondelete='CASCADE')

    op.drop_constraint(op.f('club_invites_club_id_fkey'), 'club_invites', type_='foreignkey')
    op.create_foreign_key(op.f('club_invites_club_id_fkey'), 'club_invites', 'clubs', ['club_id'], ['id'], ondelete='CASCADE')

    op.drop_constraint(op.f('club_events_club_id_fkey'), 'club_events', type_='foreignkey')
    op.create_foreign_key(op.f('club_events_club_id_fkey'), 'club_events', 'clubs', ['club_id'], ['id'], ondelete='CASCADE')

    op.drop_constraint(op.f('club_event_participants_event_id_fkey'), 'club_event_participants', type_='foreignkey')
    op.create_foreign_key(op.f('club_event_participants_event_id_fkey'), 'club_event_participants', 'club_events', ['event_id'], ['id'], ondelete='CASCADE')

    op.drop_constraint(op.f('club_teams_club_id_fkey'), 'club_teams', type_='foreignkey')
    op.create_foreign_key(op.f('club_teams_club_id_fkey'), 'club_teams', 'clubs', ['club_id'], ['id'], ondelete='CASCADE')

    op.drop_constraint(op.f('club_team_members_team_id_fkey'), 'club_team_members', type_='foreignkey')
    op.create_foreign_key(op.f('club_team_members_team_id_fkey'), 'club_team_members', 'club_teams', ['team_id'], ['id'], ondelete='CASCADE')

    # Check constraints
    op.create_check_constraint('ck_scoring_session_status', 'scoring_sessions', "status IN ('in_progress', 'completed', 'abandoned')")
    op.create_check_constraint('ck_club_member_role', 'club_members', "role IN ('member', 'admin', 'owner')")
    op.create_check_constraint('ck_event_participant_status', 'club_event_participants', "status IN ('going', 'maybe', 'not_going')")


def downgrade() -> None:
    # Drop check constraints
    op.drop_constraint('ck_event_participant_status', 'club_event_participants', type_='check')
    op.drop_constraint('ck_club_member_role', 'club_members', type_='check')
    op.drop_constraint('ck_scoring_session_status', 'scoring_sessions', type_='check')

    # Restore original FKs (no CASCADE)
    op.drop_constraint(op.f('club_team_members_team_id_fkey'), 'club_team_members', type_='foreignkey')
    op.create_foreign_key(op.f('club_team_members_team_id_fkey'), 'club_team_members', 'club_teams', ['team_id'], ['id'])

    op.drop_constraint(op.f('club_teams_club_id_fkey'), 'club_teams', type_='foreignkey')
    op.create_foreign_key(op.f('club_teams_club_id_fkey'), 'club_teams', 'clubs', ['club_id'], ['id'])

    op.drop_constraint(op.f('club_event_participants_event_id_fkey'), 'club_event_participants', type_='foreignkey')
    op.create_foreign_key(op.f('club_event_participants_event_id_fkey'), 'club_event_participants', 'club_events', ['event_id'], ['id'])

    op.drop_constraint(op.f('club_events_club_id_fkey'), 'club_events', type_='foreignkey')
    op.create_foreign_key(op.f('club_events_club_id_fkey'), 'club_events', 'clubs', ['club_id'], ['id'])

    op.drop_constraint(op.f('club_invites_club_id_fkey'), 'club_invites', type_='foreignkey')
    op.create_foreign_key(op.f('club_invites_club_id_fkey'), 'club_invites', 'clubs', ['club_id'], ['id'])

    op.drop_constraint(op.f('club_members_user_id_fkey'), 'club_members', type_='foreignkey')
    op.create_foreign_key(op.f('club_members_user_id_fkey'), 'club_members', 'users', ['user_id'], ['id'])
    op.drop_constraint(op.f('club_members_club_id_fkey'), 'club_members', type_='foreignkey')
    op.create_foreign_key(op.f('club_members_club_id_fkey'), 'club_members', 'clubs', ['club_id'], ['id'])

    op.drop_constraint(op.f('round_template_stages_template_id_fkey'), 'round_template_stages', type_='foreignkey')
    op.create_foreign_key(op.f('round_template_stages_template_id_fkey'), 'round_template_stages', 'round_templates', ['template_id'], ['id'])

    op.drop_constraint(op.f('ends_session_id_fkey'), 'ends', type_='foreignkey')
    op.create_foreign_key(op.f('ends_session_id_fkey'), 'ends', 'scoring_sessions', ['session_id'], ['id'])

    op.drop_constraint(op.f('arrows_end_id_fkey'), 'arrows', type_='foreignkey')
    op.create_foreign_key(op.f('arrows_end_id_fkey'), 'arrows', 'ends', ['end_id'], ['id'])

    # Drop all indexes
    op.drop_index(op.f('ix_setup_profiles_user_id'), table_name='setup_profiles')
    op.drop_index(op.f('ix_setup_equipment_setup_id'), table_name='setup_equipment')
    op.drop_index(op.f('ix_setup_equipment_equipment_id'), table_name='setup_equipment')
    op.drop_index(op.f('ix_scoring_sessions_user_id'), table_name='scoring_sessions')
    op.drop_index(op.f('ix_scoring_sessions_template_id'), table_name='scoring_sessions')
    op.drop_index(op.f('ix_scoring_sessions_setup_profile_id'), table_name='scoring_sessions')
    op.drop_index(op.f('ix_round_template_stages_template_id'), table_name='round_template_stages')
    op.drop_index(op.f('ix_personal_records_user_id'), table_name='personal_records')
    op.drop_index(op.f('ix_personal_records_template_id'), table_name='personal_records')
    op.drop_index(op.f('ix_personal_records_session_id'), table_name='personal_records')
    op.drop_index(op.f('ix_equipment_user_id'), table_name='equipment')
    op.drop_index(op.f('ix_ends_stage_id'), table_name='ends')
    op.drop_index(op.f('ix_ends_session_id'), table_name='ends')
    op.drop_index(op.f('ix_club_teams_leader_id'), table_name='club_teams')
    op.drop_index(op.f('ix_club_teams_club_id'), table_name='club_teams')
    op.drop_index(op.f('ix_club_team_members_user_id'), table_name='club_team_members')
    op.drop_index(op.f('ix_club_team_members_team_id'), table_name='club_team_members')
    op.drop_index(op.f('ix_club_members_user_id'), table_name='club_members')
    op.drop_index(op.f('ix_club_members_club_id'), table_name='club_members')
    op.drop_index(op.f('ix_club_invites_created_by'), table_name='club_invites')
    op.drop_index(op.f('ix_club_invites_club_id'), table_name='club_invites')
    op.drop_index(op.f('ix_club_events_template_id'), table_name='club_events')
    op.drop_index(op.f('ix_club_events_created_by'), table_name='club_events')
    op.drop_index(op.f('ix_club_events_club_id'), table_name='club_events')
    op.drop_index(op.f('ix_club_event_participants_user_id'), table_name='club_event_participants')
    op.drop_index(op.f('ix_club_event_participants_event_id'), table_name='club_event_participants')
    op.drop_index(op.f('ix_arrows_end_id'), table_name='arrows')
