package dto

import (
	"time"

	"github.com/google/uuid"
)

type SystemBackupResponse struct {
	ID          uuid.UUID  `json:"id"`
	FileName    string     `json:"file_name"`
	FileSize    int64      `json:"file_size"`
	BackupType  string     `json:"backup_type"`
	Status      string     `json:"status"`
	Checksum    string     `json:"checksum"`
	Verified    bool       `json:"verified"`
	ErrorMessage string    `json:"error_message"`
	CreatedByName string   `json:"created_by_name"`
	CreatedAt   time.Time  `json:"created_at"`
}

type PaginatedBackups struct {
	Items      []SystemBackupResponse `json:"items"`
	Total      int64                  `json:"total"`
	Page       int                    `json:"page"`
	PageSize   int                    `json:"page_size"`
	TotalPages int                    `json:"total_pages"`
}

type BackupScheduleRequest struct {
	Enabled       bool   `form:"enabled"`
	Cron          string `form:"cron"`
	RetentionDays int    `form:"retention_days"`
}

type HealthDashboardStats struct {
	UptimeSeconds   int64   `json:"uptime_seconds"`
	DatabaseStatus  string  `json:"database_status"`
	DatabaseLatency string  `json:"database_latency"`
	StorageStatus   string  `json:"storage_status"`
	StorageUsedMB   float64 `json:"storage_used_mb"`
	QueueStatus     string  `json:"queue_status"`
	QueuePending    int64   `json:"queue_pending"`
	EmailStatus     string  `json:"email_status"`
	SMSStatus       string  `json:"sms_status"`
	DiskUsageMB     float64 `json:"disk_usage_mb"`
	BackupCount     int64   `json:"backup_count"`
}

type HealthCheckResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Checks    map[string]string `json:"checks"`
}

type ReadyCheckResponse struct {
	Ready     bool      `json:"ready"`
	Timestamp time.Time `json:"timestamp"`
}

type MetricsResponse struct {
	UptimeSeconds  int64   `json:"uptime_seconds"`
	DatabaseUp     bool    `json:"database_up"`
	QueuePending   int64   `json:"queue_pending"`
	BackupTotal    int64   `json:"backup_total"`
	FailedLogins24 int64   `json:"failed_logins_24h"`
	DiskUsageMB    float64 `json:"disk_usage_mb"`
}

type AuditSearchFilter struct {
	Query      string
	Action     string
	EntityType string
	UserEmail  string
	From       time.Time
	To         time.Time
	Page       int
	PageSize   int
}

type PaginatedAuditLogs struct {
	Items      []ActivityItem `json:"items"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}

type SystemSettingsResponse struct {
	SMTP          SMTPSettings          `json:"smtp"`
	SMS           SMSSettings           `json:"sms"`
	Security      SecuritySettings      `json:"security"`
	Branding      BrandingSettings      `json:"branding"`
	BackupSchedule BackupScheduleRequest `json:"backup_schedule"`
	AppEnv        string                `json:"app_env"`
	AppURL        string                `json:"app_url"`
}

type SMTPSettings struct {
	Host     string `form:"smtp_host" json:"host"`
	Port     int    `form:"smtp_port" json:"port"`
	Username string `form:"smtp_username" json:"username"`
	Password string `form:"smtp_password" json:"password"`
	From     string `form:"smtp_from" json:"from"`
	Enabled  bool   `form:"smtp_enabled" json:"enabled"`
}

type SMSSettings struct {
	Provider string `form:"sms_provider" json:"provider"`
	APIKey   string `form:"sms_api_key" json:"api_key"`
	SenderID string `form:"sms_sender_id" json:"sender_id"`
	Enabled  bool   `form:"sms_enabled" json:"enabled"`
}

type SecuritySettings struct {
	MinLength         int  `form:"min_length" json:"min_length"`
	RequireUppercase  bool `form:"require_uppercase" json:"require_uppercase"`
	RequireNumber     bool `form:"require_number" json:"require_number"`
	RequireSpecial    bool `form:"require_special" json:"require_special"`
	MaxFailedAttempts int  `form:"max_failed_attempts" json:"max_failed_attempts"`
}

type BrandingSettings struct {
	AppName      string `form:"app_name" json:"app_name"`
	LogoURL      string `form:"logo_url" json:"logo_url"`
	FaviconURL   string `form:"favicon_url" json:"favicon_url"`
	PrimaryColor string `form:"primary_color" json:"primary_color"`
	EmailFooter  string `form:"email_footer" json:"email_footer"`
}

type EmailTemplateResponse struct {
	ID        uuid.UUID `json:"id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	Subject   string    `json:"subject"`
	BodyHTML  string    `json:"body_html"`
	UpdatedAt time.Time `json:"updated_at"`
}

type EmailTemplateRequest struct {
	Subject  string `form:"subject" validate:"required"`
	BodyHTML string `form:"body_html" validate:"required"`
}

type LicenseResponse struct {
	ID              uuid.UUID  `json:"id"`
	LicenseKey      string     `json:"license_key"`
	SchoolName      string     `json:"school_name"`
	SchoolCode      string     `json:"school_code"`
	Status          string     `json:"status"`
	RegisteredEmail string     `json:"registered_email"`
	ActivatedAt     *time.Time `json:"activated_at"`
	ExpiresAt       *time.Time `json:"expires_at"`
	DaysRemaining   int        `json:"days_remaining"`
	CreatedAt       time.Time  `json:"created_at"`
}

type LicenseActivateRequest struct {
	LicenseKey      string `form:"license_key" validate:"required"`
	SchoolName      string `form:"school_name" validate:"required"`
	SchoolCode      string `form:"school_code"`
	RegisteredEmail string `form:"registered_email" validate:"required,email"`
	ExpiresAt       string `form:"expires_at"`
}

type SecurityDashboardStats struct {
	FailedLoginsToday int64                `json:"failed_logins_today"`
	FailedLoginsWeek  int64                `json:"failed_logins_week"`
	UniqueIPsToday    int64                `json:"unique_ips_today"`
	RecentFailures    []LoginAttemptResponse `json:"recent_failures"`
	IPAudit           []IPAuditEntry       `json:"ip_audit"`
	PasswordPolicy    SecuritySettings     `json:"password_policy"`
}

type LoginAttemptResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	IPAddress string    `json:"ip_address"`
	Success   bool      `json:"success"`
	UserAgent string    `json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
}

type IPAuditEntry struct {
	IPAddress    string    `json:"ip_address"`
	AttemptCount int64     `json:"attempt_count"`
	LastAttempt  time.Time `json:"last_attempt"`
	FailedCount  int64     `json:"failed_count"`
}

type InstallRequest struct {
	AdminEmail     string `form:"admin_email" validate:"required,email"`
	AdminPassword  string `form:"admin_password" validate:"required,min=8"`
	AdminFirstName string `form:"admin_first_name" validate:"required"`
	AdminLastName  string `form:"admin_last_name" validate:"required"`
	SchoolName     string `form:"school_name" validate:"required"`
	SchoolEmail    string `form:"school_email" validate:"required,email"`
	SchoolPhone    string `form:"school_phone"`
	SchoolAddress  string `form:"school_address"`
}

type InstallStatusResponse struct {
	Installed   bool `json:"installed"`
	HasAdmin    bool `json:"has_admin"`
	HasSchool   bool `json:"has_school"`
	HasLicense  bool `json:"has_license"`
}
