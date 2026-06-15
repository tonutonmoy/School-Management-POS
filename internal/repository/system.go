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

type SystemRepository interface {
	// Backups
	CreateBackup(ctx context.Context, p CreateBackupParams) (*BackupRecord, error)
	UpdateBackup(ctx context.Context, id uuid.UUID, p UpdateBackupParams) error
	GetBackup(ctx context.Context, id uuid.UUID) (*BackupRecord, error)
	ListBackups(ctx context.Context, limit, offset int32) ([]BackupRecord, error)
	CountBackups(ctx context.Context) (int64, error)
	DeleteOldBackups(ctx context.Context, before time.Time) (int64, error)

	// Settings
	GetSetting(ctx context.Context, category, key string) (map[string]any, error)
	UpsertSetting(ctx context.Context, category, key string, value map[string]any) error
	ListSettingsByCategory(ctx context.Context, category string) (map[string]map[string]any, error)

	// License
	GetActiveLicense(ctx context.Context) (*LicenseRecord, error)
	GetLicenseByKey(ctx context.Context, key string) (*LicenseRecord, error)
	CreateLicense(ctx context.Context, p CreateLicenseParams) (*LicenseRecord, error)
	UpdateLicense(ctx context.Context, id uuid.UUID, p UpdateLicenseParams) error

	// Login attempts
	CreateLoginAttempt(ctx context.Context, email, ip, userAgent string, success bool) error
	CountFailedLogins(ctx context.Context, since time.Time) (int64, error)
	CountFailedLoginsToday(ctx context.Context) (int64, error)
	ListRecentLoginAttempts(ctx context.Context, failedOnly bool, limit int32) ([]LoginAttemptRecord, error)
	IPAudit(ctx context.Context, since time.Time, limit int32) ([]IPAuditRecord, error)
	CountUniqueIPs(ctx context.Context, since time.Time) (int64, error)

	// Email templates
	ListEmailTemplates(ctx context.Context) ([]EmailTemplateRecord, error)
	GetEmailTemplate(ctx context.Context, slug string) (*EmailTemplateRecord, error)
	UpdateEmailTemplate(ctx context.Context, slug, subject, body string) error

	// Health helpers
	PingDB(ctx context.Context) (time.Duration, error)
	CountPendingQueue(ctx context.Context) (int64, error)
}

type BackupRecord struct {
	ID           uuid.UUID
	FileName     string
	FilePath     string
	FileSize     int64
	BackupType   string
	Status       string
	Checksum     string
	Verified     bool
	ErrorMessage string
	CreatedByName string
	CreatedAt    time.Time
}

type CreateBackupParams struct {
	FileName, FilePath string
	FileSize           int64
	BackupType         string
	Status             string
	Checksum           string
	Verified           bool
	CreatedBy          *uuid.UUID
}

type UpdateBackupParams struct {
	Status, Checksum, ErrorMessage string
	FileSize                       int64
	Verified                       bool
}

type LicenseRecord struct {
	ID              uuid.UUID
	LicenseKey      string
	SchoolName      string
	SchoolCode      string
	Status          string
	RegisteredEmail string
	ActivatedAt     *time.Time
	ExpiresAt       *time.Time
	CreatedAt       time.Time
}

type CreateLicenseParams struct {
	LicenseKey, SchoolName, SchoolCode, RegisteredEmail, Status string
	ActivatedAt, ExpiresAt                                      *time.Time
}

type UpdateLicenseParams struct {
	Status          string
	SchoolName      string
	ExpiresAt       *time.Time
	RegisteredEmail string
}

type LoginAttemptRecord struct {
	ID        uuid.UUID
	Email     string
	IPAddress string
	Success   bool
	UserAgent string
	CreatedAt time.Time
}

type IPAuditRecord struct {
	IPAddress    string
	AttemptCount int64
	FailedCount  int64
	LastAttempt  time.Time
}

type EmailTemplateRecord struct {
	ID        uuid.UUID
	Slug      string
	Name      string
	Subject   string
	BodyHTML  string
	UpdatedAt time.Time
}

type systemRepository struct {
	pool *pgxpool.Pool
}

func NewSystemRepository(pool *pgxpool.Pool) SystemRepository {
	return &systemRepository{pool: pool}
}

