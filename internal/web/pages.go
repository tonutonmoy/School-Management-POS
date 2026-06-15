package web

import (
	"bytes"
	"context"
	"time"

	"github.com/a-h/templ"
	"github.com/google/uuid"

	"github.com/school-management/pos/internal/dto"
	"github.com/school-management/pos/web/pages"
)

type Page interface {
	Render(csrf string, user *dto.AuthUser, appName string) string
}

type LoginPage struct {
	Flash, FlashType       string
	DefaultEmail, DefaultPassword string
}

func (p LoginPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.Login(p.Flash, p.FlashType, csrf, p.DefaultEmail, p.DefaultPassword))
}

type ForgotPasswordPage struct {
	Flash, FlashType string
}

func (p ForgotPasswordPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.ForgotPassword(p.Flash, p.FlashType, csrf))
}

type ResetPasswordPage struct {
	Token, Flash, FlashType string
}

func (p ResetPasswordPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.ResetPassword(p.Token, p.Flash, p.FlashType, csrf))
}

type ChangePasswordPage struct {
	Flash, FlashType string
}

func (p ChangePasswordPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.ChangePassword(p.Flash, p.FlashType, csrf, user))
}

type DashboardPage struct {
	Stats            *dto.DashboardStats
	Flash, FlashType string
}

func (p DashboardPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.Dashboard(p.Stats, p.Flash, p.FlashType, csrf, user))
}

type UserListPage struct {
	Data *dto.PaginatedUsers
}

func (p UserListPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.UserList(p.Data, csrf, user))
}

type UserDetailPage struct {
	User dto.UserResponse
}

func (p UserDetailPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.UserDetail(p.User, csrf, user))
}

type UserFormPage struct {
	Title string
	User  *dto.UserResponse
	Roles []dto.RoleResponse
}

func (p UserFormPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.UserForm(p.Title, p.User, p.Roles, csrf, user))
}

type RoleListPage struct {
	Roles []dto.RoleResponse
}

func (p RoleListPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.RoleList(p.Roles, csrf, user))
}

type RoleFormPage struct {
	Title       string
	Role        *dto.RoleResponse
	Permissions []dto.PermissionResponse
}

func (p RoleFormPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.RoleForm(p.Title, p.Role, p.Permissions, csrf, user))
}

type SchoolPage struct {
	School           *dto.SchoolResponse
	Flash, FlashType string
}

func (p SchoolPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.SchoolSetup(p.School, p.Flash, p.FlashType, csrf, user))
}

type SessionListPage struct {
	Sessions []dto.AcademicSessionResponse
}

func (p SessionListPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.SessionList(p.Sessions, csrf, user))
}

type SessionFormPage struct {
	Title   string
	Session *dto.AcademicSessionResponse
}

func (p SessionFormPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.SessionForm(p.Title, p.Session, csrf, user))
}

type AuditListPage struct {
	Logs []dto.ActivityItem
}

func (p AuditListPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.AuditList(p.Logs, csrf, user))
}

func render(c templ.Component) string {
	var buf bytes.Buffer
	_ = c.Render(context.Background(), &buf)
	return buf.String()
}

type StudentFormData struct {
	Sessions    []dto.AcademicSessionResponse
	Classes     []dto.ClassResponse
	Sections    []dto.SectionResponse
	Departments []dto.DepartmentResponse
}

func toPagesFormData(f *StudentFormData) *pages.StudentFormData {
	if f == nil {
		return &pages.StudentFormData{}
	}
	return &pages.StudentFormData{
		Sessions: f.Sessions, Classes: f.Classes, Sections: f.Sections, Departments: f.Departments,
	}
}

type ClassStudentGroup struct {
	ClassName string
	Students  []dto.StudentResponse
}

type ClassListPage struct{ Classes []dto.ClassResponse }
func (p ClassListPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.ClassList(p.Classes, csrf, user))
}

type ClassFormPage struct {
	Title string; Class *dto.ClassResponse; Subjects, Assigned []dto.SubjectResponse
}
func (p ClassFormPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.ClassForm(p.Title, p.Class, p.Subjects, p.Assigned, csrf, user))
}

type SectionListPage struct{ Sections []dto.SectionResponse }
func (p SectionListPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.SectionList(p.Sections, csrf, user))
}

type SectionFormPage struct {
	Title string; Section *dto.SectionResponse; Classes []dto.ClassResponse
}
func (p SectionFormPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.SectionForm(p.Title, p.Section, p.Classes, csrf, user))
}

type SubjectListPage struct{ Subjects []dto.SubjectResponse }
func (p SubjectListPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.SubjectList(p.Subjects, csrf, user))
}

type SubjectFormPage struct{ Title string; Subject *dto.SubjectResponse }
func (p SubjectFormPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.SubjectForm(p.Title, p.Subject, csrf, user))
}

type StudentListPage struct {
	Data *dto.PaginatedStudents; Filter dto.StudentSearchFilter; FormData *StudentFormData
}
func (p StudentListPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.StudentList(p.Data, p.Filter, toPagesFormData(p.FormData), csrf, user))
}

