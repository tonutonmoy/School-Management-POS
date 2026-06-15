package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ParentRepository interface {
	Create(ctx context.Context, userID uuid.UUID, phone, address, occupation string) (*ParentRecord, error)
	Update(ctx context.Context, id uuid.UUID, phone, address, occupation string) (*ParentRecord, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*ParentRecord, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) (*ParentRecord, error)
	List(ctx context.Context, limit, offset int32) ([]ParentListRecord, error)
	Count(ctx context.Context) (int64, error)
	LinkStudent(ctx context.Context, parentID, studentID uuid.UUID, relationship string, isPrimary bool) error
	UnlinkStudent(ctx context.Context, parentID, studentID uuid.UUID) error
	ListStudents(ctx context.Context, parentID uuid.UUID) ([]ParentStudentRecord, error)
	HasStudentAccess(ctx context.Context, userID, studentID uuid.UUID) (bool, error)
	ListParentsForStudent(ctx context.Context, studentID uuid.UUID) ([]ParentRecord, error)
	RecentActivities(ctx context.Context, parentID uuid.UUID, limit int32) ([]ParentActivityRecord, error)
}

type ParentRecord struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	Phone      string
	Address    string
	Occupation string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type ParentListRecord struct {
	ParentRecord
	Email      string
	FirstName  string
	LastName   string
	IsActive   bool
	ChildCount int
}

type ParentStudentRecord struct {
	ID              uuid.UUID
	ParentID        uuid.UUID
	StudentID       uuid.UUID
	Relationship    string
	IsPrimary       bool
	FirstName       string
	LastName        string
	AdmissionNumber string
	RollNumber      string
	ClassName       string
	CreatedAt       time.Time
}

type ParentActivityRecord struct {
	ID          uuid.UUID
	Category    string
	Title       string
	Description string
	StudentName string
	CreatedAt   time.Time
}

type parentRepo struct {
	pool *pgxpool.Pool
}

func NewParentRepository(pool *pgxpool.Pool) ParentRepository {
	return &parentRepo{pool: pool}
}

func (r *parentRepo) Create(ctx context.Context, userID uuid.UUID, phone, address, occupation string) (*ParentRecord, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO parents (user_id, phone, address, occupation)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, phone, address, occupation, created_at, updated_at`,
		userID, phone, address, occupation)
	return scanParent(row)
}

func (r *parentRepo) Update(ctx context.Context, id uuid.UUID, phone, address, occupation string) (*ParentRecord, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE parents SET phone = $2, address = $3, occupation = $4, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, user_id, phone, address, occupation, created_at, updated_at`,
		id, phone, address, occupation)
	return scanParent(row)
}

func (r *parentRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE parents SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	return err
}

func (r *parentRepo) GetByID(ctx context.Context, id uuid.UUID) (*ParentRecord, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, phone, address, occupation, created_at, updated_at
		FROM parents WHERE id = $1 AND deleted_at IS NULL`, id)
	rec, err := scanParent(row)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return rec, err
}

func (r *parentRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*ParentRecord, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, phone, address, occupation, created_at, updated_at
		FROM parents WHERE user_id = $1 AND deleted_at IS NULL`, userID)
	rec, err := scanParent(row)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return rec, err
}

func (r *parentRepo) List(ctx context.Context, limit, offset int32) ([]ParentListRecord, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT p.id, p.user_id, p.phone, p.address, p.occupation, p.created_at, p.updated_at,
		       u.email, u.first_name, u.last_name, u.is_active,
		       (SELECT COUNT(*)::int FROM parent_students ps WHERE ps.parent_id = p.id)
		FROM parents p
		JOIN users u ON u.id = p.user_id AND u.deleted_at IS NULL
		WHERE p.deleted_at IS NULL
		ORDER BY u.last_name, u.first_name
		LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ParentListRecord
	for rows.Next() {
		var rec ParentListRecord
		if err := rows.Scan(&rec.ID, &rec.UserID, &rec.Phone, &rec.Address, &rec.Occupation,
			&rec.CreatedAt, &rec.UpdatedAt, &rec.Email, &rec.FirstName, &rec.LastName, &rec.IsActive, &rec.ChildCount); err != nil {
			return nil, err
		}
		items = append(items, rec)
	}
	return items, rows.Err()
}

