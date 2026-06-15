package dto

import (
	"time"

	"github.com/google/uuid"
)

type ExamRequest struct {
	Name             string    `form:"name" validate:"required,min=2,max=150"`
	ExamType         string    `form:"exam_type" validate:"required,oneof=first_terminal mid_term half_yearly annual test_exam model_test"`
	SessionID        uuid.UUID `form:"session_id" validate:"required"`
	ClassID          uuid.UUID `form:"class_id" validate:"required"`
	StartDate        time.Time `form:"start_date" validate:"required"`
	EndDate          time.Time `form:"end_date" validate:"required"`
	TotalMarks       float64   `form:"total_marks" validate:"required,gte=0"`
	PassingMarks     float64   `form:"passing_marks" validate:"required,gte=0"`
	GradingSystemID  uuid.UUID `form:"grading_system_id" validate:"omitempty"`
}

type ExamResponse struct {
	ID              uuid.UUID  `json:"id"`
	Name            string     `json:"name"`
	ExamType        string     `json:"exam_type"`
	SessionID       uuid.UUID  `json:"session_id"`
	SessionName     string     `json:"session_name"`
	ClassID         uuid.UUID  `json:"class_id"`
	ClassName       string     `json:"class_name"`
	StartDate       time.Time  `json:"start_date"`
	EndDate         time.Time  `json:"end_date"`
	TotalMarks      float64    `json:"total_marks"`
	PassingMarks    float64    `json:"passing_marks"`
	GradingSystemID *uuid.UUID `json:"grading_system_id,omitempty"`
	Status          string     `json:"status"`
	ResultStatus    string     `json:"result_status"`
	SubjectCount    int64      `json:"subject_count,omitempty"`
}

type ExamSubjectRequest struct {
	SubjectID      uuid.UUID `form:"subject_id" validate:"required"`
	FullMarks      float64   `form:"full_marks" validate:"required,gte=0"`
	PassMarks      float64   `form:"pass_marks" validate:"required,gte=0"`
	WrittenMarks   float64   `form:"written_marks" validate:"omitempty,gte=0"`
	MCQMarks       float64   `form:"mcq_marks" validate:"omitempty,gte=0"`
	PracticalMarks float64   `form:"practical_marks" validate:"omitempty,gte=0"`
}

type ExamSubjectResponse struct {
	ID             uuid.UUID `json:"id"`
	ExamID         uuid.UUID `json:"exam_id"`
	SubjectID      uuid.UUID `json:"subject_id"`
	SubjectName    string    `json:"subject_name"`
	SubjectCode    string    `json:"subject_code"`
	FullMarks      float64   `json:"full_marks"`
	PassMarks      float64   `json:"pass_marks"`
	WrittenMarks   float64   `json:"written_marks"`
	MCQMarks       float64   `json:"mcq_marks"`
	PracticalMarks float64   `json:"practical_marks"`
}

type GradingScaleRequest struct {
	Grade          string  `form:"grade" validate:"required,max=5"`
	MinPercentage  float64 `form:"min_percentage" validate:"required,gte=0,lte=100"`
	MaxPercentage  float64 `form:"max_percentage" validate:"required,gte=0,lte=100"`
	GPAPoint       float64 `form:"gpa_point" validate:"required,gte=0,lte=5"`
	SortOrder      int     `form:"sort_order" validate:"omitempty"`
}

type GradingSystemRequest struct {
	Name    string `form:"name" validate:"required,min=2,max=100"`
	Scales  []GradingScaleRequest
}

type GradingScaleResponse struct {
	Grade         string  `json:"grade"`
	MinPercentage float64 `json:"min_percentage"`
	MaxPercentage float64 `json:"max_percentage"`
	GPAPoint      float64 `json:"gpa_point"`
}

type GradingSystemResponse struct {
	ID        uuid.UUID              `json:"id"`
	Name      string                 `json:"name"`
	IsDefault bool                   `json:"is_default"`
	Scales    []GradingScaleResponse `json:"scales"`
}

