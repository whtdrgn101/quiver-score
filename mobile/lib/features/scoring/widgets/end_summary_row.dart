import 'package:flutter/material.dart';

import '../../../core/database/database.dart';

class EndSummaryRow extends StatelessWidget {
  final EndsLocalData end;
  final List<ArrowsLocalData> arrows;
  final bool hasImage;
  final VoidCallback? onImageTap;

  const EndSummaryRow({
    super.key,
    required this.end,
    required this.arrows,
    this.hasImage = false,
    this.onImageTap,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Card(
      margin: const EdgeInsets.only(bottom: 8),
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
        child: Row(
          children: [
            // End number
            SizedBox(
              width: 32,
              child: Text(
                '${end.endNumber}',
                style: theme.textTheme.titleSmall?.copyWith(
                  color: theme.colorScheme.outline,
                ),
              ),
            ),
            // Arrow values
            Expanded(
              child: Wrap(
                spacing: 6,
                children: arrows.map((a) {
                  return Container(
                    width: 28,
                    height: 28,
                    decoration: BoxDecoration(
                      color: _arrowColor(a.scoreValue).withValues(alpha: 0.2),
                      borderRadius: BorderRadius.circular(4),
                    ),
                    alignment: Alignment.center,
                    child: Text(
                      a.scoreValue,
                      style: theme.textTheme.bodySmall?.copyWith(
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  );
                }).toList(),
              ),
            ),
            // Photo indicator
            if (hasImage)
              Padding(
                padding: const EdgeInsets.only(right: 8),
                child: GestureDetector(
                  onTap: onImageTap,
                  child: Icon(
                    Icons.camera_alt,
                    size: 18,
                    color: theme.colorScheme.primary,
                  ),
                ),
              ),
            // End total
            Text(
              '${end.endTotal}',
              style: theme.textTheme.titleMedium?.copyWith(
                fontWeight: FontWeight.bold,
              ),
            ),
          ],
        ),
      ),
    );
  }

  Color _arrowColor(String value) {
    return switch (value) {
      'X' || '10' || '9' => Colors.yellow.shade700,
      '8' || '7' => Colors.red.shade400,
      '6' || '5' => Colors.blue.shade400,
      '4' || '3' => Colors.black54,
      '2' || '1' => Colors.brown.shade200,
      'M' => Colors.grey,
      _ => Colors.grey,
    };
  }
}
