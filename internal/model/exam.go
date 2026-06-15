package model

const (
	EntityExam         = "exam"
	EntityExamSubject  = "exam_subject"
	EntityStudentMark  = "student_mark"
	EntityExamResult   = "exam_result"
	EntityReportCard   = "report_card"
	EntityGradingSystem = "grading_system"
)

const (
	PermExamCreate   = "exam.create"
	PermExamUpdate   = "exam.update"
	PermExamDelete   = "exam.delete"
	PermExamPublish  = "exam.publish"
	PermMarksEntry   = "marks.entry"
	PermMarksUpdate  = "marks.update"
	PermResultProcess = "result.process"
	PermResultPublish = "result.publish"
)

const (
	ExamStatusDraft     = "draft"
	ExamStatusActive    = "active"
	ExamStatusPublished = "published"
	ExamStatusArchived  = "archived"
)

const (
	ResultStatusDraft     = "draft"
	ResultStatusPublished = "published"
)

const (
	ExamTypeFirstTerminal = "first_terminal"
	ExamTypeMidTerm       = "mid_term"
	ExamTypeHalfYearly    = "half_yearly"
	ExamTypeAnnual        = "annual"
	ExamTypeTest          = "test_exam"
	ExamTypeModelTest     = "model_test"
)
