package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/school-management/pos/internal/auth"
	"github.com/school-management/pos/internal/dto"
	"github.com/school-management/pos/internal/model"
	"github.com/school-management/pos/internal/repository"
)

type ParentService struct {
	repos *repository.Repositories
	audit *AuditService
}

func NewParentService(repos *repository.Repositories, audit *AuditService) *ParentService {
	return &ParentService{repos: repos, audit: audit}
}

func (s *ParentService) Create(ctx context.Context, req dto.CreateParentRequest, actorID uuid.UUID, ip string) (*dto.ParentResponse, error) {
	existing, err := s.repos.Users.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrEmailExists
	}
	role, err := s.repos.Roles.GetBySlug(ctx, model.RoleParent)
	if err != nil || role == nil {
		return nil, fmt.Errorf("parent role not found")
	}
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}
	user, err := s.repos.Users.Create(ctx, repository.CreateUserParams{
		Email: strings.ToLower(req.Email), PasswordHash: hash,
		FirstName: req.FirstName, LastName: req.LastName, Phone: req.Phone,
		RoleID: role.ID, IsActive: true,
	})
	if err != nil {
		return nil, err
	}
	parent, err := s.repos.Parents.Create(ctx, user.ID, req.Phone, req.Address, req.Occupation)
	if err != nil {
		return nil, err
	}
	for _, link := range req.StudentLinks {
		if link.StudentID == uuid.Nil {
			continue
		}
		_ = s.repos.Parents.LinkStudent(ctx, parent.ID, link.StudentID, link.Relationship, link.IsPrimary)
	}
	resp := mapParentList(&repository.ParentListRecord{
		ParentRecord: *parent, Email: user.Email, FirstName: user.FirstName,
		LastName: user.LastName, IsActive: user.IsActive, ChildCount: len(req.StudentLinks),
	})
	s.audit.Log(ctx, &actorID, model.ActionCreate, model.EntityParent, &parent.ID, ip, map[string]any{"email": user.Email})
	return &resp, nil
}

func (s *ParentService) Update(ctx context.Context, id uuid.UUID, req dto.UpdateParentRequest, actorID uuid.UUID, ip string) (*dto.ParentResponse, error) {
	parent, err := s.repos.Parents.GetByID(ctx, id)
	if err != nil || parent == nil {
		return nil, ErrNotFound
	}
	user, err := s.repos.Users.GetByID(ctx, parent.UserID)
	if err != nil || user == nil {
		return nil, ErrNotFound
	}
	updated, err := s.repos.Users.Update(ctx, parent.UserID, repository.UpdateUserParams{
		Email: user.Email, FirstName: req.FirstName, LastName: req.LastName,
		Phone: req.Phone, RoleID: user.RoleID, IsActive: req.IsActive,
	})
	if err != nil {
		return nil, err
	}
	p, err := s.repos.Parents.Update(ctx, id, req.Phone, req.Address, req.Occupation)
	if err != nil {
		return nil, err
	}
	children, _ := s.repos.Parents.ListStudents(ctx, id)
	resp := mapParentList(&repository.ParentListRecord{
		ParentRecord: *p, Email: updated.Email, FirstName: updated.FirstName,
		LastName: updated.LastName, IsActive: updated.IsActive, ChildCount: len(children),
	})
	s.audit.Log(ctx, &actorID, model.ActionUpdate, model.EntityParent, &id, ip, nil)
	return &resp, nil
}

func (s *ParentService) Delete(ctx context.Context, id uuid.UUID, actorID uuid.UUID, ip string) error {
	parent, err := s.repos.Parents.GetByID(ctx, id)
	if err != nil || parent == nil {
		return ErrNotFound
	}
	if err := s.repos.Parents.SoftDelete(ctx, id); err != nil {
		return err
	}
	_ = s.repos.Users.SoftDelete(ctx, parent.UserID)
	s.audit.Log(ctx, &actorID, model.ActionDelete, model.EntityParent, &id, ip, nil)
	return nil
}

