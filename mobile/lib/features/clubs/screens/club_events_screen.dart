import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../providers/club_detail_provider.dart';
import '../widgets/event_card.dart';
import 'club_event_detail_screen.dart';

class ClubEventsTab extends ConsumerWidget {
  final String clubId;

  const ClubEventsTab({super.key, required this.clubId});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final eventsAsync = ref.watch(clubEventsProvider(clubId));
    final theme = Theme.of(context);

    return eventsAsync.when(
      data: (events) {
        if (events.isEmpty) {
          return Center(
            child: Text('No events yet',
                style: theme.textTheme.bodyMedium?.copyWith(
                  color: theme.colorScheme.onSurfaceVariant,
                )),
          );
        }

        final sorted = [...events]
          ..sort((a, b) => b.eventDate.compareTo(a.eventDate));

        return RefreshIndicator(
          onRefresh: () async =>
              ref.invalidate(clubEventsProvider(clubId)),
          child: ListView.builder(
            padding: const EdgeInsets.all(16),
            itemCount: sorted.length,
            itemBuilder: (context, index) => EventCard(
              event: sorted[index],
              onTap: () => Navigator.push(
                context,
                MaterialPageRoute(
                  builder: (_) => ClubEventDetailScreen(
                    clubId: clubId,
                    eventId: sorted[index].id,
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
