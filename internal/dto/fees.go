package dto

import (
	"time"

	"github.com/google/uuid"
)

type FeeTypeRequest struct {
	Name        string `form:"name" validate:"required,min=2,max=100"`
	Slug        string `form:"slug" validate:"required,min=2,max=100,alphanumdash"`
	Description string `form:"description" validate:"omitempty,max=500"`
	IsActive    bool   `form:"is_active"`
}

type FeeTypeResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description,omitempty"`
	IsActive    bool      `json:"is_active"`
}

type FeeStructureRequest struct {
	FeeTypeID uuid.UUID `form:"fee_type_id" validate:"required"`
	SessionID uuid.UUID `form:"session_id" validate:"required"`
	ClassID   uuid.UUID `form:"class_id" validate:"required"`
	SectionID uuid.UUID `form:"section_id" validate:"omitempty"`
	Amount    float64   `form:"amount" validate:"required,gte=0"`
	DueDay    int       `form:"due_day" validate:"required,gte=1,lte=28"`
	Frequency string    `form:"frequency" validate:"required,oneof=one_time monthly quarterly half_yearly yearly"`
	IsActive  bool      `form:"is_active"`
}

type FeeStructureResponse struct {
	ID          uuid.UUID  `json:"id"`
	FeeTypeID   uuid.UUID  `json:"fee_type_id"`
	FeeTypeName string     `json:"fee_type_name"`
	SessionID   uuid.UUID  `json:"session_id"`
	SessionName string     `json:"session_name"`
	ClassID     uuid.UUID  `json:"class_id"`
	ClassName   string     `json:"class_name"`
	SectionID   *uuid.UUID `json:"section_id,omitempty"`
	SectionName string     `json:"section_name,omitempty"`
	Amount      float64    `json:"amount"`
	DueDay      int        `json:"due_day"`
	Frequency   string     `json:"frequency"`
	IsActive    bool       `json:"is_active"`
}

type StudentDiscountRequest struct {
	StudentID      uuid.UUID `form:"student_id" validate:"required"`
	SessionID      uuid.UUID `form:"session_id" validate:"required"`
	DiscountType   string    `form:"discount_type" validate:"required,oneof=fixed percentage"`
	DiscountValue  float64   `form:"discount_value" validate:"required,gte=0"`
	Reason         string    `form:"reason" validate:"required,oneof=scholarship waiver sibling special"`
	Description    string    `form:"description" validate:"omitempty,max=500"`
	IsActive       bool      `form:"is_active"`
}

type StudentDiscountResponse struct {
	ID            uuid.UUID `json:"id"`
	StudentID     uuid.UUID `json:"student_id"`
	StudentName   string    `json:"student_name"`
	SessionID     uuid.UUID `json:"session_id"`
	DiscountType  string    `json:"discount_type"`
	DiscountValue float64   `json:"discount_value"`
	Reason        string    `json:"reason"`
	Description   string    `json:"description,omitempty"`
	IsActive      bool      `json:"is_active"`
}

type BillItemResponse struct {
	ID          uuid.UUID `json:"id"`
	FeeTypeName string    `json:"fee_type_name"`
	Description string    `json:"description"`
	Amount      float64   `json:"amount"`
}

type StudentBillResponse struct {
	ID             uuid.UUID          `json:"id"`
	InvoiceNumber  string             `json:"invoice_number"`
	StudentID      uuid.UUID          `json:"student_id"`
	StudentName    string             `json:"student_name"`
	AdmissionNo    string             `json:"admission_number"`
	ClassName      string             `json:"class_name"`
	SectionName    string             `json:"section_name"`
	BillPeriod     string             `json:"bill_period"`
	DueDate        time.Time          `json:"due_date"`
	Subtotal       float64            `json:"subtotal"`
	DiscountAmount float64            `json:"discount_amount"`
	TotalAmount    float64            `json:"total_amount"`
	PaidAmount     float64            `json:"paid_amount"`
	DueAmount      float64            `json:"due_amount"`
	Status         string             `json:"status"`
	Items          []BillItemResponse `json:"items,omitempty"`
	GeneratedAt    time.Time          `json:"generated_at"`
}

