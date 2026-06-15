package dto

import (
	"time"

	"github.com/google/uuid"
)

type AttendanceFilter struct {
	SessionID uuid.UUID
	ClassID   uuid.UUID
	SectionID uuid.UUID
	Date      time.Time
	Query     string
	Status    string
	Page      int
	PerPage   int
}

type StudentAttendanceEntry struct {
	StudentID uuid.UUID `form:"student_id" validate:"required"`
	Status    string    `form:"status" validate:"required,oneof=present absent late leave"`
	Remarks   string    `form:"remarks" validate:"omitempty,max=500"`
}

type BulkStudentAttendanceRequest struct {
	SessionID uuid.UUID `form:"session_id" validate:"required"`
	ClassID   uuid.UUID `form:"class_id" validate:"required"`
	SectionID uuid.UUID `form:"section_id" validate:"required"`
	Date      time.Time `form:"date" validate:"required"`
}

type StudentAttendanceRow struct {
	StudentID       uuid.UUID
	AdmissionNumber string
	RollNumber      string
	FullName        string
	PhotoURL        string
	Status          string
	Remarks         string
	RecordID        *uuid.UUID
}

type StudentAttendanceResponse struct {
	ID              uuid.UUID `json:"id"`
	StudentID       uuid.UUID `json:"student_id"`
	StudentName     string    `json:"student_name"`
	AdmissionNumber string    `json:"admission_number"`
	RollNumber      string    `json:"roll_number"`
	SessionID       uuid.UUID `json:"session_id"`
	ClassID         uuid.UUID `json:"class_id"`
	ClassName       string    `json:"class_name"`
	SectionID       uuid.UUID `json:"section_id"`
	SectionName     string    `json:"section_name"`
	AttendanceDate  time.Time `json:"attendance_date"`
	Status          string    `json:"status"`
	Remarks         string    `json:"remarks,omitempty"`
}

type TeacherAttendanceRow struct {
	TeacherID  uuid.UUID
	EmployeeID string
	FullName   string
	PhotoURL   string
	Department string
	Status     string
	Remarks    string
	RecordID   *uuid.UUID
}

type StaffAttendanceRow struct {
	StaffID    uuid.UUID
	EmployeeID string
	Name       string
	PhotoURL   string
	Department string
	Status     string
	Remarks    string
	RecordID   *uuid.UUID
}

type AttendanceDashboardStats struct {
	StudentPresentToday int64
	StudentAbsentToday  int64
	TeacherPresentToday int64
	StaffPresentToday   int64
	MonthlyTrend        []AttendanceTrendPoint
	ClassWiseToday      []ClassAttendanceSummary
}

type AttendanceTrendPoint struct {
	Date    time.Time
	Present int64
	Absent  int64
	Late    int64
	Leave   int64
}

type ClassAttendanceSummary struct {
	ClassID   uuid.UUID
	ClassName string
	Present   int64
	Absent    int64
	Late      int64
	Leave     int64
	Total     int64
}

type LeaveApplyRequest struct {
	EntityType string    `form:"entity_type" validate:"required,oneof=teacher staff"`
	TeacherID  uuid.UUID `form:"teacher_id" validate:"omitempty"`
	StaffID    uuid.UUID `form:"staff_id" validate:"omitempty"`
	LeaveType  string    `form:"leave_type" validate:"required,oneof=casual sick annual emergency"`
	StartDate  time.Time `form:"start_date" validate:"required"`
	EndDate    time.Time `form:"end_date" validate:"required"`
	Reason     string    `form:"reason" validate:"omitempty,max=1000"`
}

type LeaveReviewRequest struct {
	ReviewRemarks string `form:"review_remarks" validate:"omitempty,max=500"`
}

type LeaveRequestResponse struct {
	ID            uuid.UUID  `json:"id"`
	EntityType    string     `json:"entity_type"`
	EmployeeName  string     `json:"employee_name"`
	EmployeeID    string     `json:"employee_id"`
	LeaveType     string     `json:"leave_type"`
	StartDate     time.Time  `json:"start_date"`
	EndDate       time.Time  `json:"end_date"`
	Reason        string     `json:"reason,omitempty"`
	Status        string     `json:"status"`
	ReviewRemarks string     `json:"review_remarks,omitempty"`
	ReviewedAt    *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

type PaginatedLeaveRequests struct {
	Items      []LeaveRequestResponse `json:"items"`
	Total      int64                `json:"total"`
	Page       int                  `json:"page"`
	PerPage    int                  `json:"per_page"`
	TotalPages int                  `json:"total_pages"`
}

type LeaveFilter struct {
	EntityType string
	Status     string
	LeaveType  string
	Query      string
	Page       int
	PerPage    int
}

type AttendanceReportFilter struct {
	SessionID uuid.UUID
	ClassID   uuid.UUID
	SectionID uuid.UUID
	StudentID uuid.UUID
	TeacherID uuid.UUID
	StaffID   uuid.UUID
	From      time.Time
	To        time.Time
	Status    string
}

type StudentAttendanceSummary struct {
	StudentID     uuid.UUID
	StudentName   string
	PresentDays   int64
	AbsentDays    int64
	LateDays      int64
	LeaveDays     int64
	TotalMarked   int64
	AttendancePct float64
}

type MonthlyAttendanceSummary struct {
	Month       string
	PresentDays int64
	AbsentDays  int64
	LateDays    int64
	LeaveDays   int64
}
