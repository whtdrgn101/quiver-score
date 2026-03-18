package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	DB *pgxpool.Pool
}

// ── Types ─────────────────────────────────────────────────────────────

type UserOut struct {
	ID             string    `json:"id"`
	Email          string    `json:"email"`
	Username       string    `json:"username"`
	DisplayName    *string   `json:"display_name"`
	BowType        *string   `json:"bow_type"`
	Classification *string   `json:"classification"`
	Bio            *string   `json:"bio"`
	Avatar         *string   `json:"avatar"`
	EmailVerified  bool      `json:"email_verified"`
	ProfilePublic  bool      `json:"profile_public"`
	CreatedAt      time.Time `json:"created_at"`
}

// ── Methods ───────────────────────────────────────────────────────────

func (r *UserRepo) ExistsByEmailOrUsername(ctx context.Context, email, username string) (bool, error) {
	var exists bool
	err := r.DB.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 OR username = $2)",
		email, username,
	).Scan(&exists)
	return exists, err
}

func (r *UserRepo) Create(ctx context.Context, id, email, username, hashedPw, displayName, verificationToken string) error {
	_, err := r.DB.Exec(ctx,
		`INSERT INTO users (id, email, username, hashed_password, display_name, email_verification_token)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		id, email, username, hashedPw, displayName, verificationToken,
	)
	return err
}

func (r *UserRepo) GetCredentialsByUsername(ctx context.Context, username string) (userID, hashedPw string, err error) {
	err = r.DB.QueryRow(ctx,
		"SELECT id, hashed_password FROM users WHERE username = $1", username,
	).Scan(&userID, &hashedPw)
	return
}

func (r *UserRepo) Exists(ctx context.Context, id string) (bool, error) {
	var exists bool
	err := r.DB.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", id,
	).Scan(&exists)
	return exists, err
}

func (r *UserRepo) VerifyEmail(ctx context.Context, email string) (bool, error) {
	tag, err := r.DB.Exec(ctx,
		`UPDATE users SET email_verified = true, email_verification_token = NULL
		 WHERE email = $1`, email,
	)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func (r *UserRepo) GetEmailInfo(ctx context.Context, userID string) (email string, verified bool, err error) {
	err = r.DB.QueryRow(ctx,
		"SELECT email, email_verified FROM users WHERE id = $1", userID,
	).Scan(&email, &verified)
	return
}

func (r *UserRepo) UpdateVerificationToken(ctx context.Context, userID, token string) error {
	_, err := r.DB.Exec(ctx,
		"UPDATE users SET email_verification_token = $1 WHERE id = $2",
		token, userID,
	)
	return err
}

func (r *UserRepo) GetHashedPassword(ctx context.Context, userID string) (string, error) {
	var hashedPw string
	err := r.DB.QueryRow(ctx,
		"SELECT hashed_password FROM users WHERE id = $1", userID,
	).Scan(&hashedPw)
	return hashedPw, err
}

func (r *UserRepo) UpdatePassword(ctx context.Context, userID, hashedPw string) error {
	_, err := r.DB.Exec(ctx,
		"UPDATE users SET hashed_password = $1 WHERE id = $2",
		hashedPw, userID,
	)
	return err
}

func (r *UserRepo) ResetPasswordByEmail(ctx context.Context, email, hashedPw string) (bool, error) {
	tag, err := r.DB.Exec(ctx,
		"UPDATE users SET hashed_password = $1 WHERE email = $2",
		hashedPw, email,
	)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func (r *UserRepo) GetMe(ctx context.Context, userID string) (*UserOut, error) {
	var u UserOut
	err := r.DB.QueryRow(ctx,
		`SELECT id, email, username, display_name, bow_type, classification,
		        bio, avatar, email_verified, profile_public, created_at
		 FROM users WHERE id = $1`, userID,
	).Scan(
		&u.ID, &u.Email, &u.Username, &u.DisplayName, &u.BowType, &u.Classification,
		&u.Bio, &u.Avatar, &u.EmailVerified, &u.ProfilePublic, &u.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) GetArcherInfo(ctx context.Context, userID string) (username string, displayName, avatar *string, err error) {
	err = r.DB.QueryRow(ctx,
		"SELECT username, display_name, avatar FROM users WHERE id = $1", userID,
	).Scan(&username, &displayName, &avatar)
	return
}

// DeleteUserData removes all data associated with a user, mirroring the Python cascade.
func (r *UserRepo) DeleteUserData(ctx context.Context, userID string) error {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := deleteUserDataTx(ctx, tx, userID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func deleteUserDataTx(ctx context.Context, tx pgx.Tx, userID string) error {
	// 1. Scoring data: arrows → ends → sessions
	sessionIDs, err := collectIDs(ctx, tx, "SELECT id FROM scoring_sessions WHERE user_id = $1", userID)
	if err != nil {
		return err
	}
	if len(sessionIDs) > 0 {
		endIDs, err := collectIDs(ctx, tx, "SELECT id FROM ends WHERE session_id = ANY($1)", sessionIDs)
		if err != nil {
			return err
		}
		if len(endIDs) > 0 {
			if _, err := tx.Exec(ctx, "DELETE FROM arrows WHERE end_id = ANY($1)", endIDs); err != nil {
				return err
			}
		}
		if _, err := tx.Exec(ctx, "DELETE FROM ends WHERE session_id = ANY($1)", sessionIDs); err != nil {
			return err
		}
	}

	// Session annotations
	if len(sessionIDs) > 0 {
		if _, err := tx.Exec(ctx,
			"DELETE FROM session_annotations WHERE author_id = $1 OR session_id = ANY($2)",
			userID, sessionIDs); err != nil {
			return err
		}
	} else {
		if _, err := tx.Exec(ctx, "DELETE FROM session_annotations WHERE author_id = $1", userID); err != nil {
			return err
		}
	}

	// Tournament participants
	if _, err := tx.Exec(ctx, "DELETE FROM tournament_participants WHERE user_id = $1", userID); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, "DELETE FROM personal_records WHERE user_id = $1", userID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, "DELETE FROM scoring_sessions WHERE user_id = $1", userID); err != nil {
		return err
	}

	// 2. Equipment and setups
	setupIDs, err := collectIDs(ctx, tx, "SELECT id FROM setup_profiles WHERE user_id = $1", userID)
	if err != nil {
		return err
	}
	if len(setupIDs) > 0 {
		if _, err := tx.Exec(ctx, "DELETE FROM setup_equipment WHERE setup_id = ANY($1)", setupIDs); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(ctx, "DELETE FROM setup_profiles WHERE user_id = $1", userID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, "DELETE FROM equipment WHERE user_id = $1", userID); err != nil {
		return err
	}

	// 3. Coaching
	if _, err := tx.Exec(ctx, "DELETE FROM coach_athlete_links WHERE coach_id = $1 OR athlete_id = $1", userID); err != nil {
		return err
	}

	// 4. Clubs owned by user — delete entire club cascade
	ownedClubIDs, err := collectIDs(ctx, tx, "SELECT id FROM clubs WHERE owner_id = $1", userID)
	if err != nil {
		return err
	}
	if len(ownedClubIDs) > 0 {
		tournamentIDs, err := collectIDs(ctx, tx, "SELECT id FROM tournaments WHERE club_id = ANY($1)", ownedClubIDs)
		if err != nil {
			return err
		}
		if len(tournamentIDs) > 0 {
			if _, err := tx.Exec(ctx, "DELETE FROM tournament_participants WHERE tournament_id = ANY($1)", tournamentIDs); err != nil {
				return err
			}
		}
		if _, err := tx.Exec(ctx, "DELETE FROM tournaments WHERE club_id = ANY($1)", ownedClubIDs); err != nil {
			return err
		}

		eventIDs, err := collectIDs(ctx, tx, "SELECT id FROM club_events WHERE club_id = ANY($1)", ownedClubIDs)
		if err != nil {
			return err
		}
		if len(eventIDs) > 0 {
			if _, err := tx.Exec(ctx, "DELETE FROM club_event_participants WHERE event_id = ANY($1)", eventIDs); err != nil {
				return err
			}
		}
		if _, err := tx.Exec(ctx, "DELETE FROM club_events WHERE club_id = ANY($1)", ownedClubIDs); err != nil {
			return err
		}

		teamIDs, err := collectIDs(ctx, tx, "SELECT id FROM club_teams WHERE club_id = ANY($1)", ownedClubIDs)
		if err != nil {
			return err
		}
		if len(teamIDs) > 0 {
			if _, err := tx.Exec(ctx, "DELETE FROM club_team_members WHERE team_id = ANY($1)", teamIDs); err != nil {
				return err
			}
		}
		if _, err := tx.Exec(ctx, "DELETE FROM club_teams WHERE club_id = ANY($1)", ownedClubIDs); err != nil {
			return err
		}

		if _, err := tx.Exec(ctx, "DELETE FROM club_shared_rounds WHERE club_id = ANY($1)", ownedClubIDs); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, "DELETE FROM club_invites WHERE club_id = ANY($1)", ownedClubIDs); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, "DELETE FROM club_members WHERE club_id = ANY($1)", ownedClubIDs); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, "DELETE FROM clubs WHERE id = ANY($1)", ownedClubIDs); err != nil {
			return err
		}
	}

	// 5. Tournaments organized by user (not in owned clubs)
	userTournamentIDs, err := collectIDs(ctx, tx, "SELECT id FROM tournaments WHERE organizer_id = $1", userID)
	if err != nil {
		return err
	}
	if len(userTournamentIDs) > 0 {
		if _, err := tx.Exec(ctx, "DELETE FROM tournament_participants WHERE tournament_id = ANY($1)", userTournamentIDs); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, "DELETE FROM tournaments WHERE id = ANY($1)", userTournamentIDs); err != nil {
			return err
		}
	}

	// 6. Club participation (non-owned clubs)
	if _, err := tx.Exec(ctx, "DELETE FROM club_event_participants WHERE user_id = $1", userID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, "DELETE FROM club_team_members WHERE user_id = $1", userID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, "DELETE FROM club_shared_rounds WHERE shared_by = $1", userID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, "DELETE FROM club_invites WHERE created_by = $1", userID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, "DELETE FROM club_members WHERE user_id = $1", userID); err != nil {
		return err
	}

	// 7. Custom round templates and stages
	templateIDs, err := collectIDs(ctx, tx,
		"SELECT id FROM round_templates WHERE created_by = $1 AND is_official = false", userID)
	if err != nil {
		return err
	}
	if len(templateIDs) > 0 {
		if _, err := tx.Exec(ctx, "DELETE FROM club_shared_rounds WHERE template_id = ANY($1)", templateIDs); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, "DELETE FROM round_template_stages WHERE template_id = ANY($1)", templateIDs); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, "DELETE FROM round_templates WHERE id = ANY($1)", templateIDs); err != nil {
			return err
		}
	}

	// 8. Remaining user data
	for _, table := range []string{
		"notifications",
		"classification_records",
		"sight_marks",
		"feed_items",
	} {
		if _, err := tx.Exec(ctx, "DELETE FROM "+table+" WHERE user_id = $1", userID); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(ctx, "DELETE FROM follows WHERE follower_id = $1 OR following_id = $1", userID); err != nil {
		return err
	}

	// 9. Delete user
	if _, err := tx.Exec(ctx, "DELETE FROM users WHERE id = $1", userID); err != nil {
		return err
	}

	return nil
}

func collectIDs(ctx context.Context, tx pgx.Tx, query string, args ...any) ([]string, error) {
	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
