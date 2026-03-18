package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RoundRepo struct {
	DB *pgxpool.Pool
}

// ── Types ─────────────────────────────────────────────────────────────

type StageOut struct {
	ID               string         `json:"id"`
	StageOrder       int            `json:"stage_order"`
	Name             string         `json:"name"`
	Distance         *string        `json:"distance"`
	NumEnds          int            `json:"num_ends"`
	ArrowsPerEnd     int            `json:"arrows_per_end"`
	AllowedValues    []string       `json:"allowed_values"`
	ValueScoreMap    map[string]int `json:"value_score_map"`
	MaxScorePerArrow int            `json:"max_score_per_arrow"`
}

type RoundTemplateOut struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	Organization string     `json:"organization"`
	Description  *string    `json:"description"`
	IsOfficial   bool       `json:"is_official"`
	CreatedBy    *string    `json:"created_by"`
	Stages       []StageOut `json:"stages"`
}

type StageParams struct {
	Name             string         `json:"name"`
	Distance         *string        `json:"distance"`
	NumEnds          int            `json:"num_ends"`
	ArrowsPerEnd     int            `json:"arrows_per_end"`
	AllowedValues    []string       `json:"allowed_values"`
	ValueScoreMap    map[string]int `json:"value_score_map"`
	MaxScorePerArrow int            `json:"max_score_per_arrow"`
}

// ── Methods ───────────────────────────────────────────────────────────

func (r *RoundRepo) List(ctx context.Context, userID string) ([]RoundTemplateOut, error) {
	var rows pgx.Rows
	var err error

	if userID != "" {
		rows, err = r.DB.Query(ctx, `
			SELECT DISTINCT rt.id, rt.name, rt.organization, rt.description,
			       rt.is_official, rt.created_by
			FROM round_templates rt
			LEFT JOIN club_shared_rounds csr ON csr.template_id = rt.id
			LEFT JOIN club_members cm ON cm.club_id = csr.club_id AND cm.user_id = $1
			WHERE rt.is_official = true
			   OR rt.created_by = $1
			   OR cm.user_id IS NOT NULL
			ORDER BY rt.name`, userID)
	} else {
		rows, err = r.DB.Query(ctx, `
			SELECT id, name, organization, description, is_official, created_by
			FROM round_templates
			WHERE is_official = true
			ORDER BY name`)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []RoundTemplateOut
	for rows.Next() {
		var t RoundTemplateOut
		if err := rows.Scan(&t.ID, &t.Name, &t.Organization, &t.Description,
			&t.IsOfficial, &t.CreatedBy); err != nil {
			return nil, err
		}
		templates = append(templates, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range templates {
		stages, err := r.LoadStages(ctx, templates[i].ID)
		if err != nil {
			return nil, err
		}
		templates[i].Stages = stages
	}

	if templates == nil {
		templates = []RoundTemplateOut{}
	}

	return templates, nil
}

func (r *RoundRepo) Get(ctx context.Context, id string) (*RoundTemplateOut, error) {
	var t RoundTemplateOut
	err := r.DB.QueryRow(ctx, `
		SELECT id, name, organization, description, is_official, created_by
		FROM round_templates WHERE id = $1`, id,
	).Scan(&t.ID, &t.Name, &t.Organization, &t.Description, &t.IsOfficial, &t.CreatedBy)
	if err != nil {
		return nil, err
	}

	stages, err := r.LoadStages(ctx, t.ID)
	if err != nil {
		return nil, err
	}
	t.Stages = stages

	return &t, nil
}

func (r *RoundRepo) Create(ctx context.Context, name, organization string, description *string, userID string, stages []StageParams) (*RoundTemplateOut, error) {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	templateID := uuid.New().String()
	var t RoundTemplateOut
	err = tx.QueryRow(ctx, `
		INSERT INTO round_templates (id, name, organization, description, is_official, created_by)
		VALUES ($1, $2, $3, $4, false, $5)
		RETURNING id, name, organization, description, is_official, created_by`,
		templateID, name, organization, description, userID,
	).Scan(&t.ID, &t.Name, &t.Organization, &t.Description, &t.IsOfficial, &t.CreatedBy)
	if err != nil {
		return nil, err
	}

	stageOuts, err := insertStages(ctx, tx, templateID, stages)
	if err != nil {
		return nil, err
	}
	t.Stages = stageOuts

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &t, nil
}

func (r *RoundRepo) Update(ctx context.Context, id, name, organization string, description *string, stages []StageParams) (*RoundTemplateOut, error) {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var t RoundTemplateOut
	err = tx.QueryRow(ctx, `
		UPDATE round_templates
		SET name = $2, organization = $3, description = $4
		WHERE id = $1
		RETURNING id, name, organization, description, is_official, created_by`,
		id, name, organization, description,
	).Scan(&t.ID, &t.Name, &t.Organization, &t.Description, &t.IsOfficial, &t.CreatedBy)
	if err != nil {
		return nil, err
	}

	if _, err := tx.Exec(ctx, "DELETE FROM round_template_stages WHERE template_id = $1", id); err != nil {
		return nil, err
	}

	stageOuts, err := insertStages(ctx, tx, id, stages)
	if err != nil {
		return nil, err
	}
	t.Stages = stageOuts

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &t, nil
}

func (r *RoundRepo) Delete(ctx context.Context, id string) error {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "DELETE FROM round_template_stages WHERE template_id = $1", id); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, "DELETE FROM round_templates WHERE id = $1", id); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *RoundRepo) GetPermissions(ctx context.Context, id string) (isOfficial bool, createdBy *string, err error) {
	err = r.DB.QueryRow(ctx,
		"SELECT is_official, created_by FROM round_templates WHERE id = $1", id,
	).Scan(&isOfficial, &createdBy)
	return
}

func (r *RoundRepo) HasInProgressSessions(ctx context.Context, templateID string) (bool, error) {
	var has bool
	err := r.DB.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM scoring_sessions
		 WHERE template_id = $1 AND status = 'in_progress')`, templateID,
	).Scan(&has)
	return has, err
}

func (r *RoundRepo) IsMemberOfClub(ctx context.Context, clubID, userID string) (bool, error) {
	var is bool
	err := r.DB.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM club_members WHERE club_id = $1 AND user_id = $2)",
		clubID, userID,
	).Scan(&is)
	return is, err
}

func (r *RoundRepo) IsSharedWithClub(ctx context.Context, clubID, templateID string) (bool, error) {
	var is bool
	err := r.DB.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM club_shared_rounds WHERE club_id = $1 AND template_id = $2)",
		clubID, templateID,
	).Scan(&is)
	return is, err
}

func (r *RoundRepo) ShareWithClub(ctx context.Context, templateID, clubID, userID string) error {
	_, err := r.DB.Exec(ctx,
		`INSERT INTO club_shared_rounds (id, club_id, template_id, shared_by, shared_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		uuid.New().String(), clubID, templateID, userID, time.Now(),
	)
	return err
}

