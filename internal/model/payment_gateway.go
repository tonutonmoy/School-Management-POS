package model

const (
	EntityPaymentGateway     = "payment_gateway"
	EntityGatewayTransaction = "gateway_transaction"
	EntityPaymentRefund      = "payment_refund"

	PermPaymentManage    = "payment.manage"
	PermPaymentRefund    = "payment.refund"
	PermPaymentReportView = "payment.report.view"

	GatewayBkash       = "bkash"
	GatewayNagad       = "nagad"
	GatewaySSLCommerz  = "sslcommerz"

	GwPaymentAdmission = "admission_fee"
	GwPaymentStudentFee = "student_fee"

	GwStatusPending    = "pending"
	GwStatusProcessing = "processing"
	GwStatusCompleted  = "completed"
	GwStatusFailed     = "failed"
	GwStatusRefunded   = "refunded"
	GwStatusCancelled  = "cancelled"

	RefundRequested = "requested"
	RefundApproved  = "approved"
	RefundProcessed = "processed"
	RefundRejected  = "rejected"
)
