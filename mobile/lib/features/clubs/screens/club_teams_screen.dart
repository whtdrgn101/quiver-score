import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../providers/club_detail_provider.dart';
import 'club_team_detail_screen.dart';

class ClubTeamsTab extends ConsumerWidget {
  final String clubId;

  const ClubTeamsTab({super.key, required this.clubId});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final teamsAsync = ref.watch(clubTeamsProvider(clubId));
    final theme = Theme.of(context);

    return teamsAsync.when(
      data: (teams) {
        if (teams.isEmpty) {
          return Center(
            child: Text('No teams yet',
                style: theme.textTheme.bodyMedium?.copyWith(
                  color: theme.colorScheme.onSurfaceVariant,
                )),
          );
        }

        return RefreshIndicator(
          onRefresh: () async =>
              ref.invalidate(clubTeamsProvider(clubId)),
          child: ListView.builder(
            padding: const EdgeInsets.all(16),
            itemCount: teams.length,
            itemBuilder: (context, index) {
              final team = teams[index];
              return Card(
                margin: const EdgeInsets.only(bottom: 8),
                child: ListTile(
                  onTap: () => Navigator.push(
                    context,
                    MaterialPageRoute(
                      builder: (_) => ClubTeamDetailScreen(
                        clubId: clubId,
                        teamId: team.id,
                      ),
                    ),
                  ),
                  leading: const Icon(Icons.group_outlined),
                  title: Text(team.name),
                  subtitle: Text(
                    [
                      'Led by ${team.leader.effectiveName}',
                      '${team.memberCount} member${team.memberCount == 1 ? '' : 's'}',
                    ].join(' · '),
                  ),
                  trailing: const Icon(Icons.chevron_right, size: 20),
                ),
              );
            },
          ),
        );
      },
      loading: () => const Center(child: CircularProgressIndicator()),
      error: (err, _) => Center(child: Text('Error: $err')),
    );
  }
}
