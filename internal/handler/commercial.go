package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/school-management/pos/internal/dto"
	"github.com/school-management/pos/internal/export"
	"github.com/school-management/pos/internal/middleware"
	"github.com/school-management/pos/internal/model"
	"github.com/school-management/pos/internal/validator"
	"github.com/school-management/pos/internal/web"
)

func (h *Handler) registerCommercialRoutes(app, auth fiber.Router, mw *middleware.Middleware) {
	app.Get("/health", h.HealthCheck)
	app.Get("/ready", h.ReadyCheck)
	app.Get("/metrics", h.Metrics)

	app.Get("/install", h.InstallPage)
	app.Post("/install", mw.CSRFProtect(), h.InstallSubmit)

	sys := auth.Group("/system", mw.RequirePermission(model.PermSystemManage))
	sys.Get("/health", h.SystemHealthDashboard)
	sys.Get("/settings", h.SystemSettingsPage)
	sys.Post("/settings/smtp", mw.CSRFProtect(), h.SystemSaveSMTP)
	sys.Post("/settings/sms", mw.CSRFProtect(), h.SystemSaveSMS)
	sys.Post("/settings/security", mw.CSRFProtect(), h.SystemSaveSecurity)
	sys.Post("/settings/branding", mw.CSRFProtect(), h.SystemSaveBranding)
	sys.Post("/settings/backup-schedule", mw.CSRFProtect(), h.SystemSaveBackupSchedule)
	sys.Get("/templates", h.EmailTemplatesPage)
	sys.Post("/templates/:slug", mw.CSRFProtect(), h.EmailTemplateUpdate)

	backup := auth.Group("/system/backups", mw.RequirePermission(model.PermSystemBackup))
	backup.Get("/", h.BackupListPage)
	backup.Post("/create", mw.CSRFProtect(), h.BackupCreate)
	backup.Post("/:id/restore", mw.CSRFProtect(), h.BackupRestore)
	backup.Post("/:id/verify", mw.CSRFProtect(), h.BackupVerify)
	backup.Get("/:id/download", h.BackupDownload)

	audit := auth.Group("/system/audit", mw.RequirePermission(model.PermSystemAudit))
	audit.Get("/", h.AuditCenterPage)
	audit.Get("/export.csv", h.AuditExportCSV)

	license := auth.Group("/system/license", mw.RequirePermission(model.PermSystemLicense))
	license.Get("/", h.LicensePage)
	license.Post("/activate", mw.CSRFProtect(), h.LicenseActivate)

	security := auth.Group("/system/security", mw.RequirePermission(model.PermSystemSecurity))
	security.Get("/", h.SecurityDashboardPage)
}

func (h *Handler) HealthCheck(c fiber.Ctx) error {
	resp := h.services.System.HealthCheck(c.Context())
	return c.JSON(resp)
}

func (h *Handler) ReadyCheck(c fiber.Ctx) error {
	resp := h.services.System.ReadyCheck(c.Context())
	status := fiber.StatusOK
	if !resp.Ready {
		status = fiber.StatusServiceUnavailable
	}
	return c.Status(status).JSON(resp)
}

func (h *Handler) Metrics(c fiber.Ctx) error {
	m, err := h.services.System.Metrics(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "metrics unavailable"})
	}
	return c.JSON(m)
}

func (h *Handler) SystemHealthDashboard(c fiber.Ctx) error {
	stats, _ := h.services.System.HealthDashboard(c.Context())
	return h.render(c, fiber.StatusOK, web.SystemHealthPage{Stats: stats})
}

func (h *Handler) SystemSettingsPage(c fiber.Ctx) error {
	settings, _ := h.services.System.GetSettings(c.Context())
	return h.render(c, fiber.StatusOK, web.SystemSettingsPage{
		Settings: settings, Flash: c.Cookies("flash"), FlashType: c.Cookies("flash_type"),
	})
}

func (h *Handler) SystemSaveSMTP(c fiber.Ctx) error {
	return h.saveSettingsFlash(c, "/system/settings", func(user uuid.UUID) error {
		req := dto.SMTPSettings{
			Host: c.FormValue("smtp_host"), Port: parseIntDefault(c.FormValue("smtp_port"), 587),
			Username: c.FormValue("smtp_username"), Password: c.FormValue("smtp_password"),
			From: c.FormValue("smtp_from"), Enabled: c.FormValue("smtp_enabled") == "on",
		}
		return h.services.System.SaveSMTP(c.Context(), req, user, c.IP())
	})
}

