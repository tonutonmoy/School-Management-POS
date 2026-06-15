package dto

import (
	"time"

	"github.com/google/uuid"
)

type PaymentGatewayRequest struct {
	Name        string `form:"name" validate:"required"`
	Slug        string `form:"slug" validate:"required,oneof=bkash nagad sslcommerz"`
	IsActive    bool   `form:"is_active"`
	IsSandbox   bool   `form:"is_sandbox"`
	APIKey      string `form:"api_key"`
	APISecret   string `form:"api_secret"`
	MerchantID  string `form:"merchant_id"`
	StoreID     string `form:"store_id"`
	CallbackURL string `form:"callback_url"`
	SuccessURL  string `form:"success_url"`
	FailURL     string `form:"fail_url"`
}

type PaymentGatewayResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	IsActive    bool      `json:"is_active"`
	IsSandbox   bool      `json:"is_sandbox"`
	MerchantID  string    `json:"merchant_id"`
	StoreID     string    `json:"store_id"`
	CallbackURL string    `json:"callback_url"`
	SuccessURL  string    `json:"success_url"`
	FailURL     string    `json:"fail_url"`
	HasCredentials bool   `json:"has_credentials"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type InitiatePaymentRequest struct {
	GatewaySlug    string    `form:"gateway" validate:"required"`
	PaymentType    string    `form:"payment_type" validate:"required"`
	StudentID      uuid.UUID `form:"student_id"`
	ApplicationID  uuid.UUID `form:"application_id"`
	Amount         float64   `form:"amount" validate:"required,gt=0"`
	IdempotencyKey string    `form:"idempotency_key"`
	BillIDs        []uuid.UUID
}

type InitiatePaymentResponse struct {
	TransactionID uuid.UUID `json:"transaction_id"`
	TransactionRef string   `json:"transaction_ref"`
	RedirectURL   string    `json:"redirect_url"`
	Amount        float64   `json:"amount"`
	Gateway       string    `json:"gateway"`
	Status        string    `json:"status"`
}

type GatewayTransactionResponse struct {
	ID              uuid.UUID  `json:"id"`
	TransactionRef  string     `json:"transaction_ref"`
	GatewayRef      string     `json:"gateway_ref"`
	GatewayName     string     `json:"gateway_name"`
	GatewaySlug     string     `json:"gateway_slug"`
	PaymentType     string     `json:"payment_type"`
	Amount          float64    `json:"amount"`
	Status          string     `json:"status"`
	StudentID       *uuid.UUID `json:"student_id"`
	StudentName     string     `json:"student_name"`
	PaymentID       *uuid.UUID `json:"payment_id"`
	ReceiptNumber   string     `json:"receipt_number"`
	ErrorMessage    string     `json:"error_message"`
	SignatureVerified bool     `json:"signature_verified"`
	CreatedAt       time.Time  `json:"created_at"`
	CompletedAt     *time.Time `json:"completed_at"`
}

type PaymentDashboardStats struct {
	TodayCollection    float64 `json:"today_collection"`
	GatewayCollection  float64 `json:"gateway_collection"`
	FailedPayments     int64   `json:"failed_payments"`
	PendingPayments    int64   `json:"pending_payments"`
	TodayTransactions  int64   `json:"today_transactions"`
}

type PaginatedGatewayTransactions struct {
	Items      []GatewayTransactionResponse `json:"items"`
	Total      int64                        `json:"total"`
	Page       int                          `json:"page"`
	PageSize   int                          `json:"page_size"`
	TotalPages int                          `json:"total_pages"`
}

type GatewayCollectionReport struct {
	GatewaySlug string  `json:"gateway_slug"`
	GatewayName string  `json:"gateway_name"`
	Count       int64   `json:"count"`
	TotalAmount float64 `json:"total_amount"`
}

type PaymentReportFilter struct {
	Query       string
	Status      string
	GatewaySlug string
	PaymentType string
	From        time.Time
	To          time.Time
	Page        int
	PageSize    int
}

type RefundRequest struct {
	PaymentID uuid.UUID `form:"payment_id" validate:"required"`
	Amount    float64   `form:"amount" validate:"required,gt=0"`
	Reason    string    `form:"reason" validate:"required"`
}

type RefundResponse struct {
	ID            uuid.UUID  `json:"id"`
	PaymentID     uuid.UUID  `json:"payment_id"`
	Amount        float64    `json:"amount"`
	Status        string     `json:"status"`
	Reason        string     `json:"reason"`
	RequestedByName string `json:"requested_by_name"`
	CreatedAt     time.Time  `json:"created_at"`
	ProcessedAt   *time.Time `json:"processed_at"`
}

type PaginatedRefunds struct {
	Items      []RefundResponse `json:"items"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalPages int              `json:"total_pages"`
}

type ParentPayNowData struct {
	StudentID   uuid.UUID
	StudentName string
	CurrentDue  float64
	Gateways    []PaymentGatewayResponse
	Bills       []StudentBillResponse
}

type WebhookPayload struct {
	TransactionRef string
	GatewayRef     string
	Status         string
	Amount         float64
	Signature      string
	RawBody        map[string]any
}
