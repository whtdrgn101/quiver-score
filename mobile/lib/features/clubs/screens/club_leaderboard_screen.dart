import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../providers/club_detail_provider.dart';

class ClubLeaderboardTab extends ConsumerWidget {
  final String clubId;

  const ClubLeaderboardTab({super.key, required this.clubId});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final lbAsync = ref.watch(clubLeaderboardProvider(clubId));
    final theme = Theme.of(context);

    return lbAsync.when(
      data: (groups) {
        if (groups.isEmpty) {
          return Center(
            child: Text('No scores recorded yet',
                style: theme.textTheme.bodyMedium?.copyWith(
                  color: theme.colorScheme.onSurfaceVariant,
                )),
          );
        }

        return RefreshIndicator(
          onRefresh: () async =>
              ref.invalidate(clubLeaderboardProvider(clubId)),
          child: ListView.builder(
            padding: const EdgeInsets.all(16),
            itemCount: groups.length,
            itemBuilder: (context, index) {
              final group = groups[index];
              return Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  if (index > 0) const SizedBox(height: 16),
                  Text(group.templateName,
                      style: theme.textTheme.titleSmall?.copyWith(
                        color: theme.colorScheme.primary,
                        fontWeight: FontWeight.w600,
                      )),
                  const SizedBox(height: 8),
                  ...List.generate(group.entries.length, (i) {
                    final entry = group.entries[i];
                    return Card(
                      margin: const EdgeInsets.only(bottom: 4),
                      child: ListTile(
                        leading: CircleAvatar(
                          radius: 16,
                          backgroundColor:
                              i < 3 ? _rankColor(i) : theme.colorScheme.surfaceContainerHighest,
                          child: Text('${i + 1}',
                              style: TextStyle(
                                fontWeight: FontWeight.bold,
                                color: i < 3
                                    ? Colors.white
                                    : theme.colorScheme.onSurface,
                              )),
                        ),
                        title: Text(entry.effectiveName),
                        subtitle: Text('@${entry.username}'),
                        trailing: Column(
                          mainAxisAlignment: MainAxisAlignment.center,
                          crossAxisAlignment: CrossAxisAlignment.end,
                          children: [
                            Text('${entry.bestScore}',
                                style: theme.textTheme.titleMedium?.copyWith(
                                  fontWeight: FontWeight.bold,
                                )),
                            if (entry.bestXCount > 0)
                              Text('${entry.bestXCount}x',
                                  style: theme.textTheme.bodySmall),
                          ],
                        ),
                      ),
                    );
                  }),
                ],
              );
            },
          ),
        );
      },
      loading: () => const Center(child: CircularProgressIndicator()),
      error: (err, _) => Center(child: Text('Error: $err')),
    );
  }

  Color _rankColor(int index) {
    switch (index) {
      case 0:
        return const Color(0xFFFFD700);
      case 1:
        return const Color(0xFFC0C0C0);
      case 2:
        return const Color(0xFFCD7F32);
      default:
        return Colors.grey;
    }
  }
}