func (r *parentRepo) Count(ctx context.Context) (int64, error) {
	var n int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM parents WHERE deleted_at IS NULL`).Scan(&n)
	return n, err
}

func (r *parentRepo) LinkStudent(ctx context.Context, parentID, studentID uuid.UUID, relationship string, isPrimary bool) error {
	if relationship == "" {
		relationship = "guardian"
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO parent_students (parent_id, student_id, relationship, is_primary)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (parent_id, student_id) DO UPDATE SET relationship = EXCLUDED.relationship, is_primary = EXCLUDED.is_primary`,
		parentID, studentID, relationship, isPrimary)
	return err
}

func (r *parentRepo) UnlinkStudent(ctx context.Context, parentID, studentID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM parent_students WHERE parent_id = $1 AND student_id = $2`, parentID, studentID)
	return err
}

func (r *parentRepo) ListStudents(ctx context.Context, parentID uuid.UUID) ([]ParentStudentRecord, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT ps.id, ps.parent_id, ps.student_id, ps.relationship, ps.is_primary, ps.created_at,
		       s.first_name, s.last_name, s.admission_number, COALESCE(s.roll_number, ''), COALESCE(c.name, '')
		FROM parent_students ps
		JOIN students s ON s.id = ps.student_id AND s.deleted_at IS NULL
		LEFT JOIN classes c ON c.id = s.class_id
		WHERE ps.parent_id = $1
		ORDER BY ps.is_primary DESC, s.first_name`, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ParentStudentRecord
	for rows.Next() {
		var rec ParentStudentRecord
		if err := rows.Scan(&rec.ID, &rec.ParentID, &rec.StudentID, &rec.Relationship, &rec.IsPrimary, &rec.CreatedAt,
			&rec.FirstName, &rec.LastName, &rec.AdmissionNumber, &rec.RollNumber, &rec.ClassName); err != nil {
			return nil, err
		}
		items = append(items, rec)
	}
	return items, rows.Err()
}

func (r *parentRepo) HasStudentAccess(ctx context.Context, userID, studentID uuid.UUID) (bool, error) {
	var ok bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM parent_students ps
			JOIN parents p ON p.id = ps.parent_id AND p.deleted_at IS NULL
			WHERE p.user_id = $1 AND ps.student_id = $2
		)`, userID, studentID).Scan(&ok)
	return ok, err
}

func (r *parentRepo) ListParentsForStudent(ctx context.Context, studentID uuid.UUID) ([]ParentRecord, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT p.id, p.user_id, p.phone, p.address, p.occupation, p.created_at, p.updated_at
		FROM parents p
		JOIN parent_students ps ON ps.parent_id = p.id
		WHERE ps.student_id = $1 AND p.deleted_at IS NULL`, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ParentRecord
	for rows.Next() {
		var rec ParentRecord
		if err := rows.Scan(&rec.ID, &rec.UserID, &rec.Phone, &rec.Address, &rec.Occupation, &rec.CreatedAt, &rec.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, rec)
	}
	return items, rows.Err()
}

