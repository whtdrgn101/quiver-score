package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

// ── Event Types ──────────────────────────────────────────────────────

type EventParticipantOut struct {
	UserID      string  `json:"user_id"`
	Username    string  `json:"username"`
	DisplayName *string `json:"display_name"`
	Avatar      *string `json:"avatar"`
	Status      string  `json:"status"`
	Score       *int    `json:"score"`
	XCount      *int    `json:"x_count"`
	SessionID   *string `json:"session_id"`
}

type EventOut struct {
	ID           string                `json:"id"`
	ClubID       string                `json:"club_id"`
	Name         string                `json:"name"`
	Description  *string               `json:"description"`
	TemplateID   string                `json:"template_id"`
	TemplateName *string               `json:"template_name"`
	EventDate    time.Time             `json:"event_date"`
	Location     *string               `json:"location"`
	CreatedBy    string                `json:"created_by"`
	Participants []EventParticipantOut `json:"participants"`
	CreatedAt    time.Time             `json:"created_at"`
}

// ── Events ────────────────────────────────────────────────────────────

func (r *ClubRepo) CreateEvent(ctx context.Context, id, clubID, userID, name string, description *string, templateID string, eventDate time.Time, location *string) (*EventOut, error) {
	role, err := r.getMemberRole(ctx, clubID, userID)
	if err != nil || (role != "owner" && role != "admin") {
		return nil, pgx.ErrNoRows
	}

	now := time.Now().UTC()
	_, err = r.DB.Exec(ctx,
		`INSERT INTO club_events (id, club_id, name, description, template_id, event_date, location, created_by, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		id, clubID, name, description, templateID, eventDate, location, userID, now,
	)
	if err != nil {
		return nil, err
	}

	return r.getEventOut(ctx, id, clubID)
}

func (r *ClubRepo) ListEvents(ctx context.Context, clubID, userID string) ([]EventOut, error) {
	_, err := r.getMemberRole(ctx, clubID, userID)
	if err != nil {
		return nil, err
	}

	rows, err := r.DB.Query(ctx,
		"SELECT id FROM club_events WHERE club_id = $1 ORDER BY event_date DESC", clubID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []EventOut
	for rows.Next() {
		var eventID string
		if err := rows.Scan(&eventID); err != nil {
			return nil, err
		}
		e, err := r.getEventOut(ctx, eventID, clubID)
		if err != nil {
			return nil, err
		}
		events = append(events, *e)
	}
	if events == nil {
		events = []EventOut{}
	}
	return events, rows.Err()
}

func (r *ClubRepo) GetEvent(ctx context.Context, clubID, eventID, userID string) (*EventOut, error) {
	_, err := r.getMemberRole(ctx, clubID, userID)
	if err != nil {
		return nil, err
	}
	return r.getEventOut(ctx, eventID, clubID)
}

func (r *ClubRepo) UpdateEvent(ctx context.Context, clubID, eventID, userID string, name *string, description *string, descriptionSet bool, eventDate *time.Time, location *string, locationSet bool) (*EventOut, error) {
	role, err := r.getMemberRole(ctx, clubID, userID)
	if err != nil || (role != "owner" && role != "admin") {
		return nil, pgx.ErrNoRows
	}

	tag, err := r.DB.Exec(ctx,
		`UPDATE club_events SET
			name = COALESCE($3, name),
			description = CASE WHEN $4 THEN $5 ELSE description END,
			event_date = COALESCE($6, event_date),
			location = CASE WHEN $7 THEN $8 ELSE location END
		 WHERE id = $1 AND club_id = $2`,
		eventID, clubID, name, descriptionSet, description, eventDate, locationSet, location,
	)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, pgx.ErrNoRows
	}

	return r.getEventOut(ctx, eventID, clubID)
}

func (r *ClubRepo) DeleteEvent(ctx context.Context, clubID, eventID, userID string) error {
	role, err := r.getMemberRole(ctx, clubID, userID)
	if err != nil || (role != "owner" && role != "admin") {
		return pgx.ErrNoRows
	}

	// Delete participants first
	r.DB.Exec(ctx, "DELETE FROM club_event_participants WHERE event_id = $1", eventID)

	tag, err := r.DB.Exec(ctx, "DELETE FROM club_events WHERE id = $1 AND club_id = $2", eventID, clubID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *ClubRepo) RSVPEvent(ctx context.Context, clubID, eventID, userID, status string) (*EventOut, error) {
	_, err := r.getMemberRole(ctx, clubID, userID)
	if err != nil {
		return nil, err
	}

	// Verify event exists
	var exists bool
	r.DB.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM club_events WHERE id = $1 AND club_id = $2)", eventID, clubID).Scan(&exists)
	if !exists {
		return nil, pgx.ErrNoRows
	}

	now := time.Now().UTC()
	// Upsert participant
	_, err = r.DB.Exec(ctx,
		`INSERT INTO club_event_participants (id, event_id, user_id, status, rsvp_at)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (event_id, user_id) DO UPDATE SET status = $4, rsvp_at = $5`,
		generateID(), eventID, userID, status, now,
	)
	if err != nil {
		return nil, err
	}

	return r.getEventOut(ctx, eventID, clubID)
}

func (r *ClubRepo) getEventOut(ctx context.Context, eventID, clubID string) (*EventOut, error) {
	var e EventOut
	err := r.DB.QueryRow(ctx,
		`SELECT ce.id, ce.club_id, ce.name, ce.description, ce.template_id,
		        rt.name, ce.event_date, ce.location, ce.created_by, ce.created_at
		 FROM club_events ce
		 LEFT JOIN round_templates rt ON rt.id = ce.template_id
		 WHERE ce.id = $1 AND ce.club_id = $2`, eventID, clubID,
	).Scan(&e.ID, &e.ClubID, &e.Name, &e.Description, &e.TemplateID,
		&e.TemplateName, &e.EventDate, &e.Location, &e.CreatedBy, &e.CreatedAt)
	if err != nil {
		return nil, err
	}

	// Load participants
	pRows, err := r.DB.Query(ctx,
		`SELECT cep.user_id, u.username, u.display_name, u.avatar, cep.status
		 FROM club_event_participants cep
		 JOIN users u ON u.id = cep.user_id
		 WHERE cep.event_id = $1`, eventID,
	)
	if err != nil {
		return nil, err
	}
	defer pRows.Close()

	isPast := e.EventDate.Before(time.Now())

	for pRows.Next() {
		var p EventParticipantOut
		if err := pRows.Scan(&p.UserID, &p.Username, &p.DisplayName, &p.Avatar, &p.Status); err != nil {
			return nil, err
		}

		// For past events, load scores for "going" participants
		if isPast && p.Status == "going" {
			var score, xcount int
			var sid string
			err := r.DB.QueryRow(ctx,
				`SELECT id, total_score, total_x_count FROM scoring_sessions
				 WHERE user_id = $1 AND template_id = $2 AND status = 'completed'
				   AND DATE(completed_at) = DATE($3)
				 ORDER BY total_score DESC LIMIT 1`,
				p.UserID, e.TemplateID, e.EventDate,
			).Scan(&sid, &score, &xcount)
			if err == nil {
				p.Score = &score
				p.XCount = &xcount
				p.SessionID = &sid
			}
		}

		e.Participants = append(e.Participants, p)
	}
	if e.Participants == nil {
		e.Participants = []EventParticipantOut{}
	}

	return &e, nil
}
