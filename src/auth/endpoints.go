package auth

import "github.com/gofiber/fiber/v2"

func Listen(router fiber.Router) {
	router.Get("/oauth2/callback/:platform", func(c *fiber.Ctx) error {
		return nil
	})
}
