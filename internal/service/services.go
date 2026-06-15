package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/school-management/pos/internal/auth"
	"github.com/school-management/pos/internal/config"
	"github.com/school-management/pos/internal/dto"
	"github.com/school-management/pos/internal/model"
	"github.com/school-management/pos/internal/notify"
	"github.com/school-management/pos/internal/repository"
)

type Services struct {
	Auth           *AuthService
	Users          *UserService
	Roles          *RoleService
	School         *SchoolService
	Sessions       *SessionService
	Academic       *AcademicService
	Students       *StudentService
	HR             *HRService
	Attendance     *AttendanceService
	Fees           *FeeService
	Exam           *ExamService
	Accounting     *AccountingService
	Parent         *ParentService
	Notice         *NoticeService
	Notification   *NotificationService
	Admission      *AdmissionService
	Website        *WebsiteService
	Payment        *PaymentService
	Backup    *BackupService
	System    *SystemService
	Dashboard *DashboardService
	Audit          *AuditService
	Seed           *SeedService
}

func NewServices(repos *repository.Repositories, cfg *config.Config, tokens *auth.TokenManager, logger *slog.Logger) *Services {
	audit := NewAuditService(repos)
	var sms notify.SMSProvider = notify.NewLogSMSProvider(logger)
	var email notify.EmailProvider = notify.NewLogEmailProvider(logger)
	if cfg.SMS.Enabled && cfg.SMS.Provider == "bulksmsbd" {
		sms = notify.NewBulkSMSBDProvider(cfg.SMS.APIKey, cfg.SMS.SenderID, "", logger)
	}
	if cfg.SMTP.Enabled {
		email = notify.NewSMTPProvider(cfg.SMTP.Host, cfg.SMTP.Port, cfg.SMTP.Username, cfg.SMTP.Password, cfg.SMTP.From, logger)
	}
	notification := NewNotificationService(repos, sms, email, audit)
	attendance := NewAttendanceService(repos, audit)
	fees := NewFeeService(repos, audit)
	exam := NewExamService(repos, audit)
	attendance.SetNotifier(notification)
	fees.SetNotifier(notification)
	exam.SetNotifier(notification)
	admission := NewAdmissionService(repos, audit)
	paymentSvc := NewPaymentService(repos, audit, cfg, logger)
	paymentSvc.SetFees(fees)
	paymentSvc.SetAdmission(admission)
	paymentSvc.SetNotifier(notification)
	startTime := time.Now()
	backupSvc := NewBackupService(repos, cfg, audit, logger)
	systemSvc := NewSystemService(repos, cfg, audit, backupSvc, startTime)
	return &Services{
		Auth:         NewAuthService(repos, cfg, tokens, audit),
		Users:        NewUserService(repos, audit),
		Roles:        NewRoleService(repos, audit),
		School:       NewSchoolService(repos, audit),
		Sessions:     NewSessionService(repos, audit),
		Academic:     NewAcademicService(repos, audit),
		Students:     NewStudentService(repos, audit),
		HR:           NewHRService(repos, audit),
		Attendance:   attendance,
		Fees:         fees,
		Exam:         exam,
		Accounting:   NewAccountingService(repos, audit),
		Parent:       NewParentService(repos, audit),
		Notice:       NewNoticeService(repos, audit, notification),
		Notification: notification,
		Admission:    admission,
		Website:      NewWebsiteService(repos, audit),
		Payment:      paymentSvc,
		Backup:       backupSvc,
		System:       systemSvc,
		Dashboard:    NewDashboardService(repos),
		Audit:        audit,
		Seed:         NewSeedService(repos, cfg),
	}
}

type AuditService struct {
	repos *repository.Repositories
}

func NewAuditService(repos *repository.Repositories) *AuditService {
	return &AuditService{repos: repos}
}

func (s *AuditService) Log(ctx context.Context, userID *uuid.UUID, action, entityType string, entityID *uuid.UUID, ip string, metadata map[string]any) {
	if metadata == nil {
		metadata = map[string]any{}
	}
	_ = s.repos.AuditLogs.Create(ctx, repository.AuditParams{
		UserID:     userID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		IPAddress:  ip,
		Metadata:   metadata,
	})
}

