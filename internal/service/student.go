package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/school-management/pos/internal/dto"
	"github.com/school-management/pos/internal/model"
	"github.com/school-management/pos/internal/repository"
)

type StudentService struct {
	repos *repository.Repositories
	audit *AuditService
}

func NewStudentService(repos *repository.Repositories, audit *AuditService) *StudentService {
	return &StudentService{repos: repos, audit: audit}
}

func (s *StudentService) Admit(ctx context.Context, req dto.StudentAdmissionRequest, photoURL string, actorID uuid.UUID, ip string) (*dto.StudentResponse, error) {
	admissionNo, err := s.repos.Students.NextAdmissionNumber(ctx, req.AdmissionDate.Year())
	if err != nil {
		return nil, err
	}
	status := req.Status
	if status == "" {
		status = model.StudentStatusActive
	}
	var deptID *uuid.UUID
	if req.DepartmentID != uuid.Nil {
		deptID = &req.DepartmentID
	}
	nationality := req.Nationality
	if nationality == "" {
		nationality = "Bangladeshi"
	}

	student, err := s.repos.Students.Create(ctx, repository.CreateStudentParams{
		AdmissionNumber: admissionNo,
		RollNumber:      req.RollNumber,
		FirstName:       req.FirstName,
		LastName:        req.LastName,
		DateOfBirth:     req.DateOfBirth,
		Gender:          req.Gender,
		BloodGroup:      req.BloodGroup,
		Religion:        req.Religion,
		Nationality:     nationality,
		PhotoURL:        photoURL,
		Phone:           req.Phone,
		Email:           req.Email,
		Address:         req.Address,
		SessionID:       req.SessionID,
		ClassID:         req.ClassID,
		SectionID:       req.SectionID,
		DepartmentID:    deptID,
		AdmissionDate:   req.AdmissionDate,
		Status:          status,
	})
	if err != nil {
		return nil, err
	}

	_ = s.repos.Students.UpsertParents(ctx, repository.StudentParentParams{
		StudentID: student.ID, FatherName: req.FatherName, FatherPhone: req.FatherPhone,
		FatherOccupation: req.FatherOccupation, MotherName: req.MotherName, MotherPhone: req.MotherPhone,
		MotherOccupation: req.MotherOccupation, GuardianName: req.GuardianName, GuardianPhone: req.GuardianPhone,
	})

	resp := s.mapStudent(student, nil, nil)
	s.audit.Log(ctx, &actorID, model.ActionCreate, model.EntityStudent, &student.ID, ip, map[string]any{"admission_number": admissionNo})
	return &resp, nil
}

func (s *StudentService) Update(ctx context.Context, id uuid.UUID, req dto.StudentAdmissionRequest, photoURL string, actorID uuid.UUID, ip string) (*dto.StudentResponse, error) {
	var deptID *uuid.UUID
	if req.DepartmentID != uuid.Nil {
		deptID = &req.DepartmentID
	}
	status := req.Status
	if status == "" {
		status = model.StudentStatusActive
	}
	student, err := s.repos.Students.Update(ctx, id, repository.UpdateStudentParams{
		RollNumber: req.RollNumber, FirstName: req.FirstName, LastName: req.LastName,
		DateOfBirth: req.DateOfBirth, Gender: req.Gender, BloodGroup: req.BloodGroup,
		Religion: req.Religion, Nationality: req.Nationality, PhotoURL: photoURL,
		Phone: req.Phone, Email: req.Email, Address: req.Address,
		SessionID: req.SessionID, ClassID: req.ClassID, SectionID: req.SectionID,
		DepartmentID: deptID, AdmissionDate: req.AdmissionDate, Status: status,
	})
	if err != nil {
		return nil, err
	}
	if student == nil {
		return nil, ErrNotFound
	}
	_ = s.repos.Students.UpsertParents(ctx, repository.StudentParentParams{
		StudentID: id, FatherName: req.FatherName, FatherPhone: req.FatherPhone,
		FatherOccupation: req.FatherOccupation, MotherName: req.MotherName, MotherPhone: req.MotherPhone,
		MotherOccupation: req.MotherOccupation, GuardianName: req.GuardianName, GuardianPhone: req.GuardianPhone,
	})
	resp, err := s.GetFull(ctx, id)
	if err != nil {
		return nil, err
	}
	s.audit.Log(ctx, &actorID, model.ActionUpdate, model.EntityStudent, &id, ip, map[string]any{"admission_number": student.AdmissionNumber})
	return resp, nil
}

func (s *StudentService) Delete(ctx context.Context, id uuid.UUID, actorID uuid.UUID, ip string) error {
	student, err := s.repos.Students.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if student == nil {
		return ErrNotFound
	}
	if err := s.repos.Students.SoftDelete(ctx, id); err != nil {
		return err
	}
	s.audit.Log(ctx, &actorID, model.ActionDelete, model.EntityStudent, &id, ip, map[string]any{"admission_number": student.AdmissionNumber})
	return nil
}

