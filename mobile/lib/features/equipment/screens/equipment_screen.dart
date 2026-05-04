import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../models/equipment.dart';
import '../models/setup.dart';
import '../providers/equipment_provider.dart';
import '../providers/setup_provider.dart';
import '../widgets/equipment_card.dart';
import '../widgets/category_header.dart';
import '../widgets/setup_card.dart';
import 'equipment_form_screen.dart';
import 'setup_form_screen.dart';
import 'setup_detail_screen.dart';

class EquipmentScreen extends ConsumerStatefulWidget {
  const EquipmentScreen({super.key});

  @override
  ConsumerState<EquipmentScreen> createState() => _EquipmentScreenState();
}

class _EquipmentScreenState extends ConsumerState<EquipmentScreen>
    with SingleTickerProviderStateMixin {
  late final TabController _tabController;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 2, vsync: this);
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Equipment & Setups'),
        bottom: TabBar(
          controller: _tabController,
          tabs: const [
            Tab(text: 'My Equipment'),
            Tab(text: 'My Setups'),
          ],
        ),
      ),
      body: TabBarView(
        controller: _tabController,
        children: const [
          _EquipmentTab(),
          _SetupsTab(),
        ],
      ),
      floatingActionButton: ListenableBuilder(
        listenable: _tabController,
        builder: (context, _) => FloatingActionButton(
          onPressed: () => _tabController.index == 0
              ? _addEquipment(context)
              : _addSetup(context),
          child: const Icon(Icons.add),
        ),
      ),
    );
  }

  void _addEquipment(BuildContext context) {
    Navigator.push(
      context,
      MaterialPageRoute(builder: (_) => const EquipmentFormScreen()),
    );
  }

  void _addSetup(BuildContext context) {
    Navigator.push(
      context,
      MaterialPageRoute(builder: (_) => const SetupFormScreen()),
    );
  }
}

class _EquipmentTab extends ConsumerWidget {
  const _EquipmentTab();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final equipmentAsync = ref.watch(equipmentProvider);

    return equipmentAsync.when(
      data: (items) {
        if (items.isEmpty) {
          return Center(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Icon(Icons.inventory_2_outlined,
                    size: 64,
                    color: Theme.of(context).colorScheme.onSurfaceVariant),
                const SizedBox(height: 16),
                Text('No equipment yet',
                    style: Theme.of(context).textTheme.titleMedium),
                const SizedBox(height: 8),
                Text('Tap + to add your first piece of gear',
                    style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                          color:
                              Theme.of(context).colorScheme.onSurfaceVariant,
                        )),
              ],
            ),
          );
        }

        final grouped = _groupByCategory(items);
        return RefreshIndicator(
          onRefresh: () => ref.read(equipmentProvider.notifier).refresh(),
          child: ListView.builder(
            padding: const EdgeInsets.all(16),
            itemCount: grouped.length,
            itemBuilder: (context, index) {
              final entry = grouped[index];
              return Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  CategoryHeader(
                    category: entry.key,
                    count: entry.value.length,
                  ),
                  ...entry.value.map((item) => EquipmentCard(
                        equipment: item,
                        onEdit: () => Navigator.push(
                          context,
                          MaterialPageRoute(
                            builder: (_) =>
                                EquipmentFormScreen(equipment: item),
                          ),
                        ),
                        onDelete: () =>
                            _confirmDelete(context, ref, item),
                      )),
                  const SizedBox(height: 8),
                ],
              );
            },
          ),
        );
      },
      loading: () => const Center(child: CircularProgressIndicator()),
      error: (err, _) => Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text('Failed to load equipment: $err'),
            const SizedBox(height: 8),
            FilledButton(
              onPressed: () =>
                  ref.read(equipmentProvider.notifier).refresh(),
              child: const Text('Retry'),
            ),
          ],
        ),
      ),
    );
  }

  List<MapEntry<String, List<Equipment>>> _groupByCategory(
      List<Equipment> items) {
    final map = <String, List<Equipment>>{};
    for (final item in items) {
      map.putIfAbsent(item.category, () => []).add(item);
    }
    final entries = map.entries.toList();
    entries.sort((a, b) => Equipment.categories
        .indexOf(a.key)
        .compareTo(Equipment.categories.indexOf(b.key)));
    return entries;
  }

  Future<void> _confirmDelete(
      BuildContext context, WidgetRef ref, Equipment item) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Delete Equipment'),
        content: Text('Delete "${item.name}"? This cannot be undone.'),
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
            child: const Text('Delete'),
          ),
        ],
      ),
    );
    if (confirmed == true) {
      await ref.read(equipmentProvider.notifier).delete(item.id);
    }
  }
}

class _SetupsTab extends ConsumerWidget {
  const _SetupsTab();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final setupsAsync = ref.watch(setupListProvider);

    return setupsAsync.when(
      data: (items) {
        if (items.isEmpty) {
          return Center(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Icon(Icons.tune_outlined,
                    size: 64,
                    color: Theme.of(context).colorScheme.onSurfaceVariant),
                const SizedBox(height: 16),
                Text('No setups yet',
                    style: Theme.of(context).textTheme.titleMedium),
                const SizedBox(height: 8),
                Text('Tap + to create your first bow setup',
                    style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                          color:
                              Theme.of(context).colorScheme.onSurfaceVariant,
                        )),
              ],
            ),
          );
        }

        return RefreshIndicator(
          onRefresh: () => ref.read(setupListProvider.notifier).refresh(),
          child: ListView.builder(
            padding: const EdgeInsets.all(16),
            itemCount: items.length,
            itemBuilder: (context, index) => SetupCard(
              setup: items[index],
              onTap: () => Navigator.push(
                context,
                MaterialPageRoute(
                  builder: (_) =>
                      SetupDetailScreen(setupId: items[index].id),
                ),
              ),
              onDelete: () =>
                  _confirmDelete(context, ref, items[index]),
            ),
          ),
        );
      },
      loading: () => const Center(child: CircularProgressIndicator()),
      error: (err, _) => Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text('Failed to load setups: $err'),
            const SizedBox(height: 8),
            FilledButton(
              onPressed: () =>
                  ref.read(setupListProvider.notifier).refresh(),
              child: const Text('Retry'),
            ),
          ],
        ),
      ),
    );
  }

  Future<void> _confirmDelete(
      BuildContext context, WidgetRef ref, SetupSummary setup) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Delete Setup'),
        content: Text('Delete "${setup.name}"? This cannot be undone.'),
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
            child: const Text('Delete'),
          ),
        ],
      ),
    );
    if (confirmed == true) {
      await ref.read(setupListProvider.notifier).delete(setup.id);
    }
  }
}
