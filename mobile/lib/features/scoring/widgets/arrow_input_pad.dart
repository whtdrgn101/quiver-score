import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

class ArrowInputPad extends StatelessWidget {
  final List<String> allowedValues;
  final ValueChanged<String> onValueTap;
  final VoidCallback onBackspace;
  final VoidCallback? onSubmit;

  const ArrowInputPad({
    super.key,
    required this.allowedValues,
    required this.onValueTap,
    required this.onBackspace,
    this.onSubmit,
  });

  Color _getValueColor(String value) {
    return switch (value) {
      'X' || '10' => Colors.yellow.shade700,
      '9' => Colors.yellow.shade600,
      '8' || '7' => Colors.red.shade400,
      '6' || '5' => Colors.blue.shade400,
      '4' || '3' => Colors.black54,
      '2' || '1' => Colors.white70,
      'M' => Colors.grey,
      _ => Colors.grey,
    };
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Column(
      children: [
        Wrap(
          spacing: 8,
          runSpacing: 8,
          alignment: WrapAlignment.center,
          children: allowedValues.map((value) {
            return SizedBox(
              width: 52,
              height: 44,
              child: ElevatedButton(
                onPressed: () {
                  HapticFeedback.lightImpact();
                  onValueTap(value);
                },
                style: ElevatedButton.styleFrom(
                  padding: EdgeInsets.zero,
                  backgroundColor: _getValueColor(value).withValues(alpha: 0.2),
                  foregroundColor: theme.colorScheme.onSurface,
                ),
                child: Text(
                  value,
                  style: const TextStyle(
                    fontWeight: FontWeight.bold,
                    fontSize: 16,
                  ),
                ),
              ),
            );
          }).toList(),
        ),
        const SizedBox(height: 12),
        Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            OutlinedButton.icon(
              onPressed: () {
                HapticFeedback.mediumImpact();
                onBackspace();
              },
              icon: const Icon(Icons.backspace_outlined, size: 18),
              label: const Text('Undo'),
            ),
            const SizedBox(width: 16),
            FilledButton.icon(
              onPressed: onSubmit != null
                  ? () {
                      HapticFeedback.heavyImpact();
                      onSubmit!();
                    }
                  : null,
              icon: const Icon(Icons.check),
              label: const Text('Submit End'),
            ),
          ],
        ),
      ],
    );
  }
}
