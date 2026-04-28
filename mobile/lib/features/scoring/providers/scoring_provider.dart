import 'package:drift/drift.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:uuid/uuid.dart';

import '../../../core/database/database.dart';
import '../../../core/sync/sync_service.dart';

final scoringProvider =
    StateNotifierProvider<ScoringNotifier, ScoringState>((ref) {
  return ScoringNotifier(
    db: ref.watch(databaseProvider),
    syncService: ref.watch(syncServiceProvider),
  );
});

/// All local sessions, ordered by most recent
final sessionsListProvider = StreamProvider<List<ScoringSessionsLocalData>>((ref) {
  final db = ref.watch(databaseProvider);
  return (db.select(db.scoringSessionsLocal)
        ..orderBy([(t) => OrderingTerm.desc(t.startedAt)]))
      .watch();
});

class ScoringState {
  final ScoringSessionsLocalData? activeSession;
  final List<EndsLocalData> ends;
  final Map<String, List<ArrowsLocalData>> arrowsByEnd;
  final StagesCompanion? currentStage;

  const ScoringState({
    this.activeSession,
    this.ends = const [],
    this.arrowsByEnd = const {},
    this.currentStage,
  });

  ScoringState copyWith({
    ScoringSessionsLocalData? activeSession,
    List<EndsLocalData>? ends,
    Map<String, List<ArrowsLocalData>>? arrowsByEnd,
    StagesCompanion? currentStage,
  }) {
    return ScoringState(
      activeSession: activeSession ?? this.activeSession,
      ends: ends ?? this.ends,
      arrowsByEnd: arrowsByEnd ?? this.arrowsByEnd,
      currentStage: currentStage ?? this.currentStage,
    );
  }
}

class ScoringNotifier extends StateNotifier<ScoringState> {
  final AppDatabase db;
  final SyncService syncService;
  static const _uuid = Uuid();

  ScoringNotifier({required this.db, required this.syncService})
      : super(const ScoringState());

  /// Start a new scoring session
  Future<String> startSession({
    required String templateId,
    String? setupProfileId,
    String? notes,
    String? location,
    String? weather,
  }) async {
    final id = _uuid.v4();
    final now = DateTime.now();

    await db.into(db.scoringSessionsLocal).insert(
      ScoringSessionsLocalCompanion.insert(
        id: id,
        templateId: templateId,
        setupProfileId: Value(setupProfileId),
        notes: Value(notes),
        location: Value(location),
        weather: Value(weather),
        startedAt: now,
      ),
    );

    // Enqueue for sync
    await syncService.enqueue(
      entityType: 'session',
      entityId: id,
      action: 'create',
      payload: {
        'template_id': templateId,
        if (setupProfileId != null) 'setup_profile_id': setupProfileId,
        if (notes != null) 'notes': notes,
        if (location != null) 'location': location,
        if (weather != null) 'weather': weather,
      },
    );

    await loadSession(id);
    return id;
  }

  /// Load an existing session and its ends
  Future<void> loadSession(String sessionId) async {
    final session = await (db.select(db.scoringSessionsLocal)
          ..where((t) => t.id.equals(sessionId)))
        .getSingle();

    final ends = await (db.select(db.endsLocal)
          ..where((t) => t.sessionId.equals(sessionId))
          ..orderBy([(t) => OrderingTerm.asc(t.endNumber)]))
        .get();

    final arrowsByEnd = <String, List<ArrowsLocalData>>{};
    for (final end in ends) {
      final arrows = await (db.select(db.arrowsLocal)
            ..where((t) => t.endId.equals(end.id))
            ..orderBy([(t) => OrderingTerm.asc(t.arrowNumber)]))
          .get();
      arrowsByEnd[end.id] = arrows;
    }

    state = ScoringState(
      activeSession: session,
      ends: ends,
      arrowsByEnd: arrowsByEnd,
    );
  }