type StudentFormPage struct {
	Title string; Student *dto.StudentResponse; FormData *StudentFormData
}
func (p StudentFormPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.StudentForm(p.Title, p.Student, toPagesFormData(p.FormData), csrf, user))
}

type StudentProfilePage struct{ Student *dto.StudentResponse }
func (p StudentProfilePage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.StudentProfile(p.Student, csrf, user))
}

type StudentPromotePage struct {
	Student *dto.StudentResponse; FormData *StudentFormData; Action string
}
func (p StudentPromotePage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.StudentPromote(p.Student, toPagesFormData(p.FormData), p.Action, csrf, user))
}

type StudentIDCardPage struct{ Data *dto.StudentIDCardData }
func (p StudentIDCardPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.StudentIDCard(p.Data, csrf, user))
}

type StudentReportPage struct {
	Title string; Students []dto.StudentResponse; Filter dto.ReportFilter
}
func (p StudentReportPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.StudentReport(p.Title, p.Students, p.Filter, csrf, user))
}

type ClassWiseReportPage struct {
	Groups []ClassStudentGroup; Filter dto.ReportFilter
}
func (p ClassWiseReportPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	groups := make([]pages.ClassStudentGroup, len(p.Groups))
	for i, g := range p.Groups {
		groups[i] = pages.ClassStudentGroup{ClassName: g.ClassName, Students: g.Students}
	}
	return render(pages.ClassWiseReport(groups, p.Filter, csrf, user))
}

type AdmissionReportPage struct {
	Students []dto.StudentResponse; From, To time.Time
}
func (p AdmissionReportPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.AdmissionReport(p.Students, p.From, p.To, csrf, user))
}

type HRFormData struct {
	Departments  []dto.DepartmentResponse
	Designations []dto.DesignationResponse
	Classes      []dto.ClassResponse
	Subjects     []dto.SubjectResponse
	Sections     []dto.SectionResponse
}

func toHRFormData(f *HRFormData) *pages.HRFormData {
	if f == nil {
		return &pages.HRFormData{}
	}
	return &pages.HRFormData{
		Departments: f.Departments, Designations: f.Designations,
		Classes: f.Classes, Subjects: f.Subjects, Sections: f.Sections,
	}
}

type DepartmentListPage struct{ Departments []dto.DepartmentResponse }
func (p DepartmentListPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.DepartmentList(p.Departments, csrf, user))
}

type DepartmentFormPage struct{ Title string; Department *dto.DepartmentResponse }
func (p DepartmentFormPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.DepartmentForm(p.Title, p.Department, csrf, user))
}

type DesignationListPage struct{ Designations []dto.DesignationResponse }
func (p DesignationListPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.DesignationList(p.Designations, csrf, user))
}

type DesignationFormPage struct{ Title string; Designation *dto.DesignationResponse }
func (p DesignationFormPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.DesignationForm(p.Title, p.Designation, csrf, user))
}

type TeacherListPage struct {
	Data *dto.PaginatedTeachers; Filter dto.TeacherSearchFilter; FormData *HRFormData
}
func (p TeacherListPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.TeacherList(p.Data, p.Filter, toHRFormData(p.FormData), csrf, user))
}

type TeacherFormPage struct {
	Title string; Teacher *dto.TeacherResponse; FormData *HRFormData
}
func (p TeacherFormPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.TeacherForm(p.Title, p.Teacher, toHRFormData(p.FormData), csrf, user))
}

type TeacherProfilePage struct{ Teacher *dto.TeacherResponse }
func (p TeacherProfilePage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.TeacherProfile(p.Teacher, csrf, user))
}

type TeacherAssignPage struct{ Teacher *dto.TeacherResponse; FormData *HRFormData }
func (p TeacherAssignPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.TeacherAssign(p.Teacher, toHRFormData(p.FormData), csrf, user))
}

type StaffListPage struct {
	Data *dto.PaginatedStaff; Filter dto.StaffSearchFilter; FormData *HRFormData
}
func (p StaffListPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.StaffList(p.Data, p.Filter, toHRFormData(p.FormData), csrf, user))
}

type StaffFormPage struct {
	Title string; Staff *dto.StaffResponse; FormData *HRFormData
}
func (p StaffFormPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.StaffForm(p.Title, p.Staff, toHRFormData(p.FormData), csrf, user))
}

type StaffProfilePage struct{ Staff *dto.StaffResponse }
func (p StaffProfilePage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.StaffProfile(p.Staff, csrf, user))
}

type TeacherPortalPage struct{ Dashboard *dto.TeacherPortalDashboard }
func (p TeacherPortalPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.TeacherPortal(p.Dashboard, csrf, user))
}

type HRTeacherReportPage struct {
	Teachers []dto.TeacherResponse; Filter dto.HRReportFilter
}
func (p HRTeacherReportPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.HRTeacherReport(p.Teachers, p.Filter, csrf, user))
}

type HRStaffReportPage struct {
	Staff []dto.StaffResponse; Filter dto.HRReportFilter
}
func (p HRStaffReportPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.HRStaffReport(p.Staff, p.Filter, csrf, user))
}

type DepartmentReportGroup struct {
	Department dto.DepartmentResponse
	Teachers   []dto.TeacherResponse
	Staff      []dto.StaffResponse
}