type GenerateBillsRequest struct {
	SessionID  uuid.UUID `form:"session_id" validate:"required"`
	ClassID    uuid.UUID `form:"class_id" validate:"omitempty"`
	SectionID  uuid.UUID `form:"section_id" validate:"omitempty"`
	StudentID  uuid.UUID `form:"student_id" validate:"omitempty"`
	BillPeriod string    `form:"bill_period" validate:"required"`
	Regenerate bool      `form:"regenerate"`
}

type PaymentRequest struct {
	StudentID     uuid.UUID `form:"student_id" validate:"required"`
	Amount        float64   `form:"amount" validate:"required,gt=0"`
	PaymentMethod string    `form:"payment_method" validate:"required,oneof=cash bank bkash nagad rocket card"`
	CollectionDate time.Time `form:"collection_date" validate:"required"`
	Remarks       string    `form:"remarks" validate:"omitempty,max=500"`
}

type PaymentAllocationInput struct {
	BillID uuid.UUID `form:"bill_id"`
	Amount float64   `form:"amount"`
}

type PaymentResponse struct {
	ID             uuid.UUID `json:"id"`
	PaymentNumber  string    `json:"payment_number"`
	StudentID      uuid.UUID `json:"student_id"`
	StudentName    string    `json:"student_name"`
	Amount         float64   `json:"amount"`
	PaymentMethod  string    `json:"payment_method"`
	CollectorName  string    `json:"collector_name"`
	CollectionDate time.Time `json:"collection_date"`
	Remarks        string    `json:"remarks,omitempty"`
	Status         string    `json:"status"`
	ReceiptID      *uuid.UUID `json:"receipt_id,omitempty"`
	ReceiptNumber  string    `json:"receipt_number,omitempty"`
}

type ReceiptResponse struct {
	ID            uuid.UUID `json:"id"`
	ReceiptNumber string    `json:"receipt_number"`
	PaymentID     uuid.UUID `json:"payment_id"`
	PaymentNumber string    `json:"payment_number"`
	StudentID     uuid.UUID `json:"student_id"`
	StudentName   string    `json:"student_name"`
	AdmissionNo   string    `json:"admission_number"`
	ClassName     string    `json:"class_name"`
	SectionName   string    `json:"section_name"`
	TotalAmount   float64   `json:"total_amount"`
	QRToken       string    `json:"qr_token"`
	IssuedAt      time.Time `json:"issued_at"`
	CollectorName string    `json:"collector_name"`
	SchoolName    string    `json:"school_name"`
	SchoolLogo    string    `json:"school_logo,omitempty"`
	Allocations   []PaymentAllocationDetail `json:"allocations,omitempty"`
}

type PaymentAllocationDetail struct {
	InvoiceNumber string  `json:"invoice_number"`
	Amount        float64 `json:"amount"`
}

type FinanceDashboardStats struct {
	TodayCollection      float64
	MonthlyCollection    float64
	OutstandingDues      float64
	StudentsWithDues     int64
	CollectionTrend      []FinanceTrendPoint
	DueTrend             []FinanceTrendPoint
	PaymentMethodBreakdown []PaymentMethodStat
}

type FinanceTrendPoint struct {
	Label  string
	Amount float64
}

type PaymentMethodStat struct {
	Method string
	Amount float64
	Count  int64
}

type BillSearchFilter struct {
	SessionID uuid.UUID
	ClassID   uuid.UUID
	SectionID uuid.UUID
	StudentID uuid.UUID
	Status    string
	Query     string
	Page      int
	PerPage   int
}

type PaginatedBills struct {
	Items      []StudentBillResponse
	Total      int64
	Page       int
	PerPage    int
	TotalPages int
}

type DueStudentRow struct {
	StudentID     uuid.UUID
	StudentName   string
	AdmissionNo   string
	ClassName     string
	SectionName   string
	TotalDue      float64
	OverdueAmount float64
	BillCount     int64
}

type FinanceReportFilter struct {
	From      time.Time
	To        time.Time
	SessionID uuid.UUID
	ClassID   uuid.UUID
	StudentID uuid.UUID
	FeeTypeID uuid.UUID
	Method    string
}

type StudentLedgerEntry struct {
	Date        time.Time
	Type        string
	Reference   string
	Description string
	Debit       float64
	Credit      float64
	Balance     float64
}

type ParentFeeSummary struct {
	StudentID      uuid.UUID
	StudentName    string
	CurrentDue     float64
	TotalPaid      float64
	RecentPayments []PaymentResponse
	Receipts       []ReceiptResponse
}
