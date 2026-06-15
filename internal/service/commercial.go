package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"

	"github.com/school-management/pos/internal/auth"
	"github.com/school-management/pos/internal/config"
	"github.com/school-management/pos/internal/dto"
	"github.com/school-management/pos/internal/model"
	"github.com/school-management/pos/internal/repository"
)

type SystemService struct {
	repos     *repository.Repositories
	cfg       *config.Config
	audit     *AuditService
	backup    *BackupService
	startTime time.Time
}

func NewSystemService(repos *repository.Repositories, cfg *config.Config, audit *AuditService, backup *BackupService, startTime time.Time) *SystemService {
	return &SystemService{repos: repos, cfg: cfg, audit: audit, backup: backup, startTime: startTime}
}

func (s *SystemService) UptimeSeconds() int64 {
	return int64(time.Since(s.startTime).Seconds())
}

func (s *SystemService) HealthDashboard(ctx context.Context) (*dto.HealthDashboardStats, error) {
	latency, dbErr := s.repos.System.PingDB(ctx)
	dbStatus := "healthy"
	if dbErr != nil {
		dbStatus = "unhealthy"
	}
	queuePending, _ := s.repos.System.CountPendingQueue(ctx)
	queueStatus := "healthy"
	if queuePending > 100 {
		queueStatus = "degraded"
	}
	storageStatus := "healthy"
	if !s.cfg.R2.Enabled {
		storageStatus = "local"
	}
	smtpSettings, _ := s.loadSMTP(ctx)
	smsSettings, _ := s.loadSMS(ctx)
	emailStatus := "log"
	if smtpSettings.Enabled {
		emailStatus = "configured"
	}
	smsStatus := smsSettings.Provider
	if !smsSettings.Enabled {
		smsStatus = "log"
	}
	backupCount, _ := s.repos.System.CountBackups(ctx)
	diskMB := s.backup.DiskUsageMB()
	return &dto.HealthDashboardStats{
		UptimeSeconds: s.UptimeSeconds(), DatabaseStatus: dbStatus,
		DatabaseLatency: latency.String(), StorageStatus: storageStatus,
		QueueStatus: queueStatus, QueuePending: queuePending,
		EmailStatus: emailStatus, SMSStatus: smsStatus,
		DiskUsageMB: diskMB, BackupCount: backupCount,
	}, nil
}

func (s *SystemService) HealthCheck(ctx context.Context) *dto.HealthCheckResponse {
	checks := map[string]string{}
	latency, err := s.repos.System.PingDB(ctx)
	if err != nil {
		checks["database"] = "down"
	} else {
		checks["database"] = fmt.Sprintf("up (%s)", latency)
	}
	if s.cfg.R2.Enabled {
		checks["storage"] = "r2"
	} else {
		checks["storage"] = "local"
	}
	queue, _ := s.repos.System.CountPendingQueue(ctx)
	checks["queue"] = fmt.Sprintf("%d pending", queue)
	status := "ok"
	if checks["database"] == "down" {
		status = "degraded"
	}
	return &dto.HealthCheckResponse{Status: status, Timestamp: time.Now(), Checks: checks}
}

func (s *SystemService) ReadyCheck(ctx context.Context) *dto.ReadyCheckResponse {
	_, err := s.repos.System.PingDB(ctx)
	return &dto.ReadyCheckResponse{Ready: err == nil, Timestamp: time.Now()}
}

func (s *SystemService) Metrics(ctx context.Context) (*dto.MetricsResponse, error) {
	_, dbErr := s.repos.System.PingDB(ctx)
	queue, _ := s.repos.System.CountPendingQueue(ctx)
	backups, _ := s.repos.System.CountBackups(ctx)
	failed, _ := s.repos.System.CountFailedLogins(ctx, time.Now().Add(-24*time.Hour))
	return &dto.MetricsResponse{
		UptimeSeconds: s.UptimeSeconds(), DatabaseUp: dbErr == nil,
		QueuePending: queue, BackupTotal: backups, FailedLogins24: failed,
		DiskUsageMB: s.backup.DiskUsageMB(),
	}, nil
}

