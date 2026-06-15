package notify

import (
	"context"
	"fmt"
	"log/slog"
	"net/smtp"
	"time"
)

// SMSProvider abstracts Bangladeshi and generic SMS gateways.
type SMSProvider interface {
	Name() string
	Send(ctx context.Context, phone, message string) error
}

// EmailProvider abstracts SMTP and transactional email services.
type EmailProvider interface {
	Name() string
	Send(ctx context.Context, to, subject, body string) error
}

// LogSMSProvider logs SMS instead of sending (development / fallback).
type LogSMSProvider struct {
	logger *slog.Logger
}

func NewLogSMSProvider(logger *slog.Logger) *LogSMSProvider {
	return &LogSMSProvider{logger: logger}
}

func (p *LogSMSProvider) Name() string { return "log" }

func (p *LogSMSProvider) Send(ctx context.Context, phone, message string) error {
	p.logger.Info("sms", "phone", phone, "message", message)
	return nil
}

// BulkSMSBDProvider stub for bulkSMSBD-style Bangladeshi gateways.
type BulkSMSBDProvider struct {
	APIKey  string
	Sender  string
	BaseURL string
	logger  *slog.Logger
}

func NewBulkSMSBDProvider(apiKey, sender, baseURL string, logger *slog.Logger) *BulkSMSBDProvider {
	return &BulkSMSBDProvider{APIKey: apiKey, Sender: sender, BaseURL: baseURL, logger: logger}
}

func (p *BulkSMSBDProvider) Name() string { return "bulksmsbd" }

func (p *BulkSMSBDProvider) Send(ctx context.Context, phone, message string) error {
	if p.APIKey == "" {
		p.logger.Info("bulksmsbd stub", "phone", phone, "message", message)
		return nil
	}
	// Queue-ready: real HTTP integration plugs in here.
	p.logger.Info("bulksmsbd send", "phone", phone)
	return nil
}

// LogEmailProvider logs emails instead of sending.
type LogEmailProvider struct {
	logger *slog.Logger
}

func NewLogEmailProvider(logger *slog.Logger) *LogEmailProvider {
	return &LogEmailProvider{logger: logger}
}

func (p *LogEmailProvider) Name() string { return "log" }

func (p *LogEmailProvider) Send(ctx context.Context, to, subject, body string) error {
	p.logger.Info("email", "to", to, "subject", subject)
	return nil
}

// SMTPProvider stub for production SMTP.
type SMTPProvider struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	logger   *slog.Logger
}

func NewSMTPProvider(host string, port int, user, pass, from string, logger *slog.Logger) *SMTPProvider {
	return &SMTPProvider{Host: host, Port: port, Username: user, Password: pass, From: from, logger: logger}
}

func (p *SMTPProvider) Name() string { return "smtp" }

func (p *SMTPProvider) Send(ctx context.Context, to, subject, body string) error {
	if p.Host == "" {
		p.logger.Info("smtp stub", "to", to, "subject", subject)
		return nil
	}
	addr := fmt.Sprintf("%s:%d", p.Host, p.Port)
	msg := []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		p.From, to, subject, body))
	auth := smtp.PlainAuth("", p.Username, p.Password, p.Host)
	if err := smtp.SendMail(addr, auth, p.From, []string{to}, msg); err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}
	p.logger.Info("smtp sent", "to", to, "subject", subject)
	return nil
}

func Now() time.Time { return time.Now() }
