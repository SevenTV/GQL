package v3

import (
	"github.com/SevenTV/GQL/src/auth"
	"github.com/SevenTV/GQL/src/global"
	"github.com/gofiber/fiber/v2"
)

func API(gCtx global.Context, app fiber.Router) {
	GQL(app)
	auth.Listen(app)
}
