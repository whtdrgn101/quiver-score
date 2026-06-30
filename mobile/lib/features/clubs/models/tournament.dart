class Tournament {
  final String id;
  final String name;
  final String? description;
  final String organizerId;
  final String? organizerName;
  final String templateId;
  final String? templateName;
  final String status;
  final int? maxParticipants;
  final DateTime? registrationDeadline;
  final DateTime? startDate;
  final DateTime? endDate;
  final int participantCount;
  final String clubId;
  final String clubName;
  final DateTime createdAt;

  const Tournament({
    required this.id,
    required this.name,
    this.description,
    required this.organizerId,
    this.organizerName,
    required this.templateId,
    this.templateName,
    required this.status,
    this.maxParticipants,
    this.registrationDeadline,
    this.startDate,
    this.endDate,
    required this.participantCount,
    required this.clubId,
    required this.clubName,
    required this.createdAt,
  });

  factory Tournament.fromJson(Map<String, dynamic> json) {
    return Tournament(
      id: json['id'] as String,
      name: json['name'] as String,
      description: json['description'] as String?,
      organizerId: json['organizer_id'] as String,
      organizerName: json['organizer_name'] as String?,
      templateId: json['template_id'] as String,
      templateName: json['template_name'] as String?,
      status: json['status'] as String,
      maxParticipants: json['max_participants'] as int?,
      registrationDeadline: json['registration_deadline'] != null
          ? DateTime.parse(json['registration_deadline'] as String)
          : null,
      startDate: json['start_date'] != null
          ? DateTime.parse(json['start_date'] as String)
          : null,
      endDate: json['end_date'] != null
          ? DateTime.parse(json['end_date'] as String)
          : null,
      participantCount: json['participant_count'] as int? ?? 0,
      clubId: json['club_id'] as String,
      clubName: json['club_name'] as String? ?? '',
      createdAt: DateTime.parse(json['created_at'] as String),
    );
  }
}

class TournamentDetail {
  final String id;
  final String name;
  final String? description;
  final String organizerId;
  final String? organizerName;
  final String templateId;
  final String? templateName;
  final String status;
  final int? maxParticipants;
  final DateTime? registrationDeadline;
  final DateTime? startDate;
  final DateTime? endDate;
  final int participantCount;
  final String clubId;
  final String clubName;
  final DateTime createdAt;
  final List<TournamentParticipant> participants;

  const TournamentDetail({
    required this.id,
    required this.name,
    this.description,
    required this.organizerId,
    this.organizerName,
    required this.templateId,
    this.templateName,
    required this.status,
    this.maxParticipants,
    this.registrationDeadline,
    this.startDate,
    this.endDate,
    required this.participantCount,
    required this.clubId,
    required this.clubName,
    required this.createdAt,
    this.participants = const [],
  });

  factory TournamentDetail.fromJson(Map<String, dynamic> json) {
    return TournamentDetail(
      id: json['id'] as String,
      name: json['name'] as String,
      description: json['description'] as String?,
      organizerId: json['organizer_id'] as String,
      organizerName: json['organizer_name'] as String?,
      templateId: json['template_id'] as String,
      templateName: json['template_name'] as String?,
      status: json['status'] as String,
      maxParticipants: json['max_participants'] as int?,
      registrationDeadline: json['registration_deadline'] != null
          ? DateTime.parse(json['registration_deadline'] as String)
          : null,
      startDate: json['start_date'] != null
          ? DateTime.parse(json['start_date'] as String)
          : null,
      endDate: json['end_date'] != null
          ? DateTime.parse(json['end_date'] as String)
          : null,
      participantCount: json['participant_count'] as int? ?? 0,
      clubId: json['club_id'] as String,
      clubName: json['club_name'] as String? ?? '',
      createdAt: DateTime.parse(json['created_at'] as String),
      participants: (json['participants'] as List?)
              ?.map((p) =>
                  TournamentParticipant.fromJson(p as Map<String, dynamic>))
              .toList() ??
          [],
    );
  }
}

class TournamentParticipant {
  final String userId;
  final String? username;
  final int? finalScore;
  final int? finalXCount;
  final String status;

  const TournamentParticipant({
    required this.userId,
    this.username,
    this.finalScore,
    this.finalXCount,
    required this.status,
  });

  factory TournamentParticipant.fromJson(Map<String, dynamic> json) {
    return TournamentParticipant(
      userId: json['user_id'] as String,
      username: json['username'] as String?,
      finalScore: json['final_score'] as int?,
      finalXCount: json['final_x_count'] as int?,
      status: json['status'] as String,
    );
  }
}

