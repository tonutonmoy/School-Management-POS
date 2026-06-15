package service

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/school-management/pos/internal/dto"
	"github.com/school-management/pos/internal/model"
	"github.com/school-management/pos/internal/repository"
)

type AcademicService struct {
	repos *repository.Repositories
	audit *AuditService
}

func NewAcademicService(repos *repository.Repositories, audit *AuditService) *AcademicService {
	return &AcademicService{repos: repos, audit: audit}
}

func (s *AcademicService) ListDepartments(ctx context.Context) ([]dto.DepartmentResponse, error) {
	recs, err := s.repos.Academic.ListDepartments(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]dto.DepartmentResponse, 0, len(recs))
	for _, r := range recs {
		items = append(items, dto.DepartmentResponse{ID: r.ID, Name: r.Name, Slug: r.Slug})
	}
	return items, nil
}

func (s *AcademicService) CreateClass(ctx context.Context, req dto.ClassRequest, actorID uuid.UUID, ip string) (*dto.ClassResponse, error) {
	rec, err := s.repos.Academic.CreateClass(ctx, req.Name, strings.ToUpper(req.Code), req.Description, req.SortOrder)
	if err != nil {
		return nil, err
	}
	resp := mapClass(rec)
	s.audit.Log(ctx, &actorID, model.ActionCreate, model.EntityClass, &rec.ID, ip, map[string]any{"name": rec.Name})
	return &resp, nil
}

func (s *AcademicService) UpdateClass(ctx context.Context, id uuid.UUID, req dto.ClassRequest, actorID uuid.UUID, ip string) (*dto.ClassResponse, error) {
	rec, err := s.repos.Academic.UpdateClass(ctx, id, req.Name, strings.ToUpper(req.Code), req.Description, req.SortOrder)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, ErrNotFound
	}
	resp := mapClass(rec)
	s.audit.Log(ctx, &actorID, model.ActionUpdate, model.EntityClass, &id, ip, map[string]any{"name": rec.Name})
	return &resp, nil
}

func (s *AcademicService) DeleteClass(ctx context.Context, id uuid.UUID, actorID uuid.UUID, ip string) error {
	rec, err := s.repos.Academic.GetClassByID(ctx, id)
	if err != nil {
		return err
	}
	if rec == nil {
		return ErrNotFound
	}
	if err := s.repos.Academic.SoftDeleteClass(ctx, id); err != nil {
		return err
	}
	s.audit.Log(ctx, &actorID, model.ActionDelete, model.EntityClass, &id, ip, map[string]any{"name": rec.Name})
	return nil
}

func (s *AcademicService) ListClasses(ctx context.Context) ([]dto.ClassResponse, error) {
	recs, err := s.repos.Academic.ListClasses(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]dto.ClassResponse, 0, len(recs))
	for i := range recs {
		items = append(items, mapClass(&recs[i]))
	}
	return items, nil
}

func (s *AcademicService) GetClass(ctx context.Context, id uuid.UUID) (*dto.ClassResponse, error) {
	rec, err := s.repos.Academic.GetClassByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, ErrNotFound
	}
	resp := mapClass(rec)
	subs, _ := s.repos.Academic.ListSubjectsByClass(ctx, id)
	resp.SectionCount = len(subs)
	return &resp, nil
}

func (s *AcademicService) CreateSection(ctx context.Context, req dto.SectionRequest, actorID uuid.UUID, ip string) (*dto.SectionResponse, error) {
	rec, err := s.repos.Academic.CreateSection(ctx, req.ClassID, req.Name, req.Capacity)
	if err != nil {
		return nil, err
	}
	resp := mapSection(rec)
	s.audit.Log(ctx, &actorID, model.ActionCreate, model.EntitySection, &rec.ID, ip, map[string]any{"name": rec.Name})
	return &resp, nil
}

func (s *AcademicService) UpdateSection(ctx context.Context, id uuid.UUID, req dto.SectionRequest, actorID uuid.UUID, ip string) (*dto.SectionResponse, error) {
	rec, err := s.repos.Academic.UpdateSection(ctx, id, req.ClassID, req.Name, req.Capacity)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, ErrNotFound
	}
	resp := mapSection(rec)
	s.audit.Log(ctx, &actorID, model.ActionUpdate, model.EntitySection, &id, ip, map[string]any{"name": rec.Name})
	return &resp, nil
}

func (s *AcademicService) DeleteSection(ctx context.Context, id uuid.UUID, actorID uuid.UUID, ip string) error {
	rec, err := s.repos.Academic.GetSectionByID(ctx, id)
	if err != nil {
		return err
	}
	if rec == nil {
		return ErrNotFound
	}
	if err := s.repos.Academic.SoftDeleteSection(ctx, id); err != nil {
		return err
	}
	s.audit.Log(ctx, &actorID, model.ActionDelete, model.EntitySection, &id, ip, map[string]any{"name": rec.Name})
	return nil
}

