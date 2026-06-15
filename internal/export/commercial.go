package export

import (
	"bytes"
	"encoding/csv"
	"time"

	"github.com/school-management/pos/internal/dto"
)

func AuditLogsCSV(items []dto.ActivityItem) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"Time", "User", "Action", "Entity", "Description"})
	for _, r := range items {
		user := r.UserEmail
		if user == "" {
			user = "System"
		}
		_ = w.Write([]string{
			r.CreatedAt.Format(time.RFC3339), user, r.Action, r.EntityType, r.Description,
		})
	}
	w.Flush()
	return buf.Bytes(), w.Error()
}
