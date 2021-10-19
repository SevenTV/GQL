package middleware

import (
	"strings"

	"github.com/SevenTV/Common/auth"
	"github.com/SevenTV/Common/mongo"
	"github.com/SevenTV/Common/structures"
	"github.com/SevenTV/GQL/src/global"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Auth(gCtx global.Context) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		// Parse token from header
		h := c.Get("Authorization")
		s := strings.Split(h, "Bearer ")
		if len(s) != 2 {
			return c.Status(401).JSON(&fiber.Map{"error": "Bad Authorization Header"})
		}
		t := s[1]

		// Verify the token
		_, claims, err := auth.VerifyJWT(gCtx.Config().Credentials.JWTSecret, strings.Split(t, "."))
		if err != nil {
			return c.Status(401).JSON(&fiber.Map{"error": err.Error()})
		}

		// User ID from parsed token
		u := claims["u"]
		if u == nil {
			return c.Status(401).JSON(&fiber.Map{"error": "Bad Token"})
		}
		userID, err := primitive.ObjectIDFromHex(u.(string))
		if err != nil {
			return c.Status(401).JSON(&fiber.Map{"error": err.Error()})
		}

		// Version of parsed token
		var user *structures.User
		v := claims["v"].(float64)
		if err = gCtx.Inst().Mongo.Collection(mongo.CollectionNameUsers).FindOne(c.Context(), bson.M{
			"_id": userID,
		}).Decode(&user); err == mongo.ErrNoDocuments {
			return c.Status(401).JSON(&fiber.Map{"error": "Token has Unknown Bound User"})
		} else if err != nil {
			logrus.WithError(err).Error("mongo")
			return c.SendStatus(500)
		}
		if user.TokenVersion != v {
			return c.Status(401).JSON(&fiber.Map{"error": "Token Version Mismatch"})
		}

		c.Locals("user", user)
		return c.Next()
	}
}
