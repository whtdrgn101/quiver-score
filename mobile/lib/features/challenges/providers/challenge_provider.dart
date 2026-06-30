import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/api_client.dart';
import '../models/challenge.dart';

final challengesProvider = FutureProvider<List<Challenge>>((ref) async {
  final api = ref.read(apiClientProvider);
  final response = await api.dio.get('/api/v1/challenges');
  final list = response.data as List;
  return list.map((json) => Challenge.fromJson(json as Map<String, dynamic>)).toList();
});

final challengeDetailProvider =
    FutureProvider.family<Challenge, String>((ref, id) async {
  final api = ref.read(apiClientProvider);
  final response = await api.dio.get('/api/v1/challenges/$id');
  return Challenge.fromJson(response.data as Map<String, dynamic>);
});

final challengeComparisonProvider = FutureProvider.family<
    ({Map<String, dynamic>? challenger, Map<String, dynamic>? challengee}),
    Challenge>((ref, challenge) async {
  final api = ref.read(apiClientProvider);
  Map<String, dynamic>? challengerSession;
  Map<String, dynamic>? challengeeSession;

  if (challenge.challengerSessionId != null) {
    try {
      final res =
          await api.dio.get('/api/v1/sessions/${challenge.challengerSessionId}');
      challengerSession = res.data as Map<String, dynamic>;
    } catch (_) {}
  }

  if (challenge.challengeeSessionId != null) {
    try {
      final res =
          await api.dio.get('/api/v1/sessions/${challenge.challengeeSessionId}');
      challengeeSession = res.data as Map<String, dynamic>;
    } catch (_) {}
  }

  return (challenger: challengerSession, challengee: challengeeSession);
});

final followingUsersProvider = FutureProvider<List<({String id, String username})>>((ref) async {
  final api = ref.read(apiClientProvider);
  final response = await api.dio.get('/api/v1/social/following');
  final list = response.data as List;
  return list.map((item) {
    final map = item as Map<String, dynamic>;
    return (
      id: map['following_id'] as String,
      username: (map['following_username'] as String?) ?? 'User',
    );
  }).toList();
});

Future<Challenge> createChallenge(
  WidgetRef ref, {
  required String challengeeId,
  required String templateId,
  int? expiresInHours,
}) async {
  final api = ref.read(apiClientProvider);
  final response = await api.dio.post(
    '/api/v1/challenges',
    data: {
      'challengee_id': challengeeId,
      'template_id': templateId,
      if (expiresInHours != null) 'expires_in_hours': expiresInHours,
    },
  );
  return Challenge.fromJson(response.data as Map<String, dynamic>);
}

Future<Challenge> acceptChallenge(WidgetRef ref, String challengeId) async {
  final api = ref.read(apiClientProvider);
  final response = await api.dio.post('/api/v1/challenges/$challengeId/accept');
  return Challenge.fromJson(response.data as Map<String, dynamic>);
}

Future<Challenge> declineChallenge(WidgetRef ref, String challengeId) async {
  final api = ref.read(apiClientProvider);
  final response = await api.dio.post('/api/v1/challenges/$challengeId/decline');
  return Challenge.fromJson(response.data as Map<String, dynamic>);
}

Future<Challenge> submitChallengeScore(
  WidgetRef ref, {
  required String challengeId,
  required String sessionId,
}) async {
  final api = ref.read(apiClientProvider);
  final response = await api.dio.post(
    '/api/v1/challenges/$challengeId/submit-score',
    data: {'session_id': sessionId},
  );
  return Challenge.fromJson(response.data as Map<String, dynamic>);
}

final userIdProvider = FutureProvider<String?>((ref) async {
  final storage = ref.read(secureStorageProvider);
  return storage.getUserId();
});