func (s *SystemService) SearchAuditLogs(ctx context.Context, f dto.AuditSearchFilter) (*dto.PaginatedAuditLogs, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 {
		f.PageSize = 50
	}
	rf := repository.AuditSearchParams{
		Query: f.Query, Action: f.Action, EntityType: f.EntityType, UserEmail: f.UserEmail,
		From: f.From, To: f.To, Limit: int32(f.PageSize), Offset: int32((f.Page - 1) * f.PageSize),
	}
	total, err := s.repos.AuditLogs.CountSearch(ctx, rf)
	if err != nil {
		return nil, err
	}
	rows, err := s.repos.AuditLogs.Search(ctx, rf)
	if err != nil {
		return nil, err
	}
	items := make([]dto.ActivityItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, dto.ActivityItem{
			ID: r.ID, Action: r.Action, EntityType: r.EntityType,
			UserEmail: r.UserEmail, UserName: strings.TrimSpace(r.UserName),
			Description: r.Description, CreatedAt: r.CreatedAt,
		})
	}
	pages := int(total) / f.PageSize
	if int(total)%f.PageSize > 0 {
		pages++
	}
	return &dto.PaginatedAuditLogs{Items: items, Total: total, Page: f.Page, PageSize: f.PageSize, TotalPages: pages}, nil
}

func (s *SystemService) GetSettings(ctx context.Context) (*dto.SystemSettingsResponse, error) {
	smtp, _ := s.loadSMTP(ctx)
	sms, _ := s.loadSMS(ctx)
	security, _ := s.loadSecurity(ctx)
	branding, _ := s.loadBranding(ctx)
	scheduleRaw, _ := s.repos.System.GetSetting(ctx, model.SettingCategoryGeneral, "backup_schedule")
	schedule := dto.BackupScheduleRequest{
		Enabled:       boolVal(scheduleRaw["enabled"]),
		Cron:          strVal(scheduleRaw["cron"], "0 2 * * *"),
		RetentionDays: intVal(scheduleRaw["retention_days"], 30),
	}
	return &dto.SystemSettingsResponse{
		SMTP: smtp, SMS: sms, Security: security, Branding: branding,
		BackupSchedule: schedule, AppEnv: s.cfg.App.Env, AppURL: s.cfg.App.URL,
	}, nil
}

func (s *SystemService) SaveSMTP(ctx context.Context, req dto.SMTPSettings, actorID uuid.UUID, ip string) error {
	val := map[string]any{
		"host": req.Host, "port": req.Port, "username": req.Username,
		"password": req.Password, "from": req.From, "enabled": req.Enabled,
	}
	if err := s.repos.System.UpsertSetting(ctx, model.SettingCategorySMTP, "config", val); err != nil {
		return err
	}
	s.audit.Log(ctx, &actorID, model.ActionUpdate, model.EntitySystemSetting, nil, ip, map[string]any{"category": "smtp"})
	return nil
}

func (s *SystemService) SaveSMS(ctx context.Context, req dto.SMSSettings, actorID uuid.UUID, ip string) error {
	val := map[string]any{
		"provider": req.Provider, "api_key": req.APIKey, "sender_id": req.SenderID, "enabled": req.Enabled,
	}
	if err := s.repos.System.UpsertSetting(ctx, model.SettingCategorySMS, "config", val); err != nil {
		return err
	}
	s.audit.Log(ctx, &actorID, model.ActionUpdate, model.EntitySystemSetting, nil, ip, map[string]any{"category": "sms"})
	return nil
}

func (s *SystemService) SaveSecurity(ctx context.Context, req dto.SecuritySettings, actorID uuid.UUID, ip string) error {
	val := map[string]any{
		"min_length": req.MinLength, "require_uppercase": req.RequireUppercase,
		"require_number": req.RequireNumber, "require_special": req.RequireSpecial,
		"max_failed_attempts": req.MaxFailedAttempts,
	}
	if err := s.repos.System.UpsertSetting(ctx, model.SettingCategorySecurity, "password_policy", val); err != nil {
		return err
	}
	s.audit.Log(ctx, &actorID, model.ActionUpdate, model.EntitySystemSetting, nil, ip, map[string]any{"category": "security"})
	return nil
}

func (s *SystemService) SaveBranding(ctx context.Context, req dto.BrandingSettings, actorID uuid.UUID, ip string) error {
	val := map[string]any{
		"app_name": req.AppName, "logo_url": req.LogoURL, "favicon_url": req.FaviconURL,
		"primary_color": req.PrimaryColor, "email_footer": req.EmailFooter,
	}
	if err := s.repos.System.UpsertSetting(ctx, model.SettingCategoryBranding, "white_label", val); err != nil {
		return err
	}
	s.audit.Log(ctx, &actorID, model.ActionUpdate, model.EntitySystemSetting, nil, ip, map[string]any{"category": "branding"})
	return nil
}

