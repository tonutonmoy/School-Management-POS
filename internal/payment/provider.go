package payment

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
)

// GatewayProvider abstracts Bangladeshi payment gateways.
type GatewayProvider interface {
	Slug() string
	Initiate(ctx context.Context, cfg GatewayConfig, req InitiateRequest) (*InitiateResult, error)
	VerifyWebhook(ctx context.Context, cfg GatewayConfig, payload WebhookData) (bool, error)
}

type GatewayConfig struct {
	IsSandbox  bool
	APIKey     string
	APISecret  string
	MerchantID string
	StoreID    string
	CallbackURL string
	SuccessURL  string
	FailURL     string
	BaseURL    string
}

type InitiateRequest struct {
	TransactionRef string
	Amount         float64
	Currency       string
	CustomerPhone  string
	CustomerEmail  string
	Description    string
}

type InitiateResult struct {
	GatewayRef  string
	RedirectURL string
	RawResponse map[string]any
}

type WebhookData struct {
	TransactionRef string
	GatewayRef     string
	Amount         float64
	Status         string
	Signature      string
	Raw            map[string]any
}

// SandboxProvider completes payments in sandbox for all gateways.
type SandboxProvider struct {
	SlugName string
	Logger   *slog.Logger
}

func NewSandboxProvider(slug string, logger *slog.Logger) *SandboxProvider {
	return &SandboxProvider{SlugName: slug, Logger: logger}
}

func (p *SandboxProvider) Slug() string { return p.SlugName }

func (p *SandboxProvider) Initiate(ctx context.Context, cfg GatewayConfig, req InitiateRequest) (*InitiateResult, error) {
	gwRef := fmt.Sprintf("%s-%s", p.SlugName, req.TransactionRef)
	redirect := fmt.Sprintf("%s?ref=%s&gateway=%s", cfg.SuccessURL, req.TransactionRef, p.SlugName)
	if cfg.SuccessURL == "" {
		redirect = fmt.Sprintf("/payments/success?ref=%s&gateway=%s", req.TransactionRef, p.SlugName)
	}
	p.Logger.Info("sandbox payment initiated", "gateway", p.SlugName, "ref", req.TransactionRef, "amount", req.Amount)
	return &InitiateResult{
		GatewayRef:  gwRef,
		RedirectURL: redirect,
		RawResponse: map[string]any{"sandbox": true, "gateway": p.SlugName},
	}, nil
}

func (p *SandboxProvider) VerifyWebhook(ctx context.Context, cfg GatewayConfig, payload WebhookData) (bool, error) {
	if cfg.APISecret == "" {
		return true, nil
	}
	expected := SignPayload(cfg.APISecret, payload.TransactionRef, payload.Amount)
	return hmac.Equal([]byte(payload.Signature), []byte(expected)), nil
}

func SignPayload(secret, ref string, amount float64) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(fmt.Sprintf("%s:%.2f", ref, amount)))
	return hex.EncodeToString(mac.Sum(nil))
}

type ProviderRegistry struct {
	providers map[string]GatewayProvider
}

func NewRegistry(logger *slog.Logger) *ProviderRegistry {
	r := &ProviderRegistry{providers: map[string]GatewayProvider{}}
	for _, slug := range []string{"bkash", "nagad", "sslcommerz"} {
		r.providers[slug] = NewSandboxProvider(slug, logger)
	}
	return r
}

func (r *ProviderRegistry) Get(slug string) GatewayProvider {
	return r.providers[slug]
}

func (r *ProviderRegistry) Register(slug string, p GatewayProvider) {
	r.providers[slug] = p
}
