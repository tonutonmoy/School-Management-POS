package dto

import (
	"time"

	"github.com/google/uuid"
)

type AdmissionApplicationRequest struct {
	FirstName        string    `form:"first_name" validate:"required,min=2,max=100"`
	LastName         string    `form:"last_name" validate:"required,min=2,max=100"`
	DateOfBirth      time.Time `form:"date_of_birth" validate:"required"`
	Gender           string    `form:"gender" validate:"required,oneof=male female other"`
	BloodGroup       string    `form:"blood_group"`
	Religion         string    `form:"religion"`
	Nationality      string    `form:"nationality"`
	Phone            string    `form:"phone"`
	Email            string    `form:"email" validate:"omitempty,email"`
	Address          string    `form:"address"`
	FatherName       string    `form:"father_name"`
	FatherPhone      string    `form:"father_phone"`
	FatherOccupation string    `form:"father_occupation"`
	MotherName       string    `form:"mother_name"`
	MotherPhone      string    `form:"mother_phone"`
	MotherOccupation string    `form:"mother_occupation"`
	GuardianName     string    `form:"guardian_name"`
	GuardianPhone    string    `form:"guardian_phone"`
	PreviousSchool   string    `form:"previous_school"`
	PreviousClass    string    `form:"previous_class"`
	PreviousBoard    string    `form:"previous_board"`
	SessionID        uuid.UUID `form:"session_id" validate:"required"`
	ClassID          uuid.UUID `form:"class_id" validate:"required"`
	SectionID        uuid.UUID `form:"section_id"`
	AdmissionFee     float64   `form:"admission_fee"`
}

