package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/school-management/pos/internal/dto"
	"github.com/school-management/pos/internal/model"
	"github.com/school-management/pos/internal/repository"
)

type AdmissionService struct {
	repos *repository.Repositories
	audit *AuditService
}

func NewAdmissionService(repos *repository.Repositories, audit *AuditService) *AdmissionService {
	return &AdmissionService{repos: repos, audit: audit}
}

func (s *AdmissionService) Submit(ctx context.Context, req dto.AdmissionApplicationRequest) (*dto.AdmissionApplicationResponse, error) {
	appNo, err := s.repos.Admissions.NextApplicationNumber(ctx, time.Now().Year())
	if err != nil {
		return nil, err
	}
	nationality := req.Nationality
	if nationality == "" {
		nationality = "Bangladeshi"
	}
	var sessID, classID, secID *uuid.UUID
	if req.SessionID != uuid.Nil {
		sessID = &req.SessionID
	}
	if req.ClassID != uuid.Nil {
		classID = &req.ClassID
	}
	if req.SectionID != uuid.Nil {
		secID = &req.SectionID
	}
	fee := req.AdmissionFee
	if fee <= 0 {
		fee = 500
	}
	rec, err := s.repos.Admissions.Create(ctx, repository.CreateAdmissionParams{
		ApplicationNumber: appNo,
		FirstName: req.FirstName, LastName: req.LastName, DateOfBirth: req.DateOfBirth,
		Gender: req.Gender, BloodGroup: req.BloodGroup, Religion: req.Religion, Nationality: nationality,
		Phone: req.Phone, Email: req.Email, Address: req.Address,
		FatherName: req.FatherName, FatherPhone: req.FatherPhone, FatherOccupation: req.FatherOccupation,
		MotherName: req.MotherName, MotherPhone: req.MotherPhone, MotherOccupation: req.MotherOccupation,
		GuardianName: req.GuardianName, GuardianPhone: req.GuardianPhone,
		PreviousSchool: req.PreviousSchool, PreviousClass: req.PreviousClass, PreviousBoard: req.PreviousBoard,
		SessionID: sessID, ClassID: classID, SectionID: secID, AdmissionFeeAmount: fee,
	})
	if err != nil {
		return nil, err
	}
	resp := mapAdmission(rec, nil)
	return &resp, nil
}

func (s *AdmissionService) Track(ctx context.Context, appNo, token string) (*dto.AdmissionApplicationResponse, error) {
	rec, err := s.repos.Admissions.GetByTracking(ctx, appNo, token)
	if err != nil || rec == nil {
		return nil, ErrNotFound
	}
	docs, _ := s.repos.Admissions.ListDocuments(ctx, rec.IDUUID)
	resp := mapAdmission(rec, docs)
	return &resp, nil
}

func (s *AdmissionService) Get(ctx context.Context, id uuid.UUID) (*dto.AdmissionApplicationResponse, error) {
	rec, err := s.repos.Admissions.GetByID(ctx, id)
	if err != nil || rec == nil {
		return nil, ErrNotFound
	}
	docs, _ := s.repos.Admissions.ListDocuments(ctx, id)
	resp := mapAdmission(rec, docs)
	return &resp, nil
}

func (s *AdmissionService) List(ctx context.Context, f dto.AdmissionSearchFilter) (*dto.PaginatedAdmissionApplications, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 || f.PageSize > 100 {
		f.PageSize = 20
	}
	params := mapAdmissionFilter(f)
	total, err := s.repos.Admissions.Count(ctx, params)
	if err != nil {
		return nil, err
	}
	recs, err := s.repos.Admissions.Search(ctx, params)
	if err != nil {
		return nil, err
	}
	items := make([]dto.AdmissionApplicationResponse, 0, len(recs))
	for i := range recs {
		items = append(items, mapAdmission(&recs[i], nil))
	}
	totalPages := int(total) / f.PageSize
	if int(total)%f.PageSize > 0 {
		totalPages++
	}
	return &dto.PaginatedAdmissionApplications{
		Items: items, Total: total, Page: f.Page, PageSize: f.PageSize, TotalPages: totalPages,
	}, nil
}

func (s *AdmissionService) DashboardStats(ctx context.Context) (*dto.AdmissionDashboardStats, error) {
	sr, err := s.repos.Admissions.Stats(ctx)
	if err != nil {
		return nil, err
	}
	return &dto.AdmissionDashboardStats{
		TotalApplications: sr.Total, PendingCount: sr.Pending, UnderReviewCount: sr.UnderReview,
		ApprovedCount: sr.Approved, AdmittedCount: sr.Admitted, RejectedCount: sr.Rejected,
		TodayApplications: sr.Today,
	}, nil
}

