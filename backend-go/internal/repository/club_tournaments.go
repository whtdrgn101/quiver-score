package repository

import (
	"context"
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
	StartedAt    *time.Time `json:"started_at"`
	CompletedAt  *time.Time `json:"completed_at"`
	CreatedAt    time.Time  `json:"created_at"`
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

func (r *ClubRepo) AddTournamentRound(ctx context.Context, id, clubID, tournamentID, userID, name string, templateID *string, advancement *int) (*TournamentRoundOut, error) {
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

	now := time.Now().UTC()
	_, err = r.DB.Exec(ctx,
		`INSERT INTO tournament_rounds (id, tournament_id, round_number, name, template_id, advancement, status, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, 'pending', $7)`,
		id, tournamentID, nextRound, name, templateID, advancement, now,
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
	err = r.DB.QueryRow(ctx,
		"SELECT status FROM tournament_rounds WHERE id = $1 AND tournament_id = $2",
		roundID, tournamentID,
	).Scan(&roundStatus)
	if err != nil || roundStatus != "pending" {
		return nil, pgx.ErrNoRows
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
	var advancement *int
	err = r.DB.QueryRow(ctx,
		"SELECT status, advancement FROM tournament_rounds WHERE id = $1 AND tournament_id = $2",
		roundID, tournamentID,
	).Scan(&roundStatus, &advancement)
	if err != nil || roundStatus != "in_progress" {
		return nil, pgx.ErrNoRows
	}

	// Rank all scores
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
				// Check for tie at the cutoff boundary
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

func (r *ClubRepo) getTournamentRoundOut(ctx context.Context, roundID string) (*TournamentRoundOut, error) {
	var ro TournamentRoundOut
	err := r.DB.QueryRow(ctx,
		`SELECT tr.id, tr.tournament_id, tr.round_number, tr.name, tr.template_id,
		        rt.name, tr.advancement, tr.status, tr.started_at, tr.completed_at, tr.created_at
		 FROM tournament_rounds tr
		 LEFT JOIN round_templates rt ON rt.id = tr.template_id
		 WHERE tr.id = $1`, roundID,
	).Scan(&ro.ID, &ro.TournamentID, &ro.RoundNumber, &ro.Name, &ro.TemplateID,
		&ro.TemplateName, &ro.Advancement, &ro.Status, &ro.StartedAt, &ro.CompletedAt, &ro.CreatedAt)
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
