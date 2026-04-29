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
  Map<String, int> _imageCountByEnd = {};
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

    final counts = <String, int>{};
    for (final img in images) {
      counts[img.endId] = (counts[img.endId] ?? 0) + 1;
    }

    setState(() {
      _session = session;
      _ends = ends;
      _arrowsByEnd = arrowsByEnd;
      _imageCountByEnd = counts;
      _loading = false;
    });
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
            final count = _imageCountByEnd[end.id] ?? 0;
            return EndSummaryRow(
              end: end,
              arrows: arrows,
              imageCount: count,
              onImageTap: count > 0
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
