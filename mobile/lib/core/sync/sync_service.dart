import 'dart:async';
import 'dart:convert';
import 'dart:developer' as dev;
import 'dart:io';

import 'package:dio/dio.dart';
import 'package:drift/drift.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_image_compress/flutter_image_compress.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../api/api_client.dart';
import '../database/database.dart';
import '../network/connectivity_service.dart';

/// Whether a sync operation is currently in progress.
final syncInProgressProvider = StateProvider<bool>((ref) => false);

final syncServiceProvider = Provider<SyncService>((ref) {
  return SyncService(
    db: ref.watch(databaseProvider),
    api: ref.watch(apiClientProvider),
    ref: ref,
  );
});

class SyncResult {
  final int synced;
  final int failed;
  final String? lastError;

  const SyncResult({this.synced = 0, this.failed = 0, this.lastError});
}

class SyncService {
  final AppDatabase db;
  final ApiClient api;
  final Ref ref;
  bool _syncing = false;
  Timer? _retryTimer;
  static const _maxRetries = 10;

  SyncService({required this.db, required this.api, required this.ref}) {
    ref.listen(connectivityProvider, (prev, next) {
      final isOnline = next.valueOrNull ?? false;
      if (isOnline) {
        syncPendingItems();
      }
    });
  }

  Duration _backoffDelay(int retryCount) {
    final seconds = 2 << retryCount.clamp(0, 6); // 2, 4, 8, 16, 32, 64, 128
    return Duration(seconds: seconds);
  }

  void _scheduleRetry(int nextRetryCount, {Duration? overrideDelay}) {
    _retryTimer?.cancel();
    final delay = overrideDelay ?? _backoffDelay(nextRetryCount);
    dev.log('Sync: scheduling retry in ${delay.inSeconds}s', name: 'SyncService');
    _retryTimer = Timer(delay, () => syncPendingItems());
  }

