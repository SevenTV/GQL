package gql

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/SevenTV/ThreeLetterAPI/src/server/v3/gql/resolvers"
	"github.com/SevenTV/ThreeLetterAPI/src/utils"
	"github.com/gobuffalo/packr/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/graph-gophers/graphql-go"
	log "github.com/sirupsen/logrus"
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

func GQL(app fiber.Router) fiber.Router {
	gql := app.Group("/gql")

	// Load the schema
	box := packr.New("gqlv3", "./schema")
	sch1, err := box.FindString("query.gql")  // query.gql: the available queries
	sch2, err := box.FindString("emotes.gql") // emotes.gql: emote-related types
	sch3, err := box.FindString("users.gql")  // users.gql: user-related types
	if err != nil {
		panic(err)
	}

	// Build & parse the schema
	s := strings.Builder{}
	_, err = s.WriteString(sch1)
	_, err = s.WriteString(sch2)
	_, err = s.WriteString(sch3)
	if err != nil {
		panic(err)
	}
	schema := graphql.MustParseSchema(s.String(), resolvers.Resolver(), graphql.UseFieldResolvers(), graphql.MaxDepth(5))

	// Define CORS rules
	gql.Use(cors.New(cors.Config{
		AllowOrigins:  "*",
		ExposeHeaders: "X-Created-ID",
		AllowMethods:  "GET,POST,PUT,PATCH,DELETE",
	}))

	// handleRequest: Process a GQL query, from either a GET or POST
	handleRequest := func(c *fiber.Ctx, req gqlRequest) error {
		ctx := context.WithValue(context.Background(), utils.UserKey, c.Locals("user")) // Add auth user to context
		ctx = context.WithValue(ctx, utils.ReqKey, c)                                   // Add request to context

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
				log.WithError(err).Error("gql, json")
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

		return c.Status(status).JSON(result)
	}

	// Handle query via POST
	gql.Post("/", func(c *fiber.Ctx) error {
		req := gqlRequest{}
		err := c.BodyParser(&req)
		if err != nil {
			log.WithError(err).Error("gql.v3, post(BodyParser)")
		}

		return handleRequest(c, req)
	})

	// Handle query via GET
	gql.Get("/", func(c *fiber.Ctx) error {
		req := gqlRequest{}
		err := c.QueryParser(&req)
		if err != nil {
			log.WithError(err).Error("gql.v3, get(QueryParser)")
		}

		return handleRequest(c, req)
	})

	return gql
}
