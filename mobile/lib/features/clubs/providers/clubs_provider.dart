import 'dart:developer' as dev;

import 'package:drift/drift.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/api_client.dart';
import '../../../core/database/database.dart';
import '../models/club.dart';

final clubsProvider =
    AsyncNotifierProvider<ClubsNotifier, List<Club>>(ClubsNotifier.new);

class ClubsNotifier extends AsyncNotifier<List<Club>> {
  @override
  Future<List<Club>> build() => _fetch();

  Future<List<Club>> _fetch() async {
    final db = ref.read(databaseProvider);

    final cached = await (db.select(db.clubCache)
          ..orderBy([(t) => OrderingTerm.asc(t.name)]))
        .get();
    if (cached.isNotEmpty) {
      state = AsyncData(cached.map(_fromCache).toList());
    }

    try {
      final api = ref.read(apiClientProvider);
      final response = await api.dio.get('/api/v1/clubs');
      final list = (response.data as List)
          .map((j) => Club.fromJson(j as Map<String, dynamic>))
          .toList();

      await _updateCache(db, list);
      return list..sort((a, b) => a.name.compareTo(b.name));
    } catch (e) {
      dev.log('Clubs: server fetch failed: $e', name: 'ClubsProvider');
      if (cached.isNotEmpty) return cached.map(_fromCache).toList();
      rethrow;
    }
  }

  Future<void> joinClub(String code) async {
    final api = ref.read(apiClientProvider);
    await api.dio.post('/api/v1/clubs/join/$code');
    await refresh();
  }

  Future<void> leaveClub(String clubId, String userId) async {
    final api = ref.read(apiClientProvider);
    await api.dio.delete('/api/v1/clubs/$clubId/members/$userId');

    final db = ref.read(databaseProvider);
    await (db.delete(db.clubCache)..where((t) => t.id.equals(clubId))).go();

    final items = <Club>[...state.valueOrNull ?? []];
    items.removeWhere((c) => c.id == clubId);
    state = AsyncData(items);
  }

  Future<void> refresh() async {
    final previous = state.valueOrNull;
    state = previous != null
        ? AsyncLoading<List<Club>>().copyWithPrevious(AsyncData(previous))
        : const AsyncLoading();
    state = await AsyncValue.guard(_fetch);
  }

  Club _fromCache(ClubCacheData c) => Club(
        id: c.id,
        name: c.name,
        description: c.description,
        avatar: c.avatar,
        ownerId: c.ownerId,
        memberCount: c.memberCount,
        myRole: c.myRole,
        createdAt: c.createdAt,
      );

  ClubCacheCompanion _toCompanion(Club c) => ClubCacheCompanion.insert(
        id: c.id,
        name: c.name,
        description: Value(c.description),
        avatar: Value(c.avatar),
        ownerId: c.ownerId,
        memberCount: Value(c.memberCount),
        myRole: Value(c.myRole),
        createdAt: c.createdAt,
        cachedAt: DateTime.now(),
      );

  Future<void> _updateCache(AppDatabase db, List<Club> items) async {
    await db.delete(db.clubCache).go();
    for (final item in items) {
      await db.into(db.clubCache).insert(_toCompanion(item));
    }
  }
}

final invitePreviewProvider =
    FutureProvider.family<Club, String>((ref, code) async {
  final api = ref.read(apiClientProvider);
  final response = await api.dio.get('/api/v1/clubs/join/$code');
  return Club.fromJson(response.data as Map<String, dynamic>);
});
