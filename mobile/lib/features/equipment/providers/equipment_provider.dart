import 'dart:convert';
import 'dart:developer' as dev;

import 'package:drift/drift.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/api_client.dart';
import '../../../core/database/database.dart';
import '../models/equipment.dart';

final equipmentProvider =
    AsyncNotifierProvider<EquipmentNotifier, List<Equipment>>(
        EquipmentNotifier.new);

class EquipmentNotifier extends AsyncNotifier<List<Equipment>> {
  @override
  Future<List<Equipment>> build() => _fetch();

  Future<List<Equipment>> _fetch() async {
    final db = ref.read(databaseProvider);

    // Load from cache first for instant UI
    final cached = await (db.select(db.equipmentCache)
          ..orderBy([(t) => OrderingTerm.asc(t.name)]))
        .get();
    if (cached.isNotEmpty) {
      state = AsyncData(cached.map(_fromCache).toList());
    }

    // Fetch from API
    try {
      final api = ref.read(apiClientProvider);
      final response = await api.dio.get('/api/v1/equipment');
      final list = (response.data as List)
          .map((j) => Equipment.fromJson(j as Map<String, dynamic>))
          .toList();

      await _updateCache(db, list);
      return list..sort((a, b) => a.name.compareTo(b.name));
    } catch (e) {
      dev.log('Equipment: server fetch failed: $e', name: 'EquipmentProvider');
      if (cached.isNotEmpty) return cached.map(_fromCache).toList();
      rethrow;
    }
  }

  Future<Equipment> create(Map<String, dynamic> data) async {
    final api = ref.read(apiClientProvider);
    final response = await api.dio.post('/api/v1/equipment', data: data);
    final item = Equipment.fromJson(response.data as Map<String, dynamic>);

    final db = ref.read(databaseProvider);
    await db.into(db.equipmentCache).insert(_toCompanion(item));

    state = AsyncData([...state.valueOrNull ?? [], item]
      ..sort((a, b) => a.name.compareTo(b.name)));
    return item;
  }

  Future<void> updateItem(String id, Map<String, dynamic> data) async {
    final api = ref.read(apiClientProvider);
    final response = await api.dio.put('/api/v1/equipment/$id', data: data);
    final updated = Equipment.fromJson(response.data as Map<String, dynamic>);

    final db = ref.read(databaseProvider);
    await (db.update(db.equipmentCache)..where((t) => t.id.equals(id)))
        .write(_toCompanion(updated));

    final items = <Equipment>[...state.valueOrNull ?? []];
    final idx = items.indexWhere((e) => e.id == id);
    if (idx >= 0) items[idx] = updated;
    state = AsyncData(items);
  }

  Future<void> delete(String id) async {
    final api = ref.read(apiClientProvider);
    await api.dio.delete('/api/v1/equipment/$id');

    final db = ref.read(databaseProvider);
    await (db.delete(db.equipmentCache)..where((t) => t.id.equals(id))).go();

    final items = <Equipment>[...state.valueOrNull ?? []];
    items.removeWhere((e) => e.id == id);
    state = AsyncData(items);
  }

  Future<void> refresh() async {
    final previous = state.valueOrNull;
    state = previous != null
        ? AsyncLoading<List<Equipment>>()
            .copyWithPrevious(AsyncData(previous))
        : const AsyncLoading();
    state = await AsyncValue.guard(_fetch);
  }

  Equipment _fromCache(EquipmentCacheData c) => Equipment(
        id: c.id,
        category: c.category,
        name: c.name,
        brand: c.brand,
        model: c.model,
        specs: c.specs != null
            ? jsonDecode(c.specs!) as Map<String, dynamic>?
            : null,
        notes: c.notes,
        createdAt: c.createdAt,
      );

  EquipmentCacheCompanion _toCompanion(Equipment e) =>
      EquipmentCacheCompanion.insert(
        id: e.id,
        category: e.category,
        name: e.name,
        brand: Value(e.brand),
        model: Value(e.model),
        specs: Value(e.specs != null ? jsonEncode(e.specs) : null),
        notes: Value(e.notes),
        createdAt: e.createdAt,
        cachedAt: DateTime.now(),
      );

  Future<void> _updateCache(AppDatabase db, List<Equipment> items) async {
    await db.delete(db.equipmentCache).go();
    for (final item in items) {
      await db.into(db.equipmentCache).insert(_toCompanion(item));
    }
  }
}

final equipmentStatsProvider =
    FutureProvider<List<EquipmentStats>>((ref) async {
  final api = ref.read(apiClientProvider);
  final response = await api.dio.get('/api/v1/equipment/stats');
  final list = response.data as List;
  return list
      .map((j) => EquipmentStats.fromJson(j as Map<String, dynamic>))
      .toList();
});
