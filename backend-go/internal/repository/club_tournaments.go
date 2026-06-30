package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

// ── Tournament Types ─────────────────────────────────────────────────

type TournamentOut struct {
	ID                   string     `json:"id"`
	Name                 string     `json:"name"`
	Description          *string    `json:"description"`
	OrganizerID          string     `json:"organizer_id"`
	OrganizerName        *string    `json:"organizer_name"`
	TemplateID           string     `json:"template_id"`
	TemplateName         *string    `json:"template_name"`
	Status               string     `json:"status"`
	MaxParticipants      *int       `json:"max_participants"`
	RegistrationDeadline *time.Time `json:"registration_deadline"`
	StartDate            *time.Time `json:"start_date"`
	EndDate              *time.Time `json:"end_date"`
	ParticipantCount     int        `json:"participant_count"`
	ClubID               string     `json:"club_id"`
	ClubName             string     `json:"club_name"`
	CreatedAt            time.Time  `json:"created_at"`
}

type TournamentParticipantOut struct {
	UserID      string  `json:"user_id"`
	Username    *string `json:"username"`
	FinalScore  *int    `json:"final_score"`
	FinalXCount *int    `json:"final_x_count"`
	Status      string  `json:"status"`
}

type TournamentDetailOut struct {
	TournamentOut
	Participants []TournamentParticipantOut `json:"participants"`
}

type TournamentLeaderboardEntry struct {
	Rank        int     `json:"rank"`
	UserID      string  `json:"user_id"`
	Username    *string `json:"username"`
	FinalScore  *int    `json:"final_score"`
	FinalXCount *int    `json:"final_x_count"`
	Status      string  `json:"status"`
}

type TournamentRoundOut struct {
	ID           string     `json:"id"`
	TournamentID string     `json:"tournament_id"`
	RoundNumber  int        `json:"round_number"`
	Name         string     `json:"name"`
	TemplateID   *string    `json:"template_id"`
	TemplateName *string    `json:"template_name"`
	Advancement  *int       `json:"advancement"`
	Status       string     `json:"status"`
	RoundType    string     `json:"round_type"`
	StartedAt    *time.Time `json:"started_at"`
	CompletedAt  *time.Time `json:"completed_at"`
	CreatedAt    time.Time  `json:"created_at"`
}

type TournamentMatchupOut struct {
	ID               string     `json:"id"`
	RoundID          string     `json:"round_id"`
	MatchNumber      int        `json:"match_number"`
	ParticipantAID   *string    `json:"participant_a_id"`
	ParticipantAName *string    `json:"participant_a_name"`
	ParticipantBID   *string    `json:"participant_b_id"`
	ParticipantBName *string    `json:"participant_b_name"`
	ScoreA           *int       `json:"score_a"`
	ScoreB           *int       `json:"score_b"`
	WinnerID         *string    `json:"winner_id"`
	WinnerName       *string    `json:"winner_name"`
	CreatedAt        time.Time  `json:"created_at"`
}

type TournamentRoundScoreOut struct {
	ID            string  `json:"id"`
	RoundID       string  `json:"round_id"`
	ParticipantID string  `json:"participant_id"`
	UserID        string  `json:"user_id"`
	Username      *string `json:"username"`
	SessionID     *string `json:"session_id"`
	Score         *int    `json:"score"`
	XCount        *int    `json:"x_count"`
	RankInRound   *int    `json:"rank_in_round"`
	Advanced      bool    `json:"advanced"`
}

// ── Tournaments ───────────────────────────────────────────────────────

func (r *ClubRepo) CreateTournament(ctx context.Context, id, clubID, userID, name string, description *string, templateID string, maxParticipants *int, registrationDeadline, startDate, endDate time.Time) (*TournamentOut, error) {
	role, err := r.getMemberRole(ctx, clubID, userID)
	if err != nil || (role != "owner" && role != "admin") {
		return nil, pgx.ErrNoRows
	}

	now := time.Now().UTC()
	_, err = r.DB.Exec(ctx,
		`INSERT INTO tournaments (id, name, description, organizer_id, template_id, status, max_participants, registration_deadline, start_date, end_date, club_id, created_at)
		 VALUES ($1, $2, $3, $4, $5, 'registration', $6, $7, $8, $9, $10, $11)`,
		id, name, description, userID, templateID, maxParticipants, registrationDeadline, startDate, endDate, clubID, now,
	)
	if err != nil {
		return nil, err
	}

	return r.getTournamentOut(ctx, id)
}

func (r *ClubRepo) ListTournaments(ctx context.Context, clubID, userID string, status *string) ([]TournamentOut, error) {
	_, err := r.getMemberRole(ctx, clubID, userID)
	if err != nil {
		return nil, err
	}

	query := "SELECT id FROM tournaments WHERE club_id = $1"
	args := []any{clubID}
	if status != nil {
		query += " AND status = $2"
		args = append(args, *status)
	}
	query += " ORDER BY created_at DESC"

	rows, err := r.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tournaments []TournamentOut
	for rows.Next() {
		var tid string
		if err := rows.Scan(&tid); err != nil {
			return nil, err
		}
		t, err := r.getTournamentOut(ctx, tid)
		if err != nil {
			return nil, err
		}
		tournaments = append(tournaments, *t)
	}
	if tournaments == nil {
		tournaments = []TournamentOut{}
	}
	return tournaments, rows.Err()
}

