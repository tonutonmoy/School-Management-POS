package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/school-management/pos/internal/dto"
	"github.com/school-management/pos/internal/model"
	"github.com/school-management/pos/internal/repository"
)

type HRService struct {
	repos *repository.Repositories
	audit *AuditService
}

func NewHRService(repos *repository.Repositories, audit *AuditService) *HRService {
	return &HRService{repos: repos, audit: audit}
}

// Departments
func (s *HRService) CreateDepartment(ctx context.Context, req dto.DepartmentRequest, actorID uuid.UUID, ip string) (*dto.DepartmentResponse, error) {
	deptType := req.DeptType
	if deptType == "" {
		deptType = model.DeptTypeEmployee
	}
	rec, err := s.repos.HR.CreateDepartment(ctx, req.Name, strings.ToLower(req.Slug), req.Description, deptType)
	if err != nil {
		return nil, err
	}
	resp := mapDepartment(rec)
	s.audit.Log(ctx, &actorID, model.ActionCreate, model.EntityDepartment, &rec.ID, ip, map[string]any{"name": rec.Name})
	return &resp, nil
}

func (s *HRService) UpdateDepartment(ctx context.Context, id uuid.UUID, req dto.DepartmentRequest, actorID uuid.UUID, ip string) (*dto.DepartmentResponse, error) {
	rec, err := s.repos.HR.UpdateDepartment(ctx, id, req.Name, strings.ToLower(req.Slug), req.Description, req.DeptType)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, ErrNotFound
	}
	resp := mapDepartment(rec)
	s.audit.Log(ctx, &actorID, model.ActionUpdate, model.EntityDepartment, &id, ip, nil)
	return &resp, nil
}

func (s *HRService) DeleteDepartment(ctx context.Context, id uuid.UUID, actorID uuid.UUID, ip string) error {
	if err := s.repos.HR.SoftDeleteDepartment(ctx, id); err != nil {
		return err
	}
	s.audit.Log(ctx, &actorID, model.ActionDelete, model.EntityDepartment, &id, ip, nil)
	return nil
}

func (s *HRService) ListDepartments(ctx context.Context, deptType string) ([]dto.DepartmentResponse, error) {
	recs, err := s.repos.HR.ListDepartments(ctx, deptType)
	if err != nil {
		return nil, err
	}
	items := make([]dto.DepartmentResponse, 0, len(recs))
	for i := range recs {
		items = append(items, mapDepartment(&recs[i]))
	}
	return items, nil
}

func (s *HRService) GetDepartment(ctx context.Context, id uuid.UUID) (*dto.DepartmentResponse, error) {
	rec, err := s.repos.HR.GetDepartmentByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, ErrNotFound
	}
	resp := mapDepartment(rec)
	return &resp, nil
}

func mapDepartment(d *repository.HRDepartmentRecord) dto.DepartmentResponse {
	return dto.DepartmentResponse{
		ID: d.ID, Name: d.Name, Slug: d.Slug, Description: d.Description,
		DeptType: d.DeptType, TeacherCount: d.TeacherCount, StaffCount: d.StaffCount,
	}
}

// Designations
func (s *HRService) CreateDesignation(ctx context.Context, req dto.DesignationRequest, actorID uuid.UUID, ip string) (*dto.DesignationResponse, error) {
	rec, err := s.repos.HR.CreateDesignation(ctx, req.Name, strings.ToLower(req.Slug), req.Category, req.Description)
	if err != nil {
		return nil, err
	}
	resp := mapDesignation(rec)
	s.audit.Log(ctx, &actorID, model.ActionCreate, model.EntityDesignation, &rec.ID, ip, nil)
	return &resp, nil
}

func (s *HRService) UpdateDesignation(ctx context.Context, id uuid.UUID, req dto.DesignationRequest, actorID uuid.UUID, ip string) (*dto.DesignationResponse, error) {
	rec, err := s.repos.HR.UpdateDesignation(ctx, id, req.Name, strings.ToLower(req.Slug), req.Category, req.Description)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, ErrNotFound
	}
	resp := mapDesignation(rec)
	s.audit.Log(ctx, &actorID, model.ActionUpdate, model.EntityDesignation, &id, ip, nil)
	return &resp, nil
}

