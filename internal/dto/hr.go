package dto

import (
	"time"

	"github.com/google/uuid"
)

type DepartmentRequest struct {
	Name        string `form:"name" validate:"required,min=2,max=100"`
	Slug        string `form:"slug" validate:"required,min=2,max=100,alphanumdash"`
	Description string `form:"description" validate:"omitempty,max=500"`
	DeptType    string `form:"dept_type" validate:"omitempty,oneof=academic employee"`
}

type DesignationRequest struct {
	Name        string `form:"name" validate:"required,min=2,max=100"`
	Slug        string `form:"slug" validate:"required,min=2,max=100,alphanumdash"`
	Category    string `form:"category" validate:"omitempty,max=50"`
	Description string `form:"description" validate:"omitempty,max=500"`
}

type DesignationResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Category    string    `json:"category"`
	Description string    `json:"description,omitempty"`
}

type TeacherRequest struct {
	FirstName      string    `form:"first_name" validate:"required,min=2,max=100"`
	LastName       string    `form:"last_name" validate:"required,min=2,max=100"`
	Gender         string    `form:"gender" validate:"required,oneof=male female other"`
	DateOfBirth    time.Time `form:"date_of_birth" validate:"omitempty"`
	BloodGroup     string    `form:"blood_group" validate:"omitempty,max=10"`
	Religion       string    `form:"religion" validate:"omitempty,max=50"`
	Nationality    string    `form:"nationality" validate:"omitempty,max=50"`
	Phone          string    `form:"phone" validate:"omitempty,max=30"`
	Email          string    `form:"email" validate:"omitempty,email"`
	Address        string    `form:"address" validate:"omitempty,max=1000"`
	NationalID     string    `form:"national_id" validate:"omitempty,max=50"`
	JoiningDate    time.Time `form:"joining_date" validate:"required"`
	DepartmentID   uuid.UUID `form:"department_id" validate:"omitempty"`
	DesignationID  uuid.UUID `form:"designation_id" validate:"omitempty"`
	Qualification  string    `form:"qualification" validate:"omitempty,max=500"`
	Experience     string    `form:"experience" validate:"omitempty,max=1000"`
	Salary         float64   `form:"salary" validate:"omitempty,min=0"`
	EmploymentType string    `form:"employment_type" validate:"required,oneof=full_time part_time contract"`
	Status         string    `form:"status" validate:"omitempty,oneof=active inactive resigned"`
}

type TeacherAssignmentRequest struct {
	SubjectIDs []uuid.UUID `form:"subject_ids"`
	ClassIDs   []uuid.UUID `form:"class_ids"`
	SectionIDs []uuid.UUID `form:"section_ids"`
}

type TeacherResponse struct {
	ID               uuid.UUID                  `json:"id"`
	EmployeeID       string                     `json:"employee_id"`
	FirstName        string                     `json:"first_name"`
	LastName         string                     `json:"last_name"`
	FullName         string                     `json:"full_name"`
	PhotoURL         string                     `json:"photo_url,omitempty"`
	Gender           string                     `json:"gender"`
	DateOfBirth      *time.Time                 `json:"date_of_birth,omitempty"`
	BloodGroup       string                     `json:"blood_group,omitempty"`
	Religion         string                     `json:"religion,omitempty"`
	Nationality      string                     `json:"nationality,omitempty"`
	Phone            string                     `json:"phone,omitempty"`
	Email            string                     `json:"email,omitempty"`
	Address          string                     `json:"address,omitempty"`
	NationalID       string                     `json:"national_id,omitempty"`
	JoiningDate      time.Time                  `json:"joining_date"`
	DepartmentID     *uuid.UUID                 `json:"department_id,omitempty"`
	DepartmentName   string                     `json:"department_name,omitempty"`
	DesignationID    *uuid.UUID                 `json:"designation_id,omitempty"`
	DesignationName  string                     `json:"designation_name,omitempty"`
	Qualification    string                     `json:"qualification,omitempty"`
	Experience       string                     `json:"experience,omitempty"`
	Salary           float64                    `json:"salary,omitempty"`
	EmploymentType   string                     `json:"employment_type"`
	Status           string                     `json:"status"`
	Assignments      []TeacherAssignmentResponse `json:"assignments,omitempty"`
	Documents        []EmployeeDocumentResponse  `json:"documents,omitempty"`
	CreatedAt        time.Time                  `json:"created_at"`
	UpdatedAt        time.Time                  `json:"updated_at"`
}