func (r *parentRepo) RecentActivities(ctx context.Context, parentID uuid.UUID, limit int32) ([]ParentActivityRecord, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, category, title, description, student_name, created_at FROM (
			SELECT p.id, 'payment' AS category, 'Fee Payment' AS title,
			       'Payment ' || p.payment_number || ' — ৳' || ROUND(p.amount::numeric, 2) AS description,
			       s.first_name || ' ' || s.last_name AS student_name, p.created_at
			FROM payments p
			JOIN parent_students ps ON ps.student_id = p.student_id AND ps.parent_id = $1
			JOIN students s ON s.id = p.student_id
			WHERE p.status = 'completed'
			UNION ALL
			SELECT sa.id, 'attendance', 'Attendance Update',
			       'Marked ' || sa.status || ' on ' || TO_CHAR(sa.attendance_date, 'YYYY-MM-DD'),
			       s.first_name || ' ' || s.last_name, sa.updated_at
			FROM student_attendance sa
			JOIN parent_students ps ON ps.student_id = sa.student_id AND ps.parent_id = $1
			JOIN students s ON s.id = sa.student_id
			UNION ALL
			SELECT er.id, 'result', 'Result Published',
			       e.name || ' — GPA ' || ROUND(er.gpa::numeric, 2),
			       s.first_name || ' ' || s.last_name, er.updated_at
			FROM exam_results er
			JOIN exams e ON e.id = er.exam_id
			JOIN parent_students ps ON ps.student_id = er.student_id AND ps.parent_id = $1
			JOIN students s ON s.id = er.student_id
			WHERE er.status = 'published'
		) combined
		ORDER BY created_at DESC
		LIMIT $2`, parentID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ParentActivityRecord
	for rows.Next() {
		var rec ParentActivityRecord
		if err := rows.Scan(&rec.ID, &rec.Category, &rec.Title, &rec.Description, &rec.StudentName, &rec.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, rec)
	}
	return items, rows.Err()
}

func scanParent(row pgx.Row) (*ParentRecord, error) {
	var rec ParentRecord
	if err := row.Scan(&rec.ID, &rec.UserID, &rec.Phone, &rec.Address, &rec.Occupation, &rec.CreatedAt, &rec.UpdatedAt); err != nil {
		return nil, err
	}
	return &rec, nil
}

type NoticeRepository interface {
	Create(ctx context.Context, params NoticeParams) (*NoticeRecord, error)
	Update(ctx context.Context, id uuid.UUID, params NoticeParams) (*NoticeRecord, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*NoticeRecord, error)
	Search(ctx context.Context, f NoticeSearchParams) ([]NoticeRecord, error)
	Count(ctx context.Context, f NoticeSearchParams) (int64, error)
	MarkRead(ctx context.Context, noticeID, parentID uuid.UUID) error
	CountUnreadForParent(ctx context.Context, parentID uuid.UUID) (int64, error)
}

type NoticeParams struct {
	Title          string
	Body           string
	NoticeType     string
	TargetAudience string
	PublishAt      time.Time
	ExpiresAt      *time.Time
	IsPublished    bool
	CreatedBy      *uuid.UUID
}

type NoticeSearchParams struct {
	Query      string
	NoticeType string
	ParentID   *uuid.UUID
	Published  *bool
	Limit      int32
	Offset     int32
}

type NoticeRecord struct {
	ID             uuid.UUID
	Title          string
	Body           string
	NoticeType     string
	TargetAudience string
	PublishAt      time.Time
	ExpiresAt      *time.Time
	IsPublished    bool
	CreatedBy      *uuid.UUID
	CreatedByName  string
	IsRead         bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type noticeRepo struct {
	pool *pgxpool.Pool
}

func NewNoticeRepository(pool *pgxpool.Pool) NoticeRepository {
	return &noticeRepo{pool: pool}
}

func (r *noticeRepo) Create(ctx context.Context, params NoticeParams) (*NoticeRecord, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO notices (title, body, notice_type, target_audience, publish_at, expires_at, is_published, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, title, body, notice_type, target_audience, publish_at, expires_at, is_published, created_by, created_at, updated_at`,
		params.Title, params.Body, params.NoticeType, params.TargetAudience, params.PublishAt, params.ExpiresAt, params.IsPublished, params.CreatedBy)
	return scanNotice(row, false)
}

func (r *noticeRepo) Update(ctx context.Context, id uuid.UUID, params NoticeParams) (*NoticeRecord, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE notices SET title = $2, body = $3, notice_type = $4, target_audience = $5,
			publish_at = $6, expires_at = $7, is_published = $8, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, title, body, notice_type, target_audience, publish_at, expires_at, is_published, created_by, created_at, updated_at`,
		id, params.Title, params.Body, params.NoticeType, params.TargetAudience, params.PublishAt, params.ExpiresAt, params.IsPublished)
	rec, err := scanNotice(row, false)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return rec, err
}

func (r *noticeRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE notices SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	return err
}

func (r *noticeRepo) GetByID(ctx context.Context, id uuid.UUID) (*NoticeRecord, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT n.id, n.title, n.body, n.notice_type, n.target_audience, n.publish_at, n.expires_at,
		       n.is_published, n.created_by, COALESCE(u.first_name || ' ' || u.last_name, ''), false,
		       n.created_at, n.updated_at
		FROM notices n
		LEFT JOIN users u ON u.id = n.created_by
		WHERE n.id = $1 AND n.deleted_at IS NULL`, id)
	rec, err := scanNoticeFull(row)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return rec, err
}

