package dto

import (
	"time"

	"github.com/google/uuid"
)

type LoginRequest struct {
	Email    string `json:"email" form:"email" validate:"required,email"`
	Password string `json:"password" form:"password" validate:"required,min=8"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" form:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Token           string `json:"token" form:"token" validate:"required"`
	Password        string `json:"password" form:"password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" form:"confirm_password" validate:"required,eqfield=Password"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" form:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" form:"new_password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" form:"confirm_password" validate:"required,eqfield=NewPassword"`
}

type CreateUserRequest struct {
	Email     string    `json:"email" form:"email" validate:"required,email"`
	Password  string    `json:"password" form:"password" validate:"required,min=8"`
	FirstName string    `json:"first_name" form:"first_name" validate:"required,min=2,max=100"`
	LastName  string    `json:"last_name" form:"last_name" validate:"required,min=2,max=100"`
	Phone     string    `json:"phone" form:"phone" validate:"omitempty,max=30"`
	RoleID    uuid.UUID `json:"role_id" form:"role_id" validate:"required"`
	IsActive  bool      `json:"is_active" form:"is_active"`
}

type UpdateUserRequest struct {
	Email     string    `json:"email" form:"email" validate:"required,email"`
	FirstName string    `json:"first_name" form:"first_name" validate:"required,min=2,max=100"`
	LastName  string    `json:"last_name" form:"last_name" validate:"required,min=2,max=100"`
	Phone     string    `json:"phone" form:"phone" validate:"omitempty,max=30"`
	RoleID    uuid.UUID `json:"role_id" form:"role_id" validate:"required"`
	IsActive  bool      `json:"is_active" form:"is_active"`
}

type CreateRoleRequest struct {
	Name        string `json:"name" form:"name" validate:"required,min=2,max=100"`
	Slug        string `json:"slug" form:"slug" validate:"required,min=2,max=100,alphanumdash"`
	Description string `json:"description" form:"description" validate:"omitempty,max=500"`
}

type UpdateRoleRequest struct {
	Name        string `json:"name" form:"name" validate:"required,min=2,max=100"`
	Slug        string `json:"slug" form:"slug" validate:"required,min=2,max=100,alphanumdash"`
	Description string `json:"description" form:"description" validate:"omitempty,max=500"`
}

type AssignPermissionsRequest struct {
	PermissionIDs []uuid.UUID `json:"permission_ids" form:"permission_ids" validate:"required,min=1,dive,required"`
}

type SchoolSetupRequest struct {
	Name    string `json:"name" form:"name" validate:"required,min=2,max=255"`
	Address string `json:"address" form:"address" validate:"omitempty,max=1000"`
	Email   string `json:"email" form:"email" validate:"omitempty,email"`
	Phone   string `json:"phone" form:"phone" validate:"omitempty,max=30"`
	Website string `json:"website" form:"website" validate:"omitempty,url"`
}

type AcademicSessionRequest struct {
	Name      string    `json:"name" form:"name" validate:"required,min=2,max=150"`
	StartDate time.Time `json:"start_date" form:"start_date" validate:"required"`
	EndDate   time.Time `json:"end_date" form:"end_date" validate:"required,gtfield=StartDate"`
	IsActive  bool      `json:"is_active" form:"is_active"`
}

type DashboardStats struct {
	TotalStudents           int64
	TotalTeachers           int64
	TotalStaff              int64
	ActiveSessionName       string
	HasActiveSession        bool
	NewAdmissionsThisMonth  int64
	StudentsByClass         []ClassStudentCount
	RecentActivities        []ActivityItem
}

type ClassStudentCount struct {
	ClassID      uuid.UUID
	ClassName    string
	StudentCount int64
}

type ActivityItem struct {
	ID          uuid.UUID
	Action      string
	EntityType  string
	UserEmail   string
	UserName    string
	Description string
	CreatedAt   time.Time
}

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Phone     string    `json:"phone,omitempty"`
	RoleID    uuid.UUID `json:"role_id"`
	RoleName  string    `json:"role_name"`
	RoleSlug  string    `json:"role_slug"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RoleResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description,omitempty"`
	IsSystem    bool      `json:"is_system"`
	Permissions []string  `json:"permissions,omitempty"`
}

type PermissionResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description,omitempty"`
	Module      string    `json:"module"`
}

type SchoolResponse struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	LogoURL string    `json:"logo_url,omitempty"`
	Address string    `json:"address,omitempty"`
	Email   string    `json:"email,omitempty"`
	Phone   string    `json:"phone,omitempty"`
	Website string    `json:"website,omitempty"`
}

type AcademicSessionResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	IsActive  bool      `json:"is_active"`
}

type PaginatedUsers struct {
	Items      []UserResponse `json:"items"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}

type AuthUser struct {
	ID          uuid.UUID
	Email       string
	FirstName   string
	LastName    string
	RoleID      uuid.UUID
	RoleName    string
	RoleSlug    string
	Permissions []string
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}
