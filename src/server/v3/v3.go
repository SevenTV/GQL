package v3

import (
	"fmt"
	"time"

	"github.com/SevenTV/GQL/src/auth"
	"github.com/gofiber/fiber/v2"
)

func API(app fiber.Router) fiber.Router {
	api := app.Group("/v3")

	gql := GQL(api)
	auth.Listen(api)

	tok, err := auth.SignJWT(auth.JWTClaimOptions{
		UserID:       "YEAHBUT7TV",
		TokenVersion: 0,
		StandardClaims: auth.StandardClaims{
			Audience:  "7tv.app",
			ExpiresAt: time.Now().Add(time.Hour * 336).UnixMilli(),
			IssuedAt:  time.Now().UnixMilli(),
			Issuer:    "7tv.api.v3",
		},
	})
	fmt.Println(tok, err)

	return gql
}