func (r *noticeRepo) Search(ctx context.Context, f NoticeSearchParams) ([]NoticeRecord, error) {
	q, args := buildNoticeQuery(f, false)
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []NoticeRecord
	for rows.Next() {
		var rec NoticeRecord
		if err := rows.Scan(&rec.ID, &rec.Title, &rec.Body, &rec.NoticeType, &rec.TargetAudience,
			&rec.PublishAt, &rec.ExpiresAt, &rec.IsPublished, &rec.CreatedBy, &rec.CreatedByName, &rec.IsRead,
			&rec.CreatedAt, &rec.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, rec)
	}
	return items, rows.Err()
}

func (r *noticeRepo) Count(ctx context.Context, f NoticeSearchParams) (int64, error) {
	q, args := buildNoticeQuery(f, true)
	var n int64
	err := r.pool.QueryRow(ctx, q, args...).Scan(&n)
	return n, err
}

func (r *noticeRepo) MarkRead(ctx context.Context, noticeID, parentID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO notice_reads (notice_id, parent_id) VALUES ($1, $2)
		ON CONFLICT (notice_id, parent_id) DO NOTHING`, noticeID, parentID)
	return err
}

func (r *noticeRepo) CountUnreadForParent(ctx context.Context, parentID uuid.UUID) (int64, error) {
	var n int64
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM notices n
		WHERE n.deleted_at IS NULL AND n.is_published = true
		AND (n.expires_at IS NULL OR n.expires_at > NOW())
		AND n.target_audience IN ('all_parents', 'all_users')
		AND NOT EXISTS (SELECT 1 FROM notice_reads nr WHERE nr.notice_id = n.id AND nr.parent_id = $1)`,
		parentID).Scan(&n)
	return n, err
}

func buildNoticeQuery(f NoticeSearchParams, countOnly bool) (string, []any) {
	var sb strings.Builder
	args := []any{}
	idx := 1
	if countOnly {
		sb.WriteString(`SELECT COUNT(*) FROM notices n WHERE n.deleted_at IS NULL`)
	} else {
		readExpr := "false"
		if f.ParentID != nil {
			readExpr = fmt.Sprintf("EXISTS(SELECT 1 FROM notice_reads nr WHERE nr.notice_id = n.id AND nr.parent_id = $%d)", idx)
			args = append(args, *f.ParentID)
			idx++
		}
		sb.WriteString(fmt.Sprintf(`
			SELECT n.id, n.title, n.body, n.notice_type, n.target_audience, n.publish_at, n.expires_at,
			       n.is_published, n.created_by, COALESCE(u.first_name || ' ' || u.last_name, ''), %s,
			       n.created_at, n.updated_at
			FROM notices n
			LEFT JOIN users u ON u.id = n.created_by
			WHERE n.deleted_at IS NULL`, readExpr))
	}
	if f.Query != "" {
		sb.WriteString(fmt.Sprintf(" AND (n.title ILIKE $%d OR n.body ILIKE $%d)", idx, idx))
		args = append(args, "%"+f.Query+"%")
		idx++
	}
	if f.NoticeType != "" {
		sb.WriteString(fmt.Sprintf(" AND n.notice_type = $%d", idx))
		args = append(args, f.NoticeType)
		idx++
	}
	if f.Published != nil {
		sb.WriteString(fmt.Sprintf(" AND n.is_published = $%d", idx))
		args = append(args, *f.Published)
		idx++
	}
	if f.ParentID != nil && !countOnly {
		sb.WriteString(fmt.Sprintf(" AND n.is_published = true AND n.target_audience IN ('all_parents', 'all_users')"))
	}
	if !countOnly {
		sb.WriteString(" ORDER BY n.publish_at DESC")
		if f.Limit > 0 {
			sb.WriteString(fmt.Sprintf(" LIMIT $%d OFFSET $%d", idx, idx+1))
			args = append(args, f.Limit, f.Offset)
		}
	}
	return sb.String(), args
}

