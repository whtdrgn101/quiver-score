import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';

import '../../more/providers/user_provider.dart';
import '../providers/club_detail_provider.dart';

class ClubEventDetailScreen extends ConsumerWidget {
  final String clubId;
  final String eventId;

  const ClubEventDetailScreen({
    super.key,
    required this.clubId,
    required this.eventId,
  });

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final eventAsync = ref.watch(
        clubEventDetailProvider((clubId: clubId, eventId: eventId)));
    final user = ref.watch(currentUserProvider).value;
    final theme = Theme.of(context);

    return Scaffold(
      appBar: AppBar(title: const Text('Event')),
      body: eventAsync.when(
        data: (event) {
          final myParticipant = user != null
              ? event.participants
                  .where((p) => p.userId == user.id)
                  .firstOrNull
              : null;
          final going =
              event.participants.where((p) => p.status == 'going').toList();
          final maybe =
              event.participants.where((p) => p.status == 'maybe').toList();
          final declined =
              event.participants.where((p) => p.status == 'declined').toList();

          return RefreshIndicator(
            onRefresh: () async => ref.invalidate(
                clubEventDetailProvider((clubId: clubId, eventId: eventId))),
            child: ListView(
              padding: const EdgeInsets.all(16),
              children: [
                Text(event.name, style: theme.textTheme.headlineSmall),
                const SizedBox(height: 8),
                if (event.templateName != null)
                  _InfoRow(
                      icon: Icons.sports_score,
                      text: event.templateName!),
                _InfoRow(
                    icon: Icons.calendar_today,
                    text: DateFormat.yMMMd()
                        .add_jm()
                        .format(event.eventDate)),
                if (event.location != null)
                  _InfoRow(
                      icon: Icons.location_on_outlined,
                      text: event.location!),
                if (event.description != null) ...[
                  const SizedBox(height: 12),
                  Text(event.description!,
                      style: theme.textTheme.bodyMedium),
                ],

                // RSVP buttons
                if (!event.isPast) ...[
                  const SizedBox(height: 24),
                  Text('Your RSVP', style: theme.textTheme.titleMedium),
                  const SizedBox(height: 8),
                  SegmentedButton<String>(
                    segments: const [
                      ButtonSegment(
                          value: 'going',
                          icon: Icon(Icons.check_circle_outline),
                          label: Text('Going')),
                      ButtonSegment(
                          value: 'maybe',
                          icon: Icon(Icons.help_outline),
                          label: Text('Maybe')),
                      ButtonSegment(
                          value: 'declined',
                          icon: Icon(Icons.cancel_outlined),
                          label: Text('Decline')),
                    ],
                    selected: {myParticipant?.status ?? ''},
                    onSelectionChanged: (selected) {
                      rsvpEvent(ref, clubId, eventId, selected.first);
                    },
                    emptySelectionAllowed: true,
                  ),
                ],

                // Participant lists
                const SizedBox(height: 24),
                if (going.isNotEmpty) ...[
                  _ParticipantSection(
                      title: 'Going (${going.length})',
                      participants: going,
                      isPast: event.isPast),
                  const SizedBox(height: 16),
                ],
                if (maybe.isNotEmpty) ...[
                  _ParticipantSection(
                      title: 'Maybe (${maybe.length})',
                      participants: maybe,
                      isPast: event.isPast),
                  const SizedBox(height: 16),
                ],
                if (declined.isNotEmpty)
                  _ParticipantSection(
                      title: 'Declined (${declined.length})',
                      participants: declined,
                      isPast: event.isPast),
              ],
            ),
          );
        },
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (err, _) => Center(child: Text('Error: $err')),
      ),
    );
  }
}

class _InfoRow extends StatelessWidget {
  final IconData icon;
  final String text;

  const _InfoRow({required this.icon, required this.text});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 2),
      child: Row(
        children: [
          Icon(icon, size: 16, color: theme.colorScheme.onSurfaceVariant),
          const SizedBox(width: 8),
          Expanded(
              child: Text(text,
                  style: theme.textTheme.bodyMedium?.copyWith(
                    color: theme.colorScheme.onSurfaceVariant,
                  ))),
        ],
      ),
    );
  }
}

class _ParticipantSection extends StatelessWidget {
  final String title;
  final List<dynamic> participants;
  final bool isPast;

  const _ParticipantSection({
    required this.title,
    required this.participants,
    required this.isPast,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(title, style: theme.textTheme.titleSmall),
        const SizedBox(height: 4),
        ...participants.map((p) => Card(
              margin: const EdgeInsets.only(bottom: 4),
              child: ListTile(
                dense: true,
                leading: CircleAvatar(
                  radius: 16,
                  backgroundImage:
                      p.avatar != null ? NetworkImage(p.avatar!) : null,
                  child: p.avatar == null
                      ? Text(p.effectiveName.substring(0, 1).toUpperCase())
                      : null,
                ),
                title: Text(p.effectiveName),
                trailing: isPast && p.score != null
                    ? Text('${p.score}${p.xCount != null && p.xCount > 0 ? ' (${p.xCount}x)' : ''}',
                        style: theme.textTheme.bodyMedium?.copyWith(
                          fontWeight: FontWeight.w600,
                        ))
                    : null,
              ),
            )),
      ],
    );
  }
}