type DepartmentReportPage struct{ Groups []DepartmentReportGroup }
func (p DepartmentReportPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	groups := make([]pages.DepartmentReportGroup, len(p.Groups))
	for i, g := range p.Groups {
		groups[i] = pages.DepartmentReportGroup{Department: g.Department, Teachers: g.Teachers, Staff: g.Staff}
	}
	return render(pages.DepartmentReport(groups, csrf, user))
}

type AssignmentReportPage struct{ Teachers []dto.TeacherResponse }
func (p AssignmentReportPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.AssignmentReport(p.Teachers, csrf, user))
}

type AttendanceFormData struct {
	Sessions []dto.AcademicSessionResponse
	Classes  []dto.ClassResponse
	Sections []dto.SectionResponse
	Teachers []dto.TeacherResponse
	Staff    []dto.StaffResponse
}

func toAttendanceFormData(f *AttendanceFormData) *pages.AttendanceFormData {
	if f == nil {
		return &pages.AttendanceFormData{}
	}
	return &pages.AttendanceFormData{
		Sessions: f.Sessions, Classes: f.Classes, Sections: f.Sections, Teachers: f.Teachers, Staff: f.Staff,
	}
}

type StudentAttendancePage struct {
	Rows []dto.StudentAttendanceRow; Filter dto.AttendanceFilter; FormData *AttendanceFormData
	Flash, FlashType string
}
func (p StudentAttendancePage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.StudentAttendance(p.Rows, p.Filter, toAttendanceFormData(p.FormData), p.Flash, p.FlashType, csrf, user))
}

type TeacherAttendancePage struct {
	Rows []dto.TeacherAttendanceRow; Date time.Time; Query string
	Flash, FlashType string
}
func (p TeacherAttendancePage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.TeacherAttendance(p.Rows, p.Date, p.Query, p.Flash, p.FlashType, csrf, user))
}

type StaffAttendancePage struct {
	Rows []dto.StaffAttendanceRow; Date time.Time; Query string
	Flash, FlashType string
}
func (p StaffAttendancePage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.StaffAttendance(p.Rows, p.Date, p.Query, p.Flash, p.FlashType, csrf, user))
}

type AttendanceDashboardPage struct{ Stats *dto.AttendanceDashboardStats }
func (p AttendanceDashboardPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.AttendanceDashboard(p.Stats, csrf, user))
}

type LeaveListPage struct{ Data *dto.PaginatedLeaveRequests; Filter dto.LeaveFilter }
func (p LeaveListPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.LeaveList(p.Data, p.Filter, csrf, user))
}

type LeaveApplyPage struct{ FormData *AttendanceFormData }
func (p LeaveApplyPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.LeaveApply(toAttendanceFormData(p.FormData), csrf, user))
}

type ParentAttendancePage struct{ Summary *dto.StudentAttendanceSummary }
func (p ParentAttendancePage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.ParentAttendance(p.Summary, csrf, user))
}

type StudentAttendanceReportPage struct {
	Title string; Records []dto.StudentAttendanceResponse; Filter dto.AttendanceReportFilter; FormData *AttendanceFormData
}
func (p StudentAttendanceReportPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.StudentAttendanceReport(p.Title, p.Records, p.Filter, toAttendanceFormData(p.FormData), csrf, user))
}

type ClassWiseAttendanceReportPage struct{ Groups []dto.ClassAttendanceSummary; Date time.Time }
func (p ClassWiseAttendanceReportPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.ClassWiseAttendanceReport(p.Groups, p.Date, csrf, user))
}

type StudentHistoryReportPage struct {
	Records []dto.StudentAttendanceResponse; Summary *dto.StudentAttendanceSummary
	Total int64; Page int; StudentID uuid.UUID; From, To time.Time
}
func (p StudentHistoryReportPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.StudentHistoryReport(p.Records, p.Summary, p.Total, p.Page, p.StudentID, p.From, p.To, csrf, user))
}

type EmployeeAttendanceReportPage struct {
	Title string; Records []dto.StudentAttendanceResponse; Filter dto.AttendanceReportFilter; Entity string
}
func (p EmployeeAttendanceReportPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.EmployeeAttendanceReport(p.Title, p.Records, p.Filter, p.Entity, csrf, user))
}

type FeeFormData struct {
	Sessions []dto.AcademicSessionResponse
	Classes  []dto.ClassResponse
	Sections []dto.SectionResponse
	FeeTypes []dto.FeeTypeResponse
}

func toFeeFormData(f *FeeFormData) *pages.FeeFormData {
	if f == nil { return &pages.FeeFormData{} }
	return &pages.FeeFormData{Sessions: f.Sessions, Classes: f.Classes, Sections: f.Sections, FeeTypes: f.FeeTypes}
}

type FeeTypeListPage struct{ Types []dto.FeeTypeResponse }
func (p FeeTypeListPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.FeeTypeList(p.Types, csrf, user))
}

type FeeTypeFormPage struct{ Title string; FeeType *dto.FeeTypeResponse }
func (p FeeTypeFormPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.FeeTypeForm(p.Title, p.FeeType, csrf, user))
}

