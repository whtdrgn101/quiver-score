class ClubTeam {
  final String id;
  final String clubId;
  final String name;
  final String? description;
  final TeamMember leader;
  final int memberCount;
  final DateTime createdAt;

  const ClubTeam({
    required this.id,
    required this.clubId,
    required this.name,
    this.description,
    required this.leader,
    required this.memberCount,
    required this.createdAt,
  });

  factory ClubTeam.fromJson(Map<String, dynamic> json) {
    return ClubTeam(
      id: json['id'] as String,
      clubId: json['club_id'] as String,
      name: json['name'] as String,
      description: json['description'] as String?,
      leader: TeamMember.fromJson(json['leader'] as Map<String, dynamic>),
      memberCount: json['member_count'] as int? ?? 0,
      createdAt: DateTime.parse(json['created_at'] as String),
    );
  }
}

class ClubTeamDetail {
  final String id;
  final String clubId;
  final String name;
  final String? description;
  final TeamMember leader;
  final int memberCount;
  final DateTime createdAt;
  final List<TeamMember> members;

  const ClubTeamDetail({
    required this.id,
    required this.clubId,
    required this.name,
    this.description,
    required this.leader,
    required this.memberCount,
    required this.createdAt,
    this.members = const [],
  });

  factory ClubTeamDetail.fromJson(Map<String, dynamic> json) {
    return ClubTeamDetail(
      id: json['id'] as String,
      clubId: json['club_id'] as String,
      name: json['name'] as String,
      description: json['description'] as String?,
      leader: TeamMember.fromJson(json['leader'] as Map<String, dynamic>),
      memberCount: json['member_count'] as int? ?? 0,
      createdAt: DateTime.parse(json['created_at'] as String),
      members: (json['members'] as List?)
              ?.map((m) => TeamMember.fromJson(m as Map<String, dynamic>))
              .toList() ??
          [],
    );
  }
}

class TeamMember {
  final String userId;
  final String username;
  final String? displayName;
  final String? avatar;
  final DateTime joinedAt;

  const TeamMember({
    required this.userId,
    required this.username,
    this.displayName,
    this.avatar,
    required this.joinedAt,
  });

  factory TeamMember.fromJson(Map<String, dynamic> json) {
    return TeamMember(
      userId: json['user_id'] as String,
      username: json['username'] as String,
      displayName: json['display_name'] as String?,
      avatar: json['avatar'] as String?,
      joinedAt: DateTime.parse(json['joined_at'] as String),
    );
  }

  String get effectiveName => displayName ?? username;
}
