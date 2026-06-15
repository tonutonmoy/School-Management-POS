package model

const (
	EntityFeeType         = "fee_type"
	EntityFeeStructure    = "fee_structure"
	EntityStudentDiscount = "student_discount"
	EntityStudentBill     = "student_bill"
	EntityPayment         = "payment"
	EntityReceipt         = "receipt"
)

const (
	PermFeeCreate     = "fee.create"
	PermFeeUpdate     = "fee.update"
	PermFeeDelete     = "fee.delete"
	PermFeeCollect    = "fee.collect"
	PermFeeRefund     = "fee.refund"
	PermFeeReportView = "fee.report.view"
)

const (
	FreqOneTime    = "one_time"
	FreqMonthly    = "monthly"
	FreqQuarterly  = "quarterly"
	FreqHalfYearly = "half_yearly"
	FreqYearly     = "yearly"
)

const (
	BillStatusPending   = "pending"
	BillStatusPartial   = "partial"
	BillStatusPaid      = "paid"
	BillStatusOverdue   = "overdue"
	BillStatusCancelled = "cancelled"
)

const (
	DiscountFixed      = "fixed"
	DiscountPercentage = "percentage"
)

const (
	DiscountScholarship = "scholarship"
	DiscountWaiver      = "waiver"
	DiscountSibling     = "sibling"
	DiscountSpecial     = "special"
)

const (
	PayMethodCash  = "cash"
	PayMethodBank  = "bank"
	PayMethodBkash = "bkash"
	PayMethodNagad = "nagad"
	PayMethodRocket = "rocket"
	PayMethodCard  = "card"
)

const (
	PaymentCompleted = "completed"
	PaymentRefunded  = "refunded"
)

const (
	SeqInvoice = "invoice"
	SeqPayment = "payment"
	SeqReceipt = "receipt"
)