  /// Submit an end with arrow scores
  Future<void> submitEnd({
    required String stageId,
    required int endNumber,
    required List<ArrowInput> arrows,
    required Map<String, int> valueScoreMap,
  }) async {
    final session = state.activeSession;
    if (session == null) return;

    final endId = _uuid.v4();
    int endTotal = 0;

    // Calculate end total
    for (final arrow in arrows) {
      endTotal += valueScoreMap[arrow.scoreValue] ?? 0;
    }

    // Insert end
    await db.into(db.endsLocal).insert(EndsLocalCompanion.insert(
      id: endId,
      sessionId: session.id,
      stageId: stageId,
      endNumber: endNumber,
      endTotal: Value(endTotal),
      createdAt: DateTime.now(),
    ));

    // Insert arrows
    final arrowRows = <ArrowsLocalData>[];
    for (var i = 0; i < arrows.length; i++) {
      final arrowId = _uuid.v4();
      final scoreNumeric = valueScoreMap[arrows[i].scoreValue] ?? 0;
      await db.into(db.arrowsLocal).insert(ArrowsLocalCompanion.insert(
        id: arrowId,
        endId: endId,
        arrowNumber: i + 1,
        scoreValue: arrows[i].scoreValue,
        scoreNumeric: scoreNumeric,
        xPos: Value(arrows[i].xPos),
        yPos: Value(arrows[i].yPos),
      ));
      arrowRows.add(ArrowsLocalData(
        id: arrowId,
        endId: endId,
        arrowNumber: i + 1,
        scoreValue: arrows[i].scoreValue,
        scoreNumeric: scoreNumeric,
        xPos: arrows[i].xPos,
        yPos: arrows[i].yPos,
      ));
    }

    // Update session totals
    final newTotalScore = session.totalScore + endTotal;
    final newTotalArrows = session.totalArrows + arrows.length;
    final newXCount = session.totalXCount +
        arrows.where((a) => a.scoreValue == 'X').length;

    await (db.update(db.scoringSessionsLocal)
          ..where((t) => t.id.equals(session.id)))
        .write(ScoringSessionsLocalCompanion(
      totalScore: Value(newTotalScore),
      totalArrows: Value(newTotalArrows),
      totalXCount: Value(newXCount),
    ));

    // Enqueue for sync
    await syncService.enqueue(
      entityType: 'end',
      entityId: endId,
      action: 'submit',
      payload: {
        'session_id': session.id,
        'stage_id': stageId,
        'arrows': arrows
            .map((a) => {
                  'score_value': a.scoreValue,
                  if (a.xPos != null) 'x_pos': a.xPos,
                  if (a.yPos != null) 'y_pos': a.yPos,
                })
            .toList(),
      },
    );

    // Reload session state
    await loadSession(session.id);
  }

  /// Complete the current session
  Future<void> completeSession({
    String? notes,
    String? location,
    String? weather,
  }) async {
    final session = state.activeSession;
    if (session == null) return;

    final now = DateTime.now();
    await (db.update(db.scoringSessionsLocal)
          ..where((t) => t.id.equals(session.id)))
        .write(ScoringSessionsLocalCompanion(
      status: const Value('completed'),
      completedAt: Value(now),
      notes: Value(notes),
      location: Value(location),
      weather: Value(weather),
    ));

    await syncService.enqueue(
      entityType: 'session',
      entityId: session.id,
      action: 'complete',
      payload: {
        if (notes != null) 'notes': notes,
        if (location != null) 'location': location,
        if (weather != null) 'weather': weather,
      },
    );

    await loadSession(session.id);
  }

  /// Abandon the current session
  Future<void> abandonSession() async {
    final session = state.activeSession;
    if (session == null) return;

    await (db.update(db.scoringSessionsLocal)
          ..where((t) => t.id.equals(session.id)))
        .write(const ScoringSessionsLocalCompanion(
      status: Value('abandoned'),
    ));

    await syncService.enqueue(
      entityType: 'session',
      entityId: session.id,
      action: 'abandon',
      payload: {},
    );

    state = const ScoringState();
  }

  void clearActiveSession() {
    state = const ScoringState();
  }
}

class ArrowInput {
  final String scoreValue;
  final double? xPos;
  final double? yPos;

  const ArrowInput({
    required this.scoreValue,
    this.xPos,
    this.yPos,
  });
}
