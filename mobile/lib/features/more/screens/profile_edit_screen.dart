import 'dart:io';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:image_picker/image_picker.dart';

import '../../../core/api/api_client.dart';
import '../providers/user_provider.dart';

const _socialPlatforms = [
  ('instagram', 'Instagram', Icons.camera_alt_outlined),
  ('twitter', 'X / Twitter', Icons.alternate_email),
  ('facebook', 'Facebook', Icons.facebook),
  ('youtube', 'YouTube', Icons.play_circle_outline),
  ('tiktok', 'TikTok', Icons.music_note_outlined),
  ('website', 'Website', Icons.language),
];

class ProfileEditScreen extends ConsumerStatefulWidget {
  final UserInfo user;

  const ProfileEditScreen({super.key, required this.user});

  @override
  ConsumerState<ProfileEditScreen> createState() => _ProfileEditScreenState();
}

class _ProfileEditScreenState extends ConsumerState<ProfileEditScreen> {
  late final TextEditingController _displayNameCtl;
  late final TextEditingController _bioCtl;
  late final TextEditingController _bowTypeCtl;
  late final TextEditingController _classificationCtl;
  late final Map<String, TextEditingController> _socialCtls;
  late bool _profilePublic;
  bool _saving = false;

  @override
  void initState() {
    super.initState();
    final u = widget.user;
    _displayNameCtl = TextEditingController(text: u.displayName ?? '');
    _bioCtl = TextEditingController(text: u.bio ?? '');
    _bowTypeCtl = TextEditingController(text: u.bowType ?? '');
    _classificationCtl = TextEditingController(text: u.classification ?? '');
    _profilePublic = u.profilePublic;

    _socialCtls = {};
    for (final (key, _, _) in _socialPlatforms) {
      _socialCtls[key] = TextEditingController(
        text: (u.socialLinks?[key] as String?) ?? '',
      );
    }
  }

  @override
  void dispose() {
    _displayNameCtl.dispose();
    _bioCtl.dispose();
    _bowTypeCtl.dispose();
    _classificationCtl.dispose();
    for (final ctl in _socialCtls.values) {
      ctl.dispose();
    }
    super.dispose();
  }

  Future<void> _save() async {
    setState(() => _saving = true);
    try {
      final api = ref.read(apiClientProvider);

      final socialLinks = <String, String>{};
      for (final (key, _, _) in _socialPlatforms) {
        final val = _socialCtls[key]!.text.trim();
        if (val.isNotEmpty) socialLinks[key] = val;
      }

      await api.dio.put('/api/v1/users/me', data: {
        'display_name': _displayNameCtl.text.trim().isEmpty
            ? null
            : _displayNameCtl.text.trim(),
        'bio':
            _bioCtl.text.trim().isEmpty ? null : _bioCtl.text.trim(),
        'bow_type': _bowTypeCtl.text.trim().isEmpty
            ? null
            : _bowTypeCtl.text.trim(),
        'classification': _classificationCtl.text.trim().isEmpty
            ? null
            : _classificationCtl.text.trim(),
        'profile_public': _profilePublic,
        'social_links': socialLinks.isEmpty ? null : socialLinks,
      });

      ref.invalidate(currentUserProvider);

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Profile updated')),
        );
        Navigator.pop(context);
      }
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

