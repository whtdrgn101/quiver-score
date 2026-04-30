import 'dart:developer' as dev;

import 'package:drift/drift.dart' hide Column;
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';
import 'package:uuid/uuid.dart';

import '../../../core/api/api_client.dart';
import '../../../core/database/database.dart';
import '../../../core/sync/sync_service.dart';
import '../providers/history_provider.dart';
import '../providers/scoring_provider.dart';
import 'scoring_screen.dart';
import 'server_session_detail_screen.dart';
import 'session_detail_screen.dart';

class HistoryScreen extends ConsumerWidget {
  const HistoryScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final historyAsync = ref.watch(historyProvider);
    final isSyncing = ref.watch(syncInProgressProvider);
    final theme = Theme.of(context);

    return Column(
      children: [
        // Sync progress indicator
        if (isSyncing)
          LinearProgressIndicator(
            backgroundColor:
                theme.colorScheme.primaryContainer.withValues(alpha: 0.3),
          ),
        Expanded(
          child: historyAsync.when(
            data: (sessions) {
              if (sessions.isEmpty) {
                return Center(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Icon(Icons.history,
                          size: 64, color: theme.colorScheme.outline),
                      const SizedBox(height: 16),
                      Text('No rounds yet',
                          style: theme.textTheme.titleMedium),
                      const SizedBox(height: 8),
                      Text(
                        'Your rounds will appear here',
                        style: theme.textTheme.bodyMedium?.copyWith(
                          color: theme.colorScheme.onSurfaceVariant,
                        ),
                      ),
                    ],
                  ),
                );
              }

              return RefreshIndicator(
                onRefresh: () =>
                    ref.read(historyProvider.notifier).refresh(),
                child: ListView.builder(
                  padding: const EdgeInsets.all(16),
                  itemCount: sessions.length,
                  itemBuilder: (context, index) {
                    final session = sessions[index];
                    return _HistoryCard(session: session);
                  },
                ),
              );
            },
            loading: () => const Center(child: CircularProgressIndicator()),
            error: (e, _) => Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(Icons.cloud_off,
                      size: 48, color: theme.colorScheme.outline),
                  const SizedBox(height: 16),
                  Text('Could not load history',
                      style: theme.textTheme.titleMedium),
                  const SizedBox(height: 8),
                  Text(
                    'Check your connection and try again',
                    style: theme.textTheme.bodyMedium?.copyWith(
                      color: theme.colorScheme.onSurfaceVariant,
                    ),
                  ),
                  const SizedBox(height: 16),
                  FilledButton.tonal(
                    onPressed: () =>
                        ref.read(historyProvider.notifier).refresh(),
                    child: const Text('Retry'),
                  ),
                ],
              ),
            ),
          ),
        ),
      ],
    );
  }
}

class _HistoryCard extends ConsumerWidget {
  final SessionSummary session;

  const _HistoryCard({required this.session});

  Future<void> _navigateToSession(
      BuildContext context, WidgetRef ref) async {
    final localId = session.localId;
    final isInProgress = session.status == 'in_progress';
    final navigator = Navigator.of(context);

    if (isInProgress && localId != null) {
      _resumeSession(context, ref, localId);
    } else if (isInProgress) {
      final downloadedId =
          await _downloadServerSession(context, ref, session.id);
      if (downloadedId != null && context.mounted) {
        _resumeSession(context, ref, downloadedId);
      } else if (context.mounted) {
        navigator.push(
          MaterialPageRoute(
            builder: (_) =>
                ServerSessionDetailScreen(sessionId: session.id),
          ),
        );
      }
    } else if (localId != null) {
      navigator.push(
        MaterialPageRoute(
          builder: (_) => SessionDetailScreen(sessionId: localId),
        ),
      );
    } else {
      navigator.push(
        MaterialPageRoute(
          builder: (_) =>
              ServerSessionDetailScreen(sessionId: session.id),
        ),
      );
    }
  }

  static const _uuid = Uuid();