type FeeStructureListPage struct{ Structures []dto.FeeStructureResponse; FormData *FeeFormData }
func (p FeeStructureListPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.FeeStructureList(p.Structures, toFeeFormData(p.FormData), csrf, user))
}

type FeeStructureFormPage struct{ Title string; Structure *dto.FeeStructureResponse; FormData *FeeFormData }
func (p FeeStructureFormPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.FeeStructureForm(p.Title, p.Structure, toFeeFormData(p.FormData), csrf, user))
}

type BillListPage struct{ Data *dto.PaginatedBills; Filter dto.BillSearchFilter; FormData *FeeFormData }
func (p BillListPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.BillList(p.Data, p.Filter, toFeeFormData(p.FormData), csrf, user))
}

type BillGeneratePage struct{ FormData *FeeFormData }
func (p BillGeneratePage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.BillGenerate(toFeeFormData(p.FormData), csrf, user))
}

type BillDetailPage struct{ Bill *dto.StudentBillResponse }
func (p BillDetailPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.BillDetail(p.Bill, csrf, user))
}

type FeeCollectPage struct{ Flash, FlashType string }
func (p FeeCollectPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.FeeCollect(p.Flash, p.FlashType, csrf, user))
}

type FinanceDashboardPage struct{ Stats *dto.FinanceDashboardStats }
func (p FinanceDashboardPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.FinanceDashboard(p.Stats, csrf, user))
}

type DueListPage struct{ Rows []dto.DueStudentRow; FormData *FeeFormData }
func (p DueListPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.DueList(p.Rows, toFeeFormData(p.FormData), csrf, user))
}

type OverdueListPage struct{ Data *dto.PaginatedBills }
func (p OverdueListPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.OverdueList(p.Data, csrf, user))
}

type ReceiptPage struct{ Receipt *dto.ReceiptResponse }
func (p ReceiptPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.ReceiptPage(p.Receipt, csrf, user))
}

type ReceiptVerifyPage struct{ Valid bool; Receipt *dto.ReceiptResponse }
func (p ReceiptVerifyPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.ReceiptVerify(p.Valid, p.Receipt, csrf, user))
}

type ParentFeePage struct{ Summary *dto.ParentFeeSummary }
func (p ParentFeePage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.ParentFee(p.Summary, csrf, user))
}

type CollectionReportPage struct{ Title string; Payments []dto.PaymentResponse; Filter dto.FinanceReportFilter }
func (p CollectionReportPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.CollectionReport(p.Title, p.Payments, p.Filter, csrf, user))
}

type DiscountListPage struct{ Discounts []dto.StudentDiscountResponse }
func (p DiscountListPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.DiscountList(p.Discounts, csrf, user))
}

type DiscountFormPage struct{ FormData *FeeFormData }
func (p DiscountFormPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.DiscountForm(toFeeFormData(p.FormData), csrf, user))
}

type StudentLedgerPage struct{ Entries []dto.StudentLedgerEntry; StudentID uuid.UUID }
func (p StudentLedgerPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.StudentLedger(p.Entries, p.StudentID, csrf, user))
}

type PaymentHistoryPage struct{ Payments []dto.PaymentResponse; StudentID uuid.UUID }
func (p PaymentHistoryPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.PaymentHistory(p.Payments, p.StudentID, csrf, user))
}

type DueStatementPage struct{ Data *dto.PaginatedBills; StudentID uuid.UUID }
func (p DueStatementPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.DueStatement(p.Data, p.StudentID, csrf, user))
}

type IncomeSummaryPage struct{ Stats *dto.FinanceDashboardStats }
func (p IncomeSummaryPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.IncomeSummary(p.Stats, csrf, user))
}

type MethodReportPage struct{ Stats *dto.FinanceDashboardStats; Filter dto.FinanceReportFilter }
func (p MethodReportPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.MethodReport(p.Stats, p.Filter, csrf, user))
}

type FeeTypeReportPage struct{ Items []dto.FinanceTrendPoint; Filter dto.FinanceReportFilter }
func (p FeeTypeReportPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.FeeTypeReport(p.Items, p.Filter, csrf, user))
}

type ExamFormData struct {
	Sessions       []dto.AcademicSessionResponse
	Classes        []dto.ClassResponse
	Sections       []dto.SectionResponse
	Subjects       []dto.SubjectResponse
	GradingSystems []dto.GradingSystemResponse
	Exams          []dto.ExamResponse
}

func toExamFormData(f *ExamFormData) *pages.ExamFormData {
	if f == nil {
		return &pages.ExamFormData{}
	}
	return &pages.ExamFormData{
		Sessions: f.Sessions, Classes: f.Classes, Sections: f.Sections,
		Subjects: f.Subjects, GradingSystems: f.GradingSystems, Exams: f.Exams,
	}
}

type ExamListPage struct {
	Data *dto.PaginatedExams; Filter dto.ExamSearchFilter; FormData *ExamFormData
	Flash, FlashType string
}
func (p ExamListPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.ExamList(p.Data, p.Filter, toExamFormData(p.FormData), p.Flash, p.FlashType, csrf, user))
}

