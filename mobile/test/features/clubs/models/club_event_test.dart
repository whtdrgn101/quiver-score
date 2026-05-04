import 'package:flutter_test/flutter_test.dart';
import 'package:quiverscore/features/clubs/models/club_event.dart';

void main() {
  group('ClubEvent', () {
    test('fromJson parses all fields', () {
      final json = {
        'id': 'evt-1',
        'club_id': 'club-1',
        'name': 'Weekly Shoot',
        'description': 'Regular practice session',
        'template_id': 'tmpl-1',
        'template_name': 'WA 720',
        'event_date': '2025-04-01T09:00:00Z',
        'location': 'Range A',
        'created_by': 'user-1',
        'participants': [
          {
            'user_id': 'user-1',
            'username': 'archer',
            'display_name': 'Pro Archer',
            'avatar': null,
            'status': 'going',
            'score': null,
            'x_count': null,
            'session_id': null,
          }
        ],
        'created_at': '2025-03-25T10:00:00Z',
      };

      final event = ClubEvent.fromJson(json);
      expect(event.id, 'evt-1');
      expect(event.name, 'Weekly Shoot');
      expect(event.templateName, 'WA 720');
      expect(event.location, 'Range A');
      expect(event.participants.length, 1);
      expect(event.participants.first.status, 'going');
    });

    test('isPast returns true for past events', () {
      final event = ClubEvent(
        id: 'evt-1',
        clubId: 'club-1',
        name: 'Past Event',
        templateId: 'tmpl-1',
        eventDate: DateTime(2020, 1, 1),
        createdBy: 'user-1',
        createdAt: DateTime(2019, 12, 1),
      );
      expect(event.isPast, true);
    });

    test('isPast returns false for future events', () {
      final event = ClubEvent(
        id: 'evt-2',
        clubId: 'club-1',
        name: 'Future Event',
        templateId: 'tmpl-1',
        eventDate: DateTime(2030, 1, 1),
        createdBy: 'user-1',
        createdAt: DateTime(2029, 12, 1),
      );
      expect(event.isPast, false);
    });
  });

  group('EventParticipant', () {
    test('fromJson parses all fields including scores', () {
      final json = {
        'user_id': 'user-1',
        'username': 'archer',
        'display_name': 'Pro Archer',
        'avatar': null,
        'status': 'going',
        'score': 650,
        'x_count': 12,
        'session_id': 'sess-1',
      };

      final p = EventParticipant.fromJson(json);
      expect(p.userId, 'user-1');
      expect(p.status, 'going');
      expect(p.score, 650);
      expect(p.xCount, 12);
      expect(p.sessionId, 'sess-1');
      expect(p.effectiveName, 'Pro Archer');
    });

    test('fromJson handles null scores', () {
      final json = {
        'user_id': 'user-2',
        'username': 'newbie',
        'display_name': null,
        'avatar': null,
        'status': 'maybe',
        'score': null,
        'x_count': null,
        'session_id': null,
      };

      final p = EventParticipant.fromJson(json);
      expect(p.score, isNull);
      expect(p.xCount, isNull);
      expect(p.effectiveName, 'newbie');
    });
  });
}
