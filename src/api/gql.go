package api

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/executor"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/SevenTV/GQL/graph/generated"
	"github.com/SevenTV/GQL/src/api/middleware"
	"github.com/SevenTV/GQL/src/api/v3/gql/cache"
	"github.com/SevenTV/GQL/src/api/v3/gql/complexity"
	"github.com/SevenTV/GQL/src/api/v3/gql/helpers"
	"github.com/SevenTV/GQL/src/api/v3/gql/loaders"
	middlewarev3 "github.com/SevenTV/GQL/src/api/v3/gql/middleware"
	"github.com/SevenTV/GQL/src/api/v3/gql/resolvers"
	"github.com/SevenTV/GQL/src/api/v3/gql/types"
	wsTransport "github.com/SevenTV/GQL/src/api/websocket"
	"github.com/SevenTV/GQL/src/global"
	"github.com/fasthttp/websocket"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

func GqlHandler(gCtx global.Context, loader *loaders.Loaders) func(ctx *fasthttp.RequestCtx) {
	schema := generated.NewExecutableSchema(generated.Config{
		Resolvers:  resolvers.New(types.Resolver{Ctx: gCtx}),
		Directives: middlewarev3.New(gCtx),
		Complexity: complexity.New(gCtx),
	})
	srv := handler.New(schema)
	exec := executor.New(schema)

	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.Use(extension.Introspection{})

	srv.Use(&extension.ComplexityLimit{
		Func: func(ctx context.Context, rc *graphql.OperationContext) int {
			return 75
		},
	})

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: cache.NewRedisCache(gCtx, "", time.Hour*6),
	})

	srv.SetRecoverFunc(func(ctx context.Context, err interface{}) (userMessage error) {
		logrus.Error("panic in handler: ", err)
		return helpers.ErrInternalServerError
	})

	wsTransport := wsTransport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
		InitFunc: func(ctx context.Context, initPayload wsTransport.InitPayload) (context.Context, error) {
			authHeader := initPayload.Authorization()

			if strings.HasPrefix(authHeader, "Bearer ") {
				tok := strings.TrimPrefix(authHeader, "Bearer ")

				user, err := middleware.DoAuth(gCtx, tok)
				if err != nil {
					goto handler
				}

				ctx = context.WithValue(ctx, helpers.UserKey, user)
			}

		handler:
			return ctx, nil
		},
		Upgrader: websocket.FastHTTPUpgrader{
			CheckOrigin: func(ctx *fasthttp.RequestCtx) bool {
				return true
			},
		},
	}

	return func(ctx *fasthttp.RequestCtx) {
		lCtx := context.WithValue(context.WithValue(gCtx, loaders.LoadersKey, loader), helpers.UserKey, ctx.UserValue("user"))

		if wsTransport.Supports(ctx) {
			wsTransport.Do(ctx, lCtx, exec)
		} else {
			fasthttpadaptor.NewFastHTTPHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				srv.ServeHTTP(w, r.WithContext(lCtx))
			}))(ctx)
		}

	}
}
