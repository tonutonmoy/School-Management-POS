package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/school-management/pos/internal/dto"
	"github.com/school-management/pos/internal/model"
	"github.com/school-management/pos/internal/repository"
)

type ExamService struct {
	repos  *repository.Repositories
	audit  *AuditService
	notify *NotificationService
}

func NewExamService(repos *repository.Repositories, audit *AuditService) *ExamService {
	return &ExamService{repos: repos, audit: audit}
}

func (s *ExamService) SetNotifier(n *NotificationService) { s.notify = n }

// --- Exams ---

func (s *ExamService) CreateExam(ctx context.Context, req dto.ExamRequest, actorID uuid.UUID, ip string) (*dto.ExamResponse, error) {
	gsID := req.GradingSystemID
	if gsID == uuid.Nil {
		if def, _ := s.repos.Exams.GetDefaultGradingSystem(ctx); def != nil {
			gsID = def.ID
		}
	}
	rec, err := s.repos.Exams.CreateExam(ctx, repository.CreateExamParams{
		Name: req.Name, ExamType: req.ExamType, SessionID: req.SessionID, ClassID: req.ClassID,
		StartDate: req.StartDate, EndDate: req.EndDate, TotalMarks: req.TotalMarks, PassingMarks: req.PassingMarks,
		GradingSystemID: &gsID, Status: model.ExamStatusDraft,
	})
	if err != nil {
		return nil, err
	}
	resp := mapExam(rec)
	s.audit.Log(ctx, &actorID, model.ActionCreate, model.EntityExam, &rec.ID, ip, nil)
	return &resp, nil
}

func (s *ExamService) UpdateExam(ctx context.Context, id uuid.UUID, req dto.ExamRequest, actorID uuid.UUID, ip string) (*dto.ExamResponse, error) {
	gsID := req.GradingSystemID
	rec, err := s.repos.Exams.UpdateExam(ctx, id, repository.CreateExamParams{
		Name: req.Name, ExamType: req.ExamType, SessionID: req.SessionID, ClassID: req.ClassID,
		StartDate: req.StartDate, EndDate: req.EndDate, TotalMarks: req.TotalMarks, PassingMarks: req.PassingMarks,
		GradingSystemID: &gsID,
	})
	if err != nil || rec == nil {
		return nil, ErrNotFound
	}
	resp := mapExam(rec)
	s.audit.Log(ctx, &actorID, model.ActionUpdate, model.EntityExam, &id, ip, nil)
	return &resp, nil
}

func (s *ExamService) DeleteExam(ctx context.Context, id uuid.UUID, actorID uuid.UUID, ip string) error {
	if err := s.repos.Exams.SoftDeleteExam(ctx, id); err != nil {
		return err
	}
	s.audit.Log(ctx, &actorID, model.ActionDelete, model.EntityExam, &id, ip, nil)
	return nil
}

func (s *ExamService) GetExam(ctx context.Context, id uuid.UUID) (*dto.ExamResponse, error) {
	rec, err := s.repos.Exams.GetExam(ctx, id)
	if err != nil || rec == nil {
		return nil, ErrNotFound
	}
	resp := mapExam(rec)
	return &resp, nil
}

func (s *ExamService) ListExams(ctx context.Context, f dto.ExamSearchFilter) (*dto.PaginatedExams, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PerPage < 1 {
		f.PerPage = 20
	}
	params := mapExamFilter(f)
	total, _ := s.repos.Exams.CountExams(ctx, params)
	recs, err := s.repos.Exams.SearchExams(ctx, params)
	if err != nil {
		return nil, err
	}
	items := make([]dto.ExamResponse, 0, len(recs))
	for _, r := range recs {
		items = append(items, mapExam(&r))
	}
	return &dto.PaginatedExams{
		Items: items, Total: total, Page: f.Page, PerPage: f.PerPage,
		TotalPages: int(math.Ceil(float64(total) / float64(f.PerPage))),
	}, nil
}

func (s *ExamService) PublishExam(ctx context.Context, id uuid.UUID, actorID uuid.UUID, ip string) error {
	if err := s.repos.Exams.UpdateExamStatus(ctx, id, model.ExamStatusPublished); err != nil {
		return err
	}
	s.audit.Log(ctx, &actorID, model.ActionUpdate, model.EntityExam, &id, ip, map[string]any{"status": "published"})
	return nil
}

