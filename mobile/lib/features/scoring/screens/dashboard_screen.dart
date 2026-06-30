import 'dart:developer' as dev;

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/api_client.dart';
import '../../clubs/screens/tournament_detail_screen.dart';
import 'new_session_screen.dart';

final _statsProvider = FutureProvider<Map<String, dynamic>>((ref) async {
  final api = ref.read(apiClientProvider);
  final response = await api.dio.get('/api/v1/sessions/stats');
  return response.data as Map<String, dynamic>;
});

final _activeTournamentsProvider =
    FutureProvider<List<Map<String, dynamic>>>((ref) async {
  final api = ref.read(apiClientProvider);
  try {
    final response = await api.dio.get('/api/v1/users/me/tournaments');
    return (response.data as List).cast<Map<String, dynamic>>();
  } catch (_) {
    return [];
  }
});

class DashboardScreen extends ConsumerWidget {
  const DashboardScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final statsAsync = ref.watch(_statsProvider);
    final tournamentsAsync = ref.watch(_activeTournamentsProvider);
    final theme = Theme.of(context);

    return Scaffold(
      body: RefreshIndicator(
        onRefresh: () async {
          ref.invalidate(_statsProvider);
          ref.invalidate(_activeTournamentsProvider);
        },
        child: ListView(
          padding: const EdgeInsets.all(16),
          children: [
            // New Round button
            FilledButton.icon(
              onPressed: () {
                Navigator.of(context).push(
                  MaterialPageRoute(builder: (_) => const NewSessionScreen()),
                );
              },
              icon: const Icon(Icons.add),
              label: const Text('New Round'),
              style: FilledButton.styleFrom(
                minimumSize: const Size.fromHeight(56),
                textStyle: theme.textTheme.titleMedium,
              ),
            ),
            const SizedBox(height: 24),

            // Active Tournaments
            tournamentsAsync.when(
              data: (tournaments) {
                if (tournaments.isEmpty) return const SizedBox.shrink();
                return Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'Active Tournaments',
                      style: theme.textTheme.titleMedium?.copyWith(
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                    const SizedBox(height: 12),
                    ...tournaments.map((t) => Card(
                      margin: const EdgeInsets.only(bottom: 8),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(12),
                        side: BorderSide(
                            color: Colors.amber.shade400, width: 2),
                      ),
                      child: Padding(
                        padding: const EdgeInsets.all(12),
                        child: Row(
                          children: [
                            Expanded(
                              child: Column(
                                crossAxisAlignment:
                                    CrossAxisAlignment.start,
                                children: [
                                  Text(
                                    t['tournament_name'] as String? ?? '',
                                    style: theme.textTheme.bodyMedium
                                        ?.copyWith(
                                      fontWeight: FontWeight.w600,
                                    ),
                                  ),
                                  const SizedBox(height: 4),
                                  Text(
                                    '${t['club_name']} · ${t['template_name'] ?? ''}',
                                    style: theme.textTheme.bodySmall
                                        ?.copyWith(
                                      color: theme
                                          .colorScheme.onSurfaceVariant,
                                    ),
                                  ),
                                ],
                              ),
                            ),
                            FilledButton(
                              onPressed: () {
                                final clubId = t['club_id'] as String?;
                                final tournamentId =
                                    t['tournament_id'] as String?;
                                if (clubId != null && tournamentId != null) {
                                  Navigator.of(context).push(
                                    MaterialPageRoute(
                                      builder: (_) =>
                                          TournamentDetailScreen(
                                        clubId: clubId,
                                        tournamentId: tournamentId,
                                      ),
                                    ),
                                  );
                                }
                              },
                              child: const Text('View'),
                            ),
                          ],
                        ),
                      ),
                    )),
                    const SizedBox(height: 24),
                  ],
                );
              },
              loading: () => const SizedBox.shrink(),
              error: (_, _) => const SizedBox.shrink(),
            ),

            // Stats
            statsAsync.when(
              data: (stats) => _StatsGrid(stats: stats),
              loading: () => const SizedBox(
                height: 200,
                child: Center(child: CircularProgressIndicator()),
              ),
              error: (e, _) {
                dev.log('Dashboard stats error: $e', name: 'Dashboard');
                return _OfflineStats();
              },
            ),
          ],
        ),
      ),
    );
  }
}

class _StatsGrid extends StatelessWidget {
  final Map<String, dynamic> stats;