func scanNotice(row pgx.Row, isRead bool) (*NoticeRecord, error) {
	var rec NoticeRecord
	var createdBy *uuid.UUID
	if err := row.Scan(&rec.ID, &rec.Title, &rec.Body, &rec.NoticeType, &rec.TargetAudience,
		&rec.PublishAt, &rec.ExpiresAt, &rec.IsPublished, &createdBy, &rec.CreatedAt, &rec.UpdatedAt); err != nil {
		return nil, err
	}
	rec.CreatedBy = createdBy
	rec.IsRead = isRead
	return &rec, nil
}

func scanNoticeFull(row pgx.Row) (*NoticeRecord, error) {
	var rec NoticeRecord
	var createdBy *uuid.UUID
	if err := row.Scan(&rec.ID, &rec.Title, &rec.Body, &rec.NoticeType, &rec.TargetAudience,
		&rec.PublishAt, &rec.ExpiresAt, &rec.IsPublished, &createdBy, &rec.CreatedByName, &rec.IsRead,
		&rec.CreatedAt, &rec.UpdatedAt); err != nil {
		return nil, err
	}
	rec.CreatedBy = createdBy
	return &rec, nil
}

type NotificationRepository interface {
	CreateInApp(ctx context.Context, params InAppNotificationParams) (*NotificationRecord, error)
	ListForParent(ctx context.Context, parentID uuid.UUID, unreadOnly bool, limit, offset int32) ([]NotificationRecord, error)
	CountForParent(ctx context.Context, parentID uuid.UUID, unreadOnly bool) (int64, error)
	MarkRead(ctx context.Context, id, parentID uuid.UUID) error
	MarkAllRead(ctx context.Context, parentID uuid.UUID) error
	CreateSMSLog(ctx context.Context, params SMSLogParams) (*SMSLogRecord, error)
	UpdateSMSStatus(ctx context.Context, id uuid.UUID, status string, sentAt *time.Time, errMsg string) error
	CreateEmailLog(ctx context.Context, params EmailLogParams) (*EmailLogRecord, error)
	UpdateEmailStatus(ctx context.Context, id uuid.UUID, status string, sentAt *time.Time, errMsg string) error
	Enqueue(ctx context.Context, channel string, payload map[string]any) error
	CommunicationStats(ctx context.Context) (*CommunicationStatsRecord, error)
	ListRecentSMS(ctx context.Context, limit int32) ([]SMSLogRecord, error)
	ListRecentEmails(ctx context.Context, limit int32) ([]EmailLogRecord, error)
	CountInAppToday(ctx context.Context) (int64, error)
	ListSMSForExport(ctx context.Context, from, to time.Time) ([]SMSLogRecord, error)
}

type InAppNotificationParams struct {
	ParentID      *uuid.UUID
	UserID        *uuid.UUID
	Title         string
	Body          string
	Category      string
	ReferenceType string
	ReferenceID   *uuid.UUID
}

type SMSLogParams struct {
	RecipientPhone string
	Message        string
	EventType      string
	Provider       string
	Status         string
	ReferenceType  string
	ReferenceID    *uuid.UUID
	ParentID       *uuid.UUID
}

type EmailLogParams struct {
	RecipientEmail string
	Subject        string
	Body           string
	EventType      string
	Provider       string
	Status         string
	ReferenceType  string
	ReferenceID    *uuid.UUID
	ParentID       *uuid.UUID
}

type NotificationRecord struct {
	ID            uuid.UUID
	Title         string
	Body          string
	Category      string
	ReferenceType string
	ReferenceID   *uuid.UUID
	IsRead        bool
	ReadAt        *time.Time
	CreatedAt     time.Time
}

type SMSLogRecord struct {
	ID             uuid.UUID
	RecipientPhone string
	Message        string
	EventType      string
	Provider       string
	Status         string
	SentAt         *time.Time
	CreatedAt      time.Time
}

type EmailLogRecord struct {
	ID             uuid.UUID
	RecipientEmail string
	Subject        string
	EventType      string
	Provider       string
	Status         string
	SentAt         *time.Time
	CreatedAt      time.Time
}

type CommunicationStatsRecord struct {
	SMSToday        int64
	EmailsToday     int64
	Notifications   int64
	DeliveryRate    float64
}

type notificationRepo struct {
	pool *pgxpool.Pool
}

