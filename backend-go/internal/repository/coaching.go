package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CoachingRepo struct {
	DB *pgxpool.Pool
}

// ── Types ───────────────────────────────────────────────────────────────

type CoachAthleteLinkOut struct {
	ID              string    `json:"id"`
	CoachID         string    `json:"coach_id"`
	AthleteID       string    `json:"athlete_id"`
	CoachUsername    *string   `json:"coach_username"`
	AthleteUsername  *string   `json:"athlete_username"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
}

type AthleteSessionOut struct {
	ID           string     `json:"id"`
	TemplateName string     `json:"template_name"`
	TotalScore   int        `json:"total_score"`
	TotalXCount  int        `json:"total_x_count"`
	TotalArrows  int        `json:"total_arrows"`
	CompletedAt  *time.Time `json:"completed_at"`
}

type AnnotationOut struct {
	ID             string    `json:"id"`
	SessionID      string    `json:"session_id"`
	AuthorID       string    `json:"author_id"`
	AuthorUsername *string   `json:"author_username"`
	EndNumber      *int      `json:"end_number"`
	ArrowNumber    *int      `json:"arrow_number"`
	Text           string    `json:"text"`
	CreatedAt      time.Time `json:"created_at"`
}

// ── Invite ──────────────────────────────────────────────────────────────

func (r *CoachingRepo) Invite(ctx context.Context, coachID, athleteUsername string) (*CoachAthleteLinkOut, error) {
	// Look up athlete by username
	var athleteID string
	err := r.DB.QueryRow(ctx, `SELECT id FROM users WHERE username = $1`, athleteUsername).Scan(&athleteID)
	if err != nil {
		return nil, ErrNotFound
	}

	if athleteID == coachID {
		return nil, ErrValidation
	}

	// Check existing link
	var exists bool
	err = r.DB.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM coach_athlete_links WHERE coach_id = $1 AND athlete_id = $2)`,
		coachID, athleteID,
	).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrAlreadyMember
	}

	id := uuid.New().String()
	now := time.Now().UTC()
	_, err = r.DB.Exec(ctx,
		`INSERT INTO coach_athlete_links (id, coach_id, athlete_id, status, created_at) VALUES ($1, $2, $3, 'pending', $4)`,
		id, coachID, athleteID, now,
	)
	if err != nil {
		return nil, err
	}

	return r.getLink(ctx, id)
}

// ── Respond ─────────────────────────────────────────────────────────────

func (r *CoachingRepo) Respond(ctx context.Context, athleteID, linkID string, accept bool) (*CoachAthleteLinkOut, error) {
	// Find the link
	var currentStatus string
	err := r.DB.QueryRow(ctx,
		`SELECT status FROM coach_athlete_links WHERE id = $1 AND athlete_id = $2`,
		linkID, athleteID,
	).Scan(&currentStatus)
	if err != nil {
		return nil, ErrNotFound
	}

	if currentStatus != "pending" {
		return nil, ErrValidation
	}

	newStatus := "revoked"
	if accept {
		newStatus = "active"
	}

	_, err = r.DB.Exec(ctx,
		`UPDATE coach_athlete_links SET status = $1 WHERE id = $2`,
		newStatus, linkID,
	)
	if err != nil {
		return nil, err
	}

	return r.getLink(ctx, linkID)
}

// ── List Athletes / Coaches ─────────────────────────────────────────────

func (r *CoachingRepo) ListAthletes(ctx context.Context, coachID string) ([]CoachAthleteLinkOut, error) {
	return r.listLinks(ctx, "coach_id", coachID)
}

func (r *CoachingRepo) ListCoaches(ctx context.Context, athleteID string) ([]CoachAthleteLinkOut, error) {
	return r.listLinks(ctx, "athlete_id", athleteID)
}

// ── Athlete Sessions ────────────────────────────────────────────────────