func (s *ParentService) Get(ctx context.Context, id uuid.UUID) (*dto.ParentResponse, error) {
	parent, err := s.repos.Parents.GetByID(ctx, id)
	if err != nil || parent == nil {
		return nil, ErrNotFound
	}
	user, err := s.repos.Users.GetByID(ctx, parent.UserID)
	if err != nil || user == nil {
		return nil, ErrNotFound
	}
	children, _ := s.repos.Parents.ListStudents(ctx, id)
	resp := mapParentList(&repository.ParentListRecord{
		ParentRecord: *parent, Email: user.Email, FirstName: user.FirstName,
		LastName: user.LastName, IsActive: user.IsActive, ChildCount: len(children),
	})
	return &resp, nil
}

func (s *ParentService) List(ctx context.Context, page, pageSize int) (*dto.PaginatedParents, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := int32((page - 1) * pageSize)
	items, err := s.repos.Parents.List(ctx, int32(pageSize), offset)
	if err != nil {
		return nil, err
	}
	total, err := s.repos.Parents.Count(ctx)
	if err != nil {
		return nil, err
	}
	resp := make([]dto.ParentResponse, 0, len(items))
	for i := range items {
		resp = append(resp, mapParentList(&items[i]))
	}
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}
	return &dto.PaginatedParents{Items: resp, Total: total, Page: page, PageSize: pageSize, TotalPages: totalPages}, nil
}

func (s *ParentService) LinkStudent(ctx context.Context, parentID uuid.UUID, link dto.ParentLinkInput, actorID uuid.UUID, ip string) error {
	parent, err := s.repos.Parents.GetByID(ctx, parentID)
	if err != nil || parent == nil {
		return ErrNotFound
	}
	st, err := s.repos.Students.GetByID(ctx, link.StudentID)
	if err != nil || st == nil {
		return ErrNotFound
	}
	if err := s.repos.Parents.LinkStudent(ctx, parentID, link.StudentID, link.Relationship, link.IsPrimary); err != nil {
		return err
	}
	s.audit.Log(ctx, &actorID, model.ActionUpdate, model.EntityParent, &parentID, ip, map[string]any{"student_id": link.StudentID})
	return nil
}

func (s *ParentService) UnlinkStudent(ctx context.Context, parentID, studentID uuid.UUID, actorID uuid.UUID, ip string) error {
	if err := s.repos.Parents.UnlinkStudent(ctx, parentID, studentID); err != nil {
		return err
	}
	s.audit.Log(ctx, &actorID, model.ActionUpdate, model.EntityParent, &parentID, ip, map[string]any{"unlink": studentID})
	return nil
}

func (s *ParentService) ListChildren(ctx context.Context, userID uuid.UUID) ([]dto.ParentChildResponse, error) {
	parent, err := s.repos.Parents.GetByUserID(ctx, userID)
	if err != nil || parent == nil {
		return nil, ErrNotFound
	}
	recs, err := s.repos.Parents.ListStudents(ctx, parent.ID)
	if err != nil {
		return nil, err
	}
	return mapChildren(recs), nil
}

func (s *ParentService) UpdateProfile(ctx context.Context, userID uuid.UUID, req dto.UpdateParentProfileRequest) (*dto.ParentResponse, error) {
	parent, err := s.repos.Parents.GetByUserID(ctx, userID)
	if err != nil || parent == nil {
		return nil, ErrNotFound
	}
	p, err := s.repos.Parents.Update(ctx, parent.ID, req.Phone, req.Address, req.Occupation)
	if err != nil {
		return nil, err
	}
	user, _ := s.repos.Users.GetByID(ctx, userID)
	children, _ := s.repos.Parents.ListStudents(ctx, parent.ID)
	resp := mapParentList(&repository.ParentListRecord{
		ParentRecord: *p, Email: user.Email, FirstName: user.FirstName,
		LastName: user.LastName, IsActive: user.IsActive, ChildCount: len(children),
	})
	return &resp, nil
}

