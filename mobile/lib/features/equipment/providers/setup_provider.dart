import 'dart:developer' as dev;

import 'package:drift/drift.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/api_client.dart';
import '../../../core/database/database.dart';
import '../models/setup.dart';

final setupListProvider =
    AsyncNotifierProvider<SetupListNotifier, List<SetupSummary>>(
        SetupListNotifier.new);

class SetupListNotifier extends AsyncNotifier<List<SetupSummary>> {
  @override
  Future<List<SetupSummary>> build() => _fetch();

  Future<List<SetupSummary>> _fetch() async {
    final db = ref.read(databaseProvider);

    final cached = await (db.select(db.setupCache)
          ..orderBy([(t) => OrderingTerm.asc(t.name)]))
        .get();
    if (cached.isNotEmpty) {
      state = AsyncData(cached.map(_fromCache).toList());
    }

    try {
      final api = ref.read(apiClientProvider);
      final response = await api.dio.get('/api/v1/setups');
      final list = (response.data as List)
          .map((j) => SetupSummary.fromJson(j as Map<String, dynamic>))
          .toList();

      await _updateCache(db, list);
      return list..sort((a, b) => a.name.compareTo(b.name));
    } catch (e) {
      dev.log('Setups: server fetch failed: $e', name: 'SetupProvider');
      if (cached.isNotEmpty) return cached.map(_fromCache).toList();
      rethrow;
    }
  }

  Future<SetupDetail> create(Map<String, dynamic> data) async {
    final api = ref.read(apiClientProvider);
    final response = await api.dio.post('/api/v1/setups', data: data);
    final detail = SetupDetail.fromJson(response.data as Map<String, dynamic>);

    final db = ref.read(databaseProvider);
    await db.into(db.setupCache).insert(_summaryCompanion(SetupSummary(
      id: detail.id,
      name: detail.name,
      description: detail.description,
      equipmentCount: detail.equipment.length,
      createdAt: detail.createdAt,
    )));

    ref.invalidate(setupListProvider);
    return detail;
  }

  Future<void> delete(String id) async {
    final api = ref.read(apiClientProvider);
    await api.dio.delete('/api/v1/setups/$id');

    final db = ref.read(databaseProvider);
    await (db.delete(db.setupCache)..where((t) => t.id.equals(id))).go();
    await (db.delete(db.setupEquipmentCache)
          ..where((t) => t.setupId.equals(id)))
        .go();

    final items = <SetupSummary>[...state.valueOrNull ?? []];
    items.removeWhere((s) => s.id == id);
    state = AsyncData(items);
  }

  Future<void> refresh() async {
    final previous = state.valueOrNull;
    state = previous != null
        ? AsyncLoading<List<SetupSummary>>()
            .copyWithPrevious(AsyncData(previous))
        : const AsyncLoading();
    state = await AsyncValue.guard(_fetch);
  }

  SetupSummary _fromCache(SetupCacheData c) => SetupSummary(
        id: c.id,
        name: c.name,
        description: c.description,
        equipmentCount: c.equipmentCount,
        createdAt: c.createdAt,
      );

  SetupCacheCompanion _summaryCompanion(SetupSummary s) =>
      SetupCacheCompanion.insert(
        id: s.id,
        name: s.name,
        description: Value(s.description),
        equipmentCount: Value(s.equipmentCount),
        createdAt: s.createdAt,
        cachedAt: DateTime.now(),
      );

  Future<void> _updateCache(AppDatabase db, List<SetupSummary> items) async {
    await db.delete(db.setupCache).go();
    for (final item in items) {
      await db.into(db.setupCache).insert(_summaryCompanion(item));
    }
  }
}

final setupDetailProvider =
    FutureProvider.family<SetupDetail, String>((ref, setupId) async {
  final api = ref.read(apiClientProvider);
  final response = await api.dio.get('/api/v1/setups/$setupId');
  return SetupDetail.fromJson(response.data as Map<String, dynamic>);
});

Future<SetupDetail> updateSetup(
    WidgetRef ref, String id, Map<String, dynamic> data) async {
  final api = ref.read(apiClientProvider);
  final response = await api.dio.put('/api/v1/setups/$id', data: data);
  final detail = SetupDetail.fromJson(response.data as Map<String, dynamic>);
  ref.invalidate(setupDetailProvider(id));
  ref.invalidate(setupListProvider);
  return detail;
}

Future<void> addEquipmentToSetup(
    WidgetRef ref, String setupId, String equipmentId) async {
  final api = ref.read(apiClientProvider);
  await api.dio.post('/api/v1/setups/$setupId/equipment/$equipmentId');
  ref.invalidate(setupDetailProvider(setupId));
  ref.invalidate(setupListProvider);
}

Future<void> removeEquipmentFromSetup(
    WidgetRef ref, String setupId, String equipmentId) async {
  final api = ref.read(apiClientProvider);
  await api.dio.delete('/api/v1/setups/$setupId/equipment/$equipmentId');
  ref.invalidate(setupDetailProvider(setupId));
  ref.invalidate(setupListProvider);
}