func (r *ClubRepo) GetTournamentDetail(ctx context.Context, clubID, tournamentID, userID string) (*TournamentDetailOut, error) {
	_, err := r.getMemberRole(ctx, clubID, userID)
	if err != nil {
		return nil, err
	}

	t, err := r.getTournamentOut(ctx, tournamentID)
	if err != nil {
		return nil, err
	}

	// Load participants
	rows, err := r.DB.Query(ctx,
		`SELECT tp.user_id, u.username, tp.final_score, tp.final_x_count, tp.status
		 FROM tournament_participants tp
		 LEFT JOIN users u ON u.id = tp.user_id
		 WHERE tp.tournament_id = $1`, tournamentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var participants []TournamentParticipantOut
	for rows.Next() {
		var p TournamentParticipantOut
		if err := rows.Scan(&p.UserID, &p.Username, &p.FinalScore, &p.FinalXCount, &p.Status); err != nil {
			return nil, err
		}
		participants = append(participants, p)
	}
	if participants == nil {
		participants = []TournamentParticipantOut{}
	}

	return &TournamentDetailOut{TournamentOut: *t, Participants: participants}, nil
}

func (r *ClubRepo) RegisterForTournament(ctx context.Context, clubID, tournamentID, userID string) error {
	_, err := r.getMemberRole(ctx, clubID, userID)
	if err != nil {
		return err
	}

	var status string
	var maxParticipants *int
	err = r.DB.QueryRow(ctx,
		"SELECT status, max_participants FROM tournaments WHERE id = $1 AND club_id = $2",
		tournamentID, clubID,
	).Scan(&status, &maxParticipants)
	if err != nil {
		return pgx.ErrNoRows
	}
	if status != "registration" {
		return pgx.ErrNoRows
	}

	// Check capacity
	if maxParticipants != nil {
		var count int
		r.DB.QueryRow(ctx, "SELECT COUNT(*) FROM tournament_participants WHERE tournament_id = $1", tournamentID).Scan(&count)
		if count >= *maxParticipants {
			return pgx.ErrNoRows
		}
	}

	// Check not already registered
	var dup bool
	r.DB.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM tournament_participants WHERE tournament_id = $1 AND user_id = $2)", tournamentID, userID).Scan(&dup)
	if dup {
		return ErrAlreadyMember
	}

	now := time.Now().UTC()
	_, err = r.DB.Exec(ctx,
		"INSERT INTO tournament_participants (id, tournament_id, user_id, status, registered_at) VALUES ($1, $2, $3, 'registered', $4)",
		generateID(), tournamentID, userID, now,
	)
	return err
}

func (r *ClubRepo) StartTournament(ctx context.Context, clubID, tournamentID, userID string) (*TournamentOut, error) {
	var organizerID, status string
	err := r.DB.QueryRow(ctx,
		"SELECT organizer_id, status FROM tournaments WHERE id = $1 AND club_id = $2",
		tournamentID, clubID,
	).Scan(&organizerID, &status)
	if err != nil {
		return nil, pgx.ErrNoRows
	}
	if organizerID != userID {
		return nil, pgx.ErrNoRows
	}
	if status != "registration" {
		return nil, pgx.ErrNoRows
	}

	_, err = r.DB.Exec(ctx, "UPDATE tournaments SET status = 'in_progress' WHERE id = $1", tournamentID)
	if err != nil {
		return nil, err
	}
	_, err = r.DB.Exec(ctx, "UPDATE tournament_participants SET status = 'active' WHERE tournament_id = $1", tournamentID)
	if err != nil {
		return nil, err
	}

	return r.getTournamentOut(ctx, tournamentID)
}

