package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"log/slog"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"

	"github.com/school-management/pos/internal/auth"
	"github.com/school-management/pos/internal/dto"
	"github.com/school-management/pos/internal/service"
)

const (
	ContextUserKey  = "auth_user"
	ContextClaimsKey = "jwt_claims"
	CSRFHeader      = "X-CSRF-Token"
	CSRFCookie      = "csrf_token"
	AccessCookie    = "access_token"
)

type Middleware struct {
	tokens  *auth.TokenManager
	authSvc *service.AuthService
	csrfKey []byte
	logger  *slog.Logger
}

func New(tokens *auth.TokenManager, authSvc *service.AuthService, csrfSecret string, logger *slog.Logger) *Middleware {
	return &Middleware{
		tokens:  tokens,
		authSvc: authSvc,
		csrfKey: []byte(csrfSecret),
		logger:  logger,
	}
}

func (m *Middleware) RequestLogger() fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		m.logger.Info("request",
			"method", c.Method(),
			"path", c.Path(),
			"status", c.Response().StatusCode(),
			"duration", time.Since(start).String(),
			"ip", c.IP(),
		)
		return err
	}
}

func (m *Middleware) Authenticate(required bool) fiber.Handler {
	return func(c fiber.Ctx) error {
		token := extractToken(c)
		if token == "" {
			if required {
				return c.Redirect().To("/login")
			}
			return c.Next()
		}

		claims, err := m.tokens.ParseToken(token)
		if err != nil {
			if required {
				clearAuthCookies(c)
				return c.Redirect().To("/login")
			}
			return c.Next()
		}

		user, err := m.authSvc.ValidateClaims(c.Context(), claims)
		if err != nil {
			if required {
				clearAuthCookies(c)
				return c.Redirect().To("/login")
			}
			return c.Next()
		}

		c.Locals(ContextUserKey, user)
		c.Locals(ContextClaimsKey, claims)
		return c.Next()
	}
}

func extractToken(c fiber.Ctx) string {
	if auth := c.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return c.Cookies(AccessCookie)
}

func clearAuthCookies(c fiber.Ctx) {
	clearSessionCookies(c)
}

func ClearStaleCookies(c fiber.Ctx) {
	clearSessionCookies(c)
}

func clearSessionCookies(c fiber.Ctx) {
	expired := time.Now().Add(-time.Hour)
	names := []string{AccessCookie, CSRFCookie, "flash", "flash_type", "last_app_no", "last_app_token"}
	for _, name := range names {
		c.Cookie(&fiber.Cookie{
			Name:     name,
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			Expires:  expired,
			HTTPOnly: name != "flash" && name != "flash_type",
		})
	}
}

func (m *Middleware) RequirePermission(permission string) fiber.Handler {
	return func(c fiber.Ctx) error {
		user, ok := c.Locals(ContextUserKey).(*dto.AuthUser)
		if !ok || user == nil {
			return c.Status(fiber.StatusUnauthorized).SendString("Unauthorized")
		}
		if user.RoleSlug == "admin" {
			return c.Next()
		}
		for _, p := range user.Permissions {
			if p == permission {
				return c.Next()
			}
		}
		return c.Status(fiber.StatusForbidden).SendString("Forbidden")
	}
}

func (m *Middleware) CSRFGenerate() fiber.Handler {
	return func(c fiber.Ctx) error {
		token := c.Cookies(CSRFCookie)
		if token == "" {
			var err error
			token, err = generateCSRFToken()
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).SendString("CSRF error")
			}
			c.Cookie(&fiber.Cookie{
				Name:     CSRFCookie,
				Value:    token,
				Path:     "/",
				HTTPOnly: true,
				SameSite: "Lax",
				MaxAge:   86400,
			})
		}
		c.Locals("csrf_token", token)
		return c.Next()
	}
}

func (m *Middleware) CSRFProtect() fiber.Handler {
	return func(c fiber.Ctx) error {
		if c.Method() == fiber.MethodGet || c.Method() == fiber.MethodHead || c.Method() == fiber.MethodOptions {
			return c.Next()
		}
		cookie := c.Cookies(CSRFCookie)
		header := c.Get(CSRFHeader)
		form := c.FormValue("csrf_token")
		token := header
		if token == "" {
			token = form
		}
		if cookie == "" || token == "" || subtle.ConstantTimeCompare([]byte(cookie), []byte(token)) != 1 {
			return c.Status(fiber.StatusForbidden).SendString("Invalid CSRF token")
		}
		return c.Next()
	}
}

func generateCSRFToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func GetUser(c fiber.Ctx) *dto.AuthUser {
	user, _ := c.Locals(ContextUserKey).(*dto.AuthUser)
	return user
}

func GetClaims(c fiber.Ctx) *auth.Claims {
	claims, _ := c.Locals(ContextClaimsKey).(*auth.Claims)
	return claims
}

func GetCSRF(c fiber.Ctx) string {
	if v, ok := c.Locals("csrf_token").(string); ok {
		return v
	}
	return c.Cookies(CSRFCookie)
}

func SetAuthCookie(c fiber.Ctx, token string, expires time.Time) {
	c.Cookie(&fiber.Cookie{
		Name:     AccessCookie,
		Value:    token,
		Path:     "/",
		HTTPOnly: true,
		SameSite: "Lax",
		Expires:  expires,
	})
}

func ParseJWTExpiry(token string) time.Time {
	t, _, err := new(jwt.Parser).ParseUnverified(token, &auth.Claims{})
	if err != nil {
		return time.Now().Add(15 * time.Minute)
	}
	if claims, ok := t.Claims.(*auth.Claims); ok && claims.ExpiresAt != nil {
		return claims.ExpiresAt.Time
	}
	return time.Now().Add(15 * time.Minute)
}