func (h *Handler) SystemSaveSMS(c fiber.Ctx) error {
	return h.saveSettingsFlash(c, "/system/settings", func(user uuid.UUID) error {
		req := dto.SMSSettings{
			Provider: c.FormValue("sms_provider"), APIKey: c.FormValue("sms_api_key"),
			SenderID: c.FormValue("sms_sender_id"), Enabled: c.FormValue("sms_enabled") == "on",
		}
		return h.services.System.SaveSMS(c.Context(), req, user, c.IP())
	})
}

func (h *Handler) SystemSaveSecurity(c fiber.Ctx) error {
	return h.saveSettingsFlash(c, "/system/settings", func(user uuid.UUID) error {
		req := dto.SecuritySettings{
			MinLength: parseIntDefault(c.FormValue("min_length"), 8),
			RequireUppercase: c.FormValue("require_uppercase") == "on",
			RequireNumber:    c.FormValue("require_number") == "on",
			RequireSpecial:   c.FormValue("require_special") == "on",
			MaxFailedAttempts: parseIntDefault(c.FormValue("max_failed_attempts"), 5),
		}
		return h.services.System.SaveSecurity(c.Context(), req, user, c.IP())
	})
}

func (h *Handler) SystemSaveBranding(c fiber.Ctx) error {
	return h.saveSettingsFlash(c, "/system/settings", func(user uuid.UUID) error {
		req := dto.BrandingSettings{
			AppName: c.FormValue("app_name"), LogoURL: c.FormValue("logo_url"),
			FaviconURL: c.FormValue("favicon_url"), PrimaryColor: c.FormValue("primary_color"),
			EmailFooter: c.FormValue("email_footer"),
		}
		return h.services.System.SaveBranding(c.Context(), req, user, c.IP())
	})
}

func (h *Handler) SystemSaveBackupSchedule(c fiber.Ctx) error {
	return h.saveSettingsFlash(c, "/system/settings", func(user uuid.UUID) error {
		req := dto.BackupScheduleRequest{
			Enabled: c.FormValue("enabled") == "on", Cron: c.FormValue("cron"),
			RetentionDays: parseIntDefault(c.FormValue("retention_days"), 30),
		}
		return h.services.System.SaveBackupSchedule(c.Context(), req, user, c.IP())
	})
}

func (h *Handler) saveSettingsFlash(c fiber.Ctx, redirect string, fn func(uuid.UUID) error) error {
	user := middleware.GetUser(c)
	if err := fn(user.ID); err != nil {
		h.flash(c, err.Error(), true)
	} else {
		h.flash(c, "Settings saved", false)
	}
	return c.Redirect().To(redirect)
}

func (h *Handler) EmailTemplatesPage(c fiber.Ctx) error {
	templates, _ := h.services.System.ListEmailTemplates(c.Context())
	return h.render(c, fiber.StatusOK, web.EmailTemplatesPage{Templates: templates})
}

func (h *Handler) EmailTemplateUpdate(c fiber.Ctx) error {
	slug := c.Params("slug")
	req := dto.EmailTemplateRequest{Subject: c.FormValue("subject"), BodyHTML: c.FormValue("body_html")}
	user := middleware.GetUser(c)
	if err := h.services.System.UpdateEmailTemplate(c.Context(), slug, req, user.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
	} else {
		h.flash(c, "Template updated", false)
	}
	return c.Redirect().To("/system/templates")
}

func (h *Handler) BackupListPage(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	data, _ := h.services.Backup.ListBackups(c.Context(), page, 20)
	return h.render(c, fiber.StatusOK, web.BackupListPage{
		Data: data, Flash: c.Cookies("flash"), FlashType: c.Cookies("flash_type"),
	})
}

func (h *Handler) BackupCreate(c fiber.Ctx) error {
	user := middleware.GetUser(c)
	if _, err := h.services.Backup.CreateBackup(c.Context(), model.BackupManual, &user.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
	} else {
		h.flash(c, "Backup created successfully", false)
	}
	return c.Redirect().To("/system/backups")
}

func (h *Handler) BackupRestore(c fiber.Ctx) error {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.Status(400).SendString("Invalid backup")
	}
	user := middleware.GetUser(c)
	if err := h.services.Backup.RestoreBackup(c.Context(), id, user.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
	} else {
		h.flash(c, "Database restored from backup", false)
	}
	return c.Redirect().To("/system/backups")
}

func (h *Handler) BackupVerify(c fiber.Ctx) error {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.Status(400).SendString("Invalid backup")
	}
	ok, err := h.services.Backup.VerifyBackup(c.Context(), id)
	if err != nil {
		h.flash(c, err.Error(), true)
	} else if ok {
		h.flash(c, "Backup verified successfully", false)
	} else {
		h.flash(c, "Backup verification failed", true)
	}
	return c.Redirect().To("/system/backups")
}

