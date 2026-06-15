package export

import (
	"bytes"
	"encoding/csv"
	"fmt"

	"github.com/xuri/excelize/v2"

	"github.com/school-management/pos/internal/dto"
)

func ResultsCSV(results []dto.ExamResultResponse) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"Roll", "Student", "Class", "Section", "Obtained", "Full", "Percentage", "GPA", "CGPA", "Grade", "Merit", "Result"})
	for _, r := range results {
		pos := ""
		if r.MeritPosition != nil {
			pos = fmt.Sprintf("%d", *r.MeritPosition)
		}
		result := "Pass"
		if !r.IsPassed {
			result = "Fail"
		}
		_ = w.Write([]string{
			r.RollNumber, r.StudentName, r.ClassName, r.SectionName,
			fmt.Sprintf("%.2f", r.TotalObtained), fmt.Sprintf("%.2f", r.TotalFull),
			fmt.Sprintf("%.2f", r.Percentage), fmt.Sprintf("%.2f", r.GPA), fmt.Sprintf("%.2f", r.CGPA),
			r.Grade, pos, result,
		})
	}
	w.Flush()
	return buf.Bytes(), w.Error()
}

func TabulationExcel(examName string, results []dto.ExamResultResponse) ([]byte, error) {
	f := excelize.NewFile()
	sheet := "Tabulation"
	_ = f.SetSheetName("Sheet1", sheet)
	_ = f.SetCellValue(sheet, "A1", examName)
	headers := []string{"Roll", "Student", "Section", "Obtained", "Full", "Percentage", "GPA", "Grade", "Merit", "Result"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 3)
		_ = f.SetCellValue(sheet, cell, h)
	}
	for row, r := range results {
		pos := ""
		if r.MeritPosition != nil {
			pos = fmt.Sprintf("%d", *r.MeritPosition)
		}
		result := "Pass"
		if !r.IsPassed {
			result = "Fail"
		}
		vals := []any{r.RollNumber, r.StudentName, r.SectionName, r.TotalObtained, r.TotalFull, r.Percentage, r.GPA, r.Grade, pos, result}
		for col, v := range vals {
			cell, _ := excelize.CoordinatesToCellName(col+1, row+4)
			_ = f.SetCellValue(sheet, cell, v)
		}
	}
	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func SubjectPerformanceCSV(items []dto.SubjectPerfPoint) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"Subject", "Average Score", "Pass Rate %"})
	for _, it := range items {
		_ = w.Write([]string{it.SubjectName, fmt.Sprintf("%.2f", it.AvgScore), fmt.Sprintf("%.1f", it.PassRate)})
	}
	w.Flush()
	return buf.Bytes(), w.Error()
}