  /// If [err] is a 429 with a Retry-After header, return the requested wait.
  /// Falls back to exponential backoff if the header is missing or unparseable.
  Duration? _retryAfterFromError(Object err) {
    if (err is! DioException) return null;
    if (err.response?.statusCode != 429) return null;
    final raw = err.response?.headers.value('retry-after');
    if (raw == null) return null;
    final seconds = int.tryParse(raw.trim());
    if (seconds == null || seconds <= 0) return null;
    return Duration(seconds: seconds);
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

  /// Process all pending sync items in order. Returns result with counts.
  Future<SyncResult> syncPendingItems() async {
    if (_syncing) return const SyncResult();
    _syncing = true;
    ref.read(syncInProgressProvider.notifier).state = true;

    int syncedCount = 0;
    int failedCount = 0;
    String? lastError;

    try {
      final pending = await (db.select(db.syncQueue)
            ..where((t) => t.syncedAt.isNull())
            ..orderBy([(t) => OrderingTerm.asc(t.createdAt)]))
          .get();

      // Sort by dependency order:
      //   session/create → end/submit → image/upload → session/complete|abandon
      // Images depend on ends (for server IDs), ends depend on sessions,
      // and completion must come after all ends/images are synced.
      int syncPriority(SyncQueueData item) {
        if (item.entityType == 'session' && item.action == 'create') return 0;
        if (item.entityType == 'end') return 1;
        if (item.entityType == 'image') return 2;
        if (item.entityType == 'session') return 3; // complete, abandon
        return 9;
      }

      pending.sort((a, b) {
        final pa = syncPriority(a);
        final pb = syncPriority(b);
        if (pa != pb) return pa.compareTo(pb);
        return a.createdAt.compareTo(b.createdAt);
      });

      dev.log('Sync: ${pending.length} pending items', name: 'SyncService');

      int highestRetry = 0;
      bool hasRetryable = false;
      Duration? retryAfterOverride;

      for (final item in pending) {
        dev.log(
          'Sync: processing ${item.entityType}/${item.action} '
          '(id: ${item.entityId}, retry: ${item.retryCount})',
          name: 'SyncService',
        );

        try {
          await _processItem(item);

          await (db.update(db.syncQueue)
                ..where((t) => t.id.equals(item.id)))
              .write(SyncQueueCompanion(
            syncedAt: Value(DateTime.now()),
          ));

          syncedCount++;
          dev.log('Sync: success ${item.entityType}/${item.action}',
              name: 'SyncService');
        } catch (e, stack) {
          final newRetry = item.retryCount + 1;
          final isPermanent = e is DioException &&
              e.response != null &&
              e.response!.statusCode != null &&
              e.response!.statusCode! >= 400 &&
              e.response!.statusCode! < 500 &&
              e.response!.statusCode != 429;

          if (isPermanent) {
            debugPrint('Sync PERMANENT FAIL: ${item.entityType}/${item.action} '
                'entity=${item.entityId}: ${e.response?.statusCode} — '
                'marking as resolved');
            await (db.update(db.syncQueue)
                  ..where((t) => t.id.equals(item.id)))
                .write(SyncQueueCompanion(
              syncedAt: Value(DateTime.now()),
              lastError: Value('permanent: $e'),
            ));
            failedCount++;
            lastError = e.toString();
            continue;
          }

          failedCount++;
          lastError = e.toString();
          debugPrint('Sync FAILED: ${item.entityType}/${item.action} '
              'entity=${item.entityId} (retry $newRetry/$_maxRetries): $e');
          dev.log(
            'Sync: FAILED ${item.entityType}/${item.action} '
            '(retry $newRetry/$_maxRetries): $e\n$stack',
            name: 'SyncService',
          );

          await (db.update(db.syncQueue)
                ..where((t) => t.id.equals(item.id)))
              .write(SyncQueueCompanion(
            retryCount: Value(newRetry),
            lastError: Value(e.toString()),
          ));

          if (newRetry >= _maxRetries) continue;
          hasRetryable = true;
          if (newRetry > highestRetry) highestRetry = newRetry;
          // If the server told us how long to wait (429 + Retry-After),
          // honor that instead of exponential backoff.
          retryAfterOverride ??= _retryAfterFromError(e);
          break;
        }
      }

      if (hasRetryable) {
        _scheduleRetry(highestRetry, overrideDelay: retryAfterOverride);
      }
    } finally {
      _syncing = false;
      ref.read(syncInProgressProvider.notifier).state = false;
    }

    dev.log('Sync: done (synced: $syncedCount, failed: $failedCount)',
        name: 'SyncService');
    return SyncResult(
        synced: syncedCount, failed: failedCount, lastError: lastError);
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
        final response = await api.dio.post('/api/v1/sessions', data: payload);
        final serverId = response.data['id'] as String;
        dev.log('Sync: session created on server, serverId=$serverId',
            name: 'SyncService');
        // Update local record with server ID
        await (db.update(db.scoringSessionsLocal)
              ..where((t) => t.id.equals(entityId)))
            .write(ScoringSessionsLocalCompanion(
          serverId: Value(serverId),
          synced: const Value(true),
        ));
      case 'complete':
        final session = await (db.select(db.scoringSessionsLocal)
              ..where((t) => t.id.equals(entityId)))
            .getSingle();
        final syncId = session.serverId ?? entityId;
        dev.log('Sync: completing session on server, syncId=$syncId',
            name: 'SyncService');
        await api.dio
            .post('/api/v1/sessions/$syncId/complete', data: payload);
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
        dev.log('Sync: abandoning session on server, syncId=$syncId',
            name: 'SyncService');
        await api.dio.post('/api/v1/sessions/$syncId/abandon');
        await (db.update(db.scoringSessionsLocal)
              ..where((t) => t.id.equals(entityId)))
            .write(const ScoringSessionsLocalCompanion(
          synced: Value(true),
        ));
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
      dev.log(
        'Sync: submitting end to server, syncSessionId=$syncSessionId',
        name: 'SyncService',
      );
      final response = await api.dio.post(
        '/api/v1/sessions/$syncSessionId/ends',
        data: payload,
      );

      // Capture server-assigned end ID
      final serverEndId = response.data['id'] as String?;
      if (serverEndId != null) {
        await (db.update(db.endsLocal)
              ..where((t) => t.id.equals(entityId)))
            .write(EndsLocalCompanion(
          serverId: Value(serverEndId),
        ));
        dev.log('Sync: end server ID=$serverEndId', name: 'SyncService');
      }
    }
  }

