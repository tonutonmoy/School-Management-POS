package middleware

import (
	"github.com/gofiber/fiber/v3"

	"github.com/school-management/pos/internal/i18n"
)

const LangCookie = "lang"

func (m *Middleware) Locale() fiber.Handler {
	return func(c fiber.Ctx) error {
		lang := i18n.Parse(c.Cookies(LangCookie))
		c.Locals("lang", lang)
		return c.Next()
	}
}

func GetLang(c fiber.Ctx) i18n.Locale {
	if v, ok := c.Locals("lang").(i18n.Locale); ok {
		return v
	}
	return i18n.EN
}

func SetLangCookie(c fiber.Ctx, lang i18n.Locale) {
	c.Cookie(&fiber.Cookie{
		Name:     LangCookie,
		Value:    string(lang),
		Path:     "/",
		MaxAge:   365 * 86400,
		SameSite: "Lax",
	})
}
