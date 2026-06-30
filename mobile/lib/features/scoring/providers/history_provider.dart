import 'dart:developer' as dev;

import 'package:drift/drift.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/api_client.dart';
import '../../../core/database/database.dart';

/// Watches the local sessions table for changes (insert / update / delete).
/// Emitting a new value causes [historyProvider] to rebuild automatically.
final _localSessionsChangeProvider = StreamProvider<List<ScoringSessionsLocalData>>((ref) {
  final db = ref.watch(databaseProvider);
  return (db.select(db.scoringSessionsLocal)
        ..orderBy([(t) => OrderingTerm.desc(t.startedAt)]))
      .watch();
});

/// Merges local completed/abandoned sessions with server-side data
final historyProvider =
    AsyncNotifierProvider<HistoryNotifier, List<SessionSummary>>(
        HistoryNotifier.new);

class HistoryNotifier extends AsyncNotifier<List<SessionSummary>> {
  @override
  Future<List<SessionSummary>> build() {
    // Watch the local sessions stream so we auto-rebuild when local data changes
    ref.watch(_localSessionsChangeProvider);
    return _fetch();
  }

  Future<List<SessionSummary>> _fetch() async {
    // Always load local sessions first (offline-first)
    final localSessions = await _loadLocal();

    // Try to merge with server data
    List<SessionSummary> serverSessions = [];
    try {
      serverSessions = await _loadServer();
    } catch (e) {
      dev.log('History: server fetch failed: $e', name: 'HistoryProvider');
    }

    return _merge(localSessions, serverSessions);
  }

  Future<List<SessionSummary>> _loadLocal() async {
    final db = ref.read(databaseProvider);
    final sessions = await (db.select(db.scoringSessionsLocal)
          ..orderBy([(t) => OrderingTerm.desc(t.startedAt)]))
        .get();

    // Look up template names
    final templateIds = sessions.map((s) => s.templateId).toSet();
    final templates = templateIds.isEmpty
        ? <RoundTemplate>[]
        : await (db.select(db.roundTemplates)
              ..where((t) => t.id.isIn(templateIds)))
            .get();
    final templateNameMap = {for (final t in templates) t.id: t.name};

    return sessions
        .map((s) => SessionSummary(
              id: s.serverId ?? s.id,
              localId: s.id,
              templateName: templateNameMap[s.templateId],
              status: s.status,
              totalScore: s.totalScore,
              totalXCount: s.totalXCount,
              totalArrows: s.totalArrows,
              startedAt: s.startedAt,
              completedAt: s.completedAt,
              synced: s.synced,
            ))
        .toList();
  }

  Future<List<SessionSummary>> _loadServer() async {
    final api = ref.read(apiClientProvider);
    final response = await api.dio.get('/api/v1/sessions');
    final list = response.data as List;
    return list
        .map((j) => SessionSummary.fromJson(j as Map<String, dynamic>))
        .toList();
  }

  /// Merge local and server sessions, preferring local data for duplicates.
  /// Matches by server ID first, then by startedAt timestamp to catch sessions
  /// whose serverId hasn't been set locally yet.
  List<SessionSummary> _merge(
    List<SessionSummary> local,
    List<SessionSummary> server,
  ) {
    final seen = <String>{};
    final merged = <SessionSummary>[];
    final localStartTimes = <int>{};

    for (final s in local) {
      seen.add(s.id);
      if (s.localId != null) seen.add(s.localId!);
      localStartTimes.add(s.startedAt.millisecondsSinceEpoch ~/ 1000);
      merged.add(s);
    }

    for (final s in server) {
      if (seen.contains(s.id)) continue;
      final serverStartSec = s.startedAt.millisecondsSinceEpoch ~/ 1000;
      if (localStartTimes.contains(serverStartSec)) continue;
      merged.add(s);
    }

    merged.sort((a, b) => b.startedAt.compareTo(a.startedAt));
    return merged;
  }

  Future<void> refresh() async {
    // Riverpod 3: keep the current data visible while reloading (guard only
    // replaces state once the fetch resolves). copyWithPrevious is now internal.
    state = await AsyncValue.guard(_fetch);
  }
}

class SessionSummary {
  final String id;
  final String? localId;
  final String? templateName;
  final String status;
  final int totalScore;
  final int totalXCount;
  final int totalArrows;
  final DateTime startedAt;
  final DateTime? completedAt;
  final bool synced;

  const SessionSummary({
    required this.id,
    this.localId,
    this.templateName,
    required this.status,
    required this.totalScore,
    required this.totalXCount,
    required this.totalArrows,
    required this.startedAt,
    this.completedAt,
    this.synced = true,
  });

  factory SessionSummary.fromJson(Map<String, dynamic> json) {
    return SessionSummary(
      id: json['id'] as String,
      templateName: json['template_name'] as String?,
      status: json['status'] as String,
      totalScore: json['total_score'] as int? ?? 0,
      totalXCount: json['total_x_count'] as int? ?? 0,
      totalArrows: json['total_arrows'] as int? ?? 0,
      startedAt: DateTime.parse(json['started_at'] as String),
      completedAt: json['completed_at'] != null
          ? DateTime.parse(json['completed_at'] as String)
          : null,
      synced: true, // Server data is always synced
    );
  }
}
