package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

// ── Team Types ───────────────────────────────────────────────────────

type TeamMemberOut struct {
	UserID      string    `json:"user_id"`
	Username    string    `json:"username"`
	DisplayName *string   `json:"display_name"`
	Avatar      *string   `json:"avatar"`
	JoinedAt    time.Time `json:"joined_at"`
}

type TeamOut struct {
	ID          string        `json:"id"`
	ClubID      string        `json:"club_id"`
	Name        string        `json:"name"`
	Description *string       `json:"description"`
	Leader      TeamMemberOut `json:"leader"`
	MemberCount int           `json:"member_count"`
	CreatedAt   time.Time     `json:"created_at"`
}

type TeamDetailOut struct {
	ID          string          `json:"id"`
	ClubID      string          `json:"club_id"`
	Name        string          `json:"name"`
	Description *string         `json:"description"`
	Leader      TeamMemberOut   `json:"leader"`
	MemberCount int             `json:"member_count"`
	CreatedAt   time.Time       `json:"created_at"`
	Members     []TeamMemberOut `json:"members"`
}

// ── Teams ─────────────────────────────────────────────────────────────

func (r *ClubRepo) CreateTeam(ctx context.Context, id, clubID, userID, name string, description *string, leaderID string) (*TeamOut, error) {
	role, err := r.getMemberRole(ctx, clubID, userID)
	if err != nil || (role != "owner" && role != "admin") {
		return nil, pgx.ErrNoRows
	}

	// Validate leader is club member
	var leaderRole string
	err = r.DB.QueryRow(ctx, "SELECT role FROM club_members WHERE club_id = $1 AND user_id = $2", clubID, leaderID).Scan(&leaderRole)
	if err != nil {
		return nil, pgx.ErrNoRows
	}

	now := time.Now().UTC()
	_, err = r.DB.Exec(ctx,
		`INSERT INTO club_teams (id, club_id, name, description, leader_id, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		id, clubID, name, description, leaderID, now,
	)
	if err != nil {
		return nil, err
	}

	return r.getTeamOut(ctx, id, clubID)
}

func (r *ClubRepo) ListTeams(ctx context.Context, clubID, userID string) ([]TeamOut, error) {
	_, err := r.getMemberRole(ctx, clubID, userID)
	if err != nil {
		return nil, err
	}

	rows, err := r.DB.Query(ctx, "SELECT id FROM club_teams WHERE club_id = $1 ORDER BY created_at", clubID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []TeamOut
	for rows.Next() {
		var teamID string
		if err := rows.Scan(&teamID); err != nil {
			return nil, err
		}
		t, err := r.getTeamOut(ctx, teamID, clubID)
		if err != nil {
			return nil, err
		}
		teams = append(teams, *t)
	}
	if teams == nil {
		teams = []TeamOut{}
	}
	return teams, rows.Err()
}

func (r *ClubRepo) GetTeamDetail(ctx context.Context, clubID, teamID, userID string) (*TeamDetailOut, error) {
	_, err := r.getMemberRole(ctx, clubID, userID)
	if err != nil {
		return nil, err
	}

	t, err := r.getTeamOut(ctx, teamID, clubID)
	if err != nil {
		return nil, err
	}

	// Load members
	rows, err := r.DB.Query(ctx,
		`SELECT ctm.user_id, u.username, u.display_name, u.avatar, ctm.joined_at
		 FROM club_team_members ctm
		 JOIN users u ON u.id = ctm.user_id
		 WHERE ctm.team_id = $1`, teamID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []TeamMemberOut
	for rows.Next() {
		var m TeamMemberOut
		if err := rows.Scan(&m.UserID, &m.Username, &m.DisplayName, &m.Avatar, &m.JoinedAt); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	if members == nil {
		members = []TeamMemberOut{}
	}

	return &TeamDetailOut{
		ID: t.ID, ClubID: t.ClubID, Name: t.Name, Description: t.Description,
		Leader: t.Leader, MemberCount: t.MemberCount, CreatedAt: t.CreatedAt,
		Members: members,
	}, nil
}

func (r *ClubRepo) UpdateTeam(ctx context.Context, clubID, teamID, userID string, name, description *string, leaderID *string) (*TeamOut, error) {
	role, err := r.getMemberRole(ctx, clubID, userID)
	if err != nil || (role != "owner" && role != "admin") {
		return nil, pgx.ErrNoRows
	}

	if leaderID != nil {
		var lr string
		err = r.DB.QueryRow(ctx, "SELECT role FROM club_members WHERE club_id = $1 AND user_id = $2", clubID, *leaderID).Scan(&lr)
		if err != nil {
			return nil, pgx.ErrNoRows
		}
	}

	tag, err := r.DB.Exec(ctx,
		`UPDATE club_teams SET
			name = COALESCE($3, name),
			description = COALESCE($4, description),
			leader_id = COALESCE($5, leader_id)
		 WHERE id = $1 AND club_id = $2`,
		teamID, clubID, name, description, leaderID,
	)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, pgx.ErrNoRows
	}

	return r.getTeamOut(ctx, teamID, clubID)
}

func (r *ClubRepo) DeleteTeam(ctx context.Context, clubID, teamID, userID string) error {
	role, err := r.getMemberRole(ctx, clubID, userID)
	if err != nil || (role != "owner" && role != "admin") {
		return pgx.ErrNoRows
	}

	r.DB.Exec(ctx, "DELETE FROM club_team_members WHERE team_id = $1", teamID)
	tag, err := r.DB.Exec(ctx, "DELETE FROM club_teams WHERE id = $1 AND club_id = $2", teamID, clubID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *ClubRepo) AddTeamMember(ctx context.Context, clubID, teamID, targetUserID, userID string) error {
	// Check permission: owner/admin or team leader
	myRole, err := r.getMemberRole(ctx, clubID, userID)
	if err != nil {
		return err
	}

	if myRole != "owner" && myRole != "admin" {
		var leaderID string
		err = r.DB.QueryRow(ctx, "SELECT leader_id FROM club_teams WHERE id = $1 AND club_id = $2", teamID, clubID).Scan(&leaderID)
		if err != nil || leaderID != userID {
			return pgx.ErrNoRows
		}
	}

	// Validate target is club member
	var tr string
	err = r.DB.QueryRow(ctx, "SELECT role FROM club_members WHERE club_id = $1 AND user_id = $2", clubID, targetUserID).Scan(&tr)
	if err != nil {
		return pgx.ErrNoRows
	}

	// Check not already on team
	var dup bool
	r.DB.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM club_team_members WHERE team_id = $1 AND user_id = $2)", teamID, targetUserID).Scan(&dup)
	if dup {
		return ErrAlreadyMember
	}

	now := time.Now().UTC()
	_, err = r.DB.Exec(ctx,
		"INSERT INTO club_team_members (id, team_id, user_id, joined_at) VALUES ($1, $2, $3, $4)",
		generateID(), teamID, targetUserID, now,
	)
	return err
}

func (r *ClubRepo) RemoveTeamMember(ctx context.Context, clubID, teamID, targetUserID, userID string) error {
	myRole, err := r.getMemberRole(ctx, clubID, userID)
	if err != nil {
		return err
	}

	if myRole != "owner" && myRole != "admin" {
		var leaderID string
		err = r.DB.QueryRow(ctx, "SELECT leader_id FROM club_teams WHERE id = $1 AND club_id = $2", teamID, clubID).Scan(&leaderID)
		if err != nil || leaderID != userID {
			return pgx.ErrNoRows
		}
	}

	tag, err := r.DB.Exec(ctx, "DELETE FROM club_team_members WHERE team_id = $1 AND user_id = $2", teamID, targetUserID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *ClubRepo) getTeamOut(ctx context.Context, teamID, clubID string) (*TeamOut, error) {
	var t TeamOut
	var leaderID string
	err := r.DB.QueryRow(ctx,
		`SELECT id, club_id, name, description, leader_id, created_at,
		        (SELECT COUNT(*) FROM club_team_members WHERE team_id = $1)
		 FROM club_teams WHERE id = $1 AND club_id = $2`, teamID, clubID,
	).Scan(&t.ID, &t.ClubID, &t.Name, &t.Description, &leaderID, &t.CreatedAt, &t.MemberCount)
	if err != nil {
		return nil, err
	}

	// Load leader info
	err = r.DB.QueryRow(ctx,
		`SELECT u.username, u.display_name, u.avatar, cm.joined_at
		 FROM users u
		 JOIN club_members cm ON cm.user_id = u.id AND cm.club_id = $2
		 WHERE u.id = $1`, leaderID, clubID,
	).Scan(&t.Leader.Username, &t.Leader.DisplayName, &t.Leader.Avatar, &t.Leader.JoinedAt)
	t.Leader.UserID = leaderID

	return &t, nil
}
