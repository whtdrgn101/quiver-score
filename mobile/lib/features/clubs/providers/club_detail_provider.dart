import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/api_client.dart';
import '../models/club.dart';
import '../models/club_event.dart';
import '../models/club_team.dart';

final clubDetailProvider =
    FutureProvider.family<ClubDetail, String>((ref, clubId) async {
  final api = ref.read(apiClientProvider);
  final response = await api.dio.get('/api/v1/clubs/$clubId');
  return ClubDetail.fromJson(response.data as Map<String, dynamic>);
});

final clubLeaderboardProvider =
    FutureProvider.family<List<LeaderboardGroup>, String>(
        (ref, clubId) async {
  final api = ref.read(apiClientProvider);
  final response = await api.dio.get('/api/v1/clubs/$clubId/leaderboard');
  final list = response.data as List;
  return list
      .map((j) => LeaderboardGroup.fromJson(j as Map<String, dynamic>))
      .toList();
});

final clubActivityProvider =
    FutureProvider.family<List<ActivityItem>, String>((ref, clubId) async {
  final api = ref.read(apiClientProvider);
  final response = await api.dio
      .get('/api/v1/clubs/$clubId/activity', queryParameters: {'limit': 50});
  final list = response.data as List;
  return list
      .map((j) => ActivityItem.fromJson(j as Map<String, dynamic>))
      .toList();
});

final clubEventsProvider =
    FutureProvider.family<List<ClubEvent>, String>((ref, clubId) async {
  final api = ref.read(apiClientProvider);
  final response = await api.dio.get('/api/v1/clubs/$clubId/events');
  final list = response.data as List;
  return list
      .map((j) => ClubEvent.fromJson(j as Map<String, dynamic>))
      .toList();
});

final clubEventDetailProvider = FutureProvider.family<ClubEvent,
    ({String clubId, String eventId})>((ref, params) async {
  final api = ref.read(apiClientProvider);
  final response = await api.dio
      .get('/api/v1/clubs/${params.clubId}/events/${params.eventId}');
  return ClubEvent.fromJson(response.data as Map<String, dynamic>);
});

final clubTeamsProvider =
    FutureProvider.family<List<ClubTeam>, String>((ref, clubId) async {
  final api = ref.read(apiClientProvider);
  final response = await api.dio.get('/api/v1/clubs/$clubId/teams');
  final list = response.data as List;
  return list
      .map((j) => ClubTeam.fromJson(j as Map<String, dynamic>))
      .toList();
});

final clubTeamDetailProvider = FutureProvider.family<ClubTeamDetail,
    ({String clubId, String teamId})>((ref, params) async {
  final api = ref.read(apiClientProvider);
  final response = await api.dio
      .get('/api/v1/clubs/${params.clubId}/teams/${params.teamId}');
  return ClubTeamDetail.fromJson(response.data as Map<String, dynamic>);
});

Future<void> rsvpEvent(
    WidgetRef ref, String clubId, String eventId, String status) async {
  final api = ref.read(apiClientProvider);
  await api.dio
      .post('/api/v1/clubs/$clubId/events/$eventId/rsvp', data: {'status': status});
  ref.invalidate(clubEventDetailProvider(
      (clubId: clubId, eventId: eventId)));
  ref.invalidate(clubEventsProvider(clubId));
}
