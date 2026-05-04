import 'package:flutter_test/flutter_test.dart';
import 'package:quiverscore/features/clubs/models/club.dart';

void main() {
  group('Club', () {
    test('fromJson parses all fields', () {
      final json = {
        'id': 'club-1',
        'name': 'City Archers',
        'description': 'Local archery club',
        'avatar': 'https://example.com/avatar.jpg',
        'owner_id': 'user-1',
        'member_count': 15,
        'my_role': 'member',
        'created_at': '2025-01-01T00:00:00Z',
      };

      final club = Club.fromJson(json);
      expect(club.id, 'club-1');
      expect(club.name, 'City Archers');
      expect(club.description, 'Local archery club');
      expect(club.avatar, 'https://example.com/avatar.jpg');
      expect(club.ownerId, 'user-1');
      expect(club.memberCount, 15);
      expect(club.myRole, 'member');
    });

    test('fromJson handles null optional fields', () {
      final json = {
        'id': 'club-2',
        'name': 'New Club',
        'description': null,
        'avatar': null,
        'owner_id': 'user-1',
        'member_count': 1,
        'my_role': null,
        'created_at': '2025-01-01T00:00:00Z',
      };

      final club = Club.fromJson(json);
      expect(club.description, isNull);
      expect(club.avatar, isNull);
      expect(club.myRole, isNull);
    });
  });

  group('ClubMember', () {
    test('fromJson parses all fields', () {
      final json = {
        'user_id': 'user-1',
        'username': 'archer42',
        'display_name': 'Jane Archer',
        'avatar': null,
        'role': 'admin',
        'joined_at': '2025-03-01T10:00:00Z',
      };

      final member = ClubMember.fromJson(json);
      expect(member.userId, 'user-1');
      expect(member.username, 'archer42');
      expect(member.displayName, 'Jane Archer');
      expect(member.role, 'admin');
      expect(member.effectiveName, 'Jane Archer');
    });

    test('effectiveName falls back to username', () {
      final member = ClubMember(
        userId: 'user-2',
        username: 'bowman',
        role: 'member',
        joinedAt: DateTime(2025),
      );
      expect(member.effectiveName, 'bowman');
    });
  });

  group('ClubDetail', () {
    test('fromJson parses members list', () {
      final json = {
        'id': 'club-1',
        'name': 'City Archers',
        'description': null,
        'avatar': null,
        'owner_id': 'user-1',
        'member_count': 2,
        'my_role': 'owner',
        'created_at': '2025-01-01T00:00:00Z',
        'members': [
          {
            'user_id': 'user-1',
            'username': 'owner',
            'display_name': null,
            'avatar': null,
            'role': 'owner',
            'joined_at': '2025-01-01T00:00:00Z',
          },
          {
            'user_id': 'user-2',
            'username': 'member',
            'display_name': 'A Member',
            'avatar': null,
            'role': 'member',
            'joined_at': '2025-02-01T00:00:00Z',
          },
        ],
      };

      final detail = ClubDetail.fromJson(json);
      expect(detail.members.length, 2);
      expect(detail.members.first.role, 'owner');
      expect(detail.members.last.effectiveName, 'A Member');
    });
  });

  group('LeaderboardGroup', () {
    test('fromJson parses entries', () {
      final json = {
        'template_id': 'tmpl-1',
        'template_name': 'WA 720',
        'entries': [
          {
            'user_id': 'user-1',
            'username': 'sharpshooter',
            'display_name': null,
            'avatar': null,
            'best_score': 680,
            'best_x_count': 15,
            'session_id': 'sess-1',
            'achieved_at': '2025-03-15T12:00:00Z',
          },
        ],
      };

      final group = LeaderboardGroup.fromJson(json);
      expect(group.templateName, 'WA 720');
      expect(group.entries.length, 1);
      expect(group.entries.first.bestScore, 680);
      expect(group.entries.first.bestXCount, 15);
    });
  });

  group('ActivityItem', () {
    test('fromJson parses all fields', () {
      final json = {
        'type': 'session_complete',
        'user_id': 'user-1',
        'username': 'archer',
        'display_name': 'Pro Archer',
        'avatar': null,
        'template_name': 'WA 720',
        'score': 650,
        'x_count': 10,
        'session_id': 'sess-1',
        'occurred_at': '2025-03-20T14:30:00Z',
      };

      final item = ActivityItem.fromJson(json);
      expect(item.type, 'session_complete');
      expect(item.score, 650);
      expect(item.xCount, 10);
      expect(item.effectiveName, 'Pro Archer');
    });

    test('personal_record type', () {
      final json = {
        'type': 'personal_record',
        'user_id': 'user-1',
        'username': 'archer',
        'display_name': null,
        'avatar': null,
        'template_name': 'Portsmouth',
        'score': 590,
        'x_count': 45,
        'session_id': 'sess-2',
        'occurred_at': '2025-03-21T09:00:00Z',
      };

      final item = ActivityItem.fromJson(json);
      expect(item.type, 'personal_record');
      expect(item.effectiveName, 'archer');
    });
  });
}
