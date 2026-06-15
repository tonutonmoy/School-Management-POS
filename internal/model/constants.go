package model

const (
	ActionLogin  = "login"
	ActionLogout = "logout"
	ActionLoginFailed = "login_failed"
	ActionCreate = "create"
	ActionUpdate = "update"
	ActionDelete = "delete"
)

const (
	EntityUser            = "user"
	EntityRole            = "role"
	EntityPermission      = "permission"
	EntitySchool          = "school"
	EntityAcademicSession = "academic_session"
	EntityClass           = "class"
	EntitySection         = "section"
	EntitySubject         = "subject"
	EntityDepartment      = "department"
	EntityStudent         = "student"
	EntityTeacher         = "teacher"
	EntityStaff           = "staff"
	EntityDesignation     = "designation"
)

const (
	PermUserManage     = "user.manage"
	PermRoleManage     = "role.manage"
	PermSchoolManage   = "school.manage"
	PermSessionManage  = "session.manage"
	PermDashboardView  = "dashboard.view"
	PermAuditView      = "audit.view"
	PermStudentCreate  = "student.create"
	PermStudentUpdate  = "student.update"
	PermStudentDelete  = "student.delete"
	PermStudentView    = "student.view"
	PermClassCreate    = "class.create"
	PermClassUpdate    = "class.update"
	PermClassDelete    = "class.delete"
	PermSubjectCreate  = "subject.create"
	PermSubjectUpdate  = "subject.update"
	PermSubjectDelete  = "subject.delete"
	PermTeacherCreate  = "teacher.create"
	PermTeacherUpdate  = "teacher.update"
	PermTeacherDelete  = "teacher.delete"
	PermTeacherView    = "teacher.view"
	PermStaffCreate    = "staff.create"
	PermStaffUpdate    = "staff.update"
	PermStaffDelete    = "staff.delete"
	PermStaffView      = "staff.view"
	PermDepartmentCreate = "department.create"
	PermDepartmentUpdate = "department.update"
	PermDepartmentDelete = "department.delete"
	PermDesignationCreate = "designation.create"
	PermDesignationUpdate = "designation.update"
	PermDesignationDelete = "designation.delete"
	PermFeesCollect    = "fees.collect"
	PermFeesRefund     = "fees.refund"
)

const (
	RoleAdmin      = "admin"
	RolePrincipal  = "principal"
	RoleAccountant = "accountant"
	RoleTeacher    = "teacher"
	RoleStaff      = "staff"
)

const (
	StudentStatusActive      = "active"
	StudentStatusInactive    = "inactive"
	StudentStatusGraduated   = "graduated"
	StudentStatusTransferred = "transferred"
)

const (
	DocBirthCertificate   = "birth_certificate"
	DocPreviousMarksheet  = "previous_marksheet"
	DocPassportPhoto      = "passport_photo"
	DocOther              = "other"
)

const (
	PromotionTypePromote  = "promote"
	PromotionTypeTransfer = "transfer"
)
