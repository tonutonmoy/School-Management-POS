package handler

import (
	"log/slog"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/school-management/pos/internal/config"
	"github.com/school-management/pos/internal/dto"
	"github.com/school-management/pos/internal/i18n"
	"github.com/school-management/pos/internal/middleware"
	"github.com/school-management/pos/internal/model"
	"github.com/school-management/pos/internal/service"
	"github.com/school-management/pos/internal/storage"
	"github.com/school-management/pos/internal/validator"
	"github.com/school-management/pos/internal/web"
)

type Handler struct {
	services  *service.Services
	validate  *validator.Validator
	cfg       *config.Config
	storage   storage.Storage
	logger    *slog.Logger
}

func New(services *service.Services, validate *validator.Validator, cfg *config.Config, store storage.Storage, logger *slog.Logger) *Handler {
	return &Handler{services: services, validate: validate, cfg: cfg, storage: store, logger: logger}
}

func (h *Handler) Register(app fiber.Router, mw *middleware.Middleware) {
	app.Use(mw.Locale())
	app.Use(mw.CSRFGenerate())

	app.Get("/lang/:code", h.SetLanguage)

	app.Get("/", func(c fiber.Ctx) error {
		if u := middleware.GetUser(c); u != nil {
			return c.Redirect().To(middleware.HomePath(u))
		}
		return c.Redirect().To("/site")
	})

	app.Get("/login", mw.Authenticate(false), h.LoginPage)
	app.Post("/login", mw.Authenticate(false), mw.CSRFProtect(), h.Login)
	app.Post("/logout", mw.Authenticate(true), mw.CSRFProtect(), h.Logout)
	app.Get("/forgot-password", h.ForgotPasswordPage)
	app.Post("/forgot-password", mw.CSRFProtect(), h.ForgotPassword)
	app.Get("/reset-password", h.ResetPasswordPage)
	app.Post("/reset-password", mw.CSRFProtect(), h.ResetPassword)

	auth := app.Group("", mw.Authenticate(true))
	auth.Get("/dashboard", h.Dashboard)
	auth.Get("/change-password", h.ChangePasswordPage)
	auth.Post("/change-password", mw.CSRFProtect(), h.ChangePassword)

	users := auth.Group("/users", mw.RequirePermission("user.manage"))
	users.Get("/", h.UserList)
	users.Get("/new", h.UserCreatePage)
	users.Post("/", mw.CSRFProtect(), h.UserCreate)
	users.Get("/:id", h.UserDetail)
	users.Get("/:id/edit", h.UserEditPage)
	users.Post("/:id", mw.CSRFProtect(), h.UserUpdate)
	users.Post("/:id/delete", mw.CSRFProtect(), h.UserDelete)
	users.Post("/:id/toggle", mw.CSRFProtect(), h.UserToggle)

	roles := auth.Group("/roles", mw.RequirePermission("role.manage"))
	roles.Get("/", h.RoleList)
	roles.Get("/new", h.RoleCreatePage)
	roles.Post("/", mw.CSRFProtect(), h.RoleCreate)
	roles.Get("/:id/edit", h.RoleEditPage)
	roles.Post("/:id", mw.CSRFProtect(), h.RoleUpdate)
	roles.Post("/:id/delete", mw.CSRFProtect(), h.RoleDelete)
	roles.Post("/:id/permissions", mw.CSRFProtect(), h.RoleAssignPermissions)

	school := auth.Group("/school", mw.RequirePermission("school.manage"))
	school.Get("/", h.SchoolPage)
	school.Post("/", mw.CSRFProtect(), h.SchoolSave)

	sessions := auth.Group("/sessions", mw.RequirePermission("session.manage"))
	sessions.Get("/", h.SessionList)
	sessions.Get("/new", h.SessionCreatePage)
	sessions.Post("/", mw.CSRFProtect(), h.SessionCreate)
	sessions.Get("/:id/edit", h.SessionEditPage)
	sessions.Post("/:id", mw.CSRFProtect(), h.SessionUpdate)
	sessions.Post("/:id/delete", mw.CSRFProtect(), h.SessionDelete)
	sessions.Post("/:id/activate", mw.CSRFProtect(), h.SessionActivate)

	auth.Get("/audit-logs", mw.RequirePermission("audit.view"), h.AuditList)

	h.registerAcademicRoutes(auth, mw)
	h.registerStudentRoutes(auth, mw)
	h.registerHRRoutes(auth, mw)
	h.registerAttendanceRoutes(auth, mw)
	h.registerFeeRoutes(auth, mw)
	h.registerExamRoutes(auth, mw)
	h.registerAccountingRoutes(auth, mw)
	h.registerParentPortalRoutes(auth, mw)
	h.registerAdmissionRoutes(app, auth, mw)
	h.registerWebsiteRoutes(app, auth, mw)
	h.registerPaymentRoutes(app, auth, mw)
	h.registerCommercialRoutes(app, auth, mw)
}

