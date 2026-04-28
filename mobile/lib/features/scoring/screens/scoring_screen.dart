import 'dart:convert';

import 'package:drift/drift.dart' hide Column;
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/database/database.dart';
import '../providers/scoring_provider.dart';
import '../widgets/arrow_input_pad.dart';
import '../widgets/end_summary_row.dart';
import '../../images/screens/capture_screen.dart';

class ScoringScreen extends ConsumerStatefulWidget {
  const ScoringScreen({super.key});

  @override
  ConsumerState<ScoringScreen> createState() => _ScoringScreenState();
}

class _ScoringScreenState extends ConsumerState<ScoringScreen> {
  List<ArrowInput> _currentArrows = [];
  int _arrowsPerEnd = 0;
  List<String> _allowedValues = [];
  Map<String, int> _valueScoreMap = {};
  String? _currentStageId;

  @override
  void initState() {
    super.initState();
    _loadStageInfo();
  }

  Future<void> _loadStageInfo() async {
    final db = ref.read(databaseProvider);
    final session = ref.read(scoringProvider).activeSession;
    if (session == null) return;

    // Get stages for this template
    final stages = await (db.select(db.stages)
          ..where((t) => t.templateId.equals(session.templateId))
          ..orderBy([(t) => OrderingTerm.asc(t.stageOrder)]))
        .get();

    if (stages.isEmpty) return;

    // Determine current stage based on ends completed
    final ends = ref.read(scoringProvider).ends;
    var endsCompleted = ends.length;

    var endsSoFar = 0;
    for (final stage in stages) {
      if (endsCompleted < endsSoFar + stage.numEnds) {
        setState(() {
          _currentStageId = stage.id;
          _arrowsPerEnd = stage.arrowsPerEnd;
          _allowedValues =
              (jsonDecode(stage.allowedValues) as List).cast<String>();
          _valueScoreMap =
              (jsonDecode(stage.valueScoreMap) as Map<String, dynamic>)
                  .map((k, v) => MapEntry(k, v as int));
        });
        return;
      }
      endsSoFar += stage.numEnds;
    }

    // All ends complete — session can be completed
    _showCompleteDialog();
  }

  void _addArrow(String value) {
    if (_currentArrows.length >= _arrowsPerEnd) return;
    setState(() {
      _currentArrows.add(ArrowInput(scoreValue: value));
    });
  }

  void _removeLastArrow() {
    if (_currentArrows.isEmpty) return;
    setState(() {
      _currentArrows.removeLast();
    });
  }

  Future<void> _submitEnd() async {
    if (_currentArrows.length != _arrowsPerEnd) return;
    if (_currentStageId == null) return;

    final scoringState = ref.read(scoringProvider);
    final endNumber = scoringState.ends.length + 1;

    await ref.read(scoringProvider.notifier).submitEnd(
          stageId: _currentStageId!,
          endNumber: endNumber,
          arrows: _currentArrows,
          valueScoreMap: _valueScoreMap,
        );

    setState(() {
      _currentArrows = [];
    });

    // Offer to take a photo
    if (mounted) {
      _offerPhotoCapture(endNumber);
    }

    // Reload stage info (may have moved to next stage)
    _loadStageInfo();
  }

