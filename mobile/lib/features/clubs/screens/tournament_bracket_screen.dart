import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../more/providers/user_provider.dart';
import '../../scoring/providers/scoring_provider.dart';
import '../../scoring/screens/scoring_screen.dart';
import '../models/tournament.dart';
import '../providers/tournament_provider.dart';

class TournamentBracketScreen extends ConsumerStatefulWidget {
  final String clubId;
  final TournamentDetail tournament;

  const TournamentBracketScreen({
    super.key,
    required this.clubId,
    required this.tournament,
  });

  @override
  ConsumerState<TournamentBracketScreen> createState() =>
      _TournamentBracketScreenState();
}

class _TournamentBracketScreenState
    extends ConsumerState<TournamentBracketScreen> {
  @override
  Widget build(BuildContext context) {
    final roundsAsync = ref.watch(
      tournamentRoundsProvider((
        clubId: widget.clubId,
        tournamentId: widget.tournament.id,
      )),
    );
    final theme = Theme.of(context);

    return Scaffold(
      appBar: AppBar(title: Text('${widget.tournament.name} Bracket')),
      body: roundsAsync.when(
        data: (rounds) {
          final elimRounds = rounds
              .where((r) => r.roundType == 'elimination')
              .toList();

          if (elimRounds.isEmpty) {
            return Center(
              child: Padding(
                padding: const EdgeInsets.all(24),
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Icon(
                      Icons.account_tree_outlined,
                      size: 64,
                      color: theme.colorScheme.onSurfaceVariant.withValues(
                        alpha: 0.5,
                      ),
                    ),
                    const SizedBox(height: 16),
                    Text(
                      'No Elimination Rounds Yet',
                      style: theme.textTheme.titleMedium,
                    ),
                    const SizedBox(height: 8),
                    Text(
                      'Matchups will appear once the organizer adds and starts elimination rounds.',
                      textAlign: TextAlign.center,
                      style: theme.textTheme.bodyMedium?.copyWith(
                        color: theme.colorScheme.onSurfaceVariant,
                      ),
                    ),
                  ],
                ),
              ),
            );
          }

          return DefaultTabController(
            length: elimRounds.length,
            child: Column(
              children: [
                TabBar(
                  isScrollable: elimRounds.length > 3,
                  tabAlignment: elimRounds.length > 3
                      ? TabAlignment.start
                      : null,
                  tabs: elimRounds.map((r) => Tab(text: r.name)).toList(),
                ),
                Expanded(
                  child: TabBarView(
                    children: elimRounds
                        .map(
                          (round) => _BracketRoundView(
                            clubId: widget.clubId,
                            tournament: widget.tournament,
                            round: round,
                          ),
                        )
                        .toList(),
                  ),
                ),
              ],
            ),
          );
        },
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (err, _) => Center(child: Text('Error loading rounds: $err')),
      ),
    );
  }
}

class _BracketRoundView extends ConsumerWidget {
  final String clubId;
  final TournamentDetail tournament;
  final TournamentRound round;

  const _BracketRoundView({
    required this.clubId,
    required this.tournament,
    required this.round,
  });

  Future<void> _scoreMatch(
    BuildContext context,
    WidgetRef ref,
    TournamentMatchup matchup,
  ) async {
    final templateId = round.templateId ?? tournament.templateId;
    await ref
        .read(scoringProvider.notifier)
        .startSession(templateId: templateId);

    if (context.mounted) {
      await Navigator.of(context).push(
        MaterialPageRoute(
          builder: (_) => ScoringScreen(
            tournamentContext: TournamentContext(
              clubId: clubId,
              tournamentId: tournament.id,
              roundId: round.id,
              tournamentName: tournament.name,
              roundName: round.name,
              matchupId: matchup.id,
            ),
          ),
        ),
      );
      ref.invalidate(
        tournamentMatchupsProvider((
          clubId: clubId,
          tournamentId: tournament.id,
          roundId: round.id,
        )),
      );
    }
  }