func (s *SystemService) SaveBackupSchedule(ctx context.Context, req dto.BackupScheduleRequest, actorID uuid.UUID, ip string) error {
	val := map[string]any{
		"enabled": req.Enabled, "cron": req.Cron, "retention_days": req.RetentionDays,
	}
	if err := s.repos.System.UpsertSetting(ctx, model.SettingCategoryGeneral, "backup_schedule", val); err != nil {
		return err
	}
	s.audit.Log(ctx, &actorID, model.ActionUpdate, model.EntitySystemSetting, nil, ip, map[string]any{"category": "backup_schedule"})
	return nil
}

func (s *SystemService) ListEmailTemplates(ctx context.Context) ([]dto.EmailTemplateResponse, error) {
	recs, err := s.repos.System.ListEmailTemplates(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]dto.EmailTemplateResponse, 0, len(recs))
	for _, r := range recs {
		items = append(items, dto.EmailTemplateResponse{
			ID: r.ID, Slug: r.Slug, Name: r.Name, Subject: r.Subject, BodyHTML: r.BodyHTML, UpdatedAt: r.UpdatedAt,
		})
	}
	return items, nil
}

func (s *SystemService) UpdateEmailTemplate(ctx context.Context, slug string, req dto.EmailTemplateRequest, actorID uuid.UUID, ip string) error {
	if err := s.repos.System.UpdateEmailTemplate(ctx, slug, req.Subject, req.BodyHTML); err != nil {
		return err
	}
	s.audit.Log(ctx, &actorID, model.ActionUpdate, model.EntityEmailTemplate, nil, ip, map[string]any{"slug": slug})
	return nil
}

func (s *SystemService) ActivateLicense(ctx context.Context, req dto.LicenseActivateRequest, actorID uuid.UUID, ip string) (*dto.LicenseResponse, error) {
	existing, _ := s.repos.System.GetLicenseByKey(ctx, req.LicenseKey)
	if existing != nil {
		return nil, fmt.Errorf("%w: license key already used", ErrValidation)
	}
	var expires *time.Time
	if req.ExpiresAt != "" {
		if t, err := time.Parse("2006-01-02", req.ExpiresAt); err == nil {
			expires = &t
		}
	} else {
		t := time.Now().AddDate(1, 0, 0)
		expires = &t
	}
	now := time.Now()
	rec, err := s.repos.System.CreateLicense(ctx, repository.CreateLicenseParams{
		LicenseKey: req.LicenseKey, SchoolName: req.SchoolName, SchoolCode: req.SchoolCode,
		RegisteredEmail: req.RegisteredEmail, Status: model.LicenseActive,
		ActivatedAt: &now, ExpiresAt: expires,
	})
	if err != nil {
		return nil, err
	}
	resp := mapLicense(rec)
	s.audit.Log(ctx, &actorID, model.ActionCreate, model.EntityLicense, &rec.ID, ip, map[string]any{"school": req.SchoolName})
	return &resp, nil
}

func (s *SystemService) GetLicense(ctx context.Context) (*dto.LicenseResponse, error) {
	rec, err := s.repos.System.GetActiveLicense(ctx)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, nil
	}
	if rec.ExpiresAt != nil && rec.ExpiresAt.Before(time.Now()) {
		_ = s.repos.System.UpdateLicense(ctx, rec.ID, repository.UpdateLicenseParams{Status: model.LicenseExpired})
		rec.Status = model.LicenseExpired
	}
	resp := mapLicense(rec)
	return &resp, nil
}

func (s *SystemService) ValidateLicense(ctx context.Context) error {
	lic, err := s.GetLicense(ctx)
	if err != nil {
		return err
	}
	if lic == nil {
		return fmt.Errorf("%w: no active license", ErrValidation)
	}
	if lic.Status != model.LicenseActive {
		return fmt.Errorf("%w: license is %s", ErrValidation, lic.Status)
	}
	if lic.DaysRemaining < 0 {
		return fmt.Errorf("%w: license expired", ErrValidation)
	}
	return nil
}