func (s *ExamService) ArchiveExam(ctx context.Context, id uuid.UUID, actorID uuid.UUID, ip string) error {
	if err := s.repos.Exams.UpdateExamStatus(ctx, id, model.ExamStatusArchived); err != nil {
		return err
	}
	s.audit.Log(ctx, &actorID, model.ActionUpdate, model.EntityExam, &id, ip, map[string]any{"status": "archived"})
	return nil
}

// --- Exam Subjects ---

func (s *ExamService) AddExamSubject(ctx context.Context, examID uuid.UUID, req dto.ExamSubjectRequest, actorID uuid.UUID, ip string) (*dto.ExamSubjectResponse, error) {
	if req.WrittenMarks+req.MCQMarks+req.PracticalMarks > req.FullMarks {
		return nil, fmt.Errorf("%w: component marks exceed full marks", ErrValidation)
	}
	rec, err := s.repos.Exams.CreateExamSubject(ctx, examID, repository.ExamSubjectParams{
		SubjectID: req.SubjectID, FullMarks: req.FullMarks, PassMarks: req.PassMarks,
		WrittenMarks: req.WrittenMarks, MCQMarks: req.MCQMarks, PracticalMarks: req.PracticalMarks,
	})
	if err != nil {
		return nil, err
	}
	resp := mapExamSubject(rec)
	s.audit.Log(ctx, &actorID, model.ActionCreate, model.EntityExamSubject, &rec.ID, ip, nil)
	return &resp, nil
}

func (s *ExamService) ListExamSubjects(ctx context.Context, examID uuid.UUID) ([]dto.ExamSubjectResponse, error) {
	recs, err := s.repos.Exams.ListExamSubjects(ctx, examID)
	if err != nil {
		return nil, err
	}
	items := make([]dto.ExamSubjectResponse, 0, len(recs))
	for _, r := range recs {
		items = append(items, mapExamSubject(&r))
	}
	return items, nil
}

// --- Grading ---

func (s *ExamService) ListGradingSystems(ctx context.Context) ([]dto.GradingSystemResponse, error) {
	systems, err := s.repos.Exams.ListGradingSystems(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]dto.GradingSystemResponse, 0, len(systems))
	for _, sys := range systems {
		scales, _ := s.repos.Exams.ListGradingScales(ctx, sys.ID)
		items = append(items, mapGradingSystem(&sys, scales))
	}
	return items, nil
}

func (s *ExamService) CreateGradingSystem(ctx context.Context, name string, scales []dto.GradingScaleRequest, actorID uuid.UUID, ip string) (*dto.GradingSystemResponse, error) {
	params := make([]repository.GradingScaleParams, 0, len(scales))
	for i, sc := range scales {
		params = append(params, repository.GradingScaleParams{
			Grade: sc.Grade, MinPercentage: sc.MinPercentage, MaxPercentage: sc.MaxPercentage,
			GPAPoint: sc.GPAPoint, SortOrder: i + 1,
		})
	}
	rec, err := s.repos.Exams.CreateGradingSystem(ctx, name, params)
	if err != nil {
		return nil, err
	}
	gs := mapGradingSystem(rec, paramsToScaleRecords(params))
	s.audit.Log(ctx, &actorID, model.ActionCreate, model.EntityGradingSystem, &rec.ID, ip, nil)
	return &gs, nil
}

func paramsToScaleRecords(params []repository.GradingScaleParams) []repository.GradingScaleRecord {
	recs := make([]repository.GradingScaleRecord, len(params))
	for i, p := range params {
		recs[i] = repository.GradingScaleRecord{Grade: p.Grade, MinPercentage: p.MinPercentage, MaxPercentage: p.MaxPercentage, GPAPoint: p.GPAPoint}
	}
	return recs
}

// --- Marks Entry ---

func (s *ExamService) MarksSheet(ctx context.Context, examSubjectID uuid.UUID) ([]dto.MarkEntryRow, error) {
	recs, err := s.repos.Exams.ListMarksSheet(ctx, examSubjectID)
	if err != nil {
		return nil, err
	}
	items := make([]dto.MarkEntryRow, 0, len(recs))
	for _, r := range recs {
		items = append(items, dto.MarkEntryRow{
			StudentID: r.StudentID, StudentName: r.StudentName, RollNumber: r.RollNumber, AdmissionNo: r.AdmissionNo,
			WrittenScore: r.WrittenScore, MCQScore: r.MCQScore, PracticalScore: r.PracticalScore,
			TotalScore: r.TotalScore, IsAbsent: r.IsAbsent, RecordID: r.RecordID,
		})
	}
	return items, nil
}

