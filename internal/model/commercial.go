package model

const (
	EntitySystemBackup  = "system_backup"
	EntitySystemSetting = "system_setting"
	EntityLicense       = "license"
	EntityLoginAttempt  = "login_attempt"
	EntityEmailTemplate = "email_template"

	PermSystemManage   = "system.manage"
	PermSystemBackup   = "system.backup"
	PermSystemAudit    = "system.audit"
	PermSystemLicense  = "system.license"
	PermSystemSecurity = "system.security"

	BackupManual    = "manual"
	BackupScheduled = "scheduled"
	BackupPending   = "pending"
	BackupCompleted = "completed"
	BackupFailed    = "failed"

	LicenseActive  = "active"
	LicenseExpired = "expired"
	LicenseRevoked = "revoked"
	LicensePending = "pending"

	SettingCategoryGeneral  = "general"
	SettingCategorySMTP     = "smtp"
	SettingCategorySMS      = "sms"
	SettingCategorySecurity = "security"
	SettingCategoryBranding = "branding"
	SettingCategoryPayment  = "payment"
)
