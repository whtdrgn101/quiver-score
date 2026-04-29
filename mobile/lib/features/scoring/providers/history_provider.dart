import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/api_client.dart';

/// Fetches completed scoring sessions from the server API
final historyProvider =
    AsyncNotifierProvider<HistoryNotifier, List<SessionSummary>>(
        HistoryNotifier.new);

class HistoryNotifier extends AsyncNotifier<List<SessionSummary>> {
  @override
  Future<List<SessionSummary>> build() => _fetch();

  Future<List<SessionSummary>> _fetch() async {
    final api = ref.read(apiClientProvider);
    final response = await api.dio.get('/api/v1/sessions');
    final list = response.data as List;
    return list.map((j) => SessionSummary.fromJson(j as Map<String, dynamic>)).toList();
  }

  Future<void> refresh() async {
    state = const AsyncLoading();
    state = await AsyncValue.guard(_fetch);
  }
}

class SessionSummary {
  final String id;
  final String? templateName;
  final String status;
  final int totalScore;
  final int totalXCount;
  final int totalArrows;
  final DateTime startedAt;
  final DateTime? completedAt;

  const SessionSummary({
    required this.id,
    this.templateName,
    required this.status,
    required this.totalScore,
    required this.totalXCount,
    required this.totalArrows,
    required this.startedAt,
    this.completedAt,
  });

  factory SessionSummary.fromJson(Map<String, dynamic> json) {
    return SessionSummary(
      id: json['id'] as String,
      templateName: json['template_name'] as String?,
      status: json['status'] as String,
      totalScore: json['total_score'] as int? ?? 0,
      totalXCount: json['total_x_count'] as int? ?? 0,
      totalArrows: json['total_arrows'] as int? ?? 0,
      startedAt: DateTime.parse(json['started_at'] as String),
      completedAt: json['completed_at'] != null
          ? DateTime.parse(json['completed_at'] as String)
          : null,
    );
  }
}
