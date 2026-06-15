package model

const (
	RoleParent = "parent"

	EntityParent       = "parent"
	EntityNotice       = "notice"
	EntityNotification = "notification"
	EntitySMSLog       = "sms_log"
	EntityEmailLog     = "email_log"

	PermParentView        = "parent.view"
	PermNoticeCreate      = "notice.create"
	PermNoticeUpdate      = "notice.update"
	PermNoticeDelete      = "notice.delete"
	PermNotificationSend  = "notification.send"

	NoticeTypeGeneral = "general"
	NoticeTypeExam    = "exam"
	NoticeTypeHoliday = "holiday"
	NoticeTypeFee     = "fee"
	NoticeTypeUrgent  = "urgent"

	NoticeAudienceAllParents = "all_parents"
	NoticeAudienceAllUsers   = "all_users"

	NotifyCategoryAttendance = "attendance"
	NotifyCategoryFee        = "fee"
	NotifyCategoryPayment    = "payment"
	NotifyCategoryResult     = "result"
	NotifyCategoryNotice     = "notice"
	NotifyCategorySystem     = "system"

	SMSEventAbsent         = "attendance_absent"
	SMSEventFeeDue         = "fee_due_reminder"
	SMSEventPayment        = "payment_received"
	SMSEventPaymentFailed  = "payment_failed"
	SMSEventRefund         = "refund_processed"
	SMSEventResult         = "result_published"
	SMSEventNotice         = "new_notice"

	EmailEventReceipt      = "payment_receipt"
	EmailEventPaymentFailed = "payment_failed"
	EmailEventRefund       = "refund_processed"
	EmailEventResult       = "result_published"
	EmailEventNotice       = "new_notice"
	EmailEventPasswordReset = "password_reset"

	DeliveryPending   = "pending"
	DeliverySent      = "sent"
	DeliveryFailed    = "failed"
	DeliveryQueued    = "queued"

	ParentRelFather   = "father"
	ParentRelMother   = "mother"
	ParentRelGuardian = "guardian"
	ParentRelOther    = "other"
)
