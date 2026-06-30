package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ChallengesRepo struct {
	DB *pgxpool.Pool
}

type ChallengeOut struct {
	ID                  string     `json:"id"`
	ChallengerID        string     `json:"challenger_id"`
	ChallengerUsername  string     `json:"challenger_username"`
	ChallengeeID        string     `json:"challengee_id"`
	ChallengeeUsername  string     `json:"challengee_username"`
	TemplateID          string     `json:"template_id"`
	TemplateName        string     `json:"template_name"`
	ChallengerSessionID *string    `json:"challenger_session_id"`
	ChallengerScore     *int       `json:"challenger_score"`
	ChallengeeSessionID *string    `json:"challengee_session_id"`
	ChallengeeScore     *int       `json:"challengee_score"`
	Status              string     `json:"status"`
	CreatedAt           time.Time  `json:"created_at"`
	ExpiresAt           *time.Time `json:"expires_at"`
}

func (r *ChallengesRepo) CreateChallenge(ctx context.Context, challengerID, challengeeID, templateID string, expiresAt *time.Time) (*ChallengeOut, error) {
	id := uuid.New().String()
	now := time.Now()

	_, err := r.DB.Exec(ctx, `
		INSERT INTO challenges (id, challenger_id, challengee_id, template_id, status, created_at, updated_at, expires_at)
		VALUES ($1, $2, $3, $4, 'pending', $5, $5, $6)
	`, id, challengerID, challengeeID, templateID, now, expiresAt)
	if err != nil {
		return nil, err
	}
	return r.GetChallenge(ctx, id)
}

func (r *ChallengesRepo) GetChallenge(ctx context.Context, challengeID string) (*ChallengeOut, error) {
	var out ChallengeOut
	err := r.DB.QueryRow(ctx, `
		SELECT c.id, c.challenger_id, COALESCE(u1.username, ''),
		       c.challengee_id, COALESCE(u2.username, ''),
		       c.template_id, COALESCE(t.name, ''),
		       c.challenger_session_id, s1.total_score,
		       c.challengee_session_id, s2.total_score,
		       c.status, c.created_at, c.expires_at
		FROM challenges c
		LEFT JOIN users u1 ON u1.id = c.challenger_id
		LEFT JOIN users u2 ON u2.id = c.challengee_id
		LEFT JOIN round_templates t ON t.id = c.template_id
		LEFT JOIN scoring_sessions s1 ON s1.id = c.challenger_session_id
		LEFT JOIN scoring_sessions s2 ON s2.id = c.challengee_session_id
		WHERE c.id = $1
	`, challengeID).Scan(
		&out.ID, &out.ChallengerID, &out.ChallengerUsername,
		&out.ChallengeeID, &out.ChallengeeUsername,
		&out.TemplateID, &out.TemplateName,
		&out.ChallengerSessionID, &out.ChallengerScore,
		&out.ChallengeeSessionID, &out.ChallengeeScore,
		&out.Status, &out.CreatedAt, &out.ExpiresAt,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *ChallengesRepo) ListChallengesForUser(ctx context.Context, userID string) ([]ChallengeOut, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT c.id, c.challenger_id, COALESCE(u1.username, ''),
		       c.challengee_id, COALESCE(u2.username, ''),
		       c.template_id, COALESCE(t.name, ''),
		       c.challenger_session_id, s1.total_score,
		       c.challengee_session_id, s2.total_score,
		       c.status, c.created_at, c.expires_at
		FROM challenges c
		LEFT JOIN users u1 ON u1.id = c.challenger_id
		LEFT JOIN users u2 ON u2.id = c.challengee_id
		LEFT JOIN round_templates t ON t.id = c.template_id
		LEFT JOIN scoring_sessions s1 ON s1.id = c.challenger_session_id
		LEFT JOIN scoring_sessions s2 ON s2.id = c.challengee_session_id
		WHERE c.challenger_id = $1 OR c.challengee_id = $1
		ORDER BY c.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []ChallengeOut
	for rows.Next() {
		var out ChallengeOut
		err = rows.Scan(
			&out.ID, &out.ChallengerID, &out.ChallengerUsername,
			&out.ChallengeeID, &out.ChallengeeUsername,
			&out.TemplateID, &out.TemplateName,
			&out.ChallengerSessionID, &out.ChallengerScore,
			&out.ChallengeeSessionID, &out.ChallengeeScore,
			&out.Status, &out.CreatedAt, &out.ExpiresAt,
		)
		if err != nil {
			return nil, err
		}
		list = append(list, out)
	}
	return list, nil
}

func (r *ChallengesRepo) AcceptChallenge(ctx context.Context, challengeID, userID string) (*ChallengeOut, error) {
	tag, err := r.DB.Exec(ctx, `
		UPDATE challenges
		SET status = 'accepted', updated_at = NOW()
		WHERE id = $1 AND challengee_id = $2 AND status = 'pending'
	`, challengeID, userID)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, errors.New("cannot accept challenge: not found, unauthorized, or not pending")
	}
	return r.GetChallenge(ctx, challengeID)
}