type ExamFormPage struct {
	Title string; Exam *dto.ExamResponse; FormData *ExamFormData
	Flash, FlashType string
}
func (p ExamFormPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.ExamForm(p.Title, p.Exam, toExamFormData(p.FormData), p.Flash, p.FlashType, csrf, user))
}

type ExamDetailPage struct {
	Exam *dto.ExamResponse; Subjects []dto.ExamSubjectResponse; FormData *ExamFormData
	Flash, FlashType string
}
func (p ExamDetailPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.ExamDetail(p.Exam, p.Subjects, toExamFormData(p.FormData), p.Flash, p.FlashType, csrf, user))
}

type MarksEntryPage struct {
	Exam *dto.ExamResponse; Subjects []dto.ExamSubjectResponse
	ActiveSubject *dto.ExamSubjectResponse; Rows []dto.MarkEntryRow
	Flash, FlashType string
}
func (p MarksEntryPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.MarksEntry(p.Exam, p.Subjects, p.ActiveSubject, p.Rows, p.Flash, p.FlashType, csrf, user))
}

type ExamDashboardPage struct {
	Stats *dto.ExamDashboardStats; ExamStats *dto.ExamDashboardStats
	Exams *dto.PaginatedExams; FormData *ExamFormData
}
func (p ExamDashboardPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.ExamDashboard(p.Stats, p.ExamStats, p.Exams, toExamFormData(p.FormData), csrf, user))
}

type ExamResultsPage struct {
	Exam *dto.ExamResponse; Data *dto.PaginatedResults; Filter dto.ExamReportFilter; FormData *ExamFormData
}
func (p ExamResultsPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.ExamResults(p.Exam, p.Data, p.Filter, toExamFormData(p.FormData), csrf, user))
}

type ExamResultDetailPage struct{ Exam *dto.ExamResponse; Result *dto.ExamResultResponse }
func (p ExamResultDetailPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.ExamResultDetail(p.Exam, p.Result, csrf, user))
}

type TabulationPage struct {
	Exam *dto.ExamResponse; Results []dto.ExamResultResponse
	SectionID *uuid.UUID; FormData *ExamFormData
}
func (p TabulationPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.Tabulation(p.Exam, p.Results, p.SectionID, toExamFormData(p.FormData), csrf, user))
}

type MeritListPage struct{ Exam *dto.ExamResponse; Results []dto.ExamResultResponse }
func (p MeritListPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.MeritList(p.Exam, p.Results, csrf, user))
}

type GradingSystemPage struct {
	Systems []dto.GradingSystemResponse; Flash, FlashType string
}
func (p GradingSystemPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.GradingSystems(p.Systems, p.Flash, p.FlashType, csrf, user))
}

type ParentResultsPage struct{ Student *dto.StudentResponse; Results []dto.ExamResultResponse }
func (p ParentResultsPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.ParentResults(p.Student, p.Results, csrf, user))
}

type ParentResultPage struct{ Result *dto.ExamResultResponse; StudentID uuid.UUID }
func (p ParentResultPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.ParentResult(p.Result, p.StudentID, csrf, user))
}

type ExamReportSummaryPage struct {
	Exam *dto.ExamResponse; Stats *dto.ExamDashboardStats; FormData *ExamFormData
}
func (p ExamReportSummaryPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.ExamReportSummary(p.Exam, p.Stats, toExamFormData(p.FormData), csrf, user))
}

type ExamReportSubjectPage struct {
	Exam *dto.ExamResponse; Stats *dto.ExamDashboardStats; FormData *ExamFormData
}
func (p ExamReportSubjectPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.ExamReportSubject(p.Exam, p.Stats, toExamFormData(p.FormData), csrf, user))
}

type ExamReportTopPage struct {
	Exam *dto.ExamResponse; Results []dto.ExamResultResponse; FormData *ExamFormData
}
func (p ExamReportTopPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.ExamReportTop(p.Exam, p.Results, toExamFormData(p.FormData), csrf, user))
}

type ExamReportFailedPage struct {
	Exam *dto.ExamResponse; Results []dto.ExamResultResponse; FormData *ExamFormData
}
func (p ExamReportFailedPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.ExamReportFailed(p.Exam, p.Results, toExamFormData(p.FormData), csrf, user))
}

type AccountingFormData struct {
	Accounts   []dto.AccountResponse
	Categories []dto.ExpenseCategoryResponse
}

func toAccountingFormData(f *AccountingFormData) *pages.AccountingFormData {
	if f == nil {
		return &pages.AccountingFormData{}
	}
	return &pages.AccountingFormData{Accounts: f.Accounts, Categories: f.Categories}
}

type AccountingDashboardPage struct{ Stats *dto.AccountingDashboardStats }
func (p AccountingDashboardPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.AccountingDashboard(p.Stats, csrf, user))
}

type AccountListPage struct{ Accounts []dto.AccountResponse; FilterType string }
func (p AccountListPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.AccountList(p.Accounts, p.FilterType, csrf, user))
}

type AccountFormPage struct{ Title string; Account *dto.AccountResponse; FormData *AccountingFormData }
func (p AccountFormPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.AccountForm(p.Title, p.Account, toAccountingFormData(p.FormData), csrf, user))
}