func (s *HRService) DeleteDesignation(ctx context.Context, id uuid.UUID, actorID uuid.UUID, ip string) error {
	if err := s.repos.HR.SoftDeleteDesignation(ctx, id); err != nil {
		return err
	}
	s.audit.Log(ctx, &actorID, model.ActionDelete, model.EntityDesignation, &id, ip, nil)
	return nil
}

func (s *HRService) ListDesignations(ctx context.Context) ([]dto.DesignationResponse, error) {
	recs, err := s.repos.HR.ListDesignations(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]dto.DesignationResponse, 0, len(recs))
	for i := range recs {
		items = append(items, mapDesignation(&recs[i]))
	}
	return items, nil
}

func mapDesignation(d *repository.DesignationRecord) dto.DesignationResponse {
	return dto.DesignationResponse{ID: d.ID, Name: d.Name, Slug: d.Slug, Category: d.Category, Description: d.Description}
}

// Teachers
func (s *HRService) CreateTeacher(ctx context.Context, req dto.TeacherRequest, photoURL string, actorID uuid.UUID, ip string) (*dto.TeacherResponse, error) {
	empID, err := s.repos.HR.NextEmployeeID(ctx, model.EntityTeacherSeq, req.JoiningDate.Year())
	if err != nil {
		return nil, err
	}
	status := req.Status
	if status == "" {
		status = model.TeacherStatusActive
	}
	p := s.teacherParams(req, photoURL, empID, status)
	rec, err := s.repos.HR.CreateTeacher(ctx, p)
	if err != nil {
		return nil, err
	}
	resp := s.mapTeacher(rec, nil, nil)
	s.audit.Log(ctx, &actorID, model.ActionCreate, model.EntityTeacher, &rec.ID, ip, map[string]any{"employee_id": empID})
	return &resp, nil
}

func (s *HRService) UpdateTeacher(ctx context.Context, id uuid.UUID, req dto.TeacherRequest, photoURL string, actorID uuid.UUID, ip string) (*dto.TeacherResponse, error) {
	existing, err := s.repos.HR.GetTeacherByID(ctx, id)
	if err != nil || existing == nil {
		return nil, ErrNotFound
	}
	status := req.Status
	if status == "" {
		status = model.TeacherStatusActive
	}
	p := s.teacherParams(req, photoURL, existing.EmployeeID, status)
	rec, err := s.repos.HR.UpdateTeacher(ctx, id, p)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, ErrNotFound
	}
	resp, _ := s.GetTeacher(ctx, id)
	s.audit.Log(ctx, &actorID, model.ActionUpdate, model.EntityTeacher, &id, ip, nil)
	return resp, nil
}

func (s *HRService) DeleteTeacher(ctx context.Context, id uuid.UUID, actorID uuid.UUID, ip string) error {
	if err := s.repos.HR.SoftDeleteTeacher(ctx, id); err != nil {
		return err
	}
	s.audit.Log(ctx, &actorID, model.ActionDelete, model.EntityTeacher, &id, ip, nil)
	return nil
}

func (s *HRService) GetTeacher(ctx context.Context, id uuid.UUID) (*dto.TeacherResponse, error) {
	rec, err := s.repos.HR.GetTeacherByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, ErrNotFound
	}
	assigns, _ := s.repos.HR.ListTeacherAssignments(ctx, id)
	docs, _ := s.repos.HR.ListTeacherDocuments(ctx, id)
	resp := s.mapTeacher(rec, assigns, docs)
	return &resp, nil
}

func (s *HRService) SearchTeachers(ctx context.Context, filter dto.TeacherSearchFilter) (*dto.PaginatedTeachers, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 20
	}
	params := repository.TeacherSearchParams{
		Query: filter.Query, DepartmentID: filter.DepartmentID, DesignationID: filter.DesignationID,
		Status: filter.Status, Limit: int32(filter.PageSize), Offset: int32((filter.Page - 1) * filter.PageSize),
	}
	total, err := s.repos.HR.CountTeachers(ctx, params)
	if err != nil {
		return nil, err
	}
	recs, err := s.repos.HR.SearchTeachers(ctx, params)
	if err != nil {
		return nil, err
	}
	items := make([]dto.TeacherResponse, 0, len(recs))
	for i := range recs {
		items = append(items, s.mapTeacher(&recs[i], nil, nil))
	}
	tp := int(total) / filter.PageSize
	if int(total)%filter.PageSize > 0 {
		tp++
	}
	return &dto.PaginatedTeachers{Items: items, Total: total, Page: filter.Page, PageSize: filter.PageSize, TotalPages: tp}, nil
}

