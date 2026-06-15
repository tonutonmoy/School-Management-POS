package export

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"time"

	"github.com/school-management/pos/internal/repository"
)

func SMSLogsCSV(recs []repository.SMSLogRecord) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"ID", "Phone", "Message", "Event", "Provider", "Status", "Sent At", "Created At"})
	for _, r := range recs {
		sent := ""
		if r.SentAt != nil {
			sent = r.SentAt.Format(time.RFC3339)
		}
		_ = w.Write([]string{
			r.ID.String(), r.RecipientPhone, r.Message, r.EventType,
			r.Provider, r.Status, sent, r.CreatedAt.Format(time.RFC3339),
		})
	}
	w.Flush()
	return buf.Bytes(), w.Error()
}

func FormatSMSExportFilename(from, to time.Time) string {
	return fmt.Sprintf("sms-logs-%s-%s.csv", from.Format("20060102"), to.Format("20060102"))
}
