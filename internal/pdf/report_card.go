package pdf

import (
	"bytes"
	"fmt"

	"github.com/go-pdf/fpdf"

	"github.com/school-management/pos/internal/dto"
)

func GenerateReportCard(data *dto.ReportCardData) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 18)
	school := data.SchoolName
	if school == "" {
		school = "School Management System"
	}
	pdf.CellFormat(0, 10, school, "", 1, "C", false, 0, "")
	pdf.SetFont("Arial", "", 12)
	pdf.CellFormat(0, 8, "Academic Report Card", "", 1, "C", false, 0, "")
	pdf.Ln(4)

	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(0, 8, data.ExamName, "", 1, "C", false, 0, "")
	pdf.Ln(6)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, "Student Information")
	pdf.Ln(8)
	pdf.SetFont("Arial", "", 10)
	info := []string{
		fmt.Sprintf("Name: %s", data.StudentName),
		fmt.Sprintf("Admission No: %s", data.AdmissionNo),
		fmt.Sprintf("Roll No: %s", data.RollNumber),
		fmt.Sprintf("Class: %s / %s", data.ClassName, data.SectionName),
	}
	for _, line := range info {
		pdf.Cell(0, 6, line)
		pdf.Ln(6)
	}
	pdf.Ln(4)

	if data.Result != nil {
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(0, 8, "Subject Marks")
		pdf.Ln(8)
		pdf.SetFont("Arial", "B", 8)
		pdf.CellFormat(50, 7, "Subject", "1", 0, "L", false, 0, "")
		pdf.CellFormat(18, 7, "Written", "1", 0, "C", false, 0, "")
		pdf.CellFormat(18, 7, "MCQ", "1", 0, "C", false, 0, "")
		pdf.CellFormat(18, 7, "Prac.", "1", 0, "C", false, 0, "")
		pdf.CellFormat(18, 7, "Total", "1", 0, "C", false, 0, "")
		pdf.CellFormat(18, 7, "Full", "1", 0, "C", false, 0, "")
		pdf.CellFormat(18, 7, "Status", "1", 1, "C", false, 0, "")
		pdf.SetFont("Arial", "", 8)
		for _, subj := range data.Result.Subjects {
			status := "Pass"
			if !subj.IsPassed {
				status = "Fail"
			}
			pdf.CellFormat(50, 6, subj.SubjectName, "1", 0, "L", false, 0, "")
			pdf.CellFormat(18, 6, fmt.Sprintf("%.0f", subj.WrittenScore), "1", 0, "C", false, 0, "")
			pdf.CellFormat(18, 6, fmt.Sprintf("%.0f", subj.MCQScore), "1", 0, "C", false, 0, "")
			pdf.CellFormat(18, 6, fmt.Sprintf("%.0f", subj.PracticalScore), "1", 0, "C", false, 0, "")
			pdf.CellFormat(18, 6, fmt.Sprintf("%.0f", subj.TotalScore), "1", 0, "C", false, 0, "")
			pdf.CellFormat(18, 6, fmt.Sprintf("%.0f", subj.FullMarks), "1", 0, "C", false, 0, "")
			pdf.CellFormat(18, 6, status, "1", 1, "C", false, 0, "")
		}
		pdf.Ln(6)

		pdf.SetFont("Arial", "B", 11)
		pdf.Cell(45, 7, "Total Obtained:")
		pdf.SetFont("Arial", "", 10)
		pdf.Cell(0, 7, fmt.Sprintf("%.2f / %.2f (%.2f%%)", data.Result.TotalObtained, data.Result.TotalFull, data.Result.Percentage))
		pdf.Ln(7)
		pdf.SetFont("Arial", "B", 11)
		pdf.Cell(45, 7, "GPA / CGPA:")
		pdf.SetFont("Arial", "", 10)
		pdf.Cell(0, 7, fmt.Sprintf("%.2f / %.2f", data.Result.GPA, data.Result.CGPA))
		pdf.Ln(7)
		pdf.SetFont("Arial", "B", 11)
		pdf.Cell(45, 7, "Grade:")
		pdf.SetFont("Arial", "", 10)
		pdf.Cell(0, 7, data.Result.Grade)
		pdf.Ln(7)
		if data.Result.MeritPosition != nil {
			pdf.SetFont("Arial", "B", 11)
			pdf.Cell(45, 7, "Position:")
			pdf.SetFont("Arial", "", 10)
			pos := fmt.Sprintf("Merit #%d", *data.Result.MeritPosition)
			if data.Result.ClassPosition != nil {
				pos += fmt.Sprintf(" | Class #%d", *data.Result.ClassPosition)
			}
			if data.Result.SectionPosition != nil {
				pos += fmt.Sprintf(" | Section #%d", *data.Result.SectionPosition)
			}
			pdf.Cell(0, 7, pos)
			pdf.Ln(7)
		}
		resultText := "PASSED"
		if !data.Result.IsPassed {
			resultText = "FAILED"
		}
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(0, 8, "Overall Result: "+resultText)
		pdf.Ln(10)
	}

	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(0, 7, "Attendance Summary")
	pdf.Ln(7)
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 6, fmt.Sprintf("Attendance: %.1f%%", data.AttendancePct))
	pdf.Ln(20)

	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(90, 6, "", "", 0, "L", false, 0, "")
	pdf.CellFormat(90, 6, "_________________________", "", 1, "C", false, 0, "")
	pdf.CellFormat(90, 6, "", "", 0, "L", false, 0, "")
	pdf.CellFormat(90, 6, "Principal Signature", "", 1, "C", false, 0, "")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
