package i18n

import "strings"

type Locale string

const (
	EN Locale = "en"
	BN Locale = "bn"
)

func Parse(s string) Locale {
	if strings.ToLower(strings.TrimSpace(s)) == "bn" {
		return BN
	}
	return EN
}

func (l Locale) HTMLLang() string {
	if l == BN {
		return "bn"
	}
	return "en"
}

func T(locale Locale, key string) string {
	if m, ok := catalog[locale][key]; ok {
		return m
	}
	if m, ok := catalog[EN][key]; ok {
		return m
	}
	return key
}

func PageTitle(locale Locale, english string) string {
	if key, ok := pageKeys[english]; ok {
		return T(locale, key)
	}
	return english
}

var pageKeys = map[string]string{
	"Dashboard":          "page.dashboard",
	"Students":           "page.students",
	"Login":              "page.login",
	"Forgot Password":    "page.forgot_password",
	"Reset Password":     "page.reset_password",
	"Change Password":    "page.change_password",
	"Users":              "page.users",
	"Teachers":           "page.teachers",
	"System Health":      "page.system_health",
	"Payment Dashboard":  "page.payment_dashboard",
	"Attendance Dashboard": "page.attendance",
}

var catalog = map[Locale]map[string]string{
	EN: enMessages,
	BN: bnMessages,
}

var enMessages = map[string]string{
	"app.name":              "School POS",
	"app.tagline":           "Management System",
	"nav.overview":          "Overview",
	"nav.academic":          "Academic",
	"nav.finance":           "Finance",
	"nav.people":            "People",
	"nav.admin":             "Admin",
	"nav.more":              "More",
	"nav.dashboard":         "Dashboard",
	"nav.students":          "Students",
	"nav.attendance":        "Attendance",
	"nav.exams":             "Exams",
	"nav.admissions":        "Admissions",
	"nav.sections":          "Sections",
	"nav.subjects":          "Subjects",
	"nav.sessions":          "Sessions",
	"nav.finance_menu":      "Finance",
	"nav.payments":          "Payments",
	"nav.accounting":      "Accounting",
	"nav.teachers":          "Teachers",
	"nav.staff":             "Staff",
	"nav.departments":       "Departments",
	"nav.designations":      "Designations",
	"nav.teacher_portal":    "Teacher Portal",
	"nav.parents":           "Parents",
	"nav.users":             "Users",
	"nav.roles":             "Roles",
	"nav.school":            "School Setup",
	"nav.notices":           "Notices",
	"nav.communications":    "Communications",
	"nav.audit_logs":        "Audit Logs",
	"nav.website":           "Website",
	"nav.system":            "System",
	"ui.password":           "Password",
	"ui.logout":             "Logout",
	"ui.open_menu":          "Open menu",
	"ui.loading":            "Loading…",
	"ui.lang.en":            "EN",
	"ui.lang.bn":            "বাং",
	"ui.previous":           "← Previous",
	"ui.next":               "Next →",
	"ui.page_of":            "Page %d of %d",
	"login.title":           "Sign in",
	"login.subtitle":        "School Management System",
	"login.email":           "Email",
	"login.password":        "Password",
	"login.submit":          "Login",
	"login.forgot":          "Forgot password?",
	"page.dashboard":        "Dashboard",
	"page.students":         "Students",
	"page.login":            "Login",
	"page.forgot_password":  "Forgot Password",
	"page.reset_password":  "Reset Password",
	"page.change_password": "Change Password",
	"page.users":            "Users",
	"page.teachers":         "Teachers",
	"page.system_health":    "System Health",
	"page.payment_dashboard": "Payment Dashboard",
	"page.attendance":       "Attendance",
}

var bnMessages = map[string]string{
	"app.name":              "স্কুল POS",
	"app.tagline":           "ম্যানেজমেন্ট সিস্টেম",
	"nav.overview":          "সারাংশ",
	"nav.academic":          "একাডেমিক",
	"nav.finance":           "অর্থ",
	"nav.people":            "মানুষ",
	"nav.admin":             "অ্যাডমিন",
	"nav.more":              "আরও",
	"nav.dashboard":         "ড্যাশবোর্ড",
	"nav.students":          "শিক্ষার্থী",
	"nav.attendance":        "উপস্থিতি",
	"nav.exams":             "পরীক্ষা",
	"nav.admissions":        "ভর্তি",
	"nav.sections":          "শাখা",
	"nav.subjects":          "বিষয়",
	"nav.sessions":          "সেশন",
	"nav.finance_menu":      "ফাইন্যান্স",
	"nav.payments":          "পেমেন্ট",
	"nav.accounting":        "হিসাব",
	"nav.teachers":          "শিক্ষক",
	"nav.staff":             "কর্মী",
	"nav.departments":       "বিভাগ",
	"nav.designations":      "পদবি",
	"nav.teacher_portal":    "শিক্ষক পোর্টাল",
	"nav.parents":           "অভিভাবক",
	"nav.users":             "ব্যবহারকারী",
	"nav.roles":             "ভূমিকা",
	"nav.school":            "স্কুল সেটআপ",
	"nav.notices":           "নোটিশ",
	"nav.communications":    "যোগাযোগ",
	"nav.audit_logs":        "অডিট লগ",
	"nav.website":           "ওয়েবসাইট",
	"nav.system":            "সিস্টেম",
	"ui.password":           "পাসওয়ার্ড",
	"ui.logout":             "লগআউট",
	"ui.open_menu":          "মেনু খুলুন",
	"ui.loading":            "লোড হচ্ছে…",
	"ui.lang.en":            "EN",
	"ui.lang.bn":            "বাং",
	"ui.previous":           "← আগে",
	"ui.next":               "পরে →",
	"ui.page_of":            "পৃষ্ঠা %d / %d",
	"login.title":           "সাইন ইন",
	"login.subtitle":        "স্কুল ম্যানেজমেন্ট সিস্টেম",
	"login.email":           "ইমেইল",
	"login.password":        "পাসওয়ার্ড",
	"login.submit":          "লগইন",
	"login.forgot":          "পাসওয়ার্ড ভুলে গেছেন?",
	"page.dashboard":        "ড্যাশবোর্ড",
	"page.students":         "শিক্ষার্থী",
	"page.login":            "লগইন",
	"page.forgot_password":  "পাসওয়ার্ড রিসেট",
	"page.reset_password":  "নতুন পাসওয়ার্ড",
	"page.change_password": "পাসওয়ার্ড পরিবর্তন",
	"page.users":            "ব্যবহারকারী",
	"page.teachers":         "শিক্ষক",
	"page.system_health":    "সিস্টেম স্বাস্থ্য",
	"page.payment_dashboard": "পেমেন্ট ড্যাশবোর্ড",
	"page.attendance":       "উপস্থিতি",
}
