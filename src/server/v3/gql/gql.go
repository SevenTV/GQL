package gql

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/SevenTV/Common/utils"
	"github.com/SevenTV/GQL/graph/generated"
	"github.com/SevenTV/GQL/src/global"
	"github.com/SevenTV/GQL/src/instance"
	"github.com/SevenTV/GQL/src/server/v3/gql/cache"
	"github.com/SevenTV/GQL/src/server/v3/gql/helpers"
	"github.com/SevenTV/GQL/src/server/v3/gql/resolvers"
	"github.com/SevenTV/GQL/src/server/v3/gql/types"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

// API v3 - GQL

type gqlRequest struct {
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables"`
	OperationName string                 `json:"operation_name"`
	RequestID     string                 `json:"request_id"`
}

func GQL(gCtx global.Context, app fiber.Router) {
	schema := NewServer(generated.NewExecutableSchema(generated.Config{
		Resolvers: resolvers.New(types.Resolver{Ctx: gCtx}),
	}))

	schema.Use(extension.FixedComplexityLimit(5))

	schema.Use(extension.Introspection{})
	schema.Use(extension.AutomaticPersistedQuery{
		Cache: cache.NewRedisCache(gCtx, instance.RedisPrefix, time.Hour*6),
	})
	// handleRequest: Process a GQL query, from either a GET or POST
	handleRequest := func(c *fiber.Ctx, req gqlRequest) error {
		ctx := context.WithValue(context.Background(), helpers.UserKey, c.Locals("user"))

		// Check if the user is allowed to make queries
		clientIP := base64.URLEncoding.EncodeToString(utils.S2B(c.Get("Cf-Connecting-IP", c.IP())))
		clientIPKey := fmt.Sprintf("%s:banned-ips:client-ip:%s", instance.RedisPrefix, clientIP)

		if badQueries, _ := gCtx.Inst().Redis.RawClient().Exists(ctx, clientIPKey).Result(); badQueries != 0 {
			return c.Status(fiber.StatusForbidden).JSON(&fiber.Map{
				"status": fiber.ErrForbidden,
				"error":  "You are temporarily blocked from using this API",
			})
		}

		// Execute the query
		result := schema.Process(ctx, graphql.RawParams{
			Query:         req.Query,
			OperationName: req.OperationName,
			Variables:     req.Variables,
		})

		// Set quota headers
		// c.Set("X-Quota-Limit", strconv.Itoa(int(quota.GetLimit())))
		// c.Set("X-Quota-Remaining", strconv.Itoa(int(quota.GetPoints())))
		// {
		// 	// Set quota info
		// 	usage := make(map[string]int32)
		// 	quota.Fields.Range(func(key, value interface{}) bool {
		// 		usage[key.(string)] = value.(int32)
		// 		return true
		// 	})
		// 	b, _ := json.Marshal(usage)

		// 	c.Set("X-Quota-Usage", utils.B2S(b))
		// }

		// if !quota.Check() {
		// 	// Temporarily block this query
		// 	pipeline := gCtx.Inst().Redis.RawClient().Pipeline()
		// 	pipeline.SetEX(ctx, redisKey, req.Query, time.Hour)
		// 	pipeline.LPush(ctx, clientIPKey, qh)
		// 	pipeline.Expire(ctx, clientIPKey, time.Hour)
		// 	if _, err := pipeline.Exec(ctx); err != nil {
		// 		logrus.WithError(err).Error("redis, pipeline.Exec")
		// 	}

		// 	return c.Status(fiber.StatusTooManyRequests).JSON(&fiber.Map{
		// 		"status": fiber.ErrTooManyRequests,
		// 		"error":  "You are being rate limited",
		// 		"reason": fmt.Sprintf("Quota Exceeded: %d points left out of %d", quota.GetPoints(), quota.GetLimit()),
		// 	})
		// }

		return c.Status(result.Status).JSON(result)
	}

	// Handle query via POST
	app.Post("/", func(c *fiber.Ctx) error {
		req := gqlRequest{}
		err := c.BodyParser(&req)
		if err != nil {
			logrus.WithError(err).Error("gql.v3, post(BodyParser)")
		}

		return handleRequest(c, req)
	})

	// Handle query via GET
	app.Get("/", func(c *fiber.Ctx) error {
		req := gqlRequest{}
		err := c.QueryParser(&req)
		if err != nil {
			logrus.WithError(err).Error("gql.v3, get(QueryParser)")
		}

		return handleRequest(c, req)
	})
}
