package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/quiverscore/backend-go/internal/config"
	"github.com/quiverscore/backend-go/internal/middleware"
)

type RoundsHandler struct {
	DB  *pgxpool.Pool
	Cfg *config.Config
}

func (h *RoundsHandler) Routes(r chi.Router) {
	r.With(middleware.OptionalAuth(h.Cfg.SecretKey)).Get("/", h.List)
	r.Get("/{id}", h.Get)

	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth(h.Cfg.SecretKey))
		r.Post("/", h.Create)
		r.Put("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
		r.Post("/{id}/share", h.Share)
		r.Delete("/{id}/share/{club_id}", h.Unshare)
	})
}

// ── Types ─────────────────────────────────────────────────────────────

type stageOut struct {
	ID              string         `json:"id"`
	StageOrder      int            `json:"stage_order"`
	Name            string         `json:"name"`
	Distance        *string        `json:"distance"`
	NumEnds         int            `json:"num_ends"`
	ArrowsPerEnd    int            `json:"arrows_per_end"`
	AllowedValues   []string       `json:"allowed_values"`
	ValueScoreMap   map[string]int `json:"value_score_map"`
	MaxScorePerArrow int           `json:"max_score_per_arrow"`
}

type roundTemplateOut struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	Organization string     `json:"organization"`
	Description  *string    `json:"description"`
	IsOfficial   bool       `json:"is_official"`
	CreatedBy    *string    `json:"created_by"`
	Stages       []stageOut `json:"stages"`
}

type stageCreate struct {
	Name             string         `json:"name"`
	Distance         *string        `json:"distance"`
	NumEnds          int            `json:"num_ends"`
	ArrowsPerEnd     int            `json:"arrows_per_end"`
	AllowedValues    []string       `json:"allowed_values"`
	ValueScoreMap    map[string]int `json:"value_score_map"`
	MaxScorePerArrow int            `json:"max_score_per_arrow"`
}

type roundTemplateCreate struct {
	Name         string        `json:"name"`
	Organization string        `json:"organization"`
	Description  *string       `json:"description"`
	Stages       []stageCreate `json:"stages"`
}

// ── List ──────────────────────────────────────────────────────────────