type AdmissionApplicationResponse struct {
	ID                 uuid.UUID  `json:"id"`
	ApplicationNumber  string     `json:"application_number"`
	TrackingToken      string     `json:"tracking_token"`
	Status             string     `json:"status"`
	FirstName          string     `json:"first_name"`
	LastName           string     `json:"last_name"`
	FullName           string     `json:"full_name"`
	DateOfBirth        time.Time  `json:"date_of_birth"`
	Gender             string     `json:"gender"`
	BloodGroup         string     `json:"blood_group"`
	Religion           string     `json:"religion"`
	Nationality        string     `json:"nationality"`
	Phone              string     `json:"phone"`
	Email              string     `json:"email"`
	Address            string     `json:"address"`
	FatherName         string     `json:"father_name"`
	FatherPhone        string     `json:"father_phone"`
	FatherOccupation   string     `json:"father_occupation"`
	MotherName         string     `json:"mother_name"`
	MotherPhone        string     `json:"mother_phone"`
	MotherOccupation   string     `json:"mother_occupation"`
	GuardianName       string     `json:"guardian_name"`
	GuardianPhone      string     `json:"guardian_phone"`
	PreviousSchool     string     `json:"previous_school"`
	PreviousClass      string     `json:"previous_class"`
	PreviousBoard      string     `json:"previous_board"`
	SessionID          *uuid.UUID `json:"session_id"`
	SessionName        string     `json:"session_name"`
	ClassID            *uuid.UUID `json:"class_id"`
	ClassName          string     `json:"class_name"`
	SectionID          *uuid.UUID `json:"section_id"`
	SectionName        string     `json:"section_name"`
	AdmissionFeeAmount float64    `json:"admission_fee_amount"`
	PaymentStatus      string     `json:"payment_status"`
	PaymentReference   string     `json:"payment_reference"`
	ReceiptNumber      string     `json:"receipt_number"`
	ReviewNotes        string     `json:"review_notes"`
	ReviewedByName     string     `json:"reviewed_by_name"`
	ReviewedAt         *time.Time `json:"reviewed_at"`
	StudentID          *uuid.UUID `json:"student_id"`
	Documents          []AdmissionDocumentResponse `json:"documents,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

type AdmissionDocumentResponse struct {
	ID         uuid.UUID `json:"id"`
	DocType    string    `json:"doc_type"`
	FileName   string    `json:"file_name"`
	FileURL    string    `json:"file_url"`
	CreatedAt  time.Time `json:"created_at"`
}

type AdmissionSearchFilter struct {
	Query         string
	Status        string
	SessionID     *uuid.UUID
	ClassID       *uuid.UUID
	PaymentStatus string
	From          time.Time
	To            time.Time
	Page          int
	PageSize      int
}

type PaginatedAdmissionApplications struct {
	Items      []AdmissionApplicationResponse `json:"items"`
	Total      int64                          `json:"total"`
	Page       int                            `json:"page"`
	PageSize   int                            `json:"page_size"`
	TotalPages int                            `json:"total_pages"`
}

type AdmissionReviewRequest struct {
	Notes string `form:"review_notes"`
}

type AdmissionPaymentRequest struct {
	PaymentReference string  `form:"payment_reference"`
	Amount           float64 `form:"amount"`
}

type AdmissionTrackRequest struct {
	ApplicationNumber string `form:"application_number" validate:"required"`
	TrackingToken     string `form:"tracking_token" validate:"required"`
}

type AdmissionDashboardStats struct {
	TotalApplications int64 `json:"total_applications"`
	PendingCount      int64 `json:"pending_count"`
	UnderReviewCount  int64 `json:"under_review_count"`
	ApprovedCount     int64 `json:"approved_count"`
	AdmittedCount     int64 `json:"admitted_count"`
	RejectedCount     int64 `json:"rejected_count"`
	TodayApplications int64 `json:"today_applications"`
}

type WebsiteDashboardStats struct {
	AdmissionApplications int64 `json:"admission_applications"`
	WebsiteVisitors       int64 `json:"website_visitors"`
	ContactInquiries      int64 `json:"contact_inquiries"`
	NewContacts           int64 `json:"new_contacts"`
	PendingAdmissions     int64 `json:"pending_admissions"`
}

type WebsiteSettingsRequest struct {
	SiteName               string `form:"site_name" validate:"required"`
	Tagline                string `form:"tagline"`
	PrimaryColor           string `form:"primary_color"`
	SecondaryColor         string `form:"secondary_color"`
	FacebookURL            string `form:"facebook_url"`
	TwitterURL             string `form:"twitter_url"`
	InstagramURL           string `form:"instagram_url"`
	YoutubeURL             string `form:"youtube_url"`
	ContactEmail           string `form:"contact_email"`
	ContactPhone           string `form:"contact_phone"`
	ContactAddress         string `form:"contact_address"`
	DefaultMetaTitle       string `form:"default_meta_title"`
	DefaultMetaDescription string `form:"default_meta_description"`
}

type WebsiteSettingsResponse struct {
	ID                     uuid.UUID `json:"id"`
	SiteName               string    `json:"site_name"`
	Tagline                string    `json:"tagline"`
	LogoURL                string    `json:"logo_url"`
	FaviconURL             string    `json:"favicon_url"`
	PrimaryColor           string    `json:"primary_color"`
	SecondaryColor         string    `json:"secondary_color"`
	FacebookURL            string    `json:"facebook_url"`
	TwitterURL             string    `json:"twitter_url"`
	InstagramURL           string    `json:"instagram_url"`
	YoutubeURL             string    `json:"youtube_url"`
	ContactEmail           string    `json:"contact_email"`
	ContactPhone           string    `json:"contact_phone"`
	ContactAddress         string    `json:"contact_address"`
	DefaultMetaTitle       string    `json:"default_meta_title"`
	DefaultMetaDescription string    `json:"default_meta_description"`
}

type WebsitePageRequest struct {
	Slug            string `form:"slug" validate:"required"`
	Title           string `form:"title" validate:"required"`
	PageType        string `form:"page_type"`
	Content         string `form:"content"`
	MetaTitle       string `form:"meta_title"`
	MetaDescription string `form:"meta_description"`
	IsPublished     bool   `form:"is_published"`
	SortOrder       int    `form:"sort_order"`
}

type WebsitePageResponse struct {
	ID              uuid.UUID `json:"id"`
	Slug            string    `json:"slug"`
	Title           string    `json:"title"`
	PageType        string    `json:"page_type"`
	Content         string    `json:"content"`
	MetaTitle       string    `json:"meta_title"`
	MetaDescription string    `json:"meta_description"`
	OGImage         string    `json:"og_image"`
	IsPublished     bool      `json:"is_published"`
	SortOrder       int       `json:"sort_order"`
	Blocks          []WebsiteBlockResponse `json:"blocks,omitempty"`
}

type WebsiteBlockRequest struct {
	PageID    *uuid.UUID `form:"page_id"`
	BlockType string     `form:"block_type"`
	Title     string     `form:"title"`
	Content   string     `form:"content"`
	SortOrder int        `form:"sort_order"`
	IsActive  bool       `form:"is_active"`
}

type WebsiteBlockResponse struct {
	ID        uuid.UUID  `json:"id"`
	PageID    *uuid.UUID `json:"page_id"`
	BlockType string     `json:"block_type"`
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	ImageURL  string     `json:"image_url"`
	SortOrder int        `json:"sort_order"`
	IsActive  bool       `json:"is_active"`
}

type WebsiteBannerRequest struct {
	Title     string `form:"title" validate:"required"`
	Subtitle  string `form:"subtitle"`
	LinkURL   string `form:"link_url"`
	SortOrder int    `form:"sort_order"`
	IsActive  bool   `form:"is_active"`
}

type WebsiteBannerResponse struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	Subtitle  string    `json:"subtitle"`
	ImageURL  string    `json:"image_url"`
	LinkURL   string    `json:"link_url"`
	SortOrder int       `json:"sort_order"`
	IsActive  bool      `json:"is_active"`
}

type WebsiteGalleryRequest struct {
	Title     string `form:"title" validate:"required"`
	Caption   string `form:"caption"`
	SortOrder int    `form:"sort_order"`
	IsActive  bool   `form:"is_active"`
}

type WebsiteGalleryResponse struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	Caption   string    `json:"caption"`
	ImageURL  string    `json:"image_url"`
	SortOrder int       `json:"sort_order"`
	IsActive  bool      `json:"is_active"`
}

type NewsRequest struct {
	Slug            string     `form:"slug" validate:"required"`
	Title           string     `form:"title" validate:"required"`
	Excerpt         string     `form:"excerpt"`
	Body            string     `form:"body" validate:"required"`
	IsPublished     bool       `form:"is_published"`
	PublishedAt     *time.Time `form:"published_at"`
	MetaTitle       string     `form:"meta_title"`
	MetaDescription string     `form:"meta_description"`
}

type NewsResponse struct {
	ID              uuid.UUID  `json:"id"`
	Slug            string     `json:"slug"`
	Title           string     `json:"title"`
	Excerpt         string     `json:"excerpt"`
	Body            string     `json:"body"`
	ImageURL        string     `json:"image_url"`
	IsPublished     bool       `json:"is_published"`
	PublishedAt     *time.Time `json:"published_at"`
	MetaTitle       string     `json:"meta_title"`
	MetaDescription string     `json:"meta_description"`
	CreatedAt       time.Time  `json:"created_at"`
}

type EventRequest struct {
	Slug            string    `form:"slug" validate:"required"`
	Title           string    `form:"title" validate:"required"`
	Description     string    `form:"description" validate:"required"`
	Location        string    `form:"location"`
	EventDate       time.Time `form:"event_date" validate:"required"`
	EndDate         *time.Time `form:"end_date"`
	IsPublished     bool      `form:"is_published"`
	MetaTitle       string    `form:"meta_title"`
	MetaDescription string    `form:"meta_description"`
}

type EventResponse struct {
	ID              uuid.UUID  `json:"id"`
	Slug            string     `json:"slug"`
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	Location        string     `json:"location"`
	EventDate       time.Time  `json:"event_date"`
	EndDate         *time.Time `json:"end_date"`
	ImageURL        string     `json:"image_url"`
	IsPublished     bool       `json:"is_published"`
	MetaTitle       string     `json:"meta_title"`
	MetaDescription string     `json:"meta_description"`
	CreatedAt       time.Time  `json:"created_at"`
}

type DownloadRequest struct {
	Slug        string `form:"slug" validate:"required"`
	Title       string `form:"title" validate:"required"`
	Description string `form:"description"`
	Category    string `form:"category"`
	IsPublished bool   `form:"is_published"`
}

type DownloadResponse struct {
	ID            uuid.UUID `json:"id"`
	Slug          string    `json:"slug"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Category      string    `json:"category"`
	FileURL       string    `json:"file_url"`
	FileName      string    `json:"file_name"`
	DownloadCount int       `json:"download_count"`
	IsPublished   bool      `json:"is_published"`
	CreatedAt     time.Time `json:"created_at"`
}

type ContactFormRequest struct {
	Name    string `form:"name" validate:"required"`
	Email   string `form:"email" validate:"required,email"`
	Phone   string `form:"phone"`
	Subject string `form:"subject" validate:"required"`
	Message string `form:"message" validate:"required"`
}

type ContactMessageResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Subject   string    `json:"subject"`
	Message   string    `json:"message"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type PaginatedContactMessages struct {
	Items      []ContactMessageResponse `json:"items"`
	Total      int64                    `json:"total"`
	Page       int                      `json:"page"`
	PageSize   int                      `json:"page_size"`
	TotalPages int                      `json:"total_pages"`
}

type PaginatedNews struct {
	Items []NewsResponse `json:"items"`
	Total int64          `json:"total"`
	Page  int            `json:"page"`
	PageSize int         `json:"page_size"`
}

type PaginatedEvents struct {
	Items []EventResponse `json:"items"`
	Total int64           `json:"total"`
	Page  int             `json:"page"`
	PageSize int          `json:"page_size"`
}

type PublicSiteData struct {
	Settings *WebsiteSettingsResponse
	Banners  []WebsiteBannerResponse
	News     []NewsResponse
	Events   []EventResponse
}