  Future<String?> _downloadServerSession(
      BuildContext context, WidgetRef ref, String serverId) async {
    try {
      final api = ref.read(apiClientProvider);
      final db = ref.read(databaseProvider);
      final response = await api.dio.get('/api/v1/sessions/$serverId');
      final data = response.data as Map<String, dynamic>;

      final templateId = data['template_id'] as String;
      final template = data['template'] as Map<String, dynamic>?;

      if (template != null) {
        await _ensureTemplateExists(db, templateId, template);
      }

      final localId = _uuid.v4();
      await db.into(db.scoringSessionsLocal).insert(
        ScoringSessionsLocalCompanion.insert(
          id: localId,
          templateId: templateId,
          setupProfileId: Value(data['setup_profile_id'] as String?),
          status: Value(data['status'] as String? ?? 'in_progress'),
          totalScore: Value(data['total_score'] as int? ?? 0),
          totalXCount: Value(data['total_x_count'] as int? ?? 0),
          totalArrows: Value(data['total_arrows'] as int? ?? 0),
          notes: Value(data['notes'] as String?),
          location: Value(data['location'] as String?),
          weather: Value(data['weather'] as String?),
          startedAt: DateTime.parse(data['started_at'] as String),
          completedAt: Value(data['completed_at'] != null
              ? DateTime.parse(data['completed_at'] as String)
              : null),
          synced: const Value(true),
          serverId: Value(serverId),
        ),
      );

      final ends = data['ends'] as List? ?? [];
      for (final endJson in ends) {
        final endMap = endJson as Map<String, dynamic>;
        final localEndId = _uuid.v4();
        await db.into(db.endsLocal).insert(EndsLocalCompanion.insert(
          id: localEndId,
          sessionId: localId,
          stageId: endMap['stage_id'] as String? ?? '',
          endNumber: endMap['end_number'] as int,
          endTotal: Value(endMap['end_total'] as int? ?? 0),
          createdAt: DateTime.parse(endMap['created_at'] as String),
          serverId: Value(endMap['id'] as String),
        ));

        final arrows = endMap['arrows'] as List? ?? [];
        for (final arrowJson in arrows) {
          final arrowMap = arrowJson as Map<String, dynamic>;
          await db.into(db.arrowsLocal).insert(ArrowsLocalCompanion.insert(
            id: _uuid.v4(),
            endId: localEndId,
            arrowNumber: arrowMap['arrow_number'] as int,
            scoreValue: arrowMap['score_value'] as String,
            scoreNumeric: arrowMap['score_numeric'] as int,
            xPos: Value(arrowMap['x_pos'] as double?),
            yPos: Value(arrowMap['y_pos'] as double?),
          ));
        }
      }

      dev.log('Downloaded session $serverId → local $localId '
          '(${ends.length} ends)', name: 'History');
      return localId;
    } catch (e) {
      dev.log('Failed to download session $serverId: $e', name: 'History');
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Could not download session: $e')),
        );
      }
      return null;
    }
  }

  Future<void> _ensureTemplateExists(
      AppDatabase db, String templateId, Map<String, dynamic> tmpl) async {
    final existing = await (db.select(db.roundTemplates)
          ..where((t) => t.id.equals(templateId)))
        .get();
    if (existing.isNotEmpty) return;

    await db.into(db.roundTemplates).insert(RoundTemplatesCompanion.insert(
      id: templateId,
      name: tmpl['name'] as String? ?? 'Unknown',
      organization: tmpl['organization'] as String? ?? '',
      description: Value(tmpl['description'] as String?),
      isOfficial: Value(tmpl['is_official'] as bool? ?? false),
      syncedAt: DateTime.now(),
    ));

    final stages = tmpl['stages'] as List? ?? [];
    for (var i = 0; i < stages.length; i++) {
      final s = stages[i] as Map<String, dynamic>;
      await db.into(db.stages).insert(StagesCompanion.insert(
        id: s['id'] as String,
        templateId: templateId,
        name: s['name'] as String,
        distance: Value(s['distance'] as String?),
        numEnds: s['num_ends'] as int,
        arrowsPerEnd: s['arrows_per_end'] as int,
        allowedValues: jsonEncode(s['allowed_values']),
        valueScoreMap: jsonEncode(s['value_score_map']),
        maxScorePerArrow: s['max_score_per_arrow'] as int,
        stageOrder: s['stage_order'] as int? ?? i,
      ));
    }
  }

  Future<void> _resumeSession(
      BuildContext context, WidgetRef ref, String localId) async {
    final navigator = Navigator.of(context);
    final notifier = ref.read(scoringProvider.notifier);
    final historyNotifier = ref.read(historyProvider.notifier);

    await notifier.loadSession(localId);

    await navigator.push(
      MaterialPageRoute(builder: (_) => const ScoringScreen()),
    );
    historyNotifier.refresh();
  }

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final theme = Theme.of(context);
    final dateFormat = DateFormat.yMMMd();
    final isInProgress = session.status == 'in_progress';

    final statusColor = switch (session.status) {
      'completed' => Colors.green,
      'abandoned' => Colors.grey,
      'in_progress' => Colors.orange,
      _ => Colors.orange,
    };

    final statusLabel = switch (session.status) {
      'in_progress' => 'In Progress',
      'abandoned' => 'Abandoned',
      _ => null,
    };

    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      shape: isInProgress
          ? RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(12),
              side: BorderSide(color: Colors.orange.shade300, width: 1.5),
            )
          : null,
      clipBehavior: Clip.antiAlias,
      child: InkWell(
        onTap: () => _navigateToSession(context, ref),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Row(
            children: [
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      children: [
                        Flexible(
                          child: Text(
                            session.templateName ?? 'Unknown Round',
                            style: theme.textTheme.titleSmall?.copyWith(
                              fontWeight: FontWeight.w600,
                            ),
                          ),
                        ),
                        if (isInProgress) ...[
                          const SizedBox(width: 8),
                          Icon(Icons.play_circle_outline,
                              size: 18, color: Colors.orange.shade700),
                        ],
                      ],
                    ),
                    const SizedBox(height: 4),
                    Row(
                      children: [
                        Container(
                          width: 8,
                          height: 8,
                          decoration: BoxDecoration(
                            color: statusColor,
                            shape: BoxShape.circle,
                          ),
                        ),
                        const SizedBox(width: 6),
                        if (statusLabel != null) ...[
                          Text(
                            statusLabel,
                            style: theme.textTheme.bodySmall?.copyWith(
                              color: statusColor,
                              fontWeight: FontWeight.w600,
                            ),
                          ),
                          const SizedBox(width: 6),
                          Text('·',
                              style: theme.textTheme.bodySmall?.copyWith(
                                color: theme.colorScheme.onSurfaceVariant,
                              )),
                          const SizedBox(width: 6),
                        ],
                        Text(
                          dateFormat.format(session.startedAt),
                          style: theme.textTheme.bodySmall?.copyWith(
                            color: theme.colorScheme.onSurfaceVariant,
                          ),
                        ),
                        if (!session.synced) ...[
                          const SizedBox(width: 6),
                          Icon(
                            Icons.cloud_off,
                            size: 14,
                            color: theme.colorScheme.outline,
                          ),
                        ],
                      ],
                    ),
                  ],
                ),
              ),
              Column(
                crossAxisAlignment: CrossAxisAlignment.end,
                children: [
                  Text(
                    '${session.totalScore}',
                    style: theme.textTheme.headlineSmall?.copyWith(
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  Text(
                    '${session.totalArrows} arrows · ${session.totalXCount}X',
                    style: theme.textTheme.bodySmall?.copyWith(
                      color: theme.colorScheme.onSurfaceVariant,
                    ),
                  ),
                  if (isInProgress) ...[
                    const SizedBox(height: 6),
                    FilledButton.tonal(
                      onPressed: () =>
                          _navigateToSession(context, ref),
                      style: FilledButton.styleFrom(
                        backgroundColor: Colors.orange.shade100,
                        foregroundColor: Colors.orange.shade800,
                        padding: const EdgeInsets.symmetric(
                            horizontal: 12, vertical: 4),
                        minimumSize: Size.zero,
                        tapTargetSize: MaterialTapTargetSize.shrinkWrap,
                      ),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Icon(Icons.play_arrow,
                              size: 14, color: Colors.orange.shade800),
                          const SizedBox(width: 2),
                          Text(
                            'Continue',
                            style: theme.textTheme.labelSmall?.copyWith(
                              color: Colors.orange.shade800,
                              fontWeight: FontWeight.w700,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }
}