func (s *ParentService) GetProfile(ctx context.Context, userID uuid.UUID) (*dto.ParentResponse, error) {
	parent, err := s.repos.Parents.GetByUserID(ctx, userID)
	if err != nil || parent == nil {
		return nil, ErrNotFound
	}
	return s.Get(ctx, parent.ID)
}

func (s *ParentService) CanViewStudent(ctx context.Context, user *dto.AuthUser, studentID uuid.UUID, staffPerms ...string) bool {
	if user == nil {
		return false
	}
	if user.RoleSlug == model.RoleAdmin {
		return true
	}
	if user.RoleSlug == model.RoleParent {
		ok, _ := s.repos.Parents.HasStudentAccess(ctx, user.ID, studentID)
		return ok
	}
	for _, perm := range staffPerms {
		for _, p := range user.Permissions {
			if p == perm {
				return true
			}
		}
	}
	return false
}

func (s *ParentService) Dashboard(ctx context.Context, userID uuid.UUID) (*dto.ParentDashboardStats, error) {
	parent, err := s.repos.Parents.GetByUserID(ctx, userID)
	if err != nil || parent == nil {
		return nil, ErrNotFound
	}
	children, err := s.repos.Parents.ListStudents(ctx, parent.ID)
	if err != nil {
		return nil, err
	}
	stats := &dto.ParentDashboardStats{
		ChildrenCount: len(children),
		Children:      mapChildren(children),
	}
	var totalPct float64
	var pctCount int
	var totalDue float64
	for _, ch := range children {
		summary, _ := s.repos.Attendance.StudentAttendanceSummary(ctx, ch.StudentID, nil, nil)
		if summary != nil {
			total := summary.PresentDays + summary.AbsentDays + summary.LateDays + summary.LeaveDays
			if total > 0 {
				totalPct += float64(summary.PresentDays+summary.LateDays) / float64(total) * 100
				pctCount++
			}
		}
		bills, _ := s.repos.Fees.SearchBills(ctx, repository.BillSearchParams{StudentID: &ch.StudentID, Limit: 50})
		for _, b := range bills {
			if b.Status != model.BillStatusCancelled && b.Status != model.BillStatusPaid {
				totalDue += b.TotalAmount - b.PaidAmount
			}
		}
		results, _ := s.repos.Exams.ListExamResults(ctx, repository.ResultSearchParams{
			StudentID: &ch.StudentID, PublishedOnly: true, Limit: 1,
		})
		if len(results) > 0 && stats.LatestExamTitle == "" {
			stats.LatestExamTitle = results[0].ExamName
			stats.LatestExamGPA = results[0].GPA
			if results[0].MeritPosition != nil {
				stats.LatestExamPosition = *results[0].MeritPosition
			}
		}
	}
	if pctCount > 0 {
		stats.AttendancePct = totalPct / float64(pctCount)
	}
	stats.CurrentDue = totalDue
	activities, _ := s.repos.Parents.RecentActivities(ctx, parent.ID, 15)
	for _, a := range activities {
		stats.RecentActivities = append(stats.RecentActivities, dto.ParentActivityItem{
			ID: a.ID, Category: a.Category, Title: a.Title,
			Description: a.Description, StudentName: a.StudentName, CreatedAt: a.CreatedAt,
		})
	}
	unreadNotices, _ := s.repos.Notices.CountUnreadForParent(ctx, parent.ID)
	stats.UnreadNotices = int(unreadNotices)
	unreadNotifs, _ := s.repos.Notifications.CountForParent(ctx, parent.ID, true)
	stats.UnreadNotifications = int(unreadNotifs)
	return stats, nil
}

