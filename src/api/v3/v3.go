package v3

import (
	"github.com/SevenTV/GQL/src/api/middleware"
	"github.com/SevenTV/GQL/src/api/v3/gql"
	"github.com/SevenTV/GQL/src/global"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func API(gCtx global.Context, app fiber.Router) { // Define CORS rules
	app.Use(cors.New(cors.Config{
		AllowOrigins:  "*",
		ExposeHeaders: "X-Created-ID",
		AllowMethods:  "GET,POST",
	}))

	app.Use(middleware.Auth(gCtx, true))
	gql.GQL(gCtx, app)
}