func (r *systemRepository) CreateBackup(ctx context.Context, p CreateBackupParams) (*BackupRecord, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
INSERT INTO system_backups (file_name, file_path, file_size, backup_type, status, checksum, verified, created_by)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`,
		p.FileName, p.FilePath, p.FileSize, p.BackupType, p.Status, p.Checksum, p.Verified, p.CreatedBy,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.GetBackup(ctx, id)
}

func (r *systemRepository) UpdateBackup(ctx context.Context, id uuid.UUID, p UpdateBackupParams) error {
	_, err := r.pool.Exec(ctx, `
UPDATE system_backups SET status=$2, checksum=COALESCE(NULLIF($3,''), checksum),
    error_message=$4, file_size=CASE WHEN $5>0 THEN $5 ELSE file_size END, verified=$6
WHERE id=$1`, id, p.Status, p.Checksum, p.ErrorMessage, p.FileSize, p.Verified)
	return err
}

func (r *systemRepository) GetBackup(ctx context.Context, id uuid.UUID) (*BackupRecord, error) {
	row := r.pool.QueryRow(ctx, `
SELECT b.id, b.file_name, b.file_path, b.file_size, b.backup_type, b.status,
    COALESCE(b.checksum,''), b.verified, COALESCE(b.error_message,''),
    COALESCE(u.first_name||' '||u.last_name,''), b.created_at
FROM system_backups b LEFT JOIN users u ON u.id=b.created_by WHERE b.id=$1`, id)
	var rec BackupRecord
	if err := row.Scan(&rec.ID, &rec.FileName, &rec.FilePath, &rec.FileSize, &rec.BackupType, &rec.Status,
		&rec.Checksum, &rec.Verified, &rec.ErrorMessage, &rec.CreatedByName, &rec.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &rec, nil
}

func (r *systemRepository) ListBackups(ctx context.Context, limit, offset int32) ([]BackupRecord, error) {
	rows, err := r.pool.Query(ctx, `
SELECT b.id, b.file_name, b.file_path, b.file_size, b.backup_type, b.status,
    COALESCE(b.checksum,''), b.verified, COALESCE(b.error_message,''),
    COALESCE(u.first_name||' '||u.last_name,''), b.created_at
FROM system_backups b LEFT JOIN users u ON u.id=b.created_by
ORDER BY b.created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanBackups(rows)
}

func scanBackups(rows pgx.Rows) ([]BackupRecord, error) {
	var items []BackupRecord
	for rows.Next() {
		var rec BackupRecord
		if err := rows.Scan(&rec.ID, &rec.FileName, &rec.FilePath, &rec.FileSize, &rec.BackupType, &rec.Status,
			&rec.Checksum, &rec.Verified, &rec.ErrorMessage, &rec.CreatedByName, &rec.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, rec)
	}
	return items, rows.Err()
}

func (r *systemRepository) CountBackups(ctx context.Context) (int64, error) {
	var n int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM system_backups`).Scan(&n)
	return n, err
}

func (r *systemRepository) DeleteOldBackups(ctx context.Context, before time.Time) (int64, error) {
	tag, err := r.pool.Exec(ctx, `DELETE FROM system_backups WHERE created_at < $1`, before)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func (r *systemRepository) GetSetting(ctx context.Context, category, key string) (map[string]any, error) {
	var raw []byte
	err := r.pool.QueryRow(ctx, `SELECT value FROM system_settings WHERE category=$1 AND setting_key=$2`, category, key).Scan(&raw)
	if err != nil {
		if err == pgx.ErrNoRows {
			return map[string]any{}, nil
		}
		return nil, err
	}
	var m map[string]any
	_ = json.Unmarshal(raw, &m)
	if m == nil {
		m = map[string]any{}
	}
	return m, nil
}

func (r *systemRepository) UpsertSetting(ctx context.Context, category, key string, value map[string]any) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `
INSERT INTO system_settings (category, setting_key, value) VALUES ($1,$2,$3)
ON CONFLICT (category, setting_key) DO UPDATE SET value=$3, updated_at=NOW()`, category, key, raw)
	return err
}

func (r *systemRepository) ListSettingsByCategory(ctx context.Context, category string) (map[string]map[string]any, error) {
	rows, err := r.pool.Query(ctx, `SELECT setting_key, value FROM system_settings WHERE category=$1`, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]map[string]any{}
	for rows.Next() {
		var key string
		var raw []byte
		if err := rows.Scan(&key, &raw); err != nil {
			return nil, err
		}
		var m map[string]any
		_ = json.Unmarshal(raw, &m)
		out[key] = m
	}
	return out, rows.Err()
}

func (r *systemRepository) GetActiveLicense(ctx context.Context) (*LicenseRecord, error) {
	row := r.pool.QueryRow(ctx, `
SELECT id, license_key, school_name, COALESCE(school_code,''), status,
    COALESCE(registered_email,''), activated_at, expires_at, created_at
FROM licenses WHERE status='active' ORDER BY activated_at DESC NULLS LAST LIMIT 1`)
	return scanLicense(row)
}

func (r *systemRepository) GetLicenseByKey(ctx context.Context, key string) (*LicenseRecord, error) {
	row := r.pool.QueryRow(ctx, `
SELECT id, license_key, school_name, COALESCE(school_code,''), status,
    COALESCE(registered_email,''), activated_at, expires_at, created_at
FROM licenses WHERE license_key=$1`, key)
	return scanLicense(row)
}

func scanLicense(row pgx.Row) (*LicenseRecord, error) {
	var rec LicenseRecord
	if err := row.Scan(&rec.ID, &rec.LicenseKey, &rec.SchoolName, &rec.SchoolCode, &rec.Status,
		&rec.RegisteredEmail, &rec.ActivatedAt, &rec.ExpiresAt, &rec.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &rec, nil
}

func (r *systemRepository) CreateLicense(ctx context.Context, p CreateLicenseParams) (*LicenseRecord, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
INSERT INTO licenses (license_key, school_name, school_code, status, registered_email, activated_at, expires_at)
VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`,
		p.LicenseKey, p.SchoolName, p.SchoolCode, p.Status, p.RegisteredEmail, p.ActivatedAt, p.ExpiresAt,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.GetLicenseByKey(ctx, p.LicenseKey)
}

