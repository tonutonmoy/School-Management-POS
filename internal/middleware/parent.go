package middleware

import (
	"github.com/gofiber/fiber/v3"

	"github.com/school-management/pos/internal/dto"
	"github.com/school-management/pos/internal/model"
)

func (m *Middleware) RequireParent() fiber.Handler {
	return func(c fiber.Ctx) error {
		user := GetUser(c)
		if user == nil {
			return c.Status(fiber.StatusUnauthorized).SendString("Unauthorized")
		}
		if user.RoleSlug == model.RoleAdmin || user.RoleSlug == model.RoleParent {
			return c.Next()
		}
		return c.Status(fiber.StatusForbidden).SendString("Forbidden")
	}
}

func HasPermission(user *dto.AuthUser, permission string) bool {
	if user == nil {
		return false
	}
	if user.RoleSlug == model.RoleAdmin {
		return true
	}
	for _, p := range user.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

func HomePath(user *dto.AuthUser) string {
	if user != nil && user.RoleSlug == model.RoleParent {
		return "/parent/dashboard"
	}
	return "/dashboard"
}
