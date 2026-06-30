package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/quiverscore/backend-go/internal/middleware"
	"github.com/quiverscore/backend-go/internal/repository"
)

// ── CreateTournament ──────────────────────────────────────────────────

func TestCreateTournament_Success(t *testing.T) {
	userID := uuid.New().String()
	clubID := uuid.New().String()
	templateID := uuid.New().String()
	now := time.Now().UTC()
	regDeadline := now.Add(48 * time.Hour).Format(time.RFC3339)
	startDate := now.Add(72 * time.Hour).Format(time.RFC3339)
	endDate := now.Add(96 * time.Hour).Format(time.RFC3339)

	mock := &mockClubRepo{
		createTournamentFn: func(_ context.Context, id, cID, uID, name string, desc *string, tplID string, maxP *int, rd, sd, ed time.Time) (*repository.TournamentOut, error) {
			return &repository.TournamentOut{
				ID:                   id,
				Name:                 name,
				OrganizerID:          uID,
				TemplateID:           tplID,
				Status:               "registration",
				RegistrationDeadline: &rd,
				StartDate:            &sd,
				EndDate:              &ed,
				ClubID:               cID,
				CreatedAt:            time.Now().UTC(),
			}, nil
		},
	}
	h := clubsHandler(mock)

	req := clubsAuthedBody(http.MethodPost, "/tournaments", userID, map[string]any{
		"name":                  "Spring Tournament",
		"template_id":           templateID,
		"registration_deadline": regDeadline,
		"start_date":            startDate,
		"end_date":              endDate,
	})
	req = withChiURLParam(req, "clubID", clubID)

	rr := httptest.NewRecorder()
	h.CreateTournament(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var result repository.TournamentOut
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result.Name != "Spring Tournament" {
		t.Errorf("expected 'Spring Tournament', got '%s'", result.Name)
	}
	if result.Status != "registration" {
		t.Errorf("expected status 'registration', got '%s'", result.Status)
	}
}

func TestCreateTournament_InvalidRegDeadline(t *testing.T) {
	h := clubsHandler(&mockClubRepo{})

	req := clubsAuthedBody(http.MethodPost, "/tournaments", uuid.New().String(), map[string]any{
		"name":                  "Bad Tournament",
		"template_id":           uuid.New().String(),
		"registration_deadline": "not-a-date",
		"start_date":            time.Now().Add(72 * time.Hour).UTC().Format(time.RFC3339),
		"end_date":              time.Now().Add(96 * time.Hour).UTC().Format(time.RFC3339),
	})
	req = withChiURLParam(req, "clubID", uuid.New().String())

	rr := httptest.NewRecorder()
	h.CreateTournament(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestCreateTournament_InvalidStartDate(t *testing.T) {
	h := clubsHandler(&mockClubRepo{})

	req := clubsAuthedBody(http.MethodPost, "/tournaments", uuid.New().String(), map[string]any{
		"name":                  "Bad Tournament",
		"template_id":           uuid.New().String(),
		"registration_deadline": time.Now().Add(48 * time.Hour).UTC().Format(time.RFC3339),
		"start_date":            "not-a-date",
		"end_date":              time.Now().Add(96 * time.Hour).UTC().Format(time.RFC3339),
	})
	req = withChiURLParam(req, "clubID", uuid.New().String())

	rr := httptest.NewRecorder()
	h.CreateTournament(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestCreateTournament_InvalidEndDate(t *testing.T) {
	h := clubsHandler(&mockClubRepo{})

	req := clubsAuthedBody(http.MethodPost, "/tournaments", uuid.New().String(), map[string]any{
		"name":                  "Bad Tournament",
		"template_id":           uuid.New().String(),
		"registration_deadline": time.Now().Add(48 * time.Hour).UTC().Format(time.RFC3339),
		"start_date":            time.Now().Add(72 * time.Hour).UTC().Format(time.RFC3339),
		"end_date":              "not-a-date",
	})
	req = withChiURLParam(req, "clubID", uuid.New().String())

	rr := httptest.NewRecorder()
	h.CreateTournament(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestCreateTournament_Unauthorized(t *testing.T) {
	now := time.Now().UTC()
	mock := &mockClubRepo{
		createTournamentFn: func(_ context.Context, _, _, _, _ string, _ *string, _ string, _ *int, _, _, _ time.Time) (*repository.TournamentOut, error) {
			return nil, errors.New("forbidden")
		},
	}
	h := clubsHandler(mock)

	req := clubsAuthedBody(http.MethodPost, "/tournaments", uuid.New().String(), map[string]any{
		"name":                  "Tourney",
		"template_id":           uuid.New().String(),
		"registration_deadline": now.Add(48 * time.Hour).Format(time.RFC3339),
		"start_date":            now.Add(72 * time.Hour).Format(time.RFC3339),
		"end_date":              now.Add(96 * time.Hour).Format(time.RFC3339),
	})
	req = withChiURLParam(req, "clubID", uuid.New().String())

	rr := httptest.NewRecorder()
	h.CreateTournament(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

// ── ListTournamentRounds error path ──────────────────────────────────

func TestListTournamentRounds_NotFound(t *testing.T) {
	mock := &mockClubRepo{
		listTournamentRoundsFn: func(_ context.Context, _, _, _ string) ([]repository.TournamentRoundOut, error) {
			return nil, errors.New("not found")
		},
	}
	h := clubsHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/rounds", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, uuid.New().String())
	req = req.WithContext(ctx)
	req = clubsChiParams(req, map[string]string{"clubID": "c1", "tournamentID": "t1"})

	rr := httptest.NewRecorder()
	h.ListTournamentRounds(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// ── SubmitTournamentRoundScore error path ─────────────────────────────

func TestSubmitTournamentRoundScore_RepoError(t *testing.T) {
	mock := &mockClubRepo{
		submitTournamentRoundScoreFn: func(_ context.Context, _, _, _, _, _ string) (*repository.TournamentRoundScoreOut, error) {
			return nil, errors.New("cannot submit")
		},
	}
	h := clubsHandler(mock)

	req := httptest.NewRequest(http.MethodPost, "/rounds/r1/submit-score?session_id=s1", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, uuid.New().String())
	req = req.WithContext(ctx)
	req = clubsChiParams(req, map[string]string{"clubID": "c1", "tournamentID": "t1", "roundID": "r1"})

	rr := httptest.NewRecorder()
	h.SubmitTournamentRoundScore(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// ── GetTournamentRoundLeaderboard error path ─────────────────────────

func TestGetTournamentRoundLeaderboard_NotFound(t *testing.T) {
	mock := &mockClubRepo{
		getTournamentRoundLeaderboardFn: func(_ context.Context, _, _, _, _ string) ([]repository.TournamentRoundScoreOut, error) {
			return nil, errors.New("not found")
		},
	}
	h := clubsHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/rounds/r1/leaderboard", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, uuid.New().String())
	req = req.WithContext(ctx)
	req = clubsChiParams(req, map[string]string{"clubID": "c1", "tournamentID": "t1", "roundID": "r1"})

	rr := httptest.NewRecorder()
	h.GetTournamentRoundLeaderboard(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// ── Tournament Rounds ────────────────────────────────────────────────

func TestAddTournamentRound_Success(t *testing.T) {
	const userID = "user-1"
	const clubID = "club-1"
	const tournamentID = "tourney-1"

	h := clubsHandler(&mockClubRepo{
		addTournamentRoundFn: func(_ context.Context, id, cID, tID, uID, name string, templateID *string, advancement *int, roundType string) (*repository.TournamentRoundOut, error) {
			return &repository.TournamentRoundOut{
				ID: id, TournamentID: tID, RoundNumber: 1, Name: name,
				TemplateID: templateID, Advancement: advancement, Status: "pending", RoundType: roundType,
			}, nil
		},
	})

	adv := 8
	req := clubsAuthedBody(http.MethodPost, "/rounds", userID, map[string]any{
		"name": "Qualifying Round", "advancement": adv,
	})
	req = clubsChiParams(req, map[string]string{"clubID": clubID, "tournamentID": tournamentID})
	rr := httptest.NewRecorder()

	h.AddTournamentRound(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
	var result repository.TournamentRoundOut
	json.NewDecoder(rr.Body).Decode(&result)
	if result.Name != "Qualifying Round" {
		t.Errorf("expected 'Qualifying Round', got '%s'", result.Name)
	}
	if result.Status != "pending" {
		t.Errorf("expected status 'pending', got '%s'", result.Status)
	}
}

func TestAddTournamentRound_MissingName(t *testing.T) {
	const userID = "user-1"
	const clubID = "club-1"
	const tournamentID = "tourney-1"

	h := clubsHandler(&mockClubRepo{})

	req := clubsAuthedBody(http.MethodPost, "/rounds", userID, map[string]any{})
	req = clubsChiParams(req, map[string]string{"clubID": clubID, "tournamentID": tournamentID})
	rr := httptest.NewRecorder()

	h.AddTournamentRound(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestAddTournamentRound_Forbidden(t *testing.T) {
	const userID = "user-1"
	const clubID = "club-1"
	const tournamentID = "tourney-1"

	h := clubsHandler(&mockClubRepo{
		addTournamentRoundFn: func(_ context.Context, _, _, _, _, _ string, _ *string, _ *int, _ string) (*repository.TournamentRoundOut, error) {
			return nil, pgx.ErrNoRows
		},
	})

	req := clubsAuthedBody(http.MethodPost, "/rounds", userID, map[string]any{"name": "Round 1"})
	req = clubsChiParams(req, map[string]string{"clubID": clubID, "tournamentID": tournamentID})
	rr := httptest.NewRecorder()

	h.AddTournamentRound(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestListTournamentRounds_Success(t *testing.T) {
	const userID = "user-1"
	const clubID = "club-1"
	const tournamentID = "tourney-1"

	h := clubsHandler(&mockClubRepo{
		listTournamentRoundsFn: func(_ context.Context, _, _, _ string) ([]repository.TournamentRoundOut, error) {
			return []repository.TournamentRoundOut{
				{ID: "r1", RoundNumber: 1, Name: "Qualifying", Status: "completed"},
				{ID: "r2", RoundNumber: 2, Name: "Final", Status: "pending"},
			}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/rounds", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	req = req.WithContext(ctx)
	req = clubsChiParams(req, map[string]string{"clubID": clubID, "tournamentID": tournamentID})
	rr := httptest.NewRecorder()

	h.ListTournamentRounds(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var result []repository.TournamentRoundOut
	json.NewDecoder(rr.Body).Decode(&result)
	if len(result) != 2 {
		t.Fatalf("expected 2 rounds, got %d", len(result))
	}
}

func TestStartTournamentRound_Success(t *testing.T) {
	const userID = "user-1"
	const clubID = "club-1"
	const tournamentID = "tourney-1"
	const roundID = "round-1"

	h := clubsHandler(&mockClubRepo{
		startTournamentRoundFn: func(_ context.Context, _, _, _, _ string) (*repository.TournamentRoundOut, error) {
			return &repository.TournamentRoundOut{
				ID: roundID, RoundNumber: 1, Name: "Qualifying", Status: "in_progress",
			}, nil
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/rounds/"+roundID+"/start", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	req = req.WithContext(ctx)
	req = clubsChiParams(req, map[string]string{"clubID": clubID, "tournamentID": tournamentID, "roundID": roundID})
	rr := httptest.NewRecorder()

	h.StartTournamentRound(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var result repository.TournamentRoundOut
	json.NewDecoder(rr.Body).Decode(&result)
	if result.Status != "in_progress" {
		t.Errorf("expected 'in_progress', got '%s'", result.Status)
	}
}

func TestStartTournamentRound_Forbidden(t *testing.T) {
	const userID = "user-1"

	h := clubsHandler(&mockClubRepo{
		startTournamentRoundFn: func(_ context.Context, _, _, _, _ string) (*repository.TournamentRoundOut, error) {
			return nil, pgx.ErrNoRows
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/rounds/r1/start", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	req = req.WithContext(ctx)
	req = clubsChiParams(req, map[string]string{"clubID": "c1", "tournamentID": "t1", "roundID": "r1"})
	rr := httptest.NewRecorder()

	h.StartTournamentRound(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestSubmitTournamentRoundScore_Success(t *testing.T) {
	const userID = "user-1"
	score := 280
	xcount := 5

	h := clubsHandler(&mockClubRepo{
		submitTournamentRoundScoreFn: func(_ context.Context, _, _, _, _, _ string) (*repository.TournamentRoundScoreOut, error) {
			return &repository.TournamentRoundScoreOut{
				ID: "score-1", RoundID: "r1", ParticipantID: "p1", UserID: userID,
				Score: &score, XCount: &xcount,
			}, nil
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/rounds/r1/submit-score?session_id=sess-1", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	req = req.WithContext(ctx)
	req = clubsChiParams(req, map[string]string{"clubID": "c1", "tournamentID": "t1", "roundID": "r1"})
	rr := httptest.NewRecorder()

	h.SubmitTournamentRoundScore(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestSubmitTournamentRoundScore_MissingSessionID(t *testing.T) {
	const userID = "user-1"

	h := clubsHandler(&mockClubRepo{})

	req := httptest.NewRequest(http.MethodPost, "/rounds/r1/submit-score", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	req = req.WithContext(ctx)
	req = clubsChiParams(req, map[string]string{"clubID": "c1", "tournamentID": "t1", "roundID": "r1"})
	rr := httptest.NewRecorder()

	h.SubmitTournamentRoundScore(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", rr.Code)
	}
}

func TestGetTournamentRoundLeaderboard_Success(t *testing.T) {
	const userID = "user-1"
	score1 := 290
	score2 := 270

	h := clubsHandler(&mockClubRepo{
		getTournamentRoundLeaderboardFn: func(_ context.Context, _, _, _, _ string) ([]repository.TournamentRoundScoreOut, error) {
			return []repository.TournamentRoundScoreOut{
				{ID: "s1", Score: &score1, Advanced: true},
				{ID: "s2", Score: &score2, Advanced: false},
			}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/rounds/r1/leaderboard", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	req = req.WithContext(ctx)
	req = clubsChiParams(req, map[string]string{"clubID": "c1", "tournamentID": "t1", "roundID": "r1"})
	rr := httptest.NewRecorder()

	h.GetTournamentRoundLeaderboard(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var result []repository.TournamentRoundScoreOut
	json.NewDecoder(rr.Body).Decode(&result)
	if len(result) != 2 {
		t.Fatalf("expected 2 scores, got %d", len(result))
	}
}

func TestCompleteTournamentRound_Success(t *testing.T) {
	const userID = "user-1"

	h := clubsHandler(&mockClubRepo{
		completeTournamentRoundFn: func(_ context.Context, _, _, _, _ string) (*repository.TournamentRoundOut, error) {
			return &repository.TournamentRoundOut{
				ID: "r1", RoundNumber: 1, Name: "Qualifying", Status: "completed",
			}, nil
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/rounds/r1/complete", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	req = req.WithContext(ctx)
	req = clubsChiParams(req, map[string]string{"clubID": "c1", "tournamentID": "t1", "roundID": "r1"})
	rr := httptest.NewRecorder()

	h.CompleteTournamentRound(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var result repository.TournamentRoundOut
	json.NewDecoder(rr.Body).Decode(&result)
	if result.Status != "completed" {
		t.Errorf("expected 'completed', got '%s'", result.Status)
	}
}

func TestCompleteTournamentRound_Forbidden(t *testing.T) {
	const userID = "user-1"

	h := clubsHandler(&mockClubRepo{
		completeTournamentRoundFn: func(_ context.Context, _, _, _, _ string) (*repository.TournamentRoundOut, error) {
			return nil, pgx.ErrNoRows
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/rounds/r1/complete", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	req = req.WithContext(ctx)
	req = clubsChiParams(req, map[string]string{"clubID": "c1", "tournamentID": "t1", "roundID": "r1"})
	rr := httptest.NewRecorder()

	h.CompleteTournamentRound(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

// ── ListTournaments ──────────────────────────────────────────────────

func TestListTournaments_Success(t *testing.T) {
	userID := uuid.New().String()
	clubID := uuid.New().String()
	now := time.Now().UTC()
	mock := &mockClubRepo{
		listTournamentsFn: func(_ context.Context, cID, uID string, status *string) ([]repository.TournamentOut, error) {
			return []repository.TournamentOut{
				{
					ID:          uuid.New().String(),
					Name:        "Spring Tournament",
					OrganizerID: uID,
					Status:      "registration",
					ClubID:      cID,
					CreatedAt:   now,
				},
				{
					ID:          uuid.New().String(),
					Name:        "Fall Tournament",
					OrganizerID: uID,
					Status:      "completed",
					ClubID:      cID,
					CreatedAt:   now,
				},
			}, nil
		},
	}
	h := clubsHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/tournaments", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	req = req.WithContext(ctx)
	req = withChiURLParam(req, "clubID", clubID)

	rr := httptest.NewRecorder()
	h.ListTournaments(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var result []repository.TournamentOut
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 tournaments, got %d", len(result))
	}
}

func TestListTournaments_WithStatusFilter(t *testing.T) {
	userID := uuid.New().String()
	clubID := uuid.New().String()
	var capturedStatus *string
	mock := &mockClubRepo{
		listTournamentsFn: func(_ context.Context, _, _ string, status *string) ([]repository.TournamentOut, error) {
			capturedStatus = status
			return []repository.TournamentOut{}, nil
		},
	}
	h := clubsHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/tournaments?status=registration", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	req = req.WithContext(ctx)
	req = withChiURLParam(req, "clubID", clubID)

	rr := httptest.NewRecorder()
	h.ListTournaments(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if capturedStatus == nil || *capturedStatus != "registration" {
		t.Error("expected status filter 'registration' to be passed to repo")
	}
}

func TestListTournaments_NotFound(t *testing.T) {
	mock := &mockClubRepo{
		listTournamentsFn: func(_ context.Context, _, _ string, _ *string) ([]repository.TournamentOut, error) {
			return nil, errors.New("not found")
		},
	}
	h := clubsHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/tournaments", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, uuid.New().String())
	req = req.WithContext(ctx)
	req = withChiURLParam(req, "clubID", uuid.New().String())

	rr := httptest.NewRecorder()
	h.ListTournaments(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// ── GetTournament ────────────────────────────────────────────────────

func TestGetTournament_Success(t *testing.T) {
	userID := uuid.New().String()
	clubID := uuid.New().String()
	tournamentID := uuid.New().String()
	now := time.Now().UTC()
	mock := &mockClubRepo{
		getTournamentDetailFn: func(_ context.Context, cID, tID, uID string) (*repository.TournamentDetailOut, error) {
			return &repository.TournamentDetailOut{
				TournamentOut: repository.TournamentOut{
					ID:          tID,
					Name:        "Spring Open",
					OrganizerID: uID,
					Status:      "in_progress",
					ClubID:      cID,
					CreatedAt:   now,
				},
				Participants: []repository.TournamentParticipantOut{
					{UserID: uuid.New().String(), Status: "registered"},
				},
			}, nil
		},
	}
	h := clubsHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/tournaments/"+tournamentID, nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	req = req.WithContext(ctx)
	req = clubsChiParams(req, map[string]string{"clubID": clubID, "tournamentID": tournamentID})

	rr := httptest.NewRecorder()
	h.GetTournament(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var result repository.TournamentDetailOut
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result.Name != "Spring Open" {
		t.Errorf("expected 'Spring Open', got '%s'", result.Name)
	}
	if len(result.Participants) != 1 {
		t.Errorf("expected 1 participant, got %d", len(result.Participants))
	}
}

func TestGetTournament_NotFound(t *testing.T) {
	mock := &mockClubRepo{
		getTournamentDetailFn: func(_ context.Context, _, _, _ string) (*repository.TournamentDetailOut, error) {
			return nil, errors.New("not found")
		},
	}
	h := clubsHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/tournaments/bad-id", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, uuid.New().String())
	req = req.WithContext(ctx)
	req = clubsChiParams(req, map[string]string{"clubID": uuid.New().String(), "tournamentID": "bad-id"})

	rr := httptest.NewRecorder()
	h.GetTournament(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// ── RegisterForTournament ────────────────────────────────────────────

func TestRegisterForTournament_Success(t *testing.T) {
	userID := uuid.New().String()
	clubID := uuid.New().String()
	tournamentID := uuid.New().String()
	mock := &mockClubRepo{
		registerForTournamentFn: func(_ context.Context, _, _, _ string) error {
			return nil
		},
	}
	h := clubsHandler(mock)

	req := httptest.NewRequest(http.MethodPost, "/tournaments/"+tournamentID+"/register", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	req = req.WithContext(ctx)
	req = clubsChiParams(req, map[string]string{"clubID": clubID, "tournamentID": tournamentID})

	rr := httptest.NewRecorder()
	h.RegisterForTournament(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var result map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result["message"] != "Registered successfully" {
		t.Errorf("expected 'Registered successfully', got '%s'", result["message"])
	}
}

func TestRegisterForTournament_AlreadyRegistered(t *testing.T) {
	mock := &mockClubRepo{
		registerForTournamentFn: func(_ context.Context, _, _, _ string) error {
			return repository.ErrAlreadyMember
		},
	}
	h := clubsHandler(mock)

	req := httptest.NewRequest(http.MethodPost, "/tournaments/t1/register", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, uuid.New().String())
	req = req.WithContext(ctx)
	req = clubsChiParams(req, map[string]string{"clubID": "c1", "tournamentID": "t1"})

	rr := httptest.NewRecorder()
	h.RegisterForTournament(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rr.Code)
	}
}

func TestRegisterForTournament_NotFound(t *testing.T) {
	mock := &mockClubRepo{
		registerForTournamentFn: func(_ context.Context, _, _, _ string) error {
			return errors.New("not found")
		},
	}
	h := clubsHandler(mock)

	req := httptest.NewRequest(http.MethodPost, "/tournaments/t1/register", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, uuid.New().String())
	req = req.WithContext(ctx)
	req = clubsChiParams(req, map[string]string{"clubID": "c1", "tournamentID": "t1"})

	rr := httptest.NewRecorder()
	h.RegisterForTournament(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// ── StartTournament ──────────────────────────────────────────────────

func TestStartTournament_Success(t *testing.T) {
	userID := uuid.New().String()
	clubID := uuid.New().String()
	tournamentID := uuid.New().String()
	mock := &mockClubRepo{
		startTournamentFn: func(_ context.Context, cID, tID, uID string) (*repository.TournamentOut, error) {
			return &repository.TournamentOut{
				ID:          tID,
				Name:        "Spring Open",
				OrganizerID: uID,
				Status:      "in_progress",
				ClubID:      cID,
				CreatedAt:   time.Now().UTC(),
			}, nil
		},
	}
	h := clubsHandler(mock)

	req := httptest.NewRequest(http.MethodPost, "/tournaments/"+tournamentID+"/start", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	req = req.WithContext(ctx)
	req = clubsChiParams(req, map[string]string{"clubID": clubID, "tournamentID": tournamentID})

	rr := httptest.NewRecorder()
	h.StartTournament(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var result repository.TournamentOut
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result.Status != "in_progress" {
		t.Errorf("expected 'in_progress', got '%s'", result.Status)
	}
}

func TestStartTournament_Forbidden(t *testing.T) {
	mock := &mockClubRepo{
		startTournamentFn: func(_ context.Context, _, _, _ string) (*repository.TournamentOut, error) {
			return nil, errors.New("forbidden")
		},
	}
	h := clubsHandler(mock)

	req := httptest.NewRequest(http.MethodPost, "/tournaments/t1/start", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, uuid.New().String())
	req = req.WithContext(ctx)
	req = clubsChiParams(req, map[string]string{"clubID": "c1", "tournamentID": "t1"})

	rr := httptest.NewRecorder()
	h.StartTournament(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

// ── TournamentLeaderboard ────────────────────────────────────────────

func TestTournamentLeaderboard_Success(t *testing.T) {
	userID := uuid.New().String()
	clubID := uuid.New().String()
	tournamentID := uuid.New().String()
	score1 := 680
	score2 := 650
	username1 := "archer1"
	username2 := "archer2"
	mock := &mockClubRepo{
		tournamentLeaderboardFn: func(_ context.Context, _, _, _ string) ([]repository.TournamentLeaderboardEntry, error) {
			return []repository.TournamentLeaderboardEntry{
				{Rank: 1, UserID: uuid.New().String(), Username: &username1, FinalScore: &score1, Status: "registered"},
				{Rank: 2, UserID: uuid.New().String(), Username: &username2, FinalScore: &score2, Status: "registered"},
			}, nil
		},
	}
	h := clubsHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/tournaments/"+tournamentID+"/leaderboard", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	req = req.WithContext(ctx)
	req = clubsChiParams(req, map[string]string{"clubID": clubID, "tournamentID": tournamentID})

	rr := httptest.NewRecorder()
	h.TournamentLeaderboard(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var result []repository.TournamentLeaderboardEntry
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result))
	}
	if result[0].Rank != 1 {
		t.Errorf("expected rank 1, got %d", result[0].Rank)
	}
}

func TestTournamentLeaderboard_NotFound(t *testing.T) {
	mock := &mockClubRepo{
		tournamentLeaderboardFn: func(_ context.Context, _, _, _ string) ([]repository.TournamentLeaderboardEntry, error) {
			return nil, errors.New("not found")
		},
	}
	h := clubsHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/tournaments/t1/leaderboard", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, uuid.New().String())
	req = req.WithContext(ctx)
	req = clubsChiParams(req, map[string]string{"clubID": "c1", "tournamentID": "t1"})

	rr := httptest.NewRecorder()
	h.TournamentLeaderboard(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// ── CompleteTournament ───────────────────────────────────────────────

func TestCompleteTournament_Success(t *testing.T) {
	userID := uuid.New().String()
	clubID := uuid.New().String()
	tournamentID := uuid.New().String()
	mock := &mockClubRepo{
		completeTournamentFn: func(_ context.Context, cID, tID, uID string) (*repository.TournamentOut, error) {
			return &repository.TournamentOut{
				ID:          tID,
				Name:        "Spring Open",
				OrganizerID: uID,
				Status:      "completed",
				ClubID:      cID,
				CreatedAt:   time.Now().UTC(),
			}, nil
		},
	}
	h := clubsHandler(mock)

	req := httptest.NewRequest(http.MethodPost, "/tournaments/"+tournamentID+"/complete", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	req = req.WithContext(ctx)
	req = clubsChiParams(req, map[string]string{"clubID": clubID, "tournamentID": tournamentID})

	rr := httptest.NewRecorder()
	h.CompleteTournament(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var result repository.TournamentOut
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result.Status != "completed" {
		t.Errorf("expected 'completed', got '%s'", result.Status)
	}
}

func TestCompleteTournament_Forbidden(t *testing.T) {
	mock := &mockClubRepo{
		completeTournamentFn: func(_ context.Context, _, _, _ string) (*repository.TournamentOut, error) {
			return nil, errors.New("forbidden")
		},
	}
	h := clubsHandler(mock)

	req := httptest.NewRequest(http.MethodPost, "/tournaments/t1/complete", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, uuid.New().String())
	req = req.WithContext(ctx)
	req = clubsChiParams(req, map[string]string{"clubID": "c1", "tournamentID": "t1"})

	rr := httptest.NewRecorder()
	h.CompleteTournament(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

// ── WithdrawFromTournament ───────────────────────────────────────────

func TestWithdrawFromTournament_Success(t *testing.T) {
	userID := uuid.New().String()
	clubID := uuid.New().String()
	tournamentID := uuid.New().String()
	mock := &mockClubRepo{
		withdrawFromTournamentFn: func(_ context.Context, _, _, _ string) error {
			return nil
		},
	}
	h := clubsHandler(mock)

	req := httptest.NewRequest(http.MethodPost, "/tournaments/"+tournamentID+"/withdraw", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	req = req.WithContext(ctx)
	req = clubsChiParams(req, map[string]string{"clubID": clubID, "tournamentID": tournamentID})

	rr := httptest.NewRecorder()
	h.WithdrawFromTournament(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var result map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result["message"] != "Withdrawn successfully" {
		t.Errorf("expected 'Withdrawn successfully', got '%s'", result["message"])
	}
}

func TestWithdrawFromTournament_NotFound(t *testing.T) {
	mock := &mockClubRepo{
		withdrawFromTournamentFn: func(_ context.Context, _, _, _ string) error {
			return errors.New("not found")
		},
	}
	h := clubsHandler(mock)

	req := httptest.NewRequest(http.MethodPost, "/tournaments/t1/withdraw", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, uuid.New().String())
	req = req.WithContext(ctx)
	req = clubsChiParams(req, map[string]string{"clubID": "c1", "tournamentID": "t1"})

	rr := httptest.NewRecorder()
	h.WithdrawFromTournament(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// ── SubmitTournamentScore ────────────────────────────────────────────

func TestSubmitTournamentScore_Success(t *testing.T) {
	userID := uuid.New().String()
	clubID := uuid.New().String()
	tournamentID := uuid.New().String()
	sessionID := uuid.New().String()
	mock := &mockClubRepo{
		submitTournamentScoreFn: func(_ context.Context, _, _, _, _ string) (int, int, error) {
			return 680, 12, nil
		},
	}
	h := clubsHandler(mock)

	req := httptest.NewRequest(http.MethodPost, "/tournaments/"+tournamentID+"/submit-score?session_id="+sessionID, nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	req = req.WithContext(ctx)
	req = clubsChiParams(req, map[string]string{"clubID": clubID, "tournamentID": tournamentID})

	rr := httptest.NewRecorder()
	h.SubmitTournamentScore(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var result map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result["final_score"].(float64) != 680 {
		t.Errorf("expected final_score 680, got %v", result["final_score"])
	}
	if result["final_x_count"].(float64) != 12 {
		t.Errorf("expected final_x_count 12, got %v", result["final_x_count"])
	}
}

func TestSubmitTournamentScore_MissingSessionID(t *testing.T) {
	h := clubsHandler(&mockClubRepo{})

	req := httptest.NewRequest(http.MethodPost, "/tournaments/t1/submit-score", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, uuid.New().String())
	req = req.WithContext(ctx)
	req = clubsChiParams(req, map[string]string{"clubID": "c1", "tournamentID": "t1"})

	rr := httptest.NewRecorder()
	h.SubmitTournamentScore(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", rr.Code)
	}
}

func TestSubmitTournamentScore_NotFound(t *testing.T) {
	mock := &mockClubRepo{
		submitTournamentScoreFn: func(_ context.Context, _, _, _, _ string) (int, int, error) {
			return 0, 0, errors.New("not found")
		},
	}
	h := clubsHandler(mock)

	req := httptest.NewRequest(http.MethodPost, "/tournaments/t1/submit-score?session_id=s1", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, uuid.New().String())
	req = req.WithContext(ctx)
	req = clubsChiParams(req, map[string]string{"clubID": "c1", "tournamentID": "t1"})

	rr := httptest.NewRecorder()
	h.SubmitTournamentScore(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// ── Tournament Matchups Tests ────────────────────────────────────────

func TestGetTournamentMatchups_Success(t *testing.T) {
	roundID := uuid.New().String()
	mock := &mockClubRepo{
		getTournamentMatchupsFn: func(_ context.Context, rID string) ([]repository.TournamentMatchupOut, error) {
			return []repository.TournamentMatchupOut{
				{
					ID:          "m1",
					RoundID:     rID,
					MatchNumber: 1,
				},
			}, nil
		},
	}
	h := clubsHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/rounds/r1/matchups", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, uuid.New().String())
	req = req.WithContext(ctx)
	req = clubsChiParams(req, map[string]string{"clubID": "c1", "tournamentID": "t1", "roundID": roundID})

	rr := httptest.NewRecorder()
	h.GetTournamentMatchups(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var result []repository.TournamentMatchupOut
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(result) != 1 || result[0].ID != "m1" {
		t.Errorf("expected 1 matchup with ID 'm1', got %v", result)
	}
}

func TestSubmitTournamentMatchupScore_Success(t *testing.T) {
	userID := uuid.New().String()
	roundID := uuid.New().String()
	matchupID := uuid.New().String()
	sessionID := uuid.New().String()

	mock := &mockClubRepo{
		submitMatchupScoreFn: func(_ context.Context, clubID, tourneyID, rID, mID, uID, sID string) (*repository.TournamentMatchupOut, error) {
			score := 335
			return &repository.TournamentMatchupOut{
				ID:          mID,
				RoundID:     rID,
				MatchNumber: 1,
				ScoreA:      &score,
			}, nil
		},
	}
	h := clubsHandler(mock)

	req := httptest.NewRequest(http.MethodPost, "/rounds/r1/matchups/m1/submit-score?session_id="+sessionID, nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	req = req.WithContext(ctx)
	req = clubsChiParams(req, map[string]string{"clubID": "c1", "tournamentID": "t1", "roundID": roundID, "matchupID": matchupID})

	rr := httptest.NewRecorder()
	h.SubmitTournamentMatchupScore(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var result repository.TournamentMatchupOut
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result.ID != matchupID || result.ScoreA == nil || *result.ScoreA != 335 {
		t.Errorf("expected matchup score 335, got %v", result)
	}
}

func TestUpdateTournamentMatchup_Success(t *testing.T) {
	userID := uuid.New().String()
	roundID := uuid.New().String()
	matchupID := uuid.New().String()

	mock := &mockClubRepo{
		updateMatchupFn: func(_ context.Context, clubID, tourneyID, rID, mID, uID string, scoreA, scoreB *int, winnerID *string) (*repository.TournamentMatchupOut, error) {
			return &repository.TournamentMatchupOut{
				ID:          mID,
				RoundID:     rID,
				MatchNumber: 1,
				ScoreA:      scoreA,
				ScoreB:      scoreB,
				WinnerID:    winnerID,
			}, nil
		},
	}
	h := clubsHandler(mock)

	scoreA := 340
	scoreB := 338
	winnerID := "partA-id"

	req := clubsAuthedBody(http.MethodPut, "/rounds/r1/matchups/m1", userID, map[string]any{
		"score_a":   scoreA,
		"score_b":   scoreB,
		"winner_id": winnerID,
	})
	req = clubsChiParams(req, map[string]string{"clubID": "c1", "tournamentID": "t1", "roundID": roundID, "matchupID": matchupID})

	rr := httptest.NewRecorder()
	h.UpdateTournamentMatchup(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var result repository.TournamentMatchupOut
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result.ID != matchupID || result.ScoreA == nil || *result.ScoreA != 340 || result.WinnerID == nil || *result.WinnerID != "partA-id" {
		t.Errorf("unexpected matchup output: %v", result)
	}
}