type MarkEntryRow struct {
	StudentID      uuid.UUID
	StudentName    string
	RollNumber     string
	AdmissionNo    string
	WrittenScore   float64
	MCQScore       float64
	PracticalScore float64
	TotalScore     float64
	IsAbsent       bool
	RecordID       *uuid.UUID
}

type StudentMarkEntry struct {
	StudentID       uuid.UUID `form:"student_id" validate:"required"`
	WrittenScore    float64   `form:"written_score" validate:"omitempty,gte=0"`
	MCQScore        float64   `form:"mcq_score" validate:"omitempty,gte=0"`
	PracticalScore  float64   `form:"practical_score" validate:"omitempty,gte=0"`
	IsAbsent        bool      `form:"is_absent"`
}

type SubjectMarkDetail struct {
	SubjectName    string
	SubjectCode    string
	FullMarks      float64
	PassMarks      float64
	WrittenScore   float64
	MCQScore       float64
	PracticalScore float64
	TotalScore     float64
	IsPassed       bool
	Grade          string
}

type ExamResultResponse struct {
	ID              uuid.UUID           `json:"id"`
	ExamID          uuid.UUID           `json:"exam_id"`
	ExamName        string              `json:"exam_name"`
	StudentID       uuid.UUID           `json:"student_id"`
	StudentName     string              `json:"student_name"`
	AdmissionNo     string              `json:"admission_number"`
	RollNumber      string              `json:"roll_number"`
	ClassName       string              `json:"class_name"`
	SectionName     string              `json:"section_name"`
	TotalObtained   float64             `json:"total_obtained"`
	TotalFull       float64             `json:"total_full"`
	Percentage      float64             `json:"percentage"`
	GPA             float64             `json:"gpa"`
	CGPA            float64             `json:"cgpa"`
	Grade           string              `json:"grade"`
	IsPassed        bool                `json:"is_passed"`
	ClassPosition   *int                `json:"class_position,omitempty"`
	SectionPosition *int                `json:"section_position,omitempty"`
	MeritPosition   *int                `json:"merit_position,omitempty"`
	ResultStatus    string              `json:"result_status"`
	Subjects        []SubjectMarkDetail `json:"subjects,omitempty"`
	AttendancePct   float64             `json:"attendance_pct,omitempty"`
}

type ReportCardData struct {
	SchoolName    string
	StudentName   string
	AdmissionNo   string
	RollNumber    string
	ClassName     string
	SectionName   string
	ExamName      string
	SessionName   string
	Result        *ExamResultResponse
	AttendancePct float64
	CardToken     string
}

type ExamDashboardStats struct {
	ActiveExams       int64
	PublishedResults  int64
	StudentsPassed    int64
	StudentsFailed    int64
	GPADistribution   []GPADistPoint
	PassRate          float64
	SubjectPerformance []SubjectPerfPoint
}

type GPADistPoint struct {
	Grade string
	Count int64
}

type SubjectPerfPoint struct {
	SubjectName string
	AvgScore    float64
	PassRate    float64
}

type ExamSearchFilter struct {
	SessionID uuid.UUID
	ClassID   uuid.UUID
	Status    string
	Query     string
	Page      int
	PerPage   int
}

type PaginatedExams struct {
	Items      []ExamResponse
	Total      int64
	Page       int
	PerPage    int
	TotalPages int
}

type PaginatedResults struct {
	Items      []ExamResultResponse
	Total      int64
	Page       int
	PerPage    int
	TotalPages int
}

type TabulationRow struct {
	StudentID   uuid.UUID
	StudentName string
	RollNumber  string
	Subjects    map[string]float64
	Total       float64
	GPA         float64
	Grade       string
	Position    int
}

type ExamReportFilter struct {
	ExamID    uuid.UUID
	ClassID   uuid.UUID
	SectionID uuid.UUID
	StudentID uuid.UUID
	PassedOnly bool
	FailedOnly bool
	PublishedOnly bool
	TopN      int
}
