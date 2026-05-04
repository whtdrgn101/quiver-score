import 'package:flutter/material.dart';
import 'package:intl/intl.dart';

import '../models/club_event.dart';

class EventCard extends StatelessWidget {
  final ClubEvent event;
  final VoidCallback onTap;

  const EventCard({super.key, required this.event, required this.onTap});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final goingCount =
        event.participants.where((p) => p.status == 'going').length;

    return Card(
      margin: const EdgeInsets.only(bottom: 8),
      child: ListTile(
        onTap: onTap,
        leading: Icon(
          event.isPast ? Icons.event_available : Icons.event,
          color: event.isPast
              ? theme.colorScheme.onSurfaceVariant
              : theme.colorScheme.primary,
        ),
        title: Text(event.name),
        subtitle: Text(
          [
            DateFormat.yMMMd().format(event.eventDate),
            if (event.location != null) event.location!,
            '$goingCount going',
          ].join(' · '),
          maxLines: 1,
          overflow: TextOverflow.ellipsis,
        ),
        trailing: event.isPast
            ? Container(
                padding:
                    const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                decoration: BoxDecoration(
                  color: theme.colorScheme.surfaceContainerHighest,
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Text('Completed',
                    style: theme.textTheme.labelSmall?.copyWith(
                      color: theme.colorScheme.onSurfaceVariant,
                    )),
              )
            : const Icon(Icons.chevron_right, size: 20),
      ),
    );
  }
}
