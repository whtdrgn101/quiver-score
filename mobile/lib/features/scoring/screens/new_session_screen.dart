import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

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
    final theme = Theme.of(context);

    return Scaffold(
      appBar: AppBar(
        title: const Text('New Round'),
      ),
      body: templates.when(
        data: (templateList) => ListView(
          padding: const EdgeInsets.all(16),
          children: [
            Text('Select Round', style: theme.textTheme.titleMedium),
            const SizedBox(height: 12),
            ...templateList.map((t) => ListTile(
                  title: Text(t.name),
                  subtitle: Text(t.organization),
                  leading: Radio<String>(
                    value: t.id,
                    groupValue: _selectedTemplateId,
                    onChanged: (v) => setState(() => _selectedTemplateId = v),
                  ),
                  onTap: () => setState(() => _selectedTemplateId = t.id),
                )),
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
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('Error loading rounds: $e')),
      ),
    );
  }
}
