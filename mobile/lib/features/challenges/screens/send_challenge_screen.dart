import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../rounds/providers/rounds_provider.dart';
import '../providers/challenge_provider.dart';

class SendChallengeScreen extends ConsumerStatefulWidget {
  const SendChallengeScreen({super.key});

  @override
  ConsumerState<SendChallengeScreen> createState() => _SendChallengeScreenState();
}

class _SendChallengeScreenState extends ConsumerState<SendChallengeScreen> {
  String? _selectedFriendId;
  String? _selectedTemplateId;
  int _expiresInHours = 24;
  bool _submitting = false;

  @override
  Widget build(BuildContext context) {
    final followingAsync = ref.watch(followingUsersProvider);
    final templatesAsync = ref.watch(roundTemplatesProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Challenge Friend'),
        centerTitle: true,
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(24.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            const Text(
              'Select a friend to challenge:',
              style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 12),
            followingAsync.when(
              data: (users) {
                if (users.isEmpty) {
                  return const Card(
                    color: Colors.amberAccent,
                    child: Padding(
                      padding: EdgeInsets.all(16.0),
                      child: Text(
                        'You are not following anyone yet! Follow someone first to challenge them.',
                        style: TextStyle(color: Colors.black87),
                      ),
                    ),
                  );
                }
                return DropdownButtonFormField<String>(
                  initialValue: _selectedFriendId,
                  hint: const Text('Choose a friend'),
                  decoration: const InputDecoration(
                    border: OutlineInputBorder(),
                    filled: true,
                  ),
                  items: users.map((u) {
                    return DropdownMenuItem<String>(
                      value: u.id,
                      child: Text(u.username),
                    );
                  }).toList(),
                  onChanged: (val) {
                    setState(() => _selectedFriendId = val);
                  },
                );
              },
              loading: () => const Center(child: CircularProgressIndicator()),
              error: (err, _) => Text('Error loading following list: $err'),
            ),
            const SizedBox(height: 24),
            const Text(
              'Select a Round Template:',
              style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 12),
            templatesAsync.when(
              data: (templates) {
                if (templates.isEmpty) {
                  return const Text('No round templates available.');
                }
                return DropdownButtonFormField<String>(
                  initialValue: _selectedTemplateId,
                  hint: const Text('Choose a round template'),
                  decoration: const InputDecoration(
                    border: OutlineInputBorder(),
                    filled: true,
                  ),
                  items: templates.map((t) {
                    return DropdownMenuItem<String>(
                      value: t.id,
                      child: Text(t.name),
                    );
                  }).toList(),
                  onChanged: (val) {
                    setState(() => _selectedTemplateId = val);
                  },
                );
              },
              loading: () => const Center(child: CircularProgressIndicator()),
              error: (err, _) => Text('Error loading templates: $err'),
            ),
            const SizedBox(height: 24),
            const Text(
              'Expiration:',
              style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 12),
            DropdownButtonFormField<int>(
              initialValue: _expiresInHours,
              decoration: const InputDecoration(
                border: OutlineInputBorder(),
                filled: true,
              ),
              items: const [
                DropdownMenuItem(value: 12, child: Text('12 Hours')),
                DropdownMenuItem(value: 24, child: Text('24 Hours')),
                DropdownMenuItem(value: 48, child: Text('48 Hours')),
                DropdownMenuItem(value: 72, child: Text('72 Hours')),
              ],
              onChanged: (val) {
                if (val != null) {
                  setState(() => _expiresInHours = val);
                }
              },
            ),
            const SizedBox(height: 48),
            FilledButton(
              style: FilledButton.styleFrom(
                padding: const EdgeInsets.symmetric(vertical: 16),
              ),
              onPressed: (_selectedFriendId == null ||
                      _selectedTemplateId == null ||
                      _submitting)
                  ? null
                  : _sendChallenge,
              child: _submitting
                  ? const SizedBox(
                      height: 20,
                      width: 20,
                      child: CircularProgressIndicator(color: Colors.white),
                    )
                  : const Text('Send Challenge', style: TextStyle(fontSize: 16)),
            ),
          ],
        ),
      ),
    );
  }

  Future<void> _sendChallenge() async {
    setState(() => _submitting = true);
    try {
      await createChallenge(
        ref,
        challengeeId: _selectedFriendId!,
        templateId: _selectedTemplateId!,
        expiresInHours: _expiresInHours,
      );
      // Refresh the challenges list
      ref.invalidate(challengesProvider);

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Challenge sent successfully!'),
            backgroundColor: Colors.green,
          ),
        );
        Navigator.of(context).pop();
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Failed to send challenge: $e'),
            backgroundColor: Colors.red,
          ),
        );
      }
    } finally {
      if (mounted) {
        setState(() => _submitting = false);
      }
    }
  }
}
