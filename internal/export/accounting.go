package export

import (
	"bytes"
	"encoding/csv"
	"fmt"

	"github.com/xuri/excelize/v2"

	"github.com/school-management/pos/internal/dto"
)

func TrialBalanceCSV(report *dto.TrialBalanceReport) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"Code", "Account", "Type", "Debit", "Credit"})
	for _, r := range report.Rows {
		_ = w.Write([]string{r.AccountCode, r.AccountName, r.AccountType, fmt.Sprintf("%.2f", r.Debit), fmt.Sprintf("%.2f", r.Credit)})
	}
	_ = w.Write([]string{"", "", "TOTAL", fmt.Sprintf("%.2f", report.TotalDebit), fmt.Sprintf("%.2f", report.TotalCredit)})
	w.Flush()
	return buf.Bytes(), w.Error()
}

func LedgerCSV(report *dto.LedgerReport) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"Date", "Entry #", "Description", "Debit", "Credit", "Balance"})
	for _, e := range report.Entries {
		_ = w.Write([]string{
			e.EntryDate.Format("2006-01-02"), e.EntryNumber, e.Description,
			fmt.Sprintf("%.2f", e.Debit), fmt.Sprintf("%.2f", e.Credit), fmt.Sprintf("%.2f", e.Balance),
		})
	}
	w.Flush()
	return buf.Bytes(), w.Error()
}

func TrialBalanceExcel(report *dto.TrialBalanceReport) ([]byte, error) {
	f := excelize.NewFile()
	sheet := "Trial Balance"
	_ = f.SetSheetName("Sheet1", sheet)
	headers := []string{"Code", "Account", "Type", "Debit", "Credit"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
	}
	for row, r := range report.Rows {
		vals := []any{r.AccountCode, r.AccountName, r.AccountType, r.Debit, r.Credit}
		for col, v := range vals {
			cell, _ := excelize.CoordinatesToCellName(col+1, row+2)
			_ = f.SetCellValue(sheet, cell, v)
		}
	}
	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