func (h *RoundsHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)

	var rows pgx.Rows
	var err error

	if userID != "" {
		rows, err = h.DB.Query(ctx, `
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
		rows, err = h.DB.Query(ctx, `
			SELECT id, name, organization, description, is_official, created_by
			FROM round_templates
			WHERE is_official = true
			ORDER BY name`)
	}
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	defer rows.Close()

	var templates []roundTemplateOut
	for rows.Next() {
		var t roundTemplateOut
		if err := rows.Scan(&t.ID, &t.Name, &t.Organization, &t.Description,
			&t.IsOfficial, &t.CreatedBy); err != nil {
			Error(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		templates = append(templates, t)
	}
	if err := rows.Err(); err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Load stages for all templates
	for i := range templates {
		stages, err := loadStages(ctx, h.DB, templates[i].ID)
		if err != nil {
			Error(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		templates[i].Stages = stages
	}

	if templates == nil {
		templates = []roundTemplateOut{}
	}

	JSON(w, http.StatusOK, templates)
}

// ── Get ───────────────────────────────────────────────────────────────

func (h *RoundsHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		Error(w, http.StatusNotFound, "Round template not found")
		return
	}

	ctx := r.Context()
	var t roundTemplateOut
	err := h.DB.QueryRow(ctx, `
		SELECT id, name, organization, description, is_official, created_by
		FROM round_templates WHERE id = $1`, id,
	).Scan(&t.ID, &t.Name, &t.Organization, &t.Description, &t.IsOfficial, &t.CreatedBy)
	if err != nil {
		Error(w, http.StatusNotFound, "Round template not found")
		return
	}

	stages, err := loadStages(ctx, h.DB, t.ID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	t.Stages = stages

	JSON(w, http.StatusOK, t)
}

// ── Create ────────────────────────────────────────────────────────────

func (h *RoundsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req roundTemplateCreate
	if !Decode(w, r, &req) {
		return
	}

	if req.Name == "" {
		ValidationError(w, "name is required")
		return
	}
	if req.Organization == "" {
		ValidationError(w, "organization is required")
		return
	}
	if len(req.Stages) == 0 {
		ValidationError(w, "At least one stage is required")
		return
	}

	userID := middleware.GetUserID(r.Context())
	ctx := r.Context()

	tx, err := h.DB.Begin(ctx)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	defer tx.Rollback(ctx)

	templateID := uuid.New().String()
	var t roundTemplateOut
	err = tx.QueryRow(ctx, `
		INSERT INTO round_templates (id, name, organization, description, is_official, created_by)
		VALUES ($1, $2, $3, $4, false, $5)
		RETURNING id, name, organization, description, is_official, created_by`,
		templateID, req.Name, req.Organization, req.Description, userID,
	).Scan(&t.ID, &t.Name, &t.Organization, &t.Description, &t.IsOfficial, &t.CreatedBy)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	stages, err := insertStages(ctx, tx, templateID, req.Stages)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	t.Stages = stages

	if err := tx.Commit(ctx); err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	JSON(w, http.StatusCreated, t)
}

// ── Update ────────────────────────────────────────────────────────────

func (h *RoundsHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		Error(w, http.StatusNotFound, "Round template not found")
		return
	}

	var req roundTemplateCreate
	if !Decode(w, r, &req) {
		return
	}

	if req.Name == "" {
		ValidationError(w, "name is required")
		return
	}
	if req.Organization == "" {
		ValidationError(w, "organization is required")
		return
	}
	if len(req.Stages) == 0 {
		ValidationError(w, "At least one stage is required")
		return
	}

	userID := middleware.GetUserID(r.Context())
	ctx := r.Context()

	// Check template exists and permissions
	var isOfficial bool
	var createdBy *string
	err := h.DB.QueryRow(ctx,
		"SELECT is_official, created_by FROM round_templates WHERE id = $1", id,
	).Scan(&isOfficial, &createdBy)
	if err != nil {
		Error(w, http.StatusNotFound, "Round template not found")
		return
	}
	if isOfficial {
		Error(w, http.StatusForbidden, "Cannot edit official round templates")
		return
	}
	if createdBy == nil || *createdBy != userID {
		Error(w, http.StatusForbidden, "You can only edit your own custom rounds")
		return
	}

	// Check for in-progress scoring sessions
	var hasInProgress bool
	err = h.DB.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM scoring_sessions
		 WHERE template_id = $1 AND status = 'in_progress')`, id,
	).Scan(&hasInProgress)
	if err == nil && hasInProgress {
		ValidationError(w, "Cannot edit a round template while a scoring session is in progress")
		return
	}

	tx, err := h.DB.Begin(ctx)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	defer tx.Rollback(ctx)

	var t roundTemplateOut
	err = tx.QueryRow(ctx, `
		UPDATE round_templates
		SET name = $2, organization = $3, description = $4
		WHERE id = $1
		RETURNING id, name, organization, description, is_official, created_by`,
		id, req.Name, req.Organization, req.Description,
	).Scan(&t.ID, &t.Name, &t.Organization, &t.Description, &t.IsOfficial, &t.CreatedBy)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Delete old stages
	if _, err := tx.Exec(ctx, "DELETE FROM round_template_stages WHERE template_id = $1", id); err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	stages, err := insertStages(ctx, tx, id, req.Stages)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	t.Stages = stages

	if err := tx.Commit(ctx); err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	JSON(w, http.StatusOK, t)
}

// ── Delete ────────────────────────────────────────────────────────────

func (h *RoundsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		Error(w, http.StatusNotFound, "Round template not found")
		return
	}

	userID := middleware.GetUserID(r.Context())
	ctx := r.Context()

	var isOfficial bool
	var createdBy *string
	err := h.DB.QueryRow(ctx,
		"SELECT is_official, created_by FROM round_templates WHERE id = $1", id,
	).Scan(&isOfficial, &createdBy)
	if err != nil {
		Error(w, http.StatusNotFound, "Round template not found")
		return
	}
	if isOfficial {
		Error(w, http.StatusForbidden, "Cannot delete official round templates")
		return
	}
	if createdBy == nil || *createdBy != userID {
		Error(w, http.StatusForbidden, "You can only delete your own custom rounds")
		return
	}

	// Stages cascade via FK, but delete explicitly to be safe
	tx, err := h.DB.Begin(ctx)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "DELETE FROM round_template_stages WHERE template_id = $1", id); err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	if _, err := tx.Exec(ctx, "DELETE FROM round_templates WHERE id = $1", id); err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if err := tx.Commit(ctx); err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ── Share ─────────────────────────────────────────────────────────────