func (s *ParentService) AttendanceView(ctx context.Context, userID, studentID uuid.UUID, month time.Time) (*dto.ParentAttendanceView, error) {
	parent, err := s.repos.Parents.GetByUserID(ctx, userID)
	if err != nil || parent == nil {
		return nil, ErrNotFound
	}
	ok, _ := s.repos.Parents.HasStudentAccess(ctx, userID, studentID)
	if !ok {
		return nil, ErrForbidden
	}
	st, err := s.repos.Students.GetByID(ctx, studentID)
	if err != nil || st == nil {
		return nil, ErrNotFound
	}
	today := time.Now().Truncate(24 * time.Hour)
	daily, _ := s.repos.Attendance.ListStudentAttendanceReport(ctx, repository.AttendanceReportParams{
		StudentID: &studentID, From: today, To: today,
	})
	mStart := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, month.Location())
	mEnd := mStart.AddDate(0, 1, -1)
	monthly, _ := s.repos.Attendance.ListStudentAttendanceReport(ctx, repository.AttendanceReportParams{
		StudentID: &studentID, From: mStart, To: mEnd,
	})
	histFrom := today.AddDate(0, -3, 0)
	history, _ := s.repos.Attendance.ListStudentAttendanceReport(ctx, repository.AttendanceReportParams{
		StudentID: &studentID, From: histFrom, To: today, Limit: 30,
	})
	summaryRec, _ := s.repos.Attendance.StudentAttendanceSummary(ctx, studentID, nil, nil)
	var summary dto.StudentAttendanceSummary
	if summaryRec != nil {
		total := summaryRec.PresentDays + summaryRec.AbsentDays + summaryRec.LateDays + summaryRec.LeaveDays
		pct := 0.0
		if total > 0 {
			pct = float64(summaryRec.PresentDays+summaryRec.LateDays) / float64(total) * 100
		}
		summary = dto.StudentAttendanceSummary{
			StudentID: studentID, StudentName: st.FirstName + " " + st.LastName,
			PresentDays: summaryRec.PresentDays, AbsentDays: summaryRec.AbsentDays,
			LateDays: summaryRec.LateDays, LeaveDays: summaryRec.LeaveDays,
			TotalMarked: total, AttendancePct: pct,
		}
	}
	return &dto.ParentAttendanceView{
		Summary: &summary,
		Daily:   mapStudentAttendanceRecords(daily),
		Monthly: mapStudentAttendanceRecords(monthly),
		History: mapStudentAttendanceRecords(history),
		Student: mapStudentBrief(st),
	}, nil
}

func (s *ParentService) StaffAttendanceView(ctx context.Context, studentID uuid.UUID, month time.Time) (*dto.ParentAttendanceView, error) {
	st, err := s.repos.Students.GetByID(ctx, studentID)
	if err != nil || st == nil {
		return nil, ErrNotFound
	}
	today := time.Now().Truncate(24 * time.Hour)
	daily, _ := s.repos.Attendance.ListStudentAttendanceReport(ctx, repository.AttendanceReportParams{
		StudentID: &studentID, From: today, To: today,
	})
	mStart := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, month.Location())
	mEnd := mStart.AddDate(0, 1, -1)
	monthly, _ := s.repos.Attendance.ListStudentAttendanceReport(ctx, repository.AttendanceReportParams{
		StudentID: &studentID, From: mStart, To: mEnd,
	})
	histFrom := today.AddDate(0, -3, 0)
	history, _ := s.repos.Attendance.ListStudentAttendanceReport(ctx, repository.AttendanceReportParams{
		StudentID: &studentID, From: histFrom, To: today, Limit: 30,
	})
	summaryRec, _ := s.repos.Attendance.StudentAttendanceSummary(ctx, studentID, nil, nil)
	var summary dto.StudentAttendanceSummary
	if summaryRec != nil {
		total := summaryRec.PresentDays + summaryRec.AbsentDays + summaryRec.LateDays + summaryRec.LeaveDays
		pct := 0.0
		if total > 0 {
			pct = float64(summaryRec.PresentDays+summaryRec.LateDays) / float64(total) * 100
		}
		summary = dto.StudentAttendanceSummary{
			StudentID: studentID, StudentName: st.FirstName + " " + st.LastName,
			PresentDays: summaryRec.PresentDays, AbsentDays: summaryRec.AbsentDays,
			LateDays: summaryRec.LateDays, LeaveDays: summaryRec.LeaveDays,
			TotalMarked: total, AttendancePct: pct,
		}
	}
	return &dto.ParentAttendanceView{
		Summary: &summary,
		Daily:   mapStudentAttendanceRecords(daily),
		Monthly: mapStudentAttendanceRecords(monthly),
		History: mapStudentAttendanceRecords(history),
		Student: mapStudentBrief(st),
	}, nil
}