func (s *StudentService) GetFull(ctx context.Context, id uuid.UUID) (*dto.StudentResponse, error) {
	student, err := s.repos.Students.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if student == nil {
		return nil, ErrNotFound
	}
	parents, _ := s.repos.Students.GetParents(ctx, id)
	docs, _ := s.repos.Students.ListDocuments(ctx, id)
	resp := s.mapStudent(student, parents, docs)
	return &resp, nil
}

func (s *StudentService) Search(ctx context.Context, filter dto.StudentSearchFilter) (*dto.PaginatedStudents, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 20
	}
	params := repository.StudentSearchParams{
		AdmissionNumber: filter.AdmissionNumber,
		RollNumber:      filter.RollNumber,
		Name:            filter.Name,
		ClassID:         filter.ClassID,
		SectionID:       filter.SectionID,
		SessionID:       filter.SessionID,
		Limit:           int32(filter.PageSize),
		Offset:          int32((filter.Page - 1) * filter.PageSize),
	}
	total, err := s.repos.Students.CountSearch(ctx, params)
	if err != nil {
		return nil, err
	}
	recs, err := s.repos.Students.Search(ctx, params)
	if err != nil {
		return nil, err
	}
	items := make([]dto.StudentResponse, 0, len(recs))
	for i := range recs {
		items = append(items, s.mapStudent(&recs[i], nil, nil))
	}
	totalPages := int(total) / filter.PageSize
	if int(total)%filter.PageSize > 0 {
		totalPages++
	}
	return &dto.PaginatedStudents{Items: items, Total: total, Page: filter.Page, PageSize: filter.PageSize, TotalPages: totalPages}, nil
}

func (s *StudentService) Promote(ctx context.Context, id uuid.UUID, req dto.PromoteStudentRequest, actorID uuid.UUID, ip string) (*dto.StudentResponse, error) {
	student, err := s.repos.Students.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if student == nil {
		return nil, ErrNotFound
	}
	if err := s.repos.Students.CreatePromotion(ctx, repository.PromotionParams{
		StudentID: id, PromotionType: model.PromotionTypePromote,
		FromSessionID: &student.SessionID, ToSessionID: req.ToSessionID,
		FromClassID: &student.ClassID, ToClassID: req.ToClassID,
		FromSectionID: &student.SectionID, ToSectionID: req.ToSectionID,
		PromotionDate: time.Now(), Notes: req.Notes, CreatedBy: actorID,
	}); err != nil {
		return nil, err
	}
	updated, err := s.repos.Students.Update(ctx, id, repository.UpdateStudentParams{
		RollNumber: student.RollNumber, FirstName: student.FirstName, LastName: student.LastName,
		DateOfBirth: student.DateOfBirth, Gender: student.Gender, BloodGroup: student.BloodGroup,
		Religion: student.Religion, Nationality: student.Nationality, PhotoURL: "",
		Phone: student.Phone, Email: student.Email, Address: student.Address,
		SessionID: req.ToSessionID, ClassID: req.ToClassID, SectionID: req.ToSectionID,
		DepartmentID: student.DepartmentID, AdmissionDate: student.AdmissionDate, Status: model.StudentStatusActive,
	})
	if err != nil {
		return nil, err
	}
	s.audit.Log(ctx, &actorID, model.ActionUpdate, model.EntityStudent, &id, ip, map[string]any{"action": "promote"})
	return s.GetFull(ctx, updated.ID)
}

func (s *StudentService) Transfer(ctx context.Context, id uuid.UUID, req dto.TransferStudentRequest, actorID uuid.UUID, ip string) (*dto.StudentResponse, error) {
	student, err := s.repos.Students.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if student == nil {
		return nil, ErrNotFound
	}
	if err := s.repos.Students.CreatePromotion(ctx, repository.PromotionParams{
		StudentID: id, PromotionType: model.PromotionTypeTransfer,
		FromSessionID: &student.SessionID, ToSessionID: req.ToSessionID,
		FromClassID: &student.ClassID, ToClassID: req.ToClassID,
		FromSectionID: &student.SectionID, ToSectionID: req.ToSectionID,
		PromotionDate: time.Now(), Notes: req.Notes, CreatedBy: actorID,
	}); err != nil {
		return nil, err
	}
	updated, err := s.repos.Students.Update(ctx, id, repository.UpdateStudentParams{
		RollNumber: student.RollNumber, FirstName: student.FirstName, LastName: student.LastName,
		DateOfBirth: student.DateOfBirth, Gender: student.Gender, BloodGroup: student.BloodGroup,
		Religion: student.Religion, Nationality: student.Nationality, PhotoURL: "",
		Phone: student.Phone, Email: student.Email, Address: student.Address,
		SessionID: req.ToSessionID, ClassID: req.ToClassID, SectionID: req.ToSectionID,
		DepartmentID: student.DepartmentID, AdmissionDate: student.AdmissionDate, Status: model.StudentStatusTransferred,
	})
	if err != nil {
		return nil, err
	}
	s.audit.Log(ctx, &actorID, model.ActionUpdate, model.EntityStudent, &id, ip, map[string]any{"action": "transfer"})
	return s.GetFull(ctx, updated.ID)
}

