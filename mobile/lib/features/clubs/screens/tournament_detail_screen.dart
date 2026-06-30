import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/api_client.dart';
import '../../more/providers/user_provider.dart';
import '../../scoring/providers/scoring_provider.dart';
import '../../scoring/screens/scoring_screen.dart';
import '../models/tournament.dart';
import '../providers/tournament_provider.dart';
import 'tournament_bracket_screen.dart';

class TournamentDetailScreen extends ConsumerStatefulWidget {
  final String clubId;
  final String tournamentId;

  const TournamentDetailScreen({
    super.key,
    required this.clubId,
    required this.tournamentId,
  });

  @override
  ConsumerState<TournamentDetailScreen> createState() =>
      _TournamentDetailScreenState();
}

class _TournamentDetailScreenState
    extends ConsumerState<TournamentDetailScreen> {
  String? _expandedRoundId;
  bool _actionLoading = false;

  ({String clubId, String tournamentId}) get _params =>
      (clubId: widget.clubId, tournamentId: widget.tournamentId);

  void _refresh() {
    ref.invalidate(tournamentDetailProvider(_params));
    ref.invalidate(tournamentRoundsProvider(_params));
    ref.invalidate(tournamentLeaderboardProvider(_params));
  }

  Future<void> _doAction(Future<void> Function() action) async {
    setState(() => _actionLoading = true);
    try {
      await action();
      _refresh();
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('$e')),
        );
      }
    } finally {
      if (mounted) setState(() => _actionLoading = false);
    }
  }

  Future<void> _register() async {
    final api = ref.read(apiClientProvider);
    await _doAction(() => api.dio.post(
          '/api/v1/clubs/${widget.clubId}/tournaments/${widget.tournamentId}/register',
        ));
  }

  Future<void> _withdraw() async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Withdraw'),
        content:
            const Text('Withdraw from this tournament? This cannot be undone.'),
        actions: [
          TextButton(
              onPressed: () => Navigator.pop(ctx, false),
              child: const Text('Cancel')),
          FilledButton(
            onPressed: () => Navigator.pop(ctx, true),
            style: FilledButton.styleFrom(
                backgroundColor: Theme.of(context).colorScheme.error),
            child: const Text('Withdraw'),
          ),
        ],
      ),
    );
    if (confirmed != true) return;

    final api = ref.read(apiClientProvider);
    await _doAction(() => api.dio.post(
          '/api/v1/clubs/${widget.clubId}/tournaments/${widget.tournamentId}/withdraw',
        ));
  }

  Future<void> _scoreRound(
    TournamentDetail tournament,
    TournamentRound round,
  ) async {
    final templateId = round.templateId ?? tournament.templateId;

    await ref
        .read(scoringProvider.notifier)
        .startSession(templateId: templateId);

    if (mounted) {
      await Navigator.of(context).push(
        MaterialPageRoute(
          builder: (_) => ScoringScreen(
            tournamentContext: TournamentContext(
              clubId: widget.clubId,
              tournamentId: tournament.id,
              roundId: round.id,
              tournamentName: tournament.name,
              roundName: round.name,
            ),
          ),
        ),
      );
      _refresh();
    }
  }

  @override
  Widget build(BuildContext context) {
    final detailAsync = ref.watch(tournamentDetailProvider(_params));
    final roundsAsync = ref.watch(tournamentRoundsProvider(_params));
    final leaderboardAsync = ref.watch(tournamentLeaderboardProvider(_params));
    final currentUser = ref.watch(currentUserProvider).valueOrNull;
    final theme = Theme.of(context);

    return Scaffold(
      appBar: AppBar(
        title: detailAsync.whenOrNull(data: (d) => Text(d.name)) ??
            const Text('Tournament'),
      ),
      body: detailAsync.when(
        data: (tournament) {
          final myParticipant = tournament.participants
              .where((p) => p.userId == currentUser?.id)
              .firstOrNull;
          final isRegistered = myParticipant != null;
          final canRegister = tournament.status == 'registration' && !isRegistered;
          final canWithdraw = isRegistered &&
              myParticipant.status != 'completed' &&
              myParticipant.status != 'withdrawn';

          return RefreshIndicator(
            onRefresh: () async => _refresh(),
            child: ListView(
              padding: const EdgeInsets.all(16),
              children: [
                _buildHeader(theme, tournament, canRegister, canWithdraw),
                const SizedBox(height: 16),
                _buildRoundsSection(
                  theme,
                  tournament,
                  roundsAsync,
                  currentUser?.id,
                  isRegistered && myParticipant.status == 'active',
                ),
                const SizedBox(height: 16),
                _buildLeaderboard(theme, leaderboardAsync, currentUser?.id),
                const SizedBox(height: 16),
                _buildParticipants(theme, tournament),
              ],
            ),
          );
        },
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (err, _) => Center(child: Text('Error: $err')),
      ),
    );
  }

  Widget _buildHeader(
    ThemeData theme,
    TournamentDetail tournament,
    bool canRegister,
    bool canWithdraw,
  ) {
    final (statusColor, statusBg) = _statusStyle(tournament.status);

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Expanded(
                  child: Text(
                    tournament.name,
                    style: theme.textTheme.titleLarge
                        ?.copyWith(fontWeight: FontWeight.bold),
                  ),
                ),
                Container(
                  padding:
                      const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                  decoration: BoxDecoration(
                    color: statusBg,
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Text(
                    tournament.status.replaceAll('_', ' '),
                    style: theme.textTheme.labelSmall?.copyWith(
                      color: statusColor,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                ),
              ],
            ),
            if (tournament.description != null) ...[
              const SizedBox(height: 8),
              Text(tournament.description!,
                  style: theme.textTheme.bodyMedium?.copyWith(
                    color: theme.colorScheme.onSurfaceVariant,
                  )),
            ],
            const SizedBox(height: 12),
            Wrap(
              spacing: 16,
              runSpacing: 4,
              children: [
                _infoChip(theme, Icons.track_changes,
                    tournament.templateName ?? 'Unknown'),
                if (tournament.organizerName != null)
                  _infoChip(
                      theme, Icons.person_outline, tournament.organizerName!),
                _infoChip(theme, Icons.people_outline,
                    '${tournament.participantCount} participants'),
                if (tournament.startDate != null)
                  _infoChip(theme, Icons.calendar_today,
                      _formatDate(tournament.startDate!)),
              ],
            ),
            if (canRegister || canWithdraw) ...[
              const SizedBox(height: 12),
              Row(
                children: [
                  if (canRegister)
                    FilledButton(
                      onPressed: _actionLoading ? null : _register,
                      child: const Text('Register'),
                    ),
                  if (canWithdraw)
                    FilledButton(
                      onPressed: _actionLoading ? null : _withdraw,
                      style: FilledButton.styleFrom(
                        backgroundColor: theme.colorScheme.error,
                      ),
                      child: const Text('Withdraw'),
                    ),
                ],
              ),
            ],
          ],
        ),
      ),
    );
  }

  Widget _buildRoundsSection(
    ThemeData theme,
    TournamentDetail tournament,
    AsyncValue<List<TournamentRound>> roundsAsync,
    String? currentUserId,
    bool canScore,
  ) {
    return Card(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Padding(
            padding: const EdgeInsets.all(16),
            child: Text('Rounds',
                style: theme.textTheme.titleMedium
                    ?.copyWith(fontWeight: FontWeight.bold)),
          ),
          roundsAsync.when(
            data: (rounds) {
              if (rounds.isEmpty) {
                return Padding(
                  padding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
                  child: Text('No rounds yet.',
                      style: theme.textTheme.bodyMedium?.copyWith(
                        color: theme.colorScheme.onSurfaceVariant,
                      )),
                );
              }

              final activeRound =
                  rounds.where((r) => r.status == 'in_progress').firstOrNull;

              return Column(
                children: [
                  if (canScore && activeRound != null)
                    Padding(
                      padding: const EdgeInsets.fromLTRB(16, 0, 16, 12),
                      child: SizedBox(
                        width: double.infinity,
                        child: FilledButton.icon(
                          onPressed: () =>
                              _scoreRound(tournament, activeRound),
                          icon: const Icon(Icons.play_arrow),
                          label: Text(
                              'Score Round ${activeRound.roundNumber}: ${activeRound.name}'),
                        ),
                      ),
                    ),
                  ...rounds.map((round) =>
                      _buildRoundRow(theme, tournament, round, currentUserId)),
                ],
              );
            },
            loading: () => const Padding(
              padding: EdgeInsets.all(16),
              child: Center(child: CircularProgressIndicator(strokeWidth: 2)),
            ),
            error: (_, _) => const SizedBox.shrink(),
          ),
        ],
      ),
    );
  }

  Widget _buildRoundRow(
      ThemeData theme, TournamentDetail tournament, TournamentRound round, String? currentUserId) {
    final isExpanded = _expandedRoundId == round.id;
    final (statusColor, statusBg) = _statusStyle(round.status);

    return Column(
      children: [
        InkWell(
          onTap: () {
            setState(() {
              _expandedRoundId = isExpanded ? null : round.id;
            });
          },
          child: Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
            child: Row(
              children: [
                CircleAvatar(
                  radius: 16,
                  backgroundColor: theme.colorScheme.surfaceContainerHighest,
                  child: Text(
                    '${round.roundNumber}',
                    style: theme.textTheme.labelMedium
                        ?.copyWith(fontWeight: FontWeight.bold),
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(round.name,
                          style: theme.textTheme.bodyMedium
                              ?.copyWith(fontWeight: FontWeight.w600)),
                      Text(
                        [
                          round.templateName,
                          if (round.advancement != null)
                            'Top ${round.advancement} advance',
                        ].whereType<String>().join(' · '),
                        style: theme.textTheme.bodySmall?.copyWith(
                          color: theme.colorScheme.onSurfaceVariant,
                        ),
                      ),
                    ],
                  ),
                ),
                Container(
                  padding:
                      const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                  decoration: BoxDecoration(
                    color: statusBg,
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Text(
                    round.status.replaceAll('_', ' '),
                    style: theme.textTheme.labelSmall?.copyWith(
                      color: statusColor,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                ),
                const SizedBox(width: 4),
                Icon(
                  isExpanded ? Icons.expand_less : Icons.expand_more,
                  color: theme.colorScheme.onSurfaceVariant,
                ),
              ],
            ),
          ),
        ),
        if (isExpanded)
          if (round.roundType == 'elimination')
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 0, 16, 12),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'Elimination round head-to-head matches.',
                    style: theme.textTheme.bodySmall?.copyWith(
                      color: theme.colorScheme.onSurfaceVariant,
                    ),
                  ),
                  const SizedBox(height: 8),
                  SizedBox(
                    width: double.infinity,
                    child: OutlinedButton.icon(
                      onPressed: () {
                        Navigator.push(
                          context,
                          MaterialPageRoute(
                            builder: (_) => TournamentBracketScreen(
                              clubId: widget.clubId,
                              tournament: tournament,
                            ),
                          ),
                        );
                      },
                      icon: const Icon(Icons.account_tree_outlined),
                      label: const Text('View Matchups & Bracket'),
                    ),
                  ),
                ],
              ),
            )
          else
            _RoundLeaderboard(
              clubId: widget.clubId,
              tournamentId: widget.tournamentId,
              roundId: round.id,
              advancement: round.advancement,
              roundCompleted: round.status == 'completed',
              currentUserId: currentUserId,
            ),
        const Divider(height: 1),
      ],
    );
  }

  Widget _buildLeaderboard(
    ThemeData theme,
    AsyncValue<List<TournamentLeaderboardEntry>> leaderboardAsync,
    String? currentUserId,
  ) {
    return Card(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Padding(
            padding: const EdgeInsets.all(16),
            child: Text('Overall Leaderboard',
                style: theme.textTheme.titleMedium
                    ?.copyWith(fontWeight: FontWeight.bold)),
          ),
          leaderboardAsync.when(
            data: (entries) {
              if (entries.isEmpty) {
                return Padding(
                  padding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
                  child: Text('No scores yet.',
                      style: theme.textTheme.bodyMedium?.copyWith(
                        color: theme.colorScheme.onSurfaceVariant,
                      )),
                );
              }
              return Column(
                children: entries
                    .map((e) => _LeaderboardRow(
                          rank: e.rank,
                          username: e.username ?? 'Unknown',
                          score: e.finalScore,
                          xCount: e.finalXCount,
                          isCurrentUser: e.userId == currentUserId,
                          trailing: e.finalScore == null
                              ? Text(e.status,
                                  style: theme.textTheme.bodySmall?.copyWith(
                                    color: theme.colorScheme.onSurfaceVariant,
                                  ))
                              : null,
                        ))
                    .toList(),
              );
            },
            loading: () => const Padding(
              padding: EdgeInsets.all(16),
              child: Center(child: CircularProgressIndicator(strokeWidth: 2)),
            ),
            error: (_, _) => const SizedBox.shrink(),
          ),
        ],
      ),
    );
  }

  Widget _buildParticipants(ThemeData theme, TournamentDetail tournament) {
    if (tournament.participants.isEmpty) return const SizedBox.shrink();

    return Card(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Padding(
            padding: const EdgeInsets.all(16),
            child: Text('Participants',
                style: theme.textTheme.titleMedium
                    ?.copyWith(fontWeight: FontWeight.bold)),
          ),
          ...tournament.participants.map((p) => ListTile(
                dense: true,
                title: Text(p.username ?? p.userId),
                trailing: _participantStatusChip(theme, p.status),
              )),
          const SizedBox(height: 8),
        ],
      ),
    );
  }

  Widget _participantStatusChip(ThemeData theme, String status) {
    final (color, bg) = switch (status) {
      'active' => (Colors.green.shade700, Colors.green.shade50),
      'withdrawn' => (Colors.red.shade700, Colors.red.shade50),
      'eliminated' => (Colors.orange.shade700, Colors.orange.shade50),
      _ => (Colors.grey.shade700, Colors.grey.shade100),
    };
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
      decoration: BoxDecoration(
        color: bg,
        borderRadius: BorderRadius.circular(12),
      ),
      child: Text(
        status,
        style:
            theme.textTheme.labelSmall?.copyWith(color: color, fontWeight: FontWeight.w600),
      ),
    );
  }

  Widget _infoChip(ThemeData theme, IconData icon, String text) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Icon(icon, size: 14, color: theme.colorScheme.onSurfaceVariant),
        const SizedBox(width: 4),
        Text(text,
            style: theme.textTheme.bodySmall?.copyWith(
              color: theme.colorScheme.onSurfaceVariant,
            )),
      ],
    );
  }

  (Color, Color) _statusStyle(String status) {
    return switch (status) {
      'registration' => (Colors.blue.shade700, Colors.blue.shade50),
      'in_progress' => (Colors.amber.shade800, Colors.amber.shade50),
      'completed' => (Colors.green.shade700, Colors.green.shade50),
      _ => (Colors.grey.shade700, Colors.grey.shade100),
    };
  }

  String _formatDate(DateTime date) {
    return '${date.month}/${date.day}/${date.year}';
  }
}

