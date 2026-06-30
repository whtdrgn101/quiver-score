import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../scoring/providers/scoring_provider.dart';
import '../../scoring/screens/scoring_screen.dart';
import '../../clubs/models/tournament.dart';
import '../models/challenge.dart';
import '../providers/challenge_provider.dart';
import 'send_challenge_screen.dart';
import 'challenge_comparison_screen.dart';

class ChallengeListScreen extends ConsumerStatefulWidget {
  const ChallengeListScreen({super.key});

  @override
  ConsumerState<ChallengeListScreen> createState() => _ChallengeListScreenState();
}

class _ChallengeListScreenState extends ConsumerState<ChallengeListScreen>
    with SingleTickerProviderStateMixin {
  late TabController _tabController;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 2, vsync: this);
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final challengesAsync = ref.watch(challengesProvider);
    final userIdAsync = ref.watch(userIdProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Challenges'),
        centerTitle: true,
        bottom: TabBar(
          controller: _tabController,
          tabs: const [
            Tab(text: 'Active'),
            Tab(text: 'Past'),
          ],
        ),
      ),
      body: userIdAsync.when(
        data: (currentUserId) {
          if (currentUserId == null) {
            return const Center(child: Text('Authentication error. Please sign in again.'));
          }

          return challengesAsync.when(
            data: (challenges) {
              final active = challenges.where((c) =>
                  c.status == 'pending' || c.status == 'accepted').toList();
              final past = challenges.where((c) =>
                  c.status == 'completed' || c.status == 'declined').toList();

              return TabBarView(
                controller: _tabController,
                children: [
                  _buildChallengeList(active, currentUserId, isActive: true),
                  _buildChallengeList(past, currentUserId, isActive: false),
                ],
              );
            },
            loading: () => const Center(child: CircularProgressIndicator()),
            error: (err, _) => Center(child: Text('Error loading challenges: $err')),
          );
        },
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (err, _) => Center(child: Text('Error: $err')),
      ),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () {
          Navigator.of(context).push(
            MaterialPageRoute(builder: (_) => const SendChallengeScreen()),
          );
        },
        icon: const Icon(Icons.add),
        label: const Text('New Challenge'),
      ),
    );
  }

  Widget _buildChallengeList(List<Challenge> list, String currentUserId,
      {required bool isActive}) {
    if (list.isEmpty) {
      return Center(
        child: Text(
          isActive ? 'No active challenges' : 'No past challenges',
          style: const TextStyle(color: Colors.grey, fontSize: 16),
        ),
      );
    }

    return RefreshIndicator(
      onRefresh: () async {
        ref.invalidate(challengesProvider);
      },
      child: ListView.builder(
        padding: const EdgeInsets.all(12),
        itemCount: list.length,
        itemBuilder: (ctx, index) {
          final challenge = list[index];
          final isChallenger = challenge.challengerId == currentUserId;
          final opponentUsername =
              isChallenger ? challenge.challengeeUsername : challenge.challengerUsername;

          return Card(
            margin: const EdgeInsets.only(bottom: 12),
            shape:
                RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
            child: Padding(
              padding: const EdgeInsets.all(16.0),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(
                        challenge.templateName,
                        style: const TextStyle(
                            fontSize: 16, fontWeight: FontWeight.bold),
                      ),
                      _buildStatusBadge(challenge.status),
                    ],
                  ),
                  const SizedBox(height: 8),
                  Text(
                    isChallenger
                        ? 'Challenged $opponentUsername'
                        : 'Challenged by $opponentUsername',
                    style: const TextStyle(fontSize: 14, color: Colors.black87),
                  ),
                  const SizedBox(height: 12),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      // Scores overview
                      Text(
                        'Scores: ${challenge.challengerUsername} (${challenge.challengerScore ?? '-'}) vs '
                        '${challenge.challengeeUsername} (${challenge.challengeeScore ?? '-'})',
                        style: const TextStyle(fontSize: 13, color: Colors.grey),
                      ),
                    ],
                  ),
                  if (isActive) ...[
                    const SizedBox(height: 12),
                    Row(
                      mainAxisAlignment: MainAxisAlignment.end,
                      children: [
                        // Pending actions for challengee
                        if (challenge.status == 'pending' && !isChallenger) ...[
                          TextButton(
                            onPressed: () => _handleDecline(challenge.id),
                            child: const Text('Decline',
                                style: TextStyle(color: Colors.red)),
                          ),
                          const SizedBox(width: 8),
                          ElevatedButton(
                            onPressed: () => _handleAccept(challenge.id),
                            child: const Text('Accept'),
                          ),
                        ],
                        // Pending message for challenger
                        if (challenge.status == 'pending' && isChallenger)
                          const Text(
                            'Awaiting acceptance...',
                            style: TextStyle(
                                fontStyle: FontStyle.italic, color: Colors.grey),
                          ),
                        // Accepted and ready to shoot
                        if (challenge.status == 'accepted') ...[
                          // Check if user has already submitted a score
                          if ((isChallenger && challenge.challengerSessionId == null) ||
                              (!isChallenger && challenge.challengeeSessionId == null)) ...[
                            ElevatedButton.icon(
                              onPressed: () => _startChallengeSession(challenge),
                              icon: const Icon(Icons.play_arrow, size: 18),
                              label: const Text('Shoot Now'),
                            ),
                            const SizedBox(width: 8),
                          ],
                          ElevatedButton(
                            onPressed: () => _viewScoreboard(challenge),
                            child: const Text('View Scoreboard'),
                          ),
                        ],
                      ],
                    ),
                  ] else ...[
                    const SizedBox(height: 12),
                    Row(
                      mainAxisAlignment: MainAxisAlignment.end,
                      children: [
                        ElevatedButton(
                          onPressed: () => _viewScoreboard(challenge),
                          child: const Text('View Scoreboard'),
                        ),
                      ],
                    ),
                  ],
                ],
              ),
            ),
          );
        },
      ),
    );
  }

  Widget _buildStatusBadge(String status) {
    Color color = Colors.grey;
    if (status == 'accepted') color = Colors.blue;
    if (status == 'completed') color = Colors.green;
    if (status == 'declined') color = Colors.red;

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.1),
        borderRadius: BorderRadius.circular(4),
        border: Border.all(color: color),
      ),
      child: Text(
        status.toUpperCase(),
        style: TextStyle(
            fontSize: 10, fontWeight: FontWeight.bold, color: color),
      ),
    );
  }

  Future<void> _handleAccept(String challengeId) async {
    try {
      await acceptChallenge(ref, challengeId);
      ref.invalidate(challengesProvider);
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Failed to accept challenge: $e')),
        );
      }
    }
  }

  Future<void> _handleDecline(String challengeId) async {
    try {
      await declineChallenge(ref, challengeId);
      ref.invalidate(challengesProvider);
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Failed to decline challenge: $e')),
        );
      }
    }
  }

  Future<void> _startChallengeSession(Challenge challenge) async {
    // 1. Initialize and start scoring session locally
    await ref.read(scoringProvider.notifier).startSession(
          templateId: challenge.templateId,
        );

    // 2. Open scoring screen passing the challenge routing context
    if (mounted) {
      Navigator.of(context).push(
        MaterialPageRoute(
          builder: (_) => ScoringScreen(
            tournamentContext: TournamentContext(
              challengeId: challenge.id,
              tournamentName: 'Challenge: ${challenge.templateName}',
            ),
          ),
        ),
      );
    }
  }

  void _viewScoreboard(Challenge challenge) {
    Navigator.of(context).push(
      MaterialPageRoute(
        builder: (_) => ChallengeComparisonScreen(challenge: challenge),
      ),
    );
  }
}