func (r *CoachingRepo) GetAthleteSessions(ctx context.Context, coachID, athleteID string) ([]AthleteSessionOut, error) {
	// Verify active coaching link
	var exists bool
	err := r.DB.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM coach_athlete_links WHERE coach_id = $1 AND athlete_id = $2 AND status = 'active')`,
		coachID, athleteID,
	).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrForbidden
	}

	rows, err := r.DB.Query(ctx, `
		SELECT ss.id, COALESCE(rt.name, 'Unknown'), ss.total_score, ss.total_x_count, ss.total_arrows, ss.completed_at
		FROM scoring_sessions ss
		LEFT JOIN round_templates rt ON rt.id = ss.template_id
		WHERE ss.user_id = $1 AND ss.status = 'completed'
		ORDER BY ss.completed_at DESC
	`, athleteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []AthleteSessionOut
	for rows.Next() {
		var s AthleteSessionOut
		if err := rows.Scan(&s.ID, &s.TemplateName, &s.TotalScore, &s.TotalXCount, &s.TotalArrows, &s.CompletedAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	if sessions == nil {
		sessions = []AthleteSessionOut{}
	}
	return sessions, nil
}

// ── Annotations ─────────────────────────────────────────────────────────

func (r *CoachingRepo) CheckSessionAccess(ctx context.Context, userID, sessionID string) (string, error) {
	// Returns session owner ID if user has access (is owner or active coach)
	var ownerID string
	err := r.DB.QueryRow(ctx, `SELECT user_id FROM scoring_sessions WHERE id = $1`, sessionID).Scan(&ownerID)
	if err != nil {
		return "", ErrNotFound
	}

	if ownerID == userID {
		return ownerID, nil
	}

	// Check active coaching link
	var hasLink bool
	err = r.DB.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM coach_athlete_links WHERE coach_id = $1 AND athlete_id = $2 AND status = 'active')`,
		userID, ownerID,
	).Scan(&hasLink)
	if err != nil {
		return "", err
	}
	if !hasLink {
		return "", ErrForbidden
	}

	return ownerID, nil
}

func (r *CoachingRepo) AddAnnotation(ctx context.Context, sessionID, authorID string, endNumber, arrowNumber *int, text string) (*AnnotationOut, error) {
	id := uuid.New().String()
	now := time.Now().UTC()

	_, err := r.DB.Exec(ctx, `
		INSERT INTO session_annotations (id, session_id, author_id, end_number, arrow_number, text, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, id, sessionID, authorID, endNumber, arrowNumber, text, now)
	if err != nil {
		return nil, err
	}

	return r.getAnnotation(ctx, id)
}

func (r *CoachingRepo) ListAnnotations(ctx context.Context, sessionID string) ([]AnnotationOut, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT sa.id, sa.session_id, sa.author_id, u.username, sa.end_number, sa.arrow_number, sa.text, sa.created_at
		FROM session_annotations sa
		LEFT JOIN users u ON u.id = sa.author_id
		WHERE sa.session_id = $1
		ORDER BY sa.created_at ASC
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []AnnotationOut
	for rows.Next() {
		var a AnnotationOut
		if err := rows.Scan(&a.ID, &a.SessionID, &a.AuthorID, &a.AuthorUsername, &a.EndNumber, &a.ArrowNumber, &a.Text, &a.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	if out == nil {
		out = []AnnotationOut{}
	}
	return out, nil
}

// ── Helpers ─────────────────────────────────────────────────────────────

func (r *CoachingRepo) getLink(ctx context.Context, id string) (*CoachAthleteLinkOut, error) {
	var link CoachAthleteLinkOut
	err := r.DB.QueryRow(ctx, `
		SELECT cal.id, cal.coach_id, cal.athlete_id,
		       uc.username AS coach_username,
		       ua.username AS athlete_username,
		       cal.status, cal.created_at
		FROM coach_athlete_links cal
		LEFT JOIN users uc ON uc.id = cal.coach_id
		LEFT JOIN users ua ON ua.id = cal.athlete_id
		WHERE cal.id = $1
	`, id).Scan(&link.ID, &link.CoachID, &link.AthleteID, &link.CoachUsername, &link.AthleteUsername, &link.Status, &link.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &link, nil
}

func (r *CoachingRepo) listLinks(ctx context.Context, column, userID string) ([]CoachAthleteLinkOut, error) {
	query := `
		SELECT cal.id, cal.coach_id, cal.athlete_id,
		       uc.username AS coach_username,
		       ua.username AS athlete_username,
		       cal.status, cal.created_at
		FROM coach_athlete_links cal
		LEFT JOIN users uc ON uc.id = cal.coach_id
		LEFT JOIN users ua ON ua.id = cal.athlete_id
		WHERE cal.` + column + ` = $1
		ORDER BY cal.created_at DESC
	`
	rows, err := r.DB.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []CoachAthleteLinkOut
	for rows.Next() {
		var link CoachAthleteLinkOut
		if err := rows.Scan(&link.ID, &link.CoachID, &link.AthleteID, &link.CoachUsername, &link.AthleteUsername, &link.Status, &link.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, link)
	}
	if out == nil {
		out = []CoachAthleteLinkOut{}
	}
	return out, nil
}

func (r *CoachingRepo) getAnnotation(ctx context.Context, id string) (*AnnotationOut, error) {
	var a AnnotationOut
	err := r.DB.QueryRow(ctx, `
		SELECT sa.id, sa.session_id, sa.author_id, u.username, sa.end_number, sa.arrow_number, sa.text, sa.created_at
		FROM session_annotations sa
		LEFT JOIN users u ON u.id = sa.author_id
		WHERE sa.id = $1
	`, id).Scan(&a.ID, &a.SessionID, &a.AuthorID, &a.AuthorUsername, &a.EndNumber, &a.ArrowNumber, &a.Text, &a.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &a, nil
}
