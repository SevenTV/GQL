package v2

import (
	"context"
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/SevenTV/Common/errors"
	"github.com/SevenTV/GQL/graph/v2/generated"
	"github.com/SevenTV/GQL/src/api/v2/complexity"
	"github.com/SevenTV/GQL/src/api/v2/helpers"
	"github.com/SevenTV/GQL/src/api/v2/loaders"
	"github.com/SevenTV/GQL/src/api/v2/middleware"
	"github.com/SevenTV/GQL/src/api/v2/resolvers"
	"github.com/SevenTV/GQL/src/api/v2/types"
	"github.com/SevenTV/GQL/src/api/v3/cache"
	"github.com/SevenTV/GQL/src/global"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

func GqlHandlerV2(gCtx global.Context) func(ctx *fasthttp.RequestCtx) {
	schema := generated.NewExecutableSchema(generated.Config{
		Resolvers:  resolvers.New(types.Resolver{Ctx: gCtx}),
		Directives: middleware.New(gCtx),
		Complexity: complexity.New(gCtx),
	})
	srv := handler.New(schema)

	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.Use(extension.Introspection{})

	srv.Use(&extension.ComplexityLimit{
		Func: func(ctx context.Context, rc *graphql.OperationContext) int {
			return 100
		},
	})

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: cache.NewRedisCache(gCtx, "", time.Hour*6),
	})

	srv.SetRecoverFunc(func(ctx context.Context, err interface{}) (userMessage error) {
		logrus.Error("panic in handler: ", err)
		return errors.ErrInternalServerError()
	})

	loader := loaders.New(gCtx)
	return func(ctx *fasthttp.RequestCtx) {
		lCtx := context.WithValue(context.WithValue(gCtx, loaders.LoadersKey, loader), helpers.UserKey, ctx.UserValue("user"))
		lCtx = context.WithValue(lCtx, helpers.RequestCtxKey, ctx)

		fasthttpadaptor.NewFastHTTPHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			srv.ServeHTTP(w, r.WithContext(lCtx))
		}))(ctx)

	}
}
