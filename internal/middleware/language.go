package middleware

import (
	"pos-backend/pkg/i18n"

	"github.com/gofiber/fiber/v2"
)

// Language detects the preferred language from ?lang= or Accept-Language header
// and stores it in c.Locals("lang") for handlers to read via i18n.Lang(c).
func Language() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Locals("lang", i18n.Resolve(c))
		return c.Next()
	}
}