func (s *AdmissionService) UnderReview(ctx context.Context, id uuid.UUID, notes string, actorID uuid.UUID, ip string) error {
	return s.setStatus(ctx, id, model.AdmissionUnderReview, notes, actorID, ip)
}

func (s *AdmissionService) Approve(ctx context.Context, id uuid.UUID, notes string, actorID uuid.UUID, ip string) error {
	return s.setStatus(ctx, id, model.AdmissionApproved, notes, actorID, ip)
}

func (s *AdmissionService) Reject(ctx context.Context, id uuid.UUID, notes string, actorID uuid.UUID, ip string) error {
	return s.setStatus(ctx, id, model.AdmissionRejected, notes, actorID, ip)
}

func (s *AdmissionService) setStatus(ctx context.Context, id uuid.UUID, status, notes string, actorID uuid.UUID, ip string) error {
	rec, err := s.repos.Admissions.GetByID(ctx, id)
	if err != nil || rec == nil {
		return ErrNotFound
	}
	if err := s.repos.Admissions.UpdateStatus(ctx, id, status, notes, &actorID); err != nil {
		return err
	}
	s.audit.Log(ctx, &actorID, model.ActionUpdate, model.EntityAdmissionApplication, &id, ip, map[string]any{"status": status})
	return nil
}

func (s *AdmissionService) RecordPayment(ctx context.Context, id uuid.UUID, req dto.AdmissionPaymentRequest, actorID uuid.UUID, ip string) (*dto.AdmissionApplicationResponse, error) {
	rec, err := s.repos.Admissions.GetByID(ctx, id)
	if err != nil || rec == nil {
		return nil, ErrNotFound
	}
	receipt := fmt.Sprintf("ADM-RCP-%d-%05d", time.Now().Year(), time.Now().Unix()%100000)
	amount := req.Amount
	if amount <= 0 {
		amount = rec.AdmissionFeeAmount
	}
	if err := s.repos.Admissions.UpdatePayment(ctx, id, model.AdmissionPaymentPaid, req.PaymentReference, receipt, amount); err != nil {
		return nil, err
	}
	s.audit.Log(ctx, &actorID, model.ActionUpdate, model.EntityAdmissionApplication, &id, ip, map[string]any{"payment": "paid"})
	return s.Get(ctx, id)
}

func (s *AdmissionService) AddDocument(ctx context.Context, appID uuid.UUID, docType, fileName, fileURL string) error {
	_, err := s.repos.Admissions.AddDocument(ctx, appID, docType, fileName, fileURL)
	return err
}

func (s *AdmissionService) ConvertToStudent(ctx context.Context, id uuid.UUID, sectionID uuid.UUID, actorID uuid.UUID, ip string) (*dto.StudentResponse, error) {
	rec, err := s.repos.Admissions.GetByID(ctx, id)
	if err != nil || rec == nil {
		return nil, ErrNotFound
	}
	if rec.StatusVal != model.AdmissionApproved && rec.StatusVal != model.AdmissionUnderReview {
		return nil, fmt.Errorf("%w: application must be approved before admission", ErrValidation)
	}
	if rec.StudentID != nil {
		return nil, fmt.Errorf("%w: already admitted", ErrValidation)
	}
	if rec.SessionID == nil || rec.ClassID == nil {
		return nil, fmt.Errorf("%w: session and class required", ErrValidation)
	}
	secID := sectionID
	if secID == uuid.Nil && rec.SectionID != nil {
		secID = *rec.SectionID
	}
	if secID == uuid.Nil {
		sections, _ := s.repos.Academic.ListSectionsByClass(ctx, *rec.ClassID)
		if len(sections) == 0 {
			return nil, fmt.Errorf("%w: no section available for class", ErrValidation)
		}
		secID = sections[0].ID
	}

	studentSvc := NewStudentService(s.repos, s.audit)
	admitReq := dto.StudentAdmissionRequest{
		FirstName: rec.FirstName, LastName: rec.LastName, DateOfBirth: rec.DateOfBirth,
		Gender: rec.Gender, BloodGroup: rec.BloodGroup, Religion: rec.Religion, Nationality: rec.Nationality,
		Phone: rec.Phone, Email: rec.Email, Address: rec.Address,
		SessionID: *rec.SessionID, ClassID: *rec.ClassID, SectionID: secID,
		AdmissionDate: time.Now(),
		FatherName: rec.FatherName, FatherPhone: rec.FatherPhone, FatherOccupation: rec.FatherOccupation,
		MotherName: rec.MotherName, MotherPhone: rec.MotherPhone, MotherOccupation: rec.MotherOccupation,
		GuardianName: rec.GuardianName, GuardianPhone: rec.GuardianPhone,
	}
	student, err := studentSvc.Admit(ctx, admitReq, "", actorID, ip)
	if err != nil {
		return nil, err
	}

	var parentUserID *uuid.UUID
	if rec.Email != "" {
		parentSvc := NewParentService(s.repos, s.audit)
		pwd := rec.Token
		if len(pwd) < 8 {
			pwd = "Parent123!"
		} else {
			pwd = pwd[:8] + "Aa1!"
		}
		fname := rec.FatherName
		if fname == "" {
			fname = "Parent"
		}
		parent, perr := parentSvc.Create(ctx, dto.CreateParentRequest{
			Email: rec.Email, Password: pwd, FirstName: fname, LastName: rec.LastName,
			Phone: rec.FatherPhone,
			StudentLinks: []dto.ParentLinkInput{{StudentID: student.ID, Relationship: "guardian", IsPrimary: true}},
		}, actorID, ip)
		if perr == nil && parent != nil {
			parentUserID = &parent.UserID
		}
	}

	if err := s.repos.Admissions.SetAdmitted(ctx, id, student.ID, parentUserID); err != nil {
		return nil, err
	}
	s.audit.Log(ctx, &actorID, model.ActionUpdate, model.EntityAdmissionApplication, &id, ip, map[string]any{
		"status": model.AdmissionAdmitted, "student_id": student.ID,
	})
	return student, nil
}

