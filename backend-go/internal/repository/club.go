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

type ClubSharedRoundOut struct {
	ID               string    `json:"id"`
	ClubID           string    `json:"club_id"`
	ClubName         string    `json:"club_name"`
	TemplateID       string    `json:"template_id"`
	TemplateName     string    `json:"template_name"`
	SharedByUsername string    `json:"shared_by_username"`
	SharedAt         time.Time `json:"shared_at"`
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
		"DELETE FROM tournament_round_scores WHERE round_id IN (SELECT id FROM tournament_rounds WHERE tournament_id IN (SELECT id FROM tournaments WHERE club_id = $1))",
		"DELETE FROM tournament_rounds WHERE tournament_id IN (SELECT id FROM tournaments WHERE club_id = $1)",
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