class _RoundLeaderboard extends ConsumerWidget {
  final String clubId;
  final String tournamentId;
  final String roundId;
  final int? advancement;
  final bool roundCompleted;
  final String? currentUserId;

  const _RoundLeaderboard({
    required this.clubId,
    required this.tournamentId,
    required this.roundId,
    this.advancement,
    required this.roundCompleted,
    this.currentUserId,
  });

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final lbAsync = ref.watch(roundLeaderboardProvider(
      (clubId: clubId, tournamentId: tournamentId, roundId: roundId),
    ));
    final theme = Theme.of(context);

    return lbAsync.when(
      data: (entries) {
        if (entries.isEmpty) {
          return Padding(
            padding: const EdgeInsets.fromLTRB(16, 0, 16, 12),
            child: Text('No scores yet.',
                style: theme.textTheme.bodySmall?.copyWith(
                  color: theme.colorScheme.onSurfaceVariant,
                )),
          );
        }
        return Padding(
          padding: const EdgeInsets.fromLTRB(16, 0, 16, 12),
          child: Column(
            children: entries
                .map((e) => _LeaderboardRow(
                      rank: e.rankInRound ?? 0,
                      username: e.username ?? 'Unknown',
                      score: e.score,
                      xCount: e.xCount,
                      isCurrentUser: e.userId == currentUserId,
                      trailing: roundCompleted
                          ? _advancementBadge(theme, e.advanced)
                          : null,
                    ))
                .toList(),
          ),
        );
      },
      loading: () => const Padding(
        padding: EdgeInsets.all(12),
        child: Center(child: CircularProgressIndicator(strokeWidth: 2)),
      ),
      error: (_, _) => const SizedBox.shrink(),
    );
  }

  Widget? _advancementBadge(ThemeData theme, bool advanced) {
    if (advancement == null) return null;
    final (text, color, bg) = advanced
        ? ('Advanced', Colors.green.shade700, Colors.green.shade50)
        : ('Eliminated', Colors.red.shade700, Colors.red.shade50);
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 1),
      decoration: BoxDecoration(
        color: bg,
        borderRadius: BorderRadius.circular(8),
      ),
      child: Text(text,
          style: theme.textTheme.labelSmall
              ?.copyWith(color: color, fontWeight: FontWeight.w600)),
    );
  }
}

