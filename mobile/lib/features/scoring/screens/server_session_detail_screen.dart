import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';

import '../../../core/api/api_client.dart';

class ServerSessionDetailScreen extends ConsumerStatefulWidget {
  final String sessionId;

  const ServerSessionDetailScreen({super.key, required this.sessionId});

  @override
  ConsumerState<ServerSessionDetailScreen> createState() =>
      _ServerSessionDetailScreenState();
}

class _ServerSessionDetailScreenState
    extends ConsumerState<ServerSessionDetailScreen> {
  Map<String, dynamic>? _session;
  bool _loading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _loadSession();
  }

  Future<void> _loadSession() async {
    try {
      final api = ref.read(apiClientProvider);
      final response =
          await api.dio.get('/api/v1/sessions/${widget.sessionId}');
      setState(() {
        _session = response.data as Map<String, dynamic>;
        _loading = false;
      });
    } catch (e) {
      setState(() {
        _error = e.toString();
        _loading = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    if (_loading) {
      return const Scaffold(body: Center(child: CircularProgressIndicator()));
    }

    if (_error != null || _session == null) {
      return Scaffold(
        appBar: AppBar(title: const Text('Round Detail')),
        body: Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(Icons.error_outline, size: 48, color: theme.colorScheme.error),
              const SizedBox(height: 16),
              Text('Failed to load round', style: theme.textTheme.titleMedium),
              const SizedBox(height: 8),
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 32),
                child: Text(
                  _error ?? 'Unknown error',
                  style: theme.textTheme.bodySmall,
                  textAlign: TextAlign.center,
                ),
              ),
              const SizedBox(height: 16),
              FilledButton.tonal(
                onPressed: () {
                  setState(() {
                    _loading = true;
                    _error = null;
                  });
                  _loadSession();
                },
                child: const Text('Retry'),
              ),
            ],
          ),
        ),
      );
    }

    final session = _session!;
    final dateFormat = DateFormat.yMMMd().add_jm();
    final startedAt = DateTime.parse(session['started_at'] as String);
    final ends = (session['ends'] as List?) ?? [];
    final templateName =
        (session['template'] as Map?)?['name'] as String? ?? 'Round';

    return Scaffold(
      appBar: AppBar(title: Text(templateName)),
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
                      _StatColumn(
                          label: 'Score',
                          value: '${session['total_score'] ?? 0}'),
                      _StatColumn(
                          label: 'Arrows',
                          value: '${session['total_arrows'] ?? 0}'),
                      _StatColumn(
                          label: 'Xs',
                          value: '${session['total_x_count'] ?? 0}'),
                      _StatColumn(
                          label: 'Ends', value: '${ends.length}'),
                    ],
                  ),
                  const SizedBox(height: 12),
                  Text(
                    dateFormat.format(startedAt),
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
          ...ends.map((endJson) {
            final end = endJson as Map<String, dynamic>;
            final arrows = (end['arrows'] as List?) ?? [];
            final endNumber = end['end_number'] as int? ?? 0;
            final endTotal = end['end_total'] as int? ?? 0;

            return Card(
              margin: const EdgeInsets.only(bottom: 8),
              child: Padding(
                padding:
                    const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
                child: Row(
                  children: [
                    SizedBox(
                      width: 32,
                      child: Text(
                        '$endNumber',
                        style: theme.textTheme.titleSmall?.copyWith(
                          color: theme.colorScheme.outline,
                        ),
                      ),
                    ),
                    Expanded(
                      child: Wrap(
                        spacing: 6,
                        children: arrows.map((a) {
                          final arrow = a as Map<String, dynamic>;
                          final scoreValue =
                              arrow['score_value'] as String? ?? '';
                          return Container(
                            width: 28,
                            height: 28,
                            decoration: BoxDecoration(
                              color: _arrowColor(scoreValue)
                                  .withValues(alpha: 0.2),
                              borderRadius: BorderRadius.circular(4),
                            ),
                            alignment: Alignment.center,
                            child: Text(
                              scoreValue,
                              style: theme.textTheme.bodySmall?.copyWith(
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                          );
                        }).toList(),
                      ),
                    ),
                    Text(
                      '$endTotal',
                      style: theme.textTheme.titleMedium?.copyWith(
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ],
                ),
              ),
            );
          }),
        ],
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
