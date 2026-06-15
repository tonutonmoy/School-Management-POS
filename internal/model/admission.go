package model

const (
	EntityAdmissionApplication = "admission_application"
	EntityWebsitePage          = "website_page"
	EntityContactMessage       = "contact_message"

	PermAdmissionReview  = "admission.review"
	PermAdmissionApprove = "admission.approve"
	PermAdmissionReject  = "admission.reject"
	PermWebsiteManage    = "website.manage"

	AdmissionPending     = "pending"
	AdmissionUnderReview = "under_review"
	AdmissionApproved    = "approved"
	AdmissionRejected    = "rejected"
	AdmissionAdmitted    = "admitted"

	AdmissionPaymentUnpaid  = "unpaid"
	AdmissionPaymentPending = "pending"
	AdmissionPaymentPaid    = "paid"
	AdmissionPaymentWaived  = "waived"

	ContactStatusNew     = "new"
	ContactStatusRead    = "read"
	ContactStatusReplied = "replied"
	ContactStatusClosed  = "closed"

	PageTypeHome       = "home"
	PageTypeAbout      = "about"
	PageTypePrincipal  = "principal"
	PageTypeAdmission  = "admission"
	PageTypeTeachers   = "teachers"
	PageTypeContact    = "contact"
	PageTypeCustom     = "custom"

	DownloadProspectus = "prospectus"
	DownloadCircular   = "circular"
	DownloadForm       = "form"
	DownloadDocument   = "document"
)
