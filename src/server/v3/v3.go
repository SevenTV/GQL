package v3

import (
	"github.com/SevenTV/ThreeLetterAPI/src/server/v3/gql"
	"github.com/gofiber/fiber/v2"
)

func API(app fiber.Router) fiber.Router {
	api := app.Group("/v3")

	gql := gql.GQL(api)
	return gql
}
