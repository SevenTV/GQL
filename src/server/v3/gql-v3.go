package v3

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/SevenTV/Common/utils"
	"github.com/SevenTV/GQL/src/server/v3/resolvers"
	"github.com/gobuffalo/packr/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
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

func GQL(app fiber.Router) {
	// Load the schema
	box := packr.New("gqlv3", "./schema")
	var (
		sch1 string
		sch2 string
		sch3 string
		err  error
	)
	if sch1, err = box.FindString("query.gql"); err != nil {
		panic(err)
	} // query.gql: the available queries
	if sch2, err = box.FindString("emotes.gql"); err != nil {
		panic(err)
	} // emotes.gql: emote-related types
	if sch3, err = box.FindString("users.gql"); err != nil {
		panic(err)
	} // users.gql: user-related types

	// Build & parse the schema
	s := strings.Builder{}
	if _, err = s.WriteString(sch1); err != nil {
		panic(err)
	}
	if _, err = s.WriteString(sch2); err != nil {
		panic(err)
	}
	if _, err = s.WriteString(sch3); err != nil {
		panic(err)
	}
	schema := graphql.MustParseSchema(s.String(), resolvers.Resolver(), graphql.UseFieldResolvers(), graphql.MaxDepth(5))

	// Define CORS rules
	app.Use(cors.New(cors.Config{
		AllowOrigins:  "*",
		ExposeHeaders: "X-Created-ID",
		AllowMethods:  "GET,POST,PUT,PATCH,DELETE",
	}))

	// handleRequest: Process a GQL query, from either a GET or POST
	handleRequest := func(c *fiber.Ctx, req gqlRequest) error {
		ctx := context.WithValue(context.Background(), utils.Key("user"), c.Locals("user")) // Add auth user to context
		ctx = context.WithValue(ctx, utils.Key("request"), c)                               // Add request to context

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