func (s *HRService) AssignTeacher(ctx context.Context, teacherID uuid.UUID, req dto.TeacherAssignmentRequest, actorID uuid.UUID, ip string) error {
	if err := s.repos.HR.ClearTeacherAssignments(ctx, teacherID); err != nil {
		return err
	}
	for _, sid := range req.SubjectIDs {
		id := sid
		if err := s.repos.HR.AddTeacherAssignment(ctx, teacherID, &id, nil, nil); err != nil {
			return err
		}
	}
	for _, cid := range req.ClassIDs {
		id := cid
		if err := s.repos.HR.AddTeacherAssignment(ctx, teacherID, nil, &id, nil); err != nil {
			return err
		}
	}
	for _, secID := range req.SectionIDs {
		id := secID
		if err := s.repos.HR.AddTeacherAssignment(ctx, teacherID, nil, nil, &id); err != nil {
			return err
		}
	}
	s.audit.Log(ctx, &actorID, model.ActionUpdate, model.EntityTeacher, &teacherID, ip, map[string]any{"action": "assign"})
	return nil
}

func (s *HRService) AddTeacherDocument(ctx context.Context, teacherID uuid.UUID, docType, fileName, fileURL string, actorID uuid.UUID, ip string) error {
	if err := s.repos.HR.CreateTeacherDocument(ctx, teacherID, docType, fileName, fileURL); err != nil {
		return err
	}
	s.audit.Log(ctx, &actorID, model.ActionCreate, model.EntityTeacher, &teacherID, ip, map[string]any{"doc": docType})
	return nil
}

func (s *HRService) TeacherPortal(ctx context.Context, userID uuid.UUID, userEmail string) (*dto.TeacherPortalDashboard, error) {
	rec, _ := s.repos.HR.GetTeacherByUserID(ctx, userID)
	if rec == nil && userEmail != "" {
		rec, _ = s.repos.HR.GetTeacherByEmail(ctx, userEmail)
	}
	if rec == nil {
		return nil, ErrNotFound
	}
	classes, _ := s.repos.HR.CountTeacherClasses(ctx, rec.ID)
	subjects, _ := s.repos.HR.CountTeacherSubjects(ctx, rec.ID)
	day := int(time.Now().Weekday())
	schedule, _ := s.repos.HR.ListTodaySchedule(ctx, rec.ID, day)
	assigns, _ := s.repos.HR.ListTeacherAssignments(ctx, rec.ID)
	docs, _ := s.repos.HR.ListTeacherDocuments(ctx, rec.ID)
	teacher := s.mapTeacher(rec, assigns, docs)
	schedResp := make([]dto.TeacherScheduleResponse, 0, len(schedule))
	for _, sc := range schedule {
		schedResp = append(schedResp, dto.TeacherScheduleResponse{
			ID: sc.ID, SubjectName: sc.SubjectName, ClassName: sc.ClassName, SectionName: sc.SectionName,
			StartTime: sc.StartTime, EndTime: sc.EndTime, Room: sc.Room,
		})
	}
	return &dto.TeacherPortalDashboard{
		Teacher: teacher, AssignedClasses: int(classes), AssignedSubjects: int(subjects), TodaySchedule: schedResp,
	}, nil
}

func (s *HRService) ListTeachersReport(ctx context.Context, filter dto.HRReportFilter) ([]dto.TeacherResponse, error) {
	params := repository.TeacherSearchParams{DepartmentID: filter.DepartmentID, DesignationID: filter.DesignationID, Status: filter.Status}
	recs, err := s.repos.HR.ListTeachersReport(ctx, params)
	if err != nil {
		return nil, err
	}
	items := make([]dto.TeacherResponse, 0, len(recs))
	for i := range recs {
		assigns, _ := s.repos.HR.ListTeacherAssignments(ctx, recs[i].ID)
		items = append(items, s.mapTeacher(&recs[i], assigns, nil))
	}
	return items, nil
}