  Future<void> _syncImage(
      String action, String entityId, Map<String, dynamic> payload) async {
    if (action != 'upload') return;

    // Idempotency: if a previous run already uploaded this image and
    // captured a server attachment ID, skip the network round-trip.
    final existing = await (db.select(db.endImages)
          ..where((t) => t.id.equals(entityId)))
        .getSingleOrNull();
    if (existing == null) {
      dev.log('Sync: image row missing locally, skipping', name: 'SyncService');
      return;
    }
    if (existing.serverAttachmentId != null) {
      dev.log('Sync: image already uploaded (serverAttachmentId set), skipping',
          name: 'SyncService');
      await (db.update(db.endImages)
            ..where((t) => t.id.equals(entityId)))
          .write(const EndImagesCompanion(synced: Value(true)));
      return;
    }

    final endId = payload['end_id'] as String;
    final filePath = payload['file_path'] as String;

    // Look up server end ID — owner_id for owner_type=session_end is the
    // server's end UUID, not the local one. End sync must have run first.
    final end = await (db.select(db.endsLocal)
          ..where((t) => t.id.equals(endId)))
        .getSingle();

    final ownerId = end.serverId;
    if (ownerId == null) {
      throw StateError(
          'Cannot upload attachment: end ${end.id} has no serverId yet — '
          'session/end sync must run first');
    }

    final file = File(filePath);
    if (!await file.exists()) {
      dev.log('Sync: image file not found: $filePath', name: 'SyncService');
      return;
    }

    debugPrint('Sync: uploading attachment owner_type=session_end '
        'owner_id=$ownerId (localEnd=${end.id})');
    dev.log(
      'Sync: uploading attachment for end=$ownerId',
      name: 'SyncService',
    );

    // Compress before upload — server also resizes, but smaller body means
    // a much faster upload over cellular.
    final originalSize = await file.length();
    final compressed = await FlutterImageCompress.compressWithFile(
      filePath,
      minWidth: 1920,
      minHeight: 1920,
      quality: 70,
      format: CompressFormat.jpeg,
    );

    String uploadPath = filePath;
    if (compressed != null && compressed.length < originalSize) {
      final compressedFile = File('$filePath.compressed.jpg');
      await compressedFile.writeAsBytes(compressed);
      uploadPath = compressedFile.path;
      dev.log(
        'Sync: compressed image ${originalSize ~/ 1024}KB → '
        '${compressed.length ~/ 1024}KB',
        name: 'SyncService',
      );
    }

    final formData = FormData.fromMap({
      'image': await MultipartFile.fromFile(
        uploadPath,
        contentType: DioMediaType.parse('image/jpeg'),
      ),
    });

    final response = await api.dio.post(
      '/api/v1/attachments',
      queryParameters: {
        'owner_type': 'session_end',
        'owner_id': ownerId,
      },
      data: formData,
    );

    // Clean up compressed temp file
    if (uploadPath != filePath) {
      try {
        await File(uploadPath).delete();
      } catch (_) {}
    }

    final serverAttachmentId = response.data['id'] as String?;
    await (db.update(db.endImages)
          ..where((t) => t.id.equals(entityId)))
        .write(EndImagesCompanion(
      synced: const Value(true),
      serverAttachmentId: Value(serverAttachmentId),
    ));
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