func (h *Handler) render(c fiber.Ctx, status int, page web.Page) error {
	c.Set("Content-Type", "text/html; charset=utf-8")
	web.SetRenderLang(middleware.GetLang(c))
	html := page.Render(middleware.GetCSRF(c), middleware.GetUser(c), h.cfg.App.Name)
	return c.Status(status).SendString(html)
}

func (h *Handler) SetLanguage(c fiber.Ctx) error {
	lang := i18n.Parse(c.Params("code"))
	middleware.SetLangCookie(c, lang)
	dest := c.Get("Referer")
	if dest == "" {
		dest = "/"
	}
	return c.Redirect().To(dest)
}

func (h *Handler) flash(c fiber.Ctx, msg string, isError bool) {
	c.Cookie(&fiber.Cookie{Name: "flash", Value: msg, Path: "/", MaxAge: 5})
	if isError {
		c.Cookie(&fiber.Cookie{Name: "flash_type", Value: "error", Path: "/", MaxAge: 5})
	} else {
		c.Cookie(&fiber.Cookie{Name: "flash_type", Value: "success", Path: "/", MaxAge: 5})
	}
}

func parseUUIDParam(c fiber.Ctx, name string) (uuid.UUID, error) {
	return uuid.Parse(c.Params(name))
}

func parseDate(s string) (time.Time, error) {
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}
	return time.Parse(time.RFC3339, s)
}

func (h *Handler) LoginPage(c fiber.Ctx) error {
	return h.render(c, fiber.StatusOK, web.LoginPage{
		Flash: c.Cookies("flash"), FlashType: c.Cookies("flash_type"),
		DefaultEmail: h.cfg.Login.Email, DefaultPassword: h.cfg.Login.Password,
	})
}

func (h *Handler) Login(c fiber.Ctx) error {
	req := dto.LoginRequest{
		Email:    c.FormValue("email"),
		Password: c.FormValue("password"),
	}
	if errs := h.validate.Validate(req); len(errs) > 0 {
		h.flash(c, validator.FirstError(errs), true)
		return c.Redirect().To("/login")
	}
	pair, authUser, err := h.services.Auth.Login(c.Context(), req, c.IP())
	if err != nil {
		h.services.System.RecordLoginAttempt(c.Context(), req.Email, c.IP(), string(c.Request().Header.UserAgent()), false)
		h.services.Audit.Log(c.Context(), nil, model.ActionLoginFailed, model.EntityUser, nil, c.IP(), map[string]any{"email": req.Email})
		h.flash(c, "Invalid email or password", true)
		return c.Redirect().To("/login")
	}
	h.services.System.RecordLoginAttempt(c.Context(), req.Email, c.IP(), string(c.Request().Header.UserAgent()), true)
	middleware.ClearStaleCookies(c)
	middleware.SetAuthCookie(c, pair.AccessToken, pair.ExpiresAt)
	return c.Redirect().To(middleware.HomePath(authUser))
}

func (h *Handler) Logout(c fiber.Ctx) error {
	claims := middleware.GetClaims(c)
	user := middleware.GetUser(c)
	if claims != nil && user != nil && claims.ExpiresAt != nil {
		_ = h.services.Auth.Logout(c.Context(), claims.ID, claims.ExpiresAt.Time, user.ID, c.IP())
	}
	middleware.SetAuthCookie(c, "", time.Now().Add(-time.Hour))
	middleware.ClearStaleCookies(c)
	return c.Redirect().To("/login")
}

