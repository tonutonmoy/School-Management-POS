package pdf

import (
	"bytes"
	"fmt"

	"github.com/go-pdf/fpdf"

	"github.com/school-management/pos/internal/dto"
)

func GenerateTabulation(examName, className string, results []dto.ExamResultResponse) ([]byte, error) {
	pdf := fpdf.New("L", "mm", "A4", "")
	pdf.SetMargins(10, 10, 10)
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(0, 10, "Tabulation Sheet", "", 1, "C", false, 0, "")
	pdf.SetFont("Arial", "", 11)
	pdf.CellFormat(0, 7, examName+" — "+className, "", 1, "C", false, 0, "")
	pdf.Ln(4)

	pdf.SetFont("Arial", "B", 8)
	pdf.CellFormat(8, 7, "#", "1", 0, "C", false, 0, "")
	pdf.CellFormat(35, 7, "Roll", "1", 0, "L", false, 0, "")
	pdf.CellFormat(55, 7, "Student", "1", 0, "L", false, 0, "")
	pdf.CellFormat(25, 7, "Section", "1", 0, "L", false, 0, "")
	pdf.CellFormat(22, 7, "Obtained", "1", 0, "R", false, 0, "")
	pdf.CellFormat(22, 7, "Full", "1", 0, "R", false, 0, "")
	pdf.CellFormat(18, 7, "%", "1", 0, "R", false, 0, "")
	pdf.CellFormat(15, 7, "GPA", "1", 0, "R", false, 0, "")
	pdf.CellFormat(12, 7, "Gr", "1", 0, "C", false, 0, "")
	pdf.CellFormat(15, 7, "Pos", "1", 0, "C", false, 0, "")
	pdf.CellFormat(18, 7, "Result", "1", 1, "C", false, 0, "")

	pdf.SetFont("Arial", "", 8)
	for i, r := range results {
		pos := "-"
		if r.MeritPosition != nil {
			pos = fmt.Sprintf("%d", *r.MeritPosition)
		}
		result := "Pass"
		if !r.IsPassed {
			result = "Fail"
		}
		pdf.CellFormat(8, 6, fmt.Sprintf("%d", i+1), "1", 0, "C", false, 0, "")
		pdf.CellFormat(35, 6, r.RollNumber, "1", 0, "L", false, 0, "")
		pdf.CellFormat(55, 6, r.StudentName, "1", 0, "L", false, 0, "")
		pdf.CellFormat(25, 6, r.SectionName, "1", 0, "L", false, 0, "")
		pdf.CellFormat(22, 6, fmt.Sprintf("%.1f", r.TotalObtained), "1", 0, "R", false, 0, "")
		pdf.CellFormat(22, 6, fmt.Sprintf("%.1f", r.TotalFull), "1", 0, "R", false, 0, "")
		pdf.CellFormat(18, 6, fmt.Sprintf("%.1f", r.Percentage), "1", 0, "R", false, 0, "")
		pdf.CellFormat(15, 6, fmt.Sprintf("%.2f", r.GPA), "1", 0, "R", false, 0, "")
		pdf.CellFormat(12, 6, r.Grade, "1", 0, "C", false, 0, "")
		pdf.CellFormat(15, 6, pos, "1", 0, "C", false, 0, "")
		pdf.CellFormat(18, 6, result, "1", 1, "C", false, 0, "")
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
