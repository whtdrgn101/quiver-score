package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ClubRepo struct {
	DB *pgxpool.Pool
}

// ── Types ─────────────────────────────────────────────────────────────

type ClubOut struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	Avatar      *string   `json:"avatar"`
	OwnerID     string    `json:"owner_id"`
	MemberCount int       `json:"member_count"`
	MyRole      *string   `json:"my_role"`
	CreatedAt   time.Time `json:"created_at"`
}

type ClubMemberOut struct {
	UserID      string    `json:"user_id"`
	Username    string    `json:"username"`
	DisplayName *string   `json:"display_name"`
	Avatar      *string   `json:"avatar"`
	Role        string    `json:"role"`
	JoinedAt    time.Time `json:"joined_at"`
}

type ClubDetailOut struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description *string         `json:"description"`
	Avatar      *string         `json:"avatar"`
	OwnerID     string          `json:"owner_id"`
	MemberCount int             `json:"member_count"`
	MyRole      *string         `json:"my_role"`
	CreatedAt   time.Time       `json:"created_at"`
	Members     []ClubMemberOut `json:"members"`
}

type InviteOut struct {
	ID        string     `json:"id"`
	Code      string     `json:"code"`
	URL       string     `json:"url"`
	MaxUses   *int       `json:"max_uses"`
	UseCount  int        `json:"use_count"`
	ExpiresAt *time.Time `json:"expires_at"`
	Active    bool       `json:"active"`
	CreatedAt time.Time  `json:"created_at"`
}

type JoinResult struct {
	ClubID   string `json:"club_id"`
	ClubName string `json:"club_name"`
	Role     string `json:"role"`
}

type LeaderboardEntry struct {
	UserID      string    `json:"user_id"`
	Username    string    `json:"username"`
	DisplayName *string   `json:"display_name"`
	Avatar      *string   `json:"avatar"`
	BestScore   int       `json:"best_score"`
	BestXCount  int       `json:"best_x_count"`
	SessionID   string    `json:"session_id"`
	AchievedAt  time.Time `json:"achieved_at"`
}

type LeaderboardOut struct {
	TemplateID   string             `json:"template_id"`
	TemplateName string             `json:"template_name"`
	Entries      []LeaderboardEntry `json:"entries"`
}

type ActivityItem struct {
	Type         string    `json:"type"`
	UserID       string    `json:"user_id"`
	Username     string    `json:"username"`
	DisplayName  *string   `json:"display_name"`
	Avatar       *string   `json:"avatar"`
	TemplateName string    `json:"template_name"`
	Score        int       `json:"score"`
	XCount       int       `json:"x_count"`
	SessionID    string    `json:"session_id"`
	OccurredAt   time.Time `json:"occurred_at"`
}

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

type ClubSharedRoundOut struct {
	ID               string    `json:"id"`
	ClubID           string    `json:"club_id"`
	ClubName         string    `json:"club_name"`
	TemplateID       string    `json:"template_id"`
	TemplateName     string    `json:"template_name"`
	SharedByUsername string    `json:"shared_by_username"`
	SharedAt         time.Time `json:"shared_at"`
}

