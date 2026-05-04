import 'package:flutter/material.dart';
import 'package:intl/intl.dart';

import '../models/club.dart';

class ActivityItemCard extends StatelessWidget {
  final ActivityItem item;

  const ActivityItemCard({super.key, required this.item});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final isRecord = item.type == 'personal_record';

    return Card(
      margin: const EdgeInsets.only(bottom: 8),
      child: ListTile(
        leading: CircleAvatar(
          radius: 20,
          backgroundImage:
              item.avatar != null ? NetworkImage(item.avatar!) : null,
          child: item.avatar == null
              ? Text(item.effectiveName.substring(0, 1).toUpperCase())
              : null,
        ),
        title: Row(
          children: [
            Expanded(child: Text(item.effectiveName)),
            if (isRecord)
              Icon(Icons.emoji_events, size: 16, color: Colors.amber.shade700),
          ],
        ),
        subtitle: Text(
          '${isRecord ? 'New PR! ' : ''}${item.templateName} · ${item.score}${item.xCount > 0 ? ' (${item.xCount}x)' : ''}',
        ),
        trailing: Text(
          _formatDate(item.occurredAt),
          style: theme.textTheme.bodySmall?.copyWith(
            color: theme.colorScheme.onSurfaceVariant,
          ),
        ),
      ),
    );
  }

  String _formatDate(DateTime date) {
    final now = DateTime.now();
    final diff = now.difference(date);
    if (diff.inMinutes < 60) return '${diff.inMinutes}m ago';
    if (diff.inHours < 24) return '${diff.inHours}h ago';
    if (diff.inDays < 7) return '${diff.inDays}d ago';
    return DateFormat.MMMd().format(date);
  }
}