func (s *StudentService) AddDocument(ctx context.Context, studentID uuid.UUID, docType, fileName, fileURL string, actorID uuid.UUID, ip string) error {
	if _, err := s.repos.Students.GetByID(ctx, studentID); err != nil {
		return err
	}
	_, err := s.repos.Students.CreateDocument(ctx, studentID, docType, fileName, fileURL)
	if err != nil {
		return err
	}
	s.audit.Log(ctx, &actorID, model.ActionCreate, model.EntityStudent, &studentID, ip, map[string]any{"document": docType})
	return nil
}

func (s *StudentService) IDCardData(ctx context.Context, id uuid.UUID) (*dto.StudentIDCardData, error) {
	student, err := s.GetFull(ctx, id)
	if err != nil {
		return nil, err
	}
	school, _ := s.repos.Schools.Get(ctx)
	data := &dto.StudentIDCardData{Student: *student, IssueDate: time.Now()}
	if school != nil {
		data.SchoolName = school.Name
		data.SchoolLogo = school.LogoURL
	}
	return data, nil
}

func (s *StudentService) StudentListReport(ctx context.Context, filter dto.ReportFilter) ([]dto.StudentResponse, error) {
	recs, err := s.repos.Students.ListForReport(ctx, filter.ClassID, filter.SessionID, filter.Status)
	if err != nil {
		return nil, err
	}
	items := make([]dto.StudentResponse, 0, len(recs))
	for i := range recs {
		items = append(items, s.mapStudent(&recs[i], nil, nil))
	}
	return items, nil
}

func (s *StudentService) AdmissionReport(ctx context.Context, from, to time.Time) ([]dto.StudentResponse, error) {
	recs, err := s.repos.Students.ListAdmissionsReport(ctx, from, to)
	if err != nil {
		return nil, err
	}
	items := make([]dto.StudentResponse, 0, len(recs))
	for i := range recs {
		items = append(items, s.mapStudent(&recs[i], nil, nil))
	}
	return items, nil
}

func (s *StudentService) mapStudent(st *repository.StudentRecord, parents *repository.StudentParentRecord, docs []repository.StudentDocumentRecord) dto.StudentResponse {
	resp := dto.StudentResponse{
		ID: st.ID, AdmissionNumber: st.AdmissionNumber, RollNumber: st.RollNumber,
		FirstName: st.FirstName, LastName: st.LastName,
		FullName: strings.TrimSpace(st.FirstName + " " + st.LastName),
		DateOfBirth: st.DateOfBirth, Gender: st.Gender, BloodGroup: st.BloodGroup,
		Religion: st.Religion, Nationality: st.Nationality, PhotoURL: st.PhotoURL,
		Phone: st.Phone, Email: st.Email, Address: st.Address,
		SessionID: st.SessionID, SessionName: st.SessionName,
		ClassID: st.ClassID, ClassName: st.ClassName,
		SectionID: st.SectionID, SectionName: st.SectionName,
		DepartmentID: st.DepartmentID, DepartmentName: st.DepartmentName,
		AdmissionDate: st.AdmissionDate, Status: st.Status,
		CreatedAt: st.CreatedAt, UpdatedAt: st.UpdatedAt,
	}
	if parents != nil {
		resp.Parents = &dto.StudentParentResponse{
			FatherName: parents.FatherName, FatherPhone: parents.FatherPhone, FatherOccupation: parents.FatherOccupation,
			MotherName: parents.MotherName, MotherPhone: parents.MotherPhone, MotherOccupation: parents.MotherOccupation,
			GuardianName: parents.GuardianName, GuardianPhone: parents.GuardianPhone,
		}
	}
	for _, d := range docs {
		resp.Documents = append(resp.Documents, dto.StudentDocumentResponse{ID: d.ID, DocType: d.DocType, FileName: d.FileName, FileURL: d.FileURL})
	}
	return resp
}

func (s *StudentService) SectionsHTMX(ctx context.Context, classID uuid.UUID, fieldName string) (string, error) {
	if fieldName == "" {
		fieldName = "section_id"
	}
	sections, err := s.repos.Academic.ListSectionsByClass(ctx, classID)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`<select name="%s" required class="w-full rounded-lg border px-3 py-2"><option value="">Select section</option>`, fieldName))
	for _, sec := range sections {
		b.WriteString(fmt.Sprintf(`<option value="%s">%s</option>`, sec.ID, sec.Name))
	}
	b.WriteString(`</select>`)
	return b.String(), nil
}