func (s *AuditService) ListRecent(ctx context.Context, limit int32) ([]dto.ActivityItem, error) {
	rows, err := s.repos.AuditLogs.ListRecent(ctx, limit)
	if err != nil {
		return nil, err
	}
	items := make([]dto.ActivityItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, dto.ActivityItem{
			ID:          r.ID,
			Action:      r.Action,
			EntityType:  r.EntityType,
			UserEmail:   r.UserEmail,
			UserName:    strings.TrimSpace(r.UserName),
			Description: r.Description,
			CreatedAt:   r.CreatedAt,
		})
	}
	return items, nil
}

type AuthService struct {
	repos  *repository.Repositories
	cfg    *config.Config
	tokens *auth.TokenManager
	audit  *AuditService
}

func NewAuthService(repos *repository.Repositories, cfg *config.Config, tokens *auth.TokenManager, audit *AuditService) *AuthService {
	return &AuthService{repos: repos, cfg: cfg, tokens: tokens, audit: audit}
}

func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest, ip string) (*dto.TokenPair, *dto.AuthUser, error) {
	user, err := s.repos.Users.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, nil, err
	}
	if user == nil || !auth.CheckPassword(user.PasswordHash, req.Password) {
		return nil, nil, ErrInvalidCredentials
	}
	if !user.IsActive {
		return nil, nil, ErrInactiveUser
	}

	perms, err := s.repos.Users.GetPermissions(ctx, user.ID)
	if err != nil {
		return nil, nil, err
	}

	authUser := dto.AuthUser{
		ID:          user.ID,
		Email:       user.Email,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		RoleID:      user.RoleID,
		RoleName:    user.RoleName,
		RoleSlug:    user.RoleSlug,
		Permissions: perms,
	}

	pair, err := s.tokens.GenerateTokenPair(authUser)
	if err != nil {
		return nil, nil, err
	}

	_ = s.repos.Users.UpdateLastLogin(ctx, user.ID)
	s.audit.Log(ctx, &user.ID, model.ActionLogin, model.EntityUser, &user.ID, ip, map[string]any{"email": user.Email})

	return pair, &authUser, nil
}

func (s *AuthService) Logout(ctx context.Context, jti string, expiresAt time.Time, userID uuid.UUID, ip string) error {
	if err := s.repos.Auth.RevokeToken(ctx, jti, expiresAt); err != nil {
		return err
	}
	s.audit.Log(ctx, &userID, model.ActionLogout, model.EntityUser, &userID, ip, nil)
	return nil
}

func (s *AuthService) ForgotPassword(ctx context.Context, email string) (string, error) {
	user, err := s.repos.Users.GetByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", nil
	}

	plain, hashed, err := auth.GenerateResetToken()
	if err != nil {
		return "", err
	}
	expires := time.Now().Add(1 * time.Hour)
	if err := s.repos.Auth.CreatePasswordResetToken(ctx, user.ID, hashed, expires); err != nil {
		return "", err
	}
	return plain, nil
}

func (s *AuthService) ResetPassword(ctx context.Context, token, password string) error {
	hashedToken := auth.HashResetToken(token)
	rec, err := s.repos.Auth.GetValidPasswordResetToken(ctx, hashedToken)
	if err != nil {
		return err
	}
	if rec == nil {
		return ErrInvalidToken
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return err
	}
	if err := s.repos.Users.UpdatePassword(ctx, rec.UserID, hash); err != nil {
		return err
	}
	return s.repos.Auth.MarkPasswordResetTokenUsed(ctx, rec.ID)
}

func (s *AuthService) ChangePassword(ctx context.Context, userID uuid.UUID, current, newPassword string) error {
	user, err := s.repos.Users.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrNotFound
	}
	if !auth.CheckPassword(user.PasswordHash, current) {
		return ErrPasswordMismatch
	}
	hash, err := auth.HashPassword(newPassword)
	if err != nil {
		return err
	}
	return s.repos.Users.UpdatePassword(ctx, userID, hash)
}