func (h *Handler) ForgotPasswordPage(c fiber.Ctx) error {
	return h.render(c, fiber.StatusOK, web.ForgotPasswordPage{Flash: c.Cookies("flash"), FlashType: c.Cookies("flash_type")})
}

func (h *Handler) ForgotPassword(c fiber.Ctx) error {
	req := dto.ForgotPasswordRequest{Email: c.FormValue("email")}
	if errs := h.validate.Validate(req); len(errs) > 0 {
		h.flash(c, validator.FirstError(errs), true)
		return c.Redirect().To("/forgot-password")
	}
	token, err := h.services.Auth.ForgotPassword(c.Context(), req.Email)
	if err != nil {
		h.flash(c, "Unable to process request", true)
		return c.Redirect().To("/forgot-password")
	}
	if token != "" && h.cfg.App.Env == "development" {
		h.logger.Info("password reset token generated", "token", token)
	}
	h.flash(c, "If the email exists, reset instructions were sent.", false)
	return c.Redirect().To("/login")
}

func (h *Handler) ResetPasswordPage(c fiber.Ctx) error {
	return h.render(c, fiber.StatusOK, web.ResetPasswordPage{Token: c.Query("token"), Flash: c.Cookies("flash"), FlashType: c.Cookies("flash_type")})
}

func (h *Handler) ResetPassword(c fiber.Ctx) error {
	req := dto.ResetPasswordRequest{
		Token:           c.FormValue("token"),
		Password:        c.FormValue("password"),
		ConfirmPassword: c.FormValue("confirm_password"),
	}
	if errs := h.validate.Validate(req); len(errs) > 0 {
		h.flash(c, validator.FirstError(errs), true)
		return c.Redirect().To("/reset-password?token=" + req.Token)
	}
	if err := h.services.Auth.ResetPassword(c.Context(), req.Token, req.Password); err != nil {
		h.flash(c, "Invalid or expired reset link", true)
		return c.Redirect().To("/forgot-password")
	}
	h.flash(c, "Password updated successfully", false)
	return c.Redirect().To("/login")
}

func (h *Handler) ChangePasswordPage(c fiber.Ctx) error {
	return h.render(c, fiber.StatusOK, web.ChangePasswordPage{Flash: c.Cookies("flash"), FlashType: c.Cookies("flash_type")})
}

func (h *Handler) ChangePassword(c fiber.Ctx) error {
	user := middleware.GetUser(c)
	req := dto.ChangePasswordRequest{
		CurrentPassword: c.FormValue("current_password"),
		NewPassword:     c.FormValue("new_password"),
		ConfirmPassword: c.FormValue("confirm_password"),
	}
	if errs := h.validate.Validate(req); len(errs) > 0 {
		h.flash(c, validator.FirstError(errs), true)
		return c.Redirect().To("/change-password")
	}
	if err := h.services.Auth.ChangePassword(c.Context(), user.ID, req.CurrentPassword, req.NewPassword); err != nil {
		h.flash(c, "Current password is incorrect", true)
		return c.Redirect().To("/change-password")
	}
	h.flash(c, "Password changed", false)
	return c.Redirect().To("/dashboard")
}

func (h *Handler) Dashboard(c fiber.Ctx) error {
	stats, err := h.services.Dashboard.Stats(c.Context())
	if err != nil {
		return c.Status(500).SendString("Failed to load dashboard")
	}
	return h.render(c, fiber.StatusOK, web.DashboardPage{Stats: stats, Flash: c.Cookies("flash"), FlashType: c.Cookies("flash_type")})
}

func (h *Handler) UserList(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	data, err := h.services.Users.List(c.Context(), page, 20)
	if err != nil {
		return c.Status(500).SendString("Failed to load users")
	}
	return h.render(c, fiber.StatusOK, web.UserListPage{Data: data})
}

func (h *Handler) UserCreatePage(c fiber.Ctx) error {
	roles, _ := h.services.Roles.List(c.Context())
	return h.render(c, fiber.StatusOK, web.UserFormPage{Title: "Create User", Roles: roles})
}

