import 'dart:io';

import 'package:drift/drift.dart' hide Column;
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';

import '../../../core/database/database.dart';
import '../widgets/end_summary_row.dart';

class SessionDetailScreen extends ConsumerStatefulWidget {
  final String sessionId;

  const SessionDetailScreen({super.key, required this.sessionId});

  @override
  ConsumerState<SessionDetailScreen> createState() =>
      _SessionDetailScreenState();
}

class _SessionDetailScreenState extends ConsumerState<SessionDetailScreen> {
  ScoringSessionsLocalData? _session;
  List<EndsLocalData> _ends = [];
  Map<String, List<ArrowsLocalData>> _arrowsByEnd = {};
  Set<String> _endIdsWithImages = {};
  bool _loading = true;

  @override
  void initState() {
    super.initState();
    _loadSession();
  }

  Future<void> _loadSession() async {
    final db = ref.read(databaseProvider);

    final session = await (db.select(db.scoringSessionsLocal)
          ..where((t) => t.id.equals(widget.sessionId)))
        .getSingle();

    final ends = await (db.select(db.endsLocal)
          ..where((t) => t.sessionId.equals(widget.sessionId))
          ..orderBy([(t) => OrderingTerm.asc(t.endNumber)]))
        .get();

    final arrowsByEnd = <String, List<ArrowsLocalData>>{};
    final endIds = <String>{};
    for (final end in ends) {
      endIds.add(end.id);
      final arrows = await (db.select(db.arrowsLocal)
            ..where((t) => t.endId.equals(end.id))
            ..orderBy([(t) => OrderingTerm.asc(t.arrowNumber)]))
          .get();
      arrowsByEnd[end.id] = arrows;
    }

    final images = endIds.isEmpty
        ? <EndImage>[]
        : await (db.select(db.endImages)
              ..where((t) => t.endId.isIn(endIds)))
            .get();

    setState(() {
      _session = session;
      _ends = ends;
      _arrowsByEnd = arrowsByEnd;
      _endIdsWithImages = images.map((i) => i.endId).toSet();
      _loading = false;
    });
  }

  Future<void> _viewEndImage(String endId) async {
    final db = ref.read(databaseProvider);
    final images = await (db.select(db.endImages)
          ..where((t) => t.endId.equals(endId))
          ..orderBy([(t) => OrderingTerm.desc(t.capturedAt)]))
        .get();

    if (images.isEmpty || !mounted) return;

    final file = File(images.first.filePath);
    if (!await file.exists()) {
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
          builder: (_) => _ImageViewer(file: file),
        ),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    if (_loading) {
      return const Scaffold(body: Center(child: CircularProgressIndicator()));
    }

    final session = _session!;
    final dateFormat = DateFormat.yMMMd().add_jm();

    return Scaffold(
      appBar: AppBar(
        title: const Text('Round Detail'),
      ),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          // Summary card
          Card(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                children: [
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceAround,
                    children: [
                      _StatColumn(label: 'Score', value: '${session.totalScore}'),
                      _StatColumn(label: 'Arrows', value: '${session.totalArrows}'),
                      _StatColumn(label: 'Xs', value: '${session.totalXCount}'),
                      _StatColumn(label: 'Ends', value: '${_ends.length}'),
                    ],
                  ),
                  const SizedBox(height: 12),
                  Text(
                    dateFormat.format(session.startedAt),
                    style: theme.textTheme.bodySmall?.copyWith(
                      color: theme.colorScheme.onSurfaceVariant,
                    ),
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 16),
          // End rows
          ..._ends.map((end) {
            final arrows = _arrowsByEnd[end.id] ?? [];
            return EndSummaryRow(
              end: end,
              arrows: arrows,
              hasImage: _endIdsWithImages.contains(end.id),
              onImageTap: _endIdsWithImages.contains(end.id)
                  ? () => _viewEndImage(end.id)
                  : null,
            );
          }),
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

class _ImageViewer extends StatelessWidget {
  final File file;

  const _ImageViewer({required this.file});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.black,
      appBar: AppBar(
        backgroundColor: Colors.black,
        foregroundColor: Colors.white,
        title: const Text('Target Photo'),
      ),
      body: Center(
        child: InteractiveViewer(
          child: Image.file(file),
        ),
      ),
    );
  }
}
