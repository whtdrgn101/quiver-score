import 'package:flutter/material.dart';

import '../../../core/database/database.dart';

class EndSummaryRow extends StatelessWidget {
  final EndsLocalData end;
  final List<ArrowsLocalData> arrows;
  final int imageCount;
  final VoidCallback? onImageTap;
  final VoidCallback? onAddPhoto;

  const EndSummaryRow({
    super.key,
    required this.end,
    required this.arrows,
    this.imageCount = 0,
    this.onImageTap,
    this.onAddPhoto,
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
            SizedBox(
              width: 32,
              child: Text(
                '${end.endNumber}',
                style: theme.textTheme.titleSmall?.copyWith(
                  color: theme.colorScheme.outline,
                ),
              ),
            ),
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
            if (imageCount > 0)
              Padding(
                padding: const EdgeInsets.only(right: 4),
                child: GestureDetector(
                  onTap: onImageTap,
                  child: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Icon(Icons.photo, size: 18,
                          color: theme.colorScheme.primary),
                      if (imageCount > 1) ...[
                        const SizedBox(width: 2),
                        Text('$imageCount',
                            style: theme.textTheme.labelSmall?.copyWith(
                              color: theme.colorScheme.primary,
                            )),
                      ],
                    ],
                  ),
                ),
              ),
            if (onAddPhoto != null)
              Padding(
                padding: const EdgeInsets.only(right: 4),
                child: GestureDetector(
                  onTap: onAddPhoto,
                  child: Icon(Icons.add_a_photo_outlined, size: 18,
                      color: theme.colorScheme.outline),
                ),
              ),
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
