package export

import (
	"bytes"
	"encoding/csv"
	"fmt"

	"github.com/xuri/excelize/v2"

	"github.com/school-management/pos/internal/dto"
)

func CollectionCSV(records []dto.PaymentResponse) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"Payment #", "Student", "Amount", "Method", "Date", "Collector"})
	for _, r := range records {
		_ = w.Write([]string{
			r.PaymentNumber, r.StudentName, fmt.Sprintf("%.2f", r.Amount),
			r.PaymentMethod, r.CollectionDate.Format("2006-01-02"), r.CollectorName,
		})
	}
	w.Flush()
	return buf.Bytes(), w.Error()
}

func BillsCSV(bills []dto.StudentBillResponse) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"Invoice", "Student", "Class", "Period", "Total", "Paid", "Due", "Status"})
	for _, b := range bills {
		_ = w.Write([]string{
			b.InvoiceNumber, b.StudentName, b.ClassName + "/" + b.SectionName, b.BillPeriod,
			fmt.Sprintf("%.2f", b.TotalAmount), fmt.Sprintf("%.2f", b.PaidAmount),
			fmt.Sprintf("%.2f", b.DueAmount), b.Status,
		})
	}
	w.Flush()
	return buf.Bytes(), w.Error()
}

func CollectionExcel(records []dto.PaymentResponse) ([]byte, error) {
	f := excelize.NewFile()
	sheet := "Collection"
	_ = f.SetSheetName("Sheet1", sheet)
	headers := []string{"Payment #", "Student", "Amount", "Method", "Date", "Collector"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
	}
	for row, r := range records {
		vals := []any{r.PaymentNumber, r.StudentName, r.Amount, r.PaymentMethod, r.CollectionDate.Format("2006-01-02"), r.CollectorName}
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