func (h *Handler) UserCreate(c fiber.Ctx) error {
	user := middleware.GetUser(c)
	roleID, err := uuid.Parse(c.FormValue("role_id"))
	if err != nil {
		h.flash(c, "Invalid role", true)
		return c.Redirect().To("/users/new")
	}
	req := dto.CreateUserRequest{
		Email:     c.FormValue("email"),
		Password:  c.FormValue("password"),
		FirstName: c.FormValue("first_name"),
		LastName:  c.FormValue("last_name"),
		Phone:     c.FormValue("phone"),
		RoleID:    roleID,
		IsActive:  c.FormValue("is_active") == "on" || c.FormValue("is_active") == "true",
	}
	if errs := h.validate.Validate(req); len(errs) > 0 {
		h.flash(c, validator.FirstError(errs), true)
		return c.Redirect().To("/users/new")
	}
	if _, err := h.services.Users.Create(c.Context(), req, user.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
		return c.Redirect().To("/users/new")
	}
	return c.Redirect().To("/users")
}

func (h *Handler) UserDetail(c fiber.Ctx) error {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}
	u, err := h.services.Users.Get(c.Context(), id)
	if err != nil {
		return c.Status(404).SendString("User not found")
	}
	return h.render(c, fiber.StatusOK, web.UserDetailPage{User: *u})
}

func (h *Handler) UserEditPage(c fiber.Ctx) error {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}
	u, err := h.services.Users.Get(c.Context(), id)
	if err != nil {
		return c.Status(404).SendString("User not found")
	}
	roles, _ := h.services.Roles.List(c.Context())
	return h.render(c, fiber.StatusOK, web.UserFormPage{Title: "Edit User", User: u, Roles: roles})
}

func (h *Handler) UserUpdate(c fiber.Ctx) error {
	actor := middleware.GetUser(c)
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}
	roleID, err := uuid.Parse(c.FormValue("role_id"))
	if err != nil {
		h.flash(c, "Invalid role", true)
		return c.Redirect().To("/users/" + id.String() + "/edit")
	}
	req := dto.UpdateUserRequest{
		Email:     c.FormValue("email"),
		FirstName: c.FormValue("first_name"),
		LastName:  c.FormValue("last_name"),
		Phone:     c.FormValue("phone"),
		RoleID:    roleID,
		IsActive:  c.FormValue("is_active") == "on" || c.FormValue("is_active") == "true",
	}
	if errs := h.validate.Validate(req); len(errs) > 0 {
		h.flash(c, validator.FirstError(errs), true)
		return c.Redirect().To("/users/" + id.String() + "/edit")
	}
	if _, err := h.services.Users.Update(c.Context(), id, req, actor.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
		return c.Redirect().To("/users/" + id.String() + "/edit")
	}
	return c.Redirect().To("/users/" + id.String())
}

func (h *Handler) UserDelete(c fiber.Ctx) error {
	actor := middleware.GetUser(c)
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}
	if err := h.services.Users.Delete(c.Context(), id, actor.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
	}
	return c.Redirect().To("/users")
}

func (h *Handler) UserToggle(c fiber.Ctx) error {
	actor := middleware.GetUser(c)
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}
	active := c.FormValue("active") == "true"
	if _, err := h.services.Users.SetActive(c.Context(), id, active, actor.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
	}
	return c.Redirect().To("/users")
}

func (h *Handler) RoleList(c fiber.Ctx) error {
	roles, err := h.services.Roles.List(c.Context())
	if err != nil {
		return c.Status(500).SendString("Failed to load roles")
	}
	return h.render(c, fiber.StatusOK, web.RoleListPage{Roles: roles})
}

func (h *Handler) RoleCreatePage(c fiber.Ctx) error {
	perms, _ := h.services.Roles.ListPermissions(c.Context())
	return h.render(c, fiber.StatusOK, web.RoleFormPage{Title: "Create Role", Permissions: perms})
}

