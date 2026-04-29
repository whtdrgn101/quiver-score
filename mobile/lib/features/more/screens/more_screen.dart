import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:url_launcher/url_launcher.dart';

import '../../auth/providers/auth_provider.dart';
import '../providers/user_provider.dart';
import 'profile_edit_screen.dart';

class MoreScreen extends ConsumerWidget {
  const MoreScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final theme = Theme.of(context);
    final userAsync = ref.watch(currentUserProvider);

    return ListView(
      padding: const EdgeInsets.all(16),
      children: [
        // User info card
        Card(
          child: Padding(
            padding: const EdgeInsets.all(16),
            child: userAsync.when(
              data: (user) => Row(
                children: [
                  CircleAvatar(
                    radius: 28,
                    backgroundColor: theme.colorScheme.primaryContainer,
                    backgroundImage: user.avatar != null
                        ? NetworkImage(user.avatar!)
                        : null,
                    child: user.avatar == null
                        ? Text(
                            (user.displayName ?? user.username)
                                .substring(0, 1)
                                .toUpperCase(),
                            style: theme.textTheme.headlineSmall?.copyWith(
                              color: theme.colorScheme.onPrimaryContainer,
                            ),
                          )
                        : null,
                  ),
                  const SizedBox(width: 16),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          user.displayName ?? user.username,
                          style: theme.textTheme.titleMedium?.copyWith(
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                        Text(
                          user.email,
                          style: theme.textTheme.bodySmall?.copyWith(
                            color: theme.colorScheme.onSurfaceVariant,
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
              loading: () => const Row(
                children: [
                  CircleAvatar(radius: 28),
                  SizedBox(width: 16),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        SizedBox(
                          height: 16,
                          width: 120,
                          child: DecoratedBox(
                            decoration: BoxDecoration(color: Colors.black12),
                          ),
                        ),
                        SizedBox(height: 4),
                        SizedBox(
                          height: 12,
                          width: 160,
                          child: DecoratedBox(
                            decoration: BoxDecoration(color: Colors.black12),
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
              error: (_, _) => const Text('Could not load profile'),
            ),
          ),
        ),

        const SizedBox(height: 8),

        // Edit profile
        userAsync.whenOrNull(
              data: (user) => _MenuTile(
                icon: Icons.edit_outlined,
                title: 'Edit Profile',
                subtitle: 'Name, bio, bow type, social links',
                onTap: () => Navigator.push(
                  context,
                  MaterialPageRoute(
                    builder: (_) => ProfileEditScreen(user: user),
                  ),
                ),
              ),
            ) ??
            const SizedBox.shrink(),

        const SizedBox(height: 24),

        // Web links section
        Text('QuiverScore Web',
            style: theme.textTheme.titleSmall?.copyWith(
              color: theme.colorScheme.onSurfaceVariant,
            )),
        const SizedBox(height: 8),

        _MenuTile(
          icon: Icons.dashboard_outlined,
          title: 'Dashboard',
          subtitle: 'Stats, trends, and personal records',
          onTap: () => _openUrl('https://quiverscore.com'),
        ),
        _MenuTile(
          icon: Icons.people_outlined,
          title: 'Social',
          subtitle: 'Friends, followers, and activity feed',
          onTap: () => _openUrl('https://quiverscore.com/social'),
        ),
        _MenuTile(
          icon: Icons.emoji_events_outlined,
          title: 'Clubs & Tournaments',
          subtitle: 'Manage clubs and tournament play',
          onTap: () => _openUrl('https://quiverscore.com/clubs'),
        ),
        _MenuTile(
          icon: Icons.settings_outlined,
          title: 'Settings & Equipment',
          subtitle: 'Equipment, setups, and sight marks',
          onTap: () => _openUrl('https://quiverscore.com/equipment'),
        ),

        const SizedBox(height: 24),

        // Account section
        Text('Account',
            style: theme.textTheme.titleSmall?.copyWith(
              color: theme.colorScheme.onSurfaceVariant,
            )),
        const SizedBox(height: 8),

        _MenuTile(
          icon: Icons.logout,
          title: 'Sign Out',
          onTap: () => ref.read(authProvider.notifier).logout(),
          color: theme.colorScheme.error,
        ),
      ],
    );
  }

  Future<void> _openUrl(String url) async {
    await launchUrl(Uri.parse(url), mode: LaunchMode.externalApplication);
  }
}

class _MenuTile extends StatelessWidget {
  final IconData icon;
  final String title;
  final String? subtitle;
  final VoidCallback onTap;
  final Color? color;

  const _MenuTile({
    required this.icon,
    required this.title,
    this.subtitle,
    required this.onTap,
    this.color,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Card(
      margin: const EdgeInsets.only(bottom: 4),
      child: ListTile(
        leading: Icon(icon, color: color),
        title: Text(title, style: TextStyle(color: color)),
        subtitle: subtitle != null
            ? Text(subtitle!,
                style: theme.textTheme.bodySmall?.copyWith(
                  color: theme.colorScheme.onSurfaceVariant,
                ))
            : null,
        trailing: const Icon(Icons.chevron_right, size: 20),
        onTap: onTap,
      ),
    );
  }
}
