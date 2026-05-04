import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../providers/tournament_provider.dart';
import '../widgets/tournament_card.dart';
import 'tournament_detail_screen.dart';

class ClubTournamentsTab extends ConsumerWidget {
  final String clubId;

  const ClubTournamentsTab({super.key, required this.clubId});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final tournamentsAsync = ref.watch(clubTournamentsProvider(clubId));
    final theme = Theme.of(context);

    return tournamentsAsync.when(
      data: (tournaments) {
        if (tournaments.isEmpty) {
          return Center(
            child: Text(
              'No tournaments yet',
              style: theme.textTheme.bodyMedium?.copyWith(
                color: theme.colorScheme.onSurfaceVariant,
              ),
            ),
          );
        }

        final sorted = [...tournaments]
          ..sort((a, b) => b.createdAt.compareTo(a.createdAt));

        return RefreshIndicator(
          onRefresh: () async =>
              ref.invalidate(clubTournamentsProvider(clubId)),
          child: ListView.builder(
            padding: const EdgeInsets.all(16),
            itemCount: sorted.length,
            itemBuilder: (context, index) => TournamentCard(
              tournament: sorted[index],
              onTap: () => Navigator.push(
                context,
                MaterialPageRoute(
                  builder: (_) => TournamentDetailScreen(
                    clubId: clubId,
                    tournamentId: sorted[index].id,
                  ),
                ),
              ),
            ),
          ),
        );
      },
      loading: () => const Center(child: CircularProgressIndicator()),
      error: (err, _) => Center(child: Text('Error: $err')),
    );
  }
}