func NewNotificationRepository(pool *pgxpool.Pool) NotificationRepository {
	return &notificationRepo{pool: pool}
}

func (r *notificationRepo) CreateInApp(ctx context.Context, params InAppNotificationParams) (*NotificationRecord, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO notifications (parent_id, user_id, title, body, category, reference_type, reference_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, title, body, category, reference_type, reference_id, is_read, read_at, created_at`,
		params.ParentID, params.UserID, params.Title, params.Body, params.Category, params.ReferenceType, params.ReferenceID)
	var rec NotificationRecord
	var refID *uuid.UUID
	if err := row.Scan(&rec.ID, &rec.Title, &rec.Body, &rec.Category, &rec.ReferenceType, &refID, &rec.IsRead, &rec.ReadAt, &rec.CreatedAt); err != nil {
		return nil, err
	}
	rec.ReferenceID = refID
	return &rec, nil
}

func (r *notificationRepo) ListForParent(ctx context.Context, parentID uuid.UUID, unreadOnly bool, limit, offset int32) ([]NotificationRecord, error) {
	q := `
		SELECT id, title, body, category, reference_type, reference_id, is_read, read_at, created_at
		FROM notifications WHERE parent_id = $1`
	args := []any{parentID}
	if unreadOnly {
		q += ` AND is_read = false`
	}
	q += ` ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	args = append(args, limit, offset)
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanNotifications(rows)
}

func (r *notificationRepo) CountForParent(ctx context.Context, parentID uuid.UUID, unreadOnly bool) (int64, error) {
	q := `SELECT COUNT(*) FROM notifications WHERE parent_id = $1`
	if unreadOnly {
		q += ` AND is_read = false`
	}
	var n int64
	err := r.pool.QueryRow(ctx, q, parentID).Scan(&n)
	return n, err
}

func (r *notificationRepo) MarkRead(ctx context.Context, id, parentID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE notifications SET is_read = true, read_at = NOW()
		WHERE id = $1 AND parent_id = $2`, id, parentID)
	return err
}

func (r *notificationRepo) MarkAllRead(ctx context.Context, parentID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE notifications SET is_read = true, read_at = NOW()
		WHERE parent_id = $1 AND is_read = false`, parentID)
	return err
}

func (r *notificationRepo) CreateSMSLog(ctx context.Context, params SMSLogParams) (*SMSLogRecord, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO sms_logs (recipient_phone, message, event_type, provider, status, reference_type, reference_id, parent_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, recipient_phone, message, event_type, provider, status, sent_at, created_at`,
		params.RecipientPhone, params.Message, params.EventType, params.Provider, params.Status,
		params.ReferenceType, params.ReferenceID, params.ParentID)
	return scanSMSLog(row)
}

func (r *notificationRepo) UpdateSMSStatus(ctx context.Context, id uuid.UUID, status string, sentAt *time.Time, errMsg string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE sms_logs SET status = $2, sent_at = $3, error_message = $4 WHERE id = $1`, id, status, sentAt, errMsg)
	return err
}

func (r *notificationRepo) CreateEmailLog(ctx context.Context, params EmailLogParams) (*EmailLogRecord, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO email_logs (recipient_email, subject, body, event_type, provider, status, reference_type, reference_id, parent_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, recipient_email, subject, event_type, provider, status, sent_at, created_at`,
		params.RecipientEmail, params.Subject, params.Body, params.EventType, params.Provider, params.Status,
		params.ReferenceType, params.ReferenceID, params.ParentID)
	return scanEmailLog(row)
}

func (r *notificationRepo) UpdateEmailStatus(ctx context.Context, id uuid.UUID, status string, sentAt *time.Time, errMsg string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE email_logs SET status = $2, sent_at = $3, error_message = $4 WHERE id = $1`, id, status, sentAt, errMsg)
	return err
}

func (r *notificationRepo) Enqueue(ctx context.Context, channel string, payload map[string]any) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `
		INSERT INTO notification_queue (channel, payload, status) VALUES ($1, $2, 'pending')`, channel, b)
	return err
}