func mapParentList(r *repository.ParentListRecord) dto.ParentResponse {
	return dto.ParentResponse{
		ID: r.ID, UserID: r.UserID, Email: r.Email, FirstName: r.FirstName, LastName: r.LastName,
		Phone: r.Phone, Address: r.Address, Occupation: r.Occupation,
		IsActive: r.IsActive, ChildCount: r.ChildCount, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
}

func mapChildren(recs []repository.ParentStudentRecord) []dto.ParentChildResponse {
	items := make([]dto.ParentChildResponse, 0, len(recs))
	for _, r := range recs {
		items = append(items, dto.ParentChildResponse{
			ID: r.ID, StudentID: r.StudentID,
			StudentName: r.FirstName + " " + r.LastName,
			AdmissionNumber: r.AdmissionNumber, RollNumber: r.RollNumber,
			ClassName: r.ClassName, Relationship: r.Relationship, IsPrimary: r.IsPrimary,
		})
	}
	return items
}

func mapStudentBrief(st *repository.StudentRecord) *dto.StudentResponse {
	return &dto.StudentResponse{
		ID: st.ID, FirstName: st.FirstName, LastName: st.LastName,
		AdmissionNumber: st.AdmissionNumber, RollNumber: st.RollNumber,
	}
}

func mapStudentAttendanceRecords(recs []repository.StudentAttendanceRecord) []dto.StudentAttendanceResponse {
	items := make([]dto.StudentAttendanceResponse, 0, len(recs))
	for _, r := range recs {
		items = append(items, dto.StudentAttendanceResponse{
			ID: r.ID, StudentID: r.StudentID, StudentName: r.StudentName,
			ClassName: r.ClassName, SectionName: r.SectionName,
			AttendanceDate: r.AttendanceDate, Status: r.Status, Remarks: r.Remarks,
		})
	}
	return items
}

type NoticeService struct {
	repos *repository.Repositories
	audit *AuditService
	notify *NotificationService
}

func NewNoticeService(repos *repository.Repositories, audit *AuditService, notify *NotificationService) *NoticeService {
	return &NoticeService{repos: repos, audit: audit, notify: notify}
}

func (s *NoticeService) Create(ctx context.Context, req dto.NoticeRequest, actorID uuid.UUID, ip string) (*dto.NoticeResponse, error) {
	if req.NoticeType == "" {
		req.NoticeType = model.NoticeTypeGeneral
	}
	if req.TargetAudience == "" {
		req.TargetAudience = model.NoticeAudienceAllParents
	}
	if req.PublishAt.IsZero() {
		req.PublishAt = time.Now()
	}
	rec, err := s.repos.Notices.Create(ctx, repository.NoticeParams{
		Title: req.Title, Body: req.Body, NoticeType: req.NoticeType,
		TargetAudience: req.TargetAudience, PublishAt: req.PublishAt,
		ExpiresAt: req.ExpiresAt, IsPublished: req.IsPublished, CreatedBy: &actorID,
	})
	if err != nil {
		return nil, err
	}
	resp := mapNotice(rec)
	s.audit.Log(ctx, &actorID, model.ActionCreate, model.EntityNotice, &rec.ID, ip, map[string]any{"title": req.Title})
	if req.IsPublished && s.notify != nil {
		go s.notify.OnNewNotice(context.Background(), rec.ID, req.Title, req.Body)
	}
	return &resp, nil
}

func (s *NoticeService) Update(ctx context.Context, id uuid.UUID, req dto.NoticeRequest, actorID uuid.UUID, ip string) (*dto.NoticeResponse, error) {
	rec, err := s.repos.Notices.Update(ctx, id, repository.NoticeParams{
		Title: req.Title, Body: req.Body, NoticeType: req.NoticeType,
		TargetAudience: req.TargetAudience, PublishAt: req.PublishAt,
		ExpiresAt: req.ExpiresAt, IsPublished: req.IsPublished,
	})
	if err != nil || rec == nil {
		return nil, ErrNotFound
	}
	resp := mapNotice(rec)
	s.audit.Log(ctx, &actorID, model.ActionUpdate, model.EntityNotice, &id, ip, nil)
	return &resp, nil
}

func (s *NoticeService) Delete(ctx context.Context, id uuid.UUID, actorID uuid.UUID, ip string) error {
	if err := s.repos.Notices.SoftDelete(ctx, id); err != nil {
		return err
	}
	s.audit.Log(ctx, &actorID, model.ActionDelete, model.EntityNotice, &id, ip, nil)
	return nil
}

func (s *NoticeService) Get(ctx context.Context, id uuid.UUID) (*dto.NoticeResponse, error) {
	rec, err := s.repos.Notices.GetByID(ctx, id)
	if err != nil || rec == nil {
		return nil, ErrNotFound
	}
	resp := mapNotice(rec)
	return &resp, nil
}

func (s *NoticeService) List(ctx context.Context, f dto.NoticeFilter) (*dto.PaginatedNotices, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 || f.PageSize > 100 {
		f.PageSize = 20
	}
	params := repository.NoticeSearchParams{
		Query: f.Query, NoticeType: f.NoticeType,
		Limit: int32(f.PageSize), Offset: int32((f.Page - 1) * f.PageSize),
	}
	total, err := s.repos.Notices.Count(ctx, params)
	if err != nil {
		return nil, err
	}
	recs, err := s.repos.Notices.Search(ctx, params)
	if err != nil {
		return nil, err
	}
	items := make([]dto.NoticeResponse, 0, len(recs))
	for i := range recs {
		items = append(items, mapNotice(&recs[i]))
	}
	totalPages := int(total) / f.PageSize
	if int(total)%f.PageSize > 0 {
		totalPages++
	}
	return &dto.PaginatedNotices{Items: items, Total: total, Page: f.Page, PageSize: f.PageSize, TotalPages: totalPages}, nil
}

func (s *NoticeService) ListForParent(ctx context.Context, parentID uuid.UUID, f dto.NoticeFilter) (*dto.PaginatedNotices, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 || f.PageSize > 100 {
		f.PageSize = 20
	}
	pub := true
	params := repository.NoticeSearchParams{
		Query: f.Query, NoticeType: f.NoticeType, ParentID: &parentID, Published: &pub,
		Limit: int32(f.PageSize), Offset: int32((f.Page - 1) * f.PageSize),
	}
	total, _ := s.repos.Notices.Count(ctx, params)
	recs, err := s.repos.Notices.Search(ctx, params)
	if err != nil {
		return nil, err
	}
	items := make([]dto.NoticeResponse, 0, len(recs))
	for i := range recs {
		items = append(items, mapNotice(&recs[i]))
	}
	totalPages := int(total) / f.PageSize
	if int(total)%f.PageSize > 0 {
		totalPages++
	}
	return &dto.PaginatedNotices{Items: items, Total: total, Page: f.Page, PageSize: f.PageSize, TotalPages: totalPages}, nil
}

func (s *NoticeService) MarkRead(ctx context.Context, noticeID, parentID uuid.UUID) error {
	return s.repos.Notices.MarkRead(ctx, noticeID, parentID)
}

func mapNotice(r *repository.NoticeRecord) dto.NoticeResponse {
	return dto.NoticeResponse{
		ID: r.ID, Title: r.Title, Body: r.Body, NoticeType: r.NoticeType,
		TargetAudience: r.TargetAudience, PublishAt: r.PublishAt, ExpiresAt: r.ExpiresAt,
		IsPublished: r.IsPublished, IsRead: r.IsRead, CreatedByName: r.CreatedByName,
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
}