  void _offerPhotoCapture(int endNumber) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text('End $endNumber submitted'),
        action: SnackBarAction(
          label: 'Take Photo',
          onPressed: () {
            final session = ref.read(scoringProvider).activeSession;
            final ends = ref.read(scoringProvider).ends;
            if (session != null && ends.isNotEmpty) {
              Navigator.of(context).push(
                MaterialPageRoute(
                  builder: (_) => CaptureScreen(endId: ends.last.id),
                ),
              );
            }
          },
        ),
        duration: const Duration(seconds: 4),
      ),
    );
  }

  void _showCompleteDialog() {
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Round Complete'),
        content: const Text('All ends have been scored. Complete this round?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(),
            child: const Text('Keep Open'),
          ),
          FilledButton(
            onPressed: () async {
              Navigator.of(ctx).pop();
              await ref.read(scoringProvider.notifier).completeSession();
              if (mounted) Navigator.of(context).pop();
            },
            child: const Text('Complete'),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final scoringState = ref.watch(scoringProvider);
    final session = scoringState.activeSession;
    final theme = Theme.of(context);

    if (session == null) {
      return const Scaffold(body: Center(child: CircularProgressIndicator()));
    }

    final endTotal = _currentArrows.fold<int>(
      0,
      (sum, a) => sum + (_valueScoreMap[a.scoreValue] ?? 0),
    );

    return Scaffold(
      appBar: AppBar(
        title: Text('Score: ${session.totalScore}'),
        actions: [
          PopupMenuButton<String>(
            onSelected: (value) async {
              if (value == 'abandon') {
                final confirm = await showDialog<bool>(
                  context: context,
                  builder: (ctx) => AlertDialog(
                    title: const Text('Abandon Round?'),
                    content: const Text('This cannot be undone.'),
                    actions: [
                      TextButton(
                        onPressed: () => Navigator.of(ctx).pop(false),
                        child: const Text('Cancel'),
                      ),
                      FilledButton(
                        onPressed: () => Navigator.of(ctx).pop(true),
                        child: const Text('Abandon'),
                      ),
                    ],
                  ),
                );
                if (confirm == true) {
                  await ref.read(scoringProvider.notifier).abandonSession();
                  if (mounted) Navigator.of(context).pop();
                }
              } else if (value == 'complete') {
                await ref.read(scoringProvider.notifier).completeSession();
                if (mounted) Navigator.of(context).pop();
              }
            },
            itemBuilder: (_) => [
              const PopupMenuItem(value: 'complete', child: Text('Complete Round')),
              const PopupMenuItem(value: 'abandon', child: Text('Abandon Round')),
            ],
          ),
        ],
      ),
      body: Column(
        children: [
          // Running score and end history
          Expanded(
            child: ListView(
              padding: const EdgeInsets.all(16),
              children: [
                // Summary header
                Card(
                  child: Padding(
                    padding: const EdgeInsets.all(16),
                    child: Row(
                      mainAxisAlignment: MainAxisAlignment.spaceAround,
                      children: [
                        _StatColumn(label: 'Score', value: '${session.totalScore}'),
                        _StatColumn(label: 'Arrows', value: '${session.totalArrows}'),
                        _StatColumn(label: 'Xs', value: '${session.totalXCount}'),
                        _StatColumn(
                          label: 'End',
                          value: '${scoringState.ends.length + 1}',
                        ),
                      ],
                    ),
                  ),
                ),
                const SizedBox(height: 16),
                // End history
                ...scoringState.ends.reversed.map((end) {
                  final arrows = scoringState.arrowsByEnd[end.id] ?? [];
                  return EndSummaryRow(end: end, arrows: arrows);
                }),
              ],
            ),
          ),

          // Current end input
          Container(
            decoration: BoxDecoration(
              color: theme.colorScheme.surfaceContainerHighest,
              borderRadius:
                  const BorderRadius.vertical(top: Radius.circular(16)),
            ),
            padding: const EdgeInsets.all(16),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                // Current arrows display
                Row(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    for (var i = 0; i < _arrowsPerEnd; i++)
                      Padding(
                        padding: const EdgeInsets.symmetric(horizontal: 4),
                        child: Container(
                          width: 40,
                          height: 40,
                          decoration: BoxDecoration(
                            color: i < _currentArrows.length
                                ? theme.colorScheme.primaryContainer
                                : theme.colorScheme.surface,
                            borderRadius: BorderRadius.circular(8),
                            border: Border.all(
                              color: theme.colorScheme.outline,
                            ),
                          ),
                          alignment: Alignment.center,
                          child: Text(
                            i < _currentArrows.length
                                ? _currentArrows[i].scoreValue
                                : '',
                            style: theme.textTheme.titleMedium?.copyWith(
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                        ),
                      ),
                    const SizedBox(width: 16),
                    Text(
                      '= $endTotal',
                      style: theme.textTheme.titleMedium?.copyWith(
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 12),
                // Arrow input pad
                ArrowInputPad(
                  allowedValues: _allowedValues,
                  onValueTap: _addArrow,
                  onBackspace: _removeLastArrow,
                  onSubmit: _currentArrows.length == _arrowsPerEnd
                      ? _submitEnd
                      : null,
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _StatColumn extends StatelessWidget {
  final String label;
  final String value;

  const _StatColumn({required this.label, required this.value});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Column(
      children: [
        Text(value,
            style: theme.textTheme.titleLarge
                ?.copyWith(fontWeight: FontWeight.bold)),
        Text(label, style: theme.textTheme.bodySmall),
      ],
    );
  }
}