  void _showEditMatchDialog(
    BuildContext context,
    WidgetRef ref,
    TournamentMatchup matchup,
  ) {
    final scoreAController = TextEditingController(
      text: matchup.scoreA?.toString() ?? '',
    );
    final scoreBController = TextEditingController(
      text: matchup.scoreB?.toString() ?? '',
    );
    String? selectedWinnerId = matchup.winnerId;

    showDialog(
      context: context,
      builder: (ctx) => StatefulBuilder(
        builder: (context, setDialogState) => AlertDialog(
          title: Text('Edit Match ${matchup.matchNumber}'),
          content: SingleChildScrollView(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text('Scores', style: Theme.of(context).textTheme.titleSmall),
                const SizedBox(height: 8),
                Row(
                  children: [
                    Expanded(
                      child: TextField(
                        controller: scoreAController,
                        decoration: InputDecoration(
                          labelText:
                              matchup.participantAName ?? 'Participant A',
                          border: const OutlineInputBorder(),
                        ),
                        keyboardType: TextInputType.number,
                      ),
                    ),
                    const SizedBox(width: 16),
                    Expanded(
                      child: TextField(
                        controller: scoreBController,
                        decoration: InputDecoration(
                          labelText:
                              matchup.participantBName ?? 'Participant B',
                          border: const OutlineInputBorder(),
                        ),
                        keyboardType: TextInputType.number,
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 24),
                Text(
                  'Select Winner',
                  style: Theme.of(context).textTheme.titleSmall,
                ),
                const SizedBox(height: 8),
                // Flutter 3.32+: a RadioGroup ancestor manages the group value
                // instead of per-Radio groupValue/onChanged.
                RadioGroup<String?>(
                  groupValue: selectedWinnerId,
                  onChanged: (val) =>
                      setDialogState(() => selectedWinnerId = val),
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      if (matchup.participantAId != null)
                        RadioListTile<String?>(
                          title: Text(
                            matchup.participantAName ?? 'Participant A',
                          ),
                          value: matchup.participantAId!,
                        ),
                      if (matchup.participantBId != null)
                        RadioListTile<String?>(
                          title: Text(
                            matchup.participantBName ?? 'Participant B',
                          ),
                          value: matchup.participantBId!,
                        ),
                      const RadioListTile<String?>(
                        title: Text('None (Pending)'),
                        value: null,
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(ctx),
              child: const Text('Cancel'),
            ),
            FilledButton(
              onPressed: () async {
                final scoreA = int.tryParse(scoreAController.text);
                final scoreB = int.tryParse(scoreBController.text);

                try {
                  await updateMatchupScore(
                    ref,
                    clubId: clubId,
                    tournamentId: tournament.id,
                    roundId: round.id,
                    matchupId: matchup.id,
                    scoreA: scoreA,
                    scoreB: scoreB,
                    winnerId: selectedWinnerId,
                  );
                  ref.invalidate(
                    tournamentMatchupsProvider((
                      clubId: clubId,
                      tournamentId: tournament.id,
                      roundId: round.id,
                    )),
                  );
                  if (ctx.mounted) Navigator.pop(ctx);
                } catch (e) {
                  if (ctx.mounted) {
                    ScaffoldMessenger.of(ctx).showSnackBar(
                      SnackBar(content: Text('Failed to update matchup: $e')),
                    );
                  }
                }
              },
              child: const Text('Save'),
            ),
          ],
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final matchupsAsync = ref.watch(
      tournamentMatchupsProvider((
        clubId: clubId,
        tournamentId: tournament.id,
        roundId: round.id,
      )),
    );
    final currentUser = ref.watch(currentUserProvider).value;
    final isOrganizer = currentUser?.id == tournament.organizerId;
    final theme = Theme.of(context);

    return matchupsAsync.when(
      data: (matchups) {
        if (matchups.isEmpty) {
          return Center(
            child: Text(
              'No matchups generated for this round yet.',
              style: theme.textTheme.bodyMedium?.copyWith(
                color: theme.colorScheme.onSurfaceVariant,
              ),
            ),
          );
        }

        return RefreshIndicator(
          onRefresh: () async => ref.invalidate(
            tournamentMatchupsProvider((
              clubId: clubId,
              tournamentId: tournament.id,
              roundId: round.id,
            )),
          ),
          child: ListView.builder(
            padding: const EdgeInsets.all(16),
            itemCount: matchups.length,
            itemBuilder: (context, index) {
              final matchup = matchups[index];

              final isPartAUser =
                  matchup.participantAId != null &&
                  currentUser?.id != null &&
                  tournament.participants.any(
                    (p) =>
                        p.userId == currentUser!.id &&
                        p.userId == matchup.participantAId,
                  );
              final isPartBUser =
                  matchup.participantBId != null &&
                  currentUser?.id != null &&
                  tournament.participants.any(
                    (p) =>
                        p.userId == currentUser!.id &&
                        p.userId == matchup.participantBId,
                  );
              final isUserInMatch = isPartAUser || isPartBUser;

              final isCompleted = matchup.winnerId != null;
              final isPartAWinner =
                  matchup.winnerId != null &&
                  matchup.winnerId == matchup.participantAId;
              final isPartBWinner =
                  matchup.winnerId != null &&
                  matchup.winnerId == matchup.participantBId;

              final canUserScore =
                  round.status == 'in_progress' &&
                  isUserInMatch &&
                  !isCompleted &&
                  ((isPartAUser && matchup.scoreA == null) ||
                      (isPartBUser && matchup.scoreB == null));

              return Card(
                margin: const EdgeInsets.only(bottom: 16),
                child: Padding(
                  padding: const EdgeInsets.all(16),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          Text(
                            'Match ${matchup.matchNumber}',
                            style: theme.textTheme.titleMedium?.copyWith(
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                          if (isCompleted)
                            Container(
                              padding: const EdgeInsets.symmetric(
                                horizontal: 8,
                                vertical: 2,
                              ),
                              decoration: BoxDecoration(
                                color: Colors.green.shade50,
                                borderRadius: BorderRadius.circular(12),
                              ),
                              child: Text(
                                'Completed',
                                style: theme.textTheme.labelSmall?.copyWith(
                                  color: Colors.green.shade800,
                                  fontWeight: FontWeight.bold,
                                ),
                              ),
                            )
                          else if (matchup.participantAId == null ||
                              matchup.participantBId == null)
                            Container(
                              padding: const EdgeInsets.symmetric(
                                horizontal: 8,
                                vertical: 2,
                              ),
                              decoration: BoxDecoration(
                                color: Colors.blue.shade50,
                                borderRadius: BorderRadius.circular(12),
                              ),
                              child: Text(
                                'BYE',
                                style: theme.textTheme.labelSmall?.copyWith(
                                  color: Colors.blue.shade800,
                                  fontWeight: FontWeight.bold,
                                ),
                              ),
                            )
                          else
                            Container(
                              padding: const EdgeInsets.symmetric(
                                horizontal: 8,
                                vertical: 2,
                              ),
                              decoration: BoxDecoration(
                                color: Colors.amber.shade50,
                                borderRadius: BorderRadius.circular(12),
                              ),
                              child: Text(
                                'Pending',
                                style: theme.textTheme.labelSmall?.copyWith(
                                  color: Colors.amber.shade900,
                                  fontWeight: FontWeight.bold,
                                ),
                              ),
                            ),
                        ],
                      ),
                      const Divider(height: 24),
                      _MatchupPlayerRow(
                        name: matchup.participantAName ?? 'BYE',
                        score: matchup.scoreA,
                        isWinner: isPartAWinner,
                        isUser: isPartAUser,
                        theme: theme,
                      ),
                      const SizedBox(height: 12),
                      _MatchupPlayerRow(
                        name: matchup.participantBName ?? 'BYE',
                        score: matchup.scoreB,
                        isWinner: isPartBWinner,
                        isUser: isPartBUser,
                        theme: theme,
                      ),
                      if (canUserScore || isOrganizer) ...[
                        const SizedBox(height: 16),
                        Row(
                          mainAxisAlignment: MainAxisAlignment.end,
                          children: [
                            if (isOrganizer)
                              TextButton.icon(
                                onPressed: () =>
                                    _showEditMatchDialog(context, ref, matchup),
                                icon: const Icon(Icons.edit, size: 18),
                                label: const Text('Edit Match'),
                              ),
                            if (canUserScore) ...[
                              const SizedBox(width: 8),
                              FilledButton.icon(
                                onPressed: () =>
                                    _scoreMatch(context, ref, matchup),
                                icon: const Icon(Icons.play_arrow, size: 18),
                                label: const Text('Score My Match'),
                              ),
                            ],
                          ],
                        ),
                      ],
                    ],
                  ),
                ),
              );
            },
          ),
        );
      },
      loading: () => const Center(child: CircularProgressIndicator()),
      error: (err, _) => Center(child: Text('Error loading matchups: $err')),
    );
  }
}

class _MatchupPlayerRow extends StatelessWidget {
  final String name;
  final int? score;
  final bool isWinner;
  final bool isUser;
  final ThemeData theme;

  const _MatchupPlayerRow({
    required this.name,
    required this.score,
    required this.isWinner,
    required this.isUser,
    required this.theme,
  });

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Row(
          children: [
            if (isWinner)
              const Padding(
                padding: EdgeInsets.only(right: 8),
                child: Icon(Icons.check_circle, size: 18, color: Colors.green),
              )
            else
              const SizedBox(width: 26),
            Text(
              name + (isUser ? ' (You)' : ''),
              style: theme.textTheme.bodyLarge?.copyWith(
                fontWeight: isWinner || isUser
                    ? FontWeight.bold
                    : FontWeight.normal,
                color: isWinner ? theme.colorScheme.primary : null,
              ),
            ),
          ],
        ),
        Text(
          score?.toString() ?? '-',
          style: theme.textTheme.bodyLarge?.copyWith(
            fontWeight: isWinner ? FontWeight.bold : FontWeight.normal,
            color: isWinner ? theme.colorScheme.primary : null,
          ),
        ),
      ],
    );
  }
}
