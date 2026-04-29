import 'dart:convert';
import 'dart:io';

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
  Map<String, int> _imageCountByEnd = {};
  Map<String, bool> _imageSyncedByEnd = {};

  @override
  void initState() {
    super.initState();
    _loadStageInfo();
    _loadImageCounts();
  }

  Future<void> _loadImageCounts() async {
    final db = ref.read(databaseProvider);
    final session = ref.read(scoringProvider).activeSession;
    if (session == null) return;

    final ends = ref.read(scoringProvider).ends;
    final endIds = ends.map((e) => e.id).toSet();
    if (endIds.isEmpty) return;

    final images = await (db.select(db.endImages)
          ..where((t) => t.endId.isIn(endIds)))
        .get();

    final counts = <String, int>{};
    final synced = <String, bool>{};
    for (final img in images) {
      counts[img.endId] = (counts[img.endId] ?? 0) + 1;
      synced[img.endId] = (synced[img.endId] ?? true) && img.synced;
    }

    setState(() {
      _imageCountByEnd = counts;
      _imageSyncedByEnd = synced;
    });
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

    // All ends complete — no more arrow input needed
    // (completion dialog is handled by _submitEnd after photo capture)
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

    // Check if all ends are now complete BEFORE offering photo
    final isRoundComplete = await _checkAllEndsComplete();

    _loadImageCounts();

    if (isRoundComplete) {
      if (mounted) _showCompleteDialog();
    } else {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('End $endNumber submitted'),
            duration: const Duration(seconds: 2),
          ),
        );
      }
      _loadStageInfo();
    }
  }

  /// Check if all ends for all stages are complete without triggering the dialog
  Future<bool> _checkAllEndsComplete() async {
    final db = ref.read(databaseProvider);
    final session = ref.read(scoringProvider).activeSession;
    if (session == null) return false;

    final stages = await (db.select(db.stages)
          ..where((t) => t.templateId.equals(session.templateId))
          ..orderBy([(t) => OrderingTerm.asc(t.stageOrder)]))
        .get();

    final totalEndsRequired = stages.fold<int>(0, (sum, s) => sum + s.numEnds);
    final endsCompleted = ref.read(scoringProvider).ends.length;

    return endsCompleted >= totalEndsRequired;
  }

  Future<void> _addPhotoToEnd(String endId) async {
    await Navigator.of(context).push(
      MaterialPageRoute(
        builder: (_) => CaptureScreen(endId: endId),
      ),
    );
    _loadImageCounts();
  }

  Future<void> _viewEndImage(String endId) async {
    final db = ref.read(databaseProvider);
    final images = await (db.select(db.endImages)
          ..where((t) => t.endId.equals(endId))
          ..orderBy([(t) => OrderingTerm.asc(t.capturedAt)]))
        .get();

    if (images.isEmpty || !mounted) return;

    final files = <File>[];
    for (final img in images) {
      final f = File(img.filePath);
      if (await f.exists()) files.add(f);
    }

    if (files.isEmpty) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Image file not found')),
        );
      }
      return;
    }

    if (mounted) {
      Navigator.of(context).push(
        MaterialPageRoute(
          builder: (_) => _ImageGallery(files: files),
        ),
      );
    }
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
          if (scoringState.ends.isNotEmpty)
            IconButton(
              onPressed: () async {
                final confirm = await showDialog<bool>(
                  context: context,
                  builder: (ctx) => AlertDialog(
                    title: const Text('Undo Last End?'),
                    content: Text(
                      'Remove end ${scoringState.ends.length} '
                      '(${scoringState.ends.last.endTotal} pts)?',
                    ),
                    actions: [
                      TextButton(
                        onPressed: () => Navigator.of(ctx).pop(false),
                        child: const Text('Cancel'),
                      ),
                      FilledButton(
                        onPressed: () => Navigator.of(ctx).pop(true),
                        child: const Text('Undo'),
                      ),
                    ],
                  ),
                );
                if (confirm == true) {
                  final removed =
                      await ref.read(scoringProvider.notifier).undoLastEnd();
                  if (removed && mounted) {
                    _loadStageInfo();
                    _loadImageCounts();
                    ScaffoldMessenger.of(context).showSnackBar(
                      const SnackBar(content: Text('End removed')),
                    );
                  }
                }
              },
              icon: const Icon(Icons.undo),
              tooltip: 'Undo last end',
            ),
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
                  final count = _imageCountByEnd[end.id] ?? 0;
                  final synced = _imageSyncedByEnd[end.id] ?? true;
                  return EndSummaryRow(
                    end: end,
                    arrows: arrows,
                    imageCount: count,
                    imageSynced: synced,
                    onImageTap: count > 0
                        ? () => _viewEndImage(end.id)
                        : null,
                    onAddPhoto: () => _addPhotoToEnd(end.id),
                  );
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

class _ImageGallery extends StatefulWidget {
  final List<File> files;

  const _ImageGallery({required this.files});

  @override
  State<_ImageGallery> createState() => _ImageGalleryState();
}

class _ImageGalleryState extends State<_ImageGallery> {
  int _current = 0;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.black,
      appBar: AppBar(
        backgroundColor: Colors.black,
        foregroundColor: Colors.white,
        title: Text(widget.files.length > 1
            ? 'Photo ${_current + 1} of ${widget.files.length}'
            : 'Target Photo'),
      ),
      body: PageView.builder(
        itemCount: widget.files.length,
        onPageChanged: (i) => setState(() => _current = i),
        itemBuilder: (_, i) => Center(
          child: InteractiveViewer(
            child: Image.file(widget.files[i]),
          ),
        ),
      ),
    );
  }
}
