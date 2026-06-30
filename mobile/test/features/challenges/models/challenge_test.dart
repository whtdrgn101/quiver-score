import 'package:flutter_test/flutter_test.dart';
import 'package:quiverscore/features/challenges/models/challenge.dart';

void main() {
  group('Challenge', () {
    test('fromJson parses all fields correctly', () {
      final json = {
        'id': 'c1',
        'challenger_id': 'u1',
        'challenger_username': 'Alice',
        'challengee_id': 'u2',
        'challengee_username': 'Bob',
        'template_id': 't1',
        'template_name': 'WA 720',
        'challenger_session_id': 's1',
        'challenger_score': 340,
        'challengee_session_id': 's2',
        'challengee_score': 338,
        'status': 'completed',
        'created_at': '2026-06-30T12:00:00Z',
        'expires_at': '2026-07-01T12:00:00Z',
      };

      final c = Challenge.fromJson(json);

      expect(c.id, 'c1');
      expect(c.challengerId, 'u1');
      expect(c.challengerUsername, 'Alice');
      expect(c.challengeeId, 'u2');
      expect(c.challengeeUsername, 'Bob');
      expect(c.templateId, 't1');
      expect(c.templateName, 'WA 720');
      expect(c.challengerSessionId, 's1');
      expect(c.challengerScore, 340);
      expect(c.challengeeSessionId, 's2');
      expect(c.challengeeScore, 338);
      expect(c.status, 'completed');
      expect(c.createdAt, isNotNull);
      expect(c.expiresAt, isNotNull);
    });

    test('fromJson handles nulls for session ids and scores', () {
      final json = {
        'id': 'c2',
        'challenger_id': 'u1',
        'challenger_username': 'Alice',
        'challengee_id': 'u2',
        'challengee_username': 'Bob',
        'template_id': 't1',
        'template_name': 'WA 720',
        'status': 'pending',
        'created_at': '2026-06-30T12:00:00Z',
      };

      final c = Challenge.fromJson(json);

      expect(c.challengerSessionId, isNull);
      expect(c.challengerScore, isNull);
      expect(c.challengeeSessionId, isNull);
      expect(c.challengeeScore, isNull);
      expect(c.expiresAt, isNull);
    });
  });
}
