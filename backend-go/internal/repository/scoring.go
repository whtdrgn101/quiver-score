package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ScoringRepo struct {
	DB *pgxpool.Pool
}

// ── Types ─────────────────────────────────────────────────────────────

type ArrowOut struct {
	ID           string   `json:"id"`
	ArrowNumber  int      `json:"arrow_number"`
	ScoreValue   string   `json:"score_value"`
	ScoreNumeric int      `json:"score_numeric"`
	XPos         *float64 `json:"x_pos"`
	YPos         *float64 `json:"y_pos"`
}

type EndOut struct {
	ID        string     `json:"id"`
	EndNumber int        `json:"end_number"`
	EndTotal  int        `json:"end_total"`
	StageID   *string    `json:"stage_id"`
	Arrows    []ArrowOut `json:"arrows"`
	CreatedAt time.Time  `json:"created_at"`
}

type SessionOut struct {
	ID               string            `json:"id"`
	TemplateID       string            `json:"template_id"`
	SetupProfileID   *string           `json:"setup_profile_id"`
	SetupProfileName *string           `json:"setup_profile_name"`
	Template         *RoundTemplateOut `json:"template"`
	Status           string            `json:"status"`
	TotalScore       int               `json:"total_score"`
	TotalXCount      int               `json:"total_x_count"`
	TotalArrows      int               `json:"total_arrows"`
	Notes            *string           `json:"notes"`
	Location         *string           `json:"location"`
	Weather          *string           `json:"weather"`
	ShareToken       *string           `json:"share_token"`
	IsPersonalBest   bool              `json:"is_personal_best"`
	StartedAt        time.Time         `json:"started_at"`
	CompletedAt      *time.Time        `json:"completed_at"`
	Ends             []EndOut          `json:"ends"`
}

type SessionSummary struct {
	ID               string     `json:"id"`
	TemplateID       string     `json:"template_id"`
	SetupProfileID   *string    `json:"setup_profile_id"`
	SetupProfileName *string    `json:"setup_profile_name"`
	Status           string     `json:"status"`
	TotalScore       int        `json:"total_score"`
	TotalXCount      int        `json:"total_x_count"`
	TotalArrows      int        `json:"total_arrows"`
	StartedAt        time.Time  `json:"started_at"`
	CompletedAt      *time.Time `json:"completed_at"`
	TemplateName     *string    `json:"template_name"`
}

type PersonalRecordOut struct {
	TemplateName string    `json:"template_name"`
	Score        int       `json:"score"`
	MaxScore     int       `json:"max_score"`
	AchievedAt   time.Time `json:"achieved_at"`
	SessionID    string    `json:"session_id"`
}

type RoundTypeAvg struct {
	TemplateName string  `json:"template_name"`
	AvgScore     float64 `json:"avg_score"`
	Count        int     `json:"count"`
}

type RecentTrendItem struct {
	Score        int       `json:"score"`
	MaxScore     int       `json:"max_score"`
	TemplateName string    `json:"template_name"`
	Date         time.Time `json:"date"`
}

type TrendDataItem struct {
	SessionID    string    `json:"session_id"`
	TemplateName string    `json:"template_name"`
	TotalScore   int       `json:"total_score"`
	MaxScore     int       `json:"max_score"`
	Percentage   float64   `json:"percentage"`
	CompletedAt  time.Time `json:"completed_at"`
}

type StatsOut struct {
	TotalSessions     int               `json:"total_sessions"`
	CompletedSessions int               `json:"completed_sessions"`
	TotalArrows       int               `json:"total_arrows"`
	TotalXCount       int               `json:"total_x_count"`
	PersonalBestScore *int              `json:"personal_best_score"`
	PersonalBestTmpl  *string           `json:"personal_best_template"`
	AvgByRoundType    []RoundTypeAvg    `json:"avg_by_round_type"`
	RecentTrend       []RecentTrendItem `json:"recent_trend"`
	PersonalRecords   []PersonalRecordOut `json:"personal_records"`
}

// ── Session CRUD ──────────────────────────────────────────────────────

func (r *ScoringRepo) CreateSession(ctx context.Context, id, userID, templateID string, setupProfileID *string, notes, location, weather *string, now time.Time) error {
	_, err := r.DB.Exec(ctx, `
		INSERT INTO scoring_sessions (id, user_id, template_id, setup_profile_id,
			notes, location, weather, status, total_score, total_x_count, total_arrows, started_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'in_progress', 0, 0, 0, $8)`,
		id, userID, templateID, setupProfileID, notes, location, weather, now,
	)
	return err
}