func (r *notificationRepo) CommunicationStats(ctx context.Context) (*CommunicationStatsRecord, error) {
	var rec CommunicationStatsRecord
	err := r.pool.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM sms_logs WHERE created_at >= CURRENT_DATE),
			(SELECT COUNT(*) FROM email_logs WHERE created_at >= CURRENT_DATE),
			(SELECT COUNT(*) FROM notifications WHERE created_at >= CURRENT_DATE),
			COALESCE((
				SELECT CASE WHEN COUNT(*) = 0 THEN 0
				ELSE ROUND(100.0 * COUNT(*) FILTER (WHERE status = 'sent') / COUNT(*), 2) END
				FROM sms_logs WHERE created_at >= CURRENT_DATE - INTERVAL '30 days'
			), 0)`).Scan(&rec.SMSToday, &rec.EmailsToday, &rec.Notifications, &rec.DeliveryRate)
	return &rec, err
}

func (r *notificationRepo) ListRecentSMS(ctx context.Context, limit int32) ([]SMSLogRecord, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, recipient_phone, message, event_type, provider, status, sent_at, created_at
		FROM sms_logs ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []SMSLogRecord
	for rows.Next() {
		rec, err := scanSMSLogRow(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *rec)
	}
	return items, rows.Err()
}

func (r *notificationRepo) ListRecentEmails(ctx context.Context, limit int32) ([]EmailLogRecord, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, recipient_email, subject, event_type, provider, status, sent_at, created_at
		FROM email_logs ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []EmailLogRecord
	for rows.Next() {
		rec, err := scanEmailLogRow(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *rec)
	}
	return items, rows.Err()
}

func (r *notificationRepo) CountInAppToday(ctx context.Context) (int64, error) {
	var n int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM notifications WHERE created_at >= CURRENT_DATE`).Scan(&n)
	return n, err
}

func (r *notificationRepo) ListSMSForExport(ctx context.Context, from, to time.Time) ([]SMSLogRecord, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, recipient_phone, message, event_type, provider, status, sent_at, created_at
		FROM sms_logs WHERE created_at >= $1 AND created_at < $2 ORDER BY created_at`, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []SMSLogRecord
	for rows.Next() {
		rec, err := scanSMSLogRow(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *rec)
	}
	return items, rows.Err()
}

func scanNotifications(rows pgx.Rows) ([]NotificationRecord, error) {
	var items []NotificationRecord
	for rows.Next() {
		var rec NotificationRecord
		var refID *uuid.UUID
		if err := rows.Scan(&rec.ID, &rec.Title, &rec.Body, &rec.Category, &rec.ReferenceType, &refID, &rec.IsRead, &rec.ReadAt, &rec.CreatedAt); err != nil {
			return nil, err
		}
		rec.ReferenceID = refID
		items = append(items, rec)
	}
	return items, rows.Err()
}

func scanSMSLog(row pgx.Row) (*SMSLogRecord, error) {
	var rec SMSLogRecord
	if err := row.Scan(&rec.ID, &rec.RecipientPhone, &rec.Message, &rec.EventType, &rec.Provider, &rec.Status, &rec.SentAt, &rec.CreatedAt); err != nil {
		return nil, err
	}
	return &rec, nil
}

func scanSMSLogRow(rows pgx.Rows) (*SMSLogRecord, error) {
	var rec SMSLogRecord
	if err := rows.Scan(&rec.ID, &rec.RecipientPhone, &rec.Message, &rec.EventType, &rec.Provider, &rec.Status, &rec.SentAt, &rec.CreatedAt); err != nil {
		return nil, err
	}
	return &rec, nil
}

func scanEmailLog(row pgx.Row) (*EmailLogRecord, error) {
	var rec EmailLogRecord
	if err := row.Scan(&rec.ID, &rec.RecipientEmail, &rec.Subject, &rec.EventType, &rec.Provider, &rec.Status, &rec.SentAt, &rec.CreatedAt); err != nil {
		return nil, err
	}
	return &rec, nil
}

func scanEmailLogRow(rows pgx.Rows) (*EmailLogRecord, error) {
	var rec EmailLogRecord
	if err := rows.Scan(&rec.ID, &rec.RecipientEmail, &rec.Subject, &rec.EventType, &rec.Provider, &rec.Status, &rec.SentAt, &rec.CreatedAt); err != nil {
		return nil, err
	}
	return &rec, nil
}