func (r *systemRepository) UpdateLicense(ctx context.Context, id uuid.UUID, p UpdateLicenseParams) error {
	_, err := r.pool.Exec(ctx, `
UPDATE licenses SET status=$2, school_name=COALESCE(NULLIF($3,''), school_name),
    expires_at=COALESCE($4, expires_at), registered_email=COALESCE(NULLIF($5,''), registered_email), updated_at=NOW()
WHERE id=$1`, id, p.Status, p.SchoolName, p.ExpiresAt, p.RegisteredEmail)
	return err
}

func (r *systemRepository) CreateLoginAttempt(ctx context.Context, email, ip, userAgent string, success bool) error {
	_, err := r.pool.Exec(ctx, `
INSERT INTO login_attempts (email, ip_address, user_agent, success) VALUES ($1,$2,$3,$4)`,
		strings.ToLower(email), ip, userAgent, success)
	return err
}

func (r *systemRepository) CountFailedLogins(ctx context.Context, since time.Time) (int64, error) {
	var n int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM login_attempts WHERE success=false AND created_at>=$1`, since).Scan(&n)
	return n, err
}

func (r *systemRepository) CountFailedLoginsToday(ctx context.Context) (int64, error) {
	return r.CountFailedLogins(ctx, time.Now().Truncate(24*time.Hour))
}

func (r *systemRepository) ListRecentLoginAttempts(ctx context.Context, failedOnly bool, limit int32) ([]LoginAttemptRecord, error) {
	q := `SELECT id, email, COALESCE(ip_address,''), success, COALESCE(user_agent,''), created_at FROM login_attempts`
	if failedOnly {
		q += ` WHERE success=false`
	}
	q += fmt.Sprintf(` ORDER BY created_at DESC LIMIT %d`, limit)
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []LoginAttemptRecord
	for rows.Next() {
		var rec LoginAttemptRecord
		if err := rows.Scan(&rec.ID, &rec.Email, &rec.IPAddress, &rec.Success, &rec.UserAgent, &rec.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, rec)
	}
	return items, rows.Err()
}

func (r *systemRepository) IPAudit(ctx context.Context, since time.Time, limit int32) ([]IPAuditRecord, error) {
	rows, err := r.pool.Query(ctx, `
SELECT COALESCE(ip_address,'unknown'), COUNT(*),
    COUNT(*) FILTER (WHERE success=false), MAX(created_at)
FROM login_attempts WHERE created_at>=$1 AND ip_address IS NOT NULL AND ip_address != ''
GROUP BY ip_address ORDER BY MAX(created_at) DESC LIMIT $2`, since, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []IPAuditRecord
	for rows.Next() {
		var rec IPAuditRecord
		if err := rows.Scan(&rec.IPAddress, &rec.AttemptCount, &rec.FailedCount, &rec.LastAttempt); err != nil {
			return nil, err
		}
		items = append(items, rec)
	}
	return items, rows.Err()
}

func (r *systemRepository) CountUniqueIPs(ctx context.Context, since time.Time) (int64, error) {
	var n int64
	err := r.pool.QueryRow(ctx, `
SELECT COUNT(DISTINCT ip_address) FROM login_attempts
WHERE created_at>=$1 AND ip_address IS NOT NULL AND ip_address != ''`, since).Scan(&n)
	return n, err
}

func (r *systemRepository) ListEmailTemplates(ctx context.Context) ([]EmailTemplateRecord, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, slug, name, subject, body_html, updated_at FROM email_templates ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []EmailTemplateRecord
	for rows.Next() {
		var rec EmailTemplateRecord
		if err := rows.Scan(&rec.ID, &rec.Slug, &rec.Name, &rec.Subject, &rec.BodyHTML, &rec.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, rec)
	}
	return items, rows.Err()
}

func (r *systemRepository) GetEmailTemplate(ctx context.Context, slug string) (*EmailTemplateRecord, error) {
	row := r.pool.QueryRow(ctx, `SELECT id, slug, name, subject, body_html, updated_at FROM email_templates WHERE slug=$1`, slug)
	var rec EmailTemplateRecord
	if err := row.Scan(&rec.ID, &rec.Slug, &rec.Name, &rec.Subject, &rec.BodyHTML, &rec.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &rec, nil
}

func (r *systemRepository) UpdateEmailTemplate(ctx context.Context, slug, subject, body string) error {
	_, err := r.pool.Exec(ctx, `UPDATE email_templates SET subject=$2, body_html=$3, updated_at=NOW() WHERE slug=$1`, slug, subject, body)
	return err
}

func (r *systemRepository) PingDB(ctx context.Context) (time.Duration, error) {
	start := time.Now()
	err := r.pool.QueryRow(ctx, `SELECT 1`).Scan(new(int))
	return time.Since(start), err
}

func (r *systemRepository) CountPendingQueue(ctx context.Context) (int64, error) {
	var n int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM notification_queue WHERE status='pending'`).Scan(&n)
	return n, err
}