func (r *ClubRepo) TournamentLeaderboard(ctx context.Context, clubID, tournamentID, userID string) ([]TournamentLeaderboardEntry, error) {
	_, err := r.getMemberRole(ctx, clubID, userID)
	if err != nil {
		return nil, err
	}

	rows, err := r.DB.Query(ctx,
		`SELECT tp.user_id, u.username, tp.final_score, tp.final_x_count, tp.status
		 FROM tournament_participants tp
		 LEFT JOIN users u ON u.id = tp.user_id
		 WHERE tp.tournament_id = $1`, tournamentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scored, unscored []TournamentLeaderboardEntry
	for rows.Next() {
		var e TournamentLeaderboardEntry
		if err := rows.Scan(&e.UserID, &e.Username, &e.FinalScore, &e.FinalXCount, &e.Status); err != nil {
			return nil, err
		}
		if e.FinalScore != nil {
			scored = append(scored, e)
		} else {
			unscored = append(unscored, e)
		}
	}

	// Sort scored by score DESC, x_count DESC
	for i := 0; i < len(scored); i++ {
		for j := i + 1; j < len(scored); j++ {
			si, sj := *scored[i].FinalScore, *scored[j].FinalScore
			xi, xj := 0, 0
			if scored[i].FinalXCount != nil {
				xi = *scored[i].FinalXCount
			}
			if scored[j].FinalXCount != nil {
				xj = *scored[j].FinalXCount
			}
			if sj > si || (sj == si && xj > xi) {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}

	// Assign ranks
	for i := range scored {
		scored[i].Rank = i + 1
	}

	result := append(scored, unscored...)
	if result == nil {
		result = []TournamentLeaderboardEntry{}
	}
	return result, nil
}

func (r *ClubRepo) CompleteTournament(ctx context.Context, clubID, tournamentID, userID string) (*TournamentOut, error) {
	var organizerID, status string
	err := r.DB.QueryRow(ctx,
		"SELECT organizer_id, status FROM tournaments WHERE id = $1 AND club_id = $2",
		tournamentID, clubID,
	).Scan(&organizerID, &status)
	if err != nil {
		return nil, pgx.ErrNoRows
	}
	if organizerID != userID {
		return nil, pgx.ErrNoRows
	}
	if status != "in_progress" {
		return nil, pgx.ErrNoRows
	}

	// Get leaderboard, assign ranks, update statuses
	lb, err := r.TournamentLeaderboard(ctx, clubID, tournamentID, userID)
	if err != nil {
		return nil, err
	}

	for _, e := range lb {
		if e.FinalScore != nil {
			r.DB.Exec(ctx,
				"UPDATE tournament_participants SET rank = $1, status = 'completed' WHERE tournament_id = $2 AND user_id = $3",
				e.Rank, tournamentID, e.UserID,
			)
		} else {
			r.DB.Exec(ctx,
				"UPDATE tournament_participants SET status = 'withdrawn' WHERE tournament_id = $1 AND user_id = $2",
				tournamentID, e.UserID,
			)
		}
	}

	_, err = r.DB.Exec(ctx, "UPDATE tournaments SET status = 'completed' WHERE id = $1", tournamentID)
	if err != nil {
		return nil, err
	}

	return r.getTournamentOut(ctx, tournamentID)
}

func (r *ClubRepo) WithdrawFromTournament(ctx context.Context, clubID, tournamentID, userID string) error {
	var pStatus string
	err := r.DB.QueryRow(ctx,
		"SELECT status FROM tournament_participants WHERE tournament_id = $1 AND user_id = $2",
		tournamentID, userID,
	).Scan(&pStatus)
	if err != nil {
		return pgx.ErrNoRows
	}
	if pStatus == "completed" || pStatus == "withdrawn" {
		return pgx.ErrNoRows
	}

	_, err = r.DB.Exec(ctx,
		"UPDATE tournament_participants SET status = 'withdrawn' WHERE tournament_id = $1 AND user_id = $2",
		tournamentID, userID,
	)
	return err
}

func (r *ClubRepo) SubmitTournamentScore(ctx context.Context, clubID, tournamentID, userID, sessionID string) (int, int, error) {
	// Verify participant
	var pStatus string
	err := r.DB.QueryRow(ctx,
		"SELECT status FROM tournament_participants WHERE tournament_id = $1 AND user_id = $2",
		tournamentID, userID,
	).Scan(&pStatus)
	if err != nil || (pStatus != "registered" && pStatus != "active") {
		return 0, 0, pgx.ErrNoRows
	}

	// Verify session is completed
	var score, xcount int
	err = r.DB.QueryRow(ctx,
		"SELECT total_score, total_x_count FROM scoring_sessions WHERE id = $1 AND user_id = $2 AND status = 'completed'",
		sessionID, userID,
	).Scan(&score, &xcount)
	if err != nil {
		return 0, 0, pgx.ErrNoRows
	}

	_, err = r.DB.Exec(ctx,
		"UPDATE tournament_participants SET final_score = $1, final_x_count = $2, status = 'completed' WHERE tournament_id = $3 AND user_id = $4",
		score, xcount, tournamentID, userID,
	)
	return score, xcount, err
}

// ── Tournament Rounds ────────────────────────────────────────────────

func (r *ClubRepo) AddTournamentRound(ctx context.Context, id, clubID, tournamentID, userID, name string, templateID *string, advancement *int, roundType string) (*TournamentRoundOut, error) {
	var organizerID, status string
	err := r.DB.QueryRow(ctx,
		"SELECT organizer_id, status FROM tournaments WHERE id = $1 AND club_id = $2",
		tournamentID, clubID,
	).Scan(&organizerID, &status)
	if err != nil {
		return nil, pgx.ErrNoRows
	}
	if organizerID != userID {
		return nil, pgx.ErrNoRows
	}
	if status != "registration" && status != "in_progress" {
		return nil, pgx.ErrNoRows
	}

	var nextRound int
	r.DB.QueryRow(ctx,
		"SELECT COALESCE(MAX(round_number), 0) + 1 FROM tournament_rounds WHERE tournament_id = $1",
		tournamentID,
	).Scan(&nextRound)

	if roundType == "" {
		roundType = "qualification"
	}

	now := time.Now().UTC()
	_, err = r.DB.Exec(ctx,
		`INSERT INTO tournament_rounds (id, tournament_id, round_number, name, template_id, advancement, status, round_type, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, 'pending', $7, $8)`,
		id, tournamentID, nextRound, name, templateID, advancement, roundType, now,
	)
	if err != nil {
		return nil, err
	}

	return r.getTournamentRoundOut(ctx, id)
}

func (r *ClubRepo) ListTournamentRounds(ctx context.Context, clubID, tournamentID, userID string) ([]TournamentRoundOut, error) {
	_, err := r.getMemberRole(ctx, clubID, userID)
	if err != nil {
		return nil, err
	}

	rows, err := r.DB.Query(ctx,
		"SELECT id FROM tournament_rounds WHERE tournament_id = $1 ORDER BY round_number",
		tournamentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rounds []TournamentRoundOut
	for rows.Next() {
		var rid string
		if err := rows.Scan(&rid); err != nil {
			return nil, err
		}
		ro, err := r.getTournamentRoundOut(ctx, rid)
		if err != nil {
			return nil, err
		}
		rounds = append(rounds, *ro)
	}
	if rounds == nil {
		rounds = []TournamentRoundOut{}
	}
	return rounds, rows.Err()
}

func (r *ClubRepo) StartTournamentRound(ctx context.Context, clubID, tournamentID, roundID, userID string) (*TournamentRoundOut, error) {
	var organizerID string
	err := r.DB.QueryRow(ctx,
		"SELECT organizer_id FROM tournaments WHERE id = $1 AND club_id = $2",
		tournamentID, clubID,
	).Scan(&organizerID)
	if err != nil || organizerID != userID {
		return nil, pgx.ErrNoRows
	}

	var roundStatus string
	var roundType string
	var roundNumber int
	err = r.DB.QueryRow(ctx,
		"SELECT status, round_type, round_number FROM tournament_rounds WHERE id = $1 AND tournament_id = $2",
		roundID, tournamentID,
	).Scan(&roundStatus, &roundType, &roundNumber)
	if err != nil || roundStatus != "pending" {
		return nil, pgx.ErrNoRows
	}

	if roundType == "elimination" {
		if roundNumber <= 1 {
			return nil, errors.New("cannot have elimination as first round")
		}

		var prevRoundID string
		var prevRoundType string
		err = r.DB.QueryRow(ctx,
			"SELECT id, round_type FROM tournament_rounds WHERE tournament_id = $1 AND round_number = $2",
			tournamentID, roundNumber-1,
		).Scan(&prevRoundID, &prevRoundType)
		if err != nil {
			return nil, errors.New("previous round not found")
		}

		var participants []string

		if prevRoundType == "qualification" {
			rows, err := r.DB.Query(ctx,
				`SELECT participant_id FROM tournament_round_scores
				 WHERE round_id = $1 AND advanced = true
				 ORDER BY rank_in_round ASC`,
				prevRoundID,
			)
			if err != nil {
				return nil, err
			}
			defer rows.Close()

			for rows.Next() {
				var pid string
				if err := rows.Scan(&pid); err != nil {
					return nil, err
				}
				participants = append(participants, pid)
			}
			if err := rows.Err(); err != nil {
				return nil, err
			}

			if len(participants) == 0 {
				return nil, errors.New("no participants advanced from the previous round")
			}

			M := len(participants)
			N := 2
			for N < M {
				N *= 2
			}

			bracketOrder := getBracketOrder(N)
			numMatches := N / 2

			tx, err := r.DB.Begin(ctx)
			if err != nil {
				return nil, err
			}
			defer tx.Rollback(ctx)

			for j := 0; j < numMatches; j++ {
				seedA := bracketOrder[2*j]
				seedB := bracketOrder[2*j+1]

				var partA, partB *string
				if seedA <= M {
					partA = &participants[seedA-1]
				}
				if seedB <= M {
					partB = &participants[seedB-1]
				}

				matchID := generateID()
				matchNum := j + 1

				var winnerID *string
				if partA != nil && partB == nil {
					winnerID = partA
				} else if partB != nil && partA == nil {
					winnerID = partB
				}

				_, err = tx.Exec(ctx,
					`INSERT INTO tournament_matchups (id, round_id, match_number, participant_a_id, participant_b_id, winner_id)
					 VALUES ($1, $2, $3, $4, $5, $6)`,
					matchID, roundID, matchNum, partA, partB, winnerID,
				)
				if err != nil {
					return nil, err
				}
			}

			if err := tx.Commit(ctx); err != nil {
				return nil, err
			}

		} else if prevRoundType == "elimination" {
			rows, err := r.DB.Query(ctx,
				`SELECT winner_id FROM tournament_matchups
				 WHERE round_id = $1
				 ORDER BY match_number ASC`,
				prevRoundID,
			)
			if err != nil {
				return nil, err
			}
			defer rows.Close()

			var winners []*string
			for rows.Next() {
				var wid *string
				if err := rows.Scan(&wid); err != nil {
					return nil, err
				}
				winners = append(winners, wid)
			}
			if err := rows.Err(); err != nil {
				return nil, err
			}

			K := len(winners)
			if K == 0 {
				return nil, errors.New("no matchups found in the previous round")
			}

			numMatches := K / 2
			if numMatches == 0 {
				return nil, errors.New("cannot advance: only 1 matchup existed in the previous round")
			}

			tx, err := r.DB.Begin(ctx)
			if err != nil {
				return nil, err
			}
			defer tx.Rollback(ctx)

			for j := 0; j < numMatches; j++ {
				partA := winners[2*j]
				partB := winners[2*j+1]

				matchID := generateID()
				matchNum := j + 1

				var winnerID *string
				if partA != nil && partB == nil {
					winnerID = partA
				} else if partB != nil && partA == nil {
					winnerID = partB
				}

				_, err = tx.Exec(ctx,
					`INSERT INTO tournament_matchups (id, round_id, match_number, participant_a_id, participant_b_id, winner_id)
					 VALUES ($1, $2, $3, $4, $5, $6)`,
					matchID, roundID, matchNum, partA, partB, winnerID,
				)
				if err != nil {
					return nil, err
				}
			}

			if err := tx.Commit(ctx); err != nil {
				return nil, err
			}
		}
	}

	now := time.Now().UTC()
	_, err = r.DB.Exec(ctx,
		"UPDATE tournament_rounds SET status = 'in_progress', started_at = $1 WHERE id = $2",
		now, roundID,
	)
	if err != nil {
		return nil, err
	}

	return r.getTournamentRoundOut(ctx, roundID)
}

func getBracketOrder(n int) []int {
	order := []int{1}
	for len(order) < n {
		nextOrder := make([]int, len(order)*2)
		target := len(order)*2 + 1
		for i, v := range order {
			nextOrder[i*2] = v
			nextOrder[i*2+1] = target - v
		}
		order = nextOrder
	}
	return order
}

func (r *ClubRepo) SubmitTournamentRoundScore(ctx context.Context, clubID, tournamentID, roundID, userID, sessionID string) (*TournamentRoundScoreOut, error) {
	_, err := r.getMemberRole(ctx, clubID, userID)
	if err != nil {
		return nil, err
	}

	var roundStatus string
	err = r.DB.QueryRow(ctx,
		"SELECT status FROM tournament_rounds WHERE id = $1 AND tournament_id = $2",
		roundID, tournamentID,
	).Scan(&roundStatus)
	if err != nil || roundStatus != "in_progress" {
		return nil, pgx.ErrNoRows
	}

	// Find participant
	var participantID string
	err = r.DB.QueryRow(ctx,
		"SELECT id FROM tournament_participants WHERE tournament_id = $1 AND user_id = $2 AND status IN ('registered', 'active')",
		tournamentID, userID,
	).Scan(&participantID)
	if err != nil {
		return nil, pgx.ErrNoRows
	}

	// Get session score
	var score, xcount int
	err = r.DB.QueryRow(ctx,
		"SELECT total_score, total_x_count FROM scoring_sessions WHERE id = $1 AND user_id = $2 AND status = 'completed'",
		sessionID, userID,
	).Scan(&score, &xcount)
	if err != nil {
		return nil, pgx.ErrNoRows
	}

	// Upsert round score
	scoreID := generateID()
	_, err = r.DB.Exec(ctx,
		`INSERT INTO tournament_round_scores (id, round_id, participant_id, session_id, score, x_count)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT ON CONSTRAINT uq_round_participant
		 DO UPDATE SET session_id = $4, score = $5, x_count = $6`,
		scoreID, roundID, participantID, sessionID, score, xcount,
	)
	if err != nil {
		return nil, err
	}

	return r.getTournamentRoundScoreOut(ctx, roundID, participantID)
}

func (r *ClubRepo) GetTournamentRoundLeaderboard(ctx context.Context, clubID, tournamentID, roundID, userID string) ([]TournamentRoundScoreOut, error) {
	_, err := r.getMemberRole(ctx, clubID, userID)
	if err != nil {
		return nil, err
	}

	rows, err := r.DB.Query(ctx,
		`SELECT trs.id, trs.round_id, trs.participant_id, tp.user_id, u.username,
		        trs.session_id, trs.score, trs.x_count, trs.rank_in_round, trs.advanced
		 FROM tournament_round_scores trs
		 JOIN tournament_participants tp ON tp.id = trs.participant_id
		 LEFT JOIN users u ON u.id = tp.user_id
		 WHERE trs.round_id = $1
		 ORDER BY trs.score DESC NULLS LAST, trs.x_count DESC NULLS LAST`,
		roundID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scores []TournamentRoundScoreOut
	for rows.Next() {
		var s TournamentRoundScoreOut
		if err := rows.Scan(&s.ID, &s.RoundID, &s.ParticipantID, &s.UserID, &s.Username,
			&s.SessionID, &s.Score, &s.XCount, &s.RankInRound, &s.Advanced); err != nil {
			return nil, err
		}
		scores = append(scores, s)
	}
	if scores == nil {
		scores = []TournamentRoundScoreOut{}
	}
	return scores, rows.Err()
}

func (r *ClubRepo) CompleteTournamentRound(ctx context.Context, clubID, tournamentID, roundID, userID string) (*TournamentRoundOut, error) {
	var organizerID string
	err := r.DB.QueryRow(ctx,
		"SELECT organizer_id FROM tournaments WHERE id = $1 AND club_id = $2",
		tournamentID, clubID,
	).Scan(&organizerID)
	if err != nil || organizerID != userID {
		return nil, pgx.ErrNoRows
	}

	var roundStatus string
	var roundType string
	var advancement *int
	err = r.DB.QueryRow(ctx,
		"SELECT status, round_type, advancement FROM tournament_rounds WHERE id = $1 AND tournament_id = $2",
		roundID, tournamentID,
	).Scan(&roundStatus, &roundType, &advancement)
	if err != nil || roundStatus != "in_progress" {
		return nil, pgx.ErrNoRows
	}

	if roundType == "elimination" {
		rows, err := r.DB.Query(ctx,
			"SELECT participant_a_id, participant_b_id, winner_id, score_a, score_b FROM tournament_matchups WHERE round_id = $1",
			roundID,
		)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		type matchupCheck struct {
			partA    *string
			partB    *string
			winnerID *string
			scoreA   *int
			scoreB   *int
		}
		var matchups []matchupCheck
		for rows.Next() {
			var m matchupCheck
			if err := rows.Scan(&m.partA, &m.partB, &m.winnerID, &m.scoreA, &m.scoreB); err != nil {
				return nil, err
			}
			matchups = append(matchups, m)
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}

		for _, m := range matchups {
			if (m.partA != nil || m.partB != nil) && m.winnerID == nil {
				return nil, errors.New("cannot complete round: some matchups are pending scores/winner resolution")
			}
		}

		tx, err := r.DB.Begin(ctx)
		if err != nil {
			return nil, err
		}
		defer tx.Rollback(ctx)

		for _, m := range matchups {
			if m.partA != nil {
				scoreID := generateID()
				advanced := m.winnerID != nil && *m.winnerID == *m.partA
				_, err = tx.Exec(ctx,
					`INSERT INTO tournament_round_scores (id, round_id, participant_id, score, x_count, advanced)
					 VALUES ($1, $2, $3, $4, 0, $5)
					 ON CONFLICT ON CONSTRAINT uq_round_participant
					 DO UPDATE SET score = $4, advanced = $5`,
					scoreID, roundID, *m.partA, m.scoreA, advanced,
				)
				if err != nil {
					return nil, err
				}
			}
			if m.partB != nil {
				scoreID := generateID()
				advanced := m.winnerID != nil && *m.winnerID == *m.partB
				_, err = tx.Exec(ctx,
					`INSERT INTO tournament_round_scores (id, round_id, participant_id, score, x_count, advanced)
					 VALUES ($1, $2, $3, $4, 0, $5)
					 ON CONFLICT ON CONSTRAINT uq_round_participant
					 DO UPDATE SET score = $4, advanced = $5`,
					scoreID, roundID, *m.partB, m.scoreB, advanced,
				)
				if err != nil {
					return nil, err
				}
			}
		}

		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}

	} else {
		scores, err := r.GetTournamentRoundLeaderboard(ctx, clubID, tournamentID, roundID, userID)
		if err != nil {
			return nil, err
		}

		for i, s := range scores {
			rank := i + 1
			advanced := false
			if advancement != nil && s.Score != nil {
				if rank <= *advancement {
					advanced = true
				} else if rank > *advancement {
					cutoffScore := scores[*advancement-1].Score
					cutoffX := scores[*advancement-1].XCount
					if s.Score != nil && cutoffScore != nil && *s.Score == *cutoffScore {
						sX := 0
						cX := 0
						if s.XCount != nil {
							sX = *s.XCount
						}
						if cutoffX != nil {
							cX = *cutoffX
						}
						if sX >= cX {
							advanced = true
						}
					}
				}
			}

			r.DB.Exec(ctx,
				"UPDATE tournament_round_scores SET rank_in_round = $1, advanced = $2 WHERE id = $3",
				rank, advanced, s.ID,
			)
		}
	}

	now := time.Now().UTC()
	_, err = r.DB.Exec(ctx,
		"UPDATE tournament_rounds SET status = 'completed', completed_at = $1 WHERE id = $2",
		now, roundID,
	)
	if err != nil {
		return nil, err
	}

	return r.getTournamentRoundOut(ctx, roundID)
}

// ── Tournament Matchups ──────────────────────────────────────────────

func (r *ClubRepo) GetTournamentMatchups(ctx context.Context, roundID string) ([]TournamentMatchupOut, error) {
	rows, err := r.DB.Query(ctx,
		`SELECT tm.id, tm.round_id, tm.match_number,
		        tm.participant_a_id, ua.username,
		        tm.participant_b_id, ub.username,
		        tm.score_a, tm.score_b,
		        tm.winner_id, uw.username,
		        tm.created_at
		 FROM tournament_matchups tm
		 LEFT JOIN tournament_participants pa ON pa.id = tm.participant_a_id
		 LEFT JOIN users ua ON ua.id = pa.user_id
		 LEFT JOIN tournament_participants pb ON pb.id = tm.participant_b_id
		 LEFT JOIN users ub ON ub.id = pb.user_id
		 LEFT JOIN tournament_participants pw ON pw.id = tm.winner_id
		 LEFT JOIN users uw ON uw.id = pw.user_id
		 WHERE tm.round_id = $1
		 ORDER BY tm.match_number`,
		roundID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matchups []TournamentMatchupOut
	for rows.Next() {
		var tm TournamentMatchupOut
		if err := rows.Scan(&tm.ID, &tm.RoundID, &tm.MatchNumber,
			&tm.ParticipantAID, &tm.ParticipantAName,
			&tm.ParticipantBID, &tm.ParticipantBName,
			&tm.ScoreA, &tm.ScoreB,
			&tm.WinnerID, &tm.WinnerName,
			&tm.CreatedAt); err != nil {
			return nil, err
		}
		matchups = append(matchups, tm)
	}
	if matchups == nil {
		matchups = []TournamentMatchupOut{}
	}
	return matchups, rows.Err()
}

func (r *ClubRepo) SubmitMatchupScore(ctx context.Context, clubID, tournamentID, roundID, matchupID, userID, sessionID string) (*TournamentMatchupOut, error) {
	_, err := r.getMemberRole(ctx, clubID, userID)
	if err != nil {
		return nil, err
	}

	var roundStatus string
	err = r.DB.QueryRow(ctx,
		"SELECT status FROM tournament_rounds WHERE id = $1 AND tournament_id = $2",
		roundID, tournamentID,
	).Scan(&roundStatus)
	if err != nil || roundStatus != "in_progress" {
		return nil, pgx.ErrNoRows
	}

	var score int
	err = r.DB.QueryRow(ctx,
		"SELECT total_score FROM scoring_sessions WHERE id = $1 AND user_id = $2 AND status = 'completed'",
		sessionID, userID,
	).Scan(&score)
	if err != nil {
		return nil, pgx.ErrNoRows
	}

	var partAID, partBID *string
	var userAID, userBID *string
	err = r.DB.QueryRow(ctx,
		`SELECT tm.participant_a_id, pa.user_id, tm.participant_b_id, pb.user_id
		 FROM tournament_matchups tm
		 LEFT JOIN tournament_participants pa ON pa.id = tm.participant_a_id
		 LEFT JOIN tournament_participants pb ON pb.id = tm.participant_b_id
		 WHERE tm.id = $1 AND tm.round_id = $2`,
		matchupID, roundID,
	).Scan(&partAID, &userAID, &partBID, &userBID)
	if err != nil {
		return nil, pgx.ErrNoRows
	}

	var updateCol string
	if userAID != nil && *userAID == userID {
		updateCol = "score_a"
	} else if userBID != nil && *userBID == userID {
		updateCol = "score_b"
	} else {
		return nil, errors.New("user is not a participant in this matchup")
	}

	_, err = r.DB.Exec(ctx,
		"UPDATE tournament_matchups SET "+updateCol+" = $1 WHERE id = $2",
		score, matchupID,
	)
	if err != nil {
		return nil, err
	}

	var scoreA, scoreB *int
	err = r.DB.QueryRow(ctx,
		"SELECT score_a, score_b FROM tournament_matchups WHERE id = $1",
		matchupID,
	).Scan(&scoreA, &scoreB)
	if err != nil {
		return nil, err
	}

	var winnerID *string
	if scoreA != nil && scoreB != nil {
		if *scoreA > *scoreB {
			winnerID = partAID
		} else if *scoreB > *scoreA {
			winnerID = partBID
		}
		if winnerID != nil {
			_, err = r.DB.Exec(ctx,
				"UPDATE tournament_matchups SET winner_id = $1 WHERE id = $2",
				winnerID, matchupID,
			)
			if err != nil {
				return nil, err
			}
		}
	}

	return r.getMatchupOut(ctx, matchupID)
}

func (r *ClubRepo) UpdateMatchup(ctx context.Context, clubID, tournamentID, roundID, matchupID, userID string, scoreA, scoreB *int, winnerID *string) (*TournamentMatchupOut, error) {
	var organizerID string
	err := r.DB.QueryRow(ctx,
		"SELECT organizer_id FROM tournaments WHERE id = $1 AND club_id = $2",
		tournamentID, clubID,
	).Scan(&organizerID)
	if err != nil || organizerID != userID {
		return nil, pgx.ErrNoRows
	}

	_, err = r.DB.Exec(ctx,
		`UPDATE tournament_matchups
		 SET score_a = $1, score_b = $2, winner_id = $3
		 WHERE id = $4 AND round_id = $5`,
		scoreA, scoreB, winnerID, matchupID, roundID,
	)
	if err != nil {
		return nil, err
	}

	return r.getMatchupOut(ctx, matchupID)
}

func (r *ClubRepo) getMatchupOut(ctx context.Context, matchupID string) (*TournamentMatchupOut, error) {
	var tm TournamentMatchupOut
	err := r.DB.QueryRow(ctx,
		`SELECT tm.id, tm.round_id, tm.match_number,
		        tm.participant_a_id, ua.username,
		        tm.participant_b_id, ub.username,
		        tm.score_a, tm.score_b,
		        tm.winner_id, uw.username,
		        tm.created_at
		 FROM tournament_matchups tm
		 LEFT JOIN tournament_participants pa ON pa.id = tm.participant_a_id
		 LEFT JOIN users ua ON ua.id = pa.user_id
		 LEFT JOIN tournament_participants pb ON pb.id = tm.participant_b_id
		 LEFT JOIN users ub ON ub.id = pb.user_id
		 LEFT JOIN tournament_participants pw ON pw.id = tm.winner_id
		 LEFT JOIN users uw ON uw.id = pw.user_id
		 WHERE tm.id = $1`,
		matchupID,
	).Scan(&tm.ID, &tm.RoundID, &tm.MatchNumber,
		&tm.ParticipantAID, &tm.ParticipantAName,
		&tm.ParticipantBID, &tm.ParticipantBName,
		&tm.ScoreA, &tm.ScoreB,
		&tm.WinnerID, &tm.WinnerName,
		&tm.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &tm, nil
}

func (r *ClubRepo) getTournamentRoundOut(ctx context.Context, roundID string) (*TournamentRoundOut, error) {
	var ro TournamentRoundOut
	err := r.DB.QueryRow(ctx,
		`SELECT tr.id, tr.tournament_id, tr.round_number, tr.name, tr.template_id,
		        rt.name, tr.advancement, tr.status, tr.started_at, tr.completed_at, tr.created_at, tr.round_type
		 FROM tournament_rounds tr
		 LEFT JOIN round_templates rt ON rt.id = tr.template_id
		 WHERE tr.id = $1`, roundID,
	).Scan(&ro.ID, &ro.TournamentID, &ro.RoundNumber, &ro.Name, &ro.TemplateID,
		&ro.TemplateName, &ro.Advancement, &ro.Status, &ro.StartedAt, &ro.CompletedAt, &ro.CreatedAt, &ro.RoundType)
	if err != nil {
		return nil, err
	}
	return &ro, nil
}

func (r *ClubRepo) getTournamentRoundScoreOut(ctx context.Context, roundID, participantID string) (*TournamentRoundScoreOut, error) {
	var s TournamentRoundScoreOut
	err := r.DB.QueryRow(ctx,
		`SELECT trs.id, trs.round_id, trs.participant_id, tp.user_id, u.username,
		        trs.session_id, trs.score, trs.x_count, trs.rank_in_round, trs.advanced
		 FROM tournament_round_scores trs
		 JOIN tournament_participants tp ON tp.id = trs.participant_id
		 LEFT JOIN users u ON u.id = tp.user_id
		 WHERE trs.round_id = $1 AND trs.participant_id = $2`, roundID, participantID,
	).Scan(&s.ID, &s.RoundID, &s.ParticipantID, &s.UserID, &s.Username,
		&s.SessionID, &s.Score, &s.XCount, &s.RankInRound, &s.Advanced)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *ClubRepo) getTournamentOut(ctx context.Context, tournamentID string) (*TournamentOut, error) {
	var t TournamentOut
	err := r.DB.QueryRow(ctx,
		`SELECT t.id, t.name, t.description, t.organizer_id, t.template_id, t.status,
		        t.max_participants, t.registration_deadline, t.start_date, t.end_date,
		        t.club_id, t.created_at,
		        u.display_name, rt.name, c.name,
		        (SELECT COUNT(*) FROM tournament_participants WHERE tournament_id = t.id)
		 FROM tournaments t
		 LEFT JOIN users u ON u.id = t.organizer_id
		 LEFT JOIN round_templates rt ON rt.id = t.template_id
		 LEFT JOIN clubs c ON c.id = t.club_id
		 WHERE t.id = $1`, tournamentID,
	).Scan(&t.ID, &t.Name, &t.Description, &t.OrganizerID, &t.TemplateID, &t.Status,
		&t.MaxParticipants, &t.RegistrationDeadline, &t.StartDate, &t.EndDate,
		&t.ClubID, &t.CreatedAt,
		&t.OrganizerName, &t.TemplateName, &t.ClubName,
		&t.ParticipantCount,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
