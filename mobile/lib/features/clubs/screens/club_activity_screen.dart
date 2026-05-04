import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../providers/club_detail_provider.dart';
import '../widgets/activity_item_card.dart';

class ClubActivityTab extends ConsumerWidget {
  final String clubId;

  const ClubActivityTab({super.key, required this.clubId});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final activityAsync = ref.watch(clubActivityProvider(clubId));
    final theme = Theme.of(context);

    return activityAsync.when(
      data: (items) {
        if (items.isEmpty) {
          return Center(
            child: Text('No recent activity',
                style: theme.textTheme.bodyMedium?.copyWith(
                  color: theme.colorScheme.onSurfaceVariant,
                )),
          );
        }

        return RefreshIndicator(
          onRefresh: () async =>
              ref.invalidate(clubActivityProvider(clubId)),
          child: ListView.builder(
            padding: const EdgeInsets.all(16),
            itemCount: items.length,
            itemBuilder: (context, index) =>
                ActivityItemCard(item: items[index]),
          ),
        );
      },
      loading: () => const Center(child: CircularProgressIndicator()),
      error: (err, _) => Center(child: Text('Error: $err')),
    );
  }
}