func (r *RoundRepo) UnshareFromClub(ctx context.Context, clubID, templateID string) (bool, error) {
	tag, err := r.DB.Exec(ctx,
		"DELETE FROM club_shared_rounds WHERE club_id = $1 AND template_id = $2",
		clubID, templateID,
	)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func (r *RoundRepo) GetTemplateName(ctx context.Context, templateID string) (string, error) {
	var name string
	err := r.DB.QueryRow(ctx, "SELECT name FROM round_templates WHERE id = $1", templateID).Scan(&name)
	return name, err
}

func (r *RoundRepo) GetTemplateMaxScore(ctx context.Context, templateID string) int {
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

// LoadStages loads stages for a template. Exported for use by other repos.
func (r *RoundRepo) LoadStages(ctx context.Context, templateID string) ([]StageOut, error) {
	return LoadStages(ctx, r.DB, templateID)
}

// LoadStages is a package-level function usable by any repo with a pool.
func LoadStages(ctx context.Context, db *pgxpool.Pool, templateID string) ([]StageOut, error) {
	rows, err := db.Query(ctx, `
		SELECT id, stage_order, name, distance, num_ends, arrows_per_end,
		       allowed_values, value_score_map, max_score_per_arrow
		FROM round_template_stages
		WHERE template_id = $1
		ORDER BY stage_order`, templateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stages []StageOut
	for rows.Next() {
		var s StageOut
		var allowedJSON, scoreMapJSON []byte
		if err := rows.Scan(&s.ID, &s.StageOrder, &s.Name, &s.Distance,
			&s.NumEnds, &s.ArrowsPerEnd, &allowedJSON, &scoreMapJSON,
			&s.MaxScorePerArrow); err != nil {
			return nil, err
		}
		json.Unmarshal(allowedJSON, &s.AllowedValues)
		json.Unmarshal(scoreMapJSON, &s.ValueScoreMap)
		stages = append(stages, s)
	}

	if stages == nil {
		stages = []StageOut{}
	}

	return stages, rows.Err()
}

func insertStages(ctx context.Context, tx pgx.Tx, templateID string, stages []StageParams) ([]StageOut, error) {
	var out []StageOut
	for i, s := range stages {
		stageID := uuid.New().String()
		allowedJSON, _ := json.Marshal(s.AllowedValues)
		scoreMapJSON, _ := json.Marshal(s.ValueScoreMap)

		_, err := tx.Exec(ctx, `
			INSERT INTO round_template_stages
			(id, template_id, stage_order, name, distance, num_ends, arrows_per_end,
			 allowed_values, value_score_map, max_score_per_arrow)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
			stageID, templateID, i+1, s.Name, s.Distance, s.NumEnds, s.ArrowsPerEnd,
			allowedJSON, scoreMapJSON, s.MaxScorePerArrow,
		)
		if err != nil {
			return nil, err
		}

		out = append(out, StageOut{
			ID:               stageID,
			StageOrder:       i + 1,
			Name:             s.Name,
			Distance:         s.Distance,
			NumEnds:          s.NumEnds,
			ArrowsPerEnd:     s.ArrowsPerEnd,
			AllowedValues:    s.AllowedValues,
			ValueScoreMap:    s.ValueScoreMap,
			MaxScorePerArrow: s.MaxScorePerArrow,
		})
	}
	return out, nil
}
