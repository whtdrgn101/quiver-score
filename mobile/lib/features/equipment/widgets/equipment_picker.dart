import 'package:flutter/material.dart';

import '../models/equipment.dart';

class EquipmentPicker extends StatefulWidget {
  final List<Equipment> available;

  const EquipmentPicker({super.key, required this.available});

  @override
  State<EquipmentPicker> createState() => _EquipmentPickerState();
}

class _EquipmentPickerState extends State<EquipmentPicker> {
  final _selected = <String>{};

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return DraggableScrollableSheet(
      initialChildSize: 0.5,
      minChildSize: 0.3,
      maxChildSize: 0.8,
      expand: false,
      builder: (context, scrollController) => Column(
        children: [
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 16, 8, 8),
            child: Row(
              children: [
                Expanded(
                  child: Text('Add Equipment',
                      style: theme.textTheme.titleMedium),
                ),
                TextButton(
                  onPressed: _selected.isEmpty
                      ? null
                      : () {
                          final items = widget.available
                              .where((e) => _selected.contains(e.id))
                              .toList();
                          Navigator.pop(context, items);
                        },
                  child: Text('Add (${_selected.length})'),
                ),
              ],
            ),
          ),
          const Divider(height: 1),
          Expanded(
            child: ListView.builder(
              controller: scrollController,
              itemCount: widget.available.length,
              itemBuilder: (context, index) {
                final item = widget.available[index];
                final isSelected = _selected.contains(item.id);
                return CheckboxListTile(
                  value: isSelected,
                  onChanged: (v) {
                    setState(() {
                      if (v == true) {
                        _selected.add(item.id);
                      } else {
                        _selected.remove(item.id);
                      }
                    });
                  },
                  secondary: Icon(item.categoryIcon),
                  title: Text(item.name),
                  subtitle: Text([
                    item.categoryLabel,
                    if (item.brand != null) item.brand!,
                  ].join(' - ')),
                );
              },
            ),
          ),
        ],
      ),
    );
  }
}