type JournalListPage struct{ Data *dto.PaginatedJournalEntries; From, To time.Time }
func (p JournalListPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.JournalList(p.Data, p.From, p.To, csrf, user))
}

type JournalFormPage struct{ FormData *AccountingFormData }
func (p JournalFormPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.JournalForm(toAccountingFormData(p.FormData), csrf, user))
}

type JournalDetailPage struct{ Entry *dto.JournalEntryResponse }
func (p JournalDetailPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.JournalDetail(p.Entry, csrf, user))
}

type LedgerPage struct {
	Report *dto.LedgerReport; FormData *AccountingFormData
	From, To time.Time; AccountID uuid.UUID
}
func (p LedgerPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.LedgerPage(p.Report, toAccountingFormData(p.FormData), p.From, p.To, p.AccountID, csrf, user))
}

type CashBookPage struct{ Entries []dto.CashBookEntry; From, To time.Time }
func (p CashBookPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.CashBook(p.Entries, p.From, p.To, csrf, user))
}

type BankBookPage struct{ Entries []dto.BankBookEntry; From, To time.Time }
func (p BankBookPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.BankBook(p.Entries, p.From, p.To, csrf, user))
}

type ExpenseListPage struct {
	Data *dto.PaginatedExpenses; Status string; From, To time.Time
}
func (p ExpenseListPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.ExpenseList(p.Data, p.Status, p.From, p.To, csrf, user))
}

type ExpenseFormPage struct{ FormData *AccountingFormData }
func (p ExpenseFormPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.ExpenseForm(toAccountingFormData(p.FormData), csrf, user))
}

type IncomeListPage struct{ Entries []dto.IncomeEntryResponse; From, To time.Time }
func (p IncomeListPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.IncomeList(p.Entries, p.From, p.To, csrf, user))
}

type IncomeFormPage struct{ FormData *AccountingFormData }
func (p IncomeFormPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.IncomeForm(toAccountingFormData(p.FormData), csrf, user))
}

type TrialBalancePage struct{ Report *dto.TrialBalanceReport; AsOf time.Time }
func (p TrialBalancePage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.TrialBalanceReport(p.Report, p.AsOf, csrf, user))
}

type IncomeStatementPage struct{ Report *dto.IncomeStatementReport; From, To time.Time }
func (p IncomeStatementPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.IncomeStatementReport(p.Report, p.From, p.To, csrf, user))
}

type BalanceSheetPage struct{ Report *dto.BalanceSheetReport; AsOf time.Time }
func (p BalanceSheetPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.BalanceSheetReport(p.Report, p.AsOf, csrf, user))
}

type CashFlowPage struct{ Report *dto.CashFlowReport; From, To time.Time }
func (p CashFlowPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.CashFlowReport(p.Report, p.From, p.To, csrf, user))
}

type PeriodListPage struct{ Periods []dto.FinancialPeriodResponse }
func (p PeriodListPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.PeriodList(p.Periods, csrf, user))
}

type PeriodFormPage struct{}
func (p PeriodFormPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.PeriodForm(csrf, user))
}

type ParentDashboardPage struct {
	Stats *dto.ParentDashboardStats
	Flash, FlashType string
}
func (p ParentDashboardPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.ParentDashboard(p.Stats, p.Flash, p.FlashType, csrf, user))
}

type ParentProfilePage struct {
	Profile *dto.ParentResponse
	Flash, FlashType string
}
func (p ParentProfilePage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.ParentProfile(p.Profile, p.Flash, p.FlashType, csrf, user))
}

type ParentChildrenPage struct{ Children []dto.ParentChildResponse }
func (p ParentChildrenPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.ParentChildren(p.Children, csrf, user))
}

type ParentPortalAttendancePage struct{ View *dto.ParentAttendanceView }
func (p ParentPortalAttendancePage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.ParentPortalAttendance(p.View, csrf, user))
}

type ParentPortalFeePage struct {
	Summary  *dto.ParentFeeSummary
	Gateways []dto.PaymentGatewayResponse
	Online   []dto.GatewayTransactionResponse
}
func (p ParentPortalFeePage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.ParentPortalFee(p.Summary, p.Gateways, p.Online, csrf, user))
}

type ParentPortalResultsPage struct {
	Student *dto.StudentResponse
	Results []dto.ExamResultResponse
}
func (p ParentPortalResultsPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.ParentPortalResults(p.Student, p.Results, csrf, user))
}

type ParentPortalResultPage struct {
	Result *dto.ExamResultResponse
	StudentID uuid.UUID
}
func (p ParentPortalResultPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.ParentPortalResult(p.Result, p.StudentID.String(), csrf, user))
}

type ParentNoticesPage struct{ Data *dto.PaginatedNotices }
func (p ParentNoticesPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.ParentNoticesList(p.Data, csrf, user))
}

type ParentNoticeDetailPage struct{ Notice *dto.NoticeResponse }
func (p ParentNoticeDetailPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.ParentNoticeDetail(p.Notice, csrf, user))
}

type ParentNotificationsPage struct{ Data *dto.PaginatedNotifications }
func (p ParentNotificationsPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.ParentNotificationsList(p.Data, csrf, user))
}

