package dto

import (
	"time"

	"github.com/google/uuid"
)

type ParentResponse struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Email       string    `json:"email"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	Phone       string    `json:"phone"`
	Address     string    `json:"address"`
	Occupation  string    `json:"occupation"`
	IsActive    bool      `json:"is_active"`
	ChildCount  int       `json:"child_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ParentChildResponse struct {
	ID              uuid.UUID `json:"id"`
	StudentID       uuid.UUID `json:"student_id"`
	StudentName     string    `json:"student_name"`
	AdmissionNumber string    `json:"admission_number"`
	RollNumber      string    `json:"roll_number"`
	ClassName       string    `json:"class_name"`
	Relationship    string    `json:"relationship"`
	IsPrimary       bool      `json:"is_primary"`
}

type CreateParentRequest struct {
	Email        string      `json:"email" validate:"required,email"`
	Password     string      `json:"password" validate:"required,min=8"`
	FirstName    string      `json:"first_name" validate:"required"`
	LastName     string      `json:"last_name" validate:"required"`
	Phone        string      `json:"phone"`
	Address      string      `json:"address"`
	Occupation   string      `json:"occupation"`
	StudentLinks []ParentLinkInput `json:"student_links"`
}

type ParentLinkInput struct {
	StudentID    uuid.UUID `json:"student_id" validate:"required"`
	Relationship string    `json:"relationship"`
	IsPrimary    bool      `json:"is_primary"`
}

type UpdateParentRequest struct {
	FirstName  string `json:"first_name" validate:"required"`
	LastName   string `json:"last_name" validate:"required"`
	Phone      string `json:"phone"`
	Address    string `json:"address"`
	Occupation string `json:"occupation"`
	IsActive   bool   `json:"is_active"`
}

type UpdateParentProfileRequest struct {
	Phone      string `json:"phone"`
	Address    string `json:"address"`
	Occupation string `json:"occupation"`
}

type ParentDashboardStats struct {
	ChildrenCount      int                        `json:"children_count"`
	AttendancePct      float64                    `json:"attendance_pct"`
	CurrentDue         float64                    `json:"current_due"`
	LatestExamTitle    string                     `json:"latest_exam_title"`
	LatestExamGPA      float64                    `json:"latest_exam_gpa"`
	LatestExamPosition int                        `json:"latest_exam_position"`
	Children           []ParentChildResponse      `json:"children"`
	RecentActivities   []ParentActivityItem       `json:"recent_activities"`
	UnreadNotices      int                        `json:"unread_notices"`
	UnreadNotifications int                       `json:"unread_notifications"`
}

type ParentActivityItem struct {
	ID          uuid.UUID `json:"id"`
	Category    string    `json:"category"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StudentName string    `json:"student_name"`
	CreatedAt   time.Time `json:"created_at"`
}

type PaginatedParents struct {
	Items      []ParentResponse `json:"items"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalPages int              `json:"total_pages"`
}

type ParentAttendanceView struct {
	Summary   *StudentAttendanceSummary     `json:"summary"`
	Daily     []StudentAttendanceResponse     `json:"daily"`
	Monthly   []StudentAttendanceResponse   `json:"monthly"`
	History   []StudentAttendanceResponse   `json:"history"`
	Student   *StudentResponse                `json:"student"`
}

type NoticeRequest struct {
	Title          string     `json:"title" validate:"required,max=200"`
	Body           string     `json:"body" validate:"required"`
	NoticeType     string     `json:"notice_type" validate:"required"`
	TargetAudience string     `json:"target_audience"`
	PublishAt      time.Time  `json:"publish_at"`
	ExpiresAt      *time.Time `json:"expires_at"`
	IsPublished    bool       `json:"is_published"`
}

type NoticeResponse struct {
	ID             uuid.UUID  `json:"id"`
	Title          string     `json:"title"`
	Body           string     `json:"body"`
	NoticeType     string     `json:"notice_type"`
	TargetAudience string     `json:"target_audience"`
	PublishAt      time.Time  `json:"publish_at"`
	ExpiresAt      *time.Time `json:"expires_at"`
	IsPublished    bool       `json:"is_published"`
	IsRead         bool       `json:"is_read"`
	CreatedByName  string     `json:"created_by_name"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type PaginatedNotices struct {
	Items      []NoticeResponse `json:"items"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalPages int              `json:"total_pages"`
}

type NotificationResponse struct {
	ID            uuid.UUID  `json:"id"`
	Title         string     `json:"title"`
	Body          string     `json:"body"`
	Category      string     `json:"category"`
	ReferenceType string     `json:"reference_type"`
	ReferenceID   *uuid.UUID `json:"reference_id"`
	IsRead        bool       `json:"is_read"`
	ReadAt        *time.Time `json:"read_at"`
	CreatedAt     time.Time  `json:"created_at"`
}

type PaginatedNotifications struct {
	Items      []NotificationResponse `json:"items"`
	Total      int64                  `json:"total"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"page_size"`
	TotalPages int                  `json:"total_pages"`
}

type CommunicationDashboardStats struct {
	SMSSentToday        int64   `json:"sms_sent_today"`
	EmailsSentToday     int64   `json:"emails_sent_today"`
	NotificationCount   int64   `json:"notification_count"`
	DeliverySuccessRate float64 `json:"delivery_success_rate"`
	RecentSMS           []SMSLogResponse    `json:"recent_sms"`
	RecentEmails        []EmailLogResponse  `json:"recent_emails"`
}

type SMSLogResponse struct {
	ID             uuid.UUID  `json:"id"`
	RecipientPhone string     `json:"recipient_phone"`
	Message        string     `json:"message"`
	EventType      string     `json:"event_type"`
	Provider       string     `json:"provider"`
	Status         string     `json:"status"`
	SentAt         *time.Time `json:"sent_at"`
	CreatedAt      time.Time  `json:"created_at"`
}

type EmailLogResponse struct {
	ID             uuid.UUID  `json:"id"`
	RecipientEmail string     `json:"recipient_email"`
	Subject        string     `json:"subject"`
	EventType      string     `json:"event_type"`
	Provider       string     `json:"provider"`
	Status         string     `json:"status"`
	SentAt         *time.Time `json:"sent_at"`
	CreatedAt      time.Time  `json:"created_at"`
}

type NoticeFilter struct {
	Query      string
	NoticeType string
	Page       int
	PageSize   int
}

type NotificationFilter struct {
	UnreadOnly bool
	Page       int
	PageSize   int
}
