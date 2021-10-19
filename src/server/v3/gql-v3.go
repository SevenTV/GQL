package v3

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/SevenTV/Common/utils"
	"github.com/SevenTV/GQL/src/global"
	"github.com/SevenTV/GQL/src/instance"
	"github.com/SevenTV/GQL/src/server/v3/helpers"
	"github.com/SevenTV/GQL/src/server/v3/resolvers"
	"github.com/gobuffalo/packr/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/graph-gophers/graphql-go"
	"github.com/sirupsen/logrus"
)

// API v3 - GQL

type gqlRequest struct {
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables"`
	OperationName string                 `json:"operation_name"`
	RequestID     string                 `json:"request_id"`
}

type Query struct{}

func (*Query) HelloWorld() string {
	return "Hello, world!!"
}

func GQL(gCtx global.Context, app fiber.Router) {
	// Load the schema
	box := packr.New("gqlv3", "./schema")

	s := strings.Builder{}
	files := []string{"query.gql", "emotes.gql", "users.gql", "mutation.gql"}
	for _, f := range files {
		sch, err := box.FindString(f)
		if err != nil {
			logrus.WithError(err).Fatal("gql, schema, box")
		}

		if _, err = s.WriteString(sch); err != nil {
			logrus.WithError(err).Error("gql, schema, strings.Builder")
		}
	}
	schema := graphql.MustParseSchema(s.String(), resolvers.Resolver(gCtx), graphql.UseFieldResolvers(), graphql.MaxDepth(5))

	// handleRequest: Process a GQL query, from either a GET or POST
	handleRequest := func(c *fiber.Ctx, req gqlRequest) error {
		defaultQuota := gCtx.Config().Http.QuotaDefaultLimit
		quota := &helpers.Quota{
			C:      c,
			Limit:  &defaultQuota,
			Points: &defaultQuota,
			Fields: sync.Map{},
		}
		ctx := context.WithValue(context.Background(), helpers.UserKey, c.Locals("user")) // Add auth user to context
		ctx = context.WithValue(ctx, utils.Key("request"), c)
		ctx = context.WithValue(ctx, helpers.QuotaKey, quota) // Add request to context

		// Check if the user is allowed to make queries
		clientIP := base64.URLEncoding.EncodeToString(utils.S2B(c.Get("Cf-Connecting-IP", c.IP())))
		if badQueries, _ := gCtx.Inst().Redis.RawClient().HLen(ctx, fmt.Sprintf("%s:blocked-queries:client-ip:%s", instance.RedisPrefix, clientIP)).Result(); badQueries >= 25 {
			return c.Status(fiber.StatusForbidden).JSON(&fiber.Map{
				"status": fiber.ErrForbidden,
				"error":  "You are temporarily blocked from using this API",
			})
		}

		// Check if the query is allowed
		h := sha256.New()
		h.Write(utils.S2B(strings.Map(func(r rune) rune {
			if unicode.IsSpace(r) {
				return -1
			}
			return r
		}, req.Query)))
		qh := hex.EncodeToString((h.Sum(nil))) // the hash of this query
		redisKey := fmt.Sprintf("%s:blocked-queries:query:%s", instance.RedisPrefix, qh)
		if exists, err := gCtx.Inst().Redis.RawClient().Exists(ctx, redisKey).Result(); err != nil {
			logrus.WithError(err).Error("redis down? this request will pass without checking if it's blocked")
		} else if exists == 1 {
			return c.Status(fiber.StatusForbidden).JSON(&fiber.Map{
				"status": fiber.ErrForbidden,
				"error":  "This query is blocked",
			})
		}

		// Execute the query
		result := schema.Exec(ctx, req.Query, req.OperationName, req.Variables)

		status := 200
		if len(result.Errors) > 0 {
			status = 400
		}

		// Insert metadata into the response
		if c.Locals("meta") != nil {
			b, err := json.Marshal(c.Locals("meta"))
			if err != nil {
				logrus.WithError(err).Error("gql, json")
				return c.Status(500).JSON(fiber.Map{
					"error": "decoding query meta failed",
				})
			}

			newData := make([]byte, 0)
			newData = append(newData, utils.S2B(`{"metadata":`)...)
			newData = append(newData, b...)
			newData = append(newData, byte(','))
			newData = append(newData, result.Data[1:]...)
			result.Data = newData
		}

		// Set quota headers
		c.Set("X-Quota-Limit", strconv.Itoa(int(quota.GetLimit())))
		c.Set("X-Quota-Remaining", strconv.Itoa(int(quota.GetPoints())))
		{
			// Set quota info
			usage := make(map[string]int32)
			quota.Fields.Range(func(key, value interface{}) bool {
				usage[key.(string)] = value.(int32)
				return true
			})
			b, _ := json.Marshal(usage)

			c.Set("X-Quota-Usage", utils.B2S(b))
		}
		if !quota.Check() {
			// Temporarily block this query
			pipeline := gCtx.Inst().Redis.RawClient().Pipeline()
			pipeline.SetEX(ctx, redisKey, "", time.Hour)
			pipeline.HSet(ctx, fmt.Sprintf("%s:blocked-queries", instance.RedisPrefix), qh, clientIP)
			pipeline.HSet(ctx, fmt.Sprintf("%s:blocked-queries:client-ip:%s", instance.RedisPrefix, clientIP), qh, req.Query)
			pipeline.Exec(ctx)

			return c.Status(fiber.StatusTooManyRequests).JSON(&fiber.Map{
				"status": fiber.ErrTooManyRequests,
				"error":  "You are being rate limited",
				"reason": fmt.Sprintf("Quota Exceeded: %d points left out of %d", quota.GetPoints(), quota.GetLimit()),
			})
		}

		return c.Status(status).JSON(result)
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