type shareRoundRequest struct {
	ClubID string `json:"club_id"`
}

func (h *RoundsHandler) Share(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		Error(w, http.StatusNotFound, "Round template not found")
		return
	}

	var req shareRoundRequest
	if !Decode(w, r, &req) {
		return
	}

	userID := middleware.GetUserID(r.Context())
	ctx := r.Context()

	var isOfficial bool
	var createdBy *string
	err := h.DB.QueryRow(ctx,
		"SELECT is_official, created_by FROM round_templates WHERE id = $1", id,
	).Scan(&isOfficial, &createdBy)
	if err != nil {
		Error(w, http.StatusNotFound, "Round template not found")
		return
	}
	if isOfficial {
		Error(w, http.StatusForbidden, "Cannot share official round templates")
		return
	}
	if createdBy == nil || *createdBy != userID {
		Error(w, http.StatusForbidden, "You can only share your own custom rounds")
		return
	}

	// Check user is a member of the club
	var isMember bool
	err = h.DB.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM club_members WHERE club_id = $1 AND user_id = $2)",
		req.ClubID, userID,
	).Scan(&isMember)
	if err != nil || !isMember {
		Error(w, http.StatusUnauthorized, "You are not a member of this club")
		return
	}

	// Check duplicate
	var alreadyShared bool
	err = h.DB.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM club_shared_rounds WHERE club_id = $1 AND template_id = $2)",
		req.ClubID, id,
	).Scan(&alreadyShared)
	if err == nil && alreadyShared {
		Error(w, http.StatusConflict, "Round is already shared with this club")
		return
	}

	_, err = h.DB.Exec(ctx,
		`INSERT INTO club_shared_rounds (id, club_id, template_id, shared_by, shared_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		uuid.New().String(), req.ClubID, id, userID, time.Now(),
	)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	JSON(w, http.StatusCreated, map[string]string{"detail": "Round shared with club"})
}

// ── Unshare ───────────────────────────────────────────────────────────

func (h *RoundsHandler) Unshare(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	clubID := chi.URLParam(r, "club_id")

	if _, err := uuid.Parse(id); err != nil {
		Error(w, http.StatusNotFound, "Round template not found")
		return
	}
	if _, err := uuid.Parse(clubID); err != nil {
		Error(w, http.StatusNotFound, "Share not found")
		return
	}

	userID := middleware.GetUserID(r.Context())
	ctx := r.Context()

	var createdBy *string
	err := h.DB.QueryRow(ctx,
		"SELECT created_by FROM round_templates WHERE id = $1", id,
	).Scan(&createdBy)
	if err != nil {
		Error(w, http.StatusNotFound, "Round template not found")
		return
	}
	if createdBy == nil || *createdBy != userID {
		Error(w, http.StatusForbidden, "You can only unshare your own custom rounds")
		return
	}

	tag, err := h.DB.Exec(ctx,
		"DELETE FROM club_shared_rounds WHERE club_id = $1 AND template_id = $2",
		clubID, id,
	)
	if err != nil || tag.RowsAffected() == 0 {
		Error(w, http.StatusNotFound, "Share not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ── Helpers ───────────────────────────────────────────────────────────

func loadStages(ctx context.Context, db *pgxpool.Pool, templateID string) ([]stageOut, error) {
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

	var stages []stageOut
	for rows.Next() {
		var s stageOut
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
		stages = []stageOut{}
	}

	return stages, rows.Err()
}

func insertStages(ctx context.Context, tx pgx.Tx, templateID string, stages []stageCreate) ([]stageOut, error) {
	var out []stageOut
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

		out = append(out, stageOut{
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
