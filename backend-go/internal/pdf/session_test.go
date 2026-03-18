package pdf

import (
	"testing"
	"time"

	"github.com/quiverscore/backend-go/internal/repository"
)

func strPtr(s string) *string { return &s }

func TestGenerateSessionPDF_FullSession(t *testing.T) {
	now := time.Now().UTC()
	completed := now.Add(30 * time.Minute)

	s := &repository.SessionOut{
		ID:          "test-id",
		TemplateID:  "tmpl-1",
		Status:      "completed",
		TotalScore:  285,
		TotalXCount: 3,
		TotalArrows: 30,
		Notes:       strPtr("Great session today"),
		Location:    strPtr("Indoor Range"),
		Weather:     strPtr("Clear"),
		StartedAt:   now,
		CompletedAt: &completed,
		Template: &repository.RoundTemplateOut{
			ID:   "tmpl-1",
			Name: "WA 18m Round (60 arrows)",
		},
		Ends: []repository.EndOut{
			{
				ID: "end-1", EndNumber: 1, EndTotal: 57,
				Arrows: []repository.ArrowOut{
					{ID: "a1", ArrowNumber: 1, ScoreValue: "X", ScoreNumeric: 10},
					{ID: "a2", ArrowNumber: 2, ScoreValue: "10", ScoreNumeric: 10},
					{ID: "a3", ArrowNumber: 3, ScoreValue: "9", ScoreNumeric: 9},
					{ID: "a4", ArrowNumber: 4, ScoreValue: "9", ScoreNumeric: 9},
					{ID: "a5", ArrowNumber: 5, ScoreValue: "10", ScoreNumeric: 10},
					{ID: "a6", ArrowNumber: 6, ScoreValue: "9", ScoreNumeric: 9},
				},
			},
			{
				ID: "end-2", EndNumber: 2, EndTotal: 52,
				Arrows: []repository.ArrowOut{
					{ID: "b1", ArrowNumber: 1, ScoreValue: "9", ScoreNumeric: 9},
					{ID: "b2", ArrowNumber: 2, ScoreValue: "8", ScoreNumeric: 8},
					{ID: "b3", ArrowNumber: 3, ScoreValue: "9", ScoreNumeric: 9},
					{ID: "b4", ArrowNumber: 4, ScoreValue: "9", ScoreNumeric: 9},
					{ID: "b5", ArrowNumber: 5, ScoreValue: "8", ScoreNumeric: 8},
					{ID: "b6", ArrowNumber: 6, ScoreValue: "9", ScoreNumeric: 9},
				},
			},
		},
	}

	data, err := GenerateSessionPDF(s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty PDF data")
	}
	if string(data[:5]) != "%PDF-" {
		t.Fatalf("expected PDF header, got %q", string(data[:5]))
	}
}

func TestGenerateSessionPDF_NoEnds(t *testing.T) {
	s := &repository.SessionOut{
		ID:          "test-id",
		Status:      "in_progress",
		TotalScore:  0,
		TotalXCount: 0,
		TotalArrows: 0,
		StartedAt:   time.Now().UTC(),
		Template: &repository.RoundTemplateOut{
			ID:   "tmpl-1",
			Name: "Practice Round",
		},
		Ends: []repository.EndOut{},
	}

	data, err := GenerateSessionPDF(s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty PDF data")
	}
}

func TestGenerateSessionPDF_NilOptionalFields(t *testing.T) {
	s := &repository.SessionOut{
		ID:          "test-id",
		Status:      "in_progress",
		TotalScore:  0,
		TotalXCount: 0,
		TotalArrows: 0,
		StartedAt:   time.Now().UTC(),
		Template:    nil, // no template
		Notes:       nil,
		Location:    nil,
		Weather:     nil,
		CompletedAt: nil,
		Ends:        nil,
	}

	data, err := GenerateSessionPDF(s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data[:5]) != "%PDF-" {
		t.Fatalf("expected PDF header, got %q", string(data[:5]))
	}
}