func (s *AuthService) ValidateClaims(ctx context.Context, claims *auth.Claims) (*dto.AuthUser, error) {
	revoked, err := s.repos.Auth.IsTokenRevoked(ctx, claims.ID)
	if err != nil {
		return nil, err
	}
	if revoked {
		return nil, ErrUnauthorized
	}

	user, err := s.repos.Users.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil || !user.IsActive {
		return nil, ErrUnauthorized
	}

	perms, err := s.repos.Users.GetPermissions(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return &dto.AuthUser{
		ID:          user.ID,
		Email:       user.Email,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		RoleID:      user.RoleID,
		RoleName:    user.RoleName,
		RoleSlug:    user.RoleSlug,
		Permissions: perms,
	}, nil
}

type UserService struct {
	repos *repository.Repositories
	audit *AuditService
}

func NewUserService(repos *repository.Repositories, audit *AuditService) *UserService {
	return &UserService{repos: repos, audit: audit}
}

func (s *UserService) Create(ctx context.Context, req dto.CreateUserRequest, actorID uuid.UUID, ip string) (*dto.UserResponse, error) {
	existing, err := s.repos.Users.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrEmailExists
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user, err := s.repos.Users.Create(ctx, repository.CreateUserParams{
		Email:        strings.ToLower(req.Email),
		PasswordHash: hash,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Phone:        req.Phone,
		RoleID:       req.RoleID,
		IsActive:     req.IsActive,
	})
	if err != nil {
		return nil, err
	}

	resp := mapUser(user)
	s.audit.Log(ctx, &actorID, model.ActionCreate, model.EntityUser, &user.ID, ip, map[string]any{"email": user.Email})
	return &resp, nil
}

func (s *UserService) Update(ctx context.Context, id uuid.UUID, req dto.UpdateUserRequest, actorID uuid.UUID, ip string) (*dto.UserResponse, error) {
	user, err := s.repos.Users.Update(ctx, id, repository.UpdateUserParams{
		Email:     strings.ToLower(req.Email),
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		RoleID:    req.RoleID,
		IsActive:  req.IsActive,
	})
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrNotFound
	}
	resp := mapUser(user)
	s.audit.Log(ctx, &actorID, model.ActionUpdate, model.EntityUser, &id, ip, map[string]any{"email": user.Email})
	return &resp, nil
}

func (s *UserService) Delete(ctx context.Context, id uuid.UUID, actorID uuid.UUID, ip string) error {
	user, err := s.repos.Users.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrNotFound
	}
	if err := s.repos.Users.SoftDelete(ctx, id); err != nil {
		return err
	}
	s.audit.Log(ctx, &actorID, model.ActionDelete, model.EntityUser, &id, ip, map[string]any{"email": user.Email})
	return nil
}

func (s *UserService) Get(ctx context.Context, id uuid.UUID) (*dto.UserResponse, error) {
	user, err := s.repos.Users.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrNotFound
	}
	resp := mapUser(user)
	return &resp, nil
}

func (s *UserService) List(ctx context.Context, page, pageSize int) (*dto.PaginatedUsers, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := int32((page - 1) * pageSize)
	items, err := s.repos.Users.List(ctx, int32(pageSize), offset)
	if err != nil {
		return nil, err
	}
	total, err := s.repos.Users.Count(ctx)
	if err != nil {
		return nil, err
	}
	respItems := make([]dto.UserResponse, 0, len(items))
	for i := range items {
		respItems = append(respItems, mapUser(&items[i]))
	}
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}
	return &dto.PaginatedUsers{
		Items:      respItems,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *UserService) SetActive(ctx context.Context, id uuid.UUID, active bool, actorID uuid.UUID, ip string) (*dto.UserResponse, error) {
	user, err := s.repos.Users.SetActive(ctx, id, active)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrNotFound
	}
	resp := mapUser(user)
	s.audit.Log(ctx, &actorID, model.ActionUpdate, model.EntityUser, &id, ip, map[string]any{"is_active": active})
	return &resp, nil
}