type AuditSearchParams struct {
	Query, Action, EntityType, UserEmail string
	From, To                             time.Time
	Limit, Offset                        int32
}

func (r *auditRepository) Search(ctx context.Context, f AuditSearchParams) ([]AuditRecord, error) {
	where, args := auditFilter(f)
	q := `
SELECT a.id,a.user_id,a.action,a.entity_type,a.entity_id,a.ip_address,a.created_at,
       u.email,u.first_name,u.last_name
FROM audit_logs a LEFT JOIN users u ON u.id=a.user_id` + where +
		fmt.Sprintf(" ORDER BY a.created_at DESC LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	args = append(args, f.Limit, f.Offset)
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAudit(rows)
}

func (r *auditRepository) CountSearch(ctx context.Context, f AuditSearchParams) (int64, error) {
	where, args := auditFilter(f)
	q := `SELECT COUNT(*) FROM audit_logs a LEFT JOIN users u ON u.id=a.user_id` + where
	var n int64
	err := r.pool.QueryRow(ctx, q, args...).Scan(&n)
	return n, err
}

func auditFilter(f AuditSearchParams) (string, []any) {
	var parts []string
	var args []any
	if f.Action != "" {
		args = append(args, f.Action)
		parts = append(parts, fmt.Sprintf("a.action=$%d", len(args)))
	}
	if f.EntityType != "" {
		args = append(args, f.EntityType)
		parts = append(parts, fmt.Sprintf("a.entity_type=$%d", len(args)))
	}
	if f.UserEmail != "" {
		args = append(args, "%"+strings.ToLower(f.UserEmail)+"%")
		parts = append(parts, fmt.Sprintf("LOWER(u.email) LIKE $%d", len(args)))
	}
	if !f.From.IsZero() {
		args = append(args, f.From)
		parts = append(parts, fmt.Sprintf("a.created_at >= $%d", len(args)))
	}
	if !f.To.IsZero() {
		args = append(args, f.To)
		parts = append(parts, fmt.Sprintf("a.created_at <= $%d", len(args)))
	}
	if f.Query != "" {
		args = append(args, "%"+strings.ToLower(f.Query)+"%")
		parts = append(parts, fmt.Sprintf("(LOWER(a.action) LIKE $%d OR LOWER(a.entity_type) LIKE $%d OR LOWER(COALESCE(u.email,'')) LIKE $%d)", len(args), len(args), len(args)))
	}
	if len(parts) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(parts, " AND "), args
}