type ParentListPage struct {
	Data *dto.PaginatedParents
	Students []dto.StudentResponse
}
func (p ParentListPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.ParentAdminList(p.Data, p.Students, csrf, user))
}

type ParentFormPage struct {
	Title string
	Parent *dto.ParentResponse
	Students []dto.StudentResponse
}
func (p ParentFormPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.ParentAdminForm(p.Title, p.Parent, p.Students, csrf, user))
}

type ParentDetailPage struct {
	Parent *dto.ParentResponse
	Children []dto.ParentChildResponse
	Students []dto.StudentResponse
}
func (p ParentDetailPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.ParentAdminDetail(p.Parent, p.Children, p.Students, csrf, user))
}

type NoticeListPage struct{ Data *dto.PaginatedNotices }
func (p NoticeListPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.NoticeListPage(p.Data, csrf, user))
}

type NoticeFormPage struct {
	Title string
	Notice *dto.NoticeResponse
}
func (p NoticeFormPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.NoticeFormPage(p.Title, p.Notice, csrf, user))
}

type CommunicationDashboardPage struct{ Stats *dto.CommunicationDashboardStats }
func (p CommunicationDashboardPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.CommunicationDashboard(p.Stats, csrf, user))
}

type AdmissionFormData struct {
	Sessions []dto.AcademicSessionResponse
	Classes  []dto.ClassResponse
}

func toAdmissionFormData(f *AdmissionFormData) *pages.AdmissionFormData {
	if f == nil {
		return &pages.AdmissionFormData{}
	}
	return &pages.AdmissionFormData{Sessions: f.Sessions, Classes: f.Classes}
}

type PublicAdmissionFormPage struct {
	FormData *AdmissionFormData
	Flash, FlashType string
}
func (p PublicAdmissionFormPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.PublicAdmissionApply(toAdmissionFormData(p.FormData), p.Flash, p.FlashType, csrf))
}

type PublicAdmissionSuccessPage struct{ Application *dto.AdmissionApplicationResponse }
func (p PublicAdmissionSuccessPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.PublicAdmissionSuccess(p.Application))
}

type PublicAdmissionTrackPage struct{ Flash, FlashType string }
func (p PublicAdmissionTrackPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.PublicAdmissionTrack(p.Flash, p.FlashType, csrf))
}

type PublicAdmissionTrackResultPage struct {
	Application *dto.AdmissionApplicationResponse
	Gateways    []dto.PaymentGatewayResponse
}
func (p PublicAdmissionTrackResultPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.PublicAdmissionTrackResult(p.Application, p.Gateways, csrf))
}

type AdmissionDashboardPage struct{ Stats *dto.AdmissionDashboardStats }
func (p AdmissionDashboardPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.AdmissionDashboard(p.Stats, csrf, user))
}

type AdmissionListPage struct{ Data *dto.PaginatedAdmissionApplications }
func (p AdmissionListPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.AdmissionList(p.Data, csrf, user))
}

type AdmissionDetailPage struct {
	Application *dto.AdmissionApplicationResponse
	Sections []dto.SectionResponse
}
func (p AdmissionDetailPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.AdmissionDetail(p.Application, p.Sections, csrf, user))
}

type PublicSiteHomePage struct{ Data *dto.PublicSiteData }
func (p PublicSiteHomePage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.PublicSiteHome(p.Data))
}

type PublicSitePageView struct {
	Page *dto.WebsitePageResponse
	Settings *dto.WebsiteSettingsResponse
}
func (p PublicSitePageView) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.PublicSitePageView(p.Page, p.Settings))
}

type PublicNewsListPage struct {
	Data *dto.PaginatedNews
	Settings *dto.WebsiteSettingsResponse
}
func (p PublicNewsListPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.PublicNewsListPage(p.Data, p.Settings))
}

type PublicNewsDetailPage struct {
	News *dto.NewsResponse
	Settings *dto.WebsiteSettingsResponse
}
func (p PublicNewsDetailPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.PublicNewsDetailPage(p.News, p.Settings))
}

type PublicEventsListPage struct {
	Data *dto.PaginatedEvents
	Settings *dto.WebsiteSettingsResponse
}
func (p PublicEventsListPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.PublicEventsListPage(p.Data, p.Settings))
}

type PublicDownloadsPage struct {
	Downloads []dto.DownloadResponse
	Settings *dto.WebsiteSettingsResponse
}
func (p PublicDownloadsPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.PublicDownloadsPage(p.Downloads, p.Settings))
}

type PublicGalleryPage struct {
	Gallery []dto.WebsiteGalleryResponse
	Settings *dto.WebsiteSettingsResponse
}
func (p PublicGalleryPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.PublicGalleryPage(p.Gallery, p.Settings))
}

type PublicContactPage struct {
	Settings *dto.WebsiteSettingsResponse
	Flash, FlashType string
}
func (p PublicContactPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.PublicContactPage(p.Settings, p.Flash, p.FlashType, csrf))
}

type WebsiteDashboardPage struct{ Stats *dto.WebsiteDashboardStats }
func (p WebsiteDashboardPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.WebsiteDashboard(p.Stats, csrf, user))
}

