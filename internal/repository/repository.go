package repository

import "github.com/jackc/pgx/v5/pgxpool"

type Repositories struct {
	Users            UserRepository
	Roles            RoleRepository
	Permissions      PermissionRepository
	Auth             AuthRepository
	Schools          SchoolRepository
	AcademicSessions AcademicSessionRepository
	AuditLogs        AuditRepository
	Academic         AcademicRepository
	Students         StudentRepository
	HR               HRRepository
	Attendance       AttendanceRepository
	Fees             FeeRepository
	Exams            ExamRepository
	Accounting       AccountingRepository
	Parents          ParentRepository
	Notices          NoticeRepository
	Notifications    NotificationRepository
	Admissions       AdmissionRepository
	Website          WebsiteRepository
	Payments         PaymentRepository
	System           SystemRepository
}

func New(pool *pgxpool.Pool) *Repositories {
	return &Repositories{
		Users:            NewUserRepository(pool),
		Roles:            NewRoleRepository(pool),
		Permissions:      NewPermissionRepository(pool),
		Auth:             NewAuthRepository(pool),
		Schools:          NewSchoolRepository(pool),
		AcademicSessions: NewAcademicSessionRepository(pool),
		AuditLogs:        NewAuditRepository(pool),
		Academic:         NewAcademicRepository(pool),
		Students:         NewStudentRepository(pool),
		HR:               NewHRRepository(pool),
		Attendance:       NewAttendanceRepository(pool),
		Fees:             NewFeeRepository(pool),
		Exams:            NewExamRepository(pool),
		Accounting:       NewAccountingRepository(pool),
		Parents:          NewParentRepository(pool),
		Notices:          NewNoticeRepository(pool),
		Notifications:    NewNotificationRepository(pool),
		Admissions:       NewAdmissionRepository(pool),
		Website:          NewWebsiteRepository(pool),
		Payments:         NewPaymentRepository(pool),
		System:           NewSystemRepository(pool),
	}
}