class TournamentRound {
  final String id;
  final String tournamentId;
  final int roundNumber;
  final String name;
  final String? templateId;
  final String? templateName;
  final int? advancement;
  final String status;
  final String roundType;
  final DateTime? startedAt;
  final DateTime? completedAt;
  final DateTime createdAt;

  const TournamentRound({
    required this.id,
    required this.tournamentId,
    required this.roundNumber,
    required this.name,
    this.templateId,
    this.templateName,
    this.advancement,
    required this.status,
    required this.roundType,
    this.startedAt,
    this.completedAt,
    required this.createdAt,
  });

  factory TournamentRound.fromJson(Map<String, dynamic> json) {
    return TournamentRound(
      id: json['id'] as String,
      tournamentId: json['tournament_id'] as String,
      roundNumber: json['round_number'] as int,
      name: json['name'] as String,
      templateId: json['template_id'] as String?,
      templateName: json['template_name'] as String?,
      advancement: json['advancement'] as int?,
      status: json['status'] as String,
      roundType: json['round_type'] as String? ?? 'qualification',
      startedAt: json['started_at'] != null
          ? DateTime.parse(json['started_at'] as String)
          : null,
      completedAt: json['completed_at'] != null
          ? DateTime.parse(json['completed_at'] as String)
          : null,
      createdAt: DateTime.parse(json['created_at'] as String),
    );
  }
}

class TournamentMatchup {
  final String id;
  final String roundId;
  final int matchNumber;
  final String? participantAId;
  final String? participantAName;
  final String? participantBId;
  final String? participantBName;
  final int? scoreA;
  final int? scoreB;
  final String? winnerId;
  final String? winnerName;
  final DateTime createdAt;

  const TournamentMatchup({
    required this.id,
    required this.roundId,
    required this.matchNumber,
    this.participantAId,
    this.participantAName,
    this.participantBId,
    this.participantBName,
    this.scoreA,
    this.scoreB,
    this.winnerId,
    this.winnerName,
    required this.createdAt,
  });

  factory TournamentMatchup.fromJson(Map<String, dynamic> json) {
    return TournamentMatchup(
      id: json['id'] as String,
      roundId: json['round_id'] as String,
      matchNumber: json['match_number'] as int,
      participantAId: json['participant_a_id'] as String?,
      participantAName: json['participant_a_name'] as String?,
      participantBId: json['participant_b_id'] as String?,
      participantBName: json['participant_b_name'] as String?,
      scoreA: json['score_a'] as int?,
      scoreB: json['score_b'] as int?,
      winnerId: json['winner_id'] as String?,
      winnerName: json['winner_name'] as String?,
      createdAt: DateTime.parse(json['created_at'] as String),
    );
  }
}


class TournamentLeaderboardEntry {
  final int rank;
  final String userId;
  final String? username;
  final int? finalScore;
  final int? finalXCount;
  final String status;

  const TournamentLeaderboardEntry({
    required this.rank,
    required this.userId,
    this.username,
    this.finalScore,
    this.finalXCount,
    required this.status,
  });

  factory TournamentLeaderboardEntry.fromJson(Map<String, dynamic> json) {
    return TournamentLeaderboardEntry(
      rank: json['rank'] as int,
      userId: json['user_id'] as String,
      username: json['username'] as String?,
      finalScore: json['final_score'] as int?,
      finalXCount: json['final_x_count'] as int?,
      status: json['status'] as String,
    );
  }
}

class TournamentRoundScore {
  final String id;
  final String roundId;
  final String participantId;
  final String userId;
  final String? username;
  final String? sessionId;
  final int? score;
  final int? xCount;
  final int? rankInRound;
  final bool advanced;

  const TournamentRoundScore({
    required this.id,
    required this.roundId,
    required this.participantId,
    required this.userId,
    this.username,
    this.sessionId,
    this.score,
    this.xCount,
    this.rankInRound,
    required this.advanced,
  });

  factory TournamentRoundScore.fromJson(Map<String, dynamic> json) {
    return TournamentRoundScore(
      id: json['id'] as String,
      roundId: json['round_id'] as String,
      participantId: json['participant_id'] as String,
      userId: json['user_id'] as String,
      username: json['username'] as String?,
      sessionId: json['session_id'] as String?,
      score: json['score'] as int?,
      xCount: json['x_count'] as int?,
      rankInRound: json['rank_in_round'] as int?,
      advanced: json['advanced'] as bool? ?? false,
    );
  }
}

class TournamentContext {
  final String clubId;
  final String tournamentId;
  final String roundId;
  final String tournamentName;
  final String roundName;
  final String? matchupId;

  const TournamentContext({
    required this.clubId,
    required this.tournamentId,
    required this.roundId,
    required this.tournamentName,
    required this.roundName,
    this.matchupId,
  });
}