class _LeaderboardRow extends StatelessWidget {
  final int rank;
  final String username;
  final int? score;
  final int? xCount;
  final bool isCurrentUser;
  final Widget? trailing;

  const _LeaderboardRow({
    required this.rank,
    required this.username,
    this.score,
    this.xCount,
    required this.isCurrentUser,
    this.trailing,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final bgColor =
        isCurrentUser ? theme.colorScheme.primaryContainer.withValues(alpha: 0.3) : null;

    return Container(
      color: bgColor,
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
      child: Row(
        children: [
          SizedBox(
            width: 28,
            child: Text(
              '$rank',
              style: theme.textTheme.bodyMedium?.copyWith(
                fontWeight: FontWeight.bold,
                color: rank <= 3 ? _rankColor(rank) : null,
              ),
            ),
          ),
          Expanded(
            child: Text(
              username,
              style: theme.textTheme.bodyMedium?.copyWith(
                fontWeight: isCurrentUser ? FontWeight.bold : FontWeight.normal,
              ),
            ),
          ),
          if (trailing != null) ...[
            trailing!,
            const SizedBox(width: 8),
          ],
          if (score != null) ...[
            Text(
              '$score',
              style: theme.textTheme.bodyMedium
                  ?.copyWith(fontWeight: FontWeight.bold),
            ),
            if (xCount != null && xCount! > 0)
              Text(
                ' (${xCount}X)',
                style: theme.textTheme.bodySmall?.copyWith(
                  color: theme.colorScheme.onSurfaceVariant,
                ),
              ),
          ],
        ],
      ),
    );
  }

  Color _rankColor(int rank) {
    return switch (rank) {
      1 => Colors.amber.shade700,
      2 => Colors.grey.shade500,
      3 => Colors.brown.shade400,
      _ => Colors.grey,
    };
  }
}