func (s *SystemService) SecurityDashboard(ctx context.Context) (*dto.SecurityDashboardStats, error) {
	today, _ := s.repos.System.CountFailedLoginsToday(ctx)
	week, _ := s.repos.System.CountFailedLogins(ctx, time.Now().AddDate(0, 0, -7))
	ips, _ := s.repos.System.CountUniqueIPs(ctx, time.Now().Truncate(24*time.Hour))
	failures, _ := s.repos.System.ListRecentLoginAttempts(ctx, true, 20)
	ipAudit, _ := s.repos.System.IPAudit(ctx, time.Now().AddDate(0, 0, -7), 20)
	policy, _ := s.loadSecurity(ctx)
	items := make([]dto.LoginAttemptResponse, 0, len(failures))
	for _, f := range failures {
		items = append(items, mapLoginAttempt(&f))
	}
	auditItems := make([]dto.IPAuditEntry, 0, len(ipAudit))
	for _, a := range ipAudit {
		auditItems = append(auditItems, dto.IPAuditEntry{
			IPAddress: a.IPAddress, AttemptCount: a.AttemptCount,
			FailedCount: a.FailedCount, LastAttempt: a.LastAttempt,
		})
	}
	return &dto.SecurityDashboardStats{
		FailedLoginsToday: today, FailedLoginsWeek: week, UniqueIPsToday: ips,
		RecentFailures: items, IPAudit: auditItems, PasswordPolicy: policy,
	}, nil
}

func (s *SystemService) RecordLoginAttempt(ctx context.Context, email, ip, userAgent string, success bool) {
	_ = s.repos.System.CreateLoginAttempt(ctx, email, ip, userAgent, success)
}

func (s *SystemService) ValidatePasswordPolicy(ctx context.Context, password string) error {
	policy, err := s.loadSecurity(ctx)
	if err != nil {
		return err
	}
	minLen := policy.MinLength
	if minLen < 8 {
		minLen = 8
	}
	if len(password) < minLen {
		return fmt.Errorf("%w: password must be at least %d characters", ErrValidation, minLen)
	}
	if policy.RequireUppercase {
		has := false
		for _, c := range password {
			if unicode.IsUpper(c) {
				has = true
				break
			}
		}
		if !has {
			return fmt.Errorf("%w: password must contain uppercase letter", ErrValidation)
		}
	}
	if policy.RequireNumber {
		has := false
		for _, c := range password {
			if unicode.IsDigit(c) {
				has = true
				break
			}
		}
		if !has {
			return fmt.Errorf("%w: password must contain a number", ErrValidation)
		}
	}
	return nil
}

func (s *SystemService) InstallStatus(ctx context.Context) (*dto.InstallStatusResponse, error) {
	admin, _ := s.repos.Users.GetByEmail(ctx, s.cfg.Admin.Email)
	schools, _ := s.repos.Schools.Get(ctx)
	lic, _ := s.GetLicense(ctx)
	settings, _ := s.repos.System.GetSetting(ctx, model.SettingCategoryGeneral, "installed")
	installed := boolVal(settings["value"])
	if !installed && admin != nil && schools != nil {
		installed = true
	}
	return &dto.InstallStatusResponse{
		Installed: installed, HasAdmin: admin != nil,
		HasSchool: schools != nil, HasLicense: lic != nil,
	}, nil
}

func (s *SystemService) RunInstall(ctx context.Context, req dto.InstallRequest, ip string) error {
	if err := s.ValidatePasswordPolicy(ctx, req.AdminPassword); err != nil {
		return err
	}
	role, err := s.repos.Roles.GetBySlug(ctx, model.RoleAdmin)
	if err != nil || role == nil {
		return fmt.Errorf("admin role not found")
	}
	existing, _ := s.repos.Users.GetByEmail(ctx, req.AdminEmail)
	if existing == nil {
		hash, err := auth.HashPassword(req.AdminPassword)
		if err != nil {
			return err
		}
		_, err = s.repos.Users.Create(ctx, repository.CreateUserParams{
			Email: strings.ToLower(req.AdminEmail), PasswordHash: hash,
			FirstName: req.AdminFirstName, LastName: req.AdminLastName,
			RoleID: role.ID, IsActive: true,
		})
		if err != nil {
			return err
		}
	}
	schools, _ := s.repos.Schools.Get(ctx)
	if schools == nil {
		_, err = s.repos.Schools.Create(ctx, repository.SchoolParams{
			Name: req.SchoolName, Email: req.SchoolEmail, Phone: req.SchoolPhone, Address: req.SchoolAddress,
		})
		if err != nil {
			return err
		}
	}
	_ = s.repos.System.UpsertSetting(ctx, model.SettingCategoryGeneral, "installed", map[string]any{"value": true})
	s.audit.Log(ctx, nil, model.ActionCreate, model.EntitySystemSetting, nil, ip, map[string]any{"action": "install"})
	return nil
}