func (s *AdmissionService) ExportCSV(ctx context.Context, f dto.AdmissionSearchFilter) ([]repository.AdmissionRecord, error) {
	params := mapAdmissionFilter(f)
	params.Limit = 10000
	return s.repos.Admissions.ExportList(ctx, params)
}

func mapAdmissionFilter(f dto.AdmissionSearchFilter) repository.AdmissionSearchParams {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 {
		f.PageSize = 20
	}
	return repository.AdmissionSearchParams{
		Query: f.Query, Status: f.Status, PaymentStatus: f.PaymentStatus,
		SessionID: f.SessionID, ClassID: f.ClassID, From: f.From, To: f.To,
		Limit: int32(f.PageSize), Offset: int32((f.Page - 1) * f.PageSize),
	}
}

func mapAdmission(r *repository.AdmissionRecord, docs []repository.AdmissionDocumentRecord) dto.AdmissionApplicationResponse {
	resp := dto.AdmissionApplicationResponse{
		ID: r.IDUUID, ApplicationNumber: r.AppNumber, TrackingToken: r.Token, Status: r.StatusVal,
		FirstName: r.FirstName, LastName: r.LastName, FullName: r.FirstName + " " + r.LastName,
		DateOfBirth: r.DateOfBirth, Gender: r.Gender, BloodGroup: r.BloodGroup,
		Religion: r.Religion, Nationality: r.Nationality, Phone: r.Phone, Email: r.Email, Address: r.Address,
		FatherName: r.FatherName, FatherPhone: r.FatherPhone, FatherOccupation: r.FatherOccupation,
		MotherName: r.MotherName, MotherPhone: r.MotherPhone, MotherOccupation: r.MotherOccupation,
		GuardianName: r.GuardianName, GuardianPhone: r.GuardianPhone,
		PreviousSchool: r.PreviousSchool, PreviousClass: r.PreviousClass, PreviousBoard: r.PreviousBoard,
		SessionID: r.SessionID, SessionName: r.SessionName, ClassID: r.ClassID, ClassName: r.ClassName,
		SectionID: r.SectionID, SectionName: r.SectionName, StudentID: r.StudentID,
		AdmissionFeeAmount: r.AdmissionFeeAmount, PaymentStatus: r.PaymentStatus,
		PaymentReference: r.PaymentReference, ReceiptNumber: r.ReceiptNumber,
		ReviewNotes: r.ReviewNotes, ReviewedByName: r.ReviewedByName, ReviewedAt: r.ReviewedAt,
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
	for _, d := range docs {
		resp.Documents = append(resp.Documents, dto.AdmissionDocumentResponse{
			ID: d.ID, DocType: d.DocType, FileName: d.FileName, FileURL: d.FileURL, CreatedAt: d.CreatedAt,
		})
	}
	return resp
}