func (s *ExamService) BulkSaveMarks(ctx context.Context, examID, examSubjectID uuid.UUID, entries []dto.StudentMarkEntry, actorID uuid.UUID, ip string) error {
	subj, err := s.repos.Exams.GetExamSubject(ctx, examSubjectID)
	if err != nil || subj == nil {
		return ErrNotFound
	}
	for _, e := range entries {
		if e.IsAbsent {
			_ = s.repos.Exams.UpsertStudentMark(ctx, repository.UpsertMarkParams{
				ExamID: examID, ExamSubjectID: examSubjectID, StudentID: e.StudentID, EnteredBy: actorID, IsAbsent: true,
			})
			continue
		}
		if e.WrittenScore > subj.WrittenMarks || e.MCQScore > subj.MCQMarks || e.PracticalScore > subj.PracticalMarks {
			return fmt.Errorf("%w: marks exceed configured maximum for student %s", ErrValidation, e.StudentID)
		}
		total := e.WrittenScore + e.MCQScore + e.PracticalScore
		if total > subj.FullMarks {
			return fmt.Errorf("%w: total exceeds full marks for student %s", ErrValidation, e.StudentID)
		}
		if err := s.repos.Exams.UpsertStudentMark(ctx, repository.UpsertMarkParams{
			ExamID: examID, ExamSubjectID: examSubjectID, StudentID: e.StudentID, EnteredBy: actorID,
			WrittenScore: e.WrittenScore, MCQScore: e.MCQScore, PracticalScore: e.PracticalScore, TotalScore: total,
		}); err != nil {
			return err
		}
	}
	s.audit.Log(ctx, &actorID, model.ActionUpdate, model.EntityStudentMark, nil, ip, map[string]any{
		"exam_id": examID, "subject_id": examSubjectID, "count": len(entries),
	})
	return nil
}

// --- Result Processing ---

func (s *ExamService) ProcessResults(ctx context.Context, examID uuid.UUID, actorID uuid.UUID, ip string) (int, error) {
	exam, err := s.repos.Exams.GetExam(ctx, examID)
	if err != nil || exam == nil {
		return 0, ErrNotFound
	}
	subjects, _ := s.repos.Exams.ListExamSubjects(ctx, examID)
	if len(subjects) == 0 {
		return 0, fmt.Errorf("%w: no subjects configured", ErrValidation)
	}
	systemID := exam.GradingSystemID
	if systemID == nil {
		if def, _ := s.repos.Exams.GetDefaultGradingSystem(ctx); def != nil {
			systemID = &def.ID
		}
	}
	scales, _ := s.repos.Exams.ListGradingScales(ctx, *systemID)
	students, _ := s.repos.Students.ListForReport(ctx, &exam.ClassID, &exam.SessionID, model.StudentStatusActive)
	allMarks, _ := s.repos.Exams.ListMarksByExam(ctx, examID)
	markIndex := map[uuid.UUID]map[uuid.UUID]repository.StudentMarkRecord{}
	for _, m := range allMarks {
		if markIndex[m.ExamSubjectID] == nil {
			markIndex[m.ExamSubjectID] = map[uuid.UUID]repository.StudentMarkRecord{}
		}
		markIndex[m.ExamSubjectID][m.StudentID] = m
	}
	count := 0
	err = s.repos.Exams.WithTx(ctx, func(tx pgx.Tx) error {
		_ = s.repos.Exams.DeleteExamResults(ctx, examID)
		for _, st := range students {
			result := s.calcStudentResult(ctx, exam, subjects, scales, &st, markIndex)
			if err := s.repos.Exams.UpsertExamResult(ctx, tx, result); err != nil {
				return err
			}
			count++
		}
		return s.repos.Exams.UpdateResultPositions(ctx, tx, examID)
	})
	if err != nil {
		return 0, err
	}
	s.audit.Log(ctx, &actorID, model.ActionUpdate, model.EntityExamResult, nil, ip, map[string]any{"exam_id": examID, "count": count})
	return count, nil
}