func (s *SystemService) GenerateLicenseKey() string {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	return "SMP-" + strings.ToUpper(hex.EncodeToString(b))
}

func (s *SystemService) loadSMTP(ctx context.Context) (dto.SMTPSettings, error) {
	raw, err := s.repos.System.GetSetting(ctx, model.SettingCategorySMTP, "config")
	if err != nil {
		return dto.SMTPSettings{}, err
	}
	if len(raw) == 0 {
		return dto.SMTPSettings{Port: 587, Host: s.cfg.SMTP.Host, From: s.cfg.SMTP.From, Enabled: s.cfg.SMTP.Enabled}, nil
	}
	return dto.SMTPSettings{
		Host: strVal(raw["host"], s.cfg.SMTP.Host), Port: intVal(raw["port"], s.cfg.SMTP.Port),
		Username: strVal(raw["username"], ""), Password: strVal(raw["password"], ""),
		From: strVal(raw["from"], s.cfg.SMTP.From), Enabled: boolVal(raw["enabled"]),
	}, nil
}

func (s *SystemService) loadSMS(ctx context.Context) (dto.SMSSettings, error) {
	raw, err := s.repos.System.GetSetting(ctx, model.SettingCategorySMS, "config")
	if err != nil {
		return dto.SMSSettings{}, err
	}
	if len(raw) == 0 {
		return dto.SMSSettings{Provider: s.cfg.SMS.Provider, Enabled: s.cfg.SMS.Enabled}, nil
	}
	return dto.SMSSettings{
		Provider: strVal(raw["provider"], "log"), APIKey: strVal(raw["api_key"], ""),
		SenderID: strVal(raw["sender_id"], ""), Enabled: boolVal(raw["enabled"]),
	}, nil
}

func (s *SystemService) loadSecurity(ctx context.Context) (dto.SecuritySettings, error) {
	raw, err := s.repos.System.GetSetting(ctx, model.SettingCategorySecurity, "password_policy")
	if err != nil {
		return dto.SecuritySettings{}, err
	}
	return dto.SecuritySettings{
		MinLength: intVal(raw["min_length"], 8), RequireUppercase: boolVal(raw["require_uppercase"]),
		RequireNumber: boolVal(raw["require_number"]), RequireSpecial: boolVal(raw["require_special"]),
		MaxFailedAttempts: intVal(raw["max_failed_attempts"], 5),
	}, nil
}

func (s *SystemService) loadBranding(ctx context.Context) (dto.BrandingSettings, error) {
	raw, err := s.repos.System.GetSetting(ctx, model.SettingCategoryBranding, "white_label")
	if err != nil {
		return dto.BrandingSettings{}, err
	}
	return dto.BrandingSettings{
		AppName: strVal(raw["app_name"], s.cfg.App.Name), LogoURL: strVal(raw["logo_url"], ""),
		FaviconURL: strVal(raw["favicon_url"], ""), PrimaryColor: strVal(raw["primary_color"], "#4f46e5"),
		EmailFooter: strVal(raw["email_footer"], ""),
	}, nil
}

func mapLicense(r *repository.LicenseRecord) dto.LicenseResponse {
	resp := dto.LicenseResponse{
		ID: r.ID, LicenseKey: r.LicenseKey, SchoolName: r.SchoolName, SchoolCode: r.SchoolCode,
		Status: r.Status, RegisteredEmail: r.RegisteredEmail,
		ActivatedAt: r.ActivatedAt, ExpiresAt: r.ExpiresAt, CreatedAt: r.CreatedAt,
	}
	if r.ExpiresAt != nil {
		resp.DaysRemaining = int(time.Until(*r.ExpiresAt).Hours() / 24)
	}
	return resp
}

func mapLoginAttempt(r *repository.LoginAttemptRecord) dto.LoginAttemptResponse {
	return dto.LoginAttemptResponse{
		ID: r.ID, Email: r.Email, IPAddress: r.IPAddress,
		Success: r.Success, UserAgent: r.UserAgent, CreatedAt: r.CreatedAt,
	}
}

func strVal(v any, fallback string) string {
	if s, ok := v.(string); ok {
		return s
	}
	return fallback
}

func boolVal(v any) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}

func intVal(v any, fallback int) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	default:
		return fallback
	}
}

func (s *SystemService) BackupDirSize() float64 { return s.backup.DiskUsageMB() }
