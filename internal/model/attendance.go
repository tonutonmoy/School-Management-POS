package model

const (
	EntityStudentAttendance = "student_attendance"
	EntityTeacherAttendance = "teacher_attendance"
	EntityStaffAttendance   = "staff_attendance"
	EntityLeaveRequest      = "leave_request"
)

const (
	AttendancePresent = "present"
	AttendanceAbsent  = "absent"
	AttendanceLate    = "late"
	AttendanceLeave   = "leave"
)

const (
	LeaveTypeCasual    = "casual"
	LeaveTypeSick      = "sick"
	LeaveTypeAnnual    = "annual"
	LeaveTypeEmergency = "emergency"
)

const (
	LeaveStatusPending  = "pending"
	LeaveStatusApproved = "approved"
	LeaveStatusRejected = "rejected"
)

const (
	LeaveEntityTeacher = "teacher"
	LeaveEntityStaff   = "staff"
)

const (
	PermAttendanceStudentMark = "attendance.student.mark"
	PermAttendanceStudentView = "attendance.student.view"
	PermAttendanceTeacherMark = "attendance.teacher.mark"
	PermAttendanceTeacherView = "attendance.teacher.view"
	PermAttendanceStaffMark   = "attendance.staff.mark"
	PermAttendanceStaffView   = "attendance.staff.view"
	PermLeaveApprove          = "leave.approve"
	PermLeaveReject           = "leave.reject"
)
