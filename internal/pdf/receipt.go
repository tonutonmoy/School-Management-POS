package pdf

import (
	"bytes"
	"fmt"

	"github.com/go-pdf/fpdf"
	qrcode "github.com/skip2/go-qrcode"

	"github.com/school-management/pos/internal/dto"
)

func GenerateReceipt(data *dto.ReceiptResponse, verifyURL string) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(0, 10, data.SchoolName, "", 1, "C", false, 0, "")
	pdf.SetFont("Arial", "", 11)
	pdf.CellFormat(0, 8, "Fee Payment Receipt", "", 1, "C", false, 0, "")
	pdf.Ln(4)

	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(40, 7, "Receipt No:")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 7, data.ReceiptNumber)
	pdf.Ln(7)
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(40, 7, "Payment No:")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 7, data.PaymentNumber)
	pdf.Ln(7)
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(40, 7, "Date:")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 7, data.IssuedAt.Format("2006-01-02 15:04"))
	pdf.Ln(10)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, "Student Information")
	pdf.Ln(8)
	pdf.SetFont("Arial", "", 10)
	lines := []string{
		fmt.Sprintf("Name: %s", data.StudentName),
		fmt.Sprintf("Admission No: %s", data.AdmissionNo),
		fmt.Sprintf("Class: %s / %s", data.ClassName, data.SectionName),
	}
	for _, line := range lines {
		pdf.Cell(0, 6, line)
		pdf.Ln(6)
	}
	pdf.Ln(4)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, "Payment Details")
	pdf.Ln(8)
	pdf.SetFont("Arial", "B", 9)
	pdf.CellFormat(90, 7, "Invoice", "1", 0, "L", false, 0, "")
	pdf.CellFormat(40, 7, "Amount", "1", 1, "R", false, 0, "")
	pdf.SetFont("Arial", "", 9)
	for _, a := range data.Allocations {
		pdf.CellFormat(90, 7, a.InvoiceNumber, "1", 0, "L", false, 0, "")
		pdf.CellFormat(40, 7, fmt.Sprintf("%.2f", a.Amount), "1", 1, "R", false, 0, "")
	}
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(90, 8, "Total Paid", "1", 0, "L", false, 0, "")
	pdf.CellFormat(40, 8, fmt.Sprintf("%.2f", data.TotalAmount), "1", 1, "R", false, 0, "")
	pdf.Ln(6)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 6, fmt.Sprintf("Collected by: %s", data.CollectorName))
	pdf.Ln(12)

	if verifyURL != "" {
		qrPNG, err := qrcode.Encode(verifyURL, qrcode.Medium, 120)
		if err == nil {
			opt := fpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}
			name := "qr.png"
			pdf.RegisterImageOptionsReader(name, opt, bytes.NewReader(qrPNG))
			pdf.ImageOptions(name, 15, pdf.GetY(), 30, 30, false, opt, 0, "")
			pdf.SetXY(50, pdf.GetY()+5)
			pdf.SetFont("Arial", "", 8)
			pdf.MultiCell(0, 4, "Scan QR to verify receipt authenticity", "", "L", false)
		}
	}

	pdf.SetY(-40)
	pdf.SetFont("Arial", "", 9)
	pdf.CellFormat(80, 6, "_______________________", "", 0, "C", false, 0, "")
	pdf.CellFormat(0, 6, "", "", 1, "C", false, 0, "")
	pdf.CellFormat(80, 6, "Authorized Signature", "", 0, "C", false, 0, "")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
