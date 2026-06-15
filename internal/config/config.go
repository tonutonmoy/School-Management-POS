package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	JWT      JWTConfig
	CSRF     CSRFConfig
	Rate     RateLimitConfig
	R2       R2Config
	Admin    AdminConfig
	Login    LoginConfig
	SMTP     SMTPConfig
	SMS      SMSConfig
	Backup   BackupConfig
	LogLevel string
}

type AppConfig struct {
	Env  string
	Port string
	URL  string
	Name string
}

type DatabaseConfig struct {
	URL string
}

type JWTConfig struct {
	Secret     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

type CSRFConfig struct {
	Secret string
}

type RateLimitConfig struct {
	Max        int
	Expiration time.Duration
}

type R2Config struct {
	AccountID       string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	Endpoint        string
	PublicURL       string
	Enabled         bool
}

type AdminConfig struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
}

type LoginConfig struct {
	Email    string
	Password string
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	Enabled  bool
}

type SMSConfig struct {
	Provider string
	APIKey   string
	SenderID string
	Enabled  bool
}

type BackupConfig struct {
	Dir string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		App: AppConfig{
			Env:  getEnv("APP_ENV", "development"),
			Port: getEnv("APP_PORT", "8085"),
			URL:  getEnv("APP_URL", "http://localhost:8085"),
			Name: getEnv("APP_NAME", "School Management System"),
		},
		Database: DatabaseConfig{
			URL: getEnv("DATABASE_URL", ""),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", ""),
		},
		CSRF: CSRFConfig{
			Secret: getEnv("CSRF_SECRET", ""),
		},
		Rate: RateLimitConfig{
			Max: getEnvInt("RATE_LIMIT_MAX", 100),
		},
		R2: R2Config{
			AccountID:       getEnv("R2_ACCOUNT_ID", ""),
			AccessKeyID:     getEnv("R2_ACCESS_KEY_ID", ""),
			SecretAccessKey: getEnv("R2_SECRET_ACCESS_KEY", ""),
			BucketName:      getEnv("R2_BUCKET_NAME", ""),
			Endpoint:        getEnv("R2_ENDPOINT", ""),
			PublicURL:       getEnv("R2_PUBLIC_URL", ""),
		},
		Admin: AdminConfig{
			Email:     getEnv("ADMIN_EMAIL", "admin@school.local"),
			Password:  getEnv("ADMIN_PASSWORD", "Admin@123456"),
			FirstName: getEnv("ADMIN_FIRST_NAME", "System"),
			LastName:  getEnv("ADMIN_LAST_NAME", "Administrator"),
		},
		Login: LoginConfig{
			Email:    getEnv("LOGIN_EMAIL", getEnv("ADMIN_EMAIL", "admin@school.local")),
			Password: getEnv("LOGIN_PASSWORD", getEnv("ADMIN_PASSWORD", "Admin@123456")),
		},
		LogLevel: getEnv("LOG_LEVEL", "info"),
	}

	cfg.SMTP = SMTPConfig{
		Host:     getEnv("SMTP_HOST", ""),
		Port:     getEnvInt("SMTP_PORT", 587),
		Username: getEnv("SMTP_USERNAME", ""),
		Password: getEnv("SMTP_PASSWORD", ""),
		From:     getEnv("SMTP_FROM", ""),
		Enabled:  getEnv("SMTP_HOST", "") != "",
	}
	cfg.SMS = SMSConfig{
		Provider: getEnv("SMS_PROVIDER", "log"),
		APIKey:   getEnv("SMS_API_KEY", ""),
		SenderID: getEnv("SMS_SENDER_ID", ""),
		Enabled:  getEnv("SMS_API_KEY", "") != "",
	}
	cfg.Backup = BackupConfig{
		Dir: getEnv("BACKUP_DIR", "./backups"),
	}

	var err error
	cfg.JWT.AccessTTL, err = time.ParseDuration(getEnv("JWT_ACCESS_TTL", "15m"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_ACCESS_TTL: %w", err)
	}
	cfg.JWT.RefreshTTL, err = time.ParseDuration(getEnv("JWT_REFRESH_TTL", "168h"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_REFRESH_TTL: %w", err)
	}
	cfg.Rate.Expiration, err = time.ParseDuration(getEnv("RATE_LIMIT_EXPIRATION", "1m"))
	if err != nil {
		return nil, fmt.Errorf("invalid RATE_LIMIT_EXPIRATION: %w", err)
	}

	cfg.R2.Enabled = cfg.R2.AccessKeyID != "" && cfg.R2.SecretAccessKey != "" && cfg.R2.BucketName != ""

	if cfg.Database.URL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if len(cfg.JWT.Secret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}
	if len(cfg.CSRF.Secret) < 16 {
		return nil, fmt.Errorf("CSRF_SECRET must be at least 16 characters")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