func (r *ScoringRepo) SetupProfileExists(ctx context.Context, setupID, userID string) (bool, error) {
	var exists bool
	err := r.DB.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM setup_profiles WHERE id = $1 AND user_id = $2)",
		setupID, userID,
	).Scan(&exists)
	return exists, err
}

func (r *ScoringRepo) LoadSessionOut(ctx context.Context, sessionID, userID string) (*SessionOut, error) {
	var s SessionOut
	var templateID string
	err := r.DB.QueryRow(ctx, `
		SELECT ss.id, ss.template_id, ss.setup_profile_id, ss.status,
		       ss.total_score, ss.total_x_count, ss.total_arrows,
		       ss.notes, ss.location, ss.weather, ss.share_token,
		       ss.started_at, ss.completed_at,
		       sp.name AS setup_profile_name
		FROM scoring_sessions ss
		LEFT JOIN setup_profiles sp ON sp.id = ss.setup_profile_id
		WHERE ss.id = $1 AND ss.user_id = $2`,
		sessionID, userID,
	).Scan(&s.ID, &templateID, &s.SetupProfileID, &s.Status,
		&s.TotalScore, &s.TotalXCount, &s.TotalArrows,
		&s.Notes, &s.Location, &s.Weather, &s.ShareToken,
		&s.StartedAt, &s.CompletedAt,
		&s.SetupProfileName)
	if err != nil {
		return nil, err
	}
	s.TemplateID = templateID

	// Load template with stages
	var t RoundTemplateOut
	err = r.DB.QueryRow(ctx, `
		SELECT id, name, organization, description, is_official, created_by
		FROM round_templates WHERE id = $1`, templateID,
	).Scan(&t.ID, &t.Name, &t.Organization, &t.Description, &t.IsOfficial, &t.CreatedBy)
	if err == nil {
		stages, err := LoadStages(ctx, r.DB, t.ID)
		if err == nil {
			t.Stages = stages
		}
		s.Template = &t
	}

	// Load ends with arrows
	s.Ends, _ = r.LoadEnds(ctx, sessionID)

	return &s, nil
}

func (r *ScoringRepo) ListSessions(ctx context.Context, userID string, templateID, dateFrom, dateTo, search *string) ([]SessionSummary, error) {
	query := `
		SELECT ss.id, ss.template_id, ss.setup_profile_id, ss.status,
		       ss.total_score, ss.total_x_count, ss.total_arrows,
		       ss.started_at, ss.completed_at,
		       rt.name AS template_name,
		       sp.name AS setup_profile_name
		FROM scoring_sessions ss
		LEFT JOIN round_templates rt ON rt.id = ss.template_id
		LEFT JOIN setup_profiles sp ON sp.id = ss.setup_profile_id
		WHERE ss.user_id = $1`
	args := []any{userID}
	argN := 2

	if templateID != nil {
		query += fmt.Sprintf(" AND ss.template_id = $%d", argN)
		args = append(args, *templateID)
		argN++
	}
	if dateFrom != nil {
		query += fmt.Sprintf(" AND ss.started_at >= $%d::date", argN)
		args = append(args, *dateFrom)
		argN++
	}
	if dateTo != nil {
		query += fmt.Sprintf(" AND ss.started_at <= ($%d::date + interval '1 day' - interval '1 second')", argN)
		args = append(args, *dateTo)
		argN++
	}
	if search != nil {
		pattern := "%" + *search + "%"
		query += fmt.Sprintf(" AND (ss.notes ILIKE $%d OR ss.location ILIKE $%d)", argN, argN)
		args = append(args, pattern)
		argN++
	}

	query += " ORDER BY ss.started_at DESC"

	rows, err := r.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []SessionSummary{}
	for rows.Next() {
		var s SessionSummary
		if err := rows.Scan(&s.ID, &s.TemplateID, &s.SetupProfileID, &s.Status,
			&s.TotalScore, &s.TotalXCount, &s.TotalArrows,
			&s.StartedAt, &s.CompletedAt,
			&s.TemplateName, &s.SetupProfileName); err != nil {
			return nil, err
		}
		items = append(items, s)
	}
	return items, rows.Err()
}

