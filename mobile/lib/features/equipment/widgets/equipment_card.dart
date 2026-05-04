import 'package:flutter/material.dart';

import '../models/equipment.dart';

class EquipmentCard extends StatelessWidget {
  final Equipment equipment;
  final VoidCallback onEdit;
  final VoidCallback onDelete;

  const EquipmentCard({
    super.key,
    required this.equipment,
    required this.onEdit,
    required this.onDelete,
  });

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: const EdgeInsets.only(bottom: 4),
      child: ListTile(
        leading: Icon(equipment.categoryIcon),
        title: Text(equipment.name),
        subtitle: Text([
          if (equipment.brand != null) equipment.brand!,
          if (equipment.model != null) equipment.model!,
        ].join(' - ')),
        trailing: PopupMenuButton<String>(
          onSelected: (value) {
            if (value == 'edit') onEdit();
            if (value == 'delete') onDelete();
          },
          itemBuilder: (context) => [
            const PopupMenuItem(value: 'edit', child: Text('Edit')),
            PopupMenuItem(
              value: 'delete',
              child: Text('Delete',
                  style: TextStyle(
                      color: Theme.of(context).colorScheme.error)),
            ),
          ],
        ),
      ),
    );
  }
}
