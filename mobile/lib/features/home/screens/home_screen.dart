import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/database/database.dart';
import '../../../core/sync/sync_service.dart';
import '../../scoring/screens/dashboard_screen.dart';
import '../../scoring/screens/history_screen.dart';
import '../../more/screens/more_screen.dart';

final pendingSyncCountProvider = StreamProvider<int>((ref) {
  final db = ref.watch(databaseProvider);
  return (db.select(db.syncQueue)..where((t) => t.syncedAt.isNull()))
      .watch()
      .map((items) => items.length);
});

class HomeScreen extends ConsumerStatefulWidget {
  const HomeScreen({super.key});

  @override
  ConsumerState<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends ConsumerState<HomeScreen> {
  int _currentIndex = 0;
  bool _syncing = false;

  static const _screens = [
    DashboardScreen(),
    HistoryScreen(),
    MoreScreen(),
  ];

  static const _titles = [
    'QuiverScore',
    'History',
    'More',
  ];

  Future<void> _syncNow() async {
    setState(() => _syncing = true);
    try {
      final result =
          await ref.read(syncServiceProvider).syncPendingItems();
      if (mounted) {
        final msg = result.failed > 0
            ? 'Synced ${result.synced}, ${result.failed} failed: ${result.lastError}'
            : result.synced > 0
                ? 'Synced ${result.synced} item${result.synced == 1 ? '' : 's'}'
                : 'Everything is up to date';
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(msg), duration: const Duration(seconds: 2)),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Sync failed: $e')),
        );
      }
    } finally {
      if (mounted) setState(() => _syncing = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final pendingCount = ref.watch(pendingSyncCountProvider).valueOrNull ?? 0;

    return Scaffold(
      appBar: AppBar(
        title: Text(_titles[_currentIndex]),
        actions: [
          IconButton(
            onPressed: _syncing ? null : _syncNow,
            icon: _syncing
                ? const SizedBox(
                    width: 20,
                    height: 20,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : Badge(
                    isLabelVisible: pendingCount > 0,
                    label: Text('$pendingCount'),
                    child: const Icon(Icons.sync),
                  ),
            tooltip: pendingCount > 0
                ? '$pendingCount pending'
                : 'Sync now',
          ),
        ],
      ),
      body: IndexedStack(
        index: _currentIndex,
        children: _screens,
      ),
      bottomNavigationBar: NavigationBar(
        selectedIndex: _currentIndex,
        onDestinationSelected: (index) => setState(() => _currentIndex = index),
        destinations: const [
          NavigationDestination(
            icon: Icon(Icons.home_outlined),
            selectedIcon: Icon(Icons.home),
            label: 'Dashboard',
          ),
          NavigationDestination(
            icon: Icon(Icons.history_outlined),
            selectedIcon: Icon(Icons.history),
            label: 'History',
          ),
          NavigationDestination(
            icon: Icon(Icons.more_horiz),
            selectedIcon: Icon(Icons.more_horiz),
            label: 'More',
          ),
        ],
      ),
    );
  }
}