func (s *ExamService) calcStudentResult(ctx context.Context, exam *repository.ExamRecord, subjects []repository.ExamSubjectRecord, scales []repository.GradingScaleRecord, st *repository.StudentRecord, markIndex map[uuid.UUID]map[uuid.UUID]repository.StudentMarkRecord) repository.ExamResultParams {
	var totalObtained, totalFull float64
	allPassed := true
	for _, subj := range subjects {
		totalFull += subj.FullMarks
		subjMarks := markIndex[subj.ID]
		m, ok := subjMarks[st.ID]
		if !ok || m.IsAbsent {
			allPassed = false
			continue
		}
		totalObtained += m.TotalScore
		if m.TotalScore < subj.PassMarks {
			allPassed = false
		}
	}
	pct := 0.0
	if totalFull > 0 {
		pct = totalObtained / totalFull * 100
	}
	grade, gpa := gradeFromPercentage(scales, pct)
	overallPassed := allPassed && totalObtained >= exam.PassingMarks
	cgpa, _ := s.repos.Exams.StudentCGPA(ctx, st.ID, exam.SessionID)
	return repository.ExamResultParams{
		ExamID: exam.ID, StudentID: st.ID, SessionID: st.SessionID, ClassID: st.ClassID, SectionID: st.SectionID,
		TotalObtained: totalObtained, TotalFull: totalFull, Percentage: pct, GPA: gpa, CGPA: cgpa,
		Grade: grade, IsPassed: overallPassed, ResultStatus: model.ResultStatusDraft,
	}
}

func gradeFromPercentage(scales []repository.GradingScaleRecord, pct float64) (string, float64) {
	for _, sc := range scales {
		if pct >= sc.MinPercentage && pct <= sc.MaxPercentage {
			return sc.Grade, sc.GPAPoint
		}
	}
	return "F", 0
}

func (s *ExamService) PublishResults(ctx context.Context, examID uuid.UUID, actorID uuid.UUID, ip string) error {
	if err := s.repos.Exams.UpdateExamResultStatus(ctx, examID, model.ResultStatusPublished); err != nil {
		return err
	}
	s.audit.Log(ctx, &actorID, model.ActionUpdate, model.EntityExamResult, nil, ip, map[string]any{"exam_id": examID, "published": true})
	if s.notify != nil {
		eid := examID
		go func() {
			bg := context.Background()
			recs, _ := s.repos.Exams.ListExamResults(bg, repository.ResultSearchParams{ExamID: &eid, PublishedOnly: true, Limit: 500})
			for _, r := range recs {
				s.notify.OnResultPublished(bg, examID, r.StudentID, r.ExamName, r.GPA)
			}
		}()
	}
	return nil
}

func (s *ExamService) ListResults(ctx context.Context, f dto.ExamReportFilter, page, perPage int) (*dto.PaginatedResults, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	params := repository.ResultSearchParams{Limit: int32(perPage), Offset: int32((page - 1) * perPage)}
	if f.ExamID != uuid.Nil {
		params.ExamID = &f.ExamID
	}
	if f.SectionID != uuid.Nil {
		params.SectionID = &f.SectionID
	}
	if f.PassedOnly {
		params.PassedOnly = true
	}
	if f.FailedOnly {
		params.FailedOnly = true
	}
	if f.PublishedOnly {
		params.PublishedOnly = true
	}
	if f.StudentID != uuid.Nil {
		params.StudentID = &f.StudentID
	}
	total, _ := s.repos.Exams.CountExamResults(ctx, params)
	recs, err := s.repos.Exams.ListExamResults(ctx, params)
	if err != nil {
		return nil, err
	}
	items := make([]dto.ExamResultResponse, 0, len(recs))
	for _, r := range recs {
		items = append(items, mapResult(&r))
	}
	return &dto.PaginatedResults{
		Items: items, Total: total, Page: page, PerPage: perPage,
		TotalPages: int(math.Ceil(float64(total) / float64(perPage))),
	}, nil
}

func (s *ExamService) GetResult(ctx context.Context, id uuid.UUID) (*dto.ExamResultResponse, error) {
	rec, err := s.repos.Exams.GetExamResult(ctx, id)
	if err != nil || rec == nil {
		return nil, ErrNotFound
	}
	resp := mapResult(rec)
	resp.Subjects = s.buildSubjectDetails(ctx, rec.ExamID, rec.StudentID)
	return &resp, nil
}

