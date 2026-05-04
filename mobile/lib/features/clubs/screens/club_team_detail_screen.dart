import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../providers/club_detail_provider.dart';
import '../widgets/member_tile.dart';
import '../models/club.dart';

class ClubTeamDetailScreen extends ConsumerWidget {
  final String clubId;
  final String teamId;

  const ClubTeamDetailScreen({
    super.key,
    required this.clubId,
    required this.teamId,
  });

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final detailAsync = ref.watch(
        clubTeamDetailProvider((clubId: clubId, teamId: teamId)));
    final theme = Theme.of(context);

    return Scaffold(
      appBar: AppBar(title: const Text('Team')),
      body: detailAsync.when(
        data: (team) => RefreshIndicator(
          onRefresh: () async => ref.invalidate(
              clubTeamDetailProvider((clubId: clubId, teamId: teamId))),
          child: ListView(
            padding: const EdgeInsets.all(16),
            children: [
              Text(team.name, style: theme.textTheme.headlineSmall),
              if (team.description != null) ...[
                const SizedBox(height: 4),
                Text(team.description!,
                    style: theme.textTheme.bodyMedium?.copyWith(
                      color: theme.colorScheme.onSurfaceVariant,
                    )),
              ],
              const SizedBox(height: 16),
              Card(
                child: ListTile(
                  leading: CircleAvatar(
                    backgroundImage: team.leader.avatar != null
                        ? NetworkImage(team.leader.avatar!)
                        : null,
                    child: team.leader.avatar == null
                        ? Text(team.leader.effectiveName
                            .substring(0, 1)
                            .toUpperCase())
                        : null,
                  ),
                  title: Text(team.leader.effectiveName),
                  subtitle: const Text('Team Leader'),
                ),
              ),
              const SizedBox(height: 16),
              Text('Members (${team.members.length})',
                  style: theme.textTheme.titleMedium),
              const SizedBox(height: 8),
              ...team.members.map((m) => MemberTile(
                    member: ClubMember(
                      userId: m.userId,
                      username: m.username,
                      displayName: m.displayName,
                      avatar: m.avatar,
                      role: m.userId == team.leader.userId
                          ? 'leader'
                          : 'member',
                      joinedAt: m.joinedAt,
                    ),
                  )),
            ],
          ),
        ),
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (err, _) => Center(child: Text('Error: $err')),
      ),
    );
  }
}
