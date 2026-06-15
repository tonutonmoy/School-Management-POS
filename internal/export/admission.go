package export

import (
	"bytes"
	"encoding/csv"
	"time"

	"github.com/school-management/pos/internal/repository"
)

func AdmissionApplicationsCSV(recs []repository.AdmissionRecord) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"Application No", "Status", "Name", "Email", "Phone", "Class", "Session", "Payment", "Created"})
	for _, r := range recs {
		_ = w.Write([]string{
			r.AppNumber, r.StatusVal, r.FirstName + " " + r.LastName, r.Email, r.Phone,
			r.ClassName, r.SessionName, r.PaymentStatus, r.CreatedAt.Format(time.RFC3339),
		})
	}
	w.Flush()
	return buf.Bytes(), w.Error()
}