func (s *ExamService) buildSubjectDetails(ctx context.Context, examID, studentID uuid.UUID) []dto.SubjectMarkDetail {
	subjects, _ := s.repos.Exams.ListExamSubjects(ctx, examID)
	marks, _ := s.repos.Exams.ListMarksByExam(ctx, examID)
	markMap := map[uuid.UUID]repository.StudentMarkRecord{}
	for _, m := range marks {
		if m.StudentID == studentID {
			markMap[m.ExamSubjectID] = m
		}
	}
	items := make([]dto.SubjectMarkDetail, 0, len(subjects))
	for _, subj := range subjects {
		m, ok := markMap[subj.ID]
		detail := dto.SubjectMarkDetail{
			SubjectName: subj.SubjectName, SubjectCode: subj.SubjectCode,
			FullMarks: subj.FullMarks, PassMarks: subj.PassMarks,
		}
		if ok && !m.IsAbsent {
			detail.WrittenScore = m.WrittenScore
			detail.MCQScore = m.MCQScore
			detail.PracticalScore = m.PracticalScore
			detail.TotalScore = m.TotalScore
			detail.IsPassed = m.TotalScore >= subj.PassMarks
		}
		items = append(items, detail)
	}
	return items
}

func (s *ExamService) DashboardStats(ctx context.Context) (*dto.ExamDashboardStats, error) {
	active, _ := s.repos.Exams.CountExamsByStatus(ctx, model.ExamStatusPublished)
	publishedResults, _ := s.repos.Exams.CountExamsByResultStatus(ctx, model.ResultStatusPublished)
	return &dto.ExamDashboardStats{ActiveExams: active, PublishedResults: publishedResults}, nil
}

func (s *ExamService) DashboardStatsForExam(ctx context.Context, examID uuid.UUID) (*dto.ExamDashboardStats, error) {
	passed, _ := s.repos.Exams.CountResultsPassed(ctx, examID, true)
	failed, _ := s.repos.Exams.CountResultsPassed(ctx, examID, false)
	gpaDist, _ := s.repos.Exams.GPADistribution(ctx, examID)
	subjPerf, _ := s.repos.Exams.SubjectPerformance(ctx, examID)
	total := passed + failed
	passRate := 0.0
	if total > 0 {
		passRate = float64(passed) / float64(total) * 100
	}
	stats := &dto.ExamDashboardStats{
		StudentsPassed: passed, StudentsFailed: failed, PassRate: passRate,
	}
	for _, g := range gpaDist {
		stats.GPADistribution = append(stats.GPADistribution, dto.GPADistPoint{Grade: g.Grade, Count: g.Count})
	}
	for _, sp := range subjPerf {
		stats.SubjectPerformance = append(stats.SubjectPerformance, dto.SubjectPerfPoint{
			SubjectName: sp.SubjectName, AvgScore: sp.AvgScore, PassRate: sp.PassRate,
		})
	}
	return stats, nil
}

func (s *ExamService) GenerateReportCard(ctx context.Context, resultID uuid.UUID, actorID uuid.UUID, ip string) (*dto.ReportCardData, error) {
	result, err := s.GetResult(ctx, resultID)
	if err != nil {
		return nil, err
	}
	token, err := generateCardToken()
	if err != nil {
		return nil, err
	}
	_, err = s.repos.Exams.CreateReportCard(ctx, resultID, result.ExamID, result.StudentID, token, actorID)
	if err != nil {
		return nil, err
	}
	school, _ := s.repos.Schools.Get(ctx)
	summary, _ := s.repos.Attendance.StudentAttendanceSummary(ctx, result.StudentID, nil, nil)
	attPct := 0.0
	if summary != nil {
		total := summary.PresentDays + summary.AbsentDays + summary.LateDays + summary.LeaveDays
		if total > 0 {
			attPct = float64(summary.PresentDays+summary.LateDays) / float64(total) * 100
		}
	}
	data := &dto.ReportCardData{
		StudentName: result.StudentName, AdmissionNo: result.AdmissionNo, RollNumber: result.RollNumber,
		ClassName: result.ClassName, SectionName: result.SectionName, ExamName: result.ExamName,
		Result: result, AttendancePct: attPct, CardToken: token,
	}
	if school != nil {
		data.SchoolName = school.Name
	}
	s.audit.Log(ctx, &actorID, model.ActionCreate, model.EntityReportCard, &resultID, ip, nil)
	return data, nil
}

