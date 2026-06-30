class Challenge {
  final String id;
  final String challengerId;
  final String challengerUsername;
  final String challengeeId;
  final String challengeeUsername;
  final String templateId;
  final String templateName;
  final String? challengerSessionId;
  final int? challengerScore;
  final String? challengeeSessionId;
  final int? challengeeScore;
  final String status;
  final DateTime createdAt;
  final DateTime? expiresAt;

  const Challenge({
    required this.id,
    required this.challengerId,
    required this.challengerUsername,
    required this.challengeeId,
    required this.challengeeUsername,
    required this.templateId,
    required this.templateName,
    this.challengerSessionId,
    this.challengerScore,
    this.challengeeSessionId,
    this.challengeeScore,
    required this.status,
    required this.createdAt,
    this.expiresAt,
  });

  factory Challenge.fromJson(Map<String, dynamic> json) {
    return Challenge(
      id: json['id'] as String,
      challengerId: json['challenger_id'] as String,
      challengerUsername: json['challenger_username'] as String,
      challengeeId: json['challengee_id'] as String,
      challengeeUsername: json['challengee_username'] as String,
      templateId: json['template_id'] as String,
      templateName: json['template_name'] as String,
      challengerSessionId: json['challenger_session_id'] as String?,
      challengerScore: json['challenger_score'] as int?,
      challengeeSessionId: json['challengee_session_id'] as String?,
      challengeeScore: json['challengee_score'] as int?,
      status: json['status'] as String,
      createdAt: DateTime.parse(json['created_at'] as String),
      expiresAt: json['expires_at'] != null
          ? DateTime.parse(json['expires_at'] as String)
          : null,
    );
  }
}
