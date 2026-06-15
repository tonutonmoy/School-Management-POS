package pdf

import (
	"bytes"
	"fmt"

	"github.com/go-pdf/fpdf"

	"github.com/school-management/pos/internal/dto"
)

func GenerateTrialBalance(report *dto.TrialBalanceReport) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(0, 10, "Trial Balance", "", 1, "C", false, 0, "")
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(0, 7, fmt.Sprintf("As of %s", report.AsOf.Format("2006-01-02")), "", 1, "C", false, 0, "")
	pdf.Ln(4)
	pdf.SetFont("Arial", "B", 9)
	pdf.CellFormat(25, 7, "Code", "1", 0, "L", false, 0, "")
	pdf.CellFormat(70, 7, "Account", "1", 0, "L", false, 0, "")
	pdf.CellFormat(30, 7, "Debit", "1", 0, "R", false, 0, "")
	pdf.CellFormat(30, 7, "Credit", "1", 1, "R", false, 0, "")
	pdf.SetFont("Arial", "", 9)
	for _, r := range report.Rows {
		pdf.CellFormat(25, 6, r.AccountCode, "1", 0, "L", false, 0, "")
		pdf.CellFormat(70, 6, r.AccountName, "1", 0, "L", false, 0, "")
		pdf.CellFormat(30, 6, fmt.Sprintf("%.2f", r.Debit), "1", 0, "R", false, 0, "")
		pdf.CellFormat(30, 6, fmt.Sprintf("%.2f", r.Credit), "1", 1, "R", false, 0, "")
	}
	pdf.SetFont("Arial", "B", 9)
	pdf.CellFormat(95, 7, "TOTAL", "1", 0, "L", false, 0, "")
	pdf.CellFormat(30, 7, fmt.Sprintf("%.2f", report.TotalDebit), "1", 0, "R", false, 0, "")
	pdf.CellFormat(30, 7, fmt.Sprintf("%.2f", report.TotalCredit), "1", 1, "R", false, 0, "")
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func GenerateIncomeStatement(report *dto.IncomeStatementReport) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(0, 10, "Income Statement", "", 1, "C", false, 0, "")
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(0, 7, fmt.Sprintf("%s to %s", report.From.Format("2006-01-02"), report.To.Format("2006-01-02")), "", 1, "C", false, 0, "")
	pdf.Ln(6)
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, "Income")
	pdf.Ln(8)
	pdf.SetFont("Arial", "", 10)
	for _, r := range report.IncomeItems {
		pdf.CellFormat(120, 6, r.AccountName, "", 0, "L", false, 0, "")
		pdf.CellFormat(40, 6, fmt.Sprintf("%.2f", r.Credit), "", 1, "R", false, 0, "")
	}
	pdf.Ln(4)
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, "Expenses")
	pdf.Ln(8)
	pdf.SetFont("Arial", "", 10)
	for _, r := range report.ExpenseItems {
		pdf.CellFormat(120, 6, r.AccountName, "", 0, "L", false, 0, "")
		pdf.CellFormat(40, 6, fmt.Sprintf("%.2f", r.Debit), "", 1, "R", false, 0, "")
	}
	pdf.Ln(6)
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(120, 8, "Net Profit", "", 0, "L", false, 0, "")
	pdf.CellFormat(40, 8, fmt.Sprintf("%.2f", report.NetProfit), "", 1, "R", false, 0, "")
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