type WebsiteSettingsPage struct {
	Settings *dto.WebsiteSettingsResponse
	Flash, FlashType string
}
func (p WebsiteSettingsPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.WebsiteSettingsForm(p.Settings, p.Flash, p.FlashType, csrf, user))
}

type WebsitePagesListPage struct{ Pages []dto.WebsitePageResponse }
func (p WebsitePagesListPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.WebsitePagesAdmin(p.Pages, csrf, user))
}

type WebsitePageFormPage struct {
	Title string
	Page *dto.WebsitePageResponse
}
func (p WebsitePageFormPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.WebsitePageForm(p.Title, p.Page, csrf, user))
}

type WebsiteBannersPage struct{ Banners []dto.WebsiteBannerResponse }
func (p WebsiteBannersPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.WebsiteBannersAdmin(p.Banners, csrf, user))
}

type WebsiteGalleryAdminPage struct{ Gallery []dto.WebsiteGalleryResponse }
func (p WebsiteGalleryAdminPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.WebsiteGalleryAdmin(p.Gallery, csrf, user))
}

type WebsiteNewsAdminPage struct{ Data *dto.PaginatedNews }
func (p WebsiteNewsAdminPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.WebsiteNewsAdmin(p.Data, csrf, user))
}

type WebsiteNewsFormPage struct{}
func (p WebsiteNewsFormPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.WebsiteNewsForm(csrf, user))
}

type WebsiteEventsAdminPage struct{ Data *dto.PaginatedEvents }
func (p WebsiteEventsAdminPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.WebsiteEventsAdmin(p.Data, csrf, user))
}

type WebsiteDownloadsAdminPage struct{ Downloads []dto.DownloadResponse }
func (p WebsiteDownloadsAdminPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.WebsiteDownloadsAdmin(p.Downloads, csrf, user))
}

type WebsiteContactsPage struct{ Data *dto.PaginatedContactMessages }
func (p WebsiteContactsPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.WebsiteContactsAdmin(p.Data, csrf, user))
}

type PaymentDashboardPage struct{ Stats *dto.PaymentDashboardStats }
func (p PaymentDashboardPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.PaymentDashboard(p.Stats, csrf, user))
}

type PaymentGatewaysPage struct {
	Gateways []dto.PaymentGatewayResponse
	Flash, FlashType string
}
func (p PaymentGatewaysPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.PaymentGateways(p.Gateways, p.Flash, p.FlashType, csrf, user))
}

type PaymentTransactionsPage struct {
	Data   *dto.PaginatedGatewayTransactions
	Filter dto.PaymentReportFilter
}
func (p PaymentTransactionsPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.PaymentTransactions(p.Data, p.Filter, csrf, user))
}

type PaymentReportsPage struct {
	Report []dto.GatewayCollectionReport
	Failed *dto.PaginatedGatewayTransactions
	From, To string
}
func (p PaymentReportsPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.PaymentReports(p.Report, p.Failed, p.From, p.To, csrf, user))
}

type PaymentRefundsPage struct {
	Data *dto.PaginatedRefunds
	Flash, FlashType string
}
func (p PaymentRefundsPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.PaymentRefunds(p.Data, p.Flash, p.FlashType, csrf, user))
}

type PaymentResultPage struct {
	Success bool
	Ref     string
}
func (p PaymentResultPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.PaymentResult(p.Success, p.Ref))
}

type ParentPayNowPage struct{ Data *dto.ParentPayNowData }
func (p ParentPayNowPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.ParentPayNow(p.Data, csrf, user))
}

type SystemHealthPage struct{ Stats *dto.HealthDashboardStats }
func (p SystemHealthPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.SystemHealth(p.Stats, csrf, user))
}

type SystemSettingsPage struct {
	Settings *dto.SystemSettingsResponse
	Flash, FlashType string
}
func (p SystemSettingsPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.SystemSettings(p.Settings, p.Flash, p.FlashType, csrf, user))
}

type BackupListPage struct {
	Data *dto.PaginatedBackups
	Flash, FlashType string
}
func (p BackupListPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.BackupList(p.Data, p.Flash, p.FlashType, csrf, user))
}

type AuditCenterPage struct {
	Data   *dto.PaginatedAuditLogs
	Filter dto.AuditSearchFilter
}
func (p AuditCenterPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.AuditCenter(p.Data, p.Filter, csrf, user))
}

type LicensePage struct {
	License *dto.LicenseResponse
	SampleKey string
	Flash, FlashType string
}
func (p LicensePage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.LicensePage(p.License, p.SampleKey, p.Flash, p.FlashType, csrf, user))
}

type SecurityDashboardPage struct{ Stats *dto.SecurityDashboardStats }
func (p SecurityDashboardPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.SecurityDashboard(p.Stats, csrf, user))
}

type EmailTemplatesPage struct{ Templates []dto.EmailTemplateResponse }
func (p EmailTemplatesPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.EmailTemplatesList(p.Templates, csrf, user))
}

type InstallPage struct{ Flash, FlashType string }
func (p InstallPage) Render(csrf string, user *dto.AuthUser, appName string) string {
	return render(pages.InstallWizard(p.Flash, p.FlashType, csrf))
}