func (s *AcademicService) ListSections(ctx context.Context) ([]dto.SectionResponse, error) {
	recs, err := s.repos.Academic.ListSections(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]dto.SectionResponse, 0, len(recs))
	for i := range recs {
		items = append(items, mapSection(&recs[i]))
	}
	return items, nil
}

func (s *AcademicService) ListSectionsByClass(ctx context.Context, classID uuid.UUID) ([]dto.SectionResponse, error) {
	recs, err := s.repos.Academic.ListSectionsByClass(ctx, classID)
	if err != nil {
		return nil, err
	}
	items := make([]dto.SectionResponse, 0, len(recs))
	for i := range recs {
		items = append(items, mapSection(&recs[i]))
	}
	return items, nil
}

func (s *AcademicService) CreateSubject(ctx context.Context, req dto.SubjectRequest, actorID uuid.UUID, ip string) (*dto.SubjectResponse, error) {
	rec, err := s.repos.Academic.CreateSubject(ctx, req.Name, strings.ToUpper(req.Code), req.Description)
	if err != nil {
		return nil, err
	}
	resp := mapSubject(rec)
	s.audit.Log(ctx, &actorID, model.ActionCreate, model.EntitySubject, &rec.ID, ip, map[string]any{"name": rec.Name})
	return &resp, nil
}

func (s *AcademicService) UpdateSubject(ctx context.Context, id uuid.UUID, req dto.SubjectRequest, actorID uuid.UUID, ip string) (*dto.SubjectResponse, error) {
	rec, err := s.repos.Academic.UpdateSubject(ctx, id, req.Name, strings.ToUpper(req.Code), req.Description)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, ErrNotFound
	}
	resp := mapSubject(rec)
	s.audit.Log(ctx, &actorID, model.ActionUpdate, model.EntitySubject, &id, ip, map[string]any{"name": rec.Name})
	return &resp, nil
}

func (s *AcademicService) DeleteSubject(ctx context.Context, id uuid.UUID, actorID uuid.UUID, ip string) error {
	rec, err := s.repos.Academic.GetSubjectByID(ctx, id)
	if err != nil {
		return err
	}
	if rec == nil {
		return ErrNotFound
	}
	if err := s.repos.Academic.SoftDeleteSubject(ctx, id); err != nil {
		return err
	}
	s.audit.Log(ctx, &actorID, model.ActionDelete, model.EntitySubject, &id, ip, map[string]any{"name": rec.Name})
	return nil
}

func (s *AcademicService) ListSubjects(ctx context.Context) ([]dto.SubjectResponse, error) {
	recs, err := s.repos.Academic.ListSubjects(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]dto.SubjectResponse, 0, len(recs))
	for i := range recs {
		items = append(items, mapSubject(&recs[i]))
	}
	return items, nil
}

func (s *AcademicService) AssignSubjectsToClass(ctx context.Context, classID uuid.UUID, subjectIDs []uuid.UUID, actorID uuid.UUID, ip string) error {
	if _, err := s.repos.Academic.GetClassByID(ctx, classID); err != nil {
		return err
	}
	if err := s.repos.Academic.ClearClassSubjects(ctx, classID); err != nil {
		return err
	}
	for _, sid := range subjectIDs {
		if err := s.repos.Academic.AssignSubjectToClass(ctx, classID, sid); err != nil {
			return err
		}
	}
	s.audit.Log(ctx, &actorID, model.ActionUpdate, model.EntityClass, &classID, ip, map[string]any{"subjects": len(subjectIDs)})
	return nil
}

func (s *AcademicService) ListSubjectsByClass(ctx context.Context, classID uuid.UUID) ([]dto.SubjectResponse, error) {
	recs, err := s.repos.Academic.ListSubjectsByClass(ctx, classID)
	if err != nil {
		return nil, err
	}
	items := make([]dto.SubjectResponse, 0, len(recs))
	for i := range recs {
		items = append(items, mapSubject(&recs[i]))
	}
	return items, nil
}

func mapClass(c *repository.ClassRecord) dto.ClassResponse {
	return dto.ClassResponse{ID: c.ID, Name: c.Name, Code: c.Code, Description: c.Description, SortOrder: c.SortOrder}
}

func mapSection(s *repository.SectionRecord) dto.SectionResponse {
	return dto.SectionResponse{ID: s.ID, ClassID: s.ClassID, ClassName: s.ClassName, Name: s.Name, Capacity: s.Capacity}
}

func mapSubject(s *repository.SubjectRecord) dto.SubjectResponse {
	return dto.SubjectResponse{ID: s.ID, Name: s.Name, Code: s.Code, Description: s.Description}
}
