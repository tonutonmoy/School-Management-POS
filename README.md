# School Management System — Foundation Module

Production-ready foundation for a school management POS built with Go, Fiber v3, PostgreSQL, HTMX, and Templ.

## Stack

- **Go 1.24+** / Fiber v3
- **PostgreSQL** with Goose migrations
- **SQLC** query definitions (`sql/queries/` — run `make sqlc` where CGO/sqlc is available)
- **Templ** + **HTMX** + **Tailwind CSS**
- **JWT** auth, bcrypt passwords, CSRF, rate limiting, structured logging
- **Clean architecture**: handler → service → repository

## Quick start

```bash
cp .env.example .env
# Edit DATABASE_URL, JWT_SECRET, CSRF_SECRET

make generate   # templ + tailwind (+ sqlc if available)
make migrate    # or migrations run automatically on server start
make run
```

Default admin (seeded on first run):

- Email: `admin@school.local`
- Password: `Admin@123456`

## Docker

```bash
docker compose up --build
```

## Modules

### Foundation (Part 1)

| Module | Routes |
|--------|--------|
| Auth | `/login`, `/logout`, `/forgot-password`, `/reset-password`, `/change-password` |
| Users | `/users` (CRUD, activate/deactivate) |
| Roles | `/roles` (CRUD, permission assignment) |
| School | `/school` (setup + logo upload to R2) |
| Sessions | `/sessions` (one active session enforced) |
| Dashboard | `/dashboard` |
| Audit | `/audit-logs` |

### Academic & Students (Part 2)

| Module | Routes |
|--------|--------|
| Classes | `/classes` (CRUD, assign subjects) |
| Sections | `/sections` (CRUD, assign to class) |
| Subjects | `/subjects` (CRUD) |
| Students | `/students` (admission, edit, profile, promote, transfer, documents, ID card) |
| Reports | `/reports/students`, `/reports/students/by-class`, `/reports/admissions` |

Departments seeded: Science, Commerce, Arts.

Admission numbers auto-generated as `ADM-YYYY-NNNNN`.

### HR — Teachers, Staff & Departments (Part 3)

