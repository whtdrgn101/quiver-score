package pdf

import (
	"bytes"
	"fmt"

	"github.com/go-pdf/fpdf"

	"github.com/quiverscore/backend-go/internal/repository"
)

// GenerateSessionPDF renders a scoring session as a PDF document.
func GenerateSessionPDF(s *repository.SessionOut) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "Letter", "")
	pdf.SetTopMargin(12.7) // 0.5 inch
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()

	templateName := "Unknown"
	if s.Template != nil {
		templateName = s.Template.Name
	}

	// Title
	pdf.SetFont("Helvetica", "B", 18)
	pdf.CellFormat(0, 10, fmt.Sprintf("QuiverScore — %s", templateName), "", 1, "L", false, 0, "")
	pdf.Ln(4)

	// Info table
	labelW := 38.0 // ~1.5 inch
	valueW := 100.0 // ~4 inch
	lineH := 7.0

	infoRows := []struct{ label, value string }{
		{"Status", s.Status},
		{"Total Score", fmt.Sprintf("%d", s.TotalScore)},
		{"X Count", fmt.Sprintf("%d", s.TotalXCount)},
		{"Total Arrows", fmt.Sprintf("%d", s.TotalArrows)},
	}
	if s.Location != nil && *s.Location != "" {
		infoRows = append(infoRows, struct{ label, value string }{"Location", *s.Location})
	}
	if s.Weather != nil && *s.Weather != "" {
		infoRows = append(infoRows, struct{ label, value string }{"Weather", *s.Weather})
	}
	infoRows = append(infoRows, struct{ label, value string }{"Started", s.StartedAt.Format("2006-01-02 15:04")})
	if s.CompletedAt != nil {
		infoRows = append(infoRows, struct{ label, value string }{"Completed", s.CompletedAt.Format("2006-01-02 15:04")})
	}

	for _, row := range infoRows {
		pdf.SetFont("Helvetica", "B", 10)
		pdf.CellFormat(labelW, lineH, row.label, "", 0, "L", false, 0, "")
		pdf.SetFont("Helvetica", "", 10)
		pdf.CellFormat(valueW, lineH, row.value, "", 1, "L", false, 0, "")
	}
	pdf.Ln(6)

	// Scorecard
	if len(s.Ends) > 0 {
		pdf.SetFont("Helvetica", "B", 14)
		pdf.CellFormat(0, 8, "Scorecard", "", 1, "L", false, 0, "")
		pdf.Ln(2)

		maxArrows := 0
		for _, end := range s.Ends {
			if len(end.Arrows) > maxArrows {
				maxArrows = len(end.Arrows)
			}
		}

		// Column widths
		numCols := maxArrows + 2 // End + arrows + Total
		colW := 15.0
		pageW, _ := pdf.GetPageSize()
		lm, _, rm, _ := pdf.GetMargins()
		usable := pageW - lm - rm
		if float64(numCols)*colW > usable {
			colW = usable / float64(numCols)
		}

		cellH := 7.0

		// Header row
		pdf.SetFont("Helvetica", "B", 9)
		pdf.SetFillColor(5, 150, 105)   // #059669
		pdf.SetTextColor(255, 255, 255) // white
		pdf.SetDrawColor(128, 128, 128) // grey grid

		pdf.CellFormat(colW, cellH, "End", "1", 0, "C", true, 0, "")
		for i := 1; i <= maxArrows; i++ {
			pdf.CellFormat(colW, cellH, fmt.Sprintf("A%d", i), "1", 0, "C", true, 0, "")
		}
		pdf.CellFormat(colW, cellH, "Total", "1", 1, "C", true, 0, "")

		// Data rows
		pdf.SetFont("Helvetica", "", 9)
		pdf.SetTextColor(0, 0, 0)

		for rowIdx, end := range s.Ends {
			if rowIdx%2 == 1 {
				pdf.SetFillColor(240, 253, 244) // #f0fdf4
			} else {
				pdf.SetFillColor(255, 255, 255)
			}

			pdf.CellFormat(colW, cellH, fmt.Sprintf("%d", end.EndNumber), "1", 0, "C", true, 0, "")
			for _, arrow := range end.Arrows {
				pdf.CellFormat(colW, cellH, arrow.ScoreValue, "1", 0, "C", true, 0, "")
			}
			// Pad empty cells if this end has fewer arrows
			for j := len(end.Arrows); j < maxArrows; j++ {
				pdf.CellFormat(colW, cellH, "", "1", 0, "C", true, 0, "")
			}
			pdf.CellFormat(colW, cellH, fmt.Sprintf("%d", end.EndTotal), "1", 1, "C", true, 0, "")
		}
	}

	// Notes
	if s.Notes != nil && *s.Notes != "" {
		pdf.Ln(4)
		pdf.SetFont("Helvetica", "B", 12)
		pdf.CellFormat(0, 8, "Notes", "", 1, "L", false, 0, "")
		pdf.SetFont("Helvetica", "", 10)
		pdf.MultiCell(0, 5, *s.Notes, "", "L", false)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), pdf.Error()
}
