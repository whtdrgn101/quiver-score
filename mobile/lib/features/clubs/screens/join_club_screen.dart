import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/api_client.dart';
import '../models/club.dart';
import '../providers/clubs_provider.dart';

class JoinClubScreen extends ConsumerStatefulWidget {
  const JoinClubScreen({super.key});

  @override
  ConsumerState<JoinClubScreen> createState() => _JoinClubScreenState();
}

class _JoinClubScreenState extends ConsumerState<JoinClubScreen> {
  final _codeController = TextEditingController();
  Club? _preview;
  bool _loading = false;
  bool _joining = false;
  String? _error;

  @override
  void dispose() {
    _codeController.dispose();
    super.dispose();
  }

  String _extractCode(String input) {
    final uri = Uri.tryParse(input.trim());
    if (uri != null && uri.pathSegments.length >= 2) {
      final joinIdx = uri.pathSegments.indexOf('join');
      if (joinIdx >= 0 && joinIdx + 1 < uri.pathSegments.length) {
        return uri.pathSegments[joinIdx + 1];
      }
    }
    return input.trim();
  }

  Future<void> _lookupCode() async {
    final code = _extractCode(_codeController.text);
    if (code.isEmpty) return;

    setState(() {
      _loading = true;
      _error = null;
      _preview = null;
    });

    try {
      final api = ref.read(apiClientProvider);
      final response = await api.dio.get('/api/v1/clubs/join/$code');
      setState(() {
        _preview = Club.fromJson(response.data as Map<String, dynamic>);
      });
    } catch (e) {
      setState(() => _error = 'Invalid or expired invite code');
    } finally {
      setState(() => _loading = false);
    }
  }

  Future<void> _join() async {
    final code = _extractCode(_codeController.text);
    setState(() => _joining = true);

    try {
      await ref.read(clubsProvider.notifier).joinClub(code);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
              content: Text('Joined ${_preview?.name ?? 'club'}')),
        );
        Navigator.pop(context);
      }
    } catch (e) {
      if (mounted) {
        setState(() => _error = 'Failed to join: $e');
      }
    } finally {
      if (mounted) setState(() => _joining = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Scaffold(
      appBar: AppBar(title: const Text('Join a Club')),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          Text(
            'Enter an invite code or paste an invite link to join a club.',
            style: theme.textTheme.bodyMedium?.copyWith(
              color: theme.colorScheme.onSurfaceVariant,
            ),
          ),
          const SizedBox(height: 16),
          TextField(
            controller: _codeController,
            decoration: InputDecoration(
              labelText: 'Invite code or link',
              border: const OutlineInputBorder(),
              suffixIcon: IconButton(
                icon: const Icon(Icons.search),
                onPressed: _loading ? null : _lookupCode,
              ),
            ),
            onSubmitted: (_) => _lookupCode(),
          ),
          if (_error != null) ...[
            const SizedBox(height: 12),
            Text(_error!, style: TextStyle(color: theme.colorScheme.error)),
          ],
          if (_loading) ...[
            const SizedBox(height: 24),
            const Center(child: CircularProgressIndicator()),
          ],
          if (_preview != null) ...[
            const SizedBox(height: 24),
            Card(
              child: Padding(
                padding: const EdgeInsets.all(16),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(_preview!.name,
                        style: theme.textTheme.titleLarge),
                    if (_preview!.description != null) ...[
                      const SizedBox(height: 8),
                      Text(_preview!.description!,
                          style: theme.textTheme.bodyMedium),
                    ],
                    const SizedBox(height: 8),
                    Text(
                      '${_preview!.memberCount} member${_preview!.memberCount == 1 ? '' : 's'}',
                      style: theme.textTheme.bodySmall?.copyWith(
                        color: theme.colorScheme.onSurfaceVariant,
                      ),
                    ),
                    const SizedBox(height: 16),
                    SizedBox(
                      width: double.infinity,
                      child: FilledButton(
                        onPressed: _joining ? null : _join,
                        child: _joining
                            ? const SizedBox(
                                width: 20,
                                height: 20,
                                child: CircularProgressIndicator(
                                    strokeWidth: 2),
                              )
                            : const Text('Join Club'),
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ],
        ],
      ),
    );
  }
}