func (s *HRService) teacherParams(req dto.TeacherRequest, photoURL, empID, status string) repository.CreateTeacherParams {
	var deptID, desID *uuid.UUID
	if req.DepartmentID != uuid.Nil {
		deptID = &req.DepartmentID
	}
	if req.DesignationID != uuid.Nil {
		desID = &req.DesignationID
	}
	var dob *time.Time
	if !req.DateOfBirth.IsZero() {
		dob = &req.DateOfBirth
	}
	nationality := req.Nationality
	if nationality == "" {
		nationality = "Bangladeshi"
	}
	return repository.CreateTeacherParams{
		EmployeeID: empID, FirstName: req.FirstName, LastName: req.LastName, PhotoURL: photoURL,
		Gender: req.Gender, DateOfBirth: dob, BloodGroup: req.BloodGroup, Religion: req.Religion,
		Nationality: nationality, Phone: req.Phone, Email: req.Email, Address: req.Address,
		NationalID: req.NationalID, JoiningDate: req.JoiningDate, DepartmentID: deptID, DesignationID: desID,
		Qualification: req.Qualification, Experience: req.Experience, Salary: req.Salary,
		EmploymentType: req.EmploymentType, Status: status,
	}
}

func (s *HRService) mapTeacher(t *repository.TeacherRecord, assigns []repository.TeacherAssignmentRecord, docs []repository.DocumentRecord) dto.TeacherResponse {
	resp := dto.TeacherResponse{
		ID: t.ID, EmployeeID: t.EmployeeID, FirstName: t.FirstName, LastName: t.LastName,
		FullName: strings.TrimSpace(t.FirstName + " " + t.LastName), PhotoURL: t.PhotoURL,
		Gender: t.Gender, DateOfBirth: t.DateOfBirth, BloodGroup: t.BloodGroup, Religion: t.Religion,
		Nationality: t.Nationality, Phone: t.Phone, Email: t.Email, Address: t.Address,
		NationalID: t.NationalID, JoiningDate: t.JoiningDate, DepartmentID: t.DepartmentID,
		DepartmentName: t.DepartmentName, DesignationID: t.DesignationID, DesignationName: t.DesignationName,
		Qualification: t.Qualification, Experience: t.Experience, Salary: t.Salary,
		EmploymentType: t.EmploymentType, Status: t.Status, CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt,
	}
	for _, a := range assigns {
		resp.Assignments = append(resp.Assignments, dto.TeacherAssignmentResponse{
			ID: a.ID, SubjectID: a.SubjectID, SubjectName: a.SubjectName,
			ClassID: a.ClassID, ClassName: a.ClassName, SectionID: a.SectionID, SectionName: a.SectionName,
		})
	}
	for _, d := range docs {
		resp.Documents = append(resp.Documents, dto.EmployeeDocumentResponse{ID: d.ID, DocType: d.DocType, FileName: d.FileName, FileURL: d.FileURL})
	}
	return resp
}

// Staff
func (s *HRService) CreateStaff(ctx context.Context, req dto.StaffRequest, photoURL string, actorID uuid.UUID, ip string) (*dto.StaffResponse, error) {
	empID, err := s.repos.HR.NextEmployeeID(ctx, model.EntityStaffSeq, req.JoiningDate.Year())
	if err != nil {
		return nil, err
	}
	status := req.Status
	if status == "" {
		status = model.TeacherStatusActive
	}
	p := s.staffParams(req, photoURL, empID, status)
	rec, err := s.repos.HR.CreateStaff(ctx, p)
	if err != nil {
		return nil, err
	}
	resp := s.mapStaff(rec, nil)
	s.audit.Log(ctx, &actorID, model.ActionCreate, model.EntityStaff, &rec.ID, ip, map[string]any{"employee_id": empID})
	return &resp, nil
}

func (s *HRService) UpdateStaff(ctx context.Context, id uuid.UUID, req dto.StaffRequest, photoURL string, actorID uuid.UUID, ip string) (*dto.StaffResponse, error) {
	existing, err := s.repos.HR.GetStaffByID(ctx, id)
	if err != nil || existing == nil {
		return nil, ErrNotFound
	}
	status := req.Status
	if status == "" {
		status = model.TeacherStatusActive
	}
	p := s.staffParams(req, photoURL, existing.EmployeeID, status)
	rec, err := s.repos.HR.UpdateStaff(ctx, id, p)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, ErrNotFound
	}
	return s.GetStaff(ctx, id)
}