type TournamentOut struct {
	ID                   string    `json:"id"`
	Name                 string    `json:"name"`
	Description          *string   `json:"description"`
	OrganizerID          string    `json:"organizer_id"`
	OrganizerName        *string   `json:"organizer_name"`
	TemplateID           string    `json:"template_id"`
	TemplateName         *string   `json:"template_name"`
	Status               string    `json:"status"`
	MaxParticipants      *int      `json:"max_participants"`
	RegistrationDeadline time.Time `json:"registration_deadline"`
	StartDate            time.Time `json:"start_date"`
	EndDate              time.Time `json:"end_date"`
	ParticipantCount     int       `json:"participant_count"`
	ClubID               string    `json:"club_id"`
	ClubName             string    `json:"club_name"`
	CreatedAt            time.Time `json:"created_at"`
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

// ── Club CRUD ─────────────────────────────────────────────────────────

func (r *ClubRepo) Create(ctx context.Context, id, name string, description *string, ownerID string) (*ClubOut, error) {
	now := time.Now().UTC()
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		"INSERT INTO clubs (id, name, description, owner_id, created_at) VALUES ($1, $2, $3, $4, $5)",
		id, name, description, ownerID, now,
	)
	if err != nil {
		return nil, err
	}

	memberID := generateID()
	_, err = tx.Exec(ctx,
		"INSERT INTO club_members (id, club_id, user_id, role, joined_at) VALUES ($1, $2, $3, 'owner', $4)",
		memberID, id, ownerID, now,
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	role := "owner"
	return &ClubOut{
		ID: id, Name: name, Description: description, OwnerID: ownerID,
		MemberCount: 1, MyRole: &role, CreatedAt: now,
	}, nil
}

func (r *ClubRepo) ListForUser(ctx context.Context, userID string) ([]ClubOut, error) {
	rows, err := r.DB.Query(ctx,
		`SELECT c.id, c.name, c.description, c.avatar, c.owner_id, c.created_at, cm.role,
		        (SELECT COUNT(*) FROM club_members WHERE club_id = c.id)
		 FROM club_members cm
		 JOIN clubs c ON c.id = cm.club_id
		 WHERE cm.user_id = $1
		 ORDER BY c.created_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clubs []ClubOut
	for rows.Next() {
		var c ClubOut
		var role string
		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.Avatar, &c.OwnerID, &c.CreatedAt, &role, &c.MemberCount); err != nil {
			return nil, err
		}
		c.MyRole = &role
		clubs = append(clubs, c)
	}
	if clubs == nil {
		clubs = []ClubOut{}
	}
	return clubs, rows.Err()
}

func (r *ClubRepo) GetDetail(ctx context.Context, clubID, userID string) (*ClubDetailOut, error) {
	myRole, err := r.getMemberRole(ctx, clubID, userID)
	if err != nil {
		return nil, err
	}

	var d ClubDetailOut
	err = r.DB.QueryRow(ctx,
		`SELECT id, name, description, avatar, owner_id, created_at,
		        (SELECT COUNT(*) FROM club_members WHERE club_id = $1)
		 FROM clubs WHERE id = $1`, clubID,
	).Scan(&d.ID, &d.Name, &d.Description, &d.Avatar, &d.OwnerID, &d.CreatedAt, &d.MemberCount)
	if err != nil {
		return nil, err
	}
	d.MyRole = &myRole

	rows, err := r.DB.Query(ctx,
		`SELECT cm.user_id, u.username, u.display_name, u.avatar, cm.role, cm.joined_at
		 FROM club_members cm
		 JOIN users u ON u.id = cm.user_id
		 WHERE cm.club_id = $1
		 ORDER BY cm.joined_at`, clubID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var m ClubMemberOut
		if err := rows.Scan(&m.UserID, &m.Username, &m.DisplayName, &m.Avatar, &m.Role, &m.JoinedAt); err != nil {
			return nil, err
		}
		d.Members = append(d.Members, m)
	}
	if d.Members == nil {
		d.Members = []ClubMemberOut{}
	}
	return &d, rows.Err()
}

func (r *ClubRepo) Update(ctx context.Context, clubID, userID string, name, description *string) (*ClubOut, error) {
	role, err := r.getMemberRole(ctx, clubID, userID)
	if err != nil || role != "owner" {
		return nil, pgx.ErrNoRows
	}

	_, err = r.DB.Exec(ctx,
		`UPDATE clubs SET
			name = COALESCE($2, name),
			description = COALESCE($3, description)
		 WHERE id = $1`, clubID, name, description,
	)
	if err != nil {
		return nil, err
	}

	var c ClubOut
	err = r.DB.QueryRow(ctx,
		`SELECT id, name, description, avatar, owner_id, created_at,
		        (SELECT COUNT(*) FROM club_members WHERE club_id = $1)
		 FROM clubs WHERE id = $1`, clubID,
	).Scan(&c.ID, &c.Name, &c.Description, &c.Avatar, &c.OwnerID, &c.CreatedAt, &c.MemberCount)
	if err != nil {
		return nil, err
	}
	c.MyRole = &role
	return &c, nil
}

func (r *ClubRepo) Delete(ctx context.Context, clubID, userID string) error {
	role, err := r.getMemberRole(ctx, clubID, userID)
	if err != nil || role != "owner" {
		return pgx.ErrNoRows
	}

	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Delete cascade
	for _, q := range []string{
		"DELETE FROM tournament_participants WHERE tournament_id IN (SELECT id FROM tournaments WHERE club_id = $1)",
		"DELETE FROM tournaments WHERE club_id = $1",
		"DELETE FROM club_event_participants WHERE event_id IN (SELECT id FROM club_events WHERE club_id = $1)",
		"DELETE FROM club_events WHERE club_id = $1",
		"DELETE FROM club_team_members WHERE team_id IN (SELECT id FROM club_teams WHERE club_id = $1)",
		"DELETE FROM club_teams WHERE club_id = $1",
		"DELETE FROM club_shared_rounds WHERE club_id = $1",
		"DELETE FROM club_invites WHERE club_id = $1",
		"DELETE FROM club_members WHERE club_id = $1",
		"DELETE FROM clubs WHERE id = $1",
	} {
		if _, err := tx.Exec(ctx, q, clubID); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// ── Invites ───────────────────────────────────────────────────────────

func (r *ClubRepo) CreateInvite(ctx context.Context, id, clubID, userID, code string, maxUses *int, expiresAt *time.Time, frontendURL string) (*InviteOut, error) {
	role, err := r.getMemberRole(ctx, clubID, userID)
	if err != nil || (role != "owner" && role != "admin") {
		return nil, pgx.ErrNoRows
	}

	now := time.Now().UTC()
	_, err = r.DB.Exec(ctx,
		`INSERT INTO club_invites (id, club_id, code, max_uses, use_count, expires_at, active, created_by, created_at)
		 VALUES ($1, $2, $3, $4, 0, $5, true, $6, $7)`,
		id, clubID, code, maxUses, expiresAt, userID, now,
	)
	if err != nil {
		return nil, err
	}

	return &InviteOut{
		ID: id, Code: code, URL: frontendURL + "/clubs/join/" + code,
		MaxUses: maxUses, UseCount: 0, ExpiresAt: expiresAt,
		Active: true, CreatedAt: now,
	}, nil
}

func (r *ClubRepo) ListInvites(ctx context.Context, clubID, userID string, frontendURL string) ([]InviteOut, error) {
	role, err := r.getMemberRole(ctx, clubID, userID)
	if err != nil || (role != "owner" && role != "admin") {
		return nil, pgx.ErrNoRows
	}

	rows, err := r.DB.Query(ctx,
		`SELECT id, code, max_uses, use_count, expires_at, active, created_at
		 FROM club_invites WHERE club_id = $1 AND active = true
		 ORDER BY created_at DESC`, clubID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invites []InviteOut
	for rows.Next() {
		var i InviteOut
		if err := rows.Scan(&i.ID, &i.Code, &i.MaxUses, &i.UseCount, &i.ExpiresAt, &i.Active, &i.CreatedAt); err != nil {
			return nil, err
		}
		i.URL = frontendURL + "/clubs/join/" + i.Code
		invites = append(invites, i)
	}
	if invites == nil {
		invites = []InviteOut{}
	}
	return invites, rows.Err()
}

func (r *ClubRepo) DeactivateInvite(ctx context.Context, clubID, inviteID, userID string) error {
	role, err := r.getMemberRole(ctx, clubID, userID)
	if err != nil || (role != "owner" && role != "admin") {
		return pgx.ErrNoRows
	}

	tag, err := r.DB.Exec(ctx,
		"UPDATE club_invites SET active = false WHERE id = $1 AND club_id = $2",
		inviteID, clubID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *ClubRepo) PreviewInvite(ctx context.Context, code, userID string) (*ClubOut, error) {
	var clubID string
	var active bool
	var maxUses *int
	var useCount int
	var expiresAt *time.Time

	err := r.DB.QueryRow(ctx,
		"SELECT club_id, active, max_uses, use_count, expires_at FROM club_invites WHERE code = $1",
		code,
	).Scan(&clubID, &active, &maxUses, &useCount, &expiresAt)
	if err != nil {
		return nil, err
	}

	if !active {
		return nil, pgx.ErrNoRows
	}
	if expiresAt != nil && time.Now().After(*expiresAt) {
		return nil, pgx.ErrNoRows
	}
	if maxUses != nil && useCount >= *maxUses {
		return nil, pgx.ErrNoRows
	}

	var c ClubOut
	err = r.DB.QueryRow(ctx,
		`SELECT id, name, description, avatar, owner_id, created_at,
		        (SELECT COUNT(*) FROM club_members WHERE club_id = clubs.id)
		 FROM clubs WHERE id = $1`, clubID,
	).Scan(&c.ID, &c.Name, &c.Description, &c.Avatar, &c.OwnerID, &c.CreatedAt, &c.MemberCount)
	if err != nil {
		return nil, err
	}

	// Check if user is already a member
	var myRole *string
	var role string
	err = r.DB.QueryRow(ctx,
		"SELECT role FROM club_members WHERE club_id = $1 AND user_id = $2",
		clubID, userID,
	).Scan(&role)
	if err == nil {
		myRole = &role
	}
	c.MyRole = myRole

	return &c, nil
}

// ErrAlreadyMember is returned when user tries to join a club they're already in.
var ErrAlreadyMember = pgx.ErrTooManyRows // reuse a sentinel

func (r *ClubRepo) JoinViaInvite(ctx context.Context, code, userID string) (*JoinResult, error) {
	var clubID string
	var active bool
	var maxUses *int
	var useCount int
	var expiresAt *time.Time

	err := r.DB.QueryRow(ctx,
		"SELECT club_id, active, max_uses, use_count, expires_at FROM club_invites WHERE code = $1",
		code,
	).Scan(&clubID, &active, &maxUses, &useCount, &expiresAt)
	if err != nil {
		return nil, err
	}

	if !active || (expiresAt != nil && time.Now().After(*expiresAt)) || (maxUses != nil && useCount >= *maxUses) {
		return nil, pgx.ErrNoRows
	}

	// Check already member
	var existing string
	err = r.DB.QueryRow(ctx, "SELECT role FROM club_members WHERE club_id = $1 AND user_id = $2", clubID, userID).Scan(&existing)
	if err == nil {
		return nil, ErrAlreadyMember
	}

	now := time.Now().UTC()
	memberID := generateID()
	_, err = r.DB.Exec(ctx,
		"INSERT INTO club_members (id, club_id, user_id, role, joined_at) VALUES ($1, $2, $3, 'member', $4)",
		memberID, clubID, userID, now,
	)
	if err != nil {
		return nil, err
	}

	// Increment use count
	r.DB.Exec(ctx, "UPDATE club_invites SET use_count = use_count + 1 WHERE code = $1", code)

	var clubName string
	r.DB.QueryRow(ctx, "SELECT name FROM clubs WHERE id = $1", clubID).Scan(&clubName)

	return &JoinResult{ClubID: clubID, ClubName: clubName, Role: "member"}, nil
}

// ── Members ───────────────────────────────────────────────────────────

func (r *ClubRepo) PromoteMember(ctx context.Context, clubID, targetUserID, userID string) error {
	role, err := r.getMemberRole(ctx, clubID, userID)
	if err != nil || role != "owner" {
		return pgx.ErrNoRows
	}

	var targetRole string
	err = r.DB.QueryRow(ctx, "SELECT role FROM club_members WHERE club_id = $1 AND user_id = $2", clubID, targetUserID).Scan(&targetRole)
	if err != nil {
		return err
	}
	if targetRole == "owner" {
		return pgx.ErrNoRows // can't promote owner
	}

	_, err = r.DB.Exec(ctx,
		"UPDATE club_members SET role = 'admin' WHERE club_id = $1 AND user_id = $2",
		clubID, targetUserID,
	)
	return err
}

func (r *ClubRepo) DemoteMember(ctx context.Context, clubID, targetUserID, userID string) error {
	role, err := r.getMemberRole(ctx, clubID, userID)
	if err != nil || role != "owner" {
		return pgx.ErrNoRows
	}

	var targetRole string
	err = r.DB.QueryRow(ctx, "SELECT role FROM club_members WHERE club_id = $1 AND user_id = $2", clubID, targetUserID).Scan(&targetRole)
	if err != nil {
		return err
	}
	if targetRole == "owner" {
		return pgx.ErrNoRows
	}

	_, err = r.DB.Exec(ctx,
		"UPDATE club_members SET role = 'member' WHERE club_id = $1 AND user_id = $2",
		clubID, targetUserID,
	)
	return err
}

func (r *ClubRepo) RemoveMember(ctx context.Context, clubID, targetUserID, userID string) error {
	// Self-removal
	if targetUserID == userID {
		var role string
		err := r.DB.QueryRow(ctx, "SELECT role FROM club_members WHERE club_id = $1 AND user_id = $2", clubID, userID).Scan(&role)
		if err != nil {
			return err
		}
		if role == "owner" {
			return pgx.ErrNoRows // owner can't leave
		}
		_, err = r.DB.Exec(ctx, "DELETE FROM club_members WHERE club_id = $1 AND user_id = $2", clubID, userID)
		return err
	}

	// Removing someone else: need owner/admin
	myRole, err := r.getMemberRole(ctx, clubID, userID)
	if err != nil || (myRole != "owner" && myRole != "admin") {
		return pgx.ErrNoRows
	}

	var targetRole string
	err = r.DB.QueryRow(ctx, "SELECT role FROM club_members WHERE club_id = $1 AND user_id = $2", clubID, targetUserID).Scan(&targetRole)
	if err != nil {
		return err
	}
	if targetRole == "owner" {
		return pgx.ErrNoRows
	}
	if targetRole == "admin" && myRole != "owner" {
		return pgx.ErrNoRows
	}

	_, err = r.DB.Exec(ctx, "DELETE FROM club_members WHERE club_id = $1 AND user_id = $2", clubID, targetUserID)
	return err
}

// ── Leaderboard ───────────────────────────────────────────────────────

func (r *ClubRepo) Leaderboard(ctx context.Context, clubID, userID string, templateID *string) ([]LeaderboardOut, error) {
	_, err := r.getMemberRole(ctx, clubID, userID)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT s.id, s.user_id, s.template_id, s.total_score, s.total_x_count,
		       COALESCE(s.completed_at, s.started_at) as achieved_at,
		       rt.name as template_name,
		       u.username, u.display_name, u.avatar
		FROM scoring_sessions s
		JOIN club_members cm ON cm.user_id = s.user_id AND cm.club_id = $1
		LEFT JOIN round_templates rt ON rt.id = s.template_id
		LEFT JOIN users u ON u.id = s.user_id
		WHERE s.status = 'completed'`
	args := []any{clubID}

	if templateID != nil {
		query += " AND s.template_id = $2"
		args = append(args, *templateID)
	}
	query += " ORDER BY s.total_score DESC"

	rows, err := r.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Group by template, keep best per user per template
	type key struct{ templateID, userID string }
	type entry struct {
		LeaderboardEntry
		templateName string
	}
	best := map[key]entry{}
	templateOrder := []string{}
	templateNames := map[string]string{}

	for rows.Next() {
		var sid, uid, tid string
		var score, xcount int
		var achievedAt time.Time
		var tname *string
		var username string
		var displayName, avatar *string

		if err := rows.Scan(&sid, &uid, &tid, &score, &xcount, &achievedAt, &tname, &username, &displayName, &avatar); err != nil {
			return nil, err
		}

		k := key{tid, uid}
		if existing, ok := best[k]; !ok || score > existing.BestScore {
			tn := ""
			if tname != nil {
				tn = *tname
			}
			if _, seen := templateNames[tid]; !seen {
				templateOrder = append(templateOrder, tid)
				templateNames[tid] = tn
			}
			best[k] = entry{
				LeaderboardEntry: LeaderboardEntry{
					UserID: uid, Username: username, DisplayName: displayName, Avatar: avatar,
					BestScore: score, BestXCount: xcount, SessionID: sid, AchievedAt: achievedAt,
				},
				templateName: tn,
			}
		}
	}

	var result []LeaderboardOut
	for _, tid := range templateOrder {
		lb := LeaderboardOut{TemplateID: tid, TemplateName: templateNames[tid]}
		for k, e := range best {
			if k.templateID == tid {
				lb.Entries = append(lb.Entries, e.LeaderboardEntry)
			}
		}
		// Sort entries by score DESC
		for i := 0; i < len(lb.Entries); i++ {
			for j := i + 1; j < len(lb.Entries); j++ {
				if lb.Entries[j].BestScore > lb.Entries[i].BestScore {
					lb.Entries[i], lb.Entries[j] = lb.Entries[j], lb.Entries[i]
				}
			}
		}
		result = append(result, lb)
	}
	if result == nil {
		result = []LeaderboardOut{}
	}
	return result, nil
}

// ── Activity ──────────────────────────────────────────────────────────

func (r *ClubRepo) Activity(ctx context.Context, clubID, userID string, limit, offset int) ([]ActivityItem, error) {
	_, err := r.getMemberRole(ctx, clubID, userID)
	if err != nil {
		return nil, err
	}

	// Sessions
	sessRows, err := r.DB.Query(ctx,
		`SELECT s.id, s.user_id, s.total_score, s.total_x_count, s.completed_at,
		        rt.name, u.username, u.display_name, u.avatar
		 FROM scoring_sessions s
		 JOIN club_members cm ON cm.user_id = s.user_id AND cm.club_id = $1
		 LEFT JOIN round_templates rt ON rt.id = s.template_id
		 LEFT JOIN users u ON u.id = s.user_id
		 WHERE s.status = 'completed'
		 ORDER BY s.completed_at DESC
		 LIMIT $2 OFFSET $3`, clubID, limit+offset, 0,
	)
	if err != nil {
		return nil, err
	}
	defer sessRows.Close()

	var items []ActivityItem
	for sessRows.Next() {
		var a ActivityItem
		var completedAt *time.Time
		var tname *string
		if err := sessRows.Scan(&a.SessionID, &a.UserID, &a.Score, &a.XCount, &completedAt, &tname, &a.Username, &a.DisplayName, &a.Avatar); err != nil {
			return nil, err
		}
		a.Type = "session_completed"
		if tname != nil {
			a.TemplateName = *tname
		}
		if completedAt != nil {
			a.OccurredAt = *completedAt
		}
		items = append(items, a)
	}

	// Personal records
	prRows, err := r.DB.Query(ctx,
		`SELECT pr.session_id, pr.user_id, pr.score, pr.achieved_at,
		        rt.name, u.username, u.display_name, u.avatar
		 FROM personal_records pr
		 JOIN club_members cm ON cm.user_id = pr.user_id AND cm.club_id = $1
		 LEFT JOIN round_templates rt ON rt.id = pr.template_id
		 LEFT JOIN users u ON u.id = pr.user_id
		 ORDER BY pr.achieved_at DESC
		 LIMIT $2 OFFSET $3`, clubID, limit+offset, 0,
	)
	if err != nil {
		return nil, err
	}
	defer prRows.Close()

	for prRows.Next() {
		var a ActivityItem
		var tname *string
		if err := prRows.Scan(&a.SessionID, &a.UserID, &a.Score, &a.OccurredAt, &tname, &a.Username, &a.DisplayName, &a.Avatar); err != nil {
			return nil, err
		}
		a.Type = "personal_record"
		if tname != nil {
			a.TemplateName = *tname
		}
		items = append(items, a)
	}

	// Sort by occurred_at DESC
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if items[j].OccurredAt.After(items[i].OccurredAt) {
				items[i], items[j] = items[j], items[i]
			}
		}
	}

	// Apply offset/limit
	if offset < len(items) {
		items = items[offset:]
	} else {
		items = nil
	}
	if limit < len(items) {
		items = items[:limit]
	}

	if items == nil {
		items = []ActivityItem{}
	}
	return items, nil
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

// ── Shared Rounds ─────────────────────────────────────────────────────

func (r *ClubRepo) ListSharedRounds(ctx context.Context, clubID, userID string) ([]ClubSharedRoundOut, error) {
	_, err := r.getMemberRole(ctx, clubID, userID)
	if err != nil {
		return nil, err
	}

	rows, err := r.DB.Query(ctx,
		`SELECT csr.id, csr.club_id, c.name, csr.template_id, rt.name, u.username, csr.shared_at
		 FROM club_shared_rounds csr
		 JOIN clubs c ON c.id = csr.club_id
		 JOIN round_templates rt ON rt.id = csr.template_id
		 JOIN users u ON u.id = csr.shared_by
		 WHERE csr.club_id = $1
		 ORDER BY csr.shared_at DESC`, clubID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rounds []ClubSharedRoundOut
	for rows.Next() {
		var r ClubSharedRoundOut
		if err := rows.Scan(&r.ID, &r.ClubID, &r.ClubName, &r.TemplateID, &r.TemplateName, &r.SharedByUsername, &r.SharedAt); err != nil {
			return nil, err
		}
		rounds = append(rounds, r)
	}
	if rounds == nil {
		rounds = []ClubSharedRoundOut{}
	}
	return rounds, rows.Err()
}

func (r *ClubRepo) RemoveSharedRound(ctx context.Context, clubID, roundID, userID string) error {
	role, err := r.getMemberRole(ctx, clubID, userID)
	if err != nil || (role != "owner" && role != "admin") {
		return pgx.ErrNoRows
	}

	tag, err := r.DB.Exec(ctx,
		"DELETE FROM club_shared_rounds WHERE id = $1 AND club_id = $2", roundID, clubID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
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

// ── Helpers ───────────────────────────────────────────────────────────

func (r *ClubRepo) getMemberRole(ctx context.Context, clubID, userID string) (string, error) {
	var role string
	err := r.DB.QueryRow(ctx,
		"SELECT role FROM club_members WHERE club_id = $1 AND user_id = $2",
		clubID, userID,
	).Scan(&role)
	return role, err
}

func generateID() string {
	return uuid.New().String()
}
