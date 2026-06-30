import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../models/challenge.dart';
import '../providers/challenge_provider.dart';

class ChallengeComparisonScreen extends ConsumerWidget {
  final Challenge challenge;

  const ChallengeComparisonScreen({super.key, required this.challenge});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final comparisonAsync = ref.watch(challengeComparisonProvider(challenge));

    return Scaffold(
      appBar: AppBar(
        title: const Text('Challenge Scoreboard'),
        centerTitle: true,
      ),
      body: comparisonAsync.when(
        data: (data) {
          final challengerSession = data.challenger;
          final challengeeSession = data.challengee;

          // Determine winner text
          String winnerText = '';
          Color winnerColor = Colors.grey;
          if (challenge.status == 'completed' &&
              challenge.challengerScore != null &&
              challenge.challengeeScore != null) {
            final diff = challenge.challengerScore! - challenge.challengeeScore!;
            if (diff > 0) {
              winnerText = '${challenge.challengerUsername} wins by ${diff.abs()} points!';
              winnerColor = Colors.green;
            } else if (diff < 0) {
              winnerText = '${challenge.challengeeUsername} wins by ${diff.abs()} points!';
              winnerColor = Colors.green;
            } else {
              winnerText = "It's a tie!";
              winnerColor = Colors.blue;
            }
          } else {
            winnerText = 'Challenge Status: ${challenge.status.toUpperCase()}';
          }

          // Build list of ends side by side
          final challengerEnds = challengerSession?['ends'] as List? ?? [];
          final challengeeEnds = challengeeSession?['ends'] as List? ?? [];

          final maxEnds = challengerEnds.length > challengeeEnds.length
              ? challengerEnds.length
              : challengeeEnds.length;

          return SingleChildScrollView(
            padding: const EdgeInsets.all(16.0),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                // Header card
                Card(
                  elevation: 2,
                  shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(12)),
                  child: Padding(
                    padding: const EdgeInsets.all(16.0),
                    child: Column(
                      children: [
                        Text(
                          challenge.templateName,
                          style: const TextStyle(
                            fontSize: 18,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                        const SizedBox(height: 16),
                        Row(
                          children: [
                            // Challenger
                            Expanded(
                              child: Column(
                                children: [
                                  CircleAvatar(
                                    radius: 24,
                                    backgroundColor: Colors.blue.shade100,
                                    child: Text(
                                      challenge.challengerUsername[0].toUpperCase(),
                                      style: TextStyle(
                                          fontWeight: FontWeight.bold,
                                          color: Colors.blue.shade800),
                                    ),
                                  ),
                                  const SizedBox(height: 8),
                                  Text(
                                    challenge.challengerUsername,
                                    style: const TextStyle(
                                        fontWeight: FontWeight.w600),
                                    textAlign: TextAlign.center,
                                  ),
                                  const SizedBox(height: 4),
                                  Text(
                                    challenge.challengerScore != null
                                        ? '${challenge.challengerScore} pts'
                                        : 'Shooting...',
                                    style: TextStyle(
                                      fontSize: 16,
                                      fontWeight: FontWeight.bold,
                                      color: challenge.challengerScore != null
                                          ? Colors.blue
                                          : Colors.grey,
                                    ),
                                  ),
                                ],
                              ),
                            ),
                            const Padding(
                              padding: EdgeInsets.symmetric(horizontal: 16),
                              child: Text(
                                'VS',
                                style: TextStyle(
                                    fontSize: 18,
                                    fontWeight: FontWeight.bold,
                                    color: Colors.grey),
                              ),
                            ),
                            // Challengee
                            Expanded(
                              child: Column(
                                children: [
                                  CircleAvatar(
                                    radius: 24,
                                    backgroundColor: Colors.purple.shade100,
                                    child: Text(
                                      challenge.challengeeUsername[0].toUpperCase(),
                                      style: TextStyle(
                                          fontWeight: FontWeight.bold,
                                          color: Colors.purple.shade800),
                                    ),
                                  ),
                                  const SizedBox(height: 8),
                                  Text(
                                    challenge.challengeeUsername,
                                    style: const TextStyle(
                                        fontWeight: FontWeight.w600),
                                    textAlign: TextAlign.center,
                                  ),
                                  const SizedBox(height: 4),
                                  Text(
                                    challenge.challengeeScore != null
                                        ? '${challenge.challengeeScore} pts'
                                        : 'Shooting...',
                                    style: TextStyle(
                                      fontSize: 16,
                                      fontWeight: FontWeight.bold,
                                      color: challenge.challengeeScore != null
                                          ? Colors.purple
                                          : Colors.grey,
                                    ),
                                  ),
                                ],
                              ),
                            ),
                          ],
                        ),
                      ],
                    ),
                  ),
                ),
                const SizedBox(height: 16),

                // Status Banner
                Container(
                  padding: const EdgeInsets.symmetric(vertical: 12),
                  decoration: BoxDecoration(
                    color: winnerColor.withOpacity(0.1),
                    borderRadius: BorderRadius.circular(8),
                    border: Border.all(color: winnerColor),
                  ),
                  child: Center(
                    child: Text(
                      winnerText,
                      style: TextStyle(
                        fontSize: 16,
                        fontWeight: FontWeight.bold,
                        color: winnerColor,
                      ),
                    ),
                  ),
                ),
                const SizedBox(height: 24),

                // Ends List Title
                const Text(
                  'End-by-End Scoreboard',
                  style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
                ),
                const SizedBox(height: 12),

                if (maxEnds == 0)
                  const Center(
                    child: Padding(
                      padding: EdgeInsets.all(32.0),
                      child: Text(
                        'No end scores submitted yet.',
                        style: TextStyle(color: Colors.grey),
                      ),
                    ),
                  )
                else
                  ListView.separated(
                    shrinkWrap: true,
                    physics: const NeverScrollableScrollPhysics(),
                    itemCount: maxEnds,
                    separatorBuilder: (_, __) => const Divider(),
                    itemBuilder: (ctx, index) {
                      final challengerEnd = index < challengerEnds.length
                          ? challengerEnds[index]
                          : null;
                      final challengeeEnd = index < challengeeEnds.length
                          ? challengeeEnds[index]
                          : null;

                      final cScore = challengerEnd != null
                          ? challengerEnd['total_score']?.toString() ?? '-'
                          : '-';
                      final eScore = challengeeEnd != null
                          ? challengeeEnd['total_score']?.toString() ?? '-'
                          : '-';

                      return Padding(
                        padding: const EdgeInsets.symmetric(vertical: 8.0),
                        child: Row(
                          children: [
                            Expanded(
                              child: Text(
                                cScore,
                                style: const TextStyle(
                                  fontSize: 16,
                                  fontWeight: FontWeight.w500,
                                ),
                                textAlign: TextAlign.center,
                              ),
                            ),
                            Container(
                              padding: const EdgeInsets.symmetric(
                                  horizontal: 12, vertical: 4),
                              decoration: BoxDecoration(
                                color: Colors.grey.shade200,
                                borderRadius: BorderRadius.circular(4),
                              ),
                              child: Text(
                                'End ${index + 1}',
                                style: TextStyle(
                                    fontSize: 12,
                                    fontWeight: FontWeight.bold,
                                    color: Colors.grey.shade700),
                              ),
                            ),
                            Expanded(
                              child: Text(
                                eScore,
                                style: const TextStyle(
                                  fontSize: 16,
                                  fontWeight: FontWeight.w500,
                                ),
                                textAlign: TextAlign.center,
                              ),
                            ),
                          ],
                        ),
                      );
                    },
                  ),
              ],
            ),
          );
        },
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (err, _) => Center(child: Text('Error loading details: $err')),
      ),
    );
  }
}