func mapUser(u *repository.UserRecord) dto.UserResponse {
	return dto.UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Phone:     u.Phone,
		RoleID:    u.RoleID,
		RoleName:  u.RoleName,
		RoleSlug:  u.RoleSlug,
		IsActive:  u.IsActive,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

type RoleService struct {
	repos *repository.Repositories
	audit *AuditService
}

func NewRoleService(repos *repository.Repositories, audit *AuditService) *RoleService {
	return &RoleService{repos: repos, audit: audit}
}

func (s *RoleService) Create(ctx context.Context, req dto.CreateRoleRequest, actorID uuid.UUID, ip string) (*dto.RoleResponse, error) {
	slug := strings.ToLower(req.Slug)
	if existing, _ := s.repos.Roles.GetBySlug(ctx, slug); existing != nil {
		return nil, ErrRoleExists
	}
	role, err := s.repos.Roles.Create(ctx, req.Name, slug, req.Description, false)
	if err != nil {
		return nil, err
	}
	resp := mapRole(role, nil)
	s.audit.Log(ctx, &actorID, model.ActionCreate, model.EntityRole, &role.ID, ip, map[string]any{"slug": slug})
	return &resp, nil
}

func (s *RoleService) Update(ctx context.Context, id uuid.UUID, req dto.UpdateRoleRequest, actorID uuid.UUID, ip string) (*dto.RoleResponse, error) {
	existing, err := s.repos.Roles.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrNotFound
	}
	if existing.IsSystem {
		return nil, ErrSystemRole
	}
	role, err := s.repos.Roles.Update(ctx, id, req.Name, strings.ToLower(req.Slug), req.Description)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, ErrNotFound
	}
	perms, _ := s.repos.Roles.GetPermissionSlugs(ctx, id)
	resp := mapRole(role, perms)
	s.audit.Log(ctx, &actorID, model.ActionUpdate, model.EntityRole, &id, ip, map[string]any{"slug": role.Slug})
	return &resp, nil
}

func (s *RoleService) Delete(ctx context.Context, id uuid.UUID, actorID uuid.UUID, ip string) error {
	role, err := s.repos.Roles.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if role == nil {
		return ErrNotFound
	}
	if role.IsSystem {
		return ErrSystemRole
	}
	if err := s.repos.Roles.SoftDelete(ctx, id); err != nil {
		return err
	}
	s.audit.Log(ctx, &actorID, model.ActionDelete, model.EntityRole, &id, ip, map[string]any{"slug": role.Slug})
	return nil
}

func (s *RoleService) Get(ctx context.Context, id uuid.UUID) (*dto.RoleResponse, error) {
	role, err := s.repos.Roles.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, ErrNotFound
	}
	perms, err := s.repos.Roles.GetPermissionSlugs(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := mapRole(role, perms)
	return &resp, nil
}

func (s *RoleService) List(ctx context.Context) ([]dto.RoleResponse, error) {
	roles, err := s.repos.Roles.List(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]dto.RoleResponse, 0, len(roles))
	for i := range roles {
		perms, _ := s.repos.Roles.GetPermissionSlugs(ctx, roles[i].ID)
		items = append(items, mapRole(&roles[i], perms))
	}
	return items, nil
}

func (s *RoleService) AssignPermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID, actorID uuid.UUID, ip string) (*dto.RoleResponse, error) {
	role, err := s.repos.Roles.GetByID(ctx, roleID)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, ErrNotFound
	}
	if err := s.repos.Roles.ClearPermissions(ctx, roleID); err != nil {
		return nil, err
	}
	for _, pid := range permissionIDs {
		if err := s.repos.Roles.AssignPermission(ctx, roleID, pid); err != nil {
			return nil, err
		}
	}
	perms, err := s.repos.Roles.GetPermissionSlugs(ctx, roleID)
	if err != nil {
		return nil, err
	}
	resp := mapRole(role, perms)
	s.audit.Log(ctx, &actorID, model.ActionUpdate, model.EntityRole, &roleID, ip, map[string]any{"permissions": len(permissionIDs)})
	return &resp, nil
}

func (s *RoleService) ListPermissions(ctx context.Context) ([]dto.PermissionResponse, error) {
	perms, err := s.repos.Permissions.List(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]dto.PermissionResponse, 0, len(perms))
	for _, p := range perms {
		items = append(items, dto.PermissionResponse{
			ID:          p.ID,
			Name:        p.Name,
			Slug:        p.Slug,
			Description: p.Description,
			Module:      p.Module,
		})
	}
	return items, nil
}