func (s *HRService) DeleteStaff(ctx context.Context, id uuid.UUID, actorID uuid.UUID, ip string) error {
	if err := s.repos.HR.SoftDeleteStaff(ctx, id); err != nil {
		return err
	}
	s.audit.Log(ctx, &actorID, model.ActionDelete, model.EntityStaff, &id, ip, nil)
	return nil
}

func (s *HRService) GetStaff(ctx context.Context, id uuid.UUID) (*dto.StaffResponse, error) {
	rec, err := s.repos.HR.GetStaffByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, ErrNotFound
	}
	docs, _ := s.repos.HR.ListStaffDocuments(ctx, id)
	resp := s.mapStaff(rec, docs)
	return &resp, nil
}

func (s *HRService) SearchStaff(ctx context.Context, filter dto.StaffSearchFilter) (*dto.PaginatedStaff, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 20
	}
	params := repository.StaffSearchParams{
		Query: filter.Query, DepartmentID: filter.DepartmentID, Status: filter.Status,
		Limit: int32(filter.PageSize), Offset: int32((filter.Page - 1) * filter.PageSize),
	}
	total, err := s.repos.HR.CountStaff(ctx, params)
	if err != nil {
		return nil, err
	}
	recs, err := s.repos.HR.SearchStaff(ctx, params)
	if err != nil {
		return nil, err
	}
	items := make([]dto.StaffResponse, 0, len(recs))
	for i := range recs {
		items = append(items, s.mapStaff(&recs[i], nil))
	}
	tp := int(total) / filter.PageSize
	if int(total)%filter.PageSize > 0 {
		tp++
	}
	return &dto.PaginatedStaff{Items: items, Total: total, Page: filter.Page, PageSize: filter.PageSize, TotalPages: tp}, nil
}

func (s *HRService) AddStaffDocument(ctx context.Context, staffID uuid.UUID, docType, fileName, fileURL string, actorID uuid.UUID, ip string) error {
	return s.repos.HR.CreateStaffDocument(ctx, staffID, docType, fileName, fileURL)
}

func (s *HRService) ListStaffReport(ctx context.Context, filter dto.HRReportFilter) ([]dto.StaffResponse, error) {
	params := repository.StaffSearchParams{DepartmentID: filter.DepartmentID, Status: filter.Status}
	recs, err := s.repos.HR.ListStaffReport(ctx, params)
	if err != nil {
		return nil, err
	}
	items := make([]dto.StaffResponse, 0, len(recs))
	for i := range recs {
		items = append(items, s.mapStaff(&recs[i], nil))
	}
	return items, nil
}

func (s *HRService) staffParams(req dto.StaffRequest, photoURL, empID, status string) repository.CreateStaffParams {
	var deptID, desID *uuid.UUID
	if req.DepartmentID != uuid.Nil {
		deptID = &req.DepartmentID
	}
	if req.DesignationID != uuid.Nil {
		desID = &req.DesignationID
	}
	return repository.CreateStaffParams{
		EmployeeID: empID, FirstName: req.FirstName, LastName: req.LastName, PhotoURL: photoURL,
		Phone: req.Phone, Email: req.Email, Address: req.Address, DepartmentID: deptID,
		DesignationID: desID, Salary: req.Salary, JoiningDate: req.JoiningDate, Status: status,
	}
}

func (s *HRService) mapStaff(st *repository.StaffRecord, docs []repository.DocumentRecord) dto.StaffResponse {
	resp := dto.StaffResponse{
		ID: st.ID, EmployeeID: st.EmployeeID, FirstName: st.FirstName, LastName: st.LastName,
		FullName: strings.TrimSpace(st.FirstName + " " + st.LastName), PhotoURL: st.PhotoURL,
		Phone: st.Phone, Email: st.Email, Address: st.Address, DepartmentID: st.DepartmentID,
		DepartmentName: st.DepartmentName, DesignationID: st.DesignationID, DesignationName: st.DesignationName,
		Salary: st.Salary, JoiningDate: st.JoiningDate, Status: st.Status, CreatedAt: st.CreatedAt, UpdatedAt: st.UpdatedAt,
	}
	for _, d := range docs {
		resp.Documents = append(resp.Documents, dto.EmployeeDocumentResponse{ID: d.ID, DocType: d.DocType, FileName: d.FileName, FileURL: d.FileURL})
	}
	return resp
}