func (s *ExamService) ParentResult(ctx context.Context, examID, studentID uuid.UUID) (*dto.ExamResultResponse, error) {
	rec, err := s.repos.Exams.GetStudentExamResult(ctx, examID, studentID)
	if err != nil || rec == nil {
		return nil, ErrNotFound
	}
	if rec.ResultStatus != model.ResultStatusPublished {
		return nil, ErrForbidden
	}
	resp := mapResult(rec)
	resp.Subjects = s.buildSubjectDetails(ctx, examID, studentID)
	return &resp, nil
}

func (s *ExamService) Tabulation(ctx context.Context, examID uuid.UUID, sectionID *uuid.UUID) ([]dto.ExamResultResponse, error) {
	params := repository.ResultSearchParams{ExamID: &examID, Limit: 500}
	if sectionID != nil {
		params.SectionID = sectionID
	}
	recs, err := s.repos.Exams.ListExamResults(ctx, params)
	if err != nil {
		return nil, err
	}
	items := make([]dto.ExamResultResponse, 0, len(recs))
	for _, r := range recs {
		items = append(items, mapResult(&r))
	}
	return items, nil
}

// --- Mappers ---

func mapExam(r *repository.ExamRecord) dto.ExamResponse {
	return dto.ExamResponse{
		ID: r.ID, Name: r.Name, ExamType: r.ExamType, SessionID: r.SessionID, SessionName: r.SessionName,
		ClassID: r.ClassID, ClassName: r.ClassName, StartDate: r.StartDate, EndDate: r.EndDate,
		TotalMarks: r.TotalMarks, PassingMarks: r.PassingMarks, GradingSystemID: r.GradingSystemID,
		Status: r.Status, ResultStatus: r.ResultStatus, SubjectCount: r.SubjectCount,
	}
}

func mapExamSubject(r *repository.ExamSubjectRecord) dto.ExamSubjectResponse {
	return dto.ExamSubjectResponse{
		ID: r.ID, ExamID: r.ExamID, SubjectID: r.SubjectID, SubjectName: r.SubjectName, SubjectCode: r.SubjectCode,
		FullMarks: r.FullMarks, PassMarks: r.PassMarks, WrittenMarks: r.WrittenMarks,
		MCQMarks: r.MCQMarks, PracticalMarks: r.PracticalMarks,
	}
}

func mapGradingSystem(r *repository.GradingSystemRecord, scales []repository.GradingScaleRecord) dto.GradingSystemResponse {
	resp := dto.GradingSystemResponse{ID: r.ID, Name: r.Name, IsDefault: r.IsDefault}
	for _, sc := range scales {
		resp.Scales = append(resp.Scales, dto.GradingScaleResponse{
			Grade: sc.Grade, MinPercentage: sc.MinPercentage, MaxPercentage: sc.MaxPercentage, GPAPoint: sc.GPAPoint,
		})
	}
	return resp
}

func mapResult(r *repository.ExamResultRecord) dto.ExamResultResponse {
	return dto.ExamResultResponse{
		ID: r.ID, ExamID: r.ExamID, ExamName: r.ExamName, StudentID: r.StudentID,
		StudentName: r.StudentName, AdmissionNo: r.AdmissionNo, RollNumber: r.RollNumber,
		ClassName: r.ClassName, SectionName: r.SectionName,
		TotalObtained: r.TotalObtained, TotalFull: r.TotalFull, Percentage: r.Percentage,
		GPA: r.GPA, CGPA: r.CGPA, Grade: r.Grade, IsPassed: r.IsPassed,
		ClassPosition: r.ClassPosition, SectionPosition: r.SectionPosition, MeritPosition: r.MeritPosition,
		ResultStatus: r.ResultStatus,
	}
}

func mapExamFilter(f dto.ExamSearchFilter) repository.ExamSearchParams {
	p := repository.ExamSearchParams{Status: f.Status, Query: f.Query}
	if f.SessionID != uuid.Nil {
		p.SessionID = &f.SessionID
	}
	if f.ClassID != uuid.Nil {
		p.ClassID = &f.ClassID
	}
	if f.PerPage > 0 {
		p.Limit = int32(f.PerPage)
		p.Offset = int32((f.Page - 1) * f.PerPage)
	}
	return p
}

func generateCardToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
