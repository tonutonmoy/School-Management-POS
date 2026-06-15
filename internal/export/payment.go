package export

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"time"

	"github.com/school-management/pos/internal/dto"
)

func GatewayTransactionsCSV(items []dto.GatewayTransactionResponse) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"Ref", "Gateway", "Type", "Student", "Amount", "Status", "Receipt", "Created", "Completed"})
	for _, r := range items {
		completed := ""
		if r.CompletedAt != nil {
			completed = r.CompletedAt.Format(time.RFC3339)
		}
		_ = w.Write([]string{
			r.TransactionRef, r.GatewayName, r.PaymentType, r.StudentName,
			fmt.Sprintf("%.2f", r.Amount), r.Status, r.ReceiptNumber,
			r.CreatedAt.Format(time.RFC3339), completed,
		})
	}
	w.Flush()
	return buf.Bytes(), w.Error()
}