func mapRole(r *repository.RoleRecord, perms []string) dto.RoleResponse {
	return dto.RoleResponse{
		ID:          r.ID,
		Name:        r.Name,
		Slug:        r.Slug,
		Description: r.Description,
		IsSystem:    r.IsSystem,
		Permissions: perms,
	}
}

type SchoolService struct {
	repos *repository.Repositories
	audit *AuditService
}

func NewSchoolService(repos *repository.Repositories, audit *AuditService) *SchoolService {
	return &SchoolService{repos: repos, audit: audit}
}

func (s *SchoolService) Get(ctx context.Context) (*dto.SchoolResponse, error) {
	school, err := s.repos.Schools.Get(ctx)
	if err != nil {
		return nil, err
	}
	if school == nil {
		return nil, nil
	}
	resp := mapSchool(school)
	return &resp, nil
}

func (s *SchoolService) Save(ctx context.Context, req dto.SchoolSetupRequest, logoURL string, actorID uuid.UUID, ip string) (*dto.SchoolResponse, error) {
	params := repository.SchoolParams{
		Name:    req.Name,
		LogoURL: logoURL,
		Address: req.Address,
		Email:   req.Email,
		Phone:   req.Phone,
		Website: req.Website,
	}
	existing, err := s.repos.Schools.Get(ctx)
	if err != nil {
		return nil, err
	}
	var school *repository.SchoolRecord
	if existing == nil {
		school, err = s.repos.Schools.Create(ctx, params)
	} else {
		if logoURL == "" {
			params.LogoURL = existing.LogoURL
		}
		school, err = s.repos.Schools.Update(ctx, existing.ID, params)
	}
	if err != nil {
		return nil, err
	}
	if school == nil {
		return nil, fmt.Errorf("save school failed")
	}
	resp := mapSchool(school)
	action := model.ActionUpdate
	if existing == nil {
		action = model.ActionCreate
	}
	s.audit.Log(ctx, &actorID, action, model.EntitySchool, &school.ID, ip, map[string]any{"name": school.Name})
	return &resp, nil
}

func mapSchool(s *repository.SchoolRecord) dto.SchoolResponse {
	return dto.SchoolResponse{
		ID:      s.ID,
		Name:    s.Name,
		LogoURL: s.LogoURL,
		Address: s.Address,
		Email:   s.Email,
		Phone:   s.Phone,
		Website: s.Website,
	}
}

type SessionService struct {
	repos *repository.Repositories
	audit *AuditService
}

func NewSessionService(repos *repository.Repositories, audit *AuditService) *SessionService {
	return &SessionService{repos: repos, audit: audit}
}

func (s *SessionService) Create(ctx context.Context, req dto.AcademicSessionRequest, actorID uuid.UUID, ip string) (*dto.AcademicSessionResponse, error) {
	rec, err := s.repos.AcademicSessions.Create(ctx, req.Name, req.StartDate, req.EndDate, req.IsActive)
	if err != nil {
		return nil, err
	}
	resp := mapSession(rec)
	s.audit.Log(ctx, &actorID, model.ActionCreate, model.EntityAcademicSession, &rec.ID, ip, map[string]any{"name": rec.Name})
	return &resp, nil
}

func (s *SessionService) Update(ctx context.Context, id uuid.UUID, req dto.AcademicSessionRequest, actorID uuid.UUID, ip string) (*dto.AcademicSessionResponse, error) {
	rec, err := s.repos.AcademicSessions.Update(ctx, id, req.Name, req.StartDate, req.EndDate, req.IsActive)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, ErrNotFound
	}
	resp := mapSession(rec)
	s.audit.Log(ctx, &actorID, model.ActionUpdate, model.EntityAcademicSession, &id, ip, map[string]any{"name": rec.Name})
	return &resp, nil
}

func (s *SessionService) Delete(ctx context.Context, id uuid.UUID, actorID uuid.UUID, ip string) error {
	rec, err := s.repos.AcademicSessions.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if rec == nil {
		return ErrNotFound
	}
	if err := s.repos.AcademicSessions.SoftDelete(ctx, id); err != nil {
		return err
	}
	s.audit.Log(ctx, &actorID, model.ActionDelete, model.EntityAcademicSession, &id, ip, map[string]any{"name": rec.Name})
	return nil
}

