import 'package:flutter_test/flutter_test.dart';
import 'package:quiverscore/features/clubs/models/tournament.dart';

void main() {
  group('Tournament', () {
    test('fromJson parses all fields', () {
      final json = {
        'id': 't1',
        'name': 'Spring Championship',
        'description': 'Annual spring event',
        'organizer_id': 'u1',
        'organizer_name': 'Alice',
        'template_id': 'tmpl1',
        'template_name': 'WA 720',
        'status': 'in_progress',
        'max_participants': 32,
        'registration_deadline': '2026-04-01T00:00:00Z',
        'start_date': '2026-04-15T00:00:00Z',
        'end_date': '2026-04-16T00:00:00Z',
        'participant_count': 24,
        'club_id': 'c1',
        'club_name': 'Bowmen Club',
        'created_at': '2026-03-01T12:00:00Z',
      };

      final t = Tournament.fromJson(json);

      expect(t.id, 't1');
      expect(t.name, 'Spring Championship');
      expect(t.description, 'Annual spring event');
      expect(t.organizerId, 'u1');
      expect(t.organizerName, 'Alice');
      expect(t.templateId, 'tmpl1');
      expect(t.templateName, 'WA 720');
      expect(t.status, 'in_progress');
      expect(t.maxParticipants, 32);
      expect(t.registrationDeadline, isNotNull);
      expect(t.startDate, isNotNull);
      expect(t.endDate, isNotNull);
      expect(t.participantCount, 24);
      expect(t.clubId, 'c1');
      expect(t.clubName, 'Bowmen Club');
    });

    test('fromJson handles nullable fields', () {
      final json = {
        'id': 't2',
        'name': 'Quick Shoot',
        'organizer_id': 'u1',
        'template_id': 'tmpl1',
        'status': 'registration',
        'club_id': 'c1',
        'created_at': '2026-03-01T12:00:00Z',
      };

      final t = Tournament.fromJson(json);

      expect(t.description, isNull);
      expect(t.organizerName, isNull);
      expect(t.templateName, isNull);
      expect(t.maxParticipants, isNull);
      expect(t.registrationDeadline, isNull);
      expect(t.startDate, isNull);
      expect(t.endDate, isNull);
      expect(t.participantCount, 0);
    });
  });

  group('TournamentDetail', () {
    test('fromJson parses participants', () {
      final json = {
        'id': 't1',
        'name': 'Championship',
        'organizer_id': 'u1',
        'template_id': 'tmpl1',
        'status': 'in_progress',
        'participant_count': 2,
        'club_id': 'c1',
        'created_at': '2026-03-01T12:00:00Z',
        'participants': [
          {
            'user_id': 'u1',
            'username': 'Alice',
            'final_score': 650,
            'final_x_count': 12,
            'status': 'active',
          },
          {
            'user_id': 'u2',
            'username': 'Bob',
            'status': 'withdrawn',
          },
        ],
      };

      final td = TournamentDetail.fromJson(json);

      expect(td.participants, hasLength(2));
      expect(td.participants[0].userId, 'u1');
      expect(td.participants[0].username, 'Alice');
      expect(td.participants[0].finalScore, 650);
      expect(td.participants[0].finalXCount, 12);
      expect(td.participants[0].status, 'active');
      expect(td.participants[1].finalScore, isNull);
      expect(td.participants[1].status, 'withdrawn');
    });

    test('fromJson handles null participants', () {
      final json = {
        'id': 't1',
        'name': 'Championship',
        'organizer_id': 'u1',
        'template_id': 'tmpl1',
        'status': 'registration',
        'club_id': 'c1',
        'created_at': '2026-03-01T12:00:00Z',
      };

      final td = TournamentDetail.fromJson(json);

      expect(td.participants, isEmpty);
    });
  });

  group('TournamentRound', () {
    test('fromJson parses all fields', () {
      final json = {
        'id': 'r1',
        'tournament_id': 't1',
        'round_number': 1,
        'name': 'Qualifying',
        'template_id': 'tmpl2',
        'template_name': 'WA 60',
        'advancement': 16,
        'status': 'completed',
        'started_at': '2026-04-15T09:00:00Z',
        'completed_at': '2026-04-15T12:00:00Z',
        'created_at': '2026-04-01T00:00:00Z',
      };

      final r = TournamentRound.fromJson(json);

      expect(r.id, 'r1');
      expect(r.tournamentId, 't1');
      expect(r.roundNumber, 1);
      expect(r.name, 'Qualifying');
      expect(r.templateId, 'tmpl2');
      expect(r.templateName, 'WA 60');
      expect(r.advancement, 16);
      expect(r.status, 'completed');
      expect(r.startedAt, isNotNull);
      expect(r.completedAt, isNotNull);
    });

    test('fromJson handles nullable fields', () {
      final json = {
        'id': 'r2',
        'tournament_id': 't1',
        'round_number': 2,
        'name': 'Finals',
        'status': 'pending',
        'created_at': '2026-04-01T00:00:00Z',
      };

      final r = TournamentRound.fromJson(json);

      expect(r.templateId, isNull);
      expect(r.templateName, isNull);
      expect(r.advancement, isNull);
      expect(r.startedAt, isNull);
      expect(r.completedAt, isNull);
    });
  });

  group('TournamentLeaderboardEntry', () {
    test('fromJson parses all fields', () {
      final json = {
        'rank': 1,
        'user_id': 'u1',
        'username': 'Alice',
        'final_score': 680,
        'final_x_count': 15,
        'status': 'active',
      };

      final e = TournamentLeaderboardEntry.fromJson(json);

      expect(e.rank, 1);
      expect(e.userId, 'u1');
      expect(e.username, 'Alice');
      expect(e.finalScore, 680);
      expect(e.finalXCount, 15);
      expect(e.status, 'active');
    });

    test('fromJson handles nullable score', () {
      final json = {
        'rank': 3,
        'user_id': 'u3',
        'status': 'registered',
      };

      final e = TournamentLeaderboardEntry.fromJson(json);

      expect(e.username, isNull);
      expect(e.finalScore, isNull);
      expect(e.finalXCount, isNull);
    });
  });

  group('TournamentRoundScore', () {
    test('fromJson parses all fields', () {
      final json = {
        'id': 'rs1',
        'round_id': 'r1',
        'participant_id': 'p1',
        'user_id': 'u1',
        'username': 'Alice',
        'session_id': 's1',
        'score': 340,
        'x_count': 8,
        'rank_in_round': 1,
        'advanced': true,
      };

      final s = TournamentRoundScore.fromJson(json);

      expect(s.id, 'rs1');
      expect(s.roundId, 'r1');
      expect(s.participantId, 'p1');
      expect(s.userId, 'u1');
      expect(s.username, 'Alice');
      expect(s.sessionId, 's1');
      expect(s.score, 340);
      expect(s.xCount, 8);
      expect(s.rankInRound, 1);
      expect(s.advanced, true);
    });

    test('fromJson defaults advanced to false', () {
      final json = {
        'id': 'rs2',
        'round_id': 'r1',
        'participant_id': 'p2',
        'user_id': 'u2',
        'status': 'pending',
      };

      final s = TournamentRoundScore.fromJson(json);

      expect(s.advanced, false);
      expect(s.score, isNull);
      expect(s.rankInRound, isNull);
    });
  });

  group('TournamentContext', () {
    test('stores all fields', () {
      const ctx = TournamentContext(
        clubId: 'c1',
        tournamentId: 't1',
        roundId: 'r1',
        tournamentName: 'Spring Championship',
        roundName: 'Qualifying',
      );

      expect(ctx.clubId, 'c1');
      expect(ctx.tournamentId, 't1');
      expect(ctx.roundId, 'r1');
      expect(ctx.tournamentName, 'Spring Championship');
      expect(ctx.roundName, 'Qualifying');
    });
  });
}
