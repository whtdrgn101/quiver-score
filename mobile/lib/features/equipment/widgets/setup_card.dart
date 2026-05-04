import 'package:flutter/material.dart';

import '../models/setup.dart';

class SetupCard extends StatelessWidget {
  final SetupSummary setup;
  final VoidCallback onTap;
  final VoidCallback onDelete;

  const SetupCard({
    super.key,
    required this.setup,
    required this.onTap,
    required this.onDelete,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Card(
      margin: const EdgeInsets.only(bottom: 8),
      child: ListTile(
        onTap: onTap,
        leading: const Icon(Icons.tune_outlined),
        title: Text(setup.name),
        subtitle: Text(
          setup.description ?? '${setup.equipmentCount} item${setup.equipmentCount == 1 ? '' : 's'}',
          maxLines: 1,
          overflow: TextOverflow.ellipsis,
        ),
        trailing: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            if (setup.equipmentCount > 0)
              Badge(
                label: Text('${setup.equipmentCount}'),
                backgroundColor: theme.colorScheme.primaryContainer,
                textColor: theme.colorScheme.onPrimaryContainer,
              ),
            const SizedBox(width: 4),
            const Icon(Icons.chevron_right, size: 20),
          ],
        ),
      ),
    );
  }
}
