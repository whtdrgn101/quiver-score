import 'package:flutter_test/flutter_test.dart';
import 'package:quiverscore/features/clubs/models/club_team.dart';

void main() {
  group('ClubTeam', () {
    test('fromJson parses all fields', () {
      final json = {
        'id': 'team-1',
        'club_id': 'club-1',
        'name': 'Alpha Squad',
        'description': 'Competition team',
        'leader': {
          'user_id': 'user-1',
          'username': 'captain',
          'display_name': 'Team Captain',
          'avatar': null,
          'joined_at': '2025-01-01T00:00:00Z',
        },
        'member_count': 5,
        'created_at': '2025-02-01T00:00:00Z',
      };

      final team = ClubTeam.fromJson(json);
      expect(team.id, 'team-1');
      expect(team.name, 'Alpha Squad');
      expect(team.description, 'Competition team');
      expect(team.leader.effectiveName, 'Team Captain');
      expect(team.memberCount, 5);
    });
  });

  group('ClubTeamDetail', () {
    test('fromJson parses members list', () {
      final json = {
        'id': 'team-1',
        'club_id': 'club-1',
        'name': 'Alpha Squad',
        'description': null,
        'leader': {
          'user_id': 'user-1',
          'username': 'captain',
          'display_name': null,
          'avatar': null,
          'joined_at': '2025-01-01T00:00:00Z',
        },
        'member_count': 2,
        'created_at': '2025-02-01T00:00:00Z',
        'members': [
          {
            'user_id': 'user-1',
            'username': 'captain',
            'display_name': null,
            'avatar': null,
            'joined_at': '2025-01-01T00:00:00Z',
          },
          {
            'user_id': 'user-2',
            'username': 'member1',
            'display_name': 'Team Member',
            'avatar': null,
            'joined_at': '2025-02-15T00:00:00Z',
          },
        ],
      };

      final detail = ClubTeamDetail.fromJson(json);
      expect(detail.members.length, 2);
      expect(detail.leader.username, 'captain');
      expect(detail.members.last.effectiveName, 'Team Member');
    });
  });

  group('TeamMember', () {
    test('effectiveName uses displayName when available', () {
      final member = TeamMember(
        userId: 'user-1',
        username: 'jdoe',
        displayName: 'John Doe',
        joinedAt: DateTime(2025),
      );
      expect(member.effectiveName, 'John Doe');
    });

    test('effectiveName falls back to username', () {
      final member = TeamMember(
        userId: 'user-1',
        username: 'jdoe',
        joinedAt: DateTime(2025),
      );
      expect(member.effectiveName, 'jdoe');
    });
  });
}