  const _StatsGrid({required this.stats});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final completedSessions = stats['completed_sessions'] as int? ?? 0;
    final totalArrows = stats['total_arrows'] as int? ?? 0;
    final totalXCount = stats['total_x_count'] as int? ?? 0;
    final personalBestScore = stats['personal_best_score'] as int?;
    final personalBestTemplate = stats['personal_best_template'] as String?;
    final personalRecords =
        (stats['personal_records'] as List?)?.cast<Map<String, dynamic>>() ??
            [];

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // Top stat cards
        Row(
          children: [
            Expanded(
              child: _StatCard(
                label: 'Rounds',
                value: '$completedSessions',
                icon: Icons.track_changes,
                color: Colors.blue,
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: _StatCard(
                label: 'Arrows Shot',
                value: _formatNumber(totalArrows),
                icon: Icons.arrow_forward,
                color: Colors.green,
              ),
            ),
          ],
        ),
        const SizedBox(height: 12),
        Row(
          children: [
            Expanded(
              child: _StatCard(
                label: 'Total Xs',
                value: _formatNumber(totalXCount),
                icon: Icons.center_focus_strong,
                color: Colors.purple,
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: _StatCard(
                label: 'Personal Best',
                value: personalBestScore != null ? '$personalBestScore' : '-',
                subtitle: personalBestTemplate,
                icon: Icons.emoji_events,
                color: Colors.amber,
              ),
            ),
          ],
        ),

        // Personal records
        if (personalRecords.isNotEmpty) ...[
          const SizedBox(height: 24),
          Text(
            'Personal Records',
            style: theme.textTheme.titleMedium?.copyWith(
              fontWeight: FontWeight.bold,
            ),
          ),
          const SizedBox(height: 12),
          ...personalRecords.map((pr) {
            final name = pr['template_name'] as String? ?? '';
            final score = pr['score'] as int? ?? 0;
            final maxScore = pr['max_score'] as int? ?? 0;
            final pct = maxScore > 0 ? score / maxScore : 0.0;

            return Card(
              margin: const EdgeInsets.only(bottom: 8),
              child: Padding(
                padding: const EdgeInsets.all(12),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Expanded(
                          child: Text(
                            name,
                            style: theme.textTheme.bodyMedium?.copyWith(
                              fontWeight: FontWeight.w600,
                            ),
                            overflow: TextOverflow.ellipsis,
                          ),
                        ),
                        Text(
                          '$score / $maxScore',
                          style: theme.textTheme.bodyMedium?.copyWith(
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 8),
                    ClipRRect(
                      borderRadius: BorderRadius.circular(4),
                      child: LinearProgressIndicator(
                        value: pct,
                        minHeight: 6,
                        backgroundColor:
                            theme.colorScheme.surfaceContainerHighest,
                      ),
                    ),
                  ],
                ),
              ),
            );
          }),
        ],
      ],
    );
  }

  String _formatNumber(int n) {
    if (n >= 1000) return '${(n / 1000).toStringAsFixed(1)}k';
    return '$n';
  }
}

class _StatCard extends StatelessWidget {
  final String label;
  final String value;
  final String? subtitle;
  final IconData icon;
  final Color color;

  const _StatCard({
    required this.label,
    required this.value,
    this.subtitle,
    required this.icon,
    required this.color,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Icon(icon, color: color, size: 24),
            const SizedBox(height: 12),
            Text(
              value,
              style: theme.textTheme.headlineSmall?.copyWith(
                fontWeight: FontWeight.bold,
              ),
            ),
            Text(
              label,
              style: theme.textTheme.bodySmall?.copyWith(
                color: theme.colorScheme.onSurfaceVariant,
              ),
            ),
            if (subtitle != null)
              Text(
                subtitle!,
                style: theme.textTheme.bodySmall?.copyWith(
                  color: theme.colorScheme.onSurfaceVariant,
                  fontSize: 11,
                ),
                overflow: TextOverflow.ellipsis,
              ),
          ],
        ),
      ),
    );
  }
}

class _OfflineStats extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          children: [
            Icon(Icons.cloud_off, size: 40, color: theme.colorScheme.outline),
            const SizedBox(height: 12),
            Text(
              'Stats unavailable offline',
              style: theme.textTheme.bodyMedium?.copyWith(
                color: theme.colorScheme.onSurfaceVariant,
              ),
            ),
            const SizedBox(height: 4),
            Text(
              'Pull down to refresh when online',
              style: theme.textTheme.bodySmall?.copyWith(
                color: theme.colorScheme.outline,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