func (r *ChallengesRepo) DeclineChallenge(ctx context.Context, challengeID, userID string) (*ChallengeOut, error) {
	tag, err := r.DB.Exec(ctx, `
		UPDATE challenges
		SET status = 'declined', updated_at = NOW()
		WHERE id = $1 AND challengee_id = $2 AND status = 'pending'
	`, challengeID, userID)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, errors.New("cannot decline challenge: not found, unauthorized, or not pending")
	}
	return r.GetChallenge(ctx, challengeID)
}

func (r *ChallengesRepo) SubmitChallengeScore(ctx context.Context, challengeID, userID, sessionID string) (*ChallengeOut, error) {
	// 1. Verify session belongs to the user and is completed
	var sessionUserID string
	var sessionTemplateID string
	var sessionStatus string
	err := r.DB.QueryRow(ctx, `
		SELECT user_id, template_id, status FROM scoring_sessions WHERE id = $1
	`, sessionID).Scan(&sessionUserID, &sessionTemplateID, &sessionStatus)
	if err == pgx.ErrNoRows {
		return nil, errors.New("session not found")
	}
	if err != nil {
		return nil, err
	}
	if sessionUserID != userID {
		return nil, errors.New("session does not belong to user")
	}
	if sessionStatus != "completed" {
		return nil, errors.New("cannot submit in-progress session to challenge")
	}

	// 2. Fetch challenge
	chall, err := r.GetChallenge(ctx, challengeID)
	if err != nil {
		return nil, err
	}
	if chall.Status != "accepted" {
		return nil, errors.New("challenge is not accepted yet or already completed")
	}
	if chall.TemplateID != sessionTemplateID {
		return nil, errors.New("session template does not match challenge template")
	}

	// 3. Determine if user is challenger or challengee
	var column string
	if chall.ChallengerID == userID {
		column = "challenger_session_id"
	} else if chall.ChallengeeID == userID {
		column = "challengee_session_id"
	} else {
		return nil, errors.New("user is not participant in this challenge")
	}

	// 4. Update the challenge
	_, err = r.DB.Exec(ctx, `
		UPDATE challenges
		SET `+column+` = $1, updated_at = NOW()
		WHERE id = $2
	`, sessionID, challengeID)
	if err != nil {
		return nil, err
	}

	// 5. Reload challenge and check if completed
	chall, err = r.GetChallenge(ctx, challengeID)
	if err != nil {
		return nil, err
	}

	if chall.ChallengerSessionID != nil && chall.ChallengeeSessionID != nil {
		_, err = r.DB.Exec(ctx, `
			UPDATE challenges
			SET status = 'completed', updated_at = NOW()
			WHERE id = $1
		`, challengeID)
		if err != nil {
			return nil, err
		}
		chall, err = r.GetChallenge(ctx, challengeID)
	}

	return chall, err
}