func (h *Handler) RoleCreate(c fiber.Ctx) error {
	actor := middleware.GetUser(c)
	req := dto.CreateRoleRequest{
		Name:        c.FormValue("name"),
		Slug:        c.FormValue("slug"),
		Description: c.FormValue("description"),
	}
	if errs := h.validate.Validate(req); len(errs) > 0 {
		h.flash(c, validator.FirstError(errs), true)
		return c.Redirect().To("/roles/new")
	}
	role, err := h.services.Roles.Create(c.Context(), req, actor.ID, c.IP())
	if err != nil {
		h.flash(c, err.Error(), true)
		return c.Redirect().To("/roles/new")
	}
	permIDs := parseUUIDList(c)
	if len(permIDs) > 0 {
		_, _ = h.services.Roles.AssignPermissions(c.Context(), role.ID, permIDs, actor.ID, c.IP())
	}
	return c.Redirect().To("/roles")
}

func (h *Handler) RoleEditPage(c fiber.Ctx) error {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}
	role, err := h.services.Roles.Get(c.Context(), id)
	if err != nil {
		return c.Status(404).SendString("Role not found")
	}
	perms, _ := h.services.Roles.ListPermissions(c.Context())
	return h.render(c, fiber.StatusOK, web.RoleFormPage{Title: "Edit Role", Role: role, Permissions: perms})
}

func (h *Handler) RoleUpdate(c fiber.Ctx) error {
	actor := middleware.GetUser(c)
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}
	req := dto.UpdateRoleRequest{
		Name:        c.FormValue("name"),
		Slug:        c.FormValue("slug"),
		Description: c.FormValue("description"),
	}
	if errs := h.validate.Validate(req); len(errs) > 0 {
		h.flash(c, validator.FirstError(errs), true)
		return c.Redirect().To("/roles/" + id.String() + "/edit")
	}
	if _, err := h.services.Roles.Update(c.Context(), id, req, actor.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
		return c.Redirect().To("/roles/" + id.String() + "/edit")
	}
	permIDs := parseUUIDList(c)
	if len(permIDs) > 0 {
		_, _ = h.services.Roles.AssignPermissions(c.Context(), id, permIDs, actor.ID, c.IP())
	}
	return c.Redirect().To("/roles")
}

func (h *Handler) RoleDelete(c fiber.Ctx) error {
	actor := middleware.GetUser(c)
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}
	if err := h.services.Roles.Delete(c.Context(), id, actor.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
	}
	return c.Redirect().To("/roles")
}

func (h *Handler) RoleAssignPermissions(c fiber.Ctx) error {
	actor := middleware.GetUser(c)
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}
	permIDs := parseUUIDList(c)
	if _, err := h.services.Roles.AssignPermissions(c.Context(), id, permIDs, actor.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
	}
	return c.Redirect().To("/roles/" + id.String() + "/edit")
}

func parseUUIDList(c fiber.Ctx) []uuid.UUID {
	var ids []uuid.UUID
	c.Request().PostArgs().VisitAll(func(key, value []byte) {
		if string(key) == "permission_ids" {
			if id, err := uuid.Parse(string(value)); err == nil {
				ids = append(ids, id)
			}
		}
	})
	return ids
}

func (h *Handler) SchoolPage(c fiber.Ctx) error {
	school, _ := h.services.School.Get(c.Context())
	return h.render(c, fiber.StatusOK, web.SchoolPage{School: school, Flash: c.Cookies("flash"), FlashType: c.Cookies("flash_type")})
}

func (h *Handler) SchoolSave(c fiber.Ctx) error {
	actor := middleware.GetUser(c)
	req := dto.SchoolSetupRequest{
		Name:    c.FormValue("name"),
		Address: c.FormValue("address"),
		Email:   c.FormValue("email"),
		Phone:   c.FormValue("phone"),
		Website: c.FormValue("website"),
	}
	if errs := h.validate.Validate(req); len(errs) > 0 {
		h.flash(c, validator.FirstError(errs), true)
		return c.Redirect().To("/school")
	}
	logoURL := ""
	file, err := c.FormFile("logo")
	if err == nil && file != nil {
		f, err := file.Open()
		if err == nil {
			defer f.Close()
			contentType := file.Header.Get("Content-Type")
			if contentType == "" {
				contentType = "application/octet-stream"
			}
			logoURL, _ = h.storage.Upload(c.Context(), file.Filename, f, contentType)
		}
	}
	if _, err := h.services.School.Save(c.Context(), req, logoURL, actor.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
		return c.Redirect().To("/school")
	}
	h.flash(c, "School settings saved", false)
	return c.Redirect().To("/school")
}

