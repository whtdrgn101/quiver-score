import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../more/providers/user_provider.dart';
import '../providers/club_detail_provider.dart';
import '../providers/clubs_provider.dart';
import '../widgets/member_tile.dart';
import 'club_leaderboard_screen.dart';
import 'club_activity_screen.dart';
import 'club_events_screen.dart';
import 'club_teams_screen.dart';

class ClubDetailScreen extends ConsumerStatefulWidget {
  final String clubId;

  const ClubDetailScreen({super.key, required this.clubId});

  @override
  ConsumerState<ClubDetailScreen> createState() => _ClubDetailScreenState();
}

class _ClubDetailScreenState extends ConsumerState<ClubDetailScreen>
    with SingleTickerProviderStateMixin {
  late final TabController _tabController;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 5, vsync: this);
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final detailAsync = ref.watch(clubDetailProvider(widget.clubId));

    return Scaffold(
      appBar: AppBar(
        title: detailAsync.whenOrNull(data: (d) => Text(d.name)) ??
            const Text('Club'),
        actions: [
          detailAsync.whenOrNull(
                data: (detail) {
                  if (detail.myRole == 'owner') return const SizedBox.shrink();
                  return PopupMenuButton<String>(
                    onSelected: (value) {
                      if (value == 'leave') _confirmLeave(context, detail.name);
                    },
                    itemBuilder: (context) => [
                      PopupMenuItem(
                        value: 'leave',
                        child: Text('Leave Club',
                            style: TextStyle(
                                color: Theme.of(context).colorScheme.error)),
                      ),
                    ],
                  );
                },
              ) ??
              const SizedBox.shrink(),
        ],
        bottom: TabBar(
          controller: _tabController,
          isScrollable: true,
          tabs: const [
            Tab(text: 'Overview'),
            Tab(text: 'Leaderboard'),
            Tab(text: 'Activity'),
            Tab(text: 'Events'),
            Tab(text: 'Teams'),
          ],
        ),
      ),
      body: detailAsync.when(
        data: (detail) => TabBarView(
          controller: _tabController,
          children: [
            _OverviewTab(detail: detail),
            ClubLeaderboardTab(clubId: widget.clubId),
            ClubActivityTab(clubId: widget.clubId),
            ClubEventsTab(clubId: widget.clubId),
            ClubTeamsTab(clubId: widget.clubId),
          ],
        ),
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (err, _) => Center(child: Text('Error: $err')),
      ),
    );
  }

  Future<void> _confirmLeave(BuildContext context, String clubName) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Leave Club'),
        content: Text('Leave "$clubName"? You will need a new invite to rejoin.'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            style: FilledButton.styleFrom(
              backgroundColor: Theme.of(context).colorScheme.error,
            ),
            child: const Text('Leave'),
          ),
        ],
      ),
    );

    if (confirmed == true && mounted) {
      final user = ref.read(currentUserProvider).valueOrNull;
      if (user == null) return;
      await ref
          .read(clubsProvider.notifier)
          .leaveClub(widget.clubId, user.id);
      if (mounted) Navigator.pop(context);
    }
  }
}

class _OverviewTab extends StatelessWidget {
  final dynamic detail;

  const _OverviewTab({required this.detail});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return ListView(
      padding: const EdgeInsets.all(16),
      children: [
        if (detail.description != null) ...[
          Text(detail.description!,
              style: theme.textTheme.bodyLarge),
          const SizedBox(height: 16),
        ],
        Row(
          children: [
            Icon(Icons.people, size: 18, color: theme.colorScheme.primary),
            const SizedBox(width: 8),
            Text('${detail.memberCount} member${detail.memberCount == 1 ? '' : 's'}',
                style: theme.textTheme.titleSmall),
          ],
        ),
        if (detail.myRole != null) ...[
          const SizedBox(height: 4),
          Text('Your role: ${detail.myRole}',
              style: theme.textTheme.bodySmall?.copyWith(
                color: theme.colorScheme.onSurfaceVariant,
              )),
        ],
        const SizedBox(height: 16),
        Text('Members', style: theme.textTheme.titleMedium),
        const SizedBox(height: 8),
        ...detail.members.map<Widget>((m) => MemberTile(member: m)),
      ],
    );
  }
}
