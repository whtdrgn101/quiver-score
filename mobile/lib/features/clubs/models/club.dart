class Club {
  final String id;
  final String name;
  final String? description;
  final String? avatar;
  final String ownerId;
  final int memberCount;
  final String? myRole;
  final DateTime createdAt;

  const Club({
    required this.id,
    required this.name,
    this.description,
    this.avatar,
    required this.ownerId,
    required this.memberCount,
    this.myRole,
    required this.createdAt,
  });

  factory Club.fromJson(Map<String, dynamic> json) {
    return Club(
      id: json['id'] as String,
      name: json['name'] as String,
      description: json['description'] as String?,
      avatar: json['avatar'] as String?,
      ownerId: json['owner_id'] as String,
      memberCount: json['member_count'] as int? ?? 0,
      myRole: json['my_role'] as String?,
      createdAt: DateTime.parse(json['created_at'] as String),
    );
  }
}

class ClubDetail {
  final String id;
  final String name;
  final String? description;
  final String? avatar;
  final String ownerId;
  final int memberCount;
  final String? myRole;
  final DateTime createdAt;
  final List<ClubMember> members;

  const ClubDetail({
    required this.id,
    required this.name,
    this.description,
    this.avatar,
    required this.ownerId,
    required this.memberCount,
    this.myRole,
    required this.createdAt,
    this.members = const [],
  });

  factory ClubDetail.fromJson(Map<String, dynamic> json) {
    return ClubDetail(
      id: json['id'] as String,
      name: json['name'] as String,
      description: json['description'] as String?,
      avatar: json['avatar'] as String?,
      ownerId: json['owner_id'] as String,
      memberCount: json['member_count'] as int? ?? 0,
      myRole: json['my_role'] as String?,
      createdAt: DateTime.parse(json['created_at'] as String),
      members: (json['members'] as List?)
              ?.map((m) => ClubMember.fromJson(m as Map<String, dynamic>))
              .toList() ??
          [],
    );
  }
}

class ClubMember {
  final String userId;
  final String username;
  final String? displayName;
  final String? avatar;
  final String role;
  final DateTime joinedAt;

  const ClubMember({
    required this.userId,
    required this.username,
    this.displayName,
    this.avatar,
    required this.role,
    required this.joinedAt,
  });

  factory ClubMember.fromJson(Map<String, dynamic> json) {
    return ClubMember(
      userId: json['user_id'] as String,
      username: json['username'] as String,
      displayName: json['display_name'] as String?,
      avatar: json['avatar'] as String?,
      role: json['role'] as String,
      joinedAt: DateTime.parse(json['joined_at'] as String),
    );
  }

  String get effectiveName => displayName ?? username;
}

class LeaderboardGroup {
  final String templateId;
  final String templateName;
  final List<LeaderboardEntry> entries;

  const LeaderboardGroup({
    required this.templateId,
    required this.templateName,
    this.entries = const [],
  });

  factory LeaderboardGroup.fromJson(Map<String, dynamic> json) {
    return LeaderboardGroup(
      templateId: json['template_id'] as String,
      templateName: json['template_name'] as String,
      entries: (json['entries'] as List?)
              ?.map((e) => LeaderboardEntry.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
    );
  }
}

class LeaderboardEntry {
  final String userId;
  final String username;
  final String? displayName;
  final String? avatar;
  final int bestScore;
  final int bestXCount;
  final String sessionId;
  final DateTime achievedAt;

  const LeaderboardEntry({
    required this.userId,
    required this.username,
    this.displayName,
    this.avatar,
    required this.bestScore,
    required this.bestXCount,
    required this.sessionId,
    required this.achievedAt,
  });

  factory LeaderboardEntry.fromJson(Map<String, dynamic> json) {
    return LeaderboardEntry(
      userId: json['user_id'] as String,
      username: json['username'] as String,
      displayName: json['display_name'] as String?,
      avatar: json['avatar'] as String?,
      bestScore: json['best_score'] as int? ?? 0,
      bestXCount: json['best_x_count'] as int? ?? 0,
      sessionId: json['session_id'] as String,
      achievedAt: DateTime.parse(json['achieved_at'] as String),
    );
  }

  String get effectiveName => displayName ?? username;
}

class ActivityItem {
  final String type;
  final String userId;
  final String username;
  final String? displayName;
  final String? avatar;
  final String templateName;
  final int score;
  final int xCount;
  final String sessionId;
  final DateTime occurredAt;

  const ActivityItem({
    required this.type,
    required this.userId,
    required this.username,
    this.displayName,
    this.avatar,
    required this.templateName,
    required this.score,
    required this.xCount,
    required this.sessionId,
    required this.occurredAt,
  });

  factory ActivityItem.fromJson(Map<String, dynamic> json) {
    return ActivityItem(
      type: json['type'] as String,
      userId: json['user_id'] as String,
      username: json['username'] as String,
      displayName: json['display_name'] as String?,
      avatar: json['avatar'] as String?,
      templateName: json['template_name'] as String,
      score: json['score'] as int? ?? 0,
      xCount: json['x_count'] as int? ?? 0,
      sessionId: json['session_id'] as String,
      occurredAt: DateTime.parse(json['occurred_at'] as String),
    );
  }

  String get effectiveName => displayName ?? username;
}
