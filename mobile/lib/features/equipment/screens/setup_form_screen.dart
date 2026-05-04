import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../models/setup.dart';
import '../providers/setup_provider.dart';

class SetupFormScreen extends ConsumerStatefulWidget {
  final SetupDetail? detail;

  const SetupFormScreen({super.key, this.detail});

  @override
  ConsumerState<SetupFormScreen> createState() => _SetupFormScreenState();
}

class _SetupFormScreenState extends ConsumerState<SetupFormScreen> {
  final _formKey = GlobalKey<FormState>();
  late final TextEditingController _nameController;
  late final TextEditingController _descriptionController;
  late final TextEditingController _braceHeightController;
  late final TextEditingController _tillerController;
  late final TextEditingController _drawWeightController;
  late final TextEditingController _drawLengthController;
  late final TextEditingController _arrowFocController;
  bool _saving = false;

  bool get _isEditing => widget.detail != null;

  @override
  void initState() {
    super.initState();
    _nameController = TextEditingController(text: widget.detail?.name);
    _descriptionController =
        TextEditingController(text: widget.detail?.description);
    _braceHeightController =
        TextEditingController(text: widget.detail?.braceHeight?.toString());
    _tillerController =
        TextEditingController(text: widget.detail?.tiller?.toString());
    _drawWeightController =
        TextEditingController(text: widget.detail?.drawWeight?.toString());
    _drawLengthController =
        TextEditingController(text: widget.detail?.drawLength?.toString());
    _arrowFocController =
        TextEditingController(text: widget.detail?.arrowFoc?.toString());
  }

  @override
  void dispose() {
    _nameController.dispose();
    _descriptionController.dispose();
    _braceHeightController.dispose();
    _tillerController.dispose();
    _drawWeightController.dispose();
    _drawLengthController.dispose();
    _arrowFocController.dispose();
    super.dispose();
  }

  double? _parseDouble(String text) {
    final trimmed = text.trim();
    if (trimmed.isEmpty) return null;
    return double.tryParse(trimmed);
  }

  Future<void> _save() async {
    if (!_formKey.currentState!.validate()) return;
    setState(() => _saving = true);

    final data = <String, dynamic>{
      'name': _nameController.text.trim(),
    };

    if (_descriptionController.text.trim().isNotEmpty) {
      data['description'] = _descriptionController.text.trim();
    }
    final bh = _parseDouble(_braceHeightController.text);
    if (bh != null) data['brace_height'] = bh;
    final tiller = _parseDouble(_tillerController.text);
    if (tiller != null) data['tiller'] = tiller;
    final dw = _parseDouble(_drawWeightController.text);
    if (dw != null) data['draw_weight'] = dw;
    final dl = _parseDouble(_drawLengthController.text);
    if (dl != null) data['draw_length'] = dl;
    final foc = _parseDouble(_arrowFocController.text);
    if (foc != null) data['arrow_foc'] = foc;

    try {
      if (_isEditing) {
        await updateSetup(ref, widget.detail!.id, data);
      } else {
        await ref.read(setupListProvider.notifier).create(data);
      }
      if (mounted) Navigator.pop(context);
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Failed to save: $e')),
        );
      }
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(_isEditing ? 'Edit Setup' : 'New Setup'),
        actions: [
          TextButton(
            onPressed: _saving ? null : _save,
            child: _saving
                ? const SizedBox(
                    width: 20,
                    height: 20,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : const Text('Save'),
          ),
        ],
      ),
      body: Form(
        key: _formKey,
        child: ListView(
          padding: const EdgeInsets.all(16),
          children: [
            TextFormField(
              controller: _nameController,
              decoration: const InputDecoration(
                labelText: 'Name *',
                border: OutlineInputBorder(),
              ),
              validator: (v) =>
                  (v == null || v.trim().isEmpty) ? 'Name is required' : null,
              textCapitalization: TextCapitalization.words,
            ),
            const SizedBox(height: 16),
            TextFormField(
              controller: _descriptionController,
              decoration: const InputDecoration(
                labelText: 'Description',
                border: OutlineInputBorder(),
                alignLabelWithHint: true,
              ),
              maxLines: 2,
              textCapitalization: TextCapitalization.sentences,
            ),
            const SizedBox(height: 24),
            Text('Tuning',
                style: Theme.of(context).textTheme.titleMedium),
            const SizedBox(height: 12),
            Row(
              children: [
                Expanded(
                  child: TextFormField(
                    controller: _braceHeightController,
                    decoration: const InputDecoration(
                      labelText: 'Brace Height (in)',
                      border: OutlineInputBorder(),
                    ),
                    keyboardType:
                        const TextInputType.numberWithOptions(decimal: true),
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: TextFormField(
                    controller: _tillerController,
                    decoration: const InputDecoration(
                      labelText: 'Tiller (in)',
                      border: OutlineInputBorder(),
                    ),
                    keyboardType:
                        const TextInputType.numberWithOptions(decimal: true),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 16),
            Row(
              children: [
                Expanded(
                  child: TextFormField(
                    controller: _drawWeightController,
                    decoration: const InputDecoration(
                      labelText: 'Draw Weight (lbs)',
                      border: OutlineInputBorder(),
                    ),
                    keyboardType:
                        const TextInputType.numberWithOptions(decimal: true),
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: TextFormField(
                    controller: _drawLengthController,
                    decoration: const InputDecoration(
                      labelText: 'Draw Length (in)',
                      border: OutlineInputBorder(),
                    ),
                    keyboardType:
                        const TextInputType.numberWithOptions(decimal: true),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 16),
            TextFormField(
              controller: _arrowFocController,
              decoration: const InputDecoration(
                labelText: 'Arrow FOC (%)',
                border: OutlineInputBorder(),
              ),
              keyboardType:
                  const TextInputType.numberWithOptions(decimal: true),
            ),
          ],
        ),
      ),
    );
  }
}