func (r *ScoringRepo) GetSessionStatus(ctx context.Context, sessionID, userID string) (string, error) {
	var status string
	err := r.DB.QueryRow(ctx,
		"SELECT status FROM scoring_sessions WHERE id = $1 AND user_id = $2",
		sessionID, userID,
	).Scan(&status)
	return status, err
}

func (r *ScoringRepo) IsPersonalBest(ctx context.Context, userID, sessionID string) bool {
	var isPB bool
	r.DB.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM personal_records WHERE user_id = $1 AND session_id = $2)",
		userID, sessionID,
	).Scan(&isPB)
	return isPB
}

// ── End Submission ────────────────────────────────────────────────────

type StageInfo struct {
	ArrowsPerEnd  int
	AllowedValues []string
	ValueScoreMap map[string]int
}

func (r *ScoringRepo) GetStageInfo(ctx context.Context, stageID string) (*StageInfo, error) {
	var arrowsPerEnd int
	var allowedJSON, scoreMapJSON []byte
	err := r.DB.QueryRow(ctx,
		"SELECT arrows_per_end, allowed_values, value_score_map FROM round_template_stages WHERE id = $1",
		stageID,
	).Scan(&arrowsPerEnd, &allowedJSON, &scoreMapJSON)
	if err != nil {
		return nil, err
	}

	var info StageInfo
	info.ArrowsPerEnd = arrowsPerEnd
	json.Unmarshal(allowedJSON, &info.AllowedValues)
	json.Unmarshal(scoreMapJSON, &info.ValueScoreMap)
	return &info, nil
}

func (r *ScoringRepo) GetEndCount(ctx context.Context, sessionID string) int {
	var count int
	r.DB.QueryRow(ctx, "SELECT COUNT(*) FROM ends WHERE session_id = $1", sessionID).Scan(&count)
	return count
}

type ArrowIn struct {
	ScoreValue string   `json:"score_value"`
	XPos       *float64 `json:"x_pos"`
	YPos       *float64 `json:"y_pos"`
}