func (s *SessionService) List(ctx context.Context) ([]dto.AcademicSessionResponse, error) {
	recs, err := s.repos.AcademicSessions.List(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]dto.AcademicSessionResponse, 0, len(recs))
	for i := range recs {
		items = append(items, mapSession(&recs[i]))
	}
	return items, nil
}

func (s *SessionService) Get(ctx context.Context, id uuid.UUID) (*dto.AcademicSessionResponse, error) {
	rec, err := s.repos.AcademicSessions.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, ErrNotFound
	}
	resp := mapSession(rec)
	return &resp, nil
}

func (s *SessionService) SetActive(ctx context.Context, id uuid.UUID, actorID uuid.UUID, ip string) (*dto.AcademicSessionResponse, error) {
	rec, err := s.repos.AcademicSessions.SetActive(ctx, id)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, ErrNotFound
	}
	resp := mapSession(rec)
	s.audit.Log(ctx, &actorID, model.ActionUpdate, model.EntityAcademicSession, &id, ip, map[string]any{"is_active": true})
	return &resp, nil
}

func mapSession(s *repository.SessionRecord) dto.AcademicSessionResponse {
	return dto.AcademicSessionResponse{
		ID:        s.ID,
		Name:      s.Name,
		StartDate: s.StartDate,
		EndDate:   s.EndDate,
		IsActive:  s.IsActive,
	}
}

type DashboardService struct {
	repos *repository.Repositories
}

func NewDashboardService(repos *repository.Repositories) *DashboardService {
	return &DashboardService{repos: repos}
}

func (s *DashboardService) Stats(ctx context.Context) (*dto.DashboardStats, error) {
	teachers, _ := s.repos.HR.CountActiveTeachers(ctx)
	staff, _ := s.repos.HR.CountActiveStaff(ctx)
	students, _ := s.repos.Students.CountActive(ctx)
	newAdmissions, _ := s.repos.Students.CountNewAdmissionsThisMonth(ctx)
	byClass, _ := s.repos.Students.CountByClass(ctx)
	active, _ := s.repos.AcademicSessions.GetActive(ctx)
	recent, _ := s.repos.AuditLogs.ListRecent(ctx, 10)

	stats := &dto.DashboardStats{
		TotalStudents:          students,
		TotalTeachers:          teachers,
		TotalStaff:             staff,
		NewAdmissionsThisMonth: newAdmissions,
		HasActiveSession:       active != nil,
	}
	if active != nil {
		stats.ActiveSessionName = active.Name
	}
	for _, c := range byClass {
		stats.StudentsByClass = append(stats.StudentsByClass, dto.ClassStudentCount{
			ClassID: c.ClassID, ClassName: c.ClassName, StudentCount: c.StudentCount,
		})
	}
	for _, r := range recent {
		stats.RecentActivities = append(stats.RecentActivities, dto.ActivityItem{
			ID: r.ID, Action: r.Action, EntityType: r.EntityType,
			UserEmail: r.UserEmail, UserName: strings.TrimSpace(r.UserName),
			Description: r.Description, CreatedAt: r.CreatedAt,
		})
	}
	return stats, nil
}

type SeedService struct {
	repos *repository.Repositories
	cfg   *config.Config
}

func NewSeedService(repos *repository.Repositories, cfg *config.Config) *SeedService {
	return &SeedService{repos: repos, cfg: cfg}
}

func (s *SeedService) EnsureAdmin(ctx context.Context) error {
	existing, err := s.repos.Users.GetByEmail(ctx, s.cfg.Admin.Email)
	if err != nil {
		return err
	}
	if existing != nil {
		return nil
	}
	role, err := s.repos.Roles.GetBySlug(ctx, model.RoleAdmin)
	if err != nil {
		return err
	}
	if role == nil {
		return fmt.Errorf("admin role not found; run migrations first")
	}
	hash, err := auth.HashPassword(s.cfg.Admin.Password)
	if err != nil {
		return err
	}
	_, err = s.repos.Users.Create(ctx, repository.CreateUserParams{
		Email:        strings.ToLower(s.cfg.Admin.Email),
		PasswordHash: hash,
		FirstName:    s.cfg.Admin.FirstName,
		LastName:     s.cfg.Admin.LastName,
		RoleID:       role.ID,
		IsActive:     true,
	})
	return err
}