  Future<void> _pickAvatar() async {
    final picker = ImagePicker();
    final picked = await picker.pickImage(
      source: ImageSource.gallery,
      maxWidth: 512,
      maxHeight: 512,
      imageQuality: 80,
    );
    if (picked == null) return;

    setState(() => _saving = true);
    try {
      final api = ref.read(apiClientProvider);
      final file = File(picked.path);
      final formData = FormData.fromMap({
        'file': await MultipartFile.fromFile(
          file.path,
          filename: 'avatar.jpg',
        ),
      });
      await api.dio.post('/api/v1/users/me/avatar', data: formData);
      ref.invalidate(currentUserProvider);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Avatar updated')),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Failed to upload avatar: $e')),
        );
      }
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final userAsync = ref.watch(currentUserProvider);
    final avatar = userAsync.value?.avatar ?? widget.user.avatar;

    return Scaffold(
      appBar: AppBar(
        title: const Text('Edit Profile'),
        actions: [
          TextButton(
            onPressed: _saving ? null : _save,
            child: _saving
                ? const SizedBox(
                    width: 16,
                    height: 16,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : const Text('Save'),
          ),
        ],
      ),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          // Avatar
          Center(
            child: GestureDetector(
              onTap: _pickAvatar,
              child: Stack(
                children: [
                  CircleAvatar(
                    radius: 48,
                    backgroundColor: theme.colorScheme.primaryContainer,
                    backgroundImage:
                        avatar != null ? NetworkImage(avatar) : null,
                    child: avatar == null
                        ? Text(
                            (widget.user.displayName ?? widget.user.username)
                                .substring(0, 1)
                                .toUpperCase(),
                            style: theme.textTheme.headlineMedium?.copyWith(
                              color: theme.colorScheme.onPrimaryContainer,
                            ),
                          )
                        : null,
                  ),
                  Positioned(
                    bottom: 0,
                    right: 0,
                    child: Container(
                      padding: const EdgeInsets.all(4),
                      decoration: BoxDecoration(
                        color: theme.colorScheme.primary,
                        shape: BoxShape.circle,
                      ),
                      child: Icon(
                        Icons.camera_alt,
                        size: 18,
                        color: theme.colorScheme.onPrimary,
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ),

          const SizedBox(height: 24),

          // Display Name
          TextField(
            controller: _displayNameCtl,
            decoration: const InputDecoration(
              labelText: 'Display Name',
              border: OutlineInputBorder(),
            ),
            textCapitalization: TextCapitalization.words,
          ),

          const SizedBox(height: 16),

          // Bio
          TextField(
            controller: _bioCtl,
            decoration: const InputDecoration(
              labelText: 'Bio',
              border: OutlineInputBorder(),
            ),
            maxLines: 3,
            textCapitalization: TextCapitalization.sentences,
          ),

          const SizedBox(height: 16),

          // Bow Type
          TextField(
            controller: _bowTypeCtl,
            decoration: const InputDecoration(
              labelText: 'Bow Type',
              hintText: 'e.g. Recurve, Compound, Barebow',
              border: OutlineInputBorder(),
            ),
          ),

          const SizedBox(height: 16),

          // Classification
          TextField(
            controller: _classificationCtl,
            decoration: const InputDecoration(
              labelText: 'Classification',
              hintText: 'e.g. Senior, Junior, Master',
              border: OutlineInputBorder(),
            ),
          ),

          const SizedBox(height: 16),

          // Public profile toggle
          Card(
            child: SwitchListTile(
              title: const Text('Public Profile'),
              subtitle: const Text('Allow others to see your profile and stats'),
              value: _profilePublic,
              onChanged: (v) => setState(() => _profilePublic = v),
            ),
          ),

          const SizedBox(height: 24),

          // Social links
          Text(
            'Social Links',
            style: theme.textTheme.titleSmall?.copyWith(
              color: theme.colorScheme.onSurfaceVariant,
            ),
          ),
          const SizedBox(height: 8),

          for (final (key, label, icon) in _socialPlatforms) ...[
            TextField(
              controller: _socialCtls[key],
              decoration: InputDecoration(
                labelText: label,
                prefixIcon: Icon(icon),
                border: const OutlineInputBorder(),
                hintText: key == 'website'
                    ? 'https://yoursite.com'
                    : 'https://$key.com/username',
              ),
              keyboardType: TextInputType.url,
            ),
            const SizedBox(height: 12),
          ],
        ],
      ),
    );
  }
}
