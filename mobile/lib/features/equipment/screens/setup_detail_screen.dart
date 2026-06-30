import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../models/equipment.dart';
import '../models/setup.dart';
import '../providers/equipment_provider.dart';
import '../providers/setup_provider.dart';
import '../widgets/equipment_picker.dart';
import 'setup_form_screen.dart';

class SetupDetailScreen extends ConsumerWidget {
  final String setupId;

  const SetupDetailScreen({super.key, required this.setupId});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final detailAsync = ref.watch(setupDetailProvider(setupId));
    final theme = Theme.of(context);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Setup Detail'),
        actions: [
          detailAsync.whenOrNull(
                data: (detail) => IconButton(
                  icon: const Icon(Icons.edit_outlined),
                  onPressed: () async {
                    await Navigator.push(
                      context,
                      MaterialPageRoute(
                        builder: (_) => SetupFormScreen(detail: detail),
                      ),
                    );
                    ref.invalidate(setupDetailProvider(setupId));
                  },
                ),
              ) ??
              const SizedBox.shrink(),
        ],
      ),
      body: detailAsync.when(
        data: (detail) => RefreshIndicator(
          onRefresh: () async =>
              ref.invalidate(setupDetailProvider(setupId)),
          child: ListView(
            padding: const EdgeInsets.all(16),
            children: [
              // Name and description
              Text(detail.name, style: theme.textTheme.headlineSmall),
              if (detail.description != null) ...[
                const SizedBox(height: 4),
                Text(detail.description!,
                    style: theme.textTheme.bodyMedium?.copyWith(
                      color: theme.colorScheme.onSurfaceVariant,
                    )),
              ],
              const SizedBox(height: 24),

              // Tuning fields
              if (_hasTuning(detail)) ...[
                Text('Tuning', style: theme.textTheme.titleMedium),
                const SizedBox(height: 8),
                Card(
                  child: Padding(
                    padding: const EdgeInsets.all(16),
                    child: Wrap(
                      spacing: 24,
                      runSpacing: 12,
                      children: [
                        if (detail.braceHeight != null)
                          _TuningField(
                              label: 'Brace Height',
                              value: '${detail.braceHeight}" '),
                        if (detail.tiller != null)
                          _TuningField(
                              label: 'Tiller',
                              value: '${detail.tiller}"'),
                        if (detail.drawWeight != null)
                          _TuningField(
                              label: 'Draw Weight',
                              value: '${detail.drawWeight} lbs'),
                        if (detail.drawLength != null)
                          _TuningField(
                              label: 'Draw Length',
                              value: '${detail.drawLength}"'),
                        if (detail.arrowFoc != null)
                          _TuningField(
                              label: 'Arrow FOC',
                              value: '${detail.arrowFoc}%'),
                      ],
                    ),
                  ),
                ),
                const SizedBox(height: 24),
              ],

              // Linked equipment
              Row(
                children: [
                  Expanded(
                    child: Text('Equipment (${detail.equipment.length})',
                        style: theme.textTheme.titleMedium),
                  ),
                  TextButton.icon(
                    icon: const Icon(Icons.add, size: 18),
                    label: const Text('Add'),
                    onPressed: () => _showEquipmentPicker(context, ref),
                  ),
                ],
              ),
              const SizedBox(height: 8),
              if (detail.equipment.isEmpty)
                Card(
                  child: Padding(
                    padding: const EdgeInsets.all(24),
                    child: Center(
                      child: Text('No equipment linked',
                          style: theme.textTheme.bodyMedium?.copyWith(
                            color: theme.colorScheme.onSurfaceVariant,
                          )),
                    ),
                  ),
                )
              else
                ...detail.equipment.map((item) => Card(
                      margin: const EdgeInsets.only(bottom: 4),
                      child: ListTile(
                        leading: Icon(item.categoryIcon),
                        title: Text(item.name),
                        subtitle: Text([
                          item.categoryLabel,
                          if (item.brand != null) item.brand!,
                        ].join(' - ')),
                        trailing: IconButton(
                          icon: Icon(Icons.remove_circle_outline,
                              color: theme.colorScheme.error),
                          onPressed: () => _confirmRemove(
                              context, ref, item),
                        ),
                      ),
                    )),
            ],
          ),
        ),
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (err, _) => Center(child: Text('Error: $err')),
      ),
    );
  }

  bool _hasTuning(SetupDetail detail) =>
      detail.braceHeight != null ||
      detail.tiller != null ||
      detail.drawWeight != null ||
      detail.drawLength != null ||
      detail.arrowFoc != null;

  Future<void> _showEquipmentPicker(
      BuildContext context, WidgetRef ref) async {
    final allEquipment = ref.read(equipmentProvider).value ?? [];
    final detail = ref.read(setupDetailProvider(setupId)).value;
    final linkedIds =
        detail?.equipment.map((e) => e.id).toSet() ?? <String>{};
    final available =
        allEquipment.where((e) => !linkedIds.contains(e.id)).toList();

    if (available.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('All equipment is already linked')),
      );
      return;
    }

    final selected = await showModalBottomSheet<List<Equipment>>(
      context: context,
      isScrollControlled: true,
      builder: (_) => EquipmentPicker(available: available),
    );

    if (selected != null && selected.isNotEmpty) {
      for (final item in selected) {
        await addEquipmentToSetup(ref, setupId, item.id);
      }
    }
  }

  Future<void> _confirmRemove(
      BuildContext context, WidgetRef ref, Equipment item) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Remove Equipment'),
        content:
            Text('Remove "${item.name}" from this setup?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            child: const Text('Remove'),
          ),
        ],
      ),
    );
    if (confirmed == true) {
      await removeEquipmentFromSetup(ref, setupId, item.id);
    }
  }
}

class _TuningField extends StatelessWidget {
  final String label;
  final String value;

  const _TuningField({required this.label, required this.value});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      mainAxisSize: MainAxisSize.min,
      children: [
        Text(label,
            style: theme.textTheme.labelSmall?.copyWith(
              color: theme.colorScheme.onSurfaceVariant,
            )),
        Text(value, style: theme.textTheme.bodyLarge),
      ],
    );
  }
}
