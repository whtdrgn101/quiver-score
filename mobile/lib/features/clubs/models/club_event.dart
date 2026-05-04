class ClubEvent {
  final String id;
  final String clubId;
  final String name;
  final String? description;
  final String templateId;
  final String? templateName;
  final DateTime eventDate;
  final String? location;
  final String createdBy;
  final List<EventParticipant> participants;
  final DateTime createdAt;

  const ClubEvent({
    required this.id,
    required this.clubId,
    required this.name,
    this.description,
    required this.templateId,
    this.templateName,
    required this.eventDate,
    this.location,
    required this.createdBy,
    this.participants = const [],
    required this.createdAt,
  });

  factory ClubEvent.fromJson(Map<String, dynamic> json) {
    return ClubEvent(
      id: json['id'] as String,
      clubId: json['club_id'] as String,
      name: json['name'] as String,
      description: json['description'] as String?,
      templateId: json['template_id'] as String,
      templateName: json['template_name'] as String?,
      eventDate: DateTime.parse(json['event_date'] as String),
      location: json['location'] as String?,
      createdBy: json['created_by'] as String,
      participants: (json['participants'] as List?)
              ?.map(
                  (p) => EventParticipant.fromJson(p as Map<String, dynamic>))
              .toList() ??
          [],
      createdAt: DateTime.parse(json['created_at'] as String),
    );
  }

  bool get isPast => eventDate.isBefore(DateTime.now());
}

class EventParticipant {
  final String userId;
  final String username;
  final String? displayName;
  final String? avatar;
  final String status;
  final int? score;
  final int? xCount;
  final String? sessionId;

  const EventParticipant({
    required this.userId,
    required this.username,
    this.displayName,
    this.avatar,
    required this.status,
    this.score,
    this.xCount,
    this.sessionId,
  });

  factory EventParticipant.fromJson(Map<String, dynamic> json) {
    return EventParticipant(
      userId: json['user_id'] as String,
      username: json['username'] as String,
      displayName: json['display_name'] as String?,
      avatar: json['avatar'] as String?,
      status: json['status'] as String,
      score: json['score'] as int?,
      xCount: json['x_count'] as int?,
      sessionId: json['session_id'] as String?,
    );
  }

  String get effectiveName => displayName ?? username;
}