func (h *Handler) BackupDownload(c fiber.Ctx) error {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.Status(400).SendString("Invalid backup")
	}
	path, name, err := h.services.Backup.GetBackupFilePath(c.Context(), id)
	if err != nil {
		return c.Status(404).SendString("Backup not found")
	}
	c.Set("Content-Disposition", "attachment; filename="+name)
	return c.SendFile(path)
}

func (h *Handler) AuditCenterPage(c fiber.Ctx) error {
	data, _ := h.services.System.SearchAuditLogs(c.Context(), h.parseAuditFilter(c))
	return h.render(c, fiber.StatusOK, web.AuditCenterPage{Data: data, Filter: h.parseAuditFilter(c)})
}

func (h *Handler) AuditExportCSV(c fiber.Ctx) error {
	f := h.parseAuditFilter(c)
	f.Page, f.PageSize = 1, 10000
	data, err := h.services.System.SearchAuditLogs(c.Context(), f)
	if err != nil {
		return c.Status(500).SendString("Export failed")
	}
	csv, err := export.AuditLogsCSV(data.Items)
	if err != nil {
		return c.Status(500).SendString("Export failed")
	}
	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", `attachment; filename="audit-logs.csv"`)
	return c.Send(csv)
}

func (h *Handler) parseAuditFilter(c fiber.Ctx) dto.AuditSearchFilter {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	return dto.AuditSearchFilter{
		Query: c.Query("q"), Action: c.Query("action"), EntityType: c.Query("module"),
		UserEmail: c.Query("user"), Page: page, PageSize: 50,
		From: parseOptionalDate(c.Query("from")), To: parseOptionalDateEnd(c.Query("to")),
	}
}

func (h *Handler) LicensePage(c fiber.Ctx) error {
	lic, _ := h.services.System.GetLicense(c.Context())
	sampleKey := h.services.System.GenerateLicenseKey()
	return h.render(c, fiber.StatusOK, web.LicensePage{
		License: lic, SampleKey: sampleKey, Flash: c.Cookies("flash"), FlashType: c.Cookies("flash_type"),
	})
}

func (h *Handler) LicenseActivate(c fiber.Ctx) error {
	req := dto.LicenseActivateRequest{
		LicenseKey: c.FormValue("license_key"), SchoolName: c.FormValue("school_name"),
		SchoolCode: c.FormValue("school_code"), RegisteredEmail: c.FormValue("registered_email"),
		ExpiresAt: c.FormValue("expires_at"),
	}
	user := middleware.GetUser(c)
	if _, err := h.services.System.ActivateLicense(c.Context(), req, user.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
	} else {
		h.flash(c, "License activated", false)
	}
	return c.Redirect().To("/system/license")
}

func (h *Handler) SecurityDashboardPage(c fiber.Ctx) error {
	stats, _ := h.services.System.SecurityDashboard(c.Context())
	return h.render(c, fiber.StatusOK, web.SecurityDashboardPage{Stats: stats})
}

func (h *Handler) InstallPage(c fiber.Ctx) error {
	status, _ := h.services.System.InstallStatus(c.Context())
	if status != nil && status.Installed {
		return c.Redirect().To("/login")
	}
	return h.render(c, fiber.StatusOK, web.InstallPage{Flash: c.Cookies("flash"), FlashType: c.Cookies("flash_type")})
}

func (h *Handler) InstallSubmit(c fiber.Ctx) error {
	req := dto.InstallRequest{
		AdminEmail: c.FormValue("admin_email"), AdminPassword: c.FormValue("admin_password"),
		AdminFirstName: c.FormValue("admin_first_name"), AdminLastName: c.FormValue("admin_last_name"),
		SchoolName: c.FormValue("school_name"), SchoolEmail: c.FormValue("school_email"),
		SchoolPhone: c.FormValue("school_phone"), SchoolAddress: c.FormValue("school_address"),
	}
	if errs := h.validate.Validate(req); len(errs) > 0 {
		h.flash(c, validator.FirstError(errs), true)
		return c.Redirect().To("/install")
	}
	if err := h.services.System.ValidatePasswordPolicy(c.Context(), req.AdminPassword); err != nil {
		h.flash(c, err.Error(), true)
		return c.Redirect().To("/install")
	}
	if err := h.services.System.RunInstall(c.Context(), req, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
		return c.Redirect().To("/install")
	}
	h.flash(c, "Installation complete. Please log in.", false)
	return c.Redirect().To("/login")
}

func parseIntDefault(s string, fallback int) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return fallback
	}
	return n
}
