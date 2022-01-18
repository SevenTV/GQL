package gql

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/SevenTV/Common/utils"
	"github.com/SevenTV/GQL/graph/generated"
	"github.com/SevenTV/GQL/src/global"
	"github.com/SevenTV/GQL/src/instance"
	"github.com/SevenTV/GQL/src/server/v3/gql/cache"
	"github.com/SevenTV/GQL/src/server/v3/gql/complexity"
	"github.com/SevenTV/GQL/src/server/v3/gql/helpers"
	"github.com/SevenTV/GQL/src/server/v3/gql/loaders"
	"github.com/SevenTV/GQL/src/server/v3/gql/middleware"
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
		Resolvers:  resolvers.New(types.Resolver{Ctx: gCtx}),
		Directives: middleware.New(gCtx),
		Complexity: complexity.New(gCtx),
	}))

	schema.Use(&extension.ComplexityLimit{
		Func: func(ctx context.Context, rc *graphql.OperationContext) int {
			// we can define limits here
			return 75
		},
	})

	schema.Use(extension.Introspection{})
	schema.Use(extension.AutomaticPersistedQuery{
		Cache: cache.NewRedisCache(gCtx, instance.RedisPrefix+":", time.Hour*6),
	})

	loader := loaders.New(gCtx)

	// handleRequest: Process a GQL query, from either a GET or POST
	handleRequest := func(c *fiber.Ctx, req gqlRequest) error {
		ctx := context.WithValue(c.Context(), helpers.UserKey, c.Locals("user"))
		clientIP := base64.URLEncoding.EncodeToString(utils.S2B(c.Get("Cf-Connecting-IP", c.IP())))

		{
			key := fmt.Sprintf("%s:rate-limits:client-ip:%s", instance.RedisPrefix, clientIP)
			bannedKey := fmt.Sprintf("%s:banned-ips:client-ip:%s", instance.RedisPrefix, clientIP)

			pipe := gCtx.Inst().Redis.Pipeline(ctx)

			// Check if the user is allowed to make queries
			bannedCmd := pipe.Exists(ctx, bannedKey)
			pipe.SetNX(ctx, key, 0, time.Minute)
			incrCmd := pipe.Incr(ctx, key)
			ttlCmd := pipe.TTL(ctx, key)
			_, err := pipe.Exec(ctx)
			if err == nil {
				if bannedCmd.Val() != 0 {
					return c.Status(fiber.StatusForbidden).JSON(&fiber.Map{
						"status": fiber.ErrForbidden,
						"error":  "You are temporarily blocked from using this API",
					})
				}

				total := incrCmd.Val()
				ttl := ttlCmd.Val()

				c.Set("X-GQL-RateLimit-Time", strconv.Itoa(int(ttl/time.Second)))
				c.Set("X-GQL-RateLimit-Limit", strconv.Itoa(int(gCtx.Config().Http.QuotaDefaultLimit)))

				if total > int64(gCtx.Config().Http.QuotaDefaultLimit) {
					c.Set("X-GQL-RateLimit-Remaining", strconv.Itoa(0))
					return c.Status(fiber.StatusTooManyRequests).JSON(&fiber.Map{
						"status": fiber.ErrTooManyRequests,
						"error":  "You are being rate limited",
						"reason": fmt.Sprintf("Quota Exceeded: 0 points left out of %d", gCtx.Config().Http.QuotaDefaultLimit),
					})
				} else {
					c.Set("X-GQL-RateLimit-Remaining", strconv.Itoa(int(int64(gCtx.Config().Http.QuotaDefaultLimit)-total)))
				}
			} else {
				logrus.WithError(err).Warn("failed to check redis for rate-limits/banned, ignoring")
			}
		}

		// Execute the query
		result := schema.Process(context.WithValue(ctx, loaders.LoadersKey, loader), graphql.RawParams{
			Query:         req.Query,
			OperationName: req.OperationName,
			Variables:     req.Variables,
		})

		return c.Status(result.Status).JSON(result.Response)
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