| Module | Routes |
|--------|--------|
| Departments | `/departments` (CRUD) |
| Designations | `/designations` (CRUD) |
| Teachers | `/teachers` (CRUD, profile, subject/class/section assignment, documents) |
| Staff | `/staff` (CRUD, profile, documents) |
| Teacher Portal | `/teacher/dashboard` (assigned classes, subjects, today's schedule) |
| HR Reports | `/reports/hr/teachers`, `/reports/hr/staff`, `/reports/hr/department`, `/reports/hr/assignments` |
| Exports | `/reports/hr/teachers/export.csv`, `.xlsx`, `/reports/hr/staff/export.csv`, `.xlsx` |

Employee IDs auto-generated as `TCH-YYYY-NNNNN` (teachers) and `STF-YYYY-NNNNN` (staff).

Permissions: `teacher.*`, `staff.*`, `department.*`, `designation.*` (see migration `006_hr_permissions.sql`).

### Attendance & Leave (Part 4)

| Module | Routes |
|--------|--------|
| Student Attendance | `/attendance/students` (bulk mark with session/class/section/date filters) |
| Teacher Attendance | `/attendance/teachers` |
| Staff Attendance | `/attendance/staff` |
| Attendance Dashboard | `/attendance/dashboard` |
| Leave Management | `/leave`, `/leave/apply`, approve/reject |
| Parent View | `/parent/students/:id/attendance` |
| Reports | `/reports/attendance/students/daily`, `/monthly`, `/by-class`, `/:id/history`, teacher/staff daily & monthly |
| Exports | `/reports/attendance/students/export.csv`, `.xlsx`, teacher/staff CSV |

One attendance record per student per day enforced via unique index. Statuses: present, absent, late, leave.

### Fees & Finance (Part 5)

| Module | Routes |
|--------|--------|
| Fee Types | `/fees/types` (CRUD, active/inactive) |
| Fee Structures | `/fees/structures` (session/class/section, frequency) |
| Billing | `/fees/bills`, `/fees/bills/generate` |
| Collection | `/fees/collect` (full/partial, multi-invoice) |
| Discounts | `/fees/discounts` |
| Due Management | `/fees/dues`, `/fees/overdue` |
| Finance Dashboard | `/finance/dashboard` |
| Receipts | `/receipts/:id`, `/receipts/:id/pdf`, `/receipts/verify/:token` |
| Parent View | `/parent/students/:id/fees` |
| Reports | `/reports/fees/collection/*`, ledger, payment history, income, by-method, by-fee-type |

Invoice format: `INV-YYYY-NNNNN`. Transaction-safe payment processing with allocations and PDF receipts with QR verification.

### Examinations & Results (Part 6)

| Module | Routes |
|--------|--------|
| Exams | `/exams` (CRUD, publish, archive) |
| Subject Config | `/exams/:id` (per-subject full/pass/written/MCQ/practical marks) |
| Marks Entry | `/exams/:id/marks`, `/exams/:id/marks/:subjectId` |
| Result Processing | `/exams/:id/process`, `/exams/:id/publish-results` |
| Results | `/exams/:id/results`, `/exams/:id/results/:resultId` |
| Tabulation | `/exams/:id/tabulation` (PDF/CSV/Excel export) |
| Merit List | `/exams/:id/merit-list` |
| Grading Systems | `/grading-systems` (default GPA + custom) |
| Report Cards | `/report-cards/:resultId/pdf` |
| Dashboard | `/exams/dashboard` |
| Parent View | `/parent/students/:id/results`, report card download |
| Reports | `/reports/exams/summary`, subject-performance, top/failed students, CSV export |

Default GPA scale: A+ (5.00) through F (0.00). Transaction-safe result processing with class/section/merit positions.

### Accounting (Part 7)

| Module | Routes |
|--------|--------|
| Dashboard | `/accounting/dashboard` |
| Chart of Accounts | `/accounting/accounts` (CRUD, hierarchy, disable) |
| Journal Entries | `/accounting/journal` (double-entry, balanced) |
| General Ledger | `/accounting/ledger` |
| Cash Book | `/accounting/cash-book` |
| Bank Book | `/accounting/bank-book` |
| Expenses | `/expenses` (create, approve, attachments) |
| Income | `/income` (donations, events, misc) |
| Reports | `/reports/accounting/trial-balance`, income-statement, balance-sheet, cash-flow |
| Period Closing | `/accounting/periods` (lock closed periods) |

Fee payments automatically post journal entries (debit Cash/Bank, credit income by fee type). Duplicate entries prevented via `source_type` + `source_id` unique constraint.

### Parent Portal & Notifications (Part 8)

| Module | Routes |
|--------|--------|
| Parent Dashboard | `/parent/dashboard` (children, attendance %, dues, latest result) |
| Parent Profile | `/parent/profile` |
| My Children | `/parent/children` |
| Attendance View | `/parent/students/:id/attendance` (daily, monthly, history, %) |
| Fees View | `/parent/students/:id/fees` (dues, payments, receipt download) |
| Results View | `/parent/students/:id/results`, report card PDF |
| Notices | `/parent/notices` |
| Notification Center | `/parent/notifications` (read/unread, mark all read) |
| Parent Admin | `/parents` (create accounts, link/unlink students) |
| Notice Management | `/notices` (CRUD, school/exam/holiday/fee types) |
| Communication Dashboard | `/communications/dashboard` (SMS/email stats, CSV export) |

Parent login redirects to `/parent/dashboard`. SMS/email use provider abstractions (`internal/notify`) with queue-ready `notification_queue` table. Events: absent attendance, payment received, result published, new notice.

Permissions: `parent.view`, `notice.create`, `notice.update`, `notice.delete`, `notification.send`.

### Online Admission & Website CMS (Part 9)

| Module | Routes |
|--------|--------|
| Public Website | `/site`, `/site/page/:slug`, `/site/news`, `/site/events`, `/site/downloads`, `/site/contact`, `/site/gallery` |
| Online Admission | `/admission/apply`, `/admission/track`, `/admission/success` |
| Admission Review | `/admissions`, `/admissions/:id`, approve/reject/admit, CSV export |
| Website CMS | `/website/dashboard`, `/website/pages`, `/website/banners`, `/website/gallery`, `/website/news`, `/website/events`, `/website/downloads`, `/website/contacts`, `/website/settings` |

Unauthenticated visitors land on `/site`. Admitting an application creates a student record and optional parent portal account.

Permissions: `admission.review`, `admission.approve`, `admission.reject`, `website.manage`.

## Project layout

```
cmd/server/          Application entrypoint
internal/
  auth/              JWT + password helpers
  config/            Environment config
  handler/           HTTP handlers
  middleware/        Auth, CSRF, logging
  repository/        PostgreSQL repositories
  service/           Business logic
  notify/            SMS & email provider abstractions
  storage/           Cloudflare R2 uploads
  validator/         Request validation
  web/               Templ render adapters
web/
  components/        Templ UI components
  layouts/           Base layouts
  pages/             Page templates
  static/            CSS assets
migrations/          Goose SQL migrations
sql/                 SQLC schema + queries
```

## Security notes

- Never commit `.env` (contains DB and R2 credentials).
- Rotate `JWT_SECRET` and `CSRF_SECRET` for production.
- Password reset tokens are logged only in `development` mode.
