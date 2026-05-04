import 'package:flutter/material.dart';

import '../models/club.dart';

class ClubCard extends StatelessWidget {
  final Club club;
  final VoidCallback onTap;

  const ClubCard({super.key, required this.club, required this.onTap});

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: const EdgeInsets.only(bottom: 8),
      child: ListTile(
        onTap: onTap,
        leading: CircleAvatar(
          backgroundImage:
              club.avatar != null ? NetworkImage(club.avatar!) : null,
          child: club.avatar == null
              ? Text(club.name.substring(0, 1).toUpperCase())
              : null,
        ),
        title: Text(club.name),
        subtitle: Text(
          [
            '${club.memberCount} member${club.memberCount == 1 ? '' : 's'}',
            if (club.myRole != null) club.myRole!,
          ].join(' · '),
        ),
        trailing: const Icon(Icons.chevron_right, size: 20),
      ),
    );
  }
}
