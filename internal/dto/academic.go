package dto

import (
	"time"

	"github.com/google/uuid"
)

type ClassRequest struct {
	Name        string `form:"name" validate:"required,min=1,max=100"`
	Code        string `form:"code" validate:"required,min=1,max=20"`
	Description string `form:"description" validate:"omitempty,max=500"`
	SortOrder   int    `form:"sort_order" validate:"omitempty,min=0"`
}

type SectionRequest struct {
	ClassID  uuid.UUID `form:"class_id" validate:"required"`
	Name     string    `form:"name" validate:"required,min=1,max=50"`
	Capacity int       `form:"capacity" validate:"omitempty,min=1"`
}

type SubjectRequest struct {
	Name        string `form:"name" validate:"required,min=1,max=150"`
	Code        string `form:"code" validate:"required,min=1,max=30"`
	Description string `form:"description" validate:"omitempty,max=500"`
}

type AssignSubjectsRequest struct {
	SubjectIDs []uuid.UUID `form:"subject_ids"`
}

type ClassResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Description string    `json:"description,omitempty"`
	SortOrder   int       `json:"sort_order"`
	SectionCount int      `json:"section_count,omitempty"`
	StudentCount int64    `json:"student_count,omitempty"`
}

type SectionResponse struct {
	ID        uuid.UUID `json:"id"`
	ClassID   uuid.UUID `json:"class_id"`
	ClassName string    `json:"class_name,omitempty"`
	Name      string    `json:"name"`
	Capacity  int       `json:"capacity,omitempty"`
}

type SubjectResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Description string    `json:"description,omitempty"`
}

type DepartmentResponse struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Slug         string    `json:"slug"`
	Description  string    `json:"description,omitempty"`
	DeptType     string    `json:"dept_type,omitempty"`
	StaffCount   int64     `json:"staff_count,omitempty"`
	TeacherCount int64     `json:"teacher_count,omitempty"`
}

type StudentAdmissionRequest struct {
	RollNumber    string    `form:"roll_number" validate:"omitempty,max=30"`
	FirstName     string    `form:"first_name" validate:"required,min=2,max=100"`
	LastName      string    `form:"last_name" validate:"required,min=2,max=100"`
	DateOfBirth   time.Time `form:"date_of_birth" validate:"required"`
	Gender        string    `form:"gender" validate:"required,oneof=male female other"`
	BloodGroup    string    `form:"blood_group" validate:"omitempty,max=10"`
	Religion      string    `form:"religion" validate:"omitempty,max=50"`
	Nationality   string    `form:"nationality" validate:"omitempty,max=50"`
	Phone         string    `form:"phone" validate:"omitempty,max=30"`
	Email         string    `form:"email" validate:"omitempty,email"`
	Address       string    `form:"address" validate:"omitempty,max=1000"`
	SessionID     uuid.UUID `form:"session_id" validate:"required"`
	ClassID       uuid.UUID `form:"class_id" validate:"required"`
	SectionID     uuid.UUID `form:"section_id" validate:"required"`
	DepartmentID  uuid.UUID `form:"department_id" validate:"omitempty"`
	AdmissionDate time.Time `form:"admission_date" validate:"required"`
	Status        string    `form:"status" validate:"omitempty,oneof=active inactive graduated transferred"`
	FatherName       string `form:"father_name" validate:"omitempty,max=150"`
	FatherPhone      string `form:"father_phone" validate:"omitempty,max=30"`
	FatherOccupation string `form:"father_occupation" validate:"omitempty,max=100"`
	MotherName       string `form:"mother_name" validate:"omitempty,max=150"`
	MotherPhone      string `form:"mother_phone" validate:"omitempty,max=30"`
	MotherOccupation string `form:"mother_occupation" validate:"omitempty,max=100"`
	GuardianName     string `form:"guardian_name" validate:"omitempty,max=150"`
	GuardianPhone    string `form:"guardian_phone" validate:"omitempty,max=30"`
}

type StudentSearchFilter struct {
	AdmissionNumber string
	RollNumber      string
	Name            string
	ClassID         *uuid.UUID
	SectionID       *uuid.UUID
	SessionID       *uuid.UUID
	Page            int
	PageSize        int
}

type PromoteStudentRequest struct {
	ToSessionID uuid.UUID `form:"to_session_id" validate:"required"`
	ToClassID   uuid.UUID `form:"to_class_id" validate:"required"`
	ToSectionID uuid.UUID `form:"to_section_id" validate:"required"`
	Notes       string    `form:"notes" validate:"omitempty,max=500"`
}

type TransferStudentRequest struct {
	ToSessionID uuid.UUID `form:"to_session_id" validate:"required"`
	ToClassID   uuid.UUID `form:"to_class_id" validate:"required"`
	ToSectionID uuid.UUID `form:"to_section_id" validate:"required"`
	Notes       string    `form:"notes" validate:"omitempty,max=500"`
}

type StudentResponse struct {
	ID               uuid.UUID  `json:"id"`
	AdmissionNumber  string     `json:"admission_number"`
	RollNumber       string     `json:"roll_number,omitempty"`
	FirstName        string     `json:"first_name"`
	LastName         string     `json:"last_name"`
	FullName         string     `json:"full_name"`
	DateOfBirth      time.Time  `json:"date_of_birth"`
	Gender           string     `json:"gender"`
	BloodGroup       string     `json:"blood_group,omitempty"`
	Religion         string     `json:"religion,omitempty"`
	Nationality      string     `json:"nationality,omitempty"`
	PhotoURL         string     `json:"photo_url,omitempty"`
	Phone            string     `json:"phone,omitempty"`
	Email            string     `json:"email,omitempty"`
	Address          string     `json:"address,omitempty"`
	SessionID        uuid.UUID  `json:"session_id"`
	SessionName      string     `json:"session_name"`
	ClassID          uuid.UUID  `json:"class_id"`
	ClassName        string     `json:"class_name"`
	SectionID        uuid.UUID  `json:"section_id"`
	SectionName      string     `json:"section_name"`
	DepartmentID     *uuid.UUID `json:"department_id,omitempty"`
	DepartmentName   string     `json:"department_name,omitempty"`
	AdmissionDate    time.Time  `json:"admission_date"`
	Status           string     `json:"status"`
	Parents          *StudentParentResponse   `json:"parents,omitempty"`
	Documents        []StudentDocumentResponse `json:"documents,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type StudentParentResponse struct {
	FatherName       string `json:"father_name,omitempty"`
	FatherPhone      string `json:"father_phone,omitempty"`
	FatherOccupation string `json:"father_occupation,omitempty"`
	MotherName       string `json:"mother_name,omitempty"`
	MotherPhone      string `json:"mother_phone,omitempty"`
	MotherOccupation string `json:"mother_occupation,omitempty"`
	GuardianName     string `json:"guardian_name,omitempty"`
	GuardianPhone    string `json:"guardian_phone,omitempty"`
}

type StudentDocumentResponse struct {
	ID       uuid.UUID `json:"id"`
	DocType  string    `json:"doc_type"`
	FileName string    `json:"file_name"`
	FileURL  string    `json:"file_url"`
}

type PaginatedStudents struct {
	Items      []StudentResponse `json:"items"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int               `json:"total_pages"`
}

type StudentIDCardData struct {
	Student      StudentResponse
	SchoolName   string
	SchoolLogo   string
	IssueDate    time.Time
}

type ReportFilter struct {
	ClassID   *uuid.UUID
	SessionID *uuid.UUID
	Status    string
	FromDate  time.Time
	ToDate    time.Time
}
