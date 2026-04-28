import 'dart:async';
import 'dart:convert';

import 'package:drift/drift.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../api/api_client.dart';
import '../database/database.dart';
import '../network/connectivity_service.dart';

final syncServiceProvider = Provider<SyncService>((ref) {
  return SyncService(
    db: ref.watch(databaseProvider),
    api: ref.watch(apiClientProvider),
    ref: ref,
  );
});

class SyncService {
  final AppDatabase db;
  final ApiClient api;
  final Ref ref;
  bool _syncing = false;

  SyncService({required this.db, required this.api, required this.ref}) {
    // Listen for connectivity changes and trigger sync
    ref.listen(connectivityProvider, (prev, next) {
      final isOnline = next.valueOrNull ?? false;
      if (isOnline) {
        syncPendingItems();
      }
    });
  }

  /// Enqueue a mutation to be synced later
  Future<void> enqueue({
    required String entityType,
    required String entityId,
    required String action,
    required Map<String, dynamic> payload,
  }) async {
    await db.into(db.syncQueue).insert(SyncQueueCompanion.insert(
      entityType: entityType,
      entityId: entityId,
      action: action,
      payloadJson: jsonEncode(payload),
      createdAt: DateTime.now(),
    ));

    // Try to sync immediately if online
    final isOnline = ref.read(connectivityProvider).valueOrNull ?? false;
    if (isOnline) {
      syncPendingItems();
    }
  }

  /// Process all pending sync items in order
  Future<void> syncPendingItems() async {
    if (_syncing) return;
    _syncing = true;

    try {
      final pending = await (db.select(db.syncQueue)
            ..where((t) => t.syncedAt.isNull())
            ..orderBy([(t) => OrderingTerm.asc(t.createdAt)]))
          .get();

      for (final item in pending) {
        try {
          await _processItem(item);

          // Mark as synced
          await (db.update(db.syncQueue)
                ..where((t) => t.id.equals(item.id)))
              .write(SyncQueueCompanion(
            syncedAt: Value(DateTime.now()),
          ));
        } catch (e) {
          // Record error and increment retry count
          await (db.update(db.syncQueue)
                ..where((t) => t.id.equals(item.id)))
              .write(SyncQueueCompanion(
            retryCount: Value(item.retryCount + 1),
            lastError: Value(e.toString()),
          ));

          // Stop processing if we hit an error (maintain order)
          if (item.retryCount >= 5) continue; // Skip items that failed too many times
          break;
        }
      }
    } finally {
      _syncing = false;
    }
  }

  Future<void> _processItem(SyncQueueData item) async {
    final payload = jsonDecode(item.payloadJson) as Map<String, dynamic>;

    switch (item.entityType) {
      case 'session':
        await _syncSession(item.action, item.entityId, payload);
      case 'end':
        await _syncEnd(item.action, item.entityId, payload);
      case 'image':
        await _syncImage(item.action, item.entityId, payload);
      default:
        throw Exception('Unknown entity type: ${item.entityType}');
    }
  }

  Future<void> _syncSession(
      String action, String entityId, Map<String, dynamic> payload) async {
    switch (action) {
      case 'create':
        final response = await api.dio.post('/api/v1/scoring', data: payload);
        final serverId = response.data['id'] as String;
        // Update local record with server ID
        await (db.update(db.scoringSessionsLocal)
              ..where((t) => t.id.equals(entityId)))
            .write(ScoringSessionsLocalCompanion(
          serverId: Value(serverId),
          synced: const Value(true),
        ));
      case 'complete':
        // Use server ID if available, otherwise client ID
        final session = await (db.select(db.scoringSessionsLocal)
              ..where((t) => t.id.equals(entityId)))
            .getSingle();
        final syncId = session.serverId ?? entityId;
        await api.dio
            .post('/api/v1/scoring/$syncId/complete', data: payload);
        await (db.update(db.scoringSessionsLocal)
              ..where((t) => t.id.equals(entityId)))
            .write(const ScoringSessionsLocalCompanion(
          synced: Value(true),
        ));
      case 'abandon':
        final session = await (db.select(db.scoringSessionsLocal)
              ..where((t) => t.id.equals(entityId)))
            .getSingle();
        final syncId = session.serverId ?? entityId;
        await api.dio.post('/api/v1/scoring/$syncId/abandon');
    }
  }

  Future<void> _syncEnd(
      String action, String entityId, Map<String, dynamic> payload) async {
    if (action == 'submit') {
      final sessionId = payload.remove('session_id') as String;
      // Look up server ID for the session
      final session = await (db.select(db.scoringSessionsLocal)
            ..where((t) => t.id.equals(sessionId)))
          .getSingle();
      final syncSessionId = session.serverId ?? sessionId;
      await api.dio.post(
        '/api/v1/scoring/$syncSessionId/ends',
        data: payload,
      );
    }
  }

  Future<void> _syncImage(
      String action, String entityId, Map<String, dynamic> payload) async {
    // Image upload will be implemented when the API endpoint is ready
    // For now, mark as synced to not block the queue
  }

  /// Pull round templates from the API and update local DB
  Future<void> pullRoundTemplates() async {
    try {
      final response = await api.dio.get('/api/v1/rounds');
      final templates = response.data as List;

      await db.batch((batch) {
        for (final t in templates) {
          batch.insert(
            db.roundTemplates,
            RoundTemplatesCompanion.insert(
              id: t['id'] as String,
              name: t['name'] as String,
              organization: t['organization'] as String,
              description: Value(t['description'] as String?),
              isOfficial: Value(t['is_official'] as bool? ?? false),
              syncedAt: DateTime.now(),
            ),
            mode: InsertMode.insertOrReplace,
          );

          // Sync stages for this template
          final stages = t['stages'] as List? ?? [];
          for (var i = 0; i < stages.length; i++) {
            final s = stages[i];
            batch.insert(
              db.stages,
              StagesCompanion.insert(
                id: s['id'] as String,
                templateId: t['id'] as String,
                name: s['name'] as String,
                distance: Value(s['distance'] as String?),
                numEnds: s['num_ends'] as int,
                arrowsPerEnd: s['arrows_per_end'] as int,
                allowedValues: jsonEncode(s['allowed_values']),
                valueScoreMap: jsonEncode(s['value_score_map']),
                maxScorePerArrow: s['max_score_per_arrow'] as int,
                stageOrder: i,
              ),
              mode: InsertMode.insertOrReplace,
            );
          }
        }
      });
    } catch (_) {
      // Offline — use cached data
    }
  }
}

String jsonEncode(Object? value) => json.encode(value);
