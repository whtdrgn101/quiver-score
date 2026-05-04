import 'package:flutter/material.dart';

import '../models/club.dart';

class MemberTile extends StatelessWidget {
  final ClubMember member;

  const MemberTile({super.key, required this.member});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Card(
      margin: const EdgeInsets.only(bottom: 4),
      child: ListTile(
        dense: true,
        leading: CircleAvatar(
          radius: 16,
          backgroundImage:
              member.avatar != null ? NetworkImage(member.avatar!) : null,
          child: member.avatar == null
              ? Text(member.effectiveName.substring(0, 1).toUpperCase())
              : null,
        ),
        title: Text(member.effectiveName),
        subtitle: Text('@${member.username}'),
        trailing: _roleBadge(theme),
      ),
    );
  }

  Widget? _roleBadge(ThemeData theme) {
    if (member.role == 'member') return null;

    final color = member.role == 'owner'
        ? theme.colorScheme.primary
        : theme.colorScheme.tertiary;

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.1),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: color.withValues(alpha: 0.3)),
      ),
      child: Text(
        member.role,
        style: theme.textTheme.labelSmall?.copyWith(color: color),
      ),
    );
  }
}
