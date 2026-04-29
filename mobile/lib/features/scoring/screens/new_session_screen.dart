import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/sync/sync_service.dart';
import '../../rounds/providers/rounds_provider.dart';
import '../providers/scoring_provider.dart';
import 'scoring_screen.dart';

class NewSessionScreen extends ConsumerStatefulWidget {
  const NewSessionScreen({super.key});

  @override
  ConsumerState<NewSessionScreen> createState() => _NewSessionScreenState();
}

class _NewSessionScreenState extends ConsumerState<NewSessionScreen> {
  String? _selectedTemplateId;
  final _notesController = TextEditingController();
  final _locationController = TextEditingController();

  @override
  void dispose() {
    _notesController.dispose();
    _locationController.dispose();
    super.dispose();
  }

  Future<void> _startSession() async {
    if (_selectedTemplateId == null) return;

    await ref.read(scoringProvider.notifier).startSession(
          templateId: _selectedTemplateId!,
          notes: _notesController.text.isNotEmpty
              ? _notesController.text
              : null,
          location: _locationController.text.isNotEmpty
              ? _locationController.text
              : null,
        );

    if (mounted) {
      Navigator.of(context).pushReplacement(
        MaterialPageRoute(builder: (_) => const ScoringScreen()),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final templates = ref.watch(roundTemplatesProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('New Round'),
      ),
      body: templates.when(
        data: (templateList) => RefreshIndicator(
          onRefresh: () async {
            await ref.read(syncServiceProvider).pullRoundTemplates();
          },
          child: ListView(
            padding: const EdgeInsets.all(16),
            children: [
              DropdownButtonFormField<String>(
                initialValue: _selectedTemplateId,
                decoration: const InputDecoration(
                  labelText: 'Round Type',
                  border: OutlineInputBorder(),
                  prefixIcon: Icon(Icons.track_changes_outlined),
                ),
                isExpanded: true,
                items: templateList
                    .map((t) => DropdownMenuItem(
                          value: t.id,
                          child: Text(t.name, overflow: TextOverflow.ellipsis),
                        ))
                    .toList(),
                onChanged: (v) => setState(() => _selectedTemplateId = v),
              ),

              if (_selectedTemplateId != null) ...[
                const SizedBox(height: 12),
                _TemplateDetails(templateId: _selectedTemplateId!),
              ],

              const SizedBox(height: 24),
              TextField(
                controller: _locationController,
                decoration: const InputDecoration(
                  labelText: 'Location (optional)',
                  border: OutlineInputBorder(),
                  prefixIcon: Icon(Icons.location_on_outlined),
                ),
              ),
              const SizedBox(height: 16),
              TextField(
                controller: _notesController,
                decoration: const InputDecoration(
                  labelText: 'Notes (optional)',
                  border: OutlineInputBorder(),
                  prefixIcon: Icon(Icons.notes_outlined),
                ),
                maxLines: 2,
              ),
              const SizedBox(height: 32),
              FilledButton.icon(
                onPressed: _selectedTemplateId != null ? _startSession : null,
                icon: const Icon(Icons.play_arrow),
                label: const Text('Start Scoring'),
              ),
            ],
          ),
        ),
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('Error loading rounds: $e')),
      ),
    );
  }
}

class _TemplateDetails extends ConsumerWidget {
  final String templateId;

  const _TemplateDetails({required this.templateId});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final stagesAsync = ref.watch(stagesForTemplateProvider(templateId));
    final theme = Theme.of(context);

    return stagesAsync.when(
      data: (stages) {
        if (stages.isEmpty) return const SizedBox.shrink();

        final totalEnds =
            stages.fold<int>(0, (sum, s) => sum + s.numEnds.value);
        final totalArrows = stages.fold<int>(
            0, (sum, s) => sum + s.numEnds.value * s.arrowsPerEnd.value);
        final maxScore = stages.fold<int>(
            0,
            (sum, s) =>
                sum +
                s.numEnds.value *
                    s.arrowsPerEnd.value *
                    s.maxScorePerArrow.value);

        return Card(
          child: Padding(
            padding: const EdgeInsets.all(12),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceAround,
                  children: [
                    _DetailChip(
                        label: 'Ends', value: '$totalEnds', theme: theme),
                    _DetailChip(
                        label: 'Arrows', value: '$totalArrows', theme: theme),
                    _DetailChip(
                        label: 'Max', value: '$maxScore', theme: theme),
                  ],
                ),
                if (stages.length > 1) ...[
                  const SizedBox(height: 12),
                  const Divider(height: 1),
                  const SizedBox(height: 8),
                  ...stages.map((s) => Padding(
                        padding: const EdgeInsets.only(bottom: 4),
                        child: Row(
                          children: [
                            Icon(Icons.straighten,
                                size: 16,
                                color: theme.colorScheme.onSurfaceVariant),
                            const SizedBox(width: 8),
                            Expanded(
                              child: Text(
                                s.distance.value ?? s.name.value,
                                style: theme.textTheme.bodyMedium,
                              ),
                            ),
                            Text(
                              '${s.numEnds.value} ends × ${s.arrowsPerEnd.value} arrows',
                              style: theme.textTheme.bodySmall?.copyWith(
                                color: theme.colorScheme.onSurfaceVariant,
                              ),
                            ),
                          ],
                        ),
                      )),
                ] else ...[
                  if (stages.first.distance.value != null) ...[
                    const SizedBox(height: 8),
                    Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Icon(Icons.straighten,
                            size: 16,
                            color: theme.colorScheme.onSurfaceVariant),
                        const SizedBox(width: 4),
                        Text(
                          stages.first.distance.value!,
                          style: theme.textTheme.bodyMedium?.copyWith(
                            color: theme.colorScheme.onSurfaceVariant,
                          ),
                        ),
                        const SizedBox(width: 12),
                        Text(
                          '${stages.first.arrowsPerEnd.value} arrows per end',
                          style: theme.textTheme.bodySmall?.copyWith(
                            color: theme.colorScheme.onSurfaceVariant,
                          ),
                        ),
                      ],
                    ),
                  ],
                ],
              ],
            ),
          ),
        );
      },
      loading: () => const Padding(
        padding: EdgeInsets.symmetric(vertical: 8),
        child: Center(child: CircularProgressIndicator(strokeWidth: 2)),
      ),
      error: (_, _) => const SizedBox.shrink(),
    );
  }
}

class _DetailChip extends StatelessWidget {
  final String label;
  final String value;
  final ThemeData theme;

  const _DetailChip({
    required this.label,
    required this.value,
    required this.theme,
  });

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        Text(value,
            style: theme.textTheme.titleMedium
                ?.copyWith(fontWeight: FontWeight.bold)),
        Text(label,
            style: theme.textTheme.bodySmall
                ?.copyWith(color: theme.colorScheme.onSurfaceVariant)),
      ],
    );
  }
}