type TeacherAssignmentResponse struct {
	ID          uuid.UUID `json:"id"`
	SubjectID   *uuid.UUID `json:"subject_id,omitempty"`
	SubjectName string    `json:"subject_name,omitempty"`
	ClassID     *uuid.UUID `json:"class_id,omitempty"`
	ClassName   string    `json:"class_name,omitempty"`
	SectionID   *uuid.UUID `json:"section_id,omitempty"`
	SectionName string    `json:"section_name,omitempty"`
}

type TeacherScheduleResponse struct {
	ID          uuid.UUID `json:"id"`
	SubjectName string    `json:"subject_name,omitempty"`
	ClassName   string    `json:"class_name,omitempty"`
	SectionName string    `json:"section_name,omitempty"`
	StartTime   string    `json:"start_time"`
	EndTime     string    `json:"end_time"`
	Room        string    `json:"room,omitempty"`
}

type StaffRequest struct {
	FirstName     string    `form:"first_name" validate:"required,min=2,max=100"`
	LastName      string    `form:"last_name" validate:"required,min=2,max=100"`
	Phone         string    `form:"phone" validate:"omitempty,max=30"`
	Email         string    `form:"email" validate:"omitempty,email"`
	Address       string    `form:"address" validate:"omitempty,max=1000"`
	DepartmentID  uuid.UUID `form:"department_id" validate:"omitempty"`
	DesignationID uuid.UUID `form:"designation_id" validate:"omitempty"`
	Salary        float64   `form:"salary" validate:"omitempty,min=0"`
	JoiningDate   time.Time `form:"joining_date" validate:"required"`
	Status        string    `form:"status" validate:"omitempty,oneof=active inactive resigned"`
}

type StaffResponse struct {
	ID              uuid.UUID                  `json:"id"`
	EmployeeID      string                     `json:"employee_id"`
	FirstName       string                     `json:"first_name"`
	LastName        string                     `json:"last_name"`
	FullName        string                     `json:"full_name"`
	PhotoURL        string                     `json:"photo_url,omitempty"`
	Phone           string                     `json:"phone,omitempty"`
	Email           string                     `json:"email,omitempty"`
	Address         string                     `json:"address,omitempty"`
	DepartmentID    *uuid.UUID                 `json:"department_id,omitempty"`
	DepartmentName  string                     `json:"department_name,omitempty"`
	DesignationID   *uuid.UUID                 `json:"designation_id,omitempty"`
	DesignationName string                     `json:"designation_name,omitempty"`
	Salary          float64                    `json:"salary,omitempty"`
	JoiningDate     time.Time                  `json:"joining_date"`
	Status          string                     `json:"status"`
	Documents       []EmployeeDocumentResponse `json:"documents,omitempty"`
	CreatedAt       time.Time                  `json:"created_at"`
	UpdatedAt       time.Time                  `json:"updated_at"`
}

type EmployeeDocumentResponse struct {
	ID       uuid.UUID `json:"id"`
	DocType  string    `json:"doc_type"`
	FileName string    `json:"file_name"`
	FileURL  string    `json:"file_url"`
}

type TeacherSearchFilter struct {
	Query        string
	DepartmentID *uuid.UUID
	DesignationID *uuid.UUID
	Status       string
	Page         int
	PageSize     int
}

type StaffSearchFilter struct {
	Query        string
	DepartmentID *uuid.UUID
	Status       string
	Page         int
	PageSize     int
}

type PaginatedTeachers struct {
	Items      []TeacherResponse `json:"items"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int               `json:"total_pages"`
}

type PaginatedStaff struct {
	Items      []StaffResponse `json:"items"`
	Total      int64           `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int               `json:"total_pages"`
}

type TeacherPortalDashboard struct {
	Teacher           TeacherResponse
	AssignedClasses   int
	AssignedSubjects  int
	TodaySchedule     []TeacherScheduleResponse
}

type HRReportFilter struct {
	DepartmentID  *uuid.UUID
	DesignationID *uuid.UUID
	Status        string
}