func (r *ScoringRepo) SubmitEnd(ctx context.Context, sessionID, stageID string, endNumber int, arrows []ArrowIn, scoreMap map[string]int) (*EndOut, error) {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	endID := uuid.New().String()
	now := time.Now().UTC()
	_, err = tx.Exec(ctx, `
		INSERT INTO ends (id, session_id, stage_id, end_number, end_total, created_at)
		VALUES ($1, $2, $3, $4, 0, $5)`,
		endID, sessionID, stageID, endNumber, now,
	)
	if err != nil {
		return nil, err
	}

	endTotal := 0
	xCount := 0
	arrowOuts := []ArrowOut{}
	for i, a := range arrows {
		arrowID := uuid.New().String()
		numeric := scoreMap[a.ScoreValue]
		_, err = tx.Exec(ctx, `
			INSERT INTO arrows (id, end_id, arrow_number, score_value, score_numeric, x_pos, y_pos)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			arrowID, endID, i+1, a.ScoreValue, numeric, a.XPos, a.YPos,
		)
		if err != nil {
			return nil, err
		}
		endTotal += numeric
		if a.ScoreValue == "X" {
			xCount++
		}
		arrowOuts = append(arrowOuts, ArrowOut{
			ID: arrowID, ArrowNumber: i + 1,
			ScoreValue: a.ScoreValue, ScoreNumeric: numeric,
			XPos: a.XPos, YPos: a.YPos,
		})
	}

	_, err = tx.Exec(ctx, "UPDATE ends SET end_total = $1 WHERE id = $2", endTotal, endID)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx, `
		UPDATE scoring_sessions
		SET total_score = total_score + $2,
		    total_x_count = total_x_count + $3,
		    total_arrows = total_arrows + $4
		WHERE id = $1`,
		sessionID, endTotal, xCount, len(arrows),
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	sid := stageID
	return &EndOut{
		ID: endID, EndNumber: endNumber, EndTotal: endTotal,
		StageID: &sid, Arrows: arrowOuts, CreatedAt: now,
	}, nil
}

// ── Undo Last End ────────────────────────────────────────────────────

func (r *ScoringRepo) UndoLastEnd(ctx context.Context, sessionID string) error {
	var endID string
	var endTotal int
	err := r.DB.QueryRow(ctx, `
		SELECT id, end_total FROM ends
		WHERE session_id = $1
		ORDER BY end_number DESC LIMIT 1`, sessionID,
	).Scan(&endID, &endTotal)
	if err != nil {
		return err
	}

	var arrowCount, xCount int
	r.DB.QueryRow(ctx,
		"SELECT COUNT(*), COALESCE(SUM(CASE WHEN score_value = 'X' THEN 1 ELSE 0 END), 0) FROM arrows WHERE end_id = $1",
		endID,
	).Scan(&arrowCount, &xCount)

	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	tx.Exec(ctx, "DELETE FROM arrows WHERE end_id = $1", endID)
	tx.Exec(ctx, "DELETE FROM ends WHERE id = $1", endID)
	tx.Exec(ctx, `
		UPDATE scoring_sessions
		SET total_score = total_score - $2,
		    total_x_count = total_x_count - $3,
		    total_arrows = total_arrows - $4
		WHERE id = $1`,
		sessionID, endTotal, xCount, arrowCount,
	)

	return tx.Commit(ctx)
}

// ── Complete Session ──────────────────────────────────────────────────

type CompleteResult struct {
	IsPersonalBest bool
}

func (r *ScoringRepo) GetSessionForComplete(ctx context.Context, sessionID, userID string) (templateID, status string, totalScore int, err error) {
	err = r.DB.QueryRow(ctx,
		"SELECT template_id, status, total_score FROM scoring_sessions WHERE id = $1 AND user_id = $2",
		sessionID, userID,
	).Scan(&templateID, &status, &totalScore)
	return
}

func (r *ScoringRepo) CompleteSession(ctx context.Context, sessionID string, now time.Time, notes, location, weather *string) error {
	_, err := r.DB.Exec(ctx, `
		UPDATE scoring_sessions
		SET status = 'completed',
		    completed_at = $2,
		    notes = CASE WHEN $3::boolean THEN $4 ELSE notes END,
		    location = CASE WHEN $5::boolean THEN $6 ELSE location END,
		    weather = CASE WHEN $7::boolean THEN $8 ELSE weather END
		WHERE id = $1`,
		sessionID, now,
		notes != nil, notes,
		location != nil, location,
		weather != nil, weather,
	)
	return err
}

func (r *ScoringRepo) UpsertPersonalRecord(ctx context.Context, userID, templateID, sessionID string, totalScore int, now time.Time) (bool, error) {
	var existingPRID *string
	var existingScore int
	err := r.DB.QueryRow(ctx,
		"SELECT id, score FROM personal_records WHERE user_id = $1 AND template_id = $2",
		userID, templateID,
	).Scan(&existingPRID, &existingScore)

	if err != nil {
		// No existing PR — create one
		prID := uuid.New().String()
		r.DB.Exec(ctx, `
			INSERT INTO personal_records (id, user_id, template_id, session_id, score, achieved_at)
			VALUES ($1, $2, $3, $4, $5, $6)`,
			prID, userID, templateID, sessionID, totalScore, now,
		)
		return true, nil
	} else if totalScore > existingScore {
		r.DB.Exec(ctx, `
			UPDATE personal_records SET session_id = $1, score = $2, achieved_at = $3 WHERE id = $4`,
			sessionID, totalScore, now, *existingPRID,
		)
		return true, nil
	}
	return false, nil
}

func (r *ScoringRepo) InsertClassification(ctx context.Context, userID, system, classification, roundType string, score int, now time.Time, sessionID string) error {
	crID := uuid.New().String()
	_, err := r.DB.Exec(ctx, `
		INSERT INTO classification_records (id, user_id, system, classification, round_type, score, achieved_at, session_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		crID, userID, system, classification, roundType, score, now, sessionID,
	)
	return err
}

func (r *ScoringRepo) InsertNotification(ctx context.Context, userID, nType, title, message, link string, now time.Time) error {
	notifID := uuid.New().String()
	_, err := r.DB.Exec(ctx, `
		INSERT INTO notifications (id, user_id, type, title, message, link, is_read, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, false, $7)`,
		notifID, userID, nType, title, message, link, now,
	)
	return err
}

func (r *ScoringRepo) InsertFeedItem(ctx context.Context, userID, feedType string, data map[string]any, now time.Time) error {
	feedID := uuid.New().String()
	feedData, _ := json.Marshal(data)
	_, err := r.DB.Exec(ctx, `
		INSERT INTO feed_items (id, user_id, type, data, created_at)
		VALUES ($1, $2, $3, $4, $5)`,
		feedID, userID, feedType, feedData, now,
	)
	return err
}

// ── Abandon / Delete ──────────────────────────────────────────────────

func (r *ScoringRepo) AbandonSession(ctx context.Context, sessionID string) error {
	_, err := r.DB.Exec(ctx, "UPDATE scoring_sessions SET status = 'abandoned' WHERE id = $1", sessionID)
	return err
}

func (r *ScoringRepo) DeleteSession(ctx context.Context, sessionID string) error {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	tx.Exec(ctx, `DELETE FROM arrows WHERE end_id IN (SELECT id FROM ends WHERE session_id = $1)`, sessionID)
	tx.Exec(ctx, "DELETE FROM ends WHERE session_id = $1", sessionID)
	tx.Exec(ctx, "DELETE FROM scoring_sessions WHERE id = $1", sessionID)

	return tx.Commit(ctx)
}

// ── Stats ─────────────────────────────────────────────────────────────

func (r *ScoringRepo) Stats(ctx context.Context, userID string) (*StatsOut, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT ss.id, ss.status, ss.total_score, ss.total_arrows, ss.total_x_count,
		       ss.completed_at, ss.started_at, rt.name AS template_name, ss.template_id
		FROM scoring_sessions ss
		LEFT JOIN round_templates rt ON rt.id = ss.template_id
		WHERE ss.user_id = $1
		ORDER BY ss.completed_at DESC NULLS LAST`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type sessionInfo struct {
		id, status, templateID string
		totalScore, totalArrows, totalXCount int
		completedAt *time.Time
		startedAt   time.Time
		templateName *string
	}

	var sessions []sessionInfo
	for rows.Next() {
		var s sessionInfo
		if err := rows.Scan(&s.id, &s.status, &s.totalScore, &s.totalArrows, &s.totalXCount,
			&s.completedAt, &s.startedAt, &s.templateName, &s.templateID); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}

	totalArrows := 0
	totalXCount := 0
	for _, s := range sessions {
		totalArrows += s.totalArrows
		totalXCount += s.totalXCount
	}

	var completed []sessionInfo
	for _, s := range sessions {
		if s.status == "completed" {
			completed = append(completed, s)
		}
	}

	var bestScore *int
	var bestTemplate *string
	for _, s := range completed {
		if bestScore == nil || s.totalScore > *bestScore {
			score := s.totalScore
			bestScore = &score
			name := "Unknown"
			if s.templateName != nil {
				name = *s.templateName
			}
			bestTemplate = &name
		}
	}

	byType := map[string][]int{}
	for _, s := range completed {
		name := "Unknown"
		if s.templateName != nil {
			name = *s.templateName
		}
		byType[name] = append(byType[name], s.totalScore)
	}
	avgByRound := []RoundTypeAvg{}
	for name, scores := range byType {
		sum := 0
		for _, sc := range scores {
			sum += sc
		}
		avg := math.Round(float64(sum)/float64(len(scores))*10) / 10
		avgByRound = append(avgByRound, RoundTypeAvg{
			TemplateName: name, AvgScore: avg, Count: len(scores),
		})
	}

	recentTrend := []RecentTrendItem{}
	limit := 10
	if len(completed) < limit {
		limit = len(completed)
	}
	for _, s := range completed[:limit] {
		maxScore := r.getTemplateMaxScore(ctx, s.templateID)
		name := "Unknown"
		if s.templateName != nil {
			name = *s.templateName
		}
		date := s.startedAt
		if s.completedAt != nil {
			date = *s.completedAt
		}
		recentTrend = append(recentTrend, RecentTrendItem{
			Score: s.totalScore, MaxScore: maxScore, TemplateName: name, Date: date,
		})
	}

	prList, err := r.loadPersonalRecords(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &StatsOut{
		TotalSessions:     len(sessions),
		CompletedSessions: len(completed),
		TotalArrows:       totalArrows,
		TotalXCount:       totalXCount,
		PersonalBestScore: bestScore,
		PersonalBestTmpl:  bestTemplate,
		AvgByRoundType:    avgByRound,
		RecentTrend:       recentTrend,
		PersonalRecords:   prList,
	}, nil
}

// ── Personal Records ──────────────────────────────────────────────────

func (r *ScoringRepo) PersonalRecords(ctx context.Context, userID string) ([]PersonalRecordOut, error) {
	return r.loadPersonalRecords(ctx, userID)
}

func (r *ScoringRepo) loadPersonalRecords(ctx context.Context, userID string) ([]PersonalRecordOut, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT pr.template_id, pr.score, pr.achieved_at, pr.session_id, rt.name
		FROM personal_records pr
		LEFT JOIN round_templates rt ON rt.id = pr.template_id
		WHERE pr.user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []PersonalRecordOut{}
	for rows.Next() {
		var templateID, sessionID string
		var score int
		var achievedAt time.Time
		var tname *string
		if err := rows.Scan(&templateID, &score, &achievedAt, &sessionID, &tname); err != nil {
			continue
		}
		name := "Unknown"
		if tname != nil {
			name = *tname
		}
		maxScore := r.getTemplateMaxScore(ctx, templateID)
		items = append(items, PersonalRecordOut{
			TemplateName: name, Score: score, MaxScore: maxScore,
			AchievedAt: achievedAt, SessionID: sessionID,
		})
	}
	return items, nil
}

// ── Trends ────────────────────────────────────────────────────────────

func (r *ScoringRepo) Trends(ctx context.Context, userID string) ([]TrendDataItem, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT ss.id, ss.total_score, ss.completed_at, ss.started_at,
		       ss.template_id, rt.name
		FROM scoring_sessions ss
		LEFT JOIN round_templates rt ON rt.id = ss.template_id
		WHERE ss.user_id = $1 AND ss.status = 'completed'
		ORDER BY ss.completed_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []TrendDataItem{}
	for rows.Next() {
		var sid, templateID string
		var totalScore int
		var completedAt *time.Time
		var startedAt time.Time
		var tname *string
		if err := rows.Scan(&sid, &totalScore, &completedAt, &startedAt, &templateID, &tname); err != nil {
			continue
		}
		name := "Unknown"
		if tname != nil {
			name = *tname
		}
		maxScore := r.getTemplateMaxScore(ctx, templateID)
		pct := 0.0
		if maxScore > 0 {
			pct = math.Round(float64(totalScore)/float64(maxScore)*1000) / 10
		}
		date := startedAt
		if completedAt != nil {
			date = *completedAt
		}
		items = append(items, TrendDataItem{
			SessionID: sid, TemplateName: name, TotalScore: totalScore,
			MaxScore: maxScore, Percentage: pct, CompletedAt: date,
		})
	}
	return items, nil
}

// ── Export ─────────────────────────────────────────────────────────────

type BulkExportRow struct {
	StartedAt    time.Time
	CompletedAt  *time.Time
	TemplateName *string
	Status       string
	TotalScore   int
	TotalXCount  int
	TotalArrows  int
	Location     *string
	Notes        *string
}

func (r *ScoringRepo) ExportBulkData(ctx context.Context, userID string, templateID, dateFrom, dateTo, search *string) ([]BulkExportRow, error) {
	query := `
		SELECT ss.status, ss.total_score, ss.total_x_count, ss.total_arrows,
		       ss.started_at, ss.completed_at, ss.location, ss.notes,
		       rt.name AS template_name
		FROM scoring_sessions ss
		LEFT JOIN round_templates rt ON rt.id = ss.template_id
		WHERE ss.user_id = $1`
	args := []any{userID}
	argN := 2

	if templateID != nil {
		query += fmt.Sprintf(" AND ss.template_id = $%d", argN)
		args = append(args, *templateID)
		argN++
	}
	if dateFrom != nil {
		query += fmt.Sprintf(" AND ss.started_at >= $%d::date", argN)
		args = append(args, *dateFrom)
		argN++
	}
	if dateTo != nil {
		query += fmt.Sprintf(" AND ss.started_at <= ($%d::date + interval '1 day' - interval '1 second')", argN)
		args = append(args, *dateTo)
		argN++
	}
	if search != nil {
		pattern := "%" + *search + "%"
		query += fmt.Sprintf(" AND (ss.notes ILIKE $%d OR ss.location ILIKE $%d)", argN, argN)
		args = append(args, pattern)
		argN++
	}
	query += " ORDER BY ss.started_at DESC"

	rows, err := r.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []BulkExportRow
	for rows.Next() {
		var b BulkExportRow
		if err := rows.Scan(&b.Status, &b.TotalScore, &b.TotalXCount, &b.TotalArrows,
			&b.StartedAt, &b.CompletedAt, &b.Location, &b.Notes, &b.TemplateName); err != nil {
			continue
		}
		items = append(items, b)
	}
	return items, nil
}

func (r *ScoringRepo) GetTemplateName(ctx context.Context, templateID string) string {
	var name string
	r.DB.QueryRow(ctx, "SELECT name FROM round_templates WHERE id = $1", templateID).Scan(&name)
	return name
}

// ── Sharing ───────────────────────────────────────────────────────────

func (r *ScoringRepo) GetShareToken(ctx context.Context, sessionID, userID string) (*string, error) {
	var shareToken *string
	err := r.DB.QueryRow(ctx,
		"SELECT share_token FROM scoring_sessions WHERE id = $1 AND user_id = $2",
		sessionID, userID,
	).Scan(&shareToken)
	return shareToken, err
}

func (r *ScoringRepo) SetShareToken(ctx context.Context, sessionID, token string) error {
	_, err := r.DB.Exec(ctx,
		"UPDATE scoring_sessions SET share_token = $1 WHERE id = $2",
		token, sessionID,
	)
	return err
}

func (r *ScoringRepo) RevokeShareToken(ctx context.Context, sessionID, userID string) (bool, error) {
	tag, err := r.DB.Exec(ctx,
		"UPDATE scoring_sessions SET share_token = NULL WHERE id = $1 AND user_id = $2",
		sessionID, userID,
	)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

type SharedSessionData struct {
	SessionID   string
	TemplateID  string
	TotalScore  int
	TotalXCount int
	TotalArrows int
	Notes       *string
	Location    *string
	Weather     *string
	StartedAt   time.Time
	CompletedAt *time.Time
	UserID      string
}

func (r *ScoringRepo) GetSharedSession(ctx context.Context, token string) (*SharedSessionData, error) {
	var d SharedSessionData
	err := r.DB.QueryRow(ctx, `
		SELECT ss.id, ss.template_id, ss.total_score, ss.total_x_count, ss.total_arrows,
		       ss.notes, ss.location, ss.weather, ss.started_at, ss.completed_at, ss.user_id
		FROM scoring_sessions ss
		WHERE ss.share_token = $1`, token,
	).Scan(&d.SessionID, &d.TemplateID, &d.TotalScore, &d.TotalXCount, &d.TotalArrows,
		&d.Notes, &d.Location, &d.Weather, &d.StartedAt, &d.CompletedAt, &d.UserID)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// ── Helpers ───────────────────────────────────────────────────────────

func (r *ScoringRepo) LoadEnds(ctx context.Context, sessionID string) ([]EndOut, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT id, end_number, end_total, stage_id, created_at
		FROM ends WHERE session_id = $1
		ORDER BY end_number`, sessionID)
	if err != nil {
		return []EndOut{}, err
	}
	defer rows.Close()

	ends := []EndOut{}
	for rows.Next() {
		var e EndOut
		if err := rows.Scan(&e.ID, &e.EndNumber, &e.EndTotal, &e.StageID, &e.CreatedAt); err != nil {
			continue
		}
		e.Arrows, _ = r.loadArrows(ctx, e.ID)
		ends = append(ends, e)
	}
	return ends, nil
}

func (r *ScoringRepo) loadArrows(ctx context.Context, endID string) ([]ArrowOut, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT id, arrow_number, score_value, score_numeric, x_pos, y_pos
		FROM arrows WHERE end_id = $1
		ORDER BY arrow_number`, endID)
	if err != nil {
		return []ArrowOut{}, err
	}
	defer rows.Close()

	arrows := []ArrowOut{}
	for rows.Next() {
		var a ArrowOut
		if err := rows.Scan(&a.ID, &a.ArrowNumber, &a.ScoreValue, &a.ScoreNumeric, &a.XPos, &a.YPos); err != nil {
			continue
		}
		arrows = append(arrows, a)
	}
	return arrows, nil
}

func (r *ScoringRepo) getTemplateMaxScore(ctx context.Context, templateID string) int {
	rows, err := r.DB.Query(ctx, `
		SELECT num_ends, arrows_per_end, max_score_per_arrow
		FROM round_template_stages WHERE template_id = $1`, templateID)
	if err != nil {
		return 0
	}
	defer rows.Close()

	total := 0
	for rows.Next() {
		var numEnds, arrowsPerEnd, maxScore int
		if err := rows.Scan(&numEnds, &arrowsPerEnd, &maxScore); err != nil {
			continue
		}
		total += numEnds * arrowsPerEnd * maxScore
	}
	return total
}
