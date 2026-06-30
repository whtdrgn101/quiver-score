import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/api_client.dart';
import '../models/tournament.dart';

final clubTournamentsProvider = FutureProvider.family<List<Tournament>, String>(
  (ref, clubId) async {
    final api = ref.read(apiClientProvider);
    final response = await api.dio.get('/api/v1/clubs/$clubId/tournaments');
    final list = response.data as List;
    return list
        .map((j) => Tournament.fromJson(j as Map<String, dynamic>))
        .toList();
  },
);

final tournamentDetailProvider =
    FutureProvider.family<
      TournamentDetail,
      ({String clubId, String tournamentId})
    >((ref, params) async {
      final api = ref.read(apiClientProvider);
      final response = await api.dio.get(
        '/api/v1/clubs/${params.clubId}/tournaments/${params.tournamentId}',
      );
      return TournamentDetail.fromJson(response.data as Map<String, dynamic>);
    });

final tournamentRoundsProvider =
    FutureProvider.family<
      List<TournamentRound>,
      ({String clubId, String tournamentId})
    >((ref, params) async {
      final api = ref.read(apiClientProvider);
      final response = await api.dio.get(
        '/api/v1/clubs/${params.clubId}/tournaments/${params.tournamentId}/rounds',
      );
      final list = response.data as List;
      return list
          .map((j) => TournamentRound.fromJson(j as Map<String, dynamic>))
          .toList();
    });

final tournamentLeaderboardProvider =
    FutureProvider.family<
      List<TournamentLeaderboardEntry>,
      ({String clubId, String tournamentId})
    >((ref, params) async {
      final api = ref.read(apiClientProvider);
      final response = await api.dio.get(
        '/api/v1/clubs/${params.clubId}/tournaments/${params.tournamentId}/leaderboard',
      );
      final list = response.data as List;
      return list
          .map(
            (j) =>
                TournamentLeaderboardEntry.fromJson(j as Map<String, dynamic>),
          )
          .toList();
    });

final roundLeaderboardProvider =
    FutureProvider.family<
      List<TournamentRoundScore>,
      ({String clubId, String tournamentId, String roundId})
    >((ref, params) async {
      final api = ref.read(apiClientProvider);
      final response = await api.dio.get(
        '/api/v1/clubs/${params.clubId}/tournaments/${params.tournamentId}'
        '/rounds/${params.roundId}/leaderboard',
      );
      final list = response.data as List;
      return list
          .map((j) => TournamentRoundScore.fromJson(j as Map<String, dynamic>))
          .toList();
    });

final tournamentMatchupsProvider =
    FutureProvider.family<
      List<TournamentMatchup>,
      ({String clubId, String tournamentId, String roundId})
    >((ref, params) async {
      final api = ref.read(apiClientProvider);
      final response = await api.dio.get(
        '/api/v1/clubs/${params.clubId}/tournaments/${params.tournamentId}'
        '/rounds/${params.roundId}/matchups',
      );
      final list = response.data as List;
      return list
          .map((j) => TournamentMatchup.fromJson(j as Map<String, dynamic>))
          .toList();
    });

Future<void> submitTournamentRoundScore(
  WidgetRef ref, {
  required String clubId,
  required String tournamentId,
  required String roundId,
  required String sessionId,
}) async {
  final api = ref.read(apiClientProvider);
  await api.dio.post(
    '/api/v1/clubs/$clubId/tournaments/$tournamentId'
    '/rounds/$roundId/submit-score?session_id=$sessionId',
  );
}

Future<void> submitMatchupScore(
  WidgetRef ref, {
  required String clubId,
  required String tournamentId,
  required String roundId,
  required String matchupId,
  required String sessionId,
}) async {
  final api = ref.read(apiClientProvider);
  await api.dio.post(
    '/api/v1/clubs/$clubId/tournaments/$tournamentId'
    '/rounds/$roundId/matchups/$matchupId/submit-score?session_id=$sessionId',
  );
}

Future<void> updateMatchupScore(
  WidgetRef ref, {
  required String clubId,
  required String tournamentId,
  required String roundId,
  required String matchupId,
  int? scoreA,
  int? scoreB,
  String? winnerId,
}) async {
  final api = ref.read(apiClientProvider);
  await api.dio.put(
    '/api/v1/clubs/$clubId/tournaments/$tournamentId'
    '/rounds/$roundId/matchups/$matchupId',
    data: {'score_a': ?scoreA, 'score_b': ?scoreB, 'winner_id': ?winnerId},
  );
}
