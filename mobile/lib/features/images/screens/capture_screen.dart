import 'dart:io';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:image_picker/image_picker.dart';
import 'package:path_provider/path_provider.dart';
import 'package:uuid/uuid.dart';
import 'package:path/path.dart' as p;

import '../../../core/database/database.dart';
import '../../../core/sync/sync_service.dart';

class CaptureScreen extends ConsumerStatefulWidget {
  final String endId;

  const CaptureScreen({super.key, required this.endId});

  @override
  ConsumerState<CaptureScreen> createState() => _CaptureScreenState();
}

class _CaptureScreenState extends ConsumerState<CaptureScreen> {
  File? _imageFile;
  bool _saving = false;

  Future<void> _takePhoto() async {
    final picker = ImagePicker();
    final photo = await picker.pickImage(
      source: ImageSource.camera,
      maxWidth: 1920,
      maxHeight: 1920,
      imageQuality: 85,
    );

    if (photo == null) return;

    // Copy to app documents for persistence
    final appDir = await getApplicationDocumentsDirectory();
    final imagesDir = Directory(p.join(appDir.path, 'end_images'));
    if (!await imagesDir.exists()) {
      await imagesDir.create(recursive: true);
    }

    final fileName = '${widget.endId}_${DateTime.now().millisecondsSinceEpoch}.jpg';
    final savedFile = await File(photo.path).copy(
      p.join(imagesDir.path, fileName),
    );

    setState(() {
      _imageFile = savedFile;
    });
  }

  Future<void> _saveImage() async {
    if (_imageFile == null) return;
    setState(() => _saving = true);

    final db = ref.read(databaseProvider);
    final syncService = ref.read(syncServiceProvider);
    final id = const Uuid().v4();

    await db.into(db.endImages).insert(EndImagesCompanion.insert(
      id: id,
      endId: widget.endId,
      filePath: _imageFile!.path,
      capturedAt: DateTime.now(),
    ));

    await syncService.enqueue(
      entityType: 'image',
      entityId: id,
      action: 'upload',
      payload: {
        'end_id': widget.endId,
        'file_path': _imageFile!.path,
      },
    );

    if (mounted) {
      Navigator.of(context).pop();
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Target Photo'),
      ),
      body: Column(
        children: [
          Expanded(
            child: _imageFile != null
                ? Image.file(_imageFile!, fit: BoxFit.contain)
                : Center(
                    child: Column(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Icon(
                          Icons.camera_alt_outlined,
                          size: 64,
                          color: theme.colorScheme.outline,
                        ),
                        const SizedBox(height: 16),
                        Text(
                          'Take a photo of your target',
                          style: theme.textTheme.bodyLarge,
                        ),
                      ],
                    ),
                  ),
          ),
          Padding(
            padding: const EdgeInsets.all(16),
            child: Row(
              children: [
                Expanded(
                  child: OutlinedButton.icon(
                    onPressed: _saving ? null : _takePhoto,
                    icon: const Icon(Icons.camera_alt),
                    label: Text(_imageFile != null ? 'Retake' : 'Take Photo'),
                  ),
                ),
                if (_imageFile != null) ...[
                  const SizedBox(width: 16),
                  Expanded(
                    child: FilledButton.icon(
                      onPressed: _saving ? null : _saveImage,
                      icon: _saving
                          ? const SizedBox(
                              width: 18,
                              height: 18,
                              child:
                                  CircularProgressIndicator(strokeWidth: 2),
                            )
                          : const Icon(Icons.save),
                      label: const Text('Save'),
                    ),
                  ),
                ],
              ],
            ),
          ),
        ],
      ),
    );
  }
}