func (h *Handler) SessionList(c fiber.Ctx) error {
	sessions, err := h.services.Sessions.List(c.Context())
	if err != nil {
		return c.Status(500).SendString("Failed to load sessions")
	}
	return h.render(c, fiber.StatusOK, web.SessionListPage{Sessions: sessions})
}

func (h *Handler) SessionCreatePage(c fiber.Ctx) error {
	return h.render(c, fiber.StatusOK, web.SessionFormPage{Title: "Create Session"})
}

func (h *Handler) SessionCreate(c fiber.Ctx) error {
	actor := middleware.GetUser(c)
	start, err1 := parseDate(c.FormValue("start_date"))
	end, err2 := parseDate(c.FormValue("end_date"))
	if err1 != nil || err2 != nil {
		h.flash(c, "Invalid dates", true)
		return c.Redirect().To("/sessions/new")
	}
	req := dto.AcademicSessionRequest{
		Name:      c.FormValue("name"),
		StartDate: start,
		EndDate:   end,
		IsActive:  c.FormValue("is_active") == "on" || c.FormValue("is_active") == "true",
	}
	if errs := h.validate.Validate(req); len(errs) > 0 {
		h.flash(c, validator.FirstError(errs), true)
		return c.Redirect().To("/sessions/new")
	}
	if _, err := h.services.Sessions.Create(c.Context(), req, actor.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
		return c.Redirect().To("/sessions/new")
	}
	return c.Redirect().To("/sessions")
}

func (h *Handler) SessionEditPage(c fiber.Ctx) error {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}
	s, err := h.services.Sessions.Get(c.Context(), id)
	if err != nil {
		return c.Status(404).SendString("Session not found")
	}
	return h.render(c, fiber.StatusOK, web.SessionFormPage{Title: "Edit Session", Session: s})
}

func (h *Handler) SessionUpdate(c fiber.Ctx) error {
	actor := middleware.GetUser(c)
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}
	start, err1 := parseDate(c.FormValue("start_date"))
	end, err2 := parseDate(c.FormValue("end_date"))
	if err1 != nil || err2 != nil {
		h.flash(c, "Invalid dates", true)
		return c.Redirect().To("/sessions/" + id.String() + "/edit")
	}
	req := dto.AcademicSessionRequest{
		Name:      c.FormValue("name"),
		StartDate: start,
		EndDate:   end,
		IsActive:  c.FormValue("is_active") == "on" || c.FormValue("is_active") == "true",
	}
	if errs := h.validate.Validate(req); len(errs) > 0 {
		h.flash(c, validator.FirstError(errs), true)
		return c.Redirect().To("/sessions/" + id.String() + "/edit")
	}
	if _, err := h.services.Sessions.Update(c.Context(), id, req, actor.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
		return c.Redirect().To("/sessions/" + id.String() + "/edit")
	}
	return c.Redirect().To("/sessions")
}

func (h *Handler) SessionDelete(c fiber.Ctx) error {
	actor := middleware.GetUser(c)
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}
	if err := h.services.Sessions.Delete(c.Context(), id, actor.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
	}
	return c.Redirect().To("/sessions")
}

func (h *Handler) SessionActivate(c fiber.Ctx) error {
	actor := middleware.GetUser(c)
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}
	if _, err := h.services.Sessions.SetActive(c.Context(), id, actor.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
	}
	return c.Redirect().To("/sessions")
}

func (h *Handler) AuditList(c fiber.Ctx) error {
	logs, err := h.services.Audit.ListRecent(c.Context(), 50)
	if err != nil {
		return c.Status(500).SendString("Failed to load audit logs")
	}
	return h.render(c, fiber.StatusOK, web.AuditListPage{Logs: logs})
}
